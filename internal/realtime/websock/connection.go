package websock

import (
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

func ServeWS(hub *Hub, chatUC CUC.ChatUsecase, msgUC MUC.MessageUseCase, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // ⚠️ Разрешаем все Origin
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	conn := &Connection{
		ws:             ws,
		send:           make(chan any, 256),
		hub:            hub,
		ctx:            ctx,
		cancel:         cancel,
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
	c.userID = init.UserID

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

	// После получения данных, передаем их в канал для отправки
	c.hub.register <- c
	c.send <- resp

	// слушаем дальнейшие сообщения
	for {
		_, rawMessage, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		// Обработка сообщения перед сохранением в базу
		var message model.Message
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Println("Failed to unmarshal data into Message:", err)
			continue
		}

		// Сохранение сообщения в базе данных
		send, err := c.messageUsecase.SaveMessage(c.ctx, &message)
		if err != nil {
			log.Println(err, message)
		}

		// Далее можно передать это сообщение в канал для дальнейшей отправки
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
