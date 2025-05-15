package transport

import (
	"JustChat/internal/messages/model"
	"context"
	"log"
	"sync"
)

type Hub struct {
	mu         sync.RWMutex
	clients    map[int64]map[*Connection]bool // chatID -> connections
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan BroadcastMessage
}

type BroadcastMessage struct {
	ChatID  int64
	Message model.Message // JSON-сообщение
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Connection]bool),
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:  make(chan BroadcastMessage),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Контекст отменён — выход из Run
			log.Println("Hub is shutting down due to context cancellation")

			// Опционально: закрываем все подключения
			h.mu.Lock()
			for chatID, conns := range h.clients {
				for conn := range conns {
					close(conn.send)
					delete(h.clients[chatID], conn)
				}
			}
			h.mu.Unlock()

			return

		case conn := <-h.register:
			h.mu.Lock()
			if h.clients[conn.chatID] == nil {
				h.clients[conn.chatID] = make(map[*Connection]bool)
			}
			h.clients[conn.chatID][conn] = true
			log.Printf("New connection registered for chat %d", conn.chatID)
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn.chatID][conn]; ok {
				delete(h.clients[conn.chatID], conn)
				close(conn.send)
				log.Printf("Connection unregistered for chat %d", conn.chatID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			log.Printf("Broadcasting message to chat %d", msg.ChatID)
			for conn := range h.clients[msg.ChatID] {
				select {
				case conn.send <- msg.Message:
				default:
					log.Println("Error: failed to send message to connection.")
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(chat model.Message) {
	// Преобразуем сообщение в BroadcastMessage и отправляем в канал broadcast
	h.broadcast <- BroadcastMessage{
		ChatID:  chat.ChatID,
		Message: chat,
	}
}
