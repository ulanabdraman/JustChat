package usecase

import (
	"JustChat/internal/messages/model"
	"JustChat/internal/messages/repository/postgres"
	"context"
)

type MessageUseCase interface {
	GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error)
	GetMessagesByChatID(ctx context.Context, chatID int64) ([]model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID int64) error
}
type messageUseCase struct {
	repo postgres.MessageRepo
}

func NewMessageUseCase(messageRepo postgres.MessageRepo) MessageUseCase {
	return &messageUseCase{repo: messageRepo}
}
func (m messageUseCase) GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error) {
	message, err := m.repo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return message, nil
}
func (m messageUseCase) DeleteMessage(ctx context.Context, messageID int64) error {
	err := m.repo.DeleteByID(ctx, messageID)
	if err != nil {
		return err
	}
	return nil
}
func (m messageUseCase) GetMessagesByChatID(ctx context.Context, chatID int64) ([]model.Message, error) {
	messages, err := m.repo.GetByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
func (m messageUseCase) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	return m.repo.SaveMessage(ctx, message)
}
