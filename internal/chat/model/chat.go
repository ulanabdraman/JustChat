package model

import "time"

type Chat struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
}
