package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// ScriptTemplateModel is the GORM model for database operations
type ScriptTemplateModel struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name          string         `gorm:"not null;index"`
	ScriptType    string         `gorm:"not null;index"`
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
func (ScriptTemplateModel) TableName() string {
	return "script_templates"
}

// ToScriptTemplateModel converts domain.ScriptTemplate to ScriptTemplateModel
func ToScriptTemplateModel(st *domain.ScriptTemplate) *ScriptTemplateModel {
	model := &ScriptTemplateModel{
		ID:            st.ID,
		Name:          st.Name,
		ScriptType:    st.ScriptType,
		ScriptContent: st.ScriptContent,
		Language:      st.Language,
		DefaultConfig: JSONB(st.DefaultConfig),
		Description:   st.Description,
		Version:       st.Version,
		Hash:          st.Hash,
		CreatedAt:     st.CreatedAt,
		UpdatedAt:     st.UpdatedAt,
	}
	if st.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *st.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts ScriptTemplateModel to domain.ScriptTemplate
func (m *ScriptTemplateModel) ToDomain() *domain.ScriptTemplate {
	st := &domain.ScriptTemplate{
		ID:            m.ID,
		Name:          m.Name,
		ScriptType:    m.ScriptType,
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
		st.DeletedAt = &deletedAt
	}

	return st
}

// ScriptTemplateRepository defines the interface for script template data access
type ScriptTemplateRepository interface {
	Create(ctx context.Context, template *domain.ScriptTemplate) error
	GetByID(ctx context.Context, id string) (*domain.ScriptTemplate, error)
	GetByName(ctx context.Context, name string) (*domain.ScriptTemplate, error)
	GetByScriptType(ctx context.Context, scriptType string) ([]domain.ScriptTemplate, error)
	ListActive(ctx context.Context) ([]domain.ScriptTemplate, error)
	Update(ctx context.Context, template *domain.ScriptTemplate) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.ScriptTemplate, int, error)
}

// scriptTemplateRepository implements ScriptTemplateRepository interface using GORM
type scriptTemplateRepository struct {
	db *gorm.DB
}

// NewScriptTemplateRepository creates a new ScriptTemplateRepository instance
func NewScriptTemplateRepository(db *gorm.DB) ScriptTemplateRepository {
	return &scriptTemplateRepository{db: db}
}

// Create inserts a new script template into the database
func (r *scriptTemplateRepository) Create(ctx context.Context, template *domain.ScriptTemplate) error {
	model := ToScriptTemplateModel(template)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*template = *model.ToDomain()
	return nil
}

// GetByID retrieves a script template by its ID
func (r *scriptTemplateRepository) GetByID(ctx context.Context, id string) (*domain.ScriptTemplate, error) {
	var model ScriptTemplateModel
	err := r.db.WithContext(ctx).
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByName retrieves a script template by its name
func (r *scriptTemplateRepository) GetByName(ctx context.Context, name string) (*domain.ScriptTemplate, error) {
	var model ScriptTemplateModel
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&model).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByScriptType retrieves all templates for a specific script type
func (r *scriptTemplateRepository) GetByScriptType(ctx context.Context, scriptType string) ([]domain.ScriptTemplate, error) {
	var models []ScriptTemplateModel
	err := r.db.WithContext(ctx).
		Where("script_type = ?", scriptType).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	templates := make([]domain.ScriptTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}
	return templates, nil
}

// ListActive retrieves all active (non-deleted) templates
func (r *scriptTemplateRepository) ListActive(ctx context.Context) ([]domain.ScriptTemplate, error) {
	var models []ScriptTemplateModel
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	templates := make([]domain.ScriptTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}
	return templates, nil
}

// Update updates an existing script template
func (r *scriptTemplateRepository) Update(ctx context.Context, template *domain.ScriptTemplate) error {
	model := ToScriptTemplateModel(template)
	return fromGormError(r.db.WithContext(ctx).
		Model(&ScriptTemplateModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on a script template (sets deleted_at timestamp)
func (r *scriptTemplateRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&ScriptTemplateModel{}, "id = ?", id).Error)
}

// Purge permanently removes a script template from the database
func (r *scriptTemplateRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&ScriptTemplateModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted script template
func (r *scriptTemplateRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&ScriptTemplateModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

// List retrieves script templates with pagination
func (r *scriptTemplateRepository) List(ctx context.Context, page, limit int) ([]domain.ScriptTemplate, int, error) {
	var models []ScriptTemplateModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ScriptTemplateModel{}).Count(&total).Error; err != nil {
		return nil, 0, fromGormError(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, 0, fromGormError(err)
	}

	templates := make([]domain.ScriptTemplate, len(models))
	for i, model := range models {
		templates[i] = *model.ToDomain()
	}

	return templates, int(total), nil
}
