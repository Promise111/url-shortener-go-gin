package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"status":  false,
		"message": msg,
	})
}

func WriteBindError(c *gin.Context, errorSlice []Error) {
	if errorSlice != nil {
		c.JSON(http.StatusBadRequest, ValidationError{
			Status:  false,
			Message: "Bad request",
			Errors: errorSlice,
		})
		return
	}

	c.JSON(http.StatusBadRequest, ValidationError{
		Status:  false,
		Message: "Bad request",
	})
}
