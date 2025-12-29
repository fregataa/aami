package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/repository"
	"gopkg.in/yaml.v3"
)

// PrometheusRuleGenerator generates Prometheus rule files from AlertRules
type PrometheusRuleGenerator struct {
	alertRuleRepo repository.AlertRuleRepository
	groupRepo     repository.GroupRepository
	fileManager   *prometheus.RuleFileManager
	logger        *slog.Logger
}

// NewPrometheusRuleGenerator creates a new PrometheusRuleGenerator
func NewPrometheusRuleGenerator(
	alertRuleRepo repository.AlertRuleRepository,
	groupRepo repository.GroupRepository,
	fileManager *prometheus.RuleFileManager,
	logger *slog.Logger,
) *PrometheusRuleGenerator {
	return &PrometheusRuleGenerator{
		alertRuleRepo: alertRuleRepo,
		groupRepo:     groupRepo,
		fileManager:   fileManager,
		logger:        logger,
	}
}

// GenerateRulesForGroup generates Prometheus rules for a specific group
func (g *PrometheusRuleGenerator) GenerateRulesForGroup(ctx context.Context, groupID string) error {
	g.logger.Info("Generating Prometheus rules for group", "group_id", groupID)

	// Get group information
	group, err := g.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("%w: %s", domainerrors.ErrGroupRetrievalFailed, err)
	}

	// Step 1: Get enabled rules for the group
	enabledRules, err := g.getEnabledRulesForGroup(ctx, groupID)
	if err != nil {
		return err
	}

	// If no enabled rules, delete the file
	if len(enabledRules) == 0 {
		g.logger.Info("No enabled rules for group, deleting rule file", "group_id", groupID)
		return g.DeleteRulesForGroup(ctx, groupID)
	}

	// Step 2: Convert to Prometheus rules
	prometheusRules, err := g.convertRulesToPrometheusRules(enabledRules)
	if err != nil {
		return err
	}

	// Step 3: Marshal rules to YAML
	yamlData, err := g.marshalRulesToYAML(groupID, group.Name, prometheusRules)
	if err != nil {
		return err
	}

	// Step 4: Write rules to file
	if err := g.writeRuleFile(groupID, yamlData); err != nil {
		return err
	}

	g.logger.Info("Successfully generated Prometheus rules",
		"group_id", groupID,
		"rule_count", len(prometheusRules))

	return nil
}

// getEnabledRulesForGroup retrieves enabled alert rules for a specific group
func (g *PrometheusRuleGenerator) getEnabledRulesForGroup(ctx context.Context, groupID string) ([]domain.AlertRule, error) {
	// Get all alert rules for this group
	alertRules, err := g.alertRuleRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrAlertRulesListFailed, err)
	}

	// Filter only enabled rules
	var enabledRules []domain.AlertRule
	for _, rule := range alertRules {
		if rule.Enabled && rule.DeletedAt == nil {
			enabledRules = append(enabledRules, rule)
		}
	}

	return enabledRules, nil
}

// convertRulesToPrometheusRules converts AlertRules to PrometheusRules
func (g *PrometheusRuleGenerator) convertRulesToPrometheusRules(alertRules []domain.AlertRule) ([]domain.PrometheusRule, error) {
	prometheusRules := make([]domain.PrometheusRule, 0, len(alertRules))

	for _, rule := range alertRules {
		promRule, err := g.convertToPrometheusRule(&rule)
		if err != nil {
			g.logger.Error("Failed to convert rule", "rule_id", rule.ID, "error", err)
			continue
		}
		prometheusRules = append(prometheusRules, *promRule)
	}

	if len(prometheusRules) == 0 {
		return nil, domainerrors.ErrNoValidRules
	}

	return prometheusRules, nil
}

// marshalRulesToYAML marshals Prometheus rules to YAML format
func (g *PrometheusRuleGenerator) marshalRulesToYAML(groupID, groupName string, rules []domain.PrometheusRule) ([]byte, error) {
	// Create Prometheus rule group
	ruleGroup := domain.NewPrometheusRuleGroup(
		fmt.Sprintf("group_%s_%s", groupName, groupID),
		rules,
	)

	// Create rule file
	ruleFile := domain.NewPrometheusRuleFile([]domain.PrometheusRuleGroup{*ruleGroup})

	// Marshal to YAML
	yamlData, err := yaml.Marshal(ruleFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrRuleMarshalingFailed, err)
	}

	return yamlData, nil
}

// writeRuleFile writes YAML data to a rule file
func (g *PrometheusRuleGenerator) writeRuleFile(groupID string, yamlData []byte) error {
	// Write to file using file manager (with atomic write, validation, and backup)
	if err := g.fileManager.WriteRuleFile(groupID, yamlData); err != nil {
		return fmt.Errorf("%w: %s", domainerrors.ErrRuleWriteFailed, err)
	}

	return nil
}

// GenerateAllRules generates Prometheus rules for all groups
func (g *PrometheusRuleGenerator) GenerateAllRules(ctx context.Context) error {
	g.logger.Info("Generating Prometheus rules for all groups")

	// Get all alert rules
	allRules, _, err := g.alertRuleRepo.List(ctx, 1, 10000) // Large limit to get all
	if err != nil {
		return fmt.Errorf("%w: %s", domainerrors.ErrAlertRulesListFailed, err)
	}

	// Group by group_id
	rulesByGroup := make(map[string][]domain.AlertRule)
	for _, rule := range allRules {
		if rule.Enabled && rule.DeletedAt == nil {
			rulesByGroup[rule.GroupID] = append(rulesByGroup[rule.GroupID], rule)
		}
	}

	// Generate for each group
	successCount := 0
	errorCount := 0
	for groupID := range rulesByGroup {
		if err := g.GenerateRulesForGroup(ctx, groupID); err != nil {
			g.logger.Error("Failed to generate rules for group", "group_id", groupID, "error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	g.logger.Info("Completed generating all rules",
		"success", successCount,
		"errors", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%w: %d groups", domainerrors.ErrBulkGenerationFailed, errorCount)
	}

	return nil
}

// DeleteRulesForGroup deletes the Prometheus rule file for a specific group
func (g *PrometheusRuleGenerator) DeleteRulesForGroup(ctx context.Context, groupID string) error {
	// Delete file using file manager (with backup if enabled)
	if err := g.fileManager.DeleteRuleFile(groupID); err != nil {
		return fmt.Errorf("%w: %s", domainerrors.ErrRuleDeleteFailed, err)
	}

	g.logger.Info("Successfully deleted rule file", "group_id", groupID)
	return nil
}

// convertToPrometheusRule converts an AlertRule to a PrometheusRule
func (g *PrometheusRuleGenerator) convertToPrometheusRule(rule *domain.AlertRule) (*domain.PrometheusRule, error) {
	// Render the query template with config
	query, err := rule.RenderQuery()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrQueryRenderingFailed, err)
	}

	// Merge default and override config
	mergedConfig := rule.DefaultConfig.Merge(rule.Config)

	return &domain.PrometheusRule{
		Alert:       fmt.Sprintf("%s_Group_%s", rule.Name, rule.GroupID),
		Expr:        query,
		For:         mergedConfig.ForDuration,
		Labels:      g.buildLabels(rule, mergedConfig),
		Annotations: g.buildAnnotations(rule, mergedConfig),
	}, nil
}

// buildLabels builds Prometheus labels for an AlertRule
func (g *PrometheusRuleGenerator) buildLabels(rule *domain.AlertRule, config domain.AlertRuleConfig) map[string]string {
	labels := map[string]string{
		"severity": string(rule.Severity),
		"group_id": rule.GroupID,
	}

	// Add custom labels from config
	for k, v := range config.Labels {
		labels[k] = v
	}

	return labels
}

// buildAnnotations builds Prometheus annotations for an AlertRule
func (g *PrometheusRuleGenerator) buildAnnotations(rule *domain.AlertRule, config domain.AlertRuleConfig) map[string]string {
	annotations := map[string]string{
		"summary":     rule.Name,
		"description": rule.Description,
	}

	// Add custom annotations from config
	for k, v := range config.Annotations {
		annotations[k] = v
	}

	return annotations
}
