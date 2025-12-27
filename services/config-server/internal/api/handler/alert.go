package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	template, err := h.templateService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAlertTemplateResponse(template))
}

// GetByID handles GET /alert-templates/:id
func (h *AlertTemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponse(template))
}

// Update handles PUT /alert-templates/:id
func (h *AlertTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAlertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	template, err := h.templateService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponse(template))
}

// DeleteResource handles POST /alert-templates/delete
func (h *AlertTemplateHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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

	templates, total, err := h.templateService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToAlertTemplateResponseList(templates), total, pagination)
}

// GetBySeverity handles GET /alert-templates/severity/:severity
func (h *AlertTemplateHandler) GetBySeverity(c *gin.Context) {
	severity := domain.AlertSeverity(c.Param("severity"))

	templates, err := h.templateService.GetBySeverity(c.Request.Context(), severity)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertTemplateResponseList(templates))
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
	var req dto.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	rule, err := h.ruleService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAlertRuleResponse(rule))
}

// GetByID handles GET /alert-rules/:id
func (h *AlertRuleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.ruleService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponse(rule))
}

// Update handles PUT /alert-rules/:id
func (h *AlertRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	rule, err := h.ruleService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponse(rule))
}

// DeleteResource handles POST /alert-rules/delete
func (h *AlertRuleHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
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

	rules, total, err := h.ruleService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToAlertRuleResponseList(rules), total, pagination)
}

// GetByGroupID handles GET /alert-rules/group/:group_id
func (h *AlertRuleHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("group_id")

	rules, err := h.ruleService.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponseList(rules))
}

// GetByTemplateID handles GET /alert-rules/template/:template_id
func (h *AlertRuleHandler) GetByTemplateID(c *gin.Context) {
	templateID := c.Param("template_id")

	rules, err := h.ruleService.GetByTemplateID(c.Request.Context(), templateID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAlertRuleResponseList(rules))
}
