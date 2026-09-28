package repository

import (
	"context"
	"time"
)

const Duration = 5

func CtxTimeout(c context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c, Duration*time.Second)
}
