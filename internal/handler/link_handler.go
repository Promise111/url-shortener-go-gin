package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/Promise111/url-shortener-go-gin/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExpiresAtInPast = errors.New("expires_at must be in the future")

type CreateLinkRequest struct {
	LongURL   string     `json:"long_url" binding:"required,url"`
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
// @Router /links [post]
func CreateLinkHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var CreateLinkReq CreateLinkRequest
		var err error
		if err = c.ShouldBindJSON(&CreateLinkReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": err.Error(),
			})
			return
		}

		if err = CreateLinkReq.ValidateExpiresAt(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "expires_at must be in the future.",
			})
			return
		}

		var link *model.Link

		shortCode, shortCodeGenErr := util.GenerateShortCode(10)
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
// @Router /links [get]
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
// @Router /links/{id} [get]
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
// @Router /links/{id} [delete]
func DeleteLinkByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var id int64
		var idString = c.Param("id")
		id, err = strconv.ParseInt(idString, 10, 64)
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
