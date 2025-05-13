package usecase

import (
	"JustChat/internal/chat/model"
	"JustChat/internal/chat/repository/postgres"
	"context"
	"errors"
)

type ChatUsecase interface {
	GetChatByID(ctx context.Context, chatID int64) (*model.Chat, error)
	CreateChat(ctx context.Context, chat *model.Chat) error
	UpdateChatName(ctx context.Context, chatID int64, name string) error
	AddUserToChat(ctx context.Context, chatID int64, userID int64) error
	RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
}

type chatUseCase struct {
	repo postgres.ChatRepository
}

func NewChatUseCase(chatRepo postgres.ChatRepository) ChatUsecase {
	return &chatUseCase{
		repo: chatRepo,
	}
}

func (uc *chatUseCase) GetChatByID(ctx context.Context, chatID int64) (*model.Chat, error) {
	chat, err := uc.repo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return chat, nil
}
func (uc *chatUseCase) CreateChat(ctx context.Context, chat *model.Chat) error {
	return uc.repo.Create(ctx, chat)
}

func (uc *chatUseCase) UpdateChatName(ctx context.Context, chatID int64, name string) error {
	chat, err := uc.repo.GetByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat.Type == "group" {
		chat.Name = name
	} else {
		return errors.New("chat is not group")
	}

	if err := uc.repo.Update(ctx, chat); err != nil {
		return err
	}
	return nil
}

func (uc *chatUseCase) AddUserToChat(ctx context.Context, chatID int64, userID int64) error {
	if err := uc.repo.AddUserToChat(ctx, chatID, userID); err != nil {
		return err
	}
	return nil
}

func (uc *chatUseCase) RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error {
	if err := uc.repo.RemoveUserFromChat(ctx, chatID, userID); err != nil {
		return err
	}
	return nil
}

func (uc *chatUseCase) DeleteChat(ctx context.Context, chatID int64) error {
	if err := uc.repo.Delete(ctx, chatID); err != nil {
		return err
	}
	return nil
}
