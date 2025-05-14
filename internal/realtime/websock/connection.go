package websock

import (
	AUC "JustChat/internal/auth/usecase"
	CUC "JustChat/internal/chat/usecase"
	"JustChat/internal/messages/model"
	MUC "JustChat/internal/messages/usecase"
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

	chatUsecase    CUC.ChatUsecase
	messageUsecase MUC.MessageUseCase
}

func ServeWS(hub *Hub, chatUC CUC.ChatUsecase, msgUC MUC.MessageUseCase, authUC AUC.JWTUsecase, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // или указать origin явно
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	token := r.Header.Get("Sec-WebSocket-Protocol")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	// Проставим обратно, иначе браузер может закрыть соединение
	w.Header().Set("Sec-WebSocket-Protocol", token)

	claims, err := authUC.ParseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // ⚠️ в проде нужно проверять Origin
		},
	}

	ws, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-WebSocket-Protocol": []string{token},
	})

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

	var init InitPayload
	if err := json.Unmarshal(raw, &init); err != nil {
		log.Println("invalid init payload:", err)
		return
	}

	c.chatID = init.ChatID

	chat, err := c.chatUsecase.GetChatByID(c.ctx, init.ChatID)
	if err != nil {
		log.Println("chat not found:", err)
		return
	}

	messages, err := c.messageUsecase.GetMessagesByChatID(c.ctx, init.ChatID)
	if err != nil {
		log.Println("failed to get messages:", err)
		return
	}

	resp := InitResponse{
		Chat:     *chat,
		Messages: messages,
	}

	c.hub.register <- c
	c.send <- resp

	for {
		_, rawMessage, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		var message model.Message
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Println("Failed to unmarshal data into Message:", err)
			continue
		}

		message.CreatorID = c.userID
		message.ChatID = c.chatID

		send, err := c.messageUsecase.SaveMessage(c.ctx, &message)
		if err != nil {
			log.Println(err, message)
		}

		c.hub.Broadcast(c.chatID, send)
	}
}

func (c *Connection) writePump() {
	defer c.ws.Close()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Если канал закрыт, завершаем соединение
				_ = c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Println("Failed to marshal message:", err)
				continue
			}
			err = c.ws.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Failed to send message:", err)
				return
			}
		}
	}
}
