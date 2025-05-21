package usecase

import (
	ChatMembersUC "JustChat/internal/chatmembers/usecase"
	"JustChat/internal/messages/model"
	"JustChat/internal/messages/repository"
	"context"
	"encoding/json"
	"errors"
	"log"
)

type MessageUseCase interface {
	GetMessageByID(ctx context.Context, messageID int64, myuserID int64) (*model.Message, error)
	GetMessagesByChatID(ctx context.Context, chatID int64, myuserID int64) ([]model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message, myuserID int64) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID int64, myuserID int64) error
}
type messageUseCase struct {
	repo              repository.MessageRepo
	messageCh         chan<- []byte
	chatMemberUsecase ChatMembersUC.ChatMemberUseCase
}

func NewMessageUseCase(messageRepo repository.MessageRepo, messageCh chan<- []byte, chatmemberUsecase ChatMembersUC.ChatMemberUseCase) MessageUseCase {
	return &messageUseCase{repo: messageRepo, messageCh: messageCh, chatMemberUsecase: chatmemberUsecase}
}
func (m messageUseCase) GetMessageByID(ctx context.Context, messageID int64, myuserID int64) (*model.Message, error) {
	message, err := m.repo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	log.Println(message)
	_, err = m.chatMemberUsecase.GetRole(ctx, message.ChatID, myuserID)
	if err != nil {
		return nil, err
	}
	return message, nil
}
func (m messageUseCase) DeleteMessage(ctx context.Context, messageID int64, myuserID int64) error {
	message, err := m.repo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	role, err := m.chatMemberUsecase.GetRole(ctx, message.ChatID, myuserID)
	if err != nil {
		return err
	}
	if role != "admin" && message.CreatorID != myuserID {
		return errors.New("Not admin or not own the message")
	}
	err = m.repo.DeleteByID(ctx, messageID)
	if err != nil {
		return err
	}
	return nil
}
func (m messageUseCase) GetMessagesByChatID(ctx context.Context, chatID int64, myuserID int64) ([]model.Message, error) {
	_, err := m.chatMemberUsecase.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return nil, err
	}
	messages, err := m.repo.GetByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
func (m messageUseCase) SaveMessage(ctx context.Context, message *model.Message, myuserID int64) (*model.Message, error) {
	_, err := m.chatMemberUsecase.GetRole(ctx, message.ChatID, myuserID)
	if err != nil {
		return nil, err
	}
	message.Type = "message"
	savedmessage, err := m.repo.SaveMessage(ctx, message)
	if err != nil {
		return nil, err
	}
	jsonMessage, err := json.Marshal(savedmessage)
	if err != nil {
		return nil, err
	}
	m.messageCh <- jsonMessage
	return message, nil
}
