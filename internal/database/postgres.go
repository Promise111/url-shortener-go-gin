package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	var ctx context.Context = context.Background()

	var err error
	var config *pgxpool.Config

	config, err = pgxpool.ParseConfig(databaseURL)
	if err != nil {
		slog.Error("Database connection error ", "message", err.Error())
		return nil, err
	}

	var pool *pgxpool.Pool

	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Unable to create connection pool ", "message", err.Error())
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		slog.Error("Unable to ping database ", "message", err.Error())
		return nil, err
	}

	slog.Info("Database connected successfully!")
	return pool, nil
}
