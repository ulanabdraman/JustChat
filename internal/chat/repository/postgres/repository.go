package postgres

import (
	"JustChat/internal/chat/model"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type ChatRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Chat, error) //escape anal вот тут проявляется есче,выбор либо за памятью или производительностью
	Create(ctx context.Context, chat *model.Chat) error
	Update(ctx context.Context, chat *model.Chat) error
	Delete(ctx context.Context, chatID int64) error
	AddUserToChat(ctx context.Context, chatID int64, userID int64) error
	RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error
}
type chatRepository struct {
	db *sqlx.DB
}

func NewChatRepo(db *sqlx.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) GetByID(ctx context.Context, id int64) (*model.Chat, error) {
	query := `SELECT id, name, type,created_by, created_at FROM chats WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var chat model.Chat
	if err := row.Scan(&chat.ID, &chat.Name, &chat.Type, &chat.CreatedBy, &chat.CreatedAt); err != nil {
		return nil, fmt.Errorf("getByID chat repo: %w", err)
	}
	return &chat, nil
}
func (r *chatRepository) Create(ctx context.Context, chat *model.Chat) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Вставляем чат и получаем его ID
	query := `INSERT INTO chats(name, type, created_by) VALUES ($1, $2, $3) RETURNING id`
	var chatID int64
	err = tx.QueryRowContext(ctx, query, chat.Name, chat.Type, chat.CreatedBy).Scan(&chatID)
	if err != nil {
		return fmt.Errorf("insert chat: %w", err)
	}

	// Добавляем создателя в chatmembers
	memberQuery := `INSERT INTO chat_members(chat_id, user_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, memberQuery, chatID, chat.CreatedBy)
	if err != nil {
		return fmt.Errorf("insert chatmember: %w", err)
	}

	return nil
}

func (r *chatRepository) Update(ctx context.Context, chat *model.Chat) error {
	query := `UPDATE chats SET name = $1, type = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, chat.Name, chat.Type, chat.ID)
	if err != nil {
		return fmt.Errorf("update chat repo: %w", err)
	}
	return nil
}

func (r *chatRepository) Delete(ctx context.Context, chatID int64) error {
	query := `DELETE FROM chats WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("delete chat repo: %w", err)
	}
	return nil
}

func (r *chatRepository) AddUserToChat(ctx context.Context, chatID int64, userID int64) error {
	query := `INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("addUserToChat chat repo: %w", err)
	}
	return nil
}

func (r *chatRepository) RemoveUserFromChat(ctx context.Context, chatID int64, userID int64) error {
	query := `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("removeUserFromChat chat repo: %w", err)
	}
	return nil
}
