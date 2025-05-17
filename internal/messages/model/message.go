package model

import "time"

type Message struct {
	ID        int64     `json:"id" db:"id" bson:"_id,omitempty"`
	Type      string    `json:"type" db:"type" bson:"type"`
	ChatID    int64     `json:"chat_id" db:"chat_id" bson:"chat_id"`
	CreatorID int64     `json:"creator_id" db:"creator_id" bson:"creator_id"`
	Text      string    `json:"text" db:"text" bson:"text"`
	SentAt    time.Time `json:"sent_at" db:"sent_at" bson:"sent_at"`
	Deleted   bool      `json:"deleted" db:"deleted" bson:"deleted"`
}
