package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// TargetGroupModel is the ORM model for target_groups junction table
type TargetGroupModel struct {
	TargetID     string    `gorm:"primaryKey;type:uuid;not null"`
	GroupID      string    `gorm:"primaryKey;type:uuid;not null"`
	IsDefaultOwn bool      `gorm:"not null;default:false;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`

	Target TargetModel `gorm:"foreignKey:TargetID;references:ID;constraint:OnDelete:CASCADE"`
	Group  GroupModel  `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE"`
}

func (TargetGroupModel) TableName() string {
	return "target_groups"
}

// Converters
func ToTargetGroupModel(d *domain.TargetGroup) *TargetGroupModel {
	return &TargetGroupModel{
		TargetID:     d.TargetID,
		GroupID:      d.GroupID,
		IsDefaultOwn: d.IsDefaultOwn,
		CreatedAt:    d.CreatedAt,
	}
}

func (m *TargetGroupModel) ToDomain() *domain.TargetGroup {
	return &domain.TargetGroup{
		TargetID:     m.TargetID,
		GroupID:      m.GroupID,
		IsDefaultOwn: m.IsDefaultOwn,
		CreatedAt:    m.CreatedAt,
	}
}

// TargetGroupRepository interface
type TargetGroupRepository interface {
	// Create adds a new target-group mapping
	Create(ctx context.Context, tg *domain.TargetGroup) error

	// CreateBatch adds multiple mappings in a transaction
	CreateBatch(ctx context.Context, tgs []domain.TargetGroup) error

	// GetByTarget retrieves all group mappings for a target
	GetByTarget(ctx context.Context, targetID string) ([]domain.TargetGroup, error)

	// GetByGroup retrieves all target mappings for a group
	GetByGroup(ctx context.Context, groupID string) ([]domain.TargetGroup, error)

	// CountByTarget returns the number of group mappings for a target
	CountByTarget(ctx context.Context, targetID string) (int64, error)

	// CountByGroup returns the number of target mappings for a group
	CountByGroup(ctx context.Context, groupID string) (int64, error)

	// Delete removes a specific target-group mapping
	Delete(ctx context.Context, targetID, groupID string) error

	// DeleteByTarget removes all mappings for a target
	DeleteByTarget(ctx context.Context, targetID string) error

	// DeleteByGroup removes all mappings for a group
	DeleteByGroup(ctx context.Context, groupID string) error

	// Exists checks if a mapping exists
	Exists(ctx context.Context, targetID, groupID string) (bool, error)
}

// targetGroupRepository implementation
type targetGroupRepository struct {
	db *gorm.DB
}

func NewTargetGroupRepository(db *gorm.DB) TargetGroupRepository {
	return &targetGroupRepository{db: db}
}

func (r *targetGroupRepository) Create(ctx context.Context, tg *domain.TargetGroup) error {
	model := ToTargetGroupModel(tg)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*tg = *model.ToDomain()
	return nil
}

func (r *targetGroupRepository) CreateBatch(ctx context.Context, tgs []domain.TargetGroup) error {
	if len(tgs) == 0 {
		return nil
	}

	models := make([]TargetGroupModel, len(tgs))
	for i, tg := range tgs {
		models[i] = *ToTargetGroupModel(&tg)
	}

	return fromGormError(r.db.WithContext(ctx).Create(&models).Error)
}

func (r *targetGroupRepository) GetByTarget(ctx context.Context, targetID string) ([]domain.TargetGroup, error) {
	var models []TargetGroupModel
	err := r.db.WithContext(ctx).
		Where("target_id = ?", targetID).
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	result := make([]domain.TargetGroup, len(models))
	for i, m := range models {
		result[i] = *m.ToDomain()
	}
	return result, nil
}

func (r *targetGroupRepository) GetByGroup(ctx context.Context, groupID string) ([]domain.TargetGroup, error) {
	var models []TargetGroupModel
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	result := make([]domain.TargetGroup, len(models))
	for i, m := range models {
		result[i] = *m.ToDomain()
	}
	return result, nil
}

func (r *targetGroupRepository) CountByTarget(ctx context.Context, targetID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TargetGroupModel{}).
		Where("target_id = ?", targetID).
		Count(&count).Error
	return count, fromGormError(err)
}

func (r *targetGroupRepository) CountByGroup(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TargetGroupModel{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, fromGormError(err)
}

func (r *targetGroupRepository) Delete(ctx context.Context, targetID, groupID string) error {
	return fromGormError(r.db.WithContext(ctx).
		Delete(&TargetGroupModel{}, "target_id = ? AND group_id = ?", targetID, groupID).
		Error)
}

func (r *targetGroupRepository) DeleteByTarget(ctx context.Context, targetID string) error {
	return fromGormError(r.db.WithContext(ctx).
		Delete(&TargetGroupModel{}, "target_id = ?", targetID).
		Error)
}

func (r *targetGroupRepository) DeleteByGroup(ctx context.Context, groupID string) error {
	return fromGormError(r.db.WithContext(ctx).
		Delete(&TargetGroupModel{}, "group_id = ?", groupID).
		Error)
}

func (r *targetGroupRepository) Exists(ctx context.Context, targetID, groupID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TargetGroupModel{}).
		Where("target_id = ? AND group_id = ?", targetID, groupID).
		Count(&count).Error
	return count > 0, fromGormError(err)
}
