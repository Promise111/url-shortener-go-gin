package repository

import (
	"context"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateLink(pool *pgxpool.Pool, longURL string, shortCode string, expiresAt *time.Time) (*model.Link, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
	INSERT INTO link (long_url, short_code, expires_at) 
	VALUES ($1, $2, $3) 
	RETURNING id, long_url, short_code, expires_at, clicks, created_at, updated_at;
	`
	var err error

	var link model.Link
	err = pool.QueryRow(ctx, query, longURL, shortCode, expiresAt).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.CreatedAt,
		&link.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}
