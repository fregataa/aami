package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/action"
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
	// Parse raw JSON to determine creation mode
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	var result action.ScriptPolicyResult
	var err error

	// Check if template_id exists to determine creation mode
	if templateID, hasTemplate := rawReq["template_id"].(string); hasTemplate && templateID != "" {
		// Create from template
		var req dto.CreateScriptPolicyFromTemplateRequest
		if err := mapToStruct(rawReq, &req); err != nil {
			respondError(c, domainerrors.NewBindingError(err))
			return
		}
		result, err = h.policyService.CreateFromTemplate(c.Request.Context(), req.ToAction())
	} else {
		// Create directly
		var req dto.CreateScriptPolicyDirectRequest
		if err := mapToStruct(rawReq, &req); err != nil {
			respondError(c, domainerrors.NewBindingError(err))
			return
		}
		result, err = h.policyService.CreateDirect(c.Request.Context(), req.ToAction())
	}

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToScriptPolicyResponse(result))
}

// GetByID handles GET /check-instances/:id
func (h *ScriptPolicyHandler) GetByID(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	result, err := h.policyService.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponse(result))
}

// GetByTemplateID handles GET /check-instances/template/:templateId
func (h *ScriptPolicyHandler) GetByTemplateID(c *gin.Context) {
	var uri dto.TemplateIdUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("templateId", err.Error()))
		return
	}

	results, err := h.policyService.GetByTemplateID(c.Request.Context(), uri.TemplateID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(results))
}

// GetGlobalInstances handles GET /check-instances/global
func (h *ScriptPolicyHandler) GetGlobalInstances(c *gin.Context) {
	results, err := h.policyService.GetGlobalInstances(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(results))
}

// GetByGroupID handles GET /check-instances/group/:groupId
func (h *ScriptPolicyHandler) GetByGroupID(c *gin.Context) {
	var uri dto.GroupIdUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("groupId", err.Error()))
		return
	}

	results, err := h.policyService.GetByGroupID(c.Request.Context(), uri.GroupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(results))
}

// GetEffectiveChecksByGroup handles GET /check-instances/effective/group/:groupId
func (h *ScriptPolicyHandler) GetEffectiveChecksByGroup(c *gin.Context) {
	var uri dto.GroupIdUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("groupId", err.Error()))
		return
	}

	results, err := h.policyService.GetEffectiveChecksByGroup(c.Request.Context(), uri.GroupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(results))
}

// GetEffectiveChecksByTargetID handles GET /checks/target/:targetId
// This endpoint is used by nodes to get their check configurations
func (h *ScriptPolicyHandler) GetEffectiveChecksByTargetID(c *gin.Context) {
	var uri dto.TargetIdUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("targetId", err.Error()))
		return
	}

	checks, err := h.policyService.GetEffectiveChecksByTargetID(c.Request.Context(), uri.TargetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToEffectiveCheckResponseList(checks))
}

// GetEffectiveChecksByHostname handles GET /checks/target/hostname/:hostname
// This endpoint is used by node scripts that only know their hostname
func (h *ScriptPolicyHandler) GetEffectiveChecksByHostname(c *gin.Context) {
	var uri dto.HostnameUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("hostname", err.Error()))
		return
	}

	checks, err := h.policyService.GetEffectiveChecksByHostname(c.Request.Context(), uri.Hostname)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToEffectiveCheckResponseList(checks))
}

// Update handles PUT /check-instances/:id
func (h *ScriptPolicyHandler) Update(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	var req dto.UpdateScriptPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.policyService.Update(c.Request.Context(), uri.ID, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponse(result))
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

	listResult, err := h.policyService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToScriptPolicyResponseList(listResult.Items), listResult.Total, pagination)
}

// ListActive handles GET /check-instances/active
func (h *ScriptPolicyHandler) ListActive(c *gin.Context) {
	results, err := h.policyService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptPolicyResponseList(results))
}
