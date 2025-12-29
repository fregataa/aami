package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// AlertTemplateHandler handles HTTP requests for alert templates
type AlertTemplateHandler struct {
	templateService *service.AlertTemplateService
}

// NewAlertTemplateHandler creates a new AlertTemplateHandler
func NewAlertTemplateHandler(templateService *service.AlertTemplateService) *AlertTemplateHandler {
	return &AlertTemplateHandler{
		templateService: templateService,
	}
}

// Create handles POST /alert-templates
func (h *AlertTemplateHandler) Create(c *gin.Context) {
	var req dto.CreateAlertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.templateService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAlertTemplateResponse(result))
}

// GetByID handles GET /alert-templates/:id
func (h *AlertTemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponse(result))
}

// Update handles PUT /alert-templates/:id
func (h *AlertTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAlertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.templateService.Update(c.Request.Context(), id, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponse(result))
}

// DeleteResource handles POST /alert-templates/delete
func (h *AlertTemplateHandler) DeleteResource(c *gin.Context) {
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

// PurgeResource handles POST /alert-templates/purge
func (h *AlertTemplateHandler) PurgeResource(c *gin.Context) {
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

// RestoreResource handles POST /alert-templates/restore
func (h *AlertTemplateHandler) RestoreResource(c *gin.Context) {
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

// List handles GET /alert-templates
func (h *AlertTemplateHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.templateService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToAlertTemplateResponseList(listResult.Items), listResult.Total, pagination)
}

// GetBySeverity handles GET /alert-templates/severity/:severity
func (h *AlertTemplateHandler) GetBySeverity(c *gin.Context) {
	severity := domain.AlertSeverity(c.Param("severity"))

	results, err := h.templateService.GetBySeverity(c.Request.Context(), severity)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponseList(results))
}

// AlertRuleHandler handles HTTP requests for alert rules
type AlertRuleHandler struct {
	ruleService *service.AlertRuleService
}

// NewAlertRuleHandler creates a new AlertRuleHandler
func NewAlertRuleHandler(ruleService *service.AlertRuleService) *AlertRuleHandler {
	return &AlertRuleHandler{
		ruleService: ruleService,
	}
}

// Create handles POST /alert-rules
func (h *AlertRuleHandler) Create(c *gin.Context) {
	// Parse raw JSON to determine creation mode
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	var result action.AlertRuleResult
	var err error

	// Check if template_id exists to determine creation mode
	if templateID, hasTemplate := rawReq["template_id"].(string); hasTemplate && templateID != "" {
		// Create from template
		var req dto.CreateAlertRuleFromTemplateRequest
		if err := mapToStruct(rawReq, &req); err != nil {
			respondError(c, domainerrors.NewBindingError(err))
			return
		}
		result, err = h.ruleService.CreateFromTemplate(c.Request.Context(), req.ToAction())
	} else {
		// Create directly
		var req dto.CreateAlertRuleDirectRequest
		if err := mapToStruct(rawReq, &req); err != nil {
			respondError(c, domainerrors.NewBindingError(err))
			return
		}
		result, err = h.ruleService.CreateDirect(c.Request.Context(), req.ToAction())
	}

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAlertRuleResponse(result))
}

// GetByID handles GET /alert-rules/:id
func (h *AlertRuleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.ruleService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponse(result))
}

// Update handles PUT /alert-rules/:id
func (h *AlertRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.ruleService.Update(c.Request.Context(), id, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponse(result))
}

// DeleteResource handles POST /alert-rules/delete
func (h *AlertRuleHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.ruleService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /alert-rules/purge
func (h *AlertRuleHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.ruleService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /alert-rules/restore
func (h *AlertRuleHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.ruleService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /alert-rules
func (h *AlertRuleHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.ruleService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToAlertRuleResponseList(listResult.Items), listResult.Total, pagination)
}

// GetByGroupID handles GET /alert-rules/group/:group_id
func (h *AlertRuleHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("group_id")

	results, err := h.ruleService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponseList(results))
}

// GetByTemplateID handles GET /alert-rules/template/:template_id
func (h *AlertRuleHandler) GetByTemplateID(c *gin.Context) {
	templateID := c.Param("template_id")

	results, err := h.ruleService.GetByTemplateID(c.Request.Context(), templateID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponseList(results))
}

// mapToStruct converts a map to a struct using JSON encoding/decoding
func mapToStruct(m map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
