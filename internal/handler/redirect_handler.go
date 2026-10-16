package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/queue"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// @Summary Redirect by short code
// @Description Redirect the client to the original long URL
// @Tags links
// @Produce json
// @Param shortCode path string true "Short code"
// @Success 307 {string} string "Temporary redirect to the long URL"
// @Header 307 {string} Location "Destination long URL"
// @Failure 404 {object} map[string]interface{}
// @Failure 410 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /{shortCode} [get]
func RedirectShortCodeHandler(pool *pgxpool.Pool, rds *redis.Client, asynqClient *asynq.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var shortCode string = c.Param("shortCode")
		var err error
		var link *model.Links
		var now time.Time = time.Now().UTC()
		var cacheKey string = "link:" + shortCode

		strCmd := rds.Get(c.Request.Context(), cacheKey)
		cacheVal, cacheErr := strCmd.Result()
		if cacheErr == nil {
			var cached model.Links
			unmarshalError := json.Unmarshal([]byte(cacheVal), &cached)
			if unmarshalError != nil {
				slog.Error("RedirectStrCmd - failed to get cache key " + unmarshalError.Error())
			} else {
				link = &cached
			}
		} else if !errors.Is(cacheErr, redis.Nil) {
			slog.Error("Redirect get key " + cacheErr.Error())
		}

		if link == nil {
			link, err = repository.GetLinkByShortCodeVisitor(c.Request.Context(), pool, shortCode)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					WriteError(c, http.StatusNotFound, "Link record not found.")
					return
				}
				WriteError(c, http.StatusInternalServerError, "Something went wrong!")
				return
			}
			var cacheString, marshalErr = json.Marshal(model.Links{
				LongURL:   link.LongURL,
				ExpiresAt: link.ExpiresAt,
				MaxClicks: link.MaxClicks,
				Status:    link.Status,
			})
			if marshalErr != nil {
				slog.Error("Redirect Short Code " + marshalErr.Error())
			} else {
				setErr := rds.Set(c.Request.Context(), cacheKey, cacheString, 1*time.Hour).Err()
				if setErr != nil {
					slog.Error("Redirect short code, failed to set key " + setErr.Error())
				}
			}
		}

		if link.Status == model.StatusDisabled {
			WriteError(c, http.StatusGone, "This link has been disabled.")
			return
		}
		if (link.ExpiresAt != nil && !link.ExpiresAt.After(now)) || (link.Status == model.StatusExpired || (link.MaxClicks != nil && *link.MaxClicks <= link.Clicks)) {
			WriteError(c, http.StatusGone, "This link has expired.")
			return
		}

		var task, taskErr = queue.IncrementLinkTask(shortCode)
		if taskErr != nil {
			slog.Error("RedirectShortCodeHandler - enqueue build", "err ", taskErr)
		} else if _, taskEnqueueErr := asynqClient.Enqueue(task); taskEnqueueErr != nil {
			slog.Error("RedirectShortCodeHandler - enqueue", "err ", taskEnqueueErr)
		}

		c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
	}
}
