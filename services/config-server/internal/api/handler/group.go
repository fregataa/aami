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

	group, err := h.groupService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToGroupResponse(group))
}

// GetByID handles GET /groups/:id
func (h *GroupHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	group, err := h.groupService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponse(group))
}

// Update handles PUT /groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	group, err := h.groupService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponse(group))
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

	groups, total, err := h.groupService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToGroupResponseList(groups), total, pagination)
}

// GetByNamespaceID handles GET /groups/namespace/:namespace_id
func (h *GroupHandler) GetByNamespaceID(c *gin.Context) {
	namespaceID := c.Param("namespace_id")

	groups, err := h.groupService.GetByNamespaceID(c.Request.Context(), namespaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponseList(groups))
}

// GetChildren handles GET /groups/:id/children
func (h *GroupHandler) GetChildren(c *gin.Context) {
	id := c.Param("id")

	children, err := h.groupService.GetChildren(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponseList(children))
}

// GetAncestors handles GET /groups/:id/ancestors
func (h *GroupHandler) GetAncestors(c *gin.Context) {
	id := c.Param("id")

	ancestors, err := h.groupService.GetAncestors(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponseList(ancestors))
}
