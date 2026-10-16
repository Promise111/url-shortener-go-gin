package router

import (
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/handler"
	"github.com/Promise111/url-shortener-go-gin/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	APIPrefix    = "/api/v1"
	HealthPrefix = "/health"
	AuthPrefix   = "/auth"
	LinkPrefix   = "/links"
)

func Router(pool *pgxpool.Pool, cfg *config.Config, rds *redis.Client, asynqClient *asynq.Client) *gin.Engine {
	var r = gin.Default()

	api := r.Group(APIPrefix)

	var redirectLimiter = middleware.NewFixLimiter(6, 1*time.Second, middleware.Redirect)
	var createLimiter = middleware.NewFixLimiter(10, 1*time.Minute, middleware.Create)

	{
		api.GET(HealthPrefix, handler.HealthHandler)
	}

	{
		link := api.Group(LinkPrefix, middleware.AuthMiddleware(cfg))
		link.POST("", createLimiter.FixWindowLimiterMiddleware(rds), handler.CreateLinkHandler(pool))
		link.GET("", handler.GetLinksHandler(pool))
		link.GET("/code/:shortCode", handler.GetLinkByShortHandler(pool))
		link.GET("/:id", handler.GetLinkByIDHandler(pool))
		link.DELETE("/:id", handler.DeleteLinkByIDHandler(pool, rds))
		link.PATCH("/:id", handler.UpdateLinksByIdHandler(pool, rds))
	}

	{
		auth := api.Group(AuthPrefix)
		auth.POST("/register", handler.RegisterUserHandler(pool))
		auth.POST("/login", handler.LoginHandler(pool, cfg))
	}

	// public
	r.GET("/:shortCode", redirectLimiter.FixWindowLimiterMiddleware(rds), handler.RedirectShortCodeHandler(pool, rds, asynqClient))

	return r
}
