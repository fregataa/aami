package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
)

// BootstrapTokenHandler handles HTTP requests for bootstrap tokens
type BootstrapTokenHandler struct {
	tokenService *service.BootstrapTokenService
}

// NewBootstrapTokenHandler creates a new BootstrapTokenHandler
func NewBootstrapTokenHandler(tokenService *service.BootstrapTokenService) *BootstrapTokenHandler {
	return &BootstrapTokenHandler{
		tokenService: tokenService,
	}
}

// Create handles POST /bootstrap-tokens
func (h *BootstrapTokenHandler) Create(c *gin.Context) {
	var req dto.CreateBootstrapTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.tokenService.Create(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToBootstrapTokenResponse(result))
}

// GetByID handles GET /bootstrap-tokens/:id
func (h *BootstrapTokenHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.tokenService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(result))
}

// GetByToken handles GET /bootstrap-tokens/token/:token
func (h *BootstrapTokenHandler) GetByToken(c *gin.Context) {
	tokenStr := c.Param("token")

	result, err := h.tokenService.GetByToken(c.Request.Context(), tokenStr)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(result))
}

// ValidateAndUse handles POST /bootstrap-tokens/validate
func (h *BootstrapTokenHandler) ValidateAndUse(c *gin.Context) {
	var req dto.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.tokenService.ValidateAndUse(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(result))
}

// RegisterNode handles POST /bootstrap-tokens/register
func (h *BootstrapTokenHandler) RegisterNode(c *gin.Context) {
	var req dto.BootstrapRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	targetResult, tokenResult, err := h.tokenService.RegisterNode(c.Request.Context(), req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	resp := dto.BootstrapRegisterResponse{
		Target:        dto.ToTargetResponse(targetResult),
		TokenUsage:    tokenResult.Uses,
		RemainingUses: tokenResult.MaxUses - tokenResult.Uses,
	}

	c.JSON(http.StatusCreated, resp)
}

// Update handles PUT /bootstrap-tokens/:id
func (h *BootstrapTokenHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateBootstrapTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	result, err := h.tokenService.Update(c.Request.Context(), id, req.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(result))
}

// DeleteResource handles POST /bootstrap-tokens/delete
func (h *BootstrapTokenHandler) DeleteResource(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.tokenService.Delete(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PurgeResource handles POST /bootstrap-tokens/purge
func (h *BootstrapTokenHandler) PurgeResource(c *gin.Context) {
	var req dto.PurgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.tokenService.Purge(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreResource handles POST /bootstrap-tokens/restore
func (h *BootstrapTokenHandler) RestoreResource(c *gin.Context) {
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	if err := h.tokenService.Restore(c.Request.Context(), req.ID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /bootstrap-tokens
func (h *BootstrapTokenHandler) List(c *gin.Context) {
	pagination := getPagination(c)

	listResult, err := h.tokenService.List(c.Request.Context(), pagination.ToAction())
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToBootstrapTokenResponseList(listResult.Items), listResult.Total, pagination)
}
