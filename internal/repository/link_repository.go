package repository

import (
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateLink(c *gin.Context, pool *pgxpool.Pool, longURL string, shortCode string, expiresAt *time.Time) (*model.Link, error) {
	ctx, cancel := CtxTimeout(c)
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

func GetLinkByID(c *gin.Context, pool *pgxpool.Pool, id int64) (*model.Link, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, created_at, updated_at 
	FROM links 
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

func GetLinksTotalCount(c *gin.Context, pool *pgxpool.Pool) (int64, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	SELECT COUNT(*) FROM links 
	WHERE expires_at IS NULL OR expires_at > NOW();
	`
	var total int64
	var err = pool.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return total, err
	}

	return total, nil
}

func GetLinks(c *gin.Context, pool *pgxpool.Pool, page int, limit int) ([]model.Link, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var offset = (page - 1) * limit
	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, created_at, updated_at
	FROM links 
	WHERE expires_at IS NULL OR expires_at > NOW()
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

func UpdateLinks(c *gin.Context, pool *pgxpool.Pool, longURL string, expiresAt *time.Time, id int64) (*model.Link, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	UPDATE links 
	SET long_url = $1, expires_at = $2, updated_at = NOW()
	WHERE id = $3 
	RETURNING id, long_url, short_code, expires_at, clicks, created_at, updated_at;
	`
	var link model.Link

	var err error = pool.QueryRow(ctx, query, longURL, expiresAt, id).Scan(
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

func DeleteLink(c *gin.Context, pool *pgxpool.Pool, id int64) error {
	ctx, cancel := CtxTimeout(c)
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
		return pgx.ErrNoRows
	}

	return nil
}

func GetLinkByShortCode(c *gin.Context, pool *pgxpool.Pool, shortCode string) (*model.Link, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var link model.Link
	var query string = `
	SELECT id, long_url, short_code, expires_at, clicks, created_at, updated_at 
	FROM links 
	WHERE short_code = $1 
	AND (expires_at IS NULL OR expires_at > NOW());
	`
	err = pool.QueryRow(ctx, query, shortCode).Scan(
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

func IncrementClickCount(c *gin.Context, pool *pgxpool.Pool, shortCode string) error {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query string = `
	UPDATE links 
	SET clicks = clicks + 1 
	WHERE short_code = $1 
	AND (expires_at IS NULL OR expires_at > NOW());
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
