package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// ScriptPolicyHandler handles HTTP requests for check instances
type ScriptPolicyHandler struct {
	policyService *service.ScriptPolicyService
}

// NewScriptPolicyHandler creates a new ScriptPolicyHandler
func NewScriptPolicyHandler(policyService *service.ScriptPolicyService) *ScriptPolicyHandler {
	return &ScriptPolicyHandler{
		policyService: policyService,
	}
}

// Create handles POST /check-instances
func (h *ScriptPolicyHandler) Create(c *gin.Context) {
	var req dto.CreateScriptPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	instance, err := h.policyService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToScriptPolicyResponse(instance))
}

// GetByID handles GET /check-instances/:id
func (h *ScriptPolicyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	instance, err := h.policyService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponse(instance))
}

// GetByTemplateID handles GET /check-instances/template/:templateId
func (h *ScriptPolicyHandler) GetByTemplateID(c *gin.Context) {
	templateID := c.Param("templateId")

	instances, err := h.policyService.GetByTemplateID(c.Request.Context(), templateID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetGlobalInstances handles GET /check-instances/global
func (h *ScriptPolicyHandler) GetGlobalInstances(c *gin.Context) {
	instances, err := h.policyService.GetGlobalInstances(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetByNamespaceID handles GET /check-instances/namespace/:namespaceId
func (h *ScriptPolicyHandler) GetByNamespaceID(c *gin.Context) {
	namespaceID := c.Param("namespaceId")

	instances, err := h.policyService.GetByNamespaceID(c.Request.Context(), namespaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetByGroupID handles GET /check-instances/group/:groupId
func (h *ScriptPolicyHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("groupId")

	instances, err := h.policyService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetEffectiveChecksByNamespace handles GET /check-instances/effective/namespace/:namespaceId
func (h *ScriptPolicyHandler) GetEffectiveChecksByNamespace(c *gin.Context) {
	namespaceID := c.Param("namespaceId")

	instances, err := h.policyService.GetEffectiveChecksByNamespace(c.Request.Context(), namespaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetEffectiveChecksByGroup handles GET /check-instances/effective/group/:namespaceId/:groupId
func (h *ScriptPolicyHandler) GetEffectiveChecksByGroup(c *gin.Context) {
	namespaceID := c.Param("namespaceId")
	groupID := c.Param("groupId")

	instances, err := h.policyService.GetEffectiveChecksByGroup(c.Request.Context(), namespaceID, groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}

// GetEffectiveChecksByTargetID handles GET /checks/target/:targetId
// This endpoint is used by nodes to get their check configurations
func (h *ScriptPolicyHandler) GetEffectiveChecksByTargetID(c *gin.Context) {
	targetID := c.Param("targetId")

	checks, err := h.policyService.GetEffectiveChecksByTargetID(c.Request.Context(), targetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToEffectiveCheckResponseList(checks))
}

// Update handles PUT /check-instances/:id
func (h *ScriptPolicyHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateScriptPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	instance, err := h.policyService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponse(instance))
}

// DeleteResource handles POST /check-instances/delete
func (h *ScriptPolicyHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.policyService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /check-instances/purge
func (h *ScriptPolicyHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.policyService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /check-instances/restore
func (h *ScriptPolicyHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.policyService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /check-instances
func (h *ScriptPolicyHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	instances, total, err := h.policyService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToScriptPolicyResponseList(instances), total, pagination)
}

// ListActive handles GET /check-instances/active
func (h *ScriptPolicyHandler) ListActive(c *gin.Context) {
	instances, err := h.policyService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(instances))
}
