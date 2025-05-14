package usecase

import (
	"JustChat/internal/users/model"
	"JustChat/internal/users/repository/postgres"
	"context"
	"errors"
)

type UserUseCase interface {
	GetCurrentUser(ctx context.Context, userID int64) (*model.User, error)
	GetUserChats(ctx context.Context, userID int64) ([]int64, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
}

type userUseCase struct {
	repo postgres.UserRepo
}

func NewUserUseCase(repo postgres.UserRepo) UserUseCase {
	return &userUseCase{repo: repo}
}

func (uc *userUseCase) GetCurrentUser(ctx context.Context, userID int64) (*model.User, error) {
	return uc.repo.GetByID(ctx, userID)
}

func (uc *userUseCase) GetUserChats(ctx context.Context, userID int64) ([]int64, error) {
	return uc.repo.GetChatsByUserID(ctx, userID)
}

func (uc *userUseCase) CreateUser(ctx context.Context, user *model.User) error {
	if user.Type == "" {
		user.Type = "user"
	}

	if user.Type != "admin" && user.Type != "user" && user.Type != "bot" {
		return errors.New("invalid user type")
	}

	return uc.repo.Create(ctx, user)
}

func (uc *userUseCase) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return uc.repo.GetByUsername(ctx, username)
}
