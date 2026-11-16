package main

import (
	"log/slog"
	"os"
	"reflect"
	"strings"

	docs "github.com/Promise111/url-shortener-go-gin/cmd/shortener-api/docs"
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/queue"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupJSONFieldNames() {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	engine.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

//	@title			URL Shortener
//	@version		1.0
//	@description	API for creating and resolving shortened URLs.

//	@contact.name	Promise
//	@contact.email	promiseihunna@gmail.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Paste: Bearer <accessToken> (capital B, then a space, then the access JWT from login/refresh — not the refresh token)

// @host		localhost:8003
// @BasePath	/
// @schemes	http
func main() {
	slog.Info("🚀 Shortener API Server Started!")

	cfg := config.Load()

	var pool *pgxpool.Pool
	var dbConnErr error
	pool, dbConnErr = database.Connect(cfg.DatabaseURL)
	if dbConnErr != nil {
		slog.Error("Database connection failed " + dbConnErr.Error())
		os.Exit(1)
	}
	defer pool.Close()

	rds, redisConErr := database.ConnectRedis(cfg.RedisADDR)
	if redisConErr != nil {
		slog.Error("Redis connection failed " + redisConErr.Error())
		os.Exit(1)
	}
	defer rds.Close()

	var asynqClient *asynq.Client = queue.NewClient(cfg.RedisADDR)
	defer asynqClient.Close()

	setupJSONFieldNames()

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.Schemes = []string{"http"}

	var r = router.Router(pool, cfg, rds, asynqClient)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":" + cfg.Port)
}
