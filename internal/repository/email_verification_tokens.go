package repository

import (
	"context"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateEmailVerificationToken(c context.Context, pool *pgxpool.Pool, tokenHash string, userID string) error {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	expiresAt := time.Now().Add(time.Hour * 24)

	var err error
	var query = `
	INSERT INTO email_verification_tokens (token_hash, user_id, expires_at) VALUES($1,$2,$3) RETURNING id, token_hash, user_id, expires_at, created_at;
	`

	_, err = pool.Exec(ctx, query, tokenHash, userID, expiresAt)

	if err != nil {
		return err
	}

	return nil
}

func GetEmailVerificationToken(c context.Context, pool *pgxpool.Pool, tokenHash string) (*model.EmailVerificationTokens, error) {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var tokenRecord model.EmailVerificationTokens
	var query = `
	SELECT id, token_hash, user_id, expires_at, created_at FROM email_verification_tokens WHERE token_hash = $1 AND expires_at > NOW();
	`

	err = pool.QueryRow(ctx, query, tokenHash).Scan(
		&tokenRecord.ID,
		&tokenRecord.TokenHash,
		&tokenRecord.UserID,
		&tokenRecord.ExpiresAt,
		&tokenRecord.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tokenRecord, nil
}

func DeleteExpiredEmailVerificationToken(c context.Context, pool *pgxpool.Pool) error {
	ctx, cancelFunc := CtxTimeout(c)
	defer cancelFunc()

	var err error
	var query = `
	DELETE FROM email_verification_tokens 
	WHERE expires_at <= NOW();
	`

	_, err = pool.Exec(ctx, query)

	if err != nil {
		return err
	}

	return nil
}
