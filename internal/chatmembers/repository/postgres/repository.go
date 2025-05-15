package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type ChatMemberRepository interface {
	GetRole(ctx context.Context, chatID, userID int64) (string, error)
	GetUsersByChat(ctx context.Context, chatID int64) ([]int, error)
	GetChatsByUser(ctx context.Context, userID int64) ([]int, error)
	AddUserToChat(ctx context.Context, chatID, userID int64, role string) error
	RemoveUserFromChat(ctx context.Context, chatID, userID int64) error
	UpdateUserRole(ctx context.Context, chatID, userID int64, role string) error
}

type chatMemberRepo struct {
	db *sqlx.DB
}

func NewChatMemberRepository(db *sqlx.DB) ChatMemberRepository {
	return &chatMemberRepo{db: db}
}

func (r *chatMemberRepo) GetRole(ctx context.Context, chatID, userID int64) (string, error) {
	// Выполняем SQL-запрос, чтобы получить роль пользователя в чате
	query := `SELECT role FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	var role string
	row := r.db.QueryRowContext(ctx, query, chatID, userID)

	err := row.Scan(&role)
	if err != nil {
		return "", fmt.Errorf("error getting role for user_id %d in chat_id %d: %w", userID, chatID, err)
	}

	return role, nil
}

func (r *chatMemberRepo) GetUsersByChat(ctx context.Context, chatID int64) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id FROM chat_members WHERE chat_id = $1`, chatID)
	if err != nil {
		return nil, fmt.Errorf("error getting users by chat: %w", err)
	}
	defer rows.Close()

	var users []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("error scanning user id: %w", err)
		}
		users = append(users, userID)
	}

	return users, nil
}

func (r *chatMemberRepo) GetChatsByUser(ctx context.Context, userID int64) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT chat_id FROM chat_members WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting chats by user: %w", err)
	}
	defer rows.Close()

	var chats []int
	for rows.Next() {
		var chatID int
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("error scanning chat id: %w", err)
		}
		chats = append(chats, chatID)
	}

	return chats, nil
}

func (r *chatMemberRepo) AddUserToChat(ctx context.Context, chatID, userID int64, role string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO chat_members (chat_id, user_id, role) VALUES ($1, $2, $3)`,
		chatID, userID, role)
	return err
}

func (r *chatMemberRepo) RemoveUserFromChat(ctx context.Context, chatID, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`,
		chatID, userID)
	return err
}

func (r *chatMemberRepo) UpdateUserRole(ctx context.Context, chatID, userID int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chat_members SET role = $1 WHERE chat_id = $2 AND user_id = $3`,
		role, chatID, userID)
	return err
}
