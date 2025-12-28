package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/domain"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check HTTP requests
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// CheckHealth handles GET /health
// Returns complete health status including all components
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	health := h.healthService.CheckHealth(c.Request.Context())

	statusCode := http.StatusOK
	if health.Status == domain.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == domain.HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	c.JSON(statusCode, health)
}

// CheckReadiness handles GET /health/ready
// Returns readiness status for Kubernetes readiness probe
func (h *HealthHandler) CheckReadiness(c *gin.Context) {
	health := h.healthService.CheckReadiness(c.Request.Context())

	statusCode := http.StatusOK
	if !health.IsReady() {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// CheckLiveness handles GET /health/live
// Returns liveness status for Kubernetes liveness probe
func (h *HealthHandler) CheckLiveness(c *gin.Context) {
	health := h.healthService.CheckLiveness(c.Request.Context())

	statusCode := http.StatusOK
	if health.Status == domain.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}
