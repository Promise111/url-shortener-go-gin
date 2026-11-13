package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Checks struct {
	Postgres string `json:"postgres" example:"ok"`
	Redis    string `json:"redis" example:"ok"`
}

type HealthResponse struct {
	Status        bool   `json:"status" example:"true"`
	ServiceStatus string `json:"service_status" example:"ok"`
	Checks        Checks `json:"checks"`
	Message       string `json:"message" example:"💪 Shortener API is up and running!"`
}

// HealthHandler reports the status of API, Postgres and Redis. 
//
//	@Summary		Health check
//	@Description	Returns the health status of the API and services it depends on.
//	@Tags			health
//	@ID				healthCheck
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Failure		503 {object}	HealthResponse
//	@Router			/api/v1/health [get]
func HealthHandler(pool *pgxpool.Pool, rds *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancelFunc := context.WithTimeout(c.Request.Context(), time.Second*5)
		defer cancelFunc()

		pgErr := pool.Ping(ctx)
		redisErr := rds.Ping(ctx).Err()

		checks := Checks{Postgres: "ok", Redis: "ok"}

		if pgErr != nil {
			checks.Postgres = "error"
		}

		if redisErr != nil {
			checks.Redis = "error"
		}

		if pgErr != nil || redisErr != nil {
			c.JSON(http.StatusServiceUnavailable, HealthResponse{
				Status:        false,
				ServiceStatus: "unavailable",
				Checks: checks,
				Message: "Service unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, HealthResponse{
			Status:        true,
			ServiceStatus: "ok",
			Checks: checks,
			Message: "💪 Shortener API is up and running!",
		})
	}
}
