package router

import (
	"github.com/Promise111/url-shortener-go-gin/internal/config"
	"github.com/gin-gonic/gin"
)

func Router(cfg *config.Config) *gin.Engine {
	var r = gin.Default()

	return r
}
