package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// ServiceDiscoveryHandler handles service discovery HTTP requests
type ServiceDiscoveryHandler struct {
	sdService *service.ServiceDiscoveryService
}

// NewServiceDiscoveryHandler creates a new ServiceDiscoveryHandler instance
func NewServiceDiscoveryHandler(sdService *service.ServiceDiscoveryService) *ServiceDiscoveryHandler {
	return &ServiceDiscoveryHandler{
		sdService: sdService,
	}
}

// GetPrometheusTargets handles GET /api/v1/sd/prometheus
// Returns all targets in Prometheus HTTP SD format with optional filters
func (h *ServiceDiscoveryHandler) GetPrometheusTargets(c *gin.Context) {
	var filterReq dto.ServiceDiscoveryFilterRequest
	if err := c.ShouldBindQuery(&filterReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := filterReq.ToDomainFilter()
	targets, err := h.sdService.GetPrometheusTargets(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPrometheusSDTargetResponseList(targets))
}

// GetPrometheusTargetsByGroup handles GET /api/v1/sd/prometheus/group/:groupId
// Returns targets for a specific group
func (h *ServiceDiscoveryHandler) GetPrometheusTargetsByGroup(c *gin.Context) {
	groupID := c.Param("groupId")
	enabledOnly := c.DefaultQuery("enabled_only", "false") == "true"

	targets, err := h.sdService.GetPrometheusTargetsForGroup(c.Request.Context(), groupID, enabledOnly)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPrometheusSDTargetResponseList(targets))
}

// GetPrometheusTargetsByNamespace handles GET /api/v1/sd/prometheus/namespace/:namespaceId
// Returns targets for a specific namespace
func (h *ServiceDiscoveryHandler) GetPrometheusTargetsByNamespace(c *gin.Context) {
	namespaceID := c.Param("namespaceId")
	enabledOnly := c.DefaultQuery("enabled_only", "false") == "true"

	targets, err := h.sdService.GetPrometheusTargetsForNamespace(c.Request.Context(), namespaceID, enabledOnly)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPrometheusSDTargetResponseList(targets))
}

// GetActivePrometheusTargets handles GET /api/v1/sd/prometheus/active
// Returns only active targets with enabled exporters
func (h *ServiceDiscoveryHandler) GetActivePrometheusTargets(c *gin.Context) {
	targets, err := h.sdService.GetActivePrometheusTargets(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPrometheusSDTargetResponseList(targets))
}

// GenerateFileSD handles POST /api/v1/sd/prometheus/file
// Generates a Prometheus file SD JSON file
func (h *ServiceDiscoveryHandler) GenerateFileSD(c *gin.Context) {
	var req struct {
		OutputPath string                            `json:"output_path" binding:"required"`
		Filter     dto.ServiceDiscoveryFilterRequest `json:"filter"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := req.Filter.ToDomainFilter()
	if err := h.sdService.GenerateFileSD(c.Request.Context(), req.OutputPath, filter); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File SD generated successfully",
		"path":    req.OutputPath,
	})
}

// GenerateActiveFileSD handles POST /api/v1/sd/prometheus/file/active
// Generates a file SD JSON containing only active targets
func (h *ServiceDiscoveryHandler) GenerateActiveFileSD(c *gin.Context) {
	var req struct {
		OutputPath string `json:"output_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sdService.GenerateActiveFileSD(c.Request.Context(), req.OutputPath); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Active targets file SD generated successfully",
		"path":    req.OutputPath,
	})
}

// GenerateGroupFileSD handles POST /api/v1/sd/prometheus/file/group/:groupId
// Generates a file SD JSON for a specific group
func (h *ServiceDiscoveryHandler) GenerateGroupFileSD(c *gin.Context) {
	groupID := c.Param("groupId")
	var req struct {
		OutputPath  string `json:"output_path" binding:"required"`
		EnabledOnly bool   `json:"enabled_only"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sdService.GenerateGroupFileSD(c.Request.Context(), groupID, req.OutputPath, req.EnabledOnly); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Group file SD generated successfully",
		"path":     req.OutputPath,
		"group_id": groupID,
	})
}

// GenerateNamespaceFileSD handles POST /api/v1/sd/prometheus/file/namespace/:namespaceId
// Generates a file SD JSON for a specific namespace
func (h *ServiceDiscoveryHandler) GenerateNamespaceFileSD(c *gin.Context) {
	namespaceID := c.Param("namespaceId")
	var req struct {
		OutputPath  string `json:"output_path" binding:"required"`
		EnabledOnly bool   `json:"enabled_only"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sdService.GenerateNamespaceFileSD(c.Request.Context(), namespaceID, req.OutputPath, req.EnabledOnly); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Namespace file SD generated successfully",
		"path":         req.OutputPath,
		"namespace_id": namespaceID,
	})
}
