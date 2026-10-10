package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr string) (*redis.Client, error) {
	var ctx context.Context = context.Background()

	var rdb *redis.Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}
