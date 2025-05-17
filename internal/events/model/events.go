package model

import "time"

type EventNameChanged struct {
	ID          int64     `json:"id" bson:"_id"`                // уникальный id (общий с Message)
	Type        string    `json:"type" bson:"type"`             // например, "event:name_changed"
	ChatID      int64     `json:"chat_id" bson:"chat_id"`       // id чата
	CreatorID   int64     `json:"creator_id" bson:"creator_id"` // кто инициировал
	NewName     string    `json:"new_name" bson:"new_name"`
	OldName     string    `json:"old_name" bson:"old_name"`
	ChangedByID int64     `json:"changed_by_id" bson:"changed_by_id"`
	ChangedAt   time.Time `json:"changed_at" bson:"changed_at"`
}

type EventUserAdded struct {
	ID        int64     `json:"id" bson:"_id"`
	Type      string    `json:"type" bson:"type"`
	ChatID    int64     `json:"chat_id" bson:"chat_id"`
	CreatorID int64     `json:"creator_id" bson:"creator_id"`
	UserID    int64     `json:"user_id" bson:"user_id"`
	AddedByID int64     `json:"added_by_id" bson:"added_by_id"`
	AddedAt   time.Time `json:"added_at" bson:"added_at"`
}

type EventUserRemoved struct {
	ID          int64     `json:"id" bson:"_id"`
	Type        string    `json:"type" bson:"type"`
	ChatID      int64     `json:"chat_id" bson:"chat_id"`
	CreatorID   int64     `json:"creator_id" bson:"creator_id"`
	UserID      int64     `json:"user_id" bson:"user_id"`
	RemovedByID int64     `json:"removed_by_id" bson:"removed_by_id"`
	RemovedAt   time.Time `json:"removed_at" bson:"removed_at"`
}
