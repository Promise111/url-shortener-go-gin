package handler

import (
	"errors"
	"net/http"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Redirect by short code
// @Description Redirect the client to the original long URL
// @Tags links
// @Produce json
// @Param shortCode path string true "Short code"
// @Success 307 {string} string "Temporary redirect to the long URL"
// @Header 307 {string} Location "Destination long URL"
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /{shortCode} [get]
func RedirectShortCodeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var shortCode string = c.Param("shortCode")
		var err error
		var link *model.Link
		link, err = repository.GetLinkByShortCode(pool, shortCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link record not found.")
				return
			}
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}
		err = repository.IncrementClickCount(pool, shortCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link record not found.")
				return
			}
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, link.LongURL)
	}
}
