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

	token, err := h.tokenService.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToBootstrapTokenResponse(token))
}

// GetByID handles GET /bootstrap-tokens/:id
func (h *BootstrapTokenHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	token, err := h.tokenService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(token))
}

// GetByToken handles GET /bootstrap-tokens/token/:token
func (h *BootstrapTokenHandler) GetByToken(c *gin.Context) {
	tokenStr := c.Param("token")

	token, err := h.tokenService.GetByToken(c.Request.Context(), tokenStr)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(token))
}

// ValidateAndUse handles POST /bootstrap-tokens/validate
func (h *BootstrapTokenHandler) ValidateAndUse(c *gin.Context) {
	var req dto.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	token, err := h.tokenService.ValidateAndUse(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(token))
}

// RegisterNode handles POST /bootstrap-tokens/register
func (h *BootstrapTokenHandler) RegisterNode(c *gin.Context) {
	var req dto.BootstrapRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, domainerrors.NewBindingError(err))
		return
	}

	target, token, err := h.tokenService.RegisterNode(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := dto.BootstrapRegisterResponse{
		Target:        dto.ToTargetResponse(target),
		TokenUsage:    token.Uses,
		RemainingUses: token.RemainingUses(),
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

	token, err := h.tokenService.Update(c.Request.Context(), id, req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBootstrapTokenResponse(token))
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

	tokens, total, err := h.tokenService.List(c.Request.Context(), pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	respondList(c, dto.ToBootstrapTokenResponseList(tokens), total, pagination)
}
