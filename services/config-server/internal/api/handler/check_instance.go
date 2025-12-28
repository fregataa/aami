package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// CheckInstanceHandler handles HTTP requests for check instances
type CheckInstanceHandler struct {
	instanceService *service.CheckInstanceService
}

// NewCheckInstanceHandler creates a new CheckInstanceHandler
func NewCheckInstanceHandler(instanceService *service.CheckInstanceService) *CheckInstanceHandler {
	return &CheckInstanceHandler{
		instanceService: instanceService,
	}
}

// Create handles POST /check-instances
func (h *CheckInstanceHandler) Create(c *gin.Context) {
	var req dto.CreateCheckInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	instance, err := h.instanceService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToCheckInstanceResponse(instance))
}

// GetByID handles GET /check-instances/:id
func (h *CheckInstanceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	instance, err := h.instanceService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponse(instance))
}

// GetByTemplateID handles GET /check-instances/template/:templateId
func (h *CheckInstanceHandler) GetByTemplateID(c *gin.Context) {
	templateID := c.Param("templateId")

	instances, err := h.instanceService.GetByTemplateID(c.Request.Context(), templateID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetGlobalInstances handles GET /check-instances/global
func (h *CheckInstanceHandler) GetGlobalInstances(c *gin.Context) {
	instances, err := h.instanceService.GetGlobalInstances(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetByNamespaceID handles GET /check-instances/namespace/:namespaceId
func (h *CheckInstanceHandler) GetByNamespaceID(c *gin.Context) {
	namespaceID := c.Param("namespaceId")

	instances, err := h.instanceService.GetByNamespaceID(c.Request.Context(), namespaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetByGroupID handles GET /check-instances/group/:groupId
func (h *CheckInstanceHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("groupId")

	instances, err := h.instanceService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetEffectiveChecksByNamespace handles GET /check-instances/effective/namespace/:namespaceId
func (h *CheckInstanceHandler) GetEffectiveChecksByNamespace(c *gin.Context) {
	namespaceID := c.Param("namespaceId")

	instances, err := h.instanceService.GetEffectiveChecksByNamespace(c.Request.Context(), namespaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetEffectiveChecksByGroup handles GET /check-instances/effective/group/:namespaceId/:groupId
func (h *CheckInstanceHandler) GetEffectiveChecksByGroup(c *gin.Context) {
	namespaceID := c.Param("namespaceId")
	groupID := c.Param("groupId")

	instances, err := h.instanceService.GetEffectiveChecksByGroup(c.Request.Context(), namespaceID, groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}

// GetEffectiveChecksByTargetID handles GET /checks/target/:targetId
// This endpoint is used by nodes to get their check configurations
func (h *CheckInstanceHandler) GetEffectiveChecksByTargetID(c *gin.Context) {
	targetID := c.Param("targetId")

	checks, err := h.instanceService.GetEffectiveChecksByTargetID(c.Request.Context(), targetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToEffectiveCheckResponseList(checks))
}

// Update handles PUT /check-instances/:id
func (h *CheckInstanceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCheckInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	instance, err := h.instanceService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponse(instance))
}

// DeleteResource handles POST /check-instances/delete
func (h *CheckInstanceHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.instanceService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /check-instances/purge
func (h *CheckInstanceHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.instanceService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /check-instances/restore
func (h *CheckInstanceHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.instanceService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /check-instances
func (h *CheckInstanceHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	instances, total, err := h.instanceService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToCheckInstanceResponseList(instances), total, pagination)
}

// ListActive handles GET /check-instances/active
func (h *CheckInstanceHandler) ListActive(c *gin.Context) {
	instances, err := h.instanceService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckInstanceResponseList(instances))
}
