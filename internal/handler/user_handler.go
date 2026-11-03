package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/model"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/Promise111/url-shortener-go-gin/internal/verificationToken"
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
	Password string `json:"password" binding:"required,min=8,max=72" example:"myPass123"`
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
	Status  bool       `json:"status" example:"true"`
	Message string     `json:"message" example:"User created successfully."`
	Data    UserSample `json:"data"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email" example:"hello@hi.com"`
	Username string `json:"username" binding:"omitempty,min=3,max=30" example:"randomgee"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"****************"`
}

func (r *LoginUserRequest) ValidateEmailUsername() error {
	var trimmedEmail = strings.TrimSpace(r.Email)
	var trimmedUsername = strings.TrimSpace(r.Username)
	if trimmedEmail == "" && trimmedUsername == "" {
		return EmailOrUsernameRequired
	}
	return nil
}

type LoginResponseData struct {
	User         UserSample `json:"user"`
	AccessToken  string     `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyNzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMC1hYjQzLTUxOTM3ZTUxMGM2OSIsInVzZ0.n5vU3bqtslWmIlCxlUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
	RefreshToken string     `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMxMGM2OSIsInVzZXJuYW1lIjoicmFuZG9tZ2VlIn0.n5vU3bqUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
}

type LoginResponse struct {
	Status bool              `json:"status" example:"true"`
	Data   LoginResponseData `json:"data"`
}

type GenerateAccessRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyNzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMC1hYjQzLTUxOTM3ZTUxMGM2OSIsInVzZ0.n5vU3bqtslWmIlCxlUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
}

type GenerateAccessResponseData struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyNzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMC1hYjQzLTUxOTM3ZTUxMGM2OSIsInVzZ0.n5vU3bqtslWmIlCxlUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMxMGM2OSIsInVzZXJuYW1lIjoicmFuZG9tZ2VlIn0.n5vU3bqUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
}

type GenerateAccessResponse struct {
	Status  bool                       `json:"status" example:"true"`
	Message string                     `json:"message" example:"Access token generated successfully"`
	Data    GenerateAccessResponseData `json:"data"`
}

type RevokeAccessResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"Refresh token revoked successfully"`
}

type VerifyEmailResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"Email verified successfully"`
}

type RevokeRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyNzkxNDg2NzgzLCJpZCI6IjUxMWE3NDNhLWFlNGItNGZhMC1hYjQzLTUxOTM3ZTUxMGM2OSIsInVzZ0.n5vU3bqtslWmIlCxlUNOlpb9V_CxXWHgM53Ha2ZbmNY"`
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

func genTokenExpiration(tokenType TokenType, accessSeconds int, refreshSeconds int) int64 {
	switch tokenType {
	case TokenTypeAccess:
		return time.Now().Add(time.Second * time.Duration(accessSeconds)).Unix()
	case TokenTypeRefresh:
		return time.Now().Add(time.Second * time.Duration(refreshSeconds)).Unix()
	default:
		return time.Now().Add(time.Minute * 15).Unix()
	}
}

func GenerateToken(user model.Users, cfg *config.Config, tokenType TokenType) (string, error) {
	var accessSeconds, accessSecondsErr = strconv.Atoi(cfg.AccessExpiration)
	if accessSecondsErr != nil {
		return "", accessSecondsErr
	}

	var refreshSeconds, refreshSecondsErr = strconv.Atoi(cfg.RefreshExpiration)
	if refreshSecondsErr != nil {
		return "", refreshSecondsErr
	}

	var expiration = genTokenExpiration(tokenType, accessSeconds, refreshSeconds)

	var claims jwt.MapClaims = jwt.MapClaims{
		"id":        user.ID,
		"email":     user.Email,
		"username":  user.Username,
		"exp":       expiration,
		"tokenType": tokenType,
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
// @Failure 409 {object} map[string]any
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
		user, err = repository.CreateUser(c.Request.Context(), pool, LowerTrim(input.Email), string(password_hash), LowerTrim(input.Username))
		if err != nil {
			slog.Error("handler", "err", err.Error())
			if database.IsUniqueViolationErr(err) {
				WriteError(c, http.StatusConflict, "Email or username taken.")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		// generate email verification token
		verificationToken, generateVerificationTokenError := verificationToken.GenerateVerificationToken()

		if generateVerificationTokenError != nil {
			slog.Error("Generate verification error", "error", generateVerificationTokenError)
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		slog.Info("Email verification token", "token", "http://localhost:8003/api/v1/auth/verify-email?token="+verificationToken)

		verificationTokenHash := hashToken(verificationToken)

		createEmailVerificationTokenError := repository.CreateEmailVerificationToken(c.Request.Context(), pool, verificationTokenHash, user.ID)
		if createEmailVerificationTokenError != nil {
			slog.Error("Create email verification token error", "error", createEmailVerificationTokenError)
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
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
// @Description Login and return access + refresh JWTs (refresh hash persisted in DB)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginUserRequest true "Login payload"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var input LoginUserRequest
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
		if email == "" && username != "" {
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

		if user.EmailVerifiedAt == nil {
			WriteError(c, http.StatusForbidden, "Please verify your email")
			return
		}

		var accessToken string
		var refreshToken string

		if accessToken, err = GenerateToken(*user, cfg, TokenTypeAccess); err != nil {
			slog.Error("login", "err", err.Error())
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		if refreshToken, err = GenerateToken(*user, cfg, TokenTypeRefresh); err != nil {
			slog.Error("login", "err", err.Error())
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		var refreshHash = hashToken(refreshToken)
		refreshExpSeconds, refreshExpErr := strconv.Atoi(cfg.RefreshExpiration)
		if refreshExpErr != nil {
			slog.Error("refresh token hash", "error", refreshExpErr)
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}
		var expiration = time.Now().Add(time.Second * time.Duration(refreshExpSeconds))

		userID := user.ID

		_, persistRefreshTokenError := repository.PersistRefreshToken(c, pool, refreshHash, userID, expiration)

		if persistRefreshTokenError != nil {
			slog.Error("persist refresh token error", "error", persistRefreshTokenError)
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusOK, LoginResponse{
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
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		})
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// @Summary Generate Access and Refresh Tokens
// @Description Allow user generate access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GenerateAccessRequest true "Generate access payload"
// @Success 200 {object} GenerateAccessResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func GenerateAccess(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GenerateAccessRequest
		if bindError := c.ShouldBindJSON(&request); bindError != nil {
			WriteError(c, http.StatusBadRequest, bindError.Error())
			return
		}

		refreshHash := hashToken(request.RefreshToken)

		refreshRecord, refreshRecordError := repository.GetRefreshTokenByHash(c.Request.Context(), pool, refreshHash)

		if refreshRecordError != nil {
			if errors.Is(refreshRecordError, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		user, getUserErr := repository.GetUserByID(c.Request.Context(), pool, refreshRecord.UserID)
		if getUserErr != nil {
			if errors.Is(getUserErr, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "User not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		// delete previous refresh
		deleteRefreshHashError := repository.DeleteRefreshToken(c.Request.Context(), pool, refreshHash)

		if deleteRefreshHashError != nil {
			if errors.Is(deleteRefreshHashError, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "Invalid or expired refresh token")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		accessToken, accessTokenErr := GenerateToken(*user, cfg, TokenTypeAccess)
		if accessTokenErr != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		refreshToken, refreshTokenErr := GenerateToken(*user, cfg, TokenTypeRefresh)
		if refreshTokenErr != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		var newRefreshHash = hashToken(refreshToken)

		expirationSeconds := cfg.RefreshExpiration

		intExpiresAt, convErr := strconv.Atoi(expirationSeconds)
		if convErr != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		_, persistTokenErr := repository.PersistRefreshToken(c.Request.Context(), pool, newRefreshHash, user.ID, time.Now().Add(time.Second*time.Duration(intExpiresAt)))

		if persistTokenErr != nil {
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusOK, GenerateAccessResponse{
			Status:  true,
			Message: "Access token generated successfully",
			Data: GenerateAccessResponseData{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		})
	}
}

// @Summary Logout
// @Description Allow user logout and invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RevokeRefreshRequest true "Logout payload"
// @Success 200 {object} RevokeAccessResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func RevokeRefreshToken(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RevokeRefreshRequest

		err := c.ShouldBindJSON(&request)
		if err != nil {
			WriteError(c, http.StatusBadRequest, err.Error())
			return
		}

		hashedToken := hashToken(request.RefreshToken)

		_, refreshHashRecordErr := repository.RevokeRefreshToken(c.Request.Context(), pool, time.Now(), hashedToken)
		if refreshHashRecordErr != nil {
			if errors.Is(refreshHashRecordErr, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, refreshHashRecordErr.Error())
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusOK, RevokeAccessResponse{
			Status:  true,
			Message: "Refresh token revoked successfully",
		})
	}
}

// @Summary Verify Email
// @Description Allow user verify email
// @Tags auth
// @Produce json
// @Param token query string true "Email verification token"
// @Success 200 {object} VerifyEmailResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/verify-email [post]
func VerifyEmail(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token = c.Query("token")
		if strings.TrimSpace(token) == "" {
			WriteError(c, http.StatusBadRequest, "Please enter a valid token query")
			return
		}
		tokenHash := hashToken(token)
		tokenHashRecord, tokenHashRecordError := repository.GetEmailVerificationToken(c.Request.Context(), pool, tokenHash)
		if tokenHashRecordError != nil {
			if errors.Is(tokenHashRecordError, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		user, getUserError := repository.GetUserByID(c.Request.Context(), pool, tokenHashRecord.UserID)
		if getUserError != nil {
			if errors.Is(getUserError, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "User not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		if user.EmailVerifiedAt != nil {
			c.JSON(http.StatusOK, VerifyEmailResponse{
				Status:  true,
				Message: "Email already verified",
			})
			return
		}

		updateEmailVerifiedAtError := repository.UpdateEmailVerifiedAt(c.Request.Context(), pool, user.ID)

		if updateEmailVerifiedAtError != nil {
			if errors.Is(updateEmailVerifiedAtError, pgx.ErrNoRows) {
				WriteError(c, http.StatusUnauthorized, "User not found")
				return
			}
			WriteError(c, http.StatusInternalServerError, InternalServerErrorMsg)
			return
		}

		c.JSON(http.StatusOK, VerifyEmailResponse{
			Status:  true,
			Message: "Email verified successfully",
		})
	}
}
