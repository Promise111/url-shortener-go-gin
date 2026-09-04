package router

import (
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/handler"
	"github.com/gin-gonic/gin"
)

const (
	APIPrefix    = "/api/v1"
	HealthPrefix = "/health"
	AuthPrefix   = "/auth"
	LinkPrefix   = "/link"
)

func Router(cfg *config.Config) *gin.Engine {
	var r = gin.Default()

	api := r.Group(APIPrefix)

	{
		api.GET(HealthPrefix, handler.HealthHandler)
	}

	return r
}
