package repository

import (
	"JustChat/internal/messages/model"
	"context"
)

type MessageRepo interface {
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	DeleteByID(ctx context.Context, id int64) error
}
