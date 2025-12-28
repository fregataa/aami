package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// ExporterModel is the GORM model for database operations
type ExporterModel struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TargetID       string         `gorm:"type:uuid;not null;index"`
	Type           string         `gorm:"not null;index"`
	Port           int            `gorm:"not null"`
	Enabled        bool           `gorm:"not null;default:true"`
	MetricsPath    string         `gorm:"not null;default:'/metrics'"`
	ScrapeInterval string         `gorm:"not null;default:'15s'"`
	ScrapeTimeout  string         `gorm:"not null;default:'10s'"`
	Config         JSONB          `gorm:"type:jsonb;default:'{}'"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (ExporterModel) TableName() string {
	return "exporters"
}

// ToExporterModel converts domain.Exporter to ExporterModel
func ToExporterModel(e *domain.Exporter) *ExporterModel {
	model := &ExporterModel{
		ID:             e.ID,
		TargetID:       e.TargetID,
		Type:           string(e.Type),
		Port:           e.Port,
		Enabled:        e.Enabled,
		MetricsPath:    e.MetricsPath,
		ScrapeInterval: e.ScrapeInterval,
		ScrapeTimeout:  e.ScrapeTimeout,
		Config:         JSONB(e.Config),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
	if e.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *e.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts ExporterModel to domain.Exporter
func (m *ExporterModel) ToDomain() *domain.Exporter {
	e := &domain.Exporter{
		ID:             m.ID,
		TargetID:       m.TargetID,
		Type:           domain.ExporterType(m.Type),
		Port:           m.Port,
		Enabled:        m.Enabled,
		MetricsPath:    m.MetricsPath,
		ScrapeInterval: m.ScrapeInterval,
		ScrapeTimeout:  m.ScrapeTimeout,
		Config:         map[string]interface{}(m.Config),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		e.DeletedAt = &deletedAt
	}
	return e
}

// ExporterRepository defines the interface for exporter data access
type ExporterRepository interface {
	Create(ctx context.Context, exporter *domain.Exporter) error
	GetByID(ctx context.Context, id string) (*domain.Exporter, error)
	GetByTargetID(ctx context.Context, targetID string) ([]domain.Exporter, error)
	GetByType(ctx context.Context, exporterType domain.ExporterType) ([]domain.Exporter, error)
	Update(ctx context.Context, exporter *domain.Exporter) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.Exporter, int, error)
}

// exporterRepository implements ExporterRepository interface using GORM
type exporterRepository struct {
	db *gorm.DB
}

// NewExporterRepository creates a new ExporterRepository instance
func NewExporterRepository(db *gorm.DB) ExporterRepository {
	return &exporterRepository{db: db}
}

func (r *exporterRepository) Create(ctx context.Context, exporter *domain.Exporter) error {
	model := ToExporterModel(exporter)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*exporter = *model.ToDomain()
	return nil
}

func (r *exporterRepository) GetByID(ctx context.Context, id string) (*domain.Exporter, error) {
	var model ExporterModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

func (r *exporterRepository) GetByTargetID(ctx context.Context, targetID string) ([]domain.Exporter, error) {
	var models []ExporterModel
	err := r.db.WithContext(ctx).
		Where("target_id = ?", targetID).
		Order("type ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	exporters := make([]domain.Exporter, len(models))
	for i, model := range models {
		exporters[i] = *model.ToDomain()
	}
	return exporters, nil
}

func (r *exporterRepository) GetByType(ctx context.Context, exporterType domain.ExporterType) ([]domain.Exporter, error) {
	var models []ExporterModel
	err := r.db.WithContext(ctx).
		Where("type = ?", string(exporterType)).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	exporters := make([]domain.Exporter, len(models))
	for i, model := range models {
		exporters[i] = *model.ToDomain()
	}
	return exporters, nil
}

func (r *exporterRepository) Update(ctx context.Context, exporter *domain.Exporter) error {
	model := ToExporterModel(exporter)
	return fromGormError(r.db.WithContext(ctx).
		Model(&ExporterModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on an exporter (sets deleted_at timestamp)
func (r *exporterRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&ExporterModel{}, "id = ?", id).Error)
}

// Purge permanently removes an exporter from the database
func (r *exporterRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&ExporterModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted exporter
func (r *exporterRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&ExporterModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

func (r *exporterRepository) List(ctx context.Context, page, limit int) ([]domain.Exporter, int, error) {
	var models []ExporterModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ExporterModel{}).Count(&total).Error; err != nil {
		return nil, 0, fromGormError(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, 0, fromGormError(err)
	}

	exporters := make([]domain.Exporter, len(models))
	for i, model := range models {
		exporters[i] = *model.ToDomain()
	}

	return exporters, int(total), nil
}
