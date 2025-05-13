package postgres

import (
	"JustChat/internal/users/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetChatsByUserID(ctx context.Context, userID int64) ([]int64, error) // можно заменить на Chat model позже
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, type, username, online, was_online, created_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Type, &user.Username, &user.Online, &user.WasOnline, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getByID user repo: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetChatsByUserID(ctx context.Context, userID int64) ([]int64, error) {
	query := `SELECT chat_id FROM chat_users WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("getChatsByUserID user repo: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		chatIDs = append(chatIDs, chatID)
	}
	return chatIDs, nil
}
func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users(type,username) VALUES ($1,$2)`
	_, err := r.db.ExecContext(ctx, query, user.Type, user.Username)
	if err != nil {
		return fmt.Errorf("create user repo: %w", err)
	}
	return nil
}
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}
