package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// CheckTemplateModel is the GORM model for database operations
type CheckTemplateModel struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string         `gorm:"not null;index"`
	CheckType     string         `gorm:"not null;index"`
	ScriptContent string         `gorm:"type:text;not null"`
	Language      string         `gorm:"not null;default:'bash'"`
	DefaultConfig JSONB          `gorm:"type:jsonb;default:'{}'"`
	Description   string         `gorm:"type:text"`
	Version       string         `gorm:"not null"`
	Hash          string         `gorm:"not null;index"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (CheckTemplateModel) TableName() string {
	return "check_templates"
}

// ToCheckTemplateModel converts domain.CheckTemplate to CheckTemplateModel
func ToCheckTemplateModel(ct *domain.CheckTemplate) *CheckTemplateModel {
	model := &CheckTemplateModel{
		ID:            ct.ID,
		Name:          ct.Name,
		CheckType:     ct.CheckType,
		ScriptContent: ct.ScriptContent,
		Language:      ct.Language,
		DefaultConfig: JSONB(ct.DefaultConfig),
		Description:   ct.Description,
		Version:       ct.Version,
		Hash:          ct.Hash,
		CreatedAt:     ct.CreatedAt,
		UpdatedAt:     ct.UpdatedAt,
	}
	if ct.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *ct.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts CheckTemplateModel to domain.CheckTemplate
func (m *CheckTemplateModel) ToDomain() *domain.CheckTemplate {
	ct := &domain.CheckTemplate{
		ID:            m.ID,
		Name:          m.Name,
		CheckType:     m.CheckType,
		ScriptContent: m.ScriptContent,
		Language:      m.Language,
		DefaultConfig: map[string]interface{}(m.DefaultConfig),
		Description:   m.Description,
		Version:       m.Version,
		Hash:          m.Hash,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		ct.DeletedAt = &deletedAt
	}

	return ct
}

// CheckTemplateRepository defines the interface for check template data access
type CheckTemplateRepository interface {
	Create(ctx context.Context, template *domain.CheckTemplate) error
	GetByID(ctx context.Context, id string) (*domain.CheckTemplate, error)
	GetByName(ctx context.Context, name string) (*domain.CheckTemplate, error)
	GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckTemplate, error)
	ListActive(ctx context.Context) ([]domain.CheckTemplate, error)
	Update(ctx context.Context, template *domain.CheckTemplate) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.CheckTemplate, int, error)
}

// checkTemplateRepository implements CheckTemplateRepository interface using GORM
type checkTemplateRepository struct {
	db *gorm.DB
}

// NewCheckTemplateRepository creates a new CheckTemplateRepository instance
func NewCheckTemplateRepository(db *gorm.DB) CheckTemplateRepository {
	return &checkTemplateRepository{db: db}
}

// Create inserts a new check template into the database
func (r *checkTemplateRepository) Create(ctx context.Context, template *domain.CheckTemplate) error {
	model := ToCheckTemplateModel(template)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	*template = *model.ToDomain()
	return nil
}

// GetByID retrieves a check template by its ID
func (r *checkTemplateRepository) GetByID(ctx context.Context, id string) (*domain.CheckTemplate, error) {
	var model CheckTemplateModel
	err := r.db.WithContext(ctx).
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// GetByName retrieves a check template by its name
func (r *checkTemplateRepository) GetByName(ctx context.Context, name string) (*domain.CheckTemplate, error) {
	var model CheckTemplateModel
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&model).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// GetByCheckType retrieves all templates for a specific check type
func (r *checkTemplateRepository) GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckTemplate, error) {
	var models []CheckTemplateModel
	err := r.db.WithContext(ctx).
		Where("check_type = ?", checkType).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	templates := make([]domain.CheckTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}
	return templates, nil
}

// ListActive retrieves all active (non-deleted) templates
func (r *checkTemplateRepository) ListActive(ctx context.Context) ([]domain.CheckTemplate, error) {
	var models []CheckTemplateModel
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	templates := make([]domain.CheckTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}
	return templates, nil
}

// Update updates an existing check template
func (r *checkTemplateRepository) Update(ctx context.Context, template *domain.CheckTemplate) error {
	model := ToCheckTemplateModel(template)
	return r.db.WithContext(ctx).
		Model(&CheckTemplateModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error
}

// Delete performs soft delete on a check template (sets deleted_at timestamp)
func (r *checkTemplateRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&CheckTemplateModel{}, "id = ?", id).Error
}

// Purge permanently removes a check template from the database
func (r *checkTemplateRepository) Purge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&CheckTemplateModel{}, "id = ?", id).Error
}

// Restore restores a soft-deleted check template
func (r *checkTemplateRepository) Restore(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&CheckTemplateModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// List retrieves check templates with pagination
func (r *checkTemplateRepository) List(ctx context.Context, page, limit int) ([]domain.CheckTemplate, int, error) {
	var models []CheckTemplateModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&CheckTemplateModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	templates := make([]domain.CheckTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}

	return templates, int(total), nil
}
