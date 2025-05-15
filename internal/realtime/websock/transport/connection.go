package transport

import (
	authUsecase "JustChat/internal/auth/usecase"
	chatUsecase "JustChat/internal/chat/usecase"
	messageUsecase "JustChat/internal/messages/usecase"
	"JustChat/internal/realtime/websock/model"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Connection struct {
	ws     *websocket.Conn
	send   chan any
	hub    *Hub
	ctx    context.Context
	cancel context.CancelFunc
	chatID int64
	userID int64

	chatUsecase    chatUsecase.ChatUsecase
	messageUsecase messageUsecase.MessageUseCase
}

func ServeWS(hub *Hub, chatUC chatUsecase.ChatUsecase, msgUC messageUsecase.MessageUseCase, authUC authUsecase.JWTUsecase, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	token := r.Header.Get("Sec-WebSocket-Protocol")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Sec-WebSocket-Protocol", token)

	claims, err := authUC.ParseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-WebSocket-Protocol": []string{token},
	})
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	conn := &Connection{
		ws:             ws,
		send:           make(chan any, 256),
		hub:            hub,
		ctx:            ctx,
		cancel:         cancel,
		userID:         claims.UserID,
		chatUsecase:    chatUC,
		messageUsecase: msgUC,
	}

	go conn.writePump()
	conn.readPump()
}

// Обновленная версия writePump с обработкой ошибок
func (c *Connection) writePump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Если канал закрыт, отправляем сообщение о закрытии соединения
				_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Проверка на целостность данных перед отправкой
			data, err := json.Marshal(msg)
			if err != nil {
				log.Println("Failed to marshal message:", err)
				_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{}) // Закрываем соединение
				return
			}

			// Пытаемся отправить сообщение
			err = c.ws.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Failed to send message:", err)
				_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{}) // Закрываем соединение
				return
			}
		}
	}
}

// Внесем проверку chatID при подключении
func (c *Connection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.ws.Close()
	}()

	_, raw, err := c.ws.ReadMessage()
	if err != nil {
		log.Println("read error:", err)
		return
	}

	var init model.InitPayload
	if err := json.Unmarshal(raw, &init); err != nil {
		log.Println("invalid init payload:", err)
		return
	}

	// Проверяем, что chatID валиден
	if init.ChatID == 0 {
		log.Println("Invalid chatID:", init.ChatID)
		return
	}

	c.chatID = init.ChatID

	chat, err := c.chatUsecase.GetChatByID(c.ctx, init.ChatID, c.userID)
	if err != nil {
		log.Println("chat not found:", err)
		return
	}

	messages, err := c.messageUsecase.GetMessagesByChatID(c.ctx, init.ChatID, c.userID)
	if err != nil {
		log.Println("failed to get messages:", err)
		return
	}

	resp := model.InitResponse{
		Chat:     *chat,
		Messages: messages,
	}

	// Register the connection to the hub
	c.hub.register <- c

	// Send initial chat and messages info
	c.send <- resp

	// Connection just listens to incoming messages but doesn't handle them
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		// Simply ignore incoming messages here. All message logic is handled in the usecase.
	}
}
