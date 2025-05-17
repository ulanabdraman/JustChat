package usecase

import (
	"JustChat/internal/users/model"
	"JustChat/internal/users/repository/postgres"
	"context"
	"errors"
)

type UserUseCase interface {
	GetCurrentUser(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error) // <-- добавлено
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

func (uc *userUseCase) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error) {
	var users []*model.User

	for _, id := range userIDs {
		user, err := uc.repo.GetByID(ctx, id)
		if err != nil || user == nil {
			continue // Пропускаем ошибку или если пользователь не найден
		}
		users = append(users, user)
	}

	return users, nil
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
