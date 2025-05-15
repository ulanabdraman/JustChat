package usecase

import (
	"JustChat/internal/messages/model"
	"JustChat/internal/realtime/websock/transport"
	"encoding/json"
	"fmt"
)

type WebSockUsecase interface {
	SendMessage(jsonMessage []byte) error
}

type webSockUsecase struct {
	hub *transport.Hub
}

// Новый конструктор для WebSockUsecase
func NewWebSockUsecase(hub *transport.Hub) WebSockUsecase {
	return &webSockUsecase{hub: hub}
}

// Метод, который обрабатывает входящий JSON и вызывает Broadcast хаба
func (w *webSockUsecase) SendMessage(jsonMessage []byte) error {
	// Структура для извлечения chatID из JSON
	var messageData model.Message

	// Преобразуем JSON в структуру
	if err := json.Unmarshal(jsonMessage, &messageData); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if messageData.ChatID == 0 {
		return fmt.Errorf("invalid chatID")
	}

	// Вызов метода Broadcast из хаба
	w.hub.Broadcast(messageData)

	return nil
}
