package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status  bool   `json:"status" example:"true"`
	Message string `json:"message" example:"💪 Shortener API is up and running!"`
}

// HealthHandler reports whether the API process is running.
//
//	@Summary		Health check
//	@Description	Returns the health status of the shortener API.
//	@Tags			health
//	@ID				healthCheck
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  true,
		Message: "💪 Shortener API is up and running!",
	})
}
