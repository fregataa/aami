package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// NamespaceHandler handles HTTP requests for namespaces
type NamespaceHandler struct {
	namespaceService *service.NamespaceService
}

// NewNamespaceHandler creates a new NamespaceHandler
func NewNamespaceHandler(namespaceService *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: namespaceService,
	}
}

// Create handles POST /namespaces
func (h *NamespaceHandler) Create(c *gin.Context) {
	var req dto.CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	namespace, err := h.namespaceService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToNamespaceResponse(namespace))
}

// GetByID handles GET /namespaces/:id
func (h *NamespaceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	namespace, err := h.namespaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToNamespaceResponse(namespace))
}

// GetByName handles GET /namespaces/name/:name
func (h *NamespaceHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	namespace, err := h.namespaceService.GetByName(c.Request.Context(), name)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToNamespaceResponse(namespace))
}

// Update handles PUT /namespaces/:id
func (h *NamespaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	namespace, err := h.namespaceService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToNamespaceResponse(namespace))
}

// DeleteResource handles POST /namespaces/delete
func (h *NamespaceHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.namespaceService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /namespaces/purge
func (h *NamespaceHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.namespaceService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /namespaces/restore
func (h *NamespaceHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if err := h.namespaceService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /namespaces
func (h *NamespaceHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	namespaces, total, err := h.namespaceService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToNamespaceResponseList(namespaces), total, pagination)
}

// GetAll handles GET /namespaces/all
func (h *NamespaceHandler) GetAll(c *gin.Context) {
	namespaces, err := h.namespaceService.GetAll(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToNamespaceResponseList(namespaces))
}

// GetStats handles GET /namespaces/:id/stats
func (h *NamespaceHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.namespaceService.GetStats(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllStats handles GET /namespaces/stats
func (h *NamespaceHandler) GetAllStats(c *gin.Context) {
	stats, err := h.namespaceService.GetAllStats(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
