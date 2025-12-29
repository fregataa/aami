package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// MonitoringScriptModel is the GORM model for database operations
type MonitoringScriptModel struct {
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
func (MonitoringScriptModel) TableName() string {
	return "monitoring_scripts"
}

// ToMonitoringScriptModel converts domain.MonitoringScript to MonitoringScriptModel
func ToMonitoringScriptModel(ms *domain.MonitoringScript) *MonitoringScriptModel {
	model := &MonitoringScriptModel{
		ID:            ms.ID,
		Name:          ms.Name,
		ScriptType:    ms.ScriptType,
		ScriptContent: ms.ScriptContent,
		Language:      ms.Language,
		DefaultConfig: JSONB(ms.DefaultConfig),
		Description:   ms.Description,
		Version:       ms.Version,
		Hash:          ms.Hash,
		CreatedAt:     ms.CreatedAt,
		UpdatedAt:     ms.UpdatedAt,
	}
	if ms.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *ms.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts MonitoringScriptModel to domain.MonitoringScript
func (m *MonitoringScriptModel) ToDomain() *domain.MonitoringScript {
	ms := &domain.MonitoringScript{
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
		ms.DeletedAt = &deletedAt
	}

	return ms
}

// MonitoringScriptRepository defines the interface for monitoring script data access
type MonitoringScriptRepository interface {
	Create(ctx context.Context, script *domain.MonitoringScript) error
	GetByID(ctx context.Context, id string) (*domain.MonitoringScript, error)
	GetByName(ctx context.Context, name string) (*domain.MonitoringScript, error)
	GetByScriptType(ctx context.Context, scriptType string) ([]domain.MonitoringScript, error)
	ListActive(ctx context.Context) ([]domain.MonitoringScript, error)
	Update(ctx context.Context, script *domain.MonitoringScript) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.MonitoringScript, int, error)
}

// monitoringScriptRepository implements MonitoringScriptRepository interface using GORM
type monitoringScriptRepository struct {
	db *gorm.DB
}

// NewMonitoringScriptRepository creates a new MonitoringScriptRepository instance
func NewMonitoringScriptRepository(db *gorm.DB) MonitoringScriptRepository {
	return &monitoringScriptRepository{db: db}
}

// Create inserts a new monitoring script into the database
func (r *monitoringScriptRepository) Create(ctx context.Context, script *domain.MonitoringScript) error {
	model := ToMonitoringScriptModel(script)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*script = *model.ToDomain()
	return nil
}

// GetByID retrieves a monitoring script by its ID
func (r *monitoringScriptRepository) GetByID(ctx context.Context, id string) (*domain.MonitoringScript, error) {
	var model MonitoringScriptModel
	err := r.db.WithContext(ctx).
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByName retrieves a monitoring script by its name
func (r *monitoringScriptRepository) GetByName(ctx context.Context, name string) (*domain.MonitoringScript, error) {
	var model MonitoringScriptModel
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&model).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByScriptType retrieves all scripts for a specific script type
func (r *monitoringScriptRepository) GetByScriptType(ctx context.Context, scriptType string) ([]domain.MonitoringScript, error) {
	var models []MonitoringScriptModel
	err := r.db.WithContext(ctx).
		Where("script_type = ?", scriptType).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	scripts := make([]domain.MonitoringScript, len(models))
	for i, model := range models {
		scripts[i] = *model.ToDomain()
	}
	return scripts, nil
}

// ListActive retrieves all active (non-deleted) scripts
func (r *monitoringScriptRepository) ListActive(ctx context.Context) ([]domain.MonitoringScript, error) {
	var models []MonitoringScriptModel
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	scripts := make([]domain.MonitoringScript, len(models))
	for i, model := range models {
		scripts[i] = *model.ToDomain()
	}
	return scripts, nil
}

// Update updates an existing monitoring script
func (r *monitoringScriptRepository) Update(ctx context.Context, script *domain.MonitoringScript) error {
	model := ToMonitoringScriptModel(script)
	return fromGormError(r.db.WithContext(ctx).
		Model(&MonitoringScriptModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on a monitoring script (sets deleted_at timestamp)
func (r *monitoringScriptRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&MonitoringScriptModel{}, "id = ?", id).Error)
}

// Purge permanently removes a monitoring script from the database
func (r *monitoringScriptRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&MonitoringScriptModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted monitoring script
func (r *monitoringScriptRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&MonitoringScriptModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

// List retrieves monitoring scripts with pagination
func (r *monitoringScriptRepository) List(ctx context.Context, page, limit int) ([]domain.MonitoringScript, int, error) {
	var models []MonitoringScriptModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&MonitoringScriptModel{}).Count(&total).Error; err != nil {
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

	scripts := make([]domain.MonitoringScript, len(models))
	for i, model := range models {
		scripts[i] = *model.ToDomain()
	}

	return scripts, int(total), nil
}
