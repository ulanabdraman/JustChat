package model

import "time"

type EventNameChanged struct {
	Type        string    `json:"type"`
	ChatID      int64     `json:"chat_id"`
	NewName     string    `json:"new_name"`
	OldName     string    `json:"old_name"`
	ChangedByID int64     `json:"changed_by_id"`
	ChangedAt   time.Time `json:"changed_at"`
}
type EventUserAdded struct {
	Type      string    `json:"type"`
	ChatID    int64     `json:"chat_id"`
	UserID    int64     `json:"user_id"`
	AddedByID int64     `json:"added_by_id"`
	AddedAt   time.Time `json:"added_at"`
}
type EventUserRemoved struct {
	Type        string    `json:"type"`
	ChatID      int64     `json:"chat_id"`
	UserID      int64     `json:"user_id"`
	RemovedByID int64     `json:"removed_by_id"`
	RemovedAt   time.Time `json:"removed_at"`
}
