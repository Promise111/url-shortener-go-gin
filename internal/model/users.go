package model

import "time"

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"_" db:"password_hash"`
	Username     string    `json:"username" db:"username"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
