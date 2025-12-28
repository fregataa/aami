package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// AlertRuleModel is the GORM model for database operations
type AlertRuleModel struct {
	ID      string      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GroupID string      `gorm:"type:uuid;not null;index"`
	Group   *GroupModel `gorm:"foreignKey:GroupID"`

	// Template fields (copied from template at creation)
	Name          string            `gorm:"not null;type:varchar(255);index"`
	Description   string            `gorm:"type:text"`
	Severity      domain.AlertSeverity `gorm:"not null;type:varchar(20);index"`
	QueryTemplate string            `gorm:"not null;type:text"`
	DefaultConfig JSONB             `gorm:"type:jsonb;default:'{}'"`

	// Rule-specific fields
	Enabled       bool   `gorm:"not null;default:true"`
	Config        JSONB  `gorm:"type:jsonb;default:'{}'"`
	MergeStrategy string `gorm:"not null;default:'merge'"`
	Priority      int    `gorm:"not null;default:100"`

	// Metadata (optional, for tracking origin)
	CreatedFromTemplateID   *string `gorm:"type:varchar(255);index"`
	CreatedFromTemplateName *string `gorm:"type:varchar(255)"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (AlertRuleModel) TableName() string {
	return "alert_rules"
}

// ToAlertRuleModel converts domain.AlertRule to AlertRuleModel
func ToAlertRuleModel(ar *domain.AlertRule) *AlertRuleModel {
	model := &AlertRuleModel{
		ID:      ar.ID,
		GroupID: ar.GroupID,

		// Template fields
		Name:          ar.Name,
		Description:   ar.Description,
		Severity:      ar.Severity,
		QueryTemplate: ar.QueryTemplate,
		DefaultConfig: JSONB(ar.DefaultConfig),

		// Rule-specific fields
		Enabled:       ar.Enabled,
		Config:        JSONB(ar.Config),
		MergeStrategy: ar.MergeStrategy,
		Priority:      ar.Priority,

		// Metadata
		CreatedFromTemplateID:   ar.CreatedFromTemplateID,
		CreatedFromTemplateName: ar.CreatedFromTemplateName,

		CreatedAt: ar.CreatedAt,
		UpdatedAt: ar.UpdatedAt,
	}
	if ar.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *ar.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts AlertRuleModel to domain.AlertRule
func (m *AlertRuleModel) ToDomain() *domain.AlertRule {
	ar := &domain.AlertRule{
		ID:      m.ID,
		GroupID: m.GroupID,

		// Template fields
		Name:          m.Name,
		Description:   m.Description,
		Severity:      m.Severity,
		QueryTemplate: m.QueryTemplate,
		DefaultConfig: map[string]interface{}(m.DefaultConfig),

		// Rule-specific fields
		Enabled:       m.Enabled,
		Config:        map[string]interface{}(m.Config),
		MergeStrategy: m.MergeStrategy,
		Priority:      m.Priority,

		// Metadata
		CreatedFromTemplateID:   m.CreatedFromTemplateID,
		CreatedFromTemplateName: m.CreatedFromTemplateName,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		ar.DeletedAt = &deletedAt
	}

	// Convert Group if loaded
	if m.Group != nil {
		ar.Group = *m.Group.ToDomain()
	}

	return ar
}

// AlertRuleRepository defines the interface for alert rule data access
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *domain.AlertRule) error
	GetByID(ctx context.Context, id string) (*domain.AlertRule, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.AlertRule, error)
	GetByTemplateID(ctx context.Context, templateID string) ([]domain.AlertRule, error)
	Update(ctx context.Context, rule *domain.AlertRule) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.AlertRule, int, error)
}

// alertRuleRepository implements AlertRuleRepository interface using GORM
type alertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository creates a new AlertRuleRepository instance
func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &alertRuleRepository{db: db}
}

func (r *alertRuleRepository) Create(ctx context.Context, rule *domain.AlertRule) error {
	model := ToAlertRuleModel(rule)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*rule = *model.ToDomain()
	return nil
}

func (r *alertRuleRepository) GetByID(ctx context.Context, id string) (*domain.AlertRule, error) {
	var model AlertRuleModel
	err := r.db.WithContext(ctx).
		Preload("Group").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

func (r *alertRuleRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.AlertRule, error) {
	var models []AlertRuleModel
	err := r.db.WithContext(ctx).
		Preload("Group").
		Where("group_id = ?", groupID).
		Order("priority DESC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	rules := make([]domain.AlertRule, len(models))
	for i, model := range models {
		rules[i] = *model.ToDomain()
	}
	return rules, nil
}

func (r *alertRuleRepository) GetByTemplateID(ctx context.Context, templateID string) ([]domain.AlertRule, error) {
	var models []AlertRuleModel
	err := r.db.WithContext(ctx).
		Preload("Group").
		Where("created_from_template_id = ?", templateID).
		Order("priority DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	rules := make([]domain.AlertRule, len(models))
	for i, model := range models {
		rules[i] = *model.ToDomain()
	}
	return rules, nil
}

func (r *alertRuleRepository) Update(ctx context.Context, rule *domain.AlertRule) error {
	model := ToAlertRuleModel(rule)
	return fromGormError(r.db.WithContext(ctx).
		Model(&AlertRuleModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on an alert rule (sets deleted_at timestamp)
func (r *alertRuleRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&AlertRuleModel{}, "id = ?", id).Error)
}

// Purge permanently removes an alert rule from the database
func (r *alertRuleRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&AlertRuleModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted alert rule
func (r *alertRuleRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&AlertRuleModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

func (r *alertRuleRepository) List(ctx context.Context, page, limit int) ([]domain.AlertRule, int, error) {
	var models []AlertRuleModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&AlertRuleModel{}).Count(&total).Error; err != nil {
		return nil, 0, fromGormError(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Preload("Group").
		Order("priority DESC, created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, 0, fromGormError(err)
	}

	rules := make([]domain.AlertRule, len(models))
	for i, model := range models {
		rules[i] = *model.ToDomain()
	}

	return rules, int(total), nil
}
