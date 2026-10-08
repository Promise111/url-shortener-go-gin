package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var InternalServerErrorMsg = "Something went wrong."
var EmailOrUsernameRequired = errors.New("At least one of Email or Username is required")
var InvalidCredentials = "Invalid credentials."

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"hello@hi.com"`
	Password string `json:"password" binding:"required,min=6,max=30" example:"myPass123"`
	Username string `json:"username" binding:"required,min=3,max=30" example:"randomgee"`
}

type UserSample struct {
	ID        string     `json:"id" example:"b0f7915c-5b4e-4c1a-8fc8-d6e9663f3709"`
	Email     string     `json:"email" example:"hello@hi.com"`
	Username  string     `json:"username" example:"randomgee"`
	DeletedAt *time.Time `json:"deleted_at" example:""`
	CreatedAt time.Time  `json:"created_at" example:"2026-02-02T00:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2027-08-02T00:00:00Z"`
}

type CreateUserResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"User created successfully."`
	Data    UserSample
}

type LoginRequest struct {
	Email    string `json:"email" binding:"omitempty,email" example:"hello@hi.com"`
	Username string `json:"username" binding:"omitempty,min=3,max=30" example:"randomgee"`
	Password string `json:"password" binding:"required,min=6,max=30" example:"****************"`
}

func (r *LoginRequest) ValidateEmailUsername() error {
	if r.Email == "" && r.Username == "" {
		return EmailOrUsernameRequired
	}
	return nil
}

type LoginResponseData struct {
	User  UserSample `json:"user"`
	Token string     `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InByb21pc2VpaHVubmFAaGkuY29tIiwiZXhwIjoxNzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMC1hYjQzLTUxOTM3ZTUxMGM2OSIsInVzZXJuYW1lIjoicmFuZG9tZ2VlIn0.n5vU3bqtslWmIlCxlUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
}

type LoginResponse struct {
	Status bool `json:"status" example:"true"`
	Data   LoginResponseData
}

func GenerateToken(user model.Users, cfg *config.Config) (string, error) {
	var claims jwt.MapClaims = jwt.MapClaims{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"exp":      time.Now().Add(5 * time.Hour).Unix(),
	}

	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// @Summary Register User
// @Description Create new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "Create user payload"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
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
		user, err = repository.CreateUser(c, pool, LowerTrim(input.Email), string(password_hash), LowerTrim(input.Username))
		if err != nil {
			slog.Error("handler", "err", err.Error())
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
				ID:        user.ID,
				Email:     user.Email,
				Username:  user.Username,
				DeletedAt: user.DeletedAt,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}

// @Summary Login User
// @Description Allow user login and generate access code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "Login payload"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var input LoginRequest
		if err = c.ShouldBindJSON(&input); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}
		if err = input.ValidateEmailUsername(); err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}
		var email = LowerTrim(input.Email)
		var username = LowerTrim(input.Username)
		var password = input.Password
		var user *model.Users
		if email != "" {
			user, err = repository.GetUserByEmail(c.Request.Context(), pool, email)
		}
		if username != "" {
			user, err = repository.GetUserByUsername(c.Request.Context(), pool, username)
		}
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, InvalidCredentials)
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			WriteError(c, http.StatusUnauthorized, InvalidCredentials)
			return
		}

		var token string

		token, err = GenerateToken(*user, cfg)
		if err != nil {
			slog.Error("login", "err", err.Error())
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusAccepted, LoginResponse{
			Status: true,
			Data: LoginResponseData{
				User: UserSample{
					ID:        user.ID,
					Email:     user.Email,
					Username:  user.Username,
					DeletedAt: user.DeletedAt,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
				},
				Token: token,
			},
		})
	}
}
