package model

import (
	chats "JustChat/internal/chat/model"
	messages "JustChat/internal/messages/model"
)

type InitPayload struct {
	ChatID int64 `json:"chat_id"`
}
type InitResponse struct {
	Chat     chats.Chat         `json:"chat"`
	Messages []messages.Message `json:"messages"`
}
type Payload struct {
	ChatID  int64  `json:"chat_id"`
	Message []byte `json:"message"`
}
