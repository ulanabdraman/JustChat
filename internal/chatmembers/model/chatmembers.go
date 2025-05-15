package model

type ChatMembers struct {
	ChatID int64  `json:"chat_id" db:"chat_id"`
	UserID int64  `json:"user_id" db:"user_id"`
	Role   string `json:"role" db:"role"`
}
