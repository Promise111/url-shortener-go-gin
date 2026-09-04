package main

import (
	"log/slog"
	"os"

	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/Promise111/url-shortener-go-gin/internal/router"
)

func main() {
	slog.Info("🚀 Shortener API Server Started!")
	cfg, configErr := config.Load()
	if configErr != nil {
		slog.Error("Failed to load configuration!")
		os.Exit(1)
	}

	var r = router.Router(cfg)
	r.Run(":" + cfg.Port)
}
