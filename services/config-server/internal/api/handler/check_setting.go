package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// CheckSettingHandler handles HTTP requests for check settings
type CheckSettingHandler struct {
	checkService *service.CheckSettingService
}

// NewCheckSettingHandler creates a new CheckSettingHandler
func NewCheckSettingHandler(checkService *service.CheckSettingService) *CheckSettingHandler {
	return &CheckSettingHandler{
		checkService: checkService,
	}
}

// Create handles POST /check-settings
func (h *CheckSettingHandler) Create(c *gin.Context) {
	var req dto.CreateCheckSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	setting, err := h.checkService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToCheckSettingResponse(setting))
}

// GetByID handles GET /check-settings/:id
func (h *CheckSettingHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	setting, err := h.checkService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckSettingResponse(setting))
}

// Update handles PUT /check-settings/:id
func (h *CheckSettingHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCheckSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	setting, err := h.checkService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckSettingResponse(setting))
}

// DeleteResource handles POST /check-settings/delete
func (h *CheckSettingHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.checkService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /check-settings/purge
func (h *CheckSettingHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.checkService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /check-settings/restore
func (h *CheckSettingHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.checkService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /check-settings
func (h *CheckSettingHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	settings, total, err := h.checkService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToCheckSettingResponseList(settings), total, pagination)
}

// GetByGroupID handles GET /check-settings/group/:group_id
func (h *CheckSettingHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("group_id")

	settings, err := h.checkService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckSettingResponseList(settings))
}

// GetByCheckType handles GET /check-settings/type/:type
func (h *CheckSettingHandler) GetByCheckType(c *gin.Context) {
	checkType := c.Param("type")

	settings, err := h.checkService.GetByCheckType(c.Request.Context(), checkType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckSettingResponseList(settings))
}
