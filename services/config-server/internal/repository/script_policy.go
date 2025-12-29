package repository

import (
	"context"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// ScriptPolicyModel is the GORM model for database operations
type ScriptPolicyModel struct {
	ID string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

	// Template fields (copied from template at creation)
	Name          string `gorm:"not null;type:varchar(255);index"`
	ScriptType     string `gorm:"not null;type:varchar(100);index"`
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
func (ScriptPolicyModel) TableName() string {
	return "script_policies"
}

// ToScriptPolicyModel converts domain.ScriptPolicy to ScriptPolicyModel
func ToScriptPolicyModel(ci *domain.ScriptPolicy) *ScriptPolicyModel {
	model := &ScriptPolicyModel{
		ID: ci.ID,

		// Template fields
		Name:          ci.Name,
		ScriptType:     ci.ScriptType,
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

// ToDomain converts ScriptPolicyModel to domain.ScriptPolicy
func (m *ScriptPolicyModel) ToDomain() *domain.ScriptPolicy {
	ci := &domain.ScriptPolicy{
		ID: m.ID,

		// Template fields
		Name:          m.Name,
		ScriptType:     m.ScriptType,
		ScriptContent: m.ScriptContent,
		Language:      m.Language,
		DefaultConfig: map[string]interface{}(m.DefaultConfig),
		Description:   m.Description,
		Version:       m.Version,
		Hash:          m.Hash,

		// Instance-specific fields
		Scope:       domain.PolicyScope(m.Scope),
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

// ScriptPolicyRepository defines the interface for script policy data access
type ScriptPolicyRepository interface {
	Create(ctx context.Context, instance *domain.ScriptPolicy) error
	GetByID(ctx context.Context, id string) (*domain.ScriptPolicy, error)
	GetByTemplateID(ctx context.Context, templateID string) ([]domain.ScriptPolicy, error)
	GetGlobalInstances(ctx context.Context) ([]domain.ScriptPolicy, error)
	GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.ScriptPolicy, error)
	GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.ScriptPolicy, error)
	GetEffectiveInstancesByNamespace(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error)
	GetEffectiveInstancesByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.ScriptPolicy, error)
	ListActive(ctx context.Context) ([]domain.ScriptPolicy, error)
	Update(ctx context.Context, instance *domain.ScriptPolicy) error
	Delete(ctx context.Context, id string) error  // Soft delete
	Purge(ctx context.Context, id string) error   // Hard delete
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.ScriptPolicy, int, error)
}

// scriptPolicyRepository implements ScriptPolicyRepository interface using GORM
type scriptPolicyRepository struct {
	db *gorm.DB
}

// NewScriptPolicyRepository creates a new ScriptPolicyRepository instance
func NewScriptPolicyRepository(db *gorm.DB) ScriptPolicyRepository {
	return &scriptPolicyRepository{db: db}
}

// Create inserts a new script policy into the database
func (r *scriptPolicyRepository) Create(ctx context.Context, instance *domain.ScriptPolicy) error {
	model := ToScriptPolicyModel(instance)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fromGormError(err)
	}
	*instance = *model.ToDomain()
	return nil
}

// GetByID retrieves a script policy by its ID
func (r *scriptPolicyRepository) GetByID(ctx context.Context, id string) (*domain.ScriptPolicy, error) {
	var model ScriptPolicyModel
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
func (r *scriptPolicyRepository) GetByTemplateID(ctx context.Context, templateID string) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel
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

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetGlobalInstances retrieves all global-scope instances
func (r *scriptPolicyRepository) GetGlobalInstances(ctx context.Context) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("scope = ?", domain.ScopeGlobal).
		Where("is_active = ?", true).
		Order("priority DESC").
		Find(&models).Error
	if err != nil {
		return nil, fromGormError(err)
	}

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetByNamespaceID retrieves all namespace-level instances for a specific namespace
func (r *scriptPolicyRepository) GetByNamespaceID(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel
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

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetByGroupID retrieves all group-level instances for a specific group
func (r *scriptPolicyRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel
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

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// GetEffectiveInstance finds the most specific active instance for a template
// Priority: Group > Namespace > Global
func (r *scriptPolicyRepository) GetEffectiveInstance(ctx context.Context, templateID, namespaceID, groupID string) (*domain.ScriptPolicy, error) {
	var model ScriptPolicyModel

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
func (r *scriptPolicyRepository) GetEffectiveInstancesByNamespace(ctx context.Context, namespaceID string) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel

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
	var uniqueInstances []domain.ScriptPolicy

	for _, model := range models {
		// Use created_from_template_id for deduplication if available, otherwise use name+checktype
		key := ""
		if model.CreatedFromTemplateID != nil {
			key = *model.CreatedFromTemplateID
		} else {
			key = model.Name + ":" + model.ScriptType
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
func (r *scriptPolicyRepository) GetEffectiveInstancesByGroup(ctx context.Context, namespaceID, groupID string) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel

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
	var uniqueInstances []domain.ScriptPolicy

	for _, model := range models {
		// Use created_from_template_id for deduplication if available, otherwise use name+checktype
		key := ""
		if model.CreatedFromTemplateID != nil {
			key = *model.CreatedFromTemplateID
		} else {
			key = model.Name + ":" + model.ScriptType
		}

		if !seen[key] {
			seen[key] = true
			uniqueInstances = append(uniqueInstances, *model.ToDomain())
		}
	}

	return uniqueInstances, nil
}

// ListActive retrieves all active (non-deleted) instances
func (r *scriptPolicyRepository) ListActive(ctx context.Context) ([]domain.ScriptPolicy, error) {
	var models []ScriptPolicyModel
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

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}
	return instances, nil
}

// Update updates an existing script policy
func (r *scriptPolicyRepository) Update(ctx context.Context, instance *domain.ScriptPolicy) error {
	model := ToScriptPolicyModel(instance)
	return fromGormError(r.db.WithContext(ctx).
		Model(&ScriptPolicyModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error)
}

// Delete performs soft delete on a script policy (sets deleted_at timestamp)
func (r *scriptPolicyRepository) Delete(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).Delete(&ScriptPolicyModel{}, "id = ?", id).Error)
}

// Purge permanently removes a script policy from the database
func (r *scriptPolicyRepository) Purge(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Delete(&ScriptPolicyModel{}, "id = ?", id).Error)
}

// Restore restores a soft-deleted script policy
func (r *scriptPolicyRepository) Restore(ctx context.Context, id string) error {
	return fromGormError(r.db.WithContext(ctx).
		Unscoped().
		Model(&ScriptPolicyModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error)
}

// List retrieves script policys with pagination
func (r *scriptPolicyRepository) List(ctx context.Context, page, limit int) ([]domain.ScriptPolicy, int, error) {
	var models []ScriptPolicyModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ScriptPolicyModel{}).Count(&total).Error; err != nil {
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

	instances := make([]domain.ScriptPolicy, len(models))
	for i, model := range models {
		instances[i] = *model.ToDomain()
	}

	return instances, int(total), nil
}
