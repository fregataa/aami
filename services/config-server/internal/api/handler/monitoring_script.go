package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// MonitoringScriptHandler handles HTTP requests for monitoring scripts
type MonitoringScriptHandler struct {
	scriptService *service.MonitoringScriptService
}

// NewMonitoringScriptHandler creates a new MonitoringScriptHandler
func NewMonitoringScriptHandler(scriptService *service.MonitoringScriptService) *MonitoringScriptHandler {
	return &MonitoringScriptHandler{
		scriptService: scriptService,
	}
}

// Create handles POST /monitoring-scripts
func (h *MonitoringScriptHandler) Create(c *gin.Context) {
	var req dto.CreateMonitoringScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.scriptService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToMonitoringScriptResponse(result))
}

// GetByID handles GET /monitoring-scripts/:id
func (h *MonitoringScriptHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.scriptService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMonitoringScriptResponse(result))
}

// GetByName handles GET /monitoring-scripts/name/:name
func (h *MonitoringScriptHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	result, err := h.scriptService.GetByName(c.Request.Context(), name)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMonitoringScriptResponse(result))
}

// GetByScriptType handles GET /monitoring-scripts/type/:scriptType
func (h *MonitoringScriptHandler) GetByScriptType(c *gin.Context) {
	scriptType := c.Param("scriptType")

	results, err := h.scriptService.GetByScriptType(c.Request.Context(), scriptType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMonitoringScriptSummaryResponseList(results))
}

// Update handles PUT /monitoring-scripts/:id
func (h *MonitoringScriptHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateMonitoringScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.scriptService.Update(c.Request.Context(), id, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMonitoringScriptResponse(result))
}

// DeleteResource handles POST /monitoring-scripts/delete
func (h *MonitoringScriptHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.scriptService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /monitoring-scripts/purge
func (h *MonitoringScriptHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.scriptService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /monitoring-scripts/restore
func (h *MonitoringScriptHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.scriptService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /monitoring-scripts
func (h *MonitoringScriptHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.scriptService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToMonitoringScriptSummaryResponseList(listResult.Items), listResult.Total, pagination)
}

// ListActive handles GET /monitoring-scripts/active
func (h *MonitoringScriptHandler) ListActive(c *gin.Context) {
	results, err := h.scriptService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMonitoringScriptSummaryResponseList(results))
}

// VerifyHash handles GET /monitoring-scripts/:id/verify-hash
func (h *MonitoringScriptHandler) VerifyHash(c *gin.Context) {
	id := c.Param("id")

	valid, err := h.scriptService.VerifyHash(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template_id": id,
		"hash_valid":  valid,
	})
}
