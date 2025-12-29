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
	ID           string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string          `gorm:"not null;index"`
	NamespaceID  string          `gorm:"type:uuid;not null;index"`
	Namespace    *NamespaceModel `gorm:"foreignKey:NamespaceID"`
	ParentID     *string         `gorm:"type:uuid;index"`
	Parent       *GroupModel     `gorm:"foreignKey:ParentID"`
	Children     []GroupModel    `gorm:"foreignKey:ParentID"`
	Description  string          `gorm:"type:text"`
	Priority     int             `gorm:"not null;default:100"`
	IsDefaultOwn bool            `gorm:"not null;default:false;index"`
	Metadata     JSONB           `gorm:"type:jsonb;default:'{}'"`
	DeletedAt    gorm.DeletedAt  `gorm:"index"`
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
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
		NamespaceID:  g.NamespaceID,
		ParentID:     g.ParentID,
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
		NamespaceID:  m.NamespaceID,
		ParentID:     m.ParentID,
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

	// Convert Namespace if loaded
	if m.Namespace != nil {
		g.Namespace = m.Namespace.ToDomain()
	}

	// Convert Parent if loaded
	if m.Parent != nil {
		parent := m.Parent.ToDomain()
		g.Parent = parent
	}

	// Convert Children if loaded
	if len(m.Children) > 0 {
		g.Children = make([]domain.Group, len(m.Children))
		for i, child := range m.Children {
			g.Children[i] = *child.ToDomain()
		}
	}

	return g
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.Group, error)
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.Group, int, error)
	GetChildren(ctx context.Context, parentID string) ([]domain.Group, error)
	GetAncestors(ctx context.Context, groupID string) ([]domain.Group, error)
	CountByNamespaceID(ctx context.Context, namespaceID string) (int64, error)
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
		Preload("Namespace").
		Preload("Parent").
		Preload("Children").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByNamespace retrieves all groups in a namespace
func (r *groupRepository) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.Group, error) {
	var models []GroupModel
	err := r.db.WithContext(ctx).
		Where("namespace_id = ?", namespaceID).
		Preload("Namespace").
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	groups := make([]domain.Group, len(models))
	for i, model := range models {
		groups[i] = *model.ToDomain()
	}
	return groups, nil
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

// GetChildren retrieves all direct children of a group
func (r *groupRepository) GetChildren(ctx context.Context, parentID string) ([]domain.Group, error) {
	var models []GroupModel
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	children := make([]domain.Group, len(models))
	for i, model := range models {
		children[i] = *model.ToDomain()
	}
	return children, nil
}

// GetAncestors retrieves all ancestors of a group (parent, grandparent, etc.)
func (r *groupRepository) GetAncestors(ctx context.Context, groupID string) ([]domain.Group, error) {
	var models []GroupModel

	// Recursive CTE to get all ancestors
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, namespace_id, parent_id, description, priority, metadata, created_at, updated_at
			FROM groups
			WHERE id = ?
			UNION ALL
			SELECT g.id, g.name, g.namespace_id, g.parent_id, g.description, g.priority, g.metadata, g.created_at, g.updated_at
			FROM groups g
			INNER JOIN ancestors a ON g.id = a.parent_id
		)
		SELECT * FROM ancestors WHERE id != ?
		ORDER BY priority DESC
	`

	err := r.db.WithContext(ctx).Raw(query, groupID, groupID).Scan(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	ancestors := make([]domain.Group, len(models))
	for i, model := range models {
		ancestors[i] = *model.ToDomain()
	}
	return ancestors, nil
}

// CountByNamespaceID counts the number of groups in a specific namespace
func (r *groupRepository) CountByNamespaceID(ctx context.Context, namespaceID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GroupModel{}).
		Where("namespace_id = ?", namespaceID).
		Count(&count).Error
	return count, fromGormError(err)
}
