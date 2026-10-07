package router

import (
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/handler"
	"github.com/Promise111/url-shortener-go-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

const (
	APIPrefix    = "/api/v1"
	HealthPrefix = "/health"
	AuthPrefix   = "/auth"
	LinkPrefix   = "/links"
)

func Router(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	var r = gin.Default()

	createLimiter := middleware.NewIPLimiter(rate.Every(6*time.Second), 3)
	redirectLimiter := middleware.NewIPLimiter(5, 10)

	api := r.Group(APIPrefix)

	{
		api.GET(HealthPrefix, handler.HealthHandler)
	}

	{
		link := api.Group(LinkPrefix)
		link.POST("", createLimiter.RateLimiterMiddleware(), handler.CreateLinkHandler(pool))
		link.GET("", handler.GetLinksHandler(pool))
		link.GET("/:id", handler.GetLinkByIDHandler(pool))
		link.DELETE("/:id", handler.DeleteLinkByIDHandler(pool))
		link.PATCH("/:id", handler.UpdateLinksByIdHandler(pool))
	}

	{
		auth := api.Group(AuthPrefix)
		auth.POST("/register", handler.RegisterUserHandler(pool))
	}

	// public
	r.GET("/:shortCode", redirectLimiter.RateLimiterMiddleware(), handler.RedirectShortCodeHandler(pool))

	return r
}
