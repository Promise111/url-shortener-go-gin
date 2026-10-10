package main

import (
	"log/slog"
	"os"

	docs "github.com/Promise111/url-shortener-go-gin/cmd/shortener-api/docs"
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/database"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
// @description Paste: Bearer <your_login_token> (capital B, then a space, then the login JWT)

// @host		localhost:8003
// @BasePath	/
// @schemes	http
func main() {
	slog.Info("🚀 Shortener API Server Started!")

	cfg, configErr := config.Load()
	if configErr != nil {
		slog.Error("Failed to load configuration!")
		os.Exit(1)
	}

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

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.Schemes = []string{"http"}

	var r = router.Router(pool, cfg, rds)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":" + cfg.Port)
}
