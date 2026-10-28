package model

import "time"

type RefreshTokens struct {
	ID        int64 `json:"id" db:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	TokenHash string `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}