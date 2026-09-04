package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": "💪 Shortener API is up and running!",
	})
}