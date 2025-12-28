package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// CheckInstanceModel is the GORM model for database operations
type CheckInstanceModel struct {
	ID string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	// Template fields (copied from template at creation)
	Name          string `gorm:"not null;type:varchar(255);index"`
	CheckType     string `gorm:"not null;type:varchar(100);index"`
	ScriptContent string `gorm:"not null;type:text"`
	Language      string `gorm:"not null;type:varchar(50)"`
	DefaultConfig JSONB  `gorm:"type:jsonb;default:'{}'"`
	Description   string `gorm:"type:text"`
	Version       string `gorm:"type:varchar(50)"`
	Hash          string `gorm:"type:varchar(64);index"`

	// Instance-specific fields
	Scope       string          `gorm:"not null;index;type:varchar(20)"`
	NamespaceID *string         `gorm:"index"`
	Namespace   *NamespaceModel `gorm:"foreignKey:NamespaceID;references:ID"`
	GroupID     *string         `gorm:"index"`
	Group       *GroupModel     `gorm:"foreignKey:GroupID;references:ID"`
	Config      JSONB           `gorm:"type:jsonb;default:'{}'"`
	Priority    int             `gorm:"not null;default:100;index"`
	IsActive    bool            `gorm:"not null;default:true;index"`

	// Metadata (optional, for tracking origin)
	CreatedFromTemplateID   *string `gorm:"type:varchar(255);index"`
	CreatedFromTemplateName *string `gorm:"type:varchar(255)"`
	TemplateVersion         *string `gorm:"type:varchar(50)"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (CheckInstanceModel) TableName() string {
	return "check_instances"
}

// ToCheckInstanceModel converts domain.CheckInstance to CheckInstanceModel
func ToCheckInstanceModel(ci *domain.CheckInstance) *CheckInstanceModel {
	model := &CheckInstanceModel{
		ID: ci.ID,

		// Template fields
		Name:          ci.Name,
		CheckType:     ci.CheckType,
		ScriptContent: ci.ScriptContent,
		Language:      ci.Language,
		DefaultConfig: JSONB(ci.DefaultConfig),
		Description:   ci.Description,
		Version:       ci.Version,
		Hash:          ci.Hash,

		// Instance-specific fields
		Scope:    string(ci.Scope),
		Config:   JSONB(ci.Config),
		Priority: ci.Priority,
		IsActive: ci.IsActive,

		// Metadata
		CreatedFromTemplateID:   ci.CreatedFromTemplateID,
		CreatedFromTemplateName: ci.CreatedFromTemplateName,
		TemplateVersion:         ci.TemplateVersion,

		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
	}

	if ci.NamespaceID != nil {
		model.NamespaceID = ci.NamespaceID
	}
	if ci.GroupID != nil {
		model.GroupID = ci.GroupID
	}
	if ci.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *ci.DeletedAt, Valid: true}
	}

	return model
}

// ToDomain converts CheckInstanceModel to domain.CheckInstance
func (m *CheckInstanceModel) ToDomain() *domain.CheckInstance {
	ci := &domain.CheckInstance{
		ID: m.ID,

		// Template fields
		Name:          m.Name,
		CheckType:     m.CheckType,
		ScriptContent: m.ScriptContent,
		Language:      m.Language,
		DefaultConfig: map[string]interface{}(m.DefaultConfig),
		Description:   m.Description,
		Version:       m.Version,
		Hash:          m.Hash,

		// Instance-specific fields
		Scope:       domain.InstanceScope(m.Scope),
		NamespaceID: m.NamespaceID,
		GroupID:     m.GroupID,
		Config:      map[string]interface{}(m.Config),
		Priority:    m.Priority,
		IsActive:    m.IsActive,

		// Metadata
		CreatedFromTemplateID:   m.CreatedFromTemplateID,
		CreatedFromTemplateName: m.CreatedFromTemplateName,
		TemplateVersion:         m.TemplateVersion,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		ci.DeletedAt = &deletedAt
	}

	// Convert nested Namespace if present
	if m.Namespace != nil {
		ci.Namespace = m.Namespace.ToDomain()
	}

	// Convert nested Group if present
	if m.Group != nil {
		ci.Group = m.Group.ToDomain()
	}

	return ci
}

// CheckInstanceRepository defines the interface for check instance data access
type CheckInstanceRepository interface {
	Create(ctx context.Context, instance *domain.CheckInstance) error
	GetByID(ctx context.Context, id string) (*domain.CheckInstance, error)
	GetByTemplateID(ctx context.Context, templateID string) ([]domain.CheckInstance, error)
	GetGlobalInstances(ctx context.Context) ([]domain.CheckInstance, error)
	GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckInstance, error)
	GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.CheckInstance, error)
	GetEffectiveInstancesByNamespace(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error)
	GetEffectiveInstancesByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.CheckInstance, error)
	ListActive(ctx context.Context) ([]domain.CheckInstance, error)
	Update(ctx context.Context, instance *domain.CheckInstance) error
	Delete(ctx context.Context, id string) error  // Soft delete
	Purge(ctx context.Context, id string) error   // Hard delete
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.CheckInstance, int, error)
}

// checkInstanceRepository implements CheckInstanceRepository interface using GORM
type checkInstanceRepository struct {
	db *gorm.DB
}

// NewCheckInstanceRepository creates a new CheckInstanceRepository instance
func NewCheckInstanceRepository(db *gorm.DB) CheckInstanceRepository {
	return &checkInstanceRepository{db: db}
}

// Create inserts a new check instance into the database
func (r *checkInstanceRepository) Create(ctx context.Context, instance *domain.CheckInstance) error {
	model := ToCheckInstanceModel(instance)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*instance = *model.ToDomain()
	return nil
}

// GetByID retrieves a check instance by its ID
func (r *checkInstanceRepository) GetByID(ctx context.Context, id string) (*domain.CheckInstance, error) {
	var model CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Preload("Group").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, fromGormError(err)
	}
	return model.ToDomain(), nil
}

// GetByTemplateID retrieves all instances for a specific template
func (r *checkInstanceRepository) GetByTemplateID(ctx context.Context, templateID string) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Preload("Group").
		Where("template_id = ?", templateID).
		Order("priority DESC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetGlobalInstances retrieves all global-scope instances
func (r *checkInstanceRepository) GetGlobalInstances(ctx context.Context) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("scope = ?", domain.ScopeGlobal).
		Where("is_active = ?", true).
		Order("priority DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetByNamespaceID retrieves all namespace-level instances for a specific namespace
func (r *checkInstanceRepository) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Where("scope = ?", domain.ScopeNamespace).
		Where("namespace_id = ?", namespaceID).
		Where("is_active = ?", true).
		Order("priority DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetByGroupID retrieves all group-level instances for a specific group
func (r *checkInstanceRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Preload("Group").
		Where("scope = ?", domain.ScopeGroup).
		Where("group_id = ?", groupID).
		Where("is_active = ?", true).
		Order("priority DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetEffectiveInstance finds the most specific active instance for a template
// Priority: Group > Namespace > Global
func (r *checkInstanceRepository) GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.CheckInstance, error) {
	var model CheckInstanceModel

	// Try Group level first (highest priority)
	if groupID != "" {
		err := r.db.WithContext(ctx).
			Preload("Template").
			Preload("Namespace").
			Preload("Group").
			Where("template_id = ?", templateID).
			Where("scope = ?", domain.ScopeGroup).
			Where("group_id = ?", groupID).
			Where("is_active = ?", true).
			Order("priority DESC").
			First(&model).Error
		if err == nil {
			return model.ToDomain(), nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fromGormError(err)
		}
	}

	// Try Namespace level (medium priority)
	if namespaceID != "" {
		err := r.db.WithContext(ctx).
			Preload("Template").
			Preload("Namespace").
			Where("template_id = ?", templateID).
			Where("scope = ?", domain.ScopeNamespace).
			Where("namespace_id = ?", namespaceID).
			Where("is_active = ?", true).
			Order("priority DESC").
			First(&model).Error
		if err == nil {
			return model.ToDomain(), nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fromGormError(err)
		}
	}

	// Try Global level (lowest priority)
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("template_id = ?", templateID).
		Where("scope = ?", domain.ScopeGlobal).
		Where("is_active = ?", true).
		Order("priority DESC").
		First(&model).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	return model.ToDomain(), nil
}

// GetEffectiveInstancesByNamespace retrieves all effective instances for a namespace
// Combines global and namespace-level instances
func (r *checkInstanceRepository) GetEffectiveInstancesByNamespace(ctx context.Context, namespaceID string) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel

	// Get both global and namespace-level instances
	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Where("is_active = ?", true).
		Where(
			r.db.Where("scope = ?", domain.ScopeGlobal).
				Or("scope = ? AND namespace_id = ?", domain.ScopeNamespace, namespaceID),
		).
		Order("scope DESC, priority DESC").  // Group first, then Namespace, then Global; higher priority first
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	// Deduplicate by created_from_template_id, keeping the highest priority (first occurrence)
	seen := make(map[string]bool)
	var uniqueInstances []domain.CheckInstance

	for _, model := range models {
		// Use created_from_template_id for deduplication if available, otherwise use name+checktype
		key := ""
		if model.CreatedFromTemplateID != nil {
			key = *model.CreatedFromTemplateID
		} else {
			key = model.Name + ":" + model.CheckType
		}

		if !seen[key] {
			seen[key] = true
			uniqueInstances = append(uniqueInstances, *model.ToDomain())
		}
	}

	return uniqueInstances, nil
}

// GetEffectiveInstancesByGroup retrieves all effective instances for a group
// Combines global, namespace-level, and group-level instances
func (r *checkInstanceRepository) GetEffectiveInstancesByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel

	// Get global, namespace-level, and group-level instances
	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("Group").
		Where("is_active = ?", true).
		Where(
			r.db.Where("scope = ?", domain.ScopeGlobal).
				Or("scope = ? AND namespace_id = ?", domain.ScopeNamespace, namespaceID).
				Or("scope = ? AND group_id = ?", domain.ScopeGroup, groupID),
		).
		Order("scope DESC, priority DESC").  // Group first, then Namespace, then Global; higher priority first
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	// Deduplicate by created_from_template_id, keeping the highest priority (first occurrence)
	seen := make(map[string]bool)
	var uniqueInstances []domain.CheckInstance

	for _, model := range models {
		// Use created_from_template_id for deduplication if available, otherwise use name+checktype
		key := ""
		if model.CreatedFromTemplateID != nil {
			key = *model.CreatedFromTemplateID
		} else {
			key = model.Name + ":" + model.CheckType
		}

		if !seen[key] {
			seen[key] = true
			uniqueInstances = append(uniqueInstances, *model.ToDomain())
		}
	}

	return uniqueInstances, nil
}

// ListActive retrieves all active (non-deleted) instances
func (r *checkInstanceRepository) ListActive(ctx context.Context) ([]domain.CheckInstance, error) {
	var models []CheckInstanceModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Preload("Group").
		Where("is_active = ?", true).
		Order("priority DESC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// Update updates an existing check instance
func (r *checkInstanceRepository) Update(ctx context.Context, instance *domain.CheckInstance) error {
	model := ToCheckInstanceModel(instance)
	return fromGormError(r.db.WithContext(ctx).
		Model(&CheckInstanceModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on a check instance (sets deleted_at timestamp)
func (r *checkInstanceRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&CheckInstanceModel{}, "id = ?", id).Error)
}

// Purge permanently removes a check instance from the database
func (r *checkInstanceRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&CheckInstanceModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted check instance
func (r *checkInstanceRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&CheckInstanceModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

// List retrieves check instances with pagination
func (r *checkInstanceRepository) List(ctx context.Context, page, limit int) ([]domain.CheckInstance, int, error) {
	var models []CheckInstanceModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&CheckInstanceModel{}).Count(&total).Error; err != nil {
		return nil, 0, fromGormError(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Namespace").
		Preload("Group").
		Offset(offset).
		Limit(limit).
		Order("priority DESC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, 0, fromGormError(err)
	}

	instances := make([]domain.CheckInstance, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}

	return instances, int(total), nil
}
