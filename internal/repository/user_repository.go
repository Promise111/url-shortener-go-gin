package repository

import (
	"context"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserByID(c context.Context, pool *pgxpool.Pool, id string) (*model.User, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query = `
	SELECT id, email, password_hash, username, deleted_at, created_at, updated_at WHERE id = $1
	`

	var user model.User
	var err error

	err = pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(c context.Context, pool *pgxpool.Pool, email string) (*model.User, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var user model.User

	var query string = `SELECT id, email, password_hash, username, deleted_at, created_at, updated_at FROM user WHERE email = $1`

	err = pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByUsername(c context.Context, pool *pgxpool.Pool, username string) (*model.User, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var user model.User

	var query string = `SELECT id, email, password_hash, username, deleted_at, created_at, updated_at FROM user WHERE username = $1`

	err = pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

