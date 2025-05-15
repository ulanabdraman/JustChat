package usecase

import (
	"JustChat/internal/chatmembers/repository/postgres"
	"context"
	"errors"
)

type ChatMemberUseCase interface {
	GetRole(ctx context.Context, chatID, userID int64) (string, error)
	GetUsersByChat(ctx context.Context, chatID int64, userID int64) ([]int, error)
	GetChatsByUser(ctx context.Context, userID int64) ([]int, error)
	AddUserToChat(ctx context.Context, chatID, userID int64, role string, myuserID int64) error
	RemoveUserFromChat(ctx context.Context, chatID, userID int64, myuserID int64) error
	UpdateUserRole(ctx context.Context, chatID, userID int64, role string, myuserID int64) error
}

type chatMemberUseCase struct {
	repo repository.ChatMemberRepository
}

func NewChatMemberUseCase(repo repository.ChatMemberRepository) ChatMemberUseCase {
	return &chatMemberUseCase{repo: repo}
}

func (uc *chatMemberUseCase) GetRole(ctx context.Context, chatID, myuserID int64) (string, error) {
	return uc.repo.GetRole(ctx, chatID, myuserID)
}

func (uc *chatMemberUseCase) GetUsersByChat(ctx context.Context, chatID int64, myuserID int64) ([]int, error) {
	_, err := uc.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetUsersByChat(ctx, chatID)
}

func (uc *chatMemberUseCase) GetChatsByUser(ctx context.Context, myuserID int64) ([]int, error) {
	return uc.repo.GetChatsByUser(ctx, myuserID)
}

func (uc *chatMemberUseCase) AddUserToChat(ctx context.Context, chatID, userID int64, role string, myuserID int64) error {
	if role != "admin" && role != "user" && role != "dialog" {
		return errors.New("invalid role")
	}
	if role != "admin" {
		myrole, err := uc.GetRole(ctx, chatID, myuserID)
		if err != nil {
			return err
		}
		if myrole != "admin" {
			return errors.New("not admin")
		}
	}
	return uc.repo.AddUserToChat(ctx, chatID, userID, role)
}

func (uc *chatMemberUseCase) RemoveUserFromChat(ctx context.Context, chatID, userID int64, myuserID int64) error {
	myrole, err := uc.GetRole(ctx, chatID, myuserID)
	if err != nil {
		return err
	}
	if myrole != "admin" {
		return errors.New("not admin")
	}
	return uc.repo.RemoveUserFromChat(ctx, chatID, userID)
}

func (uc *chatMemberUseCase) UpdateUserRole(ctx context.Context, chatID, userID int64, role string, myuserID int64) error {
	if role != "admin" && role != "user" && role != "dialog" {
		return errors.New("invalid role")
	}
	return uc.repo.UpdateUserRole(ctx, chatID, userID, role)
}
