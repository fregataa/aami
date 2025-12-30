package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	defaultsLoader *service.DefaultsLoaderService
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(defaultsLoader *service.DefaultsLoaderService) *AdminHandler {
	return &AdminHandler{
		defaultsLoader: defaultsLoader,
	}
}

// Seed handles POST /api/v1/admin/seed
// @Summary Seed default templates
// @Description Load default alert templates and script templates into the database
// @Tags admin
// @Accept json
// @Produce json
// @Param request body dto.SeedRequest false "Seed options"
// @Success 200 {object} dto.SeedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/seed [post]
func (h *AdminHandler) Seed(c *gin.Context) {
	var req dto.SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use defaults (force=false, dry_run=false)
		req = dto.SeedRequest{
			Force:  false,
			DryRun: false,
		}
	}

	result, err := h.defaultsLoader.LoadAll(c.Request.Context(), req.Force, req.DryRun)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "seed_failed",
		})
		return
	}

	response := dto.SeedResponse{
		AlertTemplates: dto.SeedStats{
			Created: result.AlertTemplatesCreated,
			Updated: result.AlertTemplatesUpdated,
			Skipped: result.AlertTemplatesSkipped,
		},
		ScriptTemplates: dto.SeedStats{
			Created: result.ScriptTemplatesCreated,
			Updated: result.ScriptTemplatesUpdated,
			Skipped: result.ScriptTemplatesSkipped,
		},
		Errors: result.Errors,
	}

	c.JSON(http.StatusOK, response)
}
