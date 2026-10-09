package repository

import (
	"context"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateLink(c context.Context, pool *pgxpool.Pool, longURL string, shortCode string, expiresAt *time.Time, status model.Status, maxClicks *int64, userID string) (*model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	INSERT INTO links (long_url, short_code, expires_at, status, max_clicks, user_id) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at;
	`
	var err error

	var link model.Links
	err = pool.QueryRow(ctx, query, longURL, shortCode, expiresAt, status, maxClicks, userID).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.Status,
		&link.MaxClicks,
		&link.UserID,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func GetLinkByID(c context.Context, pool *pgxpool.Pool, id int64, userID string) (*model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at 
	FROM links 
	WHERE id = $1 AND user_id = $2;
	`

	var link model.Links

	var err error = pool.QueryRow(ctx, query, id, userID).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.Status,
		&link.MaxClicks,
		&link.UserID,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func GetLinkByShortCodeVisitor(c context.Context, pool *pgxpool.Pool, shortCode string) (*model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var link model.Links
	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at 
	FROM links 
	WHERE short_code = $1 
	`
	err = pool.QueryRow(ctx, query, shortCode).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.Status,
		&link.MaxClicks,
		&link.UserID,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func GetLinksTotalCount(c context.Context, pool *pgxpool.Pool, userID string) (int64, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	SELECT COUNT(*) FROM links 
	WHERE user_id = $1;
	`
	var total int64
	var err = pool.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return total, err
	}

	return total, nil
}

func GetLinks(c context.Context, pool *pgxpool.Pool, page int, limit int, userID string) ([]model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var offset = (page - 1) * limit
	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at
	FROM links 
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.Links = []model.Links{}
	for rows.Next() {
		var link model.Links
		err := rows.Scan(
			&link.ID,
			&link.LongURL,
			&link.ShortCode,
			&link.ExpiresAt,
			&link.Clicks,
			&link.Status,
			&link.MaxClicks,
			&link.UserID,
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

func UpdateLinks(c context.Context, pool *pgxpool.Pool, longURL string, expiresAt *time.Time, id int64, status model.Status, maxClicks *int64, userID string) (*model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	UPDATE links 
	SET long_url = $1, expires_at = $2, status = $4, max_clicks = $5, updated_at = NOW()
	WHERE id = $3 
	AND user_id = $6 
	RETURNING id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at;
	`
	var link model.Links

	var err error = pool.QueryRow(ctx, query, longURL, expiresAt, id, status, maxClicks, userID).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.Status,
		&link.MaxClicks,
		&link.UserID,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func DeleteLink(c context.Context, pool *pgxpool.Pool, id int64, userID string) error {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	DELETE FROM links 
	WHERE id = $1 
	AND user_id = $2
	`

	cmdTag, err := pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func GetLinkByShortCode(c context.Context, pool *pgxpool.Pool, shortCode string, userID string) (*model.Links, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var link model.Links
	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, status, max_clicks, user_id, created_at, updated_at 
	FROM links 
	WHERE short_code = $1 
	AND user_id = $2;
	`
	err = pool.QueryRow(ctx, query, shortCode, userID).Scan(
		&link.ID,
		&link.LongURL,
		&link.ShortCode,
		&link.ExpiresAt,
		&link.Clicks,
		&link.Status,
		&link.MaxClicks,
		&link.UserID,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func IncrementClickCount(c context.Context, pool *pgxpool.Pool, shortCode string) error {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	UPDATE links 
	SET clicks = clicks + 1 
	WHERE short_code = $1 
	AND status = 'active' 
	AND (expires_at IS NULL OR expires_at > NOW()) 
	AND (max_clicks IS NULL OR clicks < max_clicks)
	`

	var cmdTag, err = pool.Exec(ctx, query, shortCode)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
