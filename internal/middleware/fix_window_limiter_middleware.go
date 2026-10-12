package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Prefix string

const (
	Create   Prefix = "create"
	Redirect Prefix = "redirect"
)

type Limiter struct {
	Window time.Duration
	Max    int
	Prefix Prefix
}

func NewFixLimiter(max int, window time.Duration, prefix Prefix) *Limiter {

	return &Limiter{
		Max:    max,
		Window: window,
		Prefix: prefix,
	}
}

func (l *Limiter) FixWindowLimiterMiddleware(rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ip = c.ClientIP()
		var key = "rl:" + string(l.Prefix) + ":" + ip
		var ctx context.Context = c.Request.Context()
		count, err := rds.Incr(ctx, key).Result()
		if err != nil {
			slog.Error("FixWindowLimiter ", "err ", err.Error())
			c.Next()
			return
		}
		if count == 1 {
			expireErr := rds.Expire(ctx, key, l.Window).Err()
			if expireErr != nil {
				slog.Error("FixWindowLimiter ", "err ", expireErr.Error())
			}
		}
		if count > int64(l.Max) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"message": "Too many requests.",
			})
			return
		}

		c.Next()
	}
}
