package model

import "time"

type Message struct {
	ID        int64     `json:"id" db:"id"`
	ChatID    int64     `json:"chat_id" db:"chat_id"`
	CreatorID int64     `json:"creator_id" db:"creator_id"`
	Text      string    `json:"text" db:"text"`
	SentAt    time.Time `json:"sent_at" db:"sent_at"`
	Deleted   bool      `json:"deleted" db:"deleted"`
}
