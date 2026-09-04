package main

import (
	"log/slog"
	"os"

	docs "github.com/Promise111/url-shortener-go-gin/cmd/shortener-api/docs"
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
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

// @host		localhost:8003
// @BasePath	/api/v1
// @schemes	http
func main() {
	slog.Info("🚀 Shortener API Server Started!")

	cfg, configErr := config.Load()
	if configErr != nil {
		slog.Error("Failed to load configuration!")
		os.Exit(1)
	}

	docs.SwaggerInfo.BasePath = router.APIPrefix
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.Schemes = []string{"http"}

	var r = router.Router(cfg)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":" + cfg.Port)
}
