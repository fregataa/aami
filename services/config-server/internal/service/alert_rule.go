package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fregataa/aami/config-server/internal/action"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
)

// AlertRuleService handles business logic for alert rules
type AlertRuleService struct {
	ruleRepo         repository.AlertRuleRepository
	templateRepo     repository.AlertTemplateRepository
	groupRepo        repository.GroupRepository
	targetRepo       repository.TargetRepository
	ruleGenerator    *PrometheusRuleGenerator
	prometheusClient *prometheus.PrometheusClient
}

// NewAlertRuleService creates a new AlertRuleService
func NewAlertRuleService(
	ruleRepo repository.AlertRuleRepository,
	templateRepo repository.AlertTemplateRepository,
	groupRepo repository.GroupRepository,
	ruleGenerator *PrometheusRuleGenerator,
	prometheusClient *prometheus.PrometheusClient,
) *AlertRuleService {
	return &AlertRuleService{
		ruleRepo:         ruleRepo,
		templateRepo:     templateRepo,
		groupRepo:        groupRepo,
		ruleGenerator:    ruleGenerator,
		prometheusClient: prometheusClient,
	}
}

// NewAlertRuleServiceWithTargetRepo creates a new AlertRuleService with target repository
// Use this constructor when you need effective rules by target functionality
func NewAlertRuleServiceWithTargetRepo(
	ruleRepo repository.AlertRuleRepository,
	templateRepo repository.AlertTemplateRepository,
	groupRepo repository.GroupRepository,
	targetRepo repository.TargetRepository,
	ruleGenerator *PrometheusRuleGenerator,
	prometheusClient *prometheus.PrometheusClient,
) *AlertRuleService {
	return &AlertRuleService{
		ruleRepo:         ruleRepo,
		templateRepo:     templateRepo,
		groupRepo:        groupRepo,
		targetRepo:       targetRepo,
		ruleGenerator:    ruleGenerator,
		prometheusClient: prometheusClient,
	}
}

// CreateFromTemplate creates a new alert rule from a template
func (s *AlertRuleService) CreateFromTemplate(ctx context.Context, act action.CreateAlertRuleFromTemplate) (action.AlertRuleResult, error) {
	// Validate group exists
	_, err := s.groupRepo.GetByID(ctx, act.GroupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.AlertRuleResult{}, domainerrors.ErrForeignKeyViolation
		}
		return action.AlertRuleResult{}, err
	}

	// Get template
	template, err := s.templateRepo.GetByID(ctx, act.TemplateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.AlertRuleResult{}, domainerrors.ErrForeignKeyViolation
		}
		return action.AlertRuleResult{}, err
	}

	// Set defaults
	mergeStrategy := act.MergeStrategy
	if mergeStrategy == "" {
		mergeStrategy = "override"
	}

	priority := act.Priority
	if priority == 0 {
		priority = 100
	}

	// Use domain constructor to deep copy template fields
	rule := domain.NewAlertRuleFromTemplate(template, act.GroupID, act.Config)

	// Set ID and override merge strategy/priority
	rule.ID = uuid.New().String()
	rule.Enabled = act.Enabled
	rule.MergeStrategy = mergeStrategy
	rule.Priority = priority

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return action.AlertRuleResult{}, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, act.GroupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule creation",
				"group_id", act.GroupID, "rule_id", rule.ID, "error", err)
		}
	}

	// Load with relationships
	rule, err = s.ruleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		return action.AlertRuleResult{}, err
	}
	return action.NewAlertRuleResult(rule), nil
}

// CreateDirect creates a new alert rule directly without a template
func (s *AlertRuleService) CreateDirect(ctx context.Context, act action.CreateAlertRuleDirect) (action.AlertRuleResult, error) {
	// Validate group exists
	_, err := s.groupRepo.GetByID(ctx, act.GroupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.AlertRuleResult{}, domainerrors.ErrForeignKeyViolation
		}
		return action.AlertRuleResult{}, err
	}

	// Validate severity
	if !act.Severity.IsValid() {
		return action.AlertRuleResult{}, domainerrors.NewValidationError("severity", "invalid severity value")
	}

	// Set defaults
	mergeStrategy := act.MergeStrategy
	if mergeStrategy == "" {
		mergeStrategy = "override"
	}

	priority := act.Priority
	if priority == 0 {
		priority = 100
	}

	// Create rule directly
	rule := &domain.AlertRule{
		ID:            uuid.New().String(),
		GroupID:       act.GroupID,
		Name:          act.Name,
		Description:   act.Description,
		Severity:      act.Severity,
		QueryTemplate: act.QueryTemplate,
		DefaultConfig: act.DefaultConfig,
		Enabled:       act.Enabled,
		Config:        act.Config,
		MergeStrategy: mergeStrategy,
		Priority:      priority,
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return action.AlertRuleResult{}, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, act.GroupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule creation",
				"group_id", act.GroupID, "rule_id", rule.ID, "error", err)
		}
	}

	// Load with relationships
	rule, err = s.ruleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		return action.AlertRuleResult{}, err
	}
	return action.NewAlertRuleResult(rule), nil
}

// GetByID retrieves an alert rule by ID
func (s *AlertRuleService) GetByID(ctx context.Context, id string) (action.AlertRuleResult, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.AlertRuleResult{}, domainerrors.ErrNotFound
		}
		return action.AlertRuleResult{}, err
	}
	return action.NewAlertRuleResult(rule), nil
}

// Update updates an existing alert rule
func (s *AlertRuleService) Update(ctx context.Context, id string, act action.UpdateAlertRule) (action.AlertRuleResult, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return action.AlertRuleResult{}, domainerrors.ErrNotFound
		}
		return action.AlertRuleResult{}, err
	}

	if act.Enabled != nil {
		rule.Enabled = *act.Enabled
	}

	if act.Config != nil {
		rule.Config = *act.Config
	}

	if act.MergeStrategy != nil {
		rule.MergeStrategy = *act.MergeStrategy
	}

	if act.Priority != nil {
		rule.Priority = *act.Priority
	}

	groupID := rule.GroupID

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return action.AlertRuleResult{}, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, groupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule update",
				"group_id", groupID, "rule_id", id, "error", err)
		}
	}

	rule, err = s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return action.AlertRuleResult{}, err
	}
	return action.NewAlertRuleResult(rule), nil
}

// Delete performs soft delete on an alert rule
func (s *AlertRuleService) Delete(ctx context.Context, id string) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	groupID := rule.GroupID

	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, groupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule deletion",
				"group_id", groupID, "rule_id", id, "error", err)
		}
	}

	return nil
}

// Purge permanently removes an alert rule (hard delete, admin operation)
func (s *AlertRuleService) Purge(ctx context.Context, id string) error {
	return s.ruleRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted alert rule
func (s *AlertRuleService) Restore(ctx context.Context, id string) error {
	return s.ruleRepo.Restore(ctx, id)
}

// List retrieves a paginated list of alert rules
func (s *AlertRuleService) List(ctx context.Context, pagination action.Pagination) (action.ListResult[action.AlertRuleResult], error) {
	rules, total, err := s.ruleRepo.List(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return action.ListResult[action.AlertRuleResult]{}, err
	}

	results := action.NewAlertRuleResultList(rules)
	return action.NewListResult(results, pagination, total), nil
}

// GetByGroupID retrieves alert rules by group ID
func (s *AlertRuleService) GetByGroupID(ctx context.Context, groupID string) ([]action.AlertRuleResult, error) {
	_, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}
	rules, err := s.ruleRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return action.NewAlertRuleResultList(rules), nil
}

// GetByTemplateID retrieves alert rules by template ID
func (s *AlertRuleService) GetByTemplateID(ctx context.Context, templateID string) ([]action.AlertRuleResult, error) {
	_, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}
	rules, err := s.ruleRepo.GetByTemplateID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	return action.NewAlertRuleResultList(rules), nil
}

// regenerateAndReload regenerates Prometheus rules for a group and triggers reload
func (s *AlertRuleService) regenerateAndReload(ctx context.Context, groupID string) error {
	// Generate rules for the group
	if err := s.ruleGenerator.GenerateRulesForGroup(ctx, groupID); err != nil {
		return err
	}

	// Trigger Prometheus reload if client is available
	if s.prometheusClient != nil {
		if err := s.prometheusClient.Reload(ctx); err != nil {
			slog.Warn("Failed to reload Prometheus after rule regeneration",
				"group_id", groupID, "error", err)
			// Don't return error - rules are written, reload can be done manually
		}
	}

	return nil
}

// GetEffectiveRulesByTargetID retrieves all effective alert rules for a target
// considering group membership and rule priority
func (s *AlertRuleService) GetEffectiveRulesByTargetID(ctx context.Context, targetID string) (*domain.EffectiveAlertRulesResult, error) {
	if s.targetRepo == nil {
		return nil, domainerrors.NewValidationError("service", "target repository not configured")
	}

	// Get target with groups
	target, err := s.targetRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	if len(target.Groups) == 0 {
		return &domain.EffectiveAlertRulesResult{
			Target: target,
			Rules:  []domain.EffectiveAlertRule{},
		}, nil
	}

	// Map to track effective rules by name (for deduplication)
	effectiveRulesMap := make(map[string]domain.EffectiveAlertRule)

	// Process each group the target belongs to (sorted by priority)
	// Higher priority groups override lower priority rules with the same name
	for _, group := range target.Groups {
		rules, err := s.ruleRepo.GetByGroupID(ctx, group.ID)
		if err != nil {
			return nil, err
		}

		for _, rule := range rules {
			if !rule.Enabled || rule.DeletedAt != nil {
				continue
			}

			// Render the query
			renderedQuery, err := rule.RenderQuery()
			if err != nil {
				slog.Warn("Failed to render query for rule",
					"rule_id", rule.ID, "rule_name", rule.Name, "error", err)
				renderedQuery = rule.QueryTemplate // Fallback to template
			}

			// Create a copy of the group for source tracking
			sourceGroup := group

			effectiveRule := domain.EffectiveAlertRule{
				AlertRule:     rule,
				RenderedQuery: renderedQuery,
				SourceGroup:   &sourceGroup,
			}

			// Check if we already have a rule with the same name
			if existing, exists := effectiveRulesMap[rule.Name]; exists {
				// Higher priority rule wins, or merge based on strategy
				if rule.MergeStrategy == "merge" {
					// Merge configs: existing + new
					mergedConfig := existing.Config.Merge(rule.Config)
					effectiveRule.Config = mergedConfig
				}
				// For "override" strategy, just replace entirely
			}

			effectiveRulesMap[rule.Name] = effectiveRule
		}
	}

	// Convert map to slice
	effectiveRules := make([]domain.EffectiveAlertRule, 0, len(effectiveRulesMap))
	for _, rule := range effectiveRulesMap {
		effectiveRules = append(effectiveRules, rule)
	}

	return &domain.EffectiveAlertRulesResult{
		Target: target,
		Rules:  effectiveRules,
	}, nil
}
