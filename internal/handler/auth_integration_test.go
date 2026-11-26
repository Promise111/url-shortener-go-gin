package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/queue"
	"github.com/Promise111/url-shortener-go-gin/internal/repository"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
	"github.com/Promise111/url-shortener-go-gin/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type testApp struct {
	cfg   *config.Config
	pool  *pgxpool.Pool
	asynq *asynq.Client
	rds   *redis.Client
	r     *gin.Engine
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()

	gin.SetMode(gin.TestMode)

	godotenv.Load("../../.env")
	cfg := config.Load()

	if cfg.TestDatabaseURL == "" || cfg.TestRedisURL == "" {
		t.Skip("Test DB or redis URLs not set")
	}

	pool, err := database.Connect(cfg.TestDatabaseURL)
	if err != nil {
		t.Skipf("Database unavailable %v", err)
	}

	rds, err := database.ConnectRedis(cfg.TestRedisURL)
	if err != nil {
		t.Skipf("Redis connection failed %v", err)
	}

	asynqClient := queue.NewClient(cfg.TestRedisURL)

	r := router.Router(pool, cfg, rds, asynqClient)

	t.Cleanup(func() {
		asynqClient.Close()
		rds.Close()
		pool.Close()
	})
	return &testApp{asynq: asynqClient, cfg: cfg, pool: pool, rds: rds, r: r}
}

func doJSON(r *gin.Engine, method string, path string, body io.Reader) *httptest.ResponseRecorder {
	var req = httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	var writer = httptest.NewRecorder()
	r.ServeHTTP(writer, req)
	return writer
}

func TestRegisterAndLoginGate(t *testing.T) {
	app := setupTestApp(t)

	randEmailGenerator := fmt.Sprintf("u-%d@test.local", time.Now().UnixNano())
	username := fmt.Sprintf("user%d", time.Now().UnixNano())
	password := "MyPass234#"
	r := app.r

	bodyString := fmt.Sprintf(`{"email":"%s","password":"%s","username":"%s"}`, randEmailGenerator, password, username)
	body := strings.NewReader(bodyString)
	registerWriter := doJSON(r, http.MethodPost, "/api/v1/auth/register", body)

	if registerWriter.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", registerWriter.Code, registerWriter.Body.String())
	}

	loginBodyString := fmt.Sprintf(`{"email":"%s","password":"%s"}`, randEmailGenerator, password)

	loginBody := strings.NewReader(loginBodyString)
	loginWriter := doJSON(r, http.MethodPost, "/api/v1/auth/login", loginBody)

	if loginWriter.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", loginWriter.Code, loginWriter.Body.String())
	}

	tokenString, err := token.GenerateVerificationToken()
	if err != nil {
		t.Fatalf("Generating token failed %v", err)
	}

	ctx := context.Background()
	tokenHash := token.HashToken(tokenString)
	user, err := repository.GetUserByEmail(ctx, app.pool, randEmailGenerator)
	if err != nil {
		t.Fatalf("Failed to get user by email %v", err)
	}

	err = repository.DeleteEmailVerificationTokenByUserID(ctx, app.pool, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete email verification token %v", err)
	}

	err = repository.CreateEmailVerificationToken(ctx, app.pool, tokenHash, user.ID)
	if err != nil {
		t.Fatalf("Failed to create email verification token %v", err)
	}

	verifyURLPath := fmt.Sprintf("/api/v1/auth/verify-email?token=%s", tokenString)
	verifyEmailWriter  := doJSON(r, http.MethodPost, verifyURLPath, nil)

	if verifyEmailWriter.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", verifyEmailWriter.Code, verifyEmailWriter.Body.String())
	}

	loginSuccessBody := strings.NewReader(loginBodyString)
	loginSuccessWriter  := doJSON(r, http.MethodPost, "/api/v1/auth/login", loginSuccessBody)

	if loginSuccessWriter.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", loginSuccessWriter.Code, loginSuccessWriter.Body.String())
	}
}
