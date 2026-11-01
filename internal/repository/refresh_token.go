package repository

import (
	"context"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PersistRefreshToken(c context.Context, pool *pgxpool.Pool, refreshTokenHash string, userId string, expiresAt time.Time) (*model.RefreshTokens, error) {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var refreshRecord model.RefreshTokens
	var query = `
	INSERT INTO 
	refresh_tokens (token_hash, user_id, expires_at) 
	VALUES ($1, $2, $3) 
	RETURNING id, user_id, token_hash, expires_at, revoked_at, created_At;
	`
	err = pool.QueryRow(ctx, query, refreshTokenHash, userId, expiresAt).Scan(
		&refreshRecord.ID,
		&refreshRecord.UserID,
		&refreshRecord.TokenHash,
		&refreshRecord.ExpiresAt,
		&refreshRecord.RevokedAt,
		&refreshRecord.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &refreshRecord, nil

}

func RevokeRefreshToken(c context.Context, pool *pgxpool.Pool, revokedAt time.Time, refreshHash string) (*model.RefreshTokens, error) {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var refreshHashRecord model.RefreshTokens
	var query = `
	UPDATE refresh_tokens 
	SET revoked_at = $1 
	WHERE token_hash = $2 
	RETURNING id, user_id, token_hash, expires_at, revoked_at, created_At;
	`

	err = pool.QueryRow(ctx, query, revokedAt, refreshHash).Scan(
		&refreshHashRecord.ID,
		&refreshHashRecord.UserID,
		&refreshHashRecord.TokenHash,
		&refreshHashRecord.ExpiresAt,
		&refreshHashRecord.RevokedAt,
		&refreshHashRecord.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &refreshHashRecord, nil
}

func GetRefreshTokenByHash(c context.Context, pool *pgxpool.Pool, refreshHash string) (*model.RefreshTokens, error) {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var refreshHashRecord model.RefreshTokens
	var query = `
	SELECT id, user_id, token_hash, expires_at, revoked_at, created_at 
	FROM refresh_tokens 
	WHERE token_hash = $1 
	AND revoked_at IS NULL 
	AND expires_at > NOW()`

	err = pool.QueryRow(ctx, query, refreshHash).Scan(
		&refreshHashRecord.ID,
		&refreshHashRecord.UserID,
		&refreshHashRecord.TokenHash,
		&refreshHashRecord.ExpiresAt,
		&refreshHashRecord.RevokedAt,
		&refreshHashRecord.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &refreshHashRecord, nil
}

func DeleteExpiredRefreshTokens(c context.Context, pool *pgxpool.Pool) error {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var query = `
	DELETE FROM refresh_tokens 
	WHERE expires_at <= NOW();
	`

	_, err = pool.Exec(ctx, query)

	if err != nil {
		return err
	}

	return nil
}
