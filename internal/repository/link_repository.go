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
	INSERT INTO links (long_url, short_code, expires_at) 
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
		&link.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func GetLinkByID(pool *pgxpool.Pool, id int64) (*model.Link, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
	SELECT * FROM links 
	WHERE id = $1;
	`

	var link model.Link

	var err error = pool.QueryRow(ctx, query, id).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func GetTotalLinksCount(pool *pgxpool.Pool) (*int64, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
	SELECT COUNT(*) FROM links;
	`
	var total int64
	var err = pool.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return &total, err
	}

	return &total, nil
}

func GetLinks(pool *pgxpool.Pool, page int, limit int) ([]model.Link, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var offset = (page - 1) * limit
	var query string = `
	SELECT * 
	FROM links 
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
	`

	rows, err := pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.Link = []model.Link{}
	for rows.Next() {
		var link model.Link
		err := rows.Scan(
			&link.ID,
			&link.LongURL,
			&link.ShortCode,
			&link.ExpiresAt,
			&link.Clicks,
			&link.CreatedAt,
			&link.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func UpdateLinks(pool *pgxpool.Pool, longURL string, shortCode string, expiredAt *time.Time, id string) (*model.Link, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
	UPDATE links 
	SET long_url = $1, short_code = $2, expires_at = $3 
	WHERE id = $4 
	RETURNING id, long_url, short_code, expires_at, created_at, updated_at;
	`
	var link model.Link

	var err error = pool.QueryRow(ctx, query, longURL, shortCode, expiredAt, id).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func DeleteLink(pool *pgxpool.Pool, id int64) error {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
	DELETE FROM links 
	WHERE ID = $1
	`

	cmdTag, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return nil
	}

	return nil
}
