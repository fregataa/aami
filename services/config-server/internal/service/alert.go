package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fregataa/aami/config-server/internal/api/dto"
	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/google/uuid"
)

// AlertTemplateService handles business logic for alert templates
type AlertTemplateService struct {
	templateRepo repository.AlertTemplateRepository
}

// NewAlertTemplateService creates a new AlertTemplateService
func NewAlertTemplateService(templateRepo repository.AlertTemplateRepository) *AlertTemplateService {
	return &AlertTemplateService{
		templateRepo: templateRepo,
	}
}

// Create creates a new alert template
func (s *AlertTemplateService) Create(ctx context.Context, req dto.CreateAlertTemplateRequest) (*domain.AlertTemplate, error) {
	// Validate severity
	if !req.Severity.IsValid() {
		return nil, domainerrors.NewValidationError("severity", "invalid severity value")
	}

	// Check if template ID already exists
	existing, err := s.templateRepo.GetByID(ctx, req.ID)
	if err == nil && existing != nil {
		return nil, domainerrors.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}

	config := req.DefaultConfig
	if config == nil {
		config = make(map[string]interface{})
	}

	template := &domain.AlertTemplate{
		ID:            req.ID,
		Name:          req.Name,
		Description:   req.Description,
		Severity:      req.Severity,
		QueryTemplate: req.QueryTemplate,
		DefaultConfig: config,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// GetByID retrieves an alert template by ID
func (s *AlertTemplateService) GetByID(ctx context.Context, id string) (*domain.AlertTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return template, nil
}

// Update updates an existing alert template
func (s *AlertTemplateService) Update(ctx context.Context, id string, req dto.UpdateAlertTemplateRequest) (*domain.AlertTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		template.Name = *req.Name
	}

	if req.Description != nil {
		template.Description = *req.Description
	}

	if req.Severity != nil {
		if !req.Severity.IsValid() {
			return nil, domainerrors.NewValidationError("severity", "invalid severity value")
		}
		template.Severity = *req.Severity
	}

	if req.QueryTemplate != nil {
		template.QueryTemplate = *req.QueryTemplate
	}

	if req.DefaultConfig != nil {
		template.DefaultConfig = req.DefaultConfig
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// Delete performs soft delete on an alert template
func (s *AlertTemplateService) Delete(ctx context.Context, id string) error {
	_, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return domainerrors.ErrNotFound
		}
		return err
	}

	return s.templateRepo.Delete(ctx, id)
}

// Purge permanently removes an alert template (hard delete, admin operation)
func (s *AlertTemplateService) Purge(ctx context.Context, id string) error {
	return s.templateRepo.Purge(ctx, id)
}

// Restore restores a soft-deleted alert template
func (s *AlertTemplateService) Restore(ctx context.Context, id string) error {
	return s.templateRepo.Restore(ctx, id)
}

// List retrieves a paginated list of alert templates
func (s *AlertTemplateService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.AlertTemplate, int, error) {
	pagination.Normalize()
	return s.templateRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetBySeverity retrieves alert templates by severity
func (s *AlertTemplateService) GetBySeverity(ctx context.Context, severity domain.AlertSeverity) ([]domain.AlertTemplate, error) {
	if !severity.IsValid() {
		return nil, domainerrors.NewValidationError("severity", "invalid severity value")
	}
	return s.templateRepo.GetBySeverity(ctx, severity)
}

// AlertRuleService handles business logic for alert rules
type AlertRuleService struct {
	ruleRepo       repository.AlertRuleRepository
	templateRepo   repository.AlertTemplateRepository
	groupRepo      repository.GroupRepository
	ruleGenerator  *PrometheusRuleGenerator
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

// CreateFromTemplate creates a new alert rule from a template
func (s *AlertRuleService) CreateFromTemplate(ctx context.Context, req dto.CreateAlertRuleFromTemplateRequest) (*domain.AlertRule, error) {
	// Validate group exists
	_, err := s.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}

	// Get template
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}

	// Set defaults
	mergeStrategy := req.MergeStrategy
	if mergeStrategy == "" {
		mergeStrategy = "override"
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	config := req.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// Use domain constructor to deep copy template fields
	rule := domain.NewAlertRuleFromTemplate(template, req.GroupID, config)

	// Set ID and override merge strategy/priority
	rule.ID = uuid.New().String()
	rule.Enabled = req.Enabled
	rule.MergeStrategy = mergeStrategy
	rule.Priority = priority

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, req.GroupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule creation",
				"group_id", req.GroupID, "rule_id", rule.ID, "error", err)
		}
	}

	// Load with relationships
	return s.ruleRepo.GetByID(ctx, rule.ID)
}

// CreateDirect creates a new alert rule directly without a template
func (s *AlertRuleService) CreateDirect(ctx context.Context, req dto.CreateAlertRuleDirectRequest) (*domain.AlertRule, error) {
	// Validate group exists
	_, err := s.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}

	// Validate severity
	if !req.Severity.IsValid() {
		return nil, domainerrors.NewValidationError("severity", "invalid severity value")
	}

	// Set defaults
	mergeStrategy := req.MergeStrategy
	if mergeStrategy == "" {
		mergeStrategy = "override"
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	config := req.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	defaultConfig := req.DefaultConfig
	if defaultConfig == nil {
		defaultConfig = make(map[string]interface{})
	}

	// Create rule directly
	rule := &domain.AlertRule{
		ID:            uuid.New().String(),
		GroupID:       req.GroupID,
		Name:          req.Name,
		Description:   req.Description,
		Severity:      req.Severity,
		QueryTemplate: req.QueryTemplate,
		DefaultConfig: defaultConfig,
		Enabled:       req.Enabled,
		Config:        config,
		MergeStrategy: mergeStrategy,
		Priority:      priority,
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, req.GroupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule creation",
				"group_id", req.GroupID, "rule_id", rule.ID, "error", err)
		}
	}

	// Load with relationships
	return s.ruleRepo.GetByID(ctx, rule.ID)
}

// GetByID retrieves an alert rule by ID
func (s *AlertRuleService) GetByID(ctx context.Context, id string) (*domain.AlertRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return rule, nil
}

// Update updates an existing alert rule
func (s *AlertRuleService) Update(ctx context.Context, id string, req dto.UpdateAlertRuleRequest) (*domain.AlertRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}

	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if req.Config != nil {
		rule.Config = req.Config
	}

	if req.MergeStrategy != nil {
		rule.MergeStrategy = *req.MergeStrategy
	}

	if req.Priority != nil {
		rule.Priority = *req.Priority
	}

	groupID := rule.GroupID

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	// Trigger Prometheus rule generation for the group
	if s.ruleGenerator != nil {
		if err := s.regenerateAndReload(ctx, groupID); err != nil {
			// Log error but don't fail the operation
			slog.Warn("Failed to regenerate Prometheus rules after alert rule update",
				"group_id", groupID, "rule_id", id, "error", err)
		}
	}

	return s.ruleRepo.GetByID(ctx, id)
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
func (s *AlertRuleService) List(ctx context.Context, pagination dto.PaginationRequest) ([]domain.AlertRule, int, error) {
	pagination.Normalize()
	return s.ruleRepo.List(ctx, pagination.Page, pagination.Limit)
}

// GetByGroupID retrieves alert rules by group ID
func (s *AlertRuleService) GetByGroupID(ctx context.Context, groupID string) ([]domain.AlertRule, error) {
	_, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}
	return s.ruleRepo.GetByGroupID(ctx, groupID)
}

// GetByTemplateID retrieves alert rules by template ID
func (s *AlertRuleService) GetByTemplateID(ctx context.Context, templateID string) ([]domain.AlertRule, error) {
	_, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, domainerrors.ErrForeignKeyViolation
		}
		return nil, err
	}
	return s.ruleRepo.GetByTemplateID(ctx, templateID)
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
