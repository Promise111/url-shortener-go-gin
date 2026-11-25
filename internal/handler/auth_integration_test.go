package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/queue"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
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

	godotenv.Load("../../.env")
	cfg := config.Load()

	pool, err := database.Connect(cfg.TestDatabaseURL)
	if err != nil {
		t.Skipf("Database unavailable ", err)
	}

	rds, err := database.ConnectRedis(cfg.TestRedisURL)
	if err != nil {
		t.Skipf("Redis connection failed ", err)
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

func TestRegisterAndLoginGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var cfg *config.Config
	var pool *pgxpool.Pool
	var err error
	var rds *redis.Client
	var asynqClient *asynq.Client

	_ = godotenv.Load("../../.env")
	cfg = config.Load()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	test_redis := os.Getenv("TEST_REDIS")

	if test_redis == "" {
		t.Skip("TEST_REDIS not set")
	}

	pool, err = database.Connect(dsn)
	if err != nil {
		t.Skip("database unavailable: ", err)
	}
	defer pool.Close()

	rds, err = database.ConnectRedis(test_redis)
	if err != nil {
		t.Skip("Redis connection failed: ", err)
	}
	defer rds.Close()

	asynqClient = queue.NewClient(test_redis)
	defer asynqClient.Close()

	randEmailGenerator := fmt.Sprintf("u-%d@test.local", time.Now().UnixNano())
	username := fmt.Sprintf("user%d", time.Now().UnixNano())

	bodyString := fmt.Sprintf(`{"email":"%s","password":"MyPass234#","username":"%s"}`, randEmailGenerator, username)
	body := strings.NewReader(bodyString)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := router.Router(pool, cfg, rds, asynqClient)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	loginBodyString := fmt.Sprintf(`{"Email":"%s","password":"MyPass234#"}`, randEmailGenerator)
	loginBody := strings.NewReader(loginBodyString)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	r.ServeHTTP(w, loginRequest)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}
