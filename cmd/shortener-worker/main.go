package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/queue"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		slog.Error("shortener-work", "configError ", "failed to load environment variables")
		os.Exit(1)
	}

	var pool *pgxpool.Pool
	var poolErr error
	pool, poolErr = database.Connect(cfg.DatabaseURL)
	if poolErr != nil {
		slog.Error("shortener-work", "poolError ", "failed to connect to database")
		os.Exit(1)
	}
	defer pool.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisADDR},
		asynq.Config{Concurrency: 5},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskIncrement, func(ctx context.Context, task *asynq.Task) error {
		var p queue.IncrementPayload
		var unmarshalError error
		unmarshalError = json.Unmarshal(task.Payload(), &p)
		if unmarshalError != nil {
			return unmarshalError
		}

		slog.Info("Incrementing")
		if incrementLinkClickCountError := repository.IncrementClickCount(ctx, pool, p.ShortCode); incrementLinkClickCountError != nil {
			if errors.Is(incrementLinkClickCountError, pgx.ErrNoRows) {
				slog.Info("increment skipped (gone or at cap) " + p.ShortCode)
				return nil
			}
			return incrementLinkClickCountError
		}
		return nil
	})

	if err := srv.Run(mux); err != nil {
		os.Exit(1)
	}
}
