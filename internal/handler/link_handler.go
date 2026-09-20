package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/Promise111/url-shortener-go-gin/internal/shortcode"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExpiresAtInPast = errors.New("expires_at must be in the future")

type OptionalExpiresAt struct {
	Present bool
	Time    *time.Time
}

func (o *OptionalExpiresAt) UnmarshalJSON(b []byte) error {
	o.Present = true
	if string(b) == "null" {
		o.Time = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// treat "" and " " as null
	if strings.TrimSpace(s) == "" || strings.TrimSpace(s) == " " {
		o.Time = nil
		return nil
	}

	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	o.Time = &t
	return nil
}

type CreateLinkRequest struct {
	LongURL   string     `json:"long_url" binding:"required,url,max=20448"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (r CreateLinkRequest) ValidateExpiresAt() error {
	if r.ExpiresAt == nil {
		return nil
	}
	if !r.ExpiresAt.After(time.Now().UTC()) {
		return ErrExpiresAtInPast
	}
	return nil
}

type UpdateLinkRequest struct {
	LongURL   *string           `json:"long_url" binding:"omitempty,url,max=2048"`
	ExpiresAt OptionalExpiresAt `json:"expires_at" swaggertype:"string" format:"date-time" example:"2027-08-08T10:58:29Z"`
}

func (r UpdateLinkRequest) ValidateExpiresAt() error {
	if !r.ExpiresAt.Present || r.ExpiresAt.Time == nil {
		return nil
	}
	var now = time.Now().UTC()
	if !r.ExpiresAt.Time.After(now) {
		return ErrExpiresAtInPast
	}
	return nil
}

type LinkSample struct {
	ID        int64      `json:"id" example:"1"`
	LongURL   string     `json:"long_url" example:"https://facebook.com"`
	ShortCode string     `json:"short_code" example:"1234567890"`
	ExpiresAt *time.Time `json:"expires_at" example:"2027-04-08T00:00:00Z"`
	Clicks    int64      `json:"clicks" example:"10"`
	CreatedAt time.Time  `json:"created_at" example:"2026-02-02T00:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2026-09-11T00:00:00Z"`
}

type CreateLinkResponse struct {
	Status  bool       `json:"status" example:"true"`
	Message string     `json:"message" example:"Link created successfully!"`
	Data    LinkSample `json:"data"`
}

type UpdateLinkResponse struct {
	Status  bool       `json:"status" example:"true"`
	Message string     `json:"message" example:"Link updated successfully!"`
	Data    LinkSample `json:"data"`
}

type GetLinksResponse struct {
	Status  bool         `json:"status" example:"true"`
	Message string       `json:"message" example:"Link created successfully!"`
	Data    []LinkSample `json:"data"`
}

type GetLinkResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"Records fetched successfully!"`
	Data    LinkSample
}

// @Summary Shorten URL
// @Description Create new shortened URL
// @Tags links
// @Accept json
// @Produce json
// @Param request body CreateLinkRequest true "Create Link payload"
// @Success 201 {object} CreateLinkResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links [post]
func CreateLinkHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var CreateLinkReq CreateLinkRequest
		var err error
		if err = c.ShouldBindJSON(&CreateLinkReq); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		if err = CreateLinkReq.ValidateExpiresAt(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		var link *model.Link

		shortCode, shortCodeGenErr := shortcode.Generate(10)
		if shortCodeGenErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Something went wrong.",
			})
		}

		link, err = repository.CreateLink(pool, CreateLinkReq.LongURL, shortCode, CreateLinkReq.ExpiresAt)

		c.JSON(http.StatusCreated, CreateLinkResponse{
			Status:  true,
			Message: "Link created successfuly!",
			Data: LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				CreatedAt: link.CreatedAt,
				UpdatedAt: link.UpdatedAt,
			},
		})
	}
}

// @Summary Get links
// @Description Fetch all links record
// @Tags links
// @Produce json
// @Param limit query string false "limit pagination records"
// @Param page query string false "specify pagination page"
// @Success 200 {object} []GetLinksResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links [get]
func GetLinksHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var limit int = 10
		var page int = 1
		var links []model.Link
		var err error
		if queryLimit := c.Query("limit"); queryLimit != "" {
			if l, err := strconv.Atoi(queryLimit); err == nil && l > 0 {
				if l > 100 {
					l = 100
				}
				limit = l
			}
		}
		if queryPage := c.Query("page"); queryPage != "" {
			if p, err := strconv.Atoi(queryPage); err == nil && p > 0 {
				page = p
			}
		}
		links, err = repository.GetLinks(pool, page, limit)
		if err != nil {
			slog.Error(err.Error())
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		var total int64
		total, err = repository.GetLinksTotalCount(pool)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}
		var castedLimit = int64(limit)
		var totalPage int64 = 0

		if total == 0 {
			totalPage = 0
		} else {
			totalPage = (total + castedLimit - 1) / castedLimit
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"message":   "Links fetched successfully!",
			"data":      links,
			"totalPage": totalPage,
			"total":     total,
			"page":      page,
			"limit":     limit,
		})
	}
}

// @Summary Get link
// @Description Fetch link by id
// @Tags links
// @Param id path int true "Link ID"
// @Produce json
// @Success 200 {object} GetLinkResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links/{id} [get]
func GetLinkByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		var id int64
		var err error
		id, err = strconv.ParseInt(idParam, 10, 64) // alternative to int64(id)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter valid id parameter")
			return
		}

		var link *model.Link

		link, err = repository.GetLinkByID(pool, id)

		if err != nil {
			slog.Error(err.Error())
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "URL record not found.")
				return
			}
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Records fetched successfully!",
			"data":    link,
		})
	}
}

// @Summary Delete link
// @Description Delete link by id
// @Tags links
// @Param id path int true "Link ID"
// Produce json
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links/{id} [delete]
func DeleteLinkByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var id int64
		var idParam = c.Param("id")
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter a valid id parameter")
			return
		}

		err = repository.DeleteLink(pool, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, "Something wrong!")
			return
		}

		c.Status(http.StatusNoContent)

	}
}

// @Summary Update link
// @Description Update link by id
// @Tags links
// @Param id path int true "Link id"
// @Param request body UpdateLinkRequest true "Fields to update"
// @Accept json
// @Produce json
// @Success 200 {object} UpdateLinkResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links/{id} [patch]
func UpdateLinksByIdHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var id int64
		var err error
		idParam := c.Param("id")
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter valid id param")
			return
		}

		var link *model.Link
		link, err = repository.GetLinkByID(pool, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		var req UpdateLinkRequest
		if err = c.ShouldBindJSON(&req); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}
		if err = req.ValidateExpiresAt(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		if req.LongURL == nil && !req.ExpiresAt.Present {
			WriteError(c, http.StatusBadRequest, "Expected at least one of long_url or expires_at")
			return
		}

		var longURL string = link.LongURL
		var expiresAt *time.Time = link.ExpiresAt
		if req.LongURL != nil {
			longURL = *req.LongURL
		}
		if req.ExpiresAt.Present {
			expiresAt = req.ExpiresAt.Time
		}

		link, err = repository.UpdateLinks(pool, longURL, expiresAt, id)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "Link updated successfully!",
			"data":    link,
		})
	}
}
