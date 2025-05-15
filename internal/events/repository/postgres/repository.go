package repository

import (
	"JustChat/internal/events/model"
	"context"
	"database/sql"
	"errors"
)

type EventRepo interface {
	SaveEvent(ctx context.Context, event interface{}) error
}

type eventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) EventRepo {
	return &eventRepo{db: db}
}

func (repo *eventRepo) SaveEvent(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case *model.EventNameChanged:
		_, err := repo.db.ExecContext(
			ctx,
			"INSERT INTO events (type, chat_id, old_name, new_name, changed_by_id, changed_at) VALUES ($1, $2, $3, $4, $5, $6)",
			e.Type, e.ChatID, e.OldName, e.NewName, e.ChangedByID, e.ChangedAt,
		)
		return err

	case *model.EventUserAdded:
		_, err := repo.db.ExecContext(
			ctx,
			"INSERT INTO events (type, chat_id, user_id, added_by_id, added_at) VALUES ($1, $2, $3, $4, $5)",
			e.Type, e.ChatID, e.UserID, e.AddedByID, e.AddedAt,
		)
		return err

	case *model.EventUserRemoved:
		_, err := repo.db.ExecContext(
			ctx,
			"INSERT INTO events (type, chat_id, user_id, removed_by_id, removed_at) VALUES ($1, $2, $3, $4, $5)",
			e.Type, e.ChatID, e.UserID, e.RemovedByID, e.RemovedAt,
		)
		return err

	default:
		return errors.New("unknown event type")
	}
}
