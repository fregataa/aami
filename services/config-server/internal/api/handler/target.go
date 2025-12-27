package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// TargetHandler handles HTTP requests for targets
type TargetHandler struct {
	targetService *service.TargetService
}

// NewTargetHandler creates a new TargetHandler
func NewTargetHandler(targetService *service.TargetService) *TargetHandler {
	return &TargetHandler{
		targetService: targetService,
	}
}

// Create handles POST /targets
func (h *TargetHandler) Create(c *gin.Context) {
	var req dto.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	target, err := h.targetService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToTargetResponse(target))
}

// GetByID handles GET /targets/:id
func (h *TargetHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	target, err := h.targetService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTargetResponse(target))
}

// GetByHostname handles GET /targets/hostname/:hostname
func (h *TargetHandler) GetByHostname(c *gin.Context) {
	hostname := c.Param("hostname")

	target, err := h.targetService.GetByHostname(c.Request.Context(), hostname)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTargetResponse(target))
}

// Update handles PUT /targets/:id
func (h *TargetHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	target, err := h.targetService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTargetResponse(target))
}

// DeleteResource handles POST /targets/delete
func (h *TargetHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.targetService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /targets/purge
func (h *TargetHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.targetService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /targets/restore
func (h *TargetHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.targetService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /targets
func (h *TargetHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	targets, total, err := h.targetService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToTargetResponseList(targets), total, pagination)
}

// GetByGroupID handles GET /targets/group/:group_id
func (h *TargetHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("group_id")

	targets, err := h.targetService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTargetResponseList(targets))
}

// UpdateStatus handles POST /targets/:id/status
func (h *TargetHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTargetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.targetService.UpdateStatus(c.Request.Context(), id, req); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Heartbeat handles POST /targets/:id/heartbeat
func (h *TargetHandler) Heartbeat(c *gin.Context) {
	id := c.Param("id")

	if err := h.targetService.Heartbeat(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
