package handler

import "github.com/gin-gonic/gin"

func WriteError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"status":  false,
		"message": msg,
	})
}
