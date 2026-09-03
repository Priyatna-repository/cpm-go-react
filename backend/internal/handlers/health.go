package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns service status, used for uptime checks and CI/CD deploy verification
// @Tags         system
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
