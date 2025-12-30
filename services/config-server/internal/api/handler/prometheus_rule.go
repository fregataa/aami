package handler

import (
	"net/http"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/service"
	"github.com/gin-gonic/gin"
)

// PrometheusRuleHandler handles HTTP requests for Prometheus rule management
type PrometheusRuleHandler struct {
	ruleGenerator    *service.PrometheusRuleGenerator
	fileManager      *prometheus.RuleFileManager
	prometheusClient *prometheus.PrometheusClient
	alertRuleService *service.AlertRuleService
}

// NewPrometheusRuleHandler creates a new PrometheusRuleHandler
func NewPrometheusRuleHandler(
	ruleGenerator *service.PrometheusRuleGenerator,
	fileManager *prometheus.RuleFileManager,
	prometheusClient *prometheus.PrometheusClient,
) *PrometheusRuleHandler {
	return &PrometheusRuleHandler{
		ruleGenerator:    ruleGenerator,
		fileManager:      fileManager,
		prometheusClient: prometheusClient,
	}
}

// NewPrometheusRuleHandlerWithAlertService creates a new PrometheusRuleHandler with alert rule service
func NewPrometheusRuleHandlerWithAlertService(
	ruleGenerator *service.PrometheusRuleGenerator,
	fileManager *prometheus.RuleFileManager,
	prometheusClient *prometheus.PrometheusClient,
	alertRuleService *service.AlertRuleService,
) *PrometheusRuleHandler {
	return &PrometheusRuleHandler{
		ruleGenerator:    ruleGenerator,
		fileManager:      fileManager,
		prometheusClient: prometheusClient,
		alertRuleService: alertRuleService,
	}
}

// RegenerateAllRules handles POST /api/v1/prometheus/rules/regenerate
func (h *PrometheusRuleHandler) RegenerateAllRules(c *gin.Context) {
	if h.ruleGenerator == nil {
		respondError(c, domainerrors.NewValidationError("prometheus", "Prometheus rule generator not configured"))
		return
	}

	if err := h.ruleGenerator.GenerateAllRules(c.Request.Context()); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.RegenerateRulesResponse{
		Message: "All rules regenerated successfully",
		Success: true,
	})
}

// RegenerateGroupRules handles POST /api/v1/prometheus/rules/regenerate/:group_id
func (h *PrometheusRuleHandler) RegenerateGroupRules(c *gin.Context) {
	if h.ruleGenerator == nil {
		respondError(c, domainerrors.NewValidationError("prometheus", "Prometheus rule generator not configured"))
		return
	}

	groupID := c.Param("group_id")
	if groupID == "" {
		respondError(c, domainerrors.NewValidationError("group_id", "group_id is required"))
		return
	}

	if err := h.ruleGenerator.GenerateRulesForGroup(c.Request.Context(), groupID); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.RegenerateRulesResponse{
		Message:       "Rules regenerated for group",
		GroupsUpdated: 1,
		Success:       true,
	})
}

// ListRuleFiles handles GET /api/v1/prometheus/rules/files
func (h *PrometheusRuleHandler) ListRuleFiles(c *gin.Context) {
	if h.fileManager == nil {
		respondError(c, domainerrors.NewValidationError("prometheus", "Prometheus file manager not configured"))
		return
	}

	groupIDs, err := h.fileManager.ListRuleFiles()
	if err != nil {
		respondError(c, err)
		return
	}

	files := make([]dto.RuleFileInfo, len(groupIDs))
	for i, groupID := range groupIDs {
		files[i] = dto.RuleFileInfo{
			GroupID:  groupID,
			FilePath: h.fileManager.GetFilePath(groupID),
		}
	}

	c.JSON(http.StatusOK, dto.ListRuleFilesResponse{
		Files: files,
		Total: len(files),
	})
}

// ReloadPrometheus handles POST /api/v1/prometheus/reload
func (h *PrometheusRuleHandler) ReloadPrometheus(c *gin.Context) {
	if h.prometheusClient == nil {
		respondError(c, domainerrors.NewValidationError("prometheus", "Prometheus client not configured"))
		return
	}

	if err := h.prometheusClient.Reload(c.Request.Context()); err != nil {
		respondError(c, err)
		return
	}

	// Check health after reload
	healthy := h.prometheusClient.HealthCheck(c.Request.Context()) == nil

	c.JSON(http.StatusOK, dto.ReloadPrometheusResponse{
		Message: "Prometheus reload triggered successfully",
		Success: true,
		Healthy: healthy,
	})
}

// GetStatus handles GET /api/v1/prometheus/status
func (h *PrometheusRuleHandler) GetStatus(c *gin.Context) {
	if h.prometheusClient == nil {
		c.JSON(http.StatusOK, dto.PrometheusStatusResponse{
			Reachable: false,
			Healthy:   false,
		})
		return
	}

	reachable := h.prometheusClient.IsReachable(c.Request.Context())
	healthy := h.prometheusClient.HealthCheck(c.Request.Context()) == nil

	var status map[string]interface{}
	if reachable {
		status, _ = h.prometheusClient.GetStatus(c.Request.Context())
	}

	c.JSON(http.StatusOK, dto.PrometheusStatusResponse{
		Reachable: reachable,
		Healthy:   healthy,
		Status:    status,
	})
}

// GetEffectiveRulesByTarget handles GET /api/v1/prometheus/rules/effective/:target_id
func (h *PrometheusRuleHandler) GetEffectiveRulesByTarget(c *gin.Context) {
	if h.alertRuleService == nil {
		respondError(c, domainerrors.NewValidationError("prometheus", "Alert rule service not configured"))
		return
	}

	targetID := c.Param("target_id")
	if targetID == "" {
		respondError(c, domainerrors.NewValidationError("target_id", "target_id is required"))
		return
	}

	result, err := h.alertRuleService.GetEffectiveRulesByTargetID(c.Request.Context(), targetID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toEffectiveAlertRulesResponse(result))
}

// toEffectiveAlertRulesResponse converts domain result to DTO response
func toEffectiveAlertRulesResponse(result *domain.EffectiveAlertRulesResult) dto.EffectiveAlertRulesResponse {
	rules := make([]dto.EffectiveAlertRule, len(result.Rules))

	for i, rule := range result.Rules {
		// Build config map from AlertRuleConfig
		configMap := rule.Config.ToMap()

		var sourceName string
		var sourceID string
		if rule.SourceGroup != nil {
			sourceName = rule.SourceGroup.Name
			sourceID = rule.SourceGroup.ID
		}

		rules[i] = dto.EffectiveAlertRule{
			ID:          rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    string(rule.Severity),
			Query:       rule.RenderedQuery,
			ForDuration: rule.Config.ForDuration,
			Labels:      rule.Config.Labels,
			Annotations: rule.Config.Annotations,
			Config:      configMap,
			Source:      "group",
			SourceID:    sourceID,
			SourceName:  sourceName,
		}
	}

	return dto.EffectiveAlertRulesResponse{
		TargetID: result.Target.ID,
		Hostname: result.Target.Hostname,
		Rules:    rules,
		Total:    len(rules),
	}
}
