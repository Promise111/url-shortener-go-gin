package model

import "time"

type Link struct {
	ID        string     `json:"id" db:"id"`
	LongURL   string     `json:"long_url" db:"long_url"`
	ShortCode string     `json:"short_code" db:"short_code"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	Clicks    int64        `json:"clicks" db:"clicks"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
