package postgres

import (
	"JustChat/internal/messages/model"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type MessageRepo interface {
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	DeleteByID(ctx context.Context, id int64) error
}
type messageRepo struct {
	db *sqlx.DB
}

func NewMessageRepo(db *sqlx.DB) MessageRepo {
	return &messageRepo{db: db}
}
func (m *messageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	query := `SELECT id,chat_id,creator_id,text,sent_at FROM messages WHERE id=$1 AND deleted=false`
	row := m.db.QueryRowContext(ctx, query, id)
	var message model.Message
	err := row.Scan(&message.ID, &message.ChatID, &message.CreatorID, &message.Text, &message.SentAt)
	if err != nil {
		return nil, fmt.Errorf("getByID message repo: %w", err)
	}
	return &message, nil
}
func (m *messageRepo) DeleteByID(ctx context.Context, id int64) error {
	query := `UPDATE messages SET deleted=true WHERE id=$1 AND deleted=false`
	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleteByID message repo: %w", err)
	}
	return nil
}
func (m *messageRepo) GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error) {
	query := `SELECT id, chat_id, creator_id, text, sent_at FROM messages WHERE chat_id=$1 AND deleted=false`
	rows, err := m.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("getByChatID message repo: %w", err)
	}
	defer rows.Close()
	var messages []model.Message
	for rows.Next() {
		var message model.Message

		err := rows.Scan(&message.ID, &message.ChatID, &message.CreatorID, &message.Text, &message.SentAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return messages, nil
}
func (m *messageRepo) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	query := `
		INSERT INTO messages(chat_id, creator_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, chat_id, creator_id, text, sent_at
	`

	var saved model.Message
	err := m.db.QueryRowContext(ctx, query, message.ChatID, message.CreatorID, message.Text).
		Scan(&saved.ID, &saved.ChatID, &saved.CreatorID, &saved.Text, &saved.SentAt)
	if err != nil {
		return nil, fmt.Errorf("saveMessage message repo: %w", err)
	}

	return &saved, nil
}
