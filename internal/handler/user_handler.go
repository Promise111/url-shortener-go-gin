package handler

import (
	"net/http"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var InternalServerErrorMsg = "Something went wrong."

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"hello@hi.com"`
	Password string `json:"password" binding:"required,min=6,max=30" example:"myPass123"`
	Username string `json:"username" binding:"required,min=3,max=30" example:"randomgee"`
}

func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

type UserSample struct {
	ID           string     `json:"id" example:"b0f7915c-5b4e-4c1a-8fc8-d6e9663f3709"`
	Email        string     `json:"email" example:"hello@hi.com"`
	PasswordHash string     `json:"password_hash" example:"****************"`
	Username     string     `json:"username" example:"randomgee"`
	DeletedAt    *time.Time `json:"deleted_at" example:""`
	CreatedAt    time.Time  `json:"created_at" example:"2026-02-02T00:00:00Z"`
	UpdatedAt    time.Time  `json:"updated_at" example:"2027-08-02T00:00:00Z"`
}

type CreateUserResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"User created successfully."`
	Data    UserSample
}

// @Summary Register User
// @Description Create new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "Create Link payload"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/signup [post]
func RegisterUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var input RegisterUserRequest
		if err = c.ShouldBindJSON(&input); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}
		cost := bcrypt.DefaultCost
		password_hash, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), cost)
		if hashErr != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		var user *model.Users
		user, err = repository.CreateUser(c, pool, input.Email, string(password_hash), input.Username)
		if err != nil {
			if database.IsUniqueViolationErr(err) {
				WriteError(c, http.StatusConflict, "Email or username taken.")
				return
			}
			WriteError(c, http.StatusBadRequest, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusCreated, CreateUserResponse{
			Status:  true,
			Message: "User registered successfully",
			Data: UserSample{
				ID:           user.ID,
				Email:        user.Email,
				PasswordHash: user.PasswordHash,
				Username:     user.Username,
				DeletedAt:    user.DeletedAt,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
			},
		})
	}
}
