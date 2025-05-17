package usecase

import (
	"JustChat/internal/chat/model"
	"JustChat/internal/chat/repository/postgres"
	"JustChat/internal/chatmembers/usecase"
	"context"
	"errors"
)

type ChatUsecase interface {
	GetChatByID(ctx context.Context, chatID int64, myuserID int64) (*model.Chat, error)
	GetChatsByIDs(ctx context.Context, chatIDs []int64, myuserID int64) ([]*model.Chat, error)
	CreateChat(ctx context.Context, chat *model.Chat, myuserID int64) (int64, error)
	UpdateChatName(ctx context.Context, chatID int64, name string, myuserID int64) error
	DeleteChat(ctx context.Context, chatID int64, myuserID int64) error
}

type chatUseCase struct {
	repo               postgres.ChatRepository
	chatMembersUsecase usecase.ChatMemberUseCase
}

func NewChatUseCase(chatRepo postgres.ChatRepository, chatMembersUsecase usecase.ChatMemberUseCase) ChatUsecase {
	return &chatUseCase{
		repo:               chatRepo,
		chatMembersUsecase: chatMembersUsecase,
	}
}

func (uc *chatUseCase) GetChatByID(ctx context.Context, chatID int64, myuserID int64) (*model.Chat, error) {
	_, err := uc.chatMembersUsecase.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return nil, err
	}
	chat, err := uc.repo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return chat, nil
}
func (uc *chatUseCase) GetChatsByIDs(ctx context.Context, chatIDs []int64, myuserID int64) ([]*model.Chat, error) {
	var result []*model.Chat

	for _, chatID := range chatIDs {
		_, err := uc.chatMembersUsecase.GetRole(ctx, chatID, myuserID)
		if err != nil {
			continue
		}

		chat, err := uc.repo.GetByID(ctx, chatID)
		if err != nil {
			continue
		}

		result = append(result, chat)
	}

	return result, nil
}

func (uc *chatUseCase) CreateChat(ctx context.Context, chat *model.Chat, myuserID int64) (int64, error) {
	chatID, err := uc.repo.Create(ctx, chat)
	if err != nil {
		return 0, err
	}
	return chatID, nil
}

func (uc *chatUseCase) UpdateChatName(ctx context.Context, chatID int64, name string, myuserID int64) error {
	myrole, err := uc.chatMembersUsecase.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return err
	}
	if myrole == "user" {
		return errors.New("not admin")
	}
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

func (uc *chatUseCase) DeleteChat(ctx context.Context, chatID int64, myuserID int64) error {
	myrole, err := uc.chatMembersUsecase.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return err
	}
	if myrole == "user" {
		return errors.New("not admin")
	}
	if err := uc.repo.Delete(ctx, chatID); err != nil {
		return err
	}
	return nil
}
