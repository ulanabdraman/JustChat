package usecase

import (
	"JustChat/internal/events/model"
	"JustChat/internal/events/repository/postgres"
	"context"
	"time"
)

type EventUseCase interface {
	RecordNameChangedEvent(ctx context.Context, chatID int64, oldName string, newName string, changedByID int64) error
	RecordUserAddedEvent(ctx context.Context, chatID int64, userID int64, addedByID int64) error
	RecordUserRemovedEvent(ctx context.Context, chatID int64, userID int64, removedByID int64) error
}
type eventUseCase struct {
	repo repository.EventRepo
}

func NewEventUseCase(repo repository.EventRepo) EventUseCase {
	return &eventUseCase{repo: repo}
}

// Запись события изменения имени чата
func (uc *eventUseCase) RecordNameChangedEvent(ctx context.Context, chatID int64, oldName string, newName string, changedByID int64) error {
	event := &model.EventNameChanged{
		Type:        "NameChanged",
		CreatorID:   0,
		ChatID:      chatID,
		OldName:     oldName,
		NewName:     newName,
		ChangedByID: changedByID,
		ChangedAt:   time.Now(),
	}

	return uc.repo.SaveEvent(ctx, event)
}

// Запись события добавления пользователя в чат
func (uc *eventUseCase) RecordUserAddedEvent(ctx context.Context, chatID int64, userID int64, addedByID int64) error {
	event := &model.EventUserAdded{
		Type:      "UserAdded",
		CreatorID: 0,
		ChatID:    chatID,
		UserID:    userID,
		AddedByID: addedByID,
		AddedAt:   time.Now(),
	}

	return uc.repo.SaveEvent(ctx, event)
}

// Запись события удаления пользователя из чата
func (uc *eventUseCase) RecordUserRemovedEvent(ctx context.Context, chatID int64, userID int64, removedByID int64) error {
	event := &model.EventUserRemoved{
		Type:        "UserRemoved",
		CreatorID:   0,
		ChatID:      chatID,
		UserID:      userID,
		RemovedByID: removedByID,
		RemovedAt:   time.Now(),
	}

	return uc.repo.SaveEvent(ctx, event)
}
