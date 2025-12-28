package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// NamespaceModel is the GORM model for database operations
type NamespaceModel struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string         `gorm:"unique;not null;size:50"`
	Description    string         `gorm:"type:text"`
	PolicyPriority int            `gorm:"not null"`
	MergeStrategy  string         `gorm:"not null;default:'merge';size:20"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (NamespaceModel) TableName() string {
	return "namespaces"
}

// ToNamespaceModel converts domain.Namespace to NamespaceModel
func ToNamespaceModel(n *domain.Namespace) *NamespaceModel {
	model := &NamespaceModel{
		ID:             n.ID,
		Name:           n.Name,
		Description:    n.Description,
		PolicyPriority: n.PolicyPriority,
		MergeStrategy:  n.MergeStrategy,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
	if n.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *n.DeletedAt, Valid: true}
	}
	return model
}

// ToDomain converts NamespaceModel to domain.Namespace
func (m *NamespaceModel) ToDomain() *domain.Namespace {
	n := &domain.Namespace{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		PolicyPriority: m.PolicyPriority,
		MergeStrategy:  m.MergeStrategy,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		n.DeletedAt = &deletedAt
	}
	return n
}

// NamespaceRepository defines the interface for namespace data access
type NamespaceRepository interface {
	Create(ctx context.Context, namespace *domain.Namespace) error
	GetByID(ctx context.Context, id string) (*domain.Namespace, error)
	GetByName(ctx context.Context, name string) (*domain.Namespace, error)
	Update(ctx context.Context, namespace *domain.Namespace) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.Namespace, int, error)
	GetAll(ctx context.Context) ([]domain.Namespace, error)
}

// namespaceRepository implements NamespaceRepository interface using GORM
type namespaceRepository struct {
	db *gorm.DB
}

// NewNamespaceRepository creates a new NamespaceRepository instance
func NewNamespaceRepository(db *gorm.DB) NamespaceRepository {
	return &namespaceRepository{db: db}
}

// Create inserts a new namespace into the database
func (r *namespaceRepository) Create(ctx context.Context, namespace *domain.Namespace) error {
	model := ToNamespaceModel(namespace)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	*namespace = *model.ToDomain()
	return nil
}

// GetByID retrieves a namespace by its ID
func (r *namespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	var model NamespaceModel
	err := r.db.WithContext(ctx).
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// GetByName retrieves a namespace by its name
func (r *namespaceRepository) GetByName(ctx context.Context, name string) (*domain.Namespace, error) {
	var model NamespaceModel
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&model).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// Update updates an existing namespace
func (r *namespaceRepository) Update(ctx context.Context, namespace *domain.Namespace) error {
	model := ToNamespaceModel(namespace)
	return r.db.WithContext(ctx).
		Model(&NamespaceModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error
}

// Delete performs soft delete on a namespace (sets deleted_at timestamp)
func (r *namespaceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Delete(&NamespaceModel{}, "id = ?", id).Error
}

// Purge permanently removes a namespace from the database
func (r *namespaceRepository) Purge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&NamespaceModel{}, "id = ?", id).Error
}

// Restore restores a soft-deleted namespace
func (r *namespaceRepository) Restore(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&NamespaceModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// List retrieves a paginated list of namespaces
func (r *namespaceRepository) List(ctx context.Context, page, limit int) ([]domain.Namespace, int, error) {
	var models []NamespaceModel
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&NamespaceModel{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results ordered by policy priority (higher = higher priority)
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Order("policy_priority DESC, name ASC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	namespaces := make([]domain.Namespace, len(models))
	for i, model := range models {
		namespaces[i] = *model.ToDomain()
	}

	return namespaces, int(total), nil
}

// GetAll retrieves all namespaces without pagination
func (r *namespaceRepository) GetAll(ctx context.Context) ([]domain.Namespace, error) {
	var models []NamespaceModel
	err := r.db.WithContext(ctx).
		Order("policy_priority DESC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	// Convert to domain models
	namespaces := make([]domain.Namespace, len(models))
	for i, model := range models {
		namespaces[i] = *model.ToDomain()
	}

	return namespaces, nil
}
