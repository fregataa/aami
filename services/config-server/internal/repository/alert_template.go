package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// AlertTemplateModel is the GORM model for database operations
type AlertTemplateModel struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string         `gorm:"not null;uniqueIndex"`
	Description   string         `gorm:"type:text"`
	Severity      string         `gorm:"not null;index"`
	QueryTemplate string         `gorm:"type:text;not null"`
	DefaultConfig JSONB          `gorm:"type:jsonb;default:'{}'"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (AlertTemplateModel) TableName() string {
	return "alert_templates"
}

// ToAlertTemplateModel converts domain.AlertTemplate to AlertTemplateModel
func ToAlertTemplateModel(at *domain.AlertTemplate) *AlertTemplateModel {
	model := &AlertTemplateModel{
		ID:            at.ID,
		Name:          at.Name,
		Description:   at.Description,
		Severity:      string(at.Severity),
		QueryTemplate: at.QueryTemplate,
		DefaultConfig: JSONB(at.DefaultConfig),
		CreatedAt:     at.CreatedAt,
		UpdatedAt:     at.UpdatedAt,
	}
	if at.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *at.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts AlertTemplateModel to domain.AlertTemplate
func (m *AlertTemplateModel) ToDomain() *domain.AlertTemplate {
	at := &domain.AlertTemplate{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Severity:      domain.AlertSeverity(m.Severity),
		QueryTemplate: m.QueryTemplate,
		DefaultConfig: map[string]interface{}(m.DefaultConfig),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		at.DeletedAt = &deletedAt
	}
	return at
}

// AlertTemplateRepository defines the interface for alert template data access
type AlertTemplateRepository interface {
	Create(ctx context.Context, template *domain.AlertTemplate) error
	GetByID(ctx context.Context, id string) (*domain.AlertTemplate, error)
	GetBySeverity(ctx context.Context, severity domain.AlertSeverity) ([]domain.AlertTemplate, error)
	List(ctx context.Context, page, limit int) ([]domain.AlertTemplate, int, error)
	Update(ctx context.Context, template *domain.AlertTemplate) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
}

// alertTemplateRepository implements AlertTemplateRepository interface using GORM
type alertTemplateRepository struct {
	db *gorm.DB
}

// NewAlertTemplateRepository creates a new AlertTemplateRepository instance
func NewAlertTemplateRepository(db *gorm.DB) AlertTemplateRepository {
	return &alertTemplateRepository{db: db}
}

func (r *alertTemplateRepository) Create(ctx context.Context, template *domain.AlertTemplate) error {
	model := ToAlertTemplateModel(template)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*template = *model.ToDomain()
	return nil
}

func (r *alertTemplateRepository) GetByID(ctx context.Context, id string) (*domain.AlertTemplate, error) {
	var model AlertTemplateModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

func (r *alertTemplateRepository) GetBySeverity(ctx context.Context, severity domain.AlertSeverity) ([]domain.AlertTemplate, error) {
	var models []AlertTemplateModel
	err := r.db.WithContext(ctx).
		Where("severity = ?", string(severity)).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	templates := make([]domain.AlertTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}
	return templates, nil
}

func (r *alertTemplateRepository) List(ctx context.Context, page, limit int) ([]domain.AlertTemplate, int, error) {
	var models []AlertTemplateModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&AlertTemplateModel{}).Count(&total).Error; err != nil {
		return nil, 0, fromGormError(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("severity ASC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, 0, fromGormError(err)
	}

	templates := make([]domain.AlertTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}

	return templates, int(total), nil
}

func (r *alertTemplateRepository) Update(ctx context.Context, template *domain.AlertTemplate) error {
	model := ToAlertTemplateModel(template)
	return fromGormError(r.db.WithContext(ctx).
		Model(&AlertTemplateModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on an alert template (sets deleted_at timestamp)
func (r *alertTemplateRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&AlertTemplateModel{}, "id = ?", id).Error)
}

// Purge permanently removes an alert template from the database
func (r *alertTemplateRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&AlertTemplateModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted alert template
func (r *alertTemplateRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&AlertTemplateModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}
