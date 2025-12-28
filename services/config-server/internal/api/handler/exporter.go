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

	exporter, err := h.exporterService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToExporterResponse(exporter))
}

// GetByID handles GET /exporters/:id
func (h *ExporterHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	exporter, err := h.exporterService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponse(exporter))
}

// Update handles PUT /exporters/:id
func (h *ExporterHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateExporterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	exporter, err := h.exporterService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponse(exporter))
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

	exporters, total, err := h.exporterService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToExporterResponseList(exporters), total, pagination)
}

// GetByTargetID handles GET /exporters/target/:target_id
func (h *ExporterHandler) GetByTargetID(c *gin.Context) {
	targetID := c.Param("target_id")

	exporters, err := h.exporterService.GetByTargetID(c.Request.Context(), targetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponseList(exporters))
}

// GetByType handles GET /exporters/type/:type
func (h *ExporterHandler) GetByType(c *gin.Context) {
	exporterType := domain.ExporterType(c.Param("type"))

	exporters, err := h.exporterService.GetByType(c.Request.Context(), exporterType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToExporterResponseList(exporters))
}
