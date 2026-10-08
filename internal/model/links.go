package model

import "time"

type Status string

const (
	StatusDisabled Status = "disabled"
	StatusExpired  Status = "expired"
	StatusActive   Status = "active"
)

type Links struct {
	ID        int64      `json:"id" db:"id"`
	LongURL   string     `json:"long_url" db:"long_url"`
	ShortCode string     `json:"short_code" db:"short_code"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	Clicks    int64      `json:"clicks" db:"clicks"`
	Status    Status     `json:"status" db:"status"`
	MaxClicks *int64     `json:"max_clicks" db:"max_clicks"`
	UserID    string     `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
