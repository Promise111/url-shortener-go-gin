package repository

import (
	"context"
	"time"
)

const Duration = 5

func CtxTimeout() (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), Duration*time.Second)

	return ctx, cancel
}
