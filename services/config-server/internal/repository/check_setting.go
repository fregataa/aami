package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// CheckSettingModel is the GORM model for database operations
type CheckSettingModel struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GroupID       string         `gorm:"type:uuid;not null;index"`
	Group         *GroupModel    `gorm:"foreignKey:GroupID"`
	CheckType     string         `gorm:"not null;index"`
	Config        JSONB          `gorm:"type:jsonb;default:'{}'"`
	MergeStrategy string         `gorm:"not null;default:'merge'"`
	Priority      int            `gorm:"not null;default:100"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (CheckSettingModel) TableName() string {
	return "check_settings"
}

// ToCheckSettingModel converts domain.CheckSetting to CheckSettingModel
func ToCheckSettingModel(cs *domain.CheckSetting) *CheckSettingModel {
	model := &CheckSettingModel{
		ID:            cs.ID,
		GroupID:       cs.GroupID,
		CheckType:     cs.CheckType,
		Config:        JSONB(cs.Config),
		MergeStrategy: cs.MergeStrategy,
		Priority:      cs.Priority,
		CreatedAt:     cs.CreatedAt,
		UpdatedAt:     cs.UpdatedAt,
	}
	if cs.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *cs.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts CheckSettingModel to domain.CheckSetting
func (m *CheckSettingModel) ToDomain() *domain.CheckSetting {
	cs := &domain.CheckSetting{
		ID:            m.ID,
		GroupID:       m.GroupID,
		CheckType:     m.CheckType,
		Config:        map[string]interface{}(m.Config),
		MergeStrategy: m.MergeStrategy,
		Priority:      m.Priority,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		cs.DeletedAt = &deletedAt
	}

	// Convert Group if loaded
	if m.Group != nil {
		cs.Group = *m.Group.ToDomain()
	}

	return cs
}

// CheckSettingRepository defines the interface for check setting data access
type CheckSettingRepository interface {
	Create(ctx context.Context, setting *domain.CheckSetting) error
	GetByID(ctx context.Context, id string) (*domain.CheckSetting, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckSetting, error)
	GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckSetting, error)
	Update(ctx context.Context, setting *domain.CheckSetting) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.CheckSetting, int, error)
}

// checkSettingRepository implements CheckSettingRepository interface using GORM
type checkSettingRepository struct {
	db *gorm.DB
}

// NewCheckSettingRepository creates a new CheckSettingRepository instance
func NewCheckSettingRepository(db *gorm.DB) CheckSettingRepository {
	return &checkSettingRepository{db: db}
}

func (r *checkSettingRepository) Create(ctx context.Context, setting *domain.CheckSetting) error {
	model := ToCheckSettingModel(setting)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	*setting = *model.ToDomain()
	return nil
}

func (r *checkSettingRepository) GetByID(ctx context.Context, id string) (*domain.CheckSetting, error) {
	var model CheckSettingModel
	err := r.db.WithContext(ctx).
		Preload("Group").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *checkSettingRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckSetting, error) {
	var models []CheckSettingModel
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("priority ASC, check_type ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	settings := make([]domain.CheckSetting, len(models))
	for i, model := range models {
		settings[i] = *model.ToDomain()
	}
	return settings, nil
}

func (r *checkSettingRepository) GetByCheckType(ctx context.Context, checkType string) ([]domain.CheckSetting, error) {
	var models []CheckSettingModel
	err := r.db.WithContext(ctx).
		Where("check_type = ?", checkType).
		Preload("Group").
		Order("priority ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	settings := make([]domain.CheckSetting, len(models))
	for i, model := range models {
		settings[i] = *model.ToDomain()
	}
	return settings, nil
}

func (r *checkSettingRepository) Update(ctx context.Context, setting *domain.CheckSetting) error {
	model := ToCheckSettingModel(setting)
	return r.db.WithContext(ctx).
		Model(&CheckSettingModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error
}

// Delete performs soft delete on a check setting (sets deleted_at timestamp)
func (r *checkSettingRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&CheckSettingModel{}, "id = ?", id).Error
}

// Purge permanently removes a check setting from the database
func (r *checkSettingRepository) Purge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&CheckSettingModel{}, "id = ?", id).Error
}

// Restore restores a soft-deleted check setting
func (r *checkSettingRepository) Restore(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&CheckSettingModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *checkSettingRepository) List(ctx context.Context, page, limit int) ([]domain.CheckSetting, int, error) {
	var models []CheckSettingModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&CheckSettingModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Preload("Group").
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	settings := make([]domain.CheckSetting, len(models))
	for i, model := range models {
		settings[i] = *model.ToDomain()
	}

	return settings, int(total), nil
}
