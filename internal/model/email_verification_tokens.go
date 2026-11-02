package model

import "time"

type EmailVerificationTokens struct {
	ID        string    `json:"id" db:"id"`
	TokenHash string    `json:"-" db:"token_hash"`
	UserID    string    `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}