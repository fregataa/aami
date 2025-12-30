package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// ExporterHandler handles HTTP requests for exporters
type ExporterHandler struct {
	exporterService *service.ExporterService
}

// NewExporterHandler creates a new ExporterHandler
func NewExporterHandler(exporterService *service.ExporterService) *ExporterHandler {
	return &ExporterHandler{
		exporterService: exporterService,
	}
}

// Create handles POST /exporters
func (h *ExporterHandler) Create(c *gin.Context) {
	var req dto.CreateExporterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.exporterService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToExporterResponse(result))
}

// GetByID handles GET /exporters/:id
func (h *ExporterHandler) GetByID(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	result, err := h.exporterService.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponse(result))
}

// Update handles PUT /exporters/:id
func (h *ExporterHandler) Update(c *gin.Context) {
	var uri dto.IDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("id", err.Error()))
		return
	}

	var req dto.UpdateExporterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.exporterService.Update(c.Request.Context(), uri.ID, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponse(result))
}

// DeleteResource handles POST /exporters/delete
func (h *ExporterHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.exporterService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /exporters/purge
func (h *ExporterHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.exporterService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /exporters/restore
func (h *ExporterHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.exporterService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /exporters
func (h *ExporterHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.exporterService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToExporterResponseList(listResult.Items), listResult.Total, pagination)
}

// GetByTargetID handles GET /exporters/target/:target_id
func (h *ExporterHandler) GetByTargetID(c *gin.Context) {
	var uri dto.TargetIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("target_id", err.Error()))
		return
	}

	results, err := h.exporterService.GetByTargetID(c.Request.Context(), uri.TargetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponseList(results))
}

// GetByType handles GET /exporters/type/:type
func (h *ExporterHandler) GetByType(c *gin.Context) {
	var uri dto.TypeUri
	if err := c.ShouldBindUri(&uri); err != nil {
		respondError(c, domainerrors.NewValidationError("type", err.Error()))
		return
	}

	exporterType := domain.ExporterType(uri.Type)
	exporters, err := h.exporterService.GetByType(c.Request.Context(), exporterType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponseList(exporters))
}
