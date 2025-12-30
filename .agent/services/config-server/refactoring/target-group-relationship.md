# Target-Group Relationship Refactoring

**Created**: 2025-12-28
**Status**: Planning
**Related Issue**: Primary/Secondary Group complexity and check collection bug

## Overview

This refactoring replaces the Primary/Secondary Group pattern with a unified junction table approach and adds automatic Default Own Group creation for orphan prevention.

**Current Problems**:
1. `GetEffectiveChecksByHostname` only checks Primary Group, ignoring Secondary Groups
2. Primary/Secondary distinction confuses users
3. Difficult to manage which group is "primary" when groups change
4. Orphan prevention using `ON DELETE RESTRICT` creates poor UX

**Solution**:
- Remove `primary_group_id` column from targets table
- Use `target_groups` junction table for ALL group relationships
- Auto-create Default Own Group for each target
- Enforce "minimum 1 group" constraint at application level
- Collect checks from ALL target groups equally

---

## Phase 1: Domain Model Changes

### 1.1 Target Domain Model

**File**: `internal/domain/target.go`

**Changes**:

```go
// BEFORE
type Target struct {
    ID              string    `json:"id"`
    Hostname        string    `json:"hostname"`
    IPAddress       string    `json:"ip_address"`
    PrimaryGroupID  string    `json:"primary_group_id"`   // Remove
    PrimaryGroup    Group     `json:"primary_group"`      // Remove
    SecondaryGroups []Group   `json:"secondary_groups"`   // Remove
    Status          TargetStatus `json:"status"`
    Labels          map[string]string `json:"labels"`
    Metadata        map[string]interface{} `json:"metadata"`
    LastSeen        *time.Time `json:"last_seen,omitempty"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// AFTER
type Target struct {
    ID        string    `json:"id"`
    Hostname  string    `json:"hostname"`
    IPAddress string    `json:"ip_address"`
    Groups    []Group   `json:"groups"`  // New: ALL groups (no primary/secondary)
    Status    TargetStatus `json:"status"`
    Labels    map[string]string `json:"labels"`
    Metadata  map[string]interface{} `json:"metadata"`
    LastSeen  *time.Time `json:"last_seen,omitempty"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// New helper methods
func (t *Target) GetGroupIDs() []string {
    ids := make([]string, len(t.Groups))
    for i, g := range t.Groups {
        ids[i] = g.ID
    }
    return ids
}

func (t *Target) HasDefaultOwnGroup() bool {
    for _, g := range t.Groups {
        if g.IsDefaultOwn {
            return true
        }
    }
    return false
}
```

### 1.2 Group Domain Model

**File**: `internal/domain/group.go`

**Add field**:

```go
type Group struct {
    ID             string                 `json:"id"`
    Name           string                 `json:"name"`
    NamespaceID    string                 `json:"namespace_id"`
    ParentID       *string                `json:"parent_id,omitempty"`
    Description    string                 `json:"description"`
    Priority       int                    `json:"priority"`
    Metadata       map[string]interface{} `json:"metadata"`
    IsDefaultOwn   bool                   `json:"is_default_own"`  // NEW
    DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
    CreatedAt      time.Time              `json:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at"`
}
```

### 1.3 New Domain Model: TargetGroup

**File**: `internal/domain/target_group.go` (NEW)

```go
package domain

import "time"

// TargetGroup represents the many-to-many relationship between targets and groups
type TargetGroup struct {
    TargetID     string    `json:"target_id"`
    GroupID      string    `json:"group_id"`
    IsDefaultOwn bool      `json:"is_default_own"`  // True if this is the auto-created default group
    CreatedAt    time.Time `json:"created_at"`
}

// Validation
func (tg *TargetGroup) Validate() error {
    if tg.TargetID == "" {
        return errors.New("target_id is required")
    }
    if tg.GroupID == "" {
        return errors.New("group_id is required")
    }
    return nil
}
```

---

## Phase 2: Repository Layer Changes

### 2.1 Group Repository ORM Model

**File**: `internal/repository/group.go`

**Update ORM model**:

```go
type GroupModel struct {
    ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Name         string         `gorm:"not null;type:varchar(255)"`
    NamespaceID  string         `gorm:"not null;type:uuid;index"`
    ParentID     *string        `gorm:"type:uuid;index"`
    Description  string         `gorm:"type:text"`
    Priority     int            `gorm:"not null;default:100"`
    Metadata     JSONB          `gorm:"type:jsonb;default:'{}'"`
    IsDefaultOwn bool           `gorm:"not null;default:false;index"` // NEW
    DeletedAt    gorm.DeletedAt `gorm:"index"`
    CreatedAt    time.Time      `gorm:"autoCreateTime"`
    UpdatedAt    time.Time      `gorm:"autoUpdateTime"`

    Namespace NamespaceModel `gorm:"foreignKey:NamespaceID"`
    Parent    *GroupModel    `gorm:"foreignKey:ParentID"`
}
```

**Update converters**:

```go
func ToGroupModel(d *domain.Group) *GroupModel {
    m := &GroupModel{
        ID:           d.ID,
        Name:         d.Name,
        NamespaceID:  d.NamespaceID,
        ParentID:     d.ParentID,
        Description:  d.Description,
        Priority:     d.Priority,
        Metadata:     JSONB(d.Metadata),
        IsDefaultOwn: d.IsDefaultOwn,  // NEW
        CreatedAt:    d.CreatedAt,
        UpdatedAt:    d.UpdatedAt,
    }
    if d.DeletedAt != nil {
        m.DeletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
    }
    return m
}

func (m *GroupModel) ToDomain() *domain.Group {
    d := &domain.Group{
        ID:           m.ID,
        Name:         m.Name,
        NamespaceID:  m.NamespaceID,
        ParentID:     m.ParentID,
        Description:  m.Description,
        Priority:     m.Priority,
        Metadata:     m.Metadata,
        IsDefaultOwn: m.IsDefaultOwn,  // NEW
        CreatedAt:    m.CreatedAt,
        UpdatedAt:    m.UpdatedAt,
    }
    if m.DeletedAt.Valid {
        deletedAt := m.DeletedAt.Time
        d.DeletedAt = &deletedAt
    }
    return d
}
```

### 2.2 New Repository: TargetGroup

**File**: `internal/repository/target_group.go` (NEW)

```go
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
        return err
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

    return r.db.WithContext(ctx).Create(&models).Error
}

func (r *targetGroupRepository) GetByTarget(ctx context.Context, targetID string) ([]domain.TargetGroup, error) {
    var models []TargetGroupModel
    err := r.db.WithContext(ctx).
        Where("target_id = ?", targetID).
        Find(&models).Error
    if err != nil {
        return nil, err
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
        return nil, err
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
    return count, err
}

func (r *targetGroupRepository) CountByGroup(ctx context.Context, groupID string) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&TargetGroupModel{}).
        Where("group_id = ?", groupID).
        Count(&count).Error
    return count, err
}

func (r *targetGroupRepository) Delete(ctx context.Context, targetID, groupID string) error {
    return r.db.WithContext(ctx).
        Delete(&TargetGroupModel{}, "target_id = ? AND group_id = ?", targetID, groupID).
        Error
}

func (r *targetGroupRepository) DeleteByTarget(ctx context.Context, targetID string) error {
    return r.db.WithContext(ctx).
        Delete(&TargetGroupModel{}, "target_id = ?", targetID).
        Error
}

func (r *targetGroupRepository) DeleteByGroup(ctx context.Context, groupID string) error {
    return r.db.WithContext(ctx).
        Delete(&TargetGroupModel{}, "group_id = ?", groupID).
        Error
}

func (r *targetGroupRepository) Exists(ctx context.Context, targetID, groupID string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&TargetGroupModel{}).
        Where("target_id = ? AND group_id = ?", targetID, groupID).
        Count(&count).Error
    return count > 0, err
}
```

### 2.3 Target Repository Updates

**File**: `internal/repository/target.go`

**Update ORM model**:

```go
type TargetModel struct {
    ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Hostname  string         `gorm:"not null;unique;index"`
    IPAddress string         `gorm:"not null;type:varchar(45)"`
    // Remove: PrimaryGroupID string `gorm:"not null;type:uuid;index"`
    Status    string         `gorm:"type:varchar(20);not null;default:'active'"`
    Labels    JSONB          `gorm:"type:jsonb;default:'{}'"`
    Metadata  JSONB          `gorm:"type:jsonb;default:'{}'"`
    LastSeen  *time.Time     `gorm:"type:timestamp"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`

    // Remove: PrimaryGroup    GroupModel   `gorm:"foreignKey:PrimaryGroupID"`
    // Remove: SecondaryGroups []GroupModel `gorm:"many2many:target_secondary_groups"`
    Groups []GroupModel `gorm:"many2many:target_groups"` // NEW: unified relationship
}
```

**Update GetByID to preload groups**:

```go
func (r *targetRepository) GetByID(ctx context.Context, id string) (*domain.Target, error) {
    var model TargetModel
    err := r.db.WithContext(ctx).
        Preload("Groups").  // Preload all groups
        First(&model, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return model.ToDomain(), nil
}
```

**Update List to preload groups**:

```go
func (r *targetRepository) List(ctx context.Context, page, limit int) ([]domain.Target, int, error) {
    var models []TargetModel
    var total int64

    // Count
    if err := r.db.WithContext(ctx).Model(&TargetModel{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Fetch with pagination and preload
    offset := (page - 1) * limit
    err := r.db.WithContext(ctx).
        Preload("Groups").
        Limit(limit).
        Offset(offset).
        Find(&models).Error
    if err != nil {
        return nil, 0, err
    }

    results := make([]domain.Target, len(models))
    for i, m := range models {
        results[i] = *m.ToDomain()
    }

    return results, int(total), nil
}
```

**Update converters**:

```go
func ToTargetModel(d *domain.Target) *TargetModel {
    m := &TargetModel{
        ID:        d.ID,
        Hostname:  d.Hostname,
        IPAddress: d.IPAddress,
        Status:    string(d.Status),
        Labels:    JSONB(d.Labels),
        Metadata:  JSONB(d.Metadata),
        LastSeen:  d.LastSeen,
        CreatedAt: d.CreatedAt,
        UpdatedAt: d.UpdatedAt,
    }

    // Convert groups
    if len(d.Groups) > 0 {
        m.Groups = make([]GroupModel, len(d.Groups))
        for i, g := range d.Groups {
            m.Groups[i] = *ToGroupModel(&g)
        }
    }

    if d.DeletedAt != nil {
        m.DeletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
    }
    return m
}

func (m *TargetModel) ToDomain() *domain.Target {
    d := &domain.Target{
        ID:        m.ID,
        Hostname:  m.Hostname,
        IPAddress: m.IPAddress,
        Status:    domain.TargetStatus(m.Status),
        Labels:    m.Labels,
        Metadata:  m.Metadata,
        LastSeen:  m.LastSeen,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }

    // Convert groups
    if len(m.Groups) > 0 {
        d.Groups = make([]domain.Group, len(m.Groups))
        for i, g := range m.Groups {
            d.Groups[i] = *g.ToDomain()
        }
    }

    if m.DeletedAt.Valid {
        deletedAt := m.DeletedAt.Time
        d.DeletedAt = &deletedAt
    }
    return d
}
```

---

## Phase 3: Service Layer Changes

### 3.1 Target Service Updates

**File**: `internal/service/target.go`

**Update service struct**:

```go
type TargetService struct {
    targetRepo      repository.TargetRepository
    targetGroupRepo repository.TargetGroupRepository  // NEW
    groupRepo       repository.GroupRepository
    exporterRepo    repository.ExporterRepository
}

func NewTargetService(
    targetRepo repository.TargetRepository,
    targetGroupRepo repository.TargetGroupRepository,  // NEW
    groupRepo repository.GroupRepository,
    exporterRepo repository.ExporterRepository,
) *TargetService {
    return &TargetService{
        targetRepo:      targetRepo,
        targetGroupRepo: targetGroupRepo,
        groupRepo:       groupRepo,
        exporterRepo:    exporterRepo,
    }
}
```

**Update Create method**:

```go
func (s *TargetService) Create(ctx context.Context, req dto.CreateTargetRequest) (*domain.Target, error) {
    // Validate hostname uniqueness
    existing, err := s.targetRepo.GetByHostname(ctx, req.Hostname)
    if err == nil && existing != nil {
        return nil, ErrAlreadyExists
    }

    var groupIDs []string
    var shouldCreateDefaultOwn bool

    if len(req.GroupIDs) == 0 {
        // Case A: No groups provided - create Default Own Group
        shouldCreateDefaultOwn = true
    } else {
        // Case B: Groups provided - validate all exist
        for _, gid := range req.GroupIDs {
            _, err := s.groupRepo.GetByID(ctx, gid)
            if err != nil {
                if errors.Is(err, gorm.ErrRecordNotFound) {
                    return nil, ErrForeignKeyViolation
                }
                return nil, err
            }
        }
        groupIDs = req.GroupIDs
    }

    // Start transaction
    return s.targetRepo.Transaction(ctx, func(txCtx context.Context) (*domain.Target, error) {
        // 1. Create target
        target := &domain.Target{
            ID:        uuid.New().String(),
            Hostname:  req.Hostname,
            IPAddress: req.IPAddress,
            Status:    domain.TargetStatusActive,
            Labels:    req.Labels,
            Metadata:  req.Metadata,
        }

        if err := s.targetRepo.Create(txCtx, target); err != nil {
            return nil, err
        }

        // 2. Create Default Own Group if needed
        if shouldCreateDefaultOwn {
            // Find or create default namespace
            namespace, err := s.getOrCreateDefaultNamespace(txCtx)
            if err != nil {
                return nil, fmt.Errorf("failed to get default namespace: %w", err)
            }

            // Create default own group
            defaultGroup := &domain.Group{
                ID:           uuid.New().String(),
                Name:         fmt.Sprintf("target-%s", target.Hostname),
                NamespaceID:  namespace.ID,
                Description:  fmt.Sprintf("Default group for target %s", target.Hostname),
                Priority:     100,
                IsDefaultOwn: true,
                Metadata:     make(map[string]interface{}),
            }

            if err := s.groupRepo.Create(txCtx, defaultGroup); err != nil {
                return nil, fmt.Errorf("failed to create default group: %w", err)
            }

            groupIDs = []string{defaultGroup.ID}
        }

        // 3. Create target-group mappings
        mappings := make([]domain.TargetGroup, len(groupIDs))
        for i, gid := range groupIDs {
            mappings[i] = domain.TargetGroup{
                TargetID:     target.ID,
                GroupID:      gid,
                IsDefaultOwn: shouldCreateDefaultOwn && i == 0,
            }
        }

        if err := s.targetGroupRepo.CreateBatch(txCtx, mappings); err != nil {
            return nil, fmt.Errorf("failed to create group mappings: %w", err)
        }

        // 4. Load target with groups
        return s.targetRepo.GetByID(txCtx, target.ID)
    })
}

func (s *TargetService) getOrCreateDefaultNamespace(ctx context.Context) (*domain.Namespace, error) {
    const defaultNamespaceName = "default"

    ns, err := s.namespaceRepo.GetByName(ctx, defaultNamespaceName)
    if err == nil {
        return ns, nil
    }

    if !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }

    // Create default namespace
    newNS := &domain.Namespace{
        ID:             uuid.New().String(),
        Name:           defaultNamespaceName,
        Description:    "Default namespace for auto-created groups",
        PolicyPriority: 100,
        MergeStrategy:  domain.MergeStrategyMerge,
    }

    if err := s.namespaceRepo.Create(ctx, newNS); err != nil {
        return nil, err
    }

    return newNS, nil
}
```

**Add new methods for group management**:

```go
// AddGroupMapping adds a target to a group
func (s *TargetService) AddGroupMapping(ctx context.Context, targetID, groupID string) error {
    // Validate target exists
    _, err := s.targetRepo.GetByID(ctx, targetID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrNotFound
        }
        return err
    }

    // Validate group exists
    _, err = s.groupRepo.GetByID(ctx, groupID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrForeignKeyViolation
        }
        return err
    }

    // Check if mapping already exists
    exists, err := s.targetGroupRepo.Exists(ctx, targetID, groupID)
    if err != nil {
        return err
    }
    if exists {
        return ErrAlreadyExists
    }

    // Create mapping
    mapping := &domain.TargetGroup{
        TargetID:     targetID,
        GroupID:      groupID,
        IsDefaultOwn: false,
    }

    return s.targetGroupRepo.Create(ctx, mapping)
}

// RemoveGroupMapping removes a target from a group
func (s *TargetService) RemoveGroupMapping(ctx context.Context, targetID, groupID string) error {
    // Check if target exists
    _, err := s.targetRepo.GetByID(ctx, targetID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrNotFound
        }
        return err
    }

    // Count existing mappings
    count, err := s.targetGroupRepo.CountByTarget(ctx, targetID)
    if err != nil {
        return err
    }

    // Prevent removal of last group
    if count <= 1 {
        return ErrCannotRemoveLastGroup
    }

    // Delete mapping
    return s.targetGroupRepo.Delete(ctx, targetID, groupID)
}

// GetTargetGroups retrieves all groups for a target
func (s *TargetService) GetTargetGroups(ctx context.Context, targetID string) ([]domain.Group, error) {
    // Validate target exists
    target, err := s.targetRepo.GetByID(ctx, targetID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    return target.Groups, nil
}

// ReplaceGroupMappings replaces all group mappings for a target
func (s *TargetService) ReplaceGroupMappings(ctx context.Context, targetID string, groupIDs []string) error {
    if len(groupIDs) == 0 {
        return errors.New("at least one group is required")
    }

    // Validate target exists
    _, err := s.targetRepo.GetByID(ctx, targetID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrNotFound
        }
        return err
    }

    // Validate all groups exist
    for _, gid := range groupIDs {
        _, err := s.groupRepo.GetByID(ctx, gid)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return ErrForeignKeyViolation
            }
            return err
        }
    }

    // Transaction: delete old mappings and create new ones
    return s.targetRepo.Transaction(ctx, func(txCtx context.Context) error {
        // Delete all existing mappings
        if err := s.targetGroupRepo.DeleteByTarget(txCtx, targetID); err != nil {
            return err
        }

        // Create new mappings
        mappings := make([]domain.TargetGroup, len(groupIDs))
        for i, gid := range groupIDs {
            mappings[i] = domain.TargetGroup{
                TargetID:     targetID,
                GroupID:      gid,
                IsDefaultOwn: false,
            }
        }

        return s.targetGroupRepo.CreateBatch(txCtx, mappings)
    })
}
```

**Add new error**:

```go
var (
    // ... existing errors
    ErrCannotRemoveLastGroup = errors.New("cannot remove last group from target")
)
```

### 3.2 CheckInstance Service Updates

**File**: `internal/service/check_instance.go`

**Fix GetEffectiveChecksByHostname**:

```go
func (s *CheckInstanceService) GetEffectiveChecksByHostname(ctx context.Context, hostname string) ([]domain.CheckInstance, error) {
    // 1. Find target
    target, err := s.targetRepo.GetByHostname(ctx, hostname)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    // Load target with all groups
    target, err = s.targetRepo.GetByID(ctx, target.ID)
    if err != nil {
        return nil, err
    }

    if len(target.Groups) == 0 {
        return nil, errors.New("target has no groups")
    }

    // 2. Collect instances from ALL groups
    var allInstances []domain.CheckInstance
    seenTemplateIDs := make(map[string]bool)

    // Process each group
    for _, group := range target.Groups {
        // Get instances for this group
        groupInstances, err := s.instanceRepo.GetEffectiveInstancesByGroup(
            ctx,
            group.NamespaceID,
            group.ID,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to get instances for group %s: %w", group.ID, err)
        }

        // Add to results, deduplicating by created_from_template_id or name:checktype
        for _, inst := range groupInstances {
            key := ""
            if inst.CreatedFromTemplateID != nil {
                key = *inst.CreatedFromTemplateID
            } else {
                key = inst.Name + ":" + inst.CheckType
            }

            if !seenTemplateIDs[key] {
                allInstances = append(allInstances, inst)
                seenTemplateIDs[key] = true
            }
        }
    }

    return allInstances, nil
}
```

---

## Phase 4: Database Migration

**File**: `migrations/009_refactor_target_group_relationship.sql` (NEW)

```sql
-- Migration 009: Refactor Target-Group Relationship
-- Replace Primary/Secondary pattern with unified junction table + Default Own Group

-- Step 1: Add is_default_own column to groups table
ALTER TABLE groups
    ADD COLUMN is_default_own BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_groups_is_default_own ON groups(is_default_own);

COMMENT ON COLUMN groups.is_default_own IS 'True if this is an auto-created default group for a target';

-- Step 2: Create unified target_groups junction table (if not exists)
CREATE TABLE IF NOT EXISTS target_groups (
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    is_default_own BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (target_id, group_id)
);

CREATE INDEX idx_target_groups_target_id ON target_groups(target_id);
CREATE INDEX idx_target_groups_group_id ON target_groups(group_id);
CREATE INDEX idx_target_groups_is_default_own ON target_groups(is_default_own);

COMMENT ON TABLE target_groups IS 'Junction table for many-to-many relationship between targets and groups';
COMMENT ON COLUMN target_groups.target_id IS 'Reference to target';
COMMENT ON COLUMN target_groups.group_id IS 'Reference to group';
COMMENT ON COLUMN target_groups.is_default_own IS 'True if this is the auto-created default group mapping';

-- Step 3: Create database trigger to prevent orphan targets
CREATE OR REPLACE FUNCTION prevent_target_orphan()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM target_groups WHERE target_id = OLD.target_id) <= 1 THEN
        RAISE EXCEPTION 'Cannot remove last group mapping for target (target_id: %)', OLD.target_id;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_prevent_target_orphan
    BEFORE DELETE ON target_groups
    FOR EACH ROW
    EXECUTE FUNCTION prevent_target_orphan();

-- Step 4: Migrate data from primary_group_id to target_groups
-- Note: This assumes no data exists yet. If data exists, uncomment and modify:
/*
INSERT INTO target_groups (target_id, group_id, is_default_own, created_at)
SELECT
    id as target_id,
    primary_group_id as group_id,
    false as is_default_own,
    created_at
FROM targets
ON CONFLICT (target_id, group_id) DO NOTHING;

-- Migrate secondary groups
INSERT INTO target_groups (target_id, group_id, is_default_own, created_at)
SELECT
    target_id,
    group_id,
    false as is_default_own,
    CURRENT_TIMESTAMP
FROM target_secondary_groups
ON CONFLICT (target_id, group_id) DO NOTHING;
*/

-- Step 5: Drop old tables and columns
DROP TABLE IF EXISTS target_secondary_groups;

-- Drop foreign key constraint on primary_group_id
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'targets'::regclass
    AND contype = 'f'
    AND confrelid = 'groups'::regclass
    AND conkey = ARRAY[(SELECT attnum FROM pg_attribute
                        WHERE attrelid = 'targets'::regclass
                        AND attname = 'primary_group_id')];

    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE targets DROP CONSTRAINT ' || constraint_name;
    END IF;
END $$;

-- Drop primary_group_id column
ALTER TABLE targets DROP COLUMN IF EXISTS primary_group_id;

-- Step 6: Update table comments
COMMENT ON TABLE targets IS 'Monitoring targets (servers, services). Groups are managed via target_groups junction table.';
```

**Update**: `cmd/config-server/main.go`

```go
migrations := []string{
    "migrations/001_initial_schema.sql",
    "migrations/002_refactor_namespace_to_table.sql",
    "migrations/003_add_soft_delete.sql",
    "migrations/004_add_check_templates.sql",
    "migrations/005_add_check_instances.sql",
    "migrations/006_migrate_check_settings_to_instances.sql",
    "migrations/007_decouple_check_instances_from_templates.sql",
    "migrations/008_decouple_alert_rules_from_templates.sql",
    "migrations/009_refactor_target_group_relationship.sql",  // NEW
}
```

---

## Phase 5: API Layer Changes

### 5.1 DTO Updates

**File**: `internal/api/dto/target.go`

**Update request DTOs**:

```go
type CreateTargetRequest struct {
    Hostname  string                 `json:"hostname" binding:"required,min=1,max=255"`
    IPAddress string                 `json:"ip_address" binding:"required,ip"`
    GroupIDs  []string               `json:"group_ids,omitempty"` // NEW: optional, if empty creates default own group
    Labels    map[string]string      `json:"labels,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateTargetRequest struct {
    IPAddress *string                 `json:"ip_address,omitempty" binding:"omitempty,ip"`
    Status    *domain.TargetStatus    `json:"status,omitempty"`
    Labels    map[string]string       `json:"labels,omitempty"`
    Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

// NEW: DTO for managing group mappings
type UpdateTargetGroupsRequest struct {
    GroupIDs []string `json:"group_ids" binding:"required,min=1"`
}

type AddGroupMappingRequest struct {
    TargetID string `json:"target_id" binding:"required,uuid"`
    GroupID  string `json:"group_id" binding:"required,uuid"`
}

type RemoveGroupMappingRequest struct {
    TargetID string `json:"target_id" binding:"required,uuid"`
    GroupID  string `json:"group_id" binding:"required,uuid"`
}
```

**Update response DTOs**:

```go
type TargetResponse struct {
    ID        string                 `json:"id"`
    Hostname  string                 `json:"hostname"`
    IPAddress string                 `json:"ip_address"`
    Groups    []GroupResponse        `json:"groups"`  // NEW: all groups
    Status    domain.TargetStatus    `json:"status"`
    Labels    map[string]string      `json:"labels"`
    Metadata  map[string]interface{} `json:"metadata"`
    LastSeen  *time.Time             `json:"last_seen,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}

func ToTargetResponse(t *domain.Target) TargetResponse {
    resp := TargetResponse{
        ID:        t.ID,
        Hostname:  t.Hostname,
        IPAddress: t.IPAddress,
        Status:    t.Status,
        Labels:    t.Labels,
        Metadata:  t.Metadata,
        LastSeen:  t.LastSeen,
        CreatedAt: t.CreatedAt,
        UpdatedAt: t.UpdatedAt,
    }

    // Convert groups
    if len(t.Groups) > 0 {
        resp.Groups = make([]GroupResponse, len(t.Groups))
        for i, g := range t.Groups {
            resp.Groups[i] = ToGroupResponse(&g)
        }
    }

    return resp
}
```

### 5.2 Handler Updates

**File**: `internal/api/handler/target.go`

**Add new handlers**:

```go
// GetTargetGroups retrieves all groups for a target
func (h *TargetHandler) GetTargetGroups(c *gin.Context) {
    targetID := c.Param("id")

    groups, err := h.targetService.GetTargetGroups(c.Request.Context(), targetID)
    if err != nil {
        respondError(c, err)
        return
    }

    responses := make([]dto.GroupResponse, len(groups))
    for i, g := range groups {
        responses[i] = dto.ToGroupResponse(&g)
    }

    c.JSON(http.StatusOK, responses)
}

// UpdateTargetGroups replaces all group mappings for a target
func (h *TargetHandler) UpdateTargetGroups(c *gin.Context) {
    targetID := c.Param("id")

    var req dto.UpdateTargetGroupsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: err.Error(),
            Code:  "INVALID_REQUEST",
        })
        return
    }

    if err := h.targetService.ReplaceGroupMappings(c.Request.Context(), targetID, req.GroupIDs); err != nil {
        respondError(c, err)
        return
    }

    c.Status(http.StatusNoContent)
}

// AddGroupMapping adds a target to a group
func (h *TargetHandler) AddGroupMapping(c *gin.Context) {
    var req dto.AddGroupMappingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: err.Error(),
            Code:  "INVALID_REQUEST",
        })
        return
    }

    if err := h.targetService.AddGroupMapping(c.Request.Context(), req.TargetID, req.GroupID); err != nil {
        respondError(c, err)
        return
    }

    c.Status(http.StatusCreated)
}

// RemoveGroupMapping removes a target from a group
func (h *TargetHandler) RemoveGroupMapping(c *gin.Context) {
    var req dto.RemoveGroupMappingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: err.Error(),
            Code:  "INVALID_REQUEST",
        })
        return
    }

    if err := h.targetService.RemoveGroupMapping(c.Request.Context(), req.TargetID, req.GroupID); err != nil {
        respondError(c, err)
        return
    }

    c.Status(http.StatusNoContent)
}
```

**Update error handler**:

```go
func respondError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, service.ErrNotFound):
        c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error(), Code: "NOT_FOUND"})
    case errors.Is(err, service.ErrAlreadyExists):
        c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error(), Code: "CONFLICT"})
    case errors.Is(err, service.ErrCannotRemoveLastGroup):  // NEW
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Code: "CANNOT_REMOVE_LAST_GROUP"})
    case errors.Is(err, service.ErrForeignKeyViolation):
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Code: "FOREIGN_KEY_VIOLATION"})
    case errors.Is(err, gorm.ErrRecordNotFound):
        c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Not found", Code: "NOT_FOUND"})
    default:
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error(), Code: "INTERNAL_ERROR"})
    }
}
```

### 5.3 Router Updates

**File**: `internal/api/router.go`

```go
func (s *Server) SetupRouter() *gin.Engine {
    router := gin.Default()

    v1 := router.Group("/api/v1")
    {
        // Targets
        targets := v1.Group("/targets")
        {
            targets.POST("", s.targetHandler.Create)
            targets.GET("", s.targetHandler.List)
            targets.GET("/:id", s.targetHandler.GetByID)
            targets.PUT("/:id", s.targetHandler.Update)
            targets.POST("/delete", s.targetHandler.Delete)
            targets.POST("/purge", s.targetHandler.Purge)
            targets.POST("/restore", s.targetHandler.Restore)

            // NEW: Group management endpoints
            targets.GET("/:id/groups", s.targetHandler.GetTargetGroups)
            targets.PUT("/:id/groups", s.targetHandler.UpdateTargetGroups)
            targets.POST("/groups/add", s.targetHandler.AddGroupMapping)
            targets.POST("/groups/remove", s.targetHandler.RemoveGroupMapping)
        }

        // ... other routes
    }

    return router
}
```

---

## Phase 6: Repository Manager Updates

**File**: `internal/repository/manager.go`

**Update Manager struct**:

```go
type Manager struct {
    db                    *gorm.DB
    namespaceRepo         NamespaceRepository
    groupRepo             GroupRepository
    targetRepo            TargetRepository
    targetGroupRepo       TargetGroupRepository  // NEW
    exporterRepo          ExporterRepository
    bootstrapTokenRepo    BootstrapTokenRepository
    checkTemplateRepo     CheckTemplateRepository
    checkInstanceRepo     CheckInstanceRepository
    alertTemplateRepo     AlertTemplateRepository
    alertRuleRepo         AlertRuleRepository
}

func NewManager(config Config) (*Manager, error) {
    db, err := initDB(config)
    if err != nil {
        return nil, err
    }

    return &Manager{
        db:                 db,
        namespaceRepo:      NewNamespaceRepository(db),
        groupRepo:          NewGroupRepository(db),
        targetRepo:         NewTargetRepository(db),
        targetGroupRepo:    NewTargetGroupRepository(db),  // NEW
        exporterRepo:       NewExporterRepository(db),
        bootstrapTokenRepo: NewBootstrapTokenRepository(db),
        checkTemplateRepo:  NewCheckTemplateRepository(db),
        checkInstanceRepo:  NewCheckInstanceRepository(db),
        alertTemplateRepo:  NewAlertTemplateRepository(db),
        alertRuleRepo:      NewAlertRuleRepository(db),
    }, nil
}

// Add getter
func (m *Manager) TargetGroupRepo() TargetGroupRepository {
    return m.targetGroupRepo
}
```

---

## Phase 7: API Server Updates

**File**: `internal/api/server.go`

**Update NewServer**:

```go
func NewServer(rm *repository.Manager) *Server {
    // ... existing repos

    // Services
    namespaceService := service.NewNamespaceService(rm.NamespaceRepo())
    groupService := service.NewGroupService(rm.GroupRepo(), rm.NamespaceRepo())
    targetService := service.NewTargetService(
        rm.TargetRepo(),
        rm.TargetGroupRepo(),  // NEW
        rm.GroupRepo(),
        rm.ExporterRepo(),
    )
    exporterService := service.NewExporterService(rm.ExporterRepo(), rm.TargetRepo())
    bootstrapTokenService := service.NewBootstrapTokenService(rm.BootstrapTokenRepo(), rm.GroupRepo())
    checkTemplateService := service.NewCheckTemplateService(rm.CheckTemplateRepo())
    checkInstanceService := service.NewCheckInstanceService(
        rm.CheckInstanceRepo(),
        rm.CheckTemplateRepo(),
        rm.TargetRepo(),
        rm.GroupRepo(),
        rm.NamespaceRepo(),
    )
    alertTemplateService := service.NewAlertTemplateService(rm.AlertTemplateRepo())
    alertRuleService := service.NewAlertRuleService(
        rm.AlertRuleRepo(),
        rm.AlertTemplateRepo(),
        rm.GroupRepo(),
    )

    // ... rest unchanged
}
```

---

## Phase 8: Test Updates

### 8.1 Test Fixtures

**File**: `test/testutil/fixtures.go`

**Update NewTestTarget**:

```go
// Remove: NewTestTarget(hostname, ipAddress, primaryGroupID)
// Add: NewTestTarget(hostname, ipAddress, groups)
func NewTestTarget(hostname string, ipAddress string, groups []domain.Group) *domain.Target {
    return &domain.Target{
        ID:        uuid.New().String(),
        Hostname:  hostname,
        IPAddress: ipAddress,
        Groups:    groups,  // NEW
        Status:    domain.TargetStatusActive,
        Labels:    make(map[string]string),
        Metadata:  make(map[string]interface{}),
        LastSeen:  nil,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// Add helper for creating target with default own group
func NewTestTargetWithDefaultGroup(hostname string, ipAddress string, namespace *domain.Namespace) *domain.Target {
    defaultGroup := &domain.Group{
        ID:           uuid.New().String(),
        Name:         fmt.Sprintf("target-%s", hostname),
        NamespaceID:  namespace.ID,
        Description:  fmt.Sprintf("Default group for %s", hostname),
        Priority:     100,
        IsDefaultOwn: true,
        Metadata:     make(map[string]interface{}),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    return &domain.Target{
        ID:        uuid.New().String(),
        Hostname:  hostname,
        IPAddress: ipAddress,
        Groups:    []domain.Group{*defaultGroup},
        Status:    domain.TargetStatusActive,
        Labels:    make(map[string]string),
        Metadata:  make(map[string]interface{}),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}
```

**Update NewTestGroup**:

```go
func NewTestGroup(name string, namespaceID string) *domain.Group {
    return &domain.Group{
        ID:           uuid.New().String(),
        Name:         name,
        NamespaceID:  namespaceID,
        Description:  "Test group: " + name,
        Priority:     100,
        IsDefaultOwn: false,  // NEW
        Metadata:     make(map[string]interface{}),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
}
```

---

## Implementation Order

1. ✅ **Planning Document** (this file)
2. **Domain Layer** (`internal/domain/`)
   - Update `target.go`
   - Update `group.go`
   - Create `target_group.go`
3. **Repository Layer** (`internal/repository/`)
   - Update `group.go` (ORM model + converters)
   - Create `target_group.go` (new repository)
   - Update `target.go` (ORM model + converters + queries)
4. **Database Migration**
   - Create `migrations/009_refactor_target_group_relationship.sql`
   - Update `cmd/config-server/main.go`
5. **Service Layer** (`internal/service/`)
   - Update `target.go` (Create method + new group management methods)
   - Update `check_instance.go` (GetEffectiveChecksByHostname)
6. **API Layer** (`internal/api/`)
   - Update `dto/target.go`
   - Update `handler/target.go`
   - Update `router.go`
7. **Infrastructure** (`internal/repository/`, `internal/api/`)
   - Update `manager.go`
   - Update `server.go`
8. **Tests** (`test/`)
   - Update `testutil/fixtures.go`
   - Update all affected unit tests
   - Update integration tests
9. **Build & Verify**
   - Run `go build`
   - Run `go test ./...`
   - Manual API testing

---

## Verification Checklist

After implementation:

- [ ] Build succeeds without errors
- [ ] All unit tests pass
- [ ] Migration runs successfully on clean database
- [ ] Target creation with no groups creates Default Own Group
- [ ] Target creation with groups does NOT create Default Own Group
- [ ] Cannot delete last group mapping from target
- [ ] `GetEffectiveChecksByHostname` returns checks from ALL groups
- [ ] API endpoints work correctly:
  - `POST /api/v1/targets` (with/without group_ids)
  - `GET /api/v1/targets/:id/groups`
  - `PUT /api/v1/targets/:id/groups`
  - `POST /api/v1/targets/groups/add`
  - `POST /api/v1/targets/groups/remove`
- [ ] Error handling is correct:
  - 400 when trying to remove last group
  - 404 when target/group not found
  - 409 when mapping already exists

---

## Breaking Changes

**API Changes**:
- `CreateTargetRequest.PrimaryGroupID` → `CreateTargetRequest.GroupIDs` (array, optional)
- `TargetResponse.PrimaryGroup` → `TargetResponse.Groups` (array)
- `TargetResponse.SecondaryGroups` removed

**Database Changes**:
- `targets.primary_group_id` column removed
- `target_secondary_groups` table removed
- `target_groups` table now handles all relationships

**Migration Required**:
- Existing data must be migrated from old structure to new structure
- See migration script for details

---

**End of Technical Specification**
