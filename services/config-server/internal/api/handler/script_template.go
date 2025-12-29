package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// ScriptTemplateHandler handles HTTP requests for script templates
type ScriptTemplateHandler struct {
	templateService *service.ScriptTemplateService
}

// NewScriptTemplateHandler creates a new ScriptTemplateHandler
func NewScriptTemplateHandler(templateService *service.ScriptTemplateService) *ScriptTemplateHandler {
	return &ScriptTemplateHandler{
		templateService: templateService,
	}
}

// Create handles POST /script-templates
func (h *ScriptTemplateHandler) Create(c *gin.Context) {
	var req dto.CreateScriptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.templateService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToScriptTemplateResponse(result))
}

// GetByID handles GET /script-templates/:id
func (h *ScriptTemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptTemplateResponse(result))
}

// GetByName handles GET /script-templates/name/:name
func (h *ScriptTemplateHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	result, err := h.templateService.GetByName(c.Request.Context(), name)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptTemplateResponse(result))
}

// GetByScriptType handles GET /script-templates/type/:scriptType
func (h *ScriptTemplateHandler) GetByScriptType(c *gin.Context) {
	scriptType := c.Param("scriptType")

	results, err := h.templateService.GetByScriptType(c.Request.Context(), scriptType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptTemplateSummaryResponseList(results))
}

// Update handles PUT /script-templates/:id
func (h *ScriptTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateScriptTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.templateService.Update(c.Request.Context(), id, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptTemplateResponse(result))
}

// DeleteResource handles POST /script-templates/delete
func (h *ScriptTemplateHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.templateService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /script-templates/purge
func (h *ScriptTemplateHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.templateService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /script-templates/restore
func (h *ScriptTemplateHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.templateService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /script-templates
func (h *ScriptTemplateHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.templateService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToScriptTemplateSummaryResponseList(listResult.Items), listResult.Total, pagination)
}

// ListActive handles GET /script-templates/active
func (h *ScriptTemplateHandler) ListActive(c *gin.Context) {
	results, err := h.templateService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToScriptTemplateSummaryResponseList(results))
}

// VerifyHash handles GET /script-templates/:id/verify-hash
func (h *ScriptTemplateHandler) VerifyHash(c *gin.Context) {
	id := c.Param("id")

	valid, err := h.templateService.VerifyHash(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template_id": id,
		"hash_valid":  valid,
	})
}
