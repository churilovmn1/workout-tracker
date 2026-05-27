package models

import "time"

// Role represents a user's authorization level.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a registered user.
type User struct {
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         Role      `json:"role" db:"role"`
	TelegramID   *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
