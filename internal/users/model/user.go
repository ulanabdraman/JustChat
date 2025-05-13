package model

import "time"

type User struct {
	ID        int64     `json:"id" db:"id"`
	Type      string    `json:"type" db:"type"`
	Username  string    `json:"username" db:"username"`
	Online    bool      `json:"online" db:"online"`
	WasOnline time.Time `json:"was_online" db:"was_online"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Password  string    `json:"password" db:"password"`
}
