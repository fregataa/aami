package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// CheckTemplateHandler handles HTTP requests for check templates
type CheckTemplateHandler struct {
	templateService *service.CheckTemplateService
}

// NewCheckTemplateHandler creates a new CheckTemplateHandler
func NewCheckTemplateHandler(templateService *service.CheckTemplateService) *CheckTemplateHandler {
	return &CheckTemplateHandler{
		templateService: templateService,
	}
}

// Create handles POST /check-templates
func (h *CheckTemplateHandler) Create(c *gin.Context) {
	var req dto.CreateCheckTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	template, err := h.templateService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToCheckTemplateResponse(template))
}

// GetByID handles GET /check-templates/:id
func (h *CheckTemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckTemplateResponse(template))
}

// GetByName handles GET /check-templates/name/:name
func (h *CheckTemplateHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	template, err := h.templateService.GetByName(c.Request.Context(), name)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckTemplateResponse(template))
}

// GetByCheckType handles GET /check-templates/type/:checkType
func (h *CheckTemplateHandler) GetByCheckType(c *gin.Context) {
	checkType := c.Param("checkType")

	templates, err := h.templateService.GetByCheckType(c.Request.Context(), checkType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckTemplateSummaryResponseList(templates))
}

// Update handles PUT /check-templates/:id
func (h *CheckTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCheckTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	template, err := h.templateService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckTemplateResponse(template))
}

// DeleteResource handles POST /check-templates/delete
func (h *CheckTemplateHandler) DeleteResource(c *gin.Context) {
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

// PurgeResource handles POST /check-templates/purge
func (h *CheckTemplateHandler) PurgeResource(c *gin.Context) {
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

// RestoreResource handles POST /check-templates/restore
func (h *CheckTemplateHandler) RestoreResource(c *gin.Context) {
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

// List handles GET /check-templates
func (h *CheckTemplateHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	templates, total, err := h.templateService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToCheckTemplateSummaryResponseList(templates), total, pagination)
}

// ListActive handles GET /check-templates/active
func (h *CheckTemplateHandler) ListActive(c *gin.Context) {
	templates, err := h.templateService.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCheckTemplateSummaryResponseList(templates))
}

// VerifyHash handles GET /check-templates/:id/verify-hash
func (h *CheckTemplateHandler) VerifyHash(c *gin.Context) {
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
