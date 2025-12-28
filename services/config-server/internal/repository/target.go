package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/fregataa/aami/config-server/internal/domain"
	"gorm.io/gorm"
)

// TargetModel is the GORM model for database operations
type TargetModel struct {
	ID        string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Hostname  string          `gorm:"not null;uniqueIndex"`
	IPAddress string          `gorm:"not null"`
	Groups    []GroupModel    `gorm:"many2many:target_groups;"`
	Status    string          `gorm:"not null;default:'inactive'"`
	Exporters []ExporterModel `gorm:"foreignKey:TargetID"`
	Labels    StringMap       `gorm:"type:jsonb;default:'{}'"`
	Metadata  JSONB           `gorm:"type:jsonb;default:'{}'"`
	LastSeen  *time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt  `gorm:"index"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (TargetModel) TableName() string {
	return "targets"
}

// StringMap is a custom type for map[string]string JSONB fields
type StringMap map[string]string

// Value implements the driver.Valuer interface
func (sm StringMap) Value() (driver.Value, error) {
	if sm == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(sm)
}

// Scan implements the sql.Scanner interface
func (sm *StringMap) Scan(value interface{}) error {
	if value == nil {
		*sm = make(map[string]string)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sm)
}

// ToTargetModel converts domain.Target to TargetModel
func ToTargetModel(t *domain.Target) *TargetModel {
	model := &TargetModel{
		ID:        t.ID,
		Hostname:  t.Hostname,
		IPAddress: t.IPAddress,
		Status:    string(t.Status),
		Labels:    StringMap(t.Labels),
		Metadata:  JSONB(t.Metadata),
		LastSeen:  t.LastSeen,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *t.DeletedAt, Valid: true}
	}

	// Convert Groups
	if len(t.Groups) > 0 {
		model.Groups = make([]GroupModel, len(t.Groups))
		for i, g := range t.Groups {
			model.Groups[i] = *ToGroupModel(&g)
		}
	}

	return model
}

// ToDomain converts TargetModel to domain.Target
func (m *TargetModel) ToDomain() *domain.Target {
	t := &domain.Target{
		ID:        m.ID,
		Hostname:  m.Hostname,
		IPAddress: m.IPAddress,
		Status:    domain.TargetStatus(m.Status),
		Labels:    map[string]string(m.Labels),
		Metadata:  map[string]interface{}(m.Metadata),
		LastSeen:  m.LastSeen,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.DeletedAt.Valid {
		deletedAt := m.DeletedAt.Time
		t.DeletedAt = &deletedAt
	}

	// Convert Groups
	if len(m.Groups) > 0 {
		t.Groups = make([]domain.Group, len(m.Groups))
		for i, group := range m.Groups {
			t.Groups[i] = *group.ToDomain()
		}
	}

	// Convert Exporters if loaded
	if len(m.Exporters) > 0 {
		t.Exporters = make([]domain.Exporter, len(m.Exporters))
		for i, exporter := range m.Exporters {
			t.Exporters[i] = *exporter.ToDomain()
		}
	}

	return t
}

// TargetRepository defines the interface for target data access
type TargetRepository interface {
	Create(ctx context.Context, target *domain.Target) error
	GetByID(ctx context.Context, id string) (*domain.Target, error)
	GetByHostname(ctx context.Context, hostname string) (*domain.Target, error)
	Update(ctx context.Context, target *domain.Target) error
	Delete(ctx context.Context, id string) error  // Soft delete (sets deleted_at)
	Purge(ctx context.Context, id string) error   // Hard delete (permanent removal)
	Restore(ctx context.Context, id string) error // Restore soft-deleted record
	List(ctx context.Context, page, limit int) ([]domain.Target, int, error)
	GetByGroupID(ctx context.Context, groupID string) ([]domain.Target, error)
	UpdateStatus(ctx context.Context, id string, status domain.TargetStatus) error
	UpdateLastSeen(ctx context.Context, id string, lastSeen time.Time) error
	CountByNamespaceID(ctx context.Context, namespaceID string) (int64, error)
	GetEffectiveCheckInstances(ctx context.Context, targetID string) (*domain.EffectiveCheckInstancesResult, error)
}

// targetRepository implements TargetRepository interface using GORM
type targetRepository struct {
	db *gorm.DB
}

// NewTargetRepository creates a new TargetRepository instance
func NewTargetRepository(db *gorm.DB) TargetRepository {
	return &targetRepository{db: db}
}

// Create inserts a new target into the database
func (r *targetRepository) Create(ctx context.Context, target *domain.Target) error {
	model := ToTargetModel(target)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	*target = *model.ToDomain()
	return nil
}

// GetByID retrieves a target by its ID with all relationships
func (r *targetRepository) GetByID(ctx context.Context, id string) (*domain.Target, error) {
	var model TargetModel
	err := r.db.WithContext(ctx).
		Preload("Groups").
		Preload("Exporters").
		First(&model, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// GetByHostname retrieves a target by its hostname
func (r *targetRepository) GetByHostname(ctx context.Context, hostname string) (*domain.Target, error) {
	var model TargetModel
	err := r.db.WithContext(ctx).
		Preload("Groups").
		Preload("Exporters").
		First(&model, "hostname = ?", hostname).Error
	if err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// Update updates an existing target
func (r *targetRepository) Update(ctx context.Context, target *domain.Target) error {
	model := ToTargetModel(target)
	return r.db.WithContext(ctx).
		Model(&TargetModel{}).
		Where("id = ?", model.ID).
		Updates(model).Error
}

// Delete performs soft delete on a target (sets deleted_at timestamp)
func (r *targetRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TargetModel{}, "id = ?", id).Error
}

// Purge permanently removes a target from the database
func (r *targetRepository) Purge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&TargetModel{}, "id = ?", id).Error
}

// Restore restores a soft-deleted target
func (r *targetRepository) Restore(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&TargetModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// List retrieves targets with pagination
func (r *targetRepository) List(ctx context.Context, page, limit int) ([]domain.Target, int, error) {
	var models []TargetModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&TargetModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Preload("Groups").
		Preload("Exporters").
		Order("hostname ASC").
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	targets := make([]domain.Target, len(models))
	for i, model := range models {
		targets[i] = *model.ToDomain()
	}

	return targets, int(total), nil
}

// GetByGroupID retrieves all targets belonging to a group
func (r *targetRepository) GetByGroupID(ctx context.Context, groupID string) ([]domain.Target, error) {
	var models []TargetModel
	err := r.db.WithContext(ctx).
		Joins("JOIN target_groups ON target_groups.target_id = targets.id").
		Where("target_groups.group_id = ?", groupID).
		Preload("Groups").
		Preload("Exporters").
		Order("hostname ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	targets := make([]domain.Target, len(models))
	for i, model := range models {
		targets[i] = *model.ToDomain()
	}
	return targets, nil
}

// UpdateStatus updates the status of a target
func (r *targetRepository) UpdateStatus(ctx context.Context, id string, status domain.TargetStatus) error {
	return r.db.WithContext(ctx).
		Model(&TargetModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": time.Now(),
		}).Error
}

// UpdateLastSeen updates the last_seen timestamp of a target
func (r *targetRepository) UpdateLastSeen(ctx context.Context, id string, lastSeen time.Time) error {
	return r.db.WithContext(ctx).
		Model(&TargetModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_seen":  lastSeen,
			"updated_at": time.Now(),
		}).Error
}

// CountByNamespaceID counts the number of targets that belong to at least one group in a specific namespace
func (r *targetRepository) CountByNamespaceID(ctx context.Context, namespaceID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TargetModel{}).
		Joins("JOIN target_groups ON target_groups.target_id = targets.id").
		Joins("JOIN groups ON target_groups.group_id = groups.id").
		Where("groups.namespace_id = ?", namespaceID).
		Distinct("targets.id").
		Count(&count).Error
	return count, err
}

// GetEffectiveCheckInstances retrieves all effective check instances for a target
// Returns namespace-level and group-level instances separately, properly sorted by priority
func (r *targetRepository) GetEffectiveCheckInstances(ctx context.Context, targetID string) (*domain.EffectiveCheckInstancesResult, error) {
	// First, get target with all groups sorted by priority
	var target TargetModel
	err := r.db.WithContext(ctx).
		Preload("Groups", func(db *gorm.DB) *gorm.DB {
			return db.Order("priority DESC") // Higher priority number = higher priority
		}).
		First(&target, "id = ?", targetID).Error
	if err != nil {
		return nil, err
	}

	if len(target.Groups) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Collect unique namespace IDs from all groups
	namespaceIDs := make(map[string]bool)
	for _, group := range target.Groups {
		namespaceIDs[group.NamespaceID] = true
	}

	// Get namespace-level instances
	seenKeys := make(map[string]bool)
	namespaceInstances, err := r.getNamespaceInstances(ctx, namespaceIDs, seenKeys)
	if err != nil {
		return nil, err
	}

	// Get group-level instances
	groupInstances, err := r.getGroupInstances(ctx, target.Groups, seenKeys)
	if err != nil {
		return nil, err
	}

	return &domain.EffectiveCheckInstancesResult{
		NamespaceInstances: namespaceInstances,
		GroupInstances:     groupInstances,
	}, nil
}

// getNamespaceInstances retrieves and deduplicates namespace-level instances (global + namespace scope)
func (r *targetRepository) getNamespaceInstances(ctx context.Context, namespaceIDs map[string]bool, seenKeys map[string]bool) ([]domain.CheckInstance, error) {
	var namespaceInstanceModels []CheckInstanceModel

	// Collect instances from all namespaces
	for nsID := range namespaceIDs {
		var instances []CheckInstanceModel
		err := r.db.WithContext(ctx).
			Where("(scope = ? OR (scope = ? AND namespace_id = ?)) AND is_active = ? AND deleted_at IS NULL",
				"global", "namespace", nsID, true).
			Order("scope DESC, priority DESC"). // Group > Namespace > Global, then by priority (higher = higher priority)
			Find(&instances).Error
		if err != nil {
			return nil, err
		}
		namespaceInstanceModels = append(namespaceInstanceModels, instances...)
	}

	// Deduplicate by template_id or name:checktype (keep first = highest priority)
	var uniqueInstances []CheckInstanceModel
	for _, inst := range namespaceInstanceModels {
		key := r.getInstanceKey(inst)
		if !seenKeys[key] {
			uniqueInstances = append(uniqueInstances, inst)
			seenKeys[key] = true
		}
	}

	// Convert to domain objects
	result := make([]domain.CheckInstance, len(uniqueInstances))
	for i, model := range uniqueInstances {
		result[i] = *model.ToDomain()
	}

	return result, nil
}

// getGroupInstances retrieves and deduplicates group-level instances
// Processes groups in priority order and respects already seen instances
func (r *targetRepository) getGroupInstances(ctx context.Context, groups []GroupModel, seenKeys map[string]bool) ([]domain.CheckInstance, error) {
	var groupInstanceModels []CheckInstanceModel

	// Process groups in priority order (already sorted)
	for _, group := range groups {
		var instances []CheckInstanceModel
		err := r.db.WithContext(ctx).
			Where("(scope = ? OR (scope = ? AND namespace_id = ?) OR (scope = ? AND group_id = ?)) AND is_active = ? AND deleted_at IS NULL",
				"global", "namespace", group.NamespaceID, "group", group.ID, true).
			Order("scope DESC, priority DESC"). // Group > Namespace > Global, then by priority (higher = higher priority)
			Find(&instances).Error
		if err != nil {
			return nil, err
		}

		// Add to results with deduplication
		for _, inst := range instances {
			key := r.getInstanceKey(inst)
			// Only add if not seen yet (first group with this check wins due to priority order)
			if !seenKeys[key] {
				groupInstanceModels = append(groupInstanceModels, inst)
				seenKeys[key] = true
			}
		}
	}

	// Convert to domain objects
	result := make([]domain.CheckInstance, len(groupInstanceModels))
	for i, model := range groupInstanceModels {
		result[i] = *model.ToDomain()
	}

	return result, nil
}

// getInstanceKey generates a unique key for deduplication
// Uses template ID if available, otherwise uses name:checktype
func (r *targetRepository) getInstanceKey(inst CheckInstanceModel) string {
	if inst.CreatedFromTemplateID != nil {
		return *inst.CreatedFromTemplateID
	}
	return inst.Name + ":" + inst.CheckType
}
