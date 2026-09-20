package router

import (
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	APIPrefix    = "/api/v1"
	HealthPrefix = "/health"
	AuthPrefix   = "/auth"
	LinkPrefix   = "/links"
)

func Router(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	var r = gin.Default()

	api := r.Group(APIPrefix)

	{
		api.GET(HealthPrefix, handler.HealthHandler)
	}

	{
		link := api.Group(LinkPrefix)
		link.POST("", handler.CreateLinkHandler(pool))
		link.GET("", handler.GetLinksHandler(pool))
		link.GET("/:id", handler.GetLinkByIDHandler(pool))
		link.DELETE("/:id", handler.DeleteLinkByIDHandler(pool))
		link.PATCH("/:id", handler.UpdateLinksByIdHandler(pool))
	}

	// public
	r.GET("/:shortCode", handler.RedirectShortCodeHandler(pool))

	return r
}
