package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/Promise111/url-shortener-go-gin/internal/shortcode"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExpiresAtInPast = errors.New("expires_at must be in the future")
var InternalServerErrorMessage = "Something went wrong!"
var InvalidStatusMessage = errors.New("status must be either active or disabled")
var MaxClicksCannotBeLess = errors.New("MaxClicks must be 5 or greater")
var UnauthorizedError = errors.New("Unauthorized")

var reserved = map[string]string{
	"api":     "used by the API",
	"health":  "used by health checks",
	"links":   "used by the links API",
	"swagger": "used by API docs",
}

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
	if strings.TrimSpace(s) == "" {
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

type OptionalMaxClicks struct {
	Present bool
	Value   *int64
}

func (m *OptionalMaxClicks) UnmarshalJSON(b []byte) error {
	m.Present = true // UnmarshalJSON is only called because key was present

	if string(b) == "null" {
		m.Value = nil
		return nil
	}

	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	m.Value = &n
	return nil
}

type CreateLinkRequest struct {
	LongURL   string       `json:"long_url" binding:"required,url,max=2048" example:"https://example.com"`
	ExpiresAt *time.Time   `json:"expires_at" example:"2027-04-08T00:00:00Z" format:"date-time"`
	ShortCode *string      `json:"short_code" binding:"omitempty,alphanum,min=3,max=20" example:"docs" maxLength:"20"`
	MaxClicks *int64       `json:"max_clicks" binding:"omitempty,gte=5" example:"100"`
	Status    model.Status `json:"status" binding:"omitempty,oneof=active disabled" example:"disabled"`
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

func (r CreateLinkRequest) AllowedOnWrite() error {
	if r.Status == model.StatusActive || r.Status == model.StatusDisabled {
		return nil
	}

	return InvalidStatusMessage
}

type UpdateLinkRequest struct {
	LongURL   *string           `json:"long_url" binding:"omitempty,url,max=2048" example:"https://facebook.com"`
	ExpiresAt OptionalExpiresAt `json:"expires_at" swaggertype:"string" format:"date-time" example:"2027-08-08T10:58:29Z"`
	MaxClicks OptionalMaxClicks `json:"max_clicks" binding:"omitempty" swaggertype:"integer" format:"int64" example:"50"`
	Status    *model.Status     `json:"status" binding:"omitempty,oneof=active disabled"`
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

func (r *UpdateLinkRequest) ValidateMaxClicks() error {
	if !r.MaxClicks.Present || r.MaxClicks.Value == nil {
		return nil
	}

	if *r.MaxClicks.Value < 5 {
		return MaxClicksCannotBeLess
	}

	return nil
}

type LinkSample struct {
	ID        int64        `json:"id" example:"1"`
	LongURL   string       `json:"long_url" example:"https://facebook.com"`
	ShortCode string       `json:"short_code" example:"6aAwoxoksd"`
	ExpiresAt *time.Time   `json:"expires_at" example:"2027-04-08T00:00:00Z"`
	Clicks    int64        `json:"clicks" example:"10"`
	Status    model.Status `json:"status" example:"disabled"`
	MaxClicks *int64       `json:"max_clicks" example:"50"`
	UserID    string       `json:"user_id" example:"b88a21e7-8524-428f-a109-f9eefdab3129"`
	CreatedAt time.Time    `json:"created_at" example:"2026-02-02T00:00:00Z"`
	UpdatedAt time.Time    `json:"updated_at" example:"2026-09-11T00:00:00Z"`
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
	Status    bool         `json:"status" example:"true"`
	Message   string       `json:"message" example:"Link created successfully!"`
	Data      []LinkSample `json:"data"`
	TotalPage int64        `json:"totalPage" example:"10"`
	Total     int64        `json:"total" example:"100"`
	Page      int          `json:"page" example:"1"`
	Limit     int          `json:"limit" example:"10"`
}

type GetLinkResponse struct {
	Status  bool       `json:"status" example:"true"`
	Message string     `json:"message" example:"Records fetched successfully!"`
	Data    LinkSample `json:"data"`
}

func LowerTrim(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

func getUserIdFromContext(c *gin.Context) (string, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return "", UnauthorizedError
	}
	userID, ok := userIDValue.(string)
	if !ok {
		return "", UnauthorizedError
	}
	return userID, nil
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
		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

		if err = c.ShouldBindJSON(&CreateLinkReq); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		if err = CreateLinkReq.ValidateExpiresAt(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		if CreateLinkReq.Status == "" {
			CreateLinkReq.Status = model.StatusActive
		}

		if err = CreateLinkReq.AllowedOnWrite(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		var link *model.Links
		code := ""
		if CreateLinkReq.ShortCode != nil {
			code = LowerTrim(*CreateLinkReq.ShortCode)
		}

		if code != "" {
			val, taken := reserved[code]
			if taken {
				WriteError(c, http.StatusUnprocessableEntity, "I saw it coming, only you will hit a reserved keyword error: "+val)
				return
			}
			link, err = repository.CreateLink(c.Request.Context(), pool, CreateLinkReq.LongURL, code, CreateLinkReq.ExpiresAt, CreateLinkReq.Status, CreateLinkReq.MaxClicks, userID)
			if err != nil {
				if database.IsUniqueViolationErr(err) {
					WriteError(c, http.StatusConflict, "The short code you entered is taken")
					return
				}
				WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
				return
			}
		} else {
			for range 3 {
				short, shortGenErr := shortcode.Generate(10)
				if shortGenErr != nil {
					WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
					return
				}
				link, err = repository.CreateLink(c.Request.Context(), pool, CreateLinkReq.LongURL, short, CreateLinkReq.ExpiresAt, CreateLinkReq.Status, CreateLinkReq.MaxClicks, userID)
				if err == nil {
					break
				}
				if !database.IsUniqueViolationErr(err) {
					WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
					return
				}
			}
		}

		if err != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
			return
		}

		c.JSON(http.StatusCreated, CreateLinkResponse{
			Status:  true,
			Message: "Link created successfully!",
			Data: LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				Status:    link.Status,
				MaxClicks: link.MaxClicks,
				UserID:    link.UserID,
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
// @Success 200 {object} GetLinksResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links [get]
func GetLinksHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var limit int = 10
		var page int = 1
		var links []model.Links
		var err error

		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

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
		links, err = repository.GetLinks(c.Request.Context(), pool, page, limit, userID)
		if err != nil {
			slog.Error(err.Error())
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
			return
		}

		var total int64
		total, err = repository.GetLinksTotalCount(c.Request.Context(), pool, userID)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
			return
		}
		var castedLimit = int64(limit)
		var totalPage int64 = 0

		if total == 0 {
			totalPage = 0
		} else {
			totalPage = (total + castedLimit - 1) / castedLimit
		}

		samples := make([]LinkSample, 0, len(links))
		for _, link := range links {
			samples = append(samples, LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				Status:    link.Status,
				MaxClicks: link.MaxClicks,
				UserID:    link.UserID,
				CreatedAt: link.CreatedAt,
				UpdatedAt: link.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, GetLinksResponse{
			Status:    true,
			Message:   "Links fetched successfully!",
			Data:      samples,
			TotalPage: totalPage,
			Total:     total,
			Page:      page,
			Limit:     limit,
		})
	}
}

// @Summary Get link by id
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

		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

		id, err = strconv.ParseInt(idParam, 10, 64) // alternative to int64(id)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter valid id parameter")
			return
		}

		var link *model.Links

		link, err = repository.GetLinkByID(c.Request.Context(), pool, id, userID)

		if err != nil {
			slog.Error(err.Error())
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "URL record not found.")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
			return
		}

		c.JSON(http.StatusOK, GetLinkResponse{
			Status:  true,
			Message: "Records fetched successfully!",
			Data: LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				Status:    link.Status,
				MaxClicks: link.MaxClicks,
				UserID:    link.UserID,
				CreatedAt: link.CreatedAt,
				UpdatedAt: link.UpdatedAt,
			},
		})
	}
}

// @Summary Get link by shortCode
// @Description Fetch link by short_code
// @Tags links
// @Param shortCode path string true "Short code"
// @Produce json
// @Success 200 {object} GetLinkResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links/code/{shortCode} [get]
func GetLinkByShortHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortParam := c.Param("shortCode")
		var err error

		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

		var link *model.Links

		link, err = repository.GetLinkByShortCode(c.Request.Context(), pool, shortParam, userID)

		if err != nil {
			slog.Error(err.Error())
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "URL record not found.")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
			return
		}

		c.JSON(http.StatusOK, GetLinkResponse{
			Status:  true,
			Message: "Records fetched successfully!",
			Data: LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				Status:    link.Status,
				MaxClicks: link.MaxClicks,
				UserID:    link.UserID,
				CreatedAt: link.CreatedAt,
				UpdatedAt: link.UpdatedAt,
			},
		})
	}
}

// @Summary Delete link
// @Description Delete link by id
// @Tags links
// @Param id path int true "Link ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/links/{id} [delete]
func DeleteLinkByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var id int64

		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

		var idParam = c.Param("id")
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter a valid id parameter")
			return
		}

		err = repository.DeleteLink(c.Request.Context(), pool, id, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
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

		userID, err := getUserIdFromContext(c)
		if err != nil {
			WriteError(c, http.StatusUnauthorized, err.Error())
			return
		}

		idParam := c.Param("id")
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			WriteError(c, http.StatusBadRequest, "Enter valid id param")
			return
		}

		var link *model.Links
		link, err = repository.GetLinkByID(c.Request.Context(), pool, id, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusNotFound, "Link not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMessage)
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
		if err = req.ValidateMaxClicks(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		if req.LongURL == nil && !req.ExpiresAt.Present && !req.MaxClicks.Present && req.Status == nil {
			WriteError(c, http.StatusBadRequest, "Expected at least one of long_url, expires_at, status or max_clicks")
			return
		}

		var longURL string = link.LongURL
		var expiresAt *time.Time = link.ExpiresAt
		var status model.Status = link.Status
		var maxClicks *int64 = link.MaxClicks
		if req.LongURL != nil {
			longURL = *req.LongURL
		}
		if req.Status != nil {
			status = *req.Status
		}
		if req.MaxClicks.Present {
			maxClicks = req.MaxClicks.Value
		}
		if req.ExpiresAt.Present {
			expiresAt = req.ExpiresAt.Time
		}

		link, err = repository.UpdateLinks(c.Request.Context(), pool, longURL, expiresAt, id, status, maxClicks, userID)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, UpdateLinkResponse{
			Status:  true,
			Message: "Link updated successfully!",
			Data: LinkSample{
				ID:        link.ID,
				LongURL:   link.LongURL,
				ShortCode: link.ShortCode,
				ExpiresAt: link.ExpiresAt,
				Clicks:    link.Clicks,
				Status:    link.Status,
				MaxClicks: link.MaxClicks,
				UserID:    link.UserID,
				CreatedAt: link.CreatedAt,
				UpdatedAt: link.UpdatedAt,
			},
		})
	}
}
