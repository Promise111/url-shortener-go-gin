package repository

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

const Duration = 5

func CtxTimeout(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), Duration*time.Second)
}
