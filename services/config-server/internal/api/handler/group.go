package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// GroupHandler handles HTTP requests for groups
type GroupHandler struct {
	groupService *service.GroupService
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// Create handles POST /groups
func (h *GroupHandler) Create(c *gin.Context) {
	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.groupService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToGroupResponse(result))
}

// GetByID handles GET /groups/:id
func (h *GroupHandler) GetByID(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	result, err := h.groupService.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponse(result))
}

// Update handles PUT /groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.groupService.Update(c.Request.Context(), uri.ID, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponse(result))
}

// DeleteResource handles POST /groups/delete
func (h *GroupHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.groupService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /groups/purge
func (h *GroupHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.groupService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /groups/restore
func (h *GroupHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.groupService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /groups
func (h *GroupHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.groupService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToGroupResponseList(listResult.Items), listResult.Total, pagination)
}
