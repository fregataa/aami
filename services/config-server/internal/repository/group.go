package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// GroupModel is the GORM model for database operations
type GroupModel struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string         `gorm:"not null;index"`
	Description  string         `gorm:"type:text"`
	Priority     int            `gorm:"not null;default:100"`
	IsDefaultOwn bool           `gorm:"not null;default:false;index"`
	Metadata     JSONB          `gorm:"type:jsonb;default:'{}'"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (GroupModel) TableName() string {
	return "groups"
}

// JSONB is a custom type for JSONB fields
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ToGroupModel converts domain.Group to GroupModel
func ToGroupModel(g *domain.Group) *GroupModel {
	// Convert map[string]string to map[string]interface{} for JSONB
	metadata := make(map[string]interface{})
	for k, v := range g.Metadata {
		metadata[k] = v
	}

	model := &GroupModel{
		ID:           g.ID,
		Name:         g.Name,
		Description:  g.Description,
		Priority:     g.Priority,
		IsDefaultOwn: g.IsDefaultOwn,
		Metadata:     JSONB(metadata),
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
	if g.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *g.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts GroupModel to domain.Group
func (m *GroupModel) ToDomain() *domain.Group {
	// Convert map[string]interface{} to map[string]string for Metadata
	metadata := make(map[string]string)
	for k, v := range m.Metadata {
		if strVal, ok := v.(string); ok {
			metadata[k] = strVal
		}
	}

	g := &domain.Group{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		Priority:     m.Priority,
		IsDefaultOwn: m.IsDefaultOwn,
		Metadata:     metadata,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		g.DeletedAt = &deletedAt
	}

	return g
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.Group, int, error)
}

// groupRepository implements GroupRepository interface using GORM
type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new GroupRepository instance
func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

// Create inserts a new group into the database
func (r *groupRepository) Create(ctx context.Context, group *domain.Group) error {
	model := ToGroupModel(group)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*group = *model.ToDomain()
	return nil
}

// GetByID retrieves a group by its ID
func (r *groupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	var model GroupModel
	err := r.db.WithContext(ctx).
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// Update updates an existing group
func (r *groupRepository) Update(ctx context.Context, group *domain.Group) error {
	model := ToGroupModel(group)
	return fromGormError(r.db.WithContext(ctx).
		Model(&GroupModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on a group (sets deleted_at timestamp)
func (r *groupRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&GroupModel{}, "id = ?", id).Error)
}

// Purge permanently removes a group from the database
func (r *groupRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&GroupModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted group
func (r *groupRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&GroupModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

// List retrieves groups with pagination
func (r *groupRepository) List(ctx context.Context, page, limit int) ([]domain.Group, int, error) {
	var models []GroupModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&GroupModel{}).Count(&total).Error; err != nil {
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

	groups := make([]domain.Group, len(models))
	for i, model := range models {
		groups[i] = *model.ToDomain()
	}

	return groups, int(total), nil
}
