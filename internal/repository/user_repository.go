package repository

import (
	"context"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserByID(c context.Context, pool *pgxpool.Pool, id string) (*model.Users, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var query = `
	SELECT id, email, password_hash, username, deleted_at, created_at, updated_at WHERE id = $1
	`

	var user model.Users
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

func GetUserByEmail(c context.Context, pool *pgxpool.Pool, email string) (*model.Users, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var user model.Users

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

func GetUserByUsername(c context.Context, pool *pgxpool.Pool, username string) (*model.Users, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var user model.Users

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

func CreateUser(c context.Context, pool *pgxpool.Pool, email string, password_hash string, username string) (*model.Users, error) {
	ctx, cancel := CtxTimeout(c)
	defer cancel()

	var err error
	var user model.Users

	var query string = `
	INSERT INTO users ('email','password_hash','username') 
	VALUES ($1, $2, $3) 
	RETURNING id, email, password_hash, deleted_at, created_at, updated_at;
	`

	err = pool.QueryRow(ctx, query, email, password_hash, username).Scan(
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
