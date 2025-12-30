# Template-Instance Decoupling Refactoring Plan

## Overview

**목적**: Template과 Instance 간의 참조 관계를 제거하고, Instance 생성 시 Template 내용을 deep copy하는 방식으로 변경

**적용 대상**:
- CheckTemplate ↔ CheckInstance
- AlertTemplate ↔ AlertRule

**핵심 원칙**:
- Template은 "생성 시점의 블루프린트"일 뿐
- Instance/Rule은 생성 후 완전히 독립적인 엔티티
- Template 수정이 기존 Instance/Rule에 영향을 주지 않음
- Template 삭제가 가능 (Instance/Rule은 독립적으로 존재)

## Current Architecture Problems

### 문제점

1. **강한 결합**:
   ```go
   type CheckInstance struct {
       TemplateID string          // 외래키 참조
       Template   *CheckTemplate  // 런타임 조인
   }
   ```
   - Template 수정 시 모든 Instance에 영향
   - Template 삭제 시 Instance도 삭제되거나 orphan 상태

2. **의존성**:
   - Instance의 effective configuration을 계산하려면 Template을 조회해야 함
   - N+1 query 문제 발생 가능
   - Template이 삭제되면 Instance가 작동 불가

3. **버전 관리 어려움**:
   - Template 업데이트 시 어떤 Instance가 영향받는지 추적 어려움
   - Instance별로 다른 Template 버전을 사용할 수 없음

## Target Architecture

### 변경 후 구조

```go
// CheckTemplate은 그대로 유지 (블루프린트 역할)
type CheckTemplate struct {
    ID            string
    Name          string
    CheckType     string
    ScriptContent string
    Language      string
    DefaultConfig map[string]interface{}
    // ...
}

// CheckInstance는 독립적인 엔티티 (Template 내용을 복사)
type CheckInstance struct {
    ID        string
    // TemplateID 제거!

    // Template의 모든 필드를 직접 포함
    Name          string
    CheckType     string
    ScriptContent string
    Language      string
    DefaultConfig map[string]interface{}

    // Instance 고유 필드
    Scope         InstanceScope
    NamespaceID   *string
    GroupID       *string
    Config        map[string]interface{}  // Override config
    Priority      int
    IsActive      bool

    // 메타데이터 (생성 시점 정보 기록)
    CreatedFromTemplateID   *string    // Optional: 어떤 템플릿에서 생성되었는지 기록
    CreatedFromTemplateName *string    // Optional: 생성 시점 템플릿 이름
    TemplateVersion         *string    // Optional: 생성 시점 템플릿 버전

    DeletedAt  *time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

## Refactoring Steps

### Phase 1: Domain Model Updates

#### 1.1 CheckInstance Domain Model

**파일**: `internal/domain/check_instance.go`

변경사항:
```go
type CheckInstance struct {
    ID        string                 `json:"id"`

    // ❌ 제거: TemplateID, Template
    // TemplateID  string           `json:"template_id"`
    // Template    *CheckTemplate   `json:"template,omitempty"`

    // ✅ 추가: Template의 모든 필드
    Name          string                 `json:"name"`
    CheckType     string                 `json:"check_type"`
    ScriptContent string                 `json:"script_content"`
    Language      string                 `json:"language"`
    DefaultConfig map[string]interface{} `json:"default_config"`
    Description   string                 `json:"description"`
    Version       string                 `json:"version"`
    Hash          string                 `json:"hash"`

    // Instance 고유 필드 (기존 유지)
    Scope       InstanceScope          `json:"scope"`
    NamespaceID *string                `json:"namespace_id,omitempty"`
    Namespace   *Namespace             `json:"namespace,omitempty"`
    GroupID     *string                `json:"group_id,omitempty"`
    Group       *Group                 `json:"group,omitempty"`
    Config      map[string]interface{} `json:"config"`
    Priority    int                    `json:"priority"`
    IsActive    bool                   `json:"is_active"`

    // ✅ 추가: 생성 메타데이터 (optional, for tracking)
    CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
    CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`
    TemplateVersion         *string `json:"template_version,omitempty"`

    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

// ✅ 추가: Template에서 Instance 생성하는 생성자
func NewCheckInstanceFromTemplate(
    template *CheckTemplate,
    scope InstanceScope,
    namespaceID *string,
    groupID *string,
    overrideConfig map[string]interface{},
) *CheckInstance {
    // Deep copy template fields
    defaultConfig := make(map[string]interface{})
    for k, v := range template.DefaultConfig {
        defaultConfig[k] = deepCopyValue(v)
    }

    config := make(map[string]interface{})
    for k, v := range overrideConfig {
        config[k] = deepCopyValue(v)
    }

    return &CheckInstance{
        // Template fields (deep copied)
        Name:          template.Name,
        CheckType:     template.CheckType,
        ScriptContent: template.ScriptContent,
        Language:      template.Language,
        DefaultConfig: defaultConfig,
        Description:   template.Description,
        Version:       template.Version,
        Hash:          template.Hash,

        // Instance fields
        Scope:       scope,
        NamespaceID: namespaceID,
        GroupID:     groupID,
        Config:      config,
        Priority:    0,
        IsActive:    true,

        // Metadata
        CreatedFromTemplateID:   &template.ID,
        CreatedFromTemplateName: &template.Name,
        TemplateVersion:         &template.Version,

        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// ✅ 수정: Validate - TemplateID 검증 제거
func (ci *CheckInstance) Validate() error {
    // ❌ 제거
    // if ci.TemplateID == "" {
    //     return NewValidationError("template_id", "template_id is required")
    // }

    // ✅ 추가: 필수 필드 검증
    if ci.Name == "" {
        return NewValidationError("name", "name is required")
    }
    if ci.CheckType == "" {
        return NewValidationError("check_type", "check_type is required")
    }
    if ci.ScriptContent == "" {
        return NewValidationError("script_content", "script_content is required")
    }
    if ci.Language == "" {
        return NewValidationError("language", "language is required")
    }

    if ci.Scope == "" {
        return NewValidationError("scope", "scope is required")
    }

    // ... 나머지 검증 로직 유지

    return nil
}

// ✅ 수정: MergeConfig - template 파라미터 제거
func (ci *CheckInstance) MergeConfig() map[string]interface{} {
    merged := make(map[string]interface{})

    // Start with instance's default config (copied from template at creation)
    for k, v := range ci.DefaultConfig {
        merged[k] = v
    }

    // Override with instance config
    for k, v := range ci.Config {
        merged[k] = v
    }

    return merged
}

// ✅ 수정: GetEffectiveScriptHash - 자체 hash 사용
func (ci *CheckInstance) GetEffectiveScriptHash() string {
    return ci.Hash
}

// ✅ 추가: Helper function for deep copy
func deepCopyValue(v interface{}) interface{} {
    switch val := v.(type) {
    case map[string]interface{}:
        newMap := make(map[string]interface{})
        for k, v := range val {
            newMap[k] = deepCopyValue(v)
        }
        return newMap
    case []interface{}:
        newSlice := make([]interface{}, len(val))
        for i, item := range val {
            newSlice[i] = deepCopyValue(item)
        }
        return newSlice
    default:
        return v
    }
}
```

#### 1.2 AlertRule Domain Model

**파일**: `internal/domain/alert.go`

변경사항:
```go
type AlertRule struct {
    ID      string `json:"id"`
    GroupID string `json:"group_id"`
    Group   Group  `json:"group,omitempty"`

    // ❌ 제거: TemplateID, Template
    // TemplateID string        `json:"template_id"`
    // Template   AlertTemplate `json:"template,omitempty"`

    // ✅ 추가: Template의 모든 필드
    Name          string                 `json:"name"`
    Description   string                 `json:"description"`
    Severity      AlertSeverity          `json:"severity"`
    QueryTemplate string                 `json:"query_template"`
    DefaultConfig map[string]interface{} `json:"default_config"`

    // Rule 고유 필드 (기존 유지)
    Enabled       bool                   `json:"enabled"`
    Config        map[string]interface{} `json:"config"`
    MergeStrategy string                 `json:"merge_strategy"`
    Priority      int                    `json:"priority"`

    // ✅ 추가: 생성 메타데이터
    CreatedFromTemplateID   *string `json:"created_from_template_id,omitempty"`
    CreatedFromTemplateName *string `json:"created_from_template_name,omitempty"`

    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

// ✅ 추가: Template에서 Rule 생성하는 생성자
func NewAlertRuleFromTemplate(
    template *AlertTemplate,
    groupID string,
    overrideConfig map[string]interface{},
) *AlertRule {
    // Deep copy template fields
    defaultConfig := make(map[string]interface{})
    for k, v := range template.DefaultConfig {
        defaultConfig[k] = deepCopyValue(v)
    }

    config := make(map[string]interface{})
    for k, v := range overrideConfig {
        config[k] = deepCopyValue(v)
    }

    return &AlertRule{
        GroupID: groupID,

        // Template fields (deep copied)
        Name:          template.Name,
        Description:   template.Description,
        Severity:      template.Severity,
        QueryTemplate: template.QueryTemplate,
        DefaultConfig: defaultConfig,

        // Rule fields
        Enabled:       true,
        Config:        config,
        MergeStrategy: "merge",
        Priority:      0,

        // Metadata
        CreatedFromTemplateID:   &template.ID,
        CreatedFromTemplateName: &template.Name,

        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// ✅ 수정: RenderQuery - 자체 필드 사용
func (ar *AlertRule) RenderQuery() (string, error) {
    // Merge default config with rule config
    mergedConfig := make(map[string]interface{})
    for k, v := range ar.DefaultConfig {
        mergedConfig[k] = v
    }
    for k, v := range ar.Config {
        mergedConfig[k] = v
    }

    // Parse and execute template
    tmpl, err := template.New("query").Parse(ar.QueryTemplate)
    if err != nil {
        return "", fmt.Errorf("failed to parse query template: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, mergedConfig); err != nil {
        return "", fmt.Errorf("failed to render query template: %w", err)
    }

    return buf.String(), nil
}
```

### Phase 2: Database Schema Updates

#### 2.1 Migration: check_instances Table

**파일**: `migrations/XXX_decouple_check_template_instance.sql`

```sql
-- Add new columns to check_instances table
ALTER TABLE check_instances
ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN check_type VARCHAR(100) NOT NULL DEFAULT '',
ADD COLUMN script_content TEXT NOT NULL DEFAULT '',
ADD COLUMN language VARCHAR(50) NOT NULL DEFAULT 'bash',
ADD COLUMN default_config JSONB DEFAULT '{}',
ADD COLUMN description TEXT DEFAULT '',
ADD COLUMN version VARCHAR(50) DEFAULT '1.0.0',
ADD COLUMN hash VARCHAR(64) DEFAULT '',
ADD COLUMN created_from_template_id VARCHAR(255),
ADD COLUMN created_from_template_name VARCHAR(255),
ADD COLUMN template_version VARCHAR(50);

-- Migrate data: copy from templates to instances
UPDATE check_instances ci
SET
    name = ct.name,
    check_type = ct.check_type,
    script_content = ct.script_content,
    language = ct.language,
    default_config = ct.default_config,
    description = ct.description,
    version = ct.version,
    hash = ct.hash,
    created_from_template_id = ct.id,
    created_from_template_name = ct.name,
    template_version = ct.version
FROM check_templates ct
WHERE ci.template_id = ct.id;

-- Remove foreign key constraint
ALTER TABLE check_instances
DROP CONSTRAINT IF EXISTS fk_check_instances_template;

-- Drop template_id column
ALTER TABLE check_instances
DROP COLUMN template_id;

-- Remove default values (they were for migration only)
ALTER TABLE check_instances
ALTER COLUMN name DROP DEFAULT,
ALTER COLUMN check_type DROP DEFAULT,
ALTER COLUMN script_content DROP DEFAULT,
ALTER COLUMN language DROP DEFAULT;

-- Add indexes for performance
CREATE INDEX idx_check_instances_check_type ON check_instances(check_type);
CREATE INDEX idx_check_instances_hash ON check_instances(hash);
CREATE INDEX idx_check_instances_created_from_template_id ON check_instances(created_from_template_id);
```

#### 2.2 Migration: alert_rules Table

**파일**: `migrations/XXX_decouple_alert_template_rule.sql`

```sql
-- Add new columns to alert_rules table
ALTER TABLE alert_rules
ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN description TEXT DEFAULT '',
ADD COLUMN severity VARCHAR(50) NOT NULL DEFAULT 'warning',
ADD COLUMN query_template TEXT NOT NULL DEFAULT '',
ADD COLUMN default_config JSONB DEFAULT '{}',
ADD COLUMN created_from_template_id VARCHAR(255),
ADD COLUMN created_from_template_name VARCHAR(255);

-- Migrate data: copy from templates to rules
UPDATE alert_rules ar
SET
    name = at.name,
    description = at.description,
    severity = at.severity,
    query_template = at.query_template,
    default_config = at.default_config,
    created_from_template_id = at.id,
    created_from_template_name = at.name
FROM alert_templates at
WHERE ar.template_id = at.id;

-- Remove foreign key constraint
ALTER TABLE alert_rules
DROP CONSTRAINT IF EXISTS fk_alert_rules_template;

-- Drop template_id column
ALTER TABLE alert_rules
DROP COLUMN template_id;

-- Remove default values
ALTER TABLE alert_rules
ALTER COLUMN name DROP DEFAULT,
ALTER COLUMN severity DROP DEFAULT,
ALTER COLUMN query_template DROP DEFAULT;

-- Add indexes
CREATE INDEX idx_alert_rules_name ON alert_rules(name);
CREATE INDEX idx_alert_rules_severity ON alert_rules(severity);
CREATE INDEX idx_alert_rules_created_from_template_id ON alert_rules(created_from_template_id);
```

### Phase 3: Repository Layer Updates

#### 3.1 CheckInstance Repository

**파일**: `internal/repository/check_instance.go`

변경사항:
```go
// ✅ 수정: Create - Template 참조 제거
func (r *checkInstanceRepository) Create(ctx context.Context, instance *domain.CheckInstance) error {
    ormInstance := &models.CheckInstance{
        // ❌ 제거: TemplateID

        // ✅ 추가: Template fields
        Name:          instance.Name,
        CheckType:     instance.CheckType,
        ScriptContent: instance.ScriptContent,
        Language:      instance.Language,
        DefaultConfig: instance.DefaultConfig,
        Description:   instance.Description,
        Version:       instance.Version,
        Hash:          instance.Hash,

        // Instance fields
        Scope:       string(instance.Scope),
        NamespaceID: instance.NamespaceID,
        GroupID:     instance.GroupID,
        Config:      instance.Config,
        Priority:    instance.Priority,
        IsActive:    instance.IsActive,

        // Metadata
        CreatedFromTemplateID:   instance.CreatedFromTemplateID,
        CreatedFromTemplateName: instance.CreatedFromTemplateName,
        TemplateVersion:         instance.TemplateVersion,
    }

    result := r.db.WithContext(ctx).Create(ormInstance)
    if result.Error != nil {
        return result.Error
    }

    instance.ID = ormInstance.ID
    return nil
}

// ✅ 수정: FindByID - Template join 제거
func (r *checkInstanceRepository) FindByID(ctx context.Context, id string) (*domain.CheckInstance, error) {
    var ormInstance models.CheckInstance

    // ❌ 제거: .Preload("Template")
    result := r.db.WithContext(ctx).
        Preload("Namespace").
        Preload("Group").
        First(&ormInstance, "id = ? AND deleted_at IS NULL", id)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, result.Error
    }

    return r.toDomain(&ormInstance), nil
}

// ✅ 수정: toDomain - Template 변환 제거
func (r *checkInstanceRepository) toDomain(orm *models.CheckInstance) *domain.CheckInstance {
    instance := &domain.CheckInstance{
        ID: orm.ID,

        // ❌ 제거: TemplateID, Template

        // ✅ 추가: Template fields
        Name:          orm.Name,
        CheckType:     orm.CheckType,
        ScriptContent: orm.ScriptContent,
        Language:      orm.Language,
        DefaultConfig: orm.DefaultConfig,
        Description:   orm.Description,
        Version:       orm.Version,
        Hash:          orm.Hash,

        // Instance fields
        Scope:       domain.InstanceScope(orm.Scope),
        NamespaceID: orm.NamespaceID,
        GroupID:     orm.GroupID,
        Config:      orm.Config,
        Priority:    orm.Priority,
        IsActive:    orm.IsActive,

        // Metadata
        CreatedFromTemplateID:   orm.CreatedFromTemplateID,
        CreatedFromTemplateName: orm.CreatedFromTemplateName,
        TemplateVersion:         orm.TemplateVersion,

        DeletedAt: orm.DeletedAt,
        CreatedAt: orm.CreatedAt,
        UpdatedAt: orm.UpdatedAt,
    }

    if orm.Namespace != nil {
        instance.Namespace = &domain.Namespace{ /* ... */ }
    }

    if orm.Group != nil {
        instance.Group = &domain.Group{ /* ... */ }
    }

    return instance
}
```

#### 3.2 ORM Model Updates

**파일**: `internal/repository/models/check_instance.go`

```go
type CheckInstance struct {
    ID string `gorm:"primaryKey;type:varchar(255)"`

    // ❌ 제거
    // TemplateID string         `gorm:"column:template_id;not null"`
    // Template   *CheckTemplate `gorm:"foreignKey:TemplateID"`

    // ✅ 추가: Template fields
    Name          string                 `gorm:"column:name;not null"`
    CheckType     string                 `gorm:"column:check_type;not null"`
    ScriptContent string                 `gorm:"column:script_content;type:text;not null"`
    Language      string                 `gorm:"column:language;not null"`
    DefaultConfig map[string]interface{} `gorm:"column:default_config;type:jsonb"`
    Description   string                 `gorm:"column:description;type:text"`
    Version       string                 `gorm:"column:version"`
    Hash          string                 `gorm:"column:hash"`

    // Instance fields
    Scope       string                 `gorm:"column:scope;not null"`
    NamespaceID *string                `gorm:"column:namespace_id"`
    Namespace   *Namespace             `gorm:"foreignKey:NamespaceID"`
    GroupID     *string                `gorm:"column:group_id"`
    Group       *Group                 `gorm:"foreignKey:GroupID"`
    Config      map[string]interface{} `gorm:"column:config;type:jsonb"`
    Priority    int                    `gorm:"column:priority;default:0"`
    IsActive    bool                   `gorm:"column:is_active;default:true"`

    // ✅ 추가: Metadata
    CreatedFromTemplateID   *string `gorm:"column:created_from_template_id"`
    CreatedFromTemplateName *string `gorm:"column:created_from_template_name"`
    TemplateVersion         *string `gorm:"column:template_version"`

    DeletedAt *time.Time `gorm:"column:deleted_at;index"`
    CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}
```

### Phase 4: Service Layer Updates

#### 4.1 CheckInstance Service

**파일**: `internal/service/check_instance.go`

변경사항:
```go
// ✅ 추가: CreateFromTemplate - Template에서 Instance 생성
func (s *checkInstanceService) CreateFromTemplate(
    ctx context.Context,
    templateID string,
    scope domain.InstanceScope,
    namespaceID *string,
    groupID *string,
    overrideConfig map[string]interface{},
) (*domain.CheckInstance, error) {
    // 1. Get template
    template, err := s.templateRepo.FindByID(ctx, templateID)
    if err != nil {
        return nil, fmt.Errorf("template not found: %w", err)
    }

    // 2. Create instance from template (deep copy)
    instance := domain.NewCheckInstanceFromTemplate(
        template,
        scope,
        namespaceID,
        groupID,
        overrideConfig,
    )

    // 3. Validate
    if err := instance.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // 4. Save
    if err := s.repo.Create(ctx, instance); err != nil {
        return nil, fmt.Errorf("failed to create instance: %w", err)
    }

    return instance, nil
}

// ✅ 수정: GetEffectiveChecksForNode - Template join 불필요
func (s *checkInstanceService) GetEffectiveChecksForNode(
    ctx context.Context,
    hostname string,
) ([]*domain.EffectiveCheck, error) {
    // 1. Get target
    target, err := s.targetRepo.FindByHostname(ctx, hostname)
    if err != nil {
        return nil, fmt.Errorf("target not found: %w", err)
    }

    // 2. Get all applicable instances (scoped query)
    instances, err := s.repo.FindApplicableInstances(ctx, target)
    if err != nil {
        return nil, err
    }

    // 3. Resolve scope priorities and build effective checks
    effectiveChecks := make(map[string]*domain.EffectiveCheck)

    for _, instance := range instances {
        // Instance already has all needed fields (no template join needed!)
        check := &domain.EffectiveCheck{
            Name:          instance.Name,
            CheckType:     instance.CheckType,
            ScriptContent: instance.ScriptContent,
            Language:      instance.Language,
            Config:        instance.MergeConfig(), // Uses instance's own DefaultConfig and Config
            Version:       instance.Version,
            Hash:          instance.Hash,
            InstanceID:    instance.ID,
        }

        // Apply scope priority logic
        existing, exists := effectiveChecks[instance.CheckType]
        if !exists || instance.Priority < existing.Priority {
            effectiveChecks[instance.CheckType] = check
        }
    }

    result := make([]*domain.EffectiveCheck, 0, len(effectiveChecks))
    for _, check := range effectiveChecks {
        result = append(result, check)
    }

    return result, nil
}
```

#### 4.2 AlertRule Service

Similar changes for alert rule service:

```go
func (s *alertRuleService) CreateFromTemplate(
    ctx context.Context,
    templateID string,
    groupID string,
    overrideConfig map[string]interface{},
) (*domain.AlertRule, error) {
    template, err := s.templateRepo.FindByID(ctx, templateID)
    if err != nil {
        return nil, fmt.Errorf("template not found: %w", err)
    }

    rule := domain.NewAlertRuleFromTemplate(template, groupID, overrideConfig)

    if err := s.repo.Create(ctx, rule); err != nil {
        return nil, err
    }

    return rule, nil
}
```

### Phase 5: API Handler Updates

#### 5.1 CheckInstance Handler

**파일**: `internal/api/handler/check_instance.go`

변경사항:
```go
// ✅ 수정: CreateCheckInstance request DTO
type CreateCheckInstanceRequest struct {
    // Option 1: From template
    TemplateID *string `json:"template_id,omitempty"`

    // Option 2: Direct creation (all fields required)
    Name          *string                 `json:"name,omitempty"`
    CheckType     *string                 `json:"check_type,omitempty"`
    ScriptContent *string                 `json:"script_content,omitempty"`
    Language      *string                 `json:"language,omitempty"`
    DefaultConfig *map[string]interface{} `json:"default_config,omitempty"`
    Description   *string                 `json:"description,omitempty"`
    Version       *string                 `json:"version,omitempty"`

    // Common fields
    Scope       string                 `json:"scope" binding:"required"`
    NamespaceID *string                `json:"namespace_id,omitempty"`
    GroupID     *string                `json:"group_id,omitempty"`
    Config      map[string]interface{} `json:"config"`
}

// ✅ 수정: CreateCheckInstance handler
func (h *checkInstanceHandler) CreateCheckInstance(c *gin.Context) {
    var req CreateCheckInstanceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    var instance *domain.CheckInstance
    var err error

    if req.TemplateID != nil {
        // Create from template
        instance, err = h.service.CreateFromTemplate(
            c.Request.Context(),
            *req.TemplateID,
            domain.InstanceScope(req.Scope),
            req.NamespaceID,
            req.GroupID,
            req.Config,
        )
    } else {
        // Direct creation (all fields provided)
        if req.Name == nil || req.CheckType == nil || req.ScriptContent == nil {
            c.JSON(400, gin.H{"error": "name, check_type, and script_content are required"})
            return
        }

        instance = &domain.CheckInstance{
            Name:          *req.Name,
            CheckType:     *req.CheckType,
            ScriptContent: *req.ScriptContent,
            Language:      getOrDefault(req.Language, "bash"),
            DefaultConfig: getOrDefault(req.DefaultConfig, make(map[string]interface{})),
            Description:   getOrDefault(req.Description, ""),
            Version:       getOrDefault(req.Version, "1.0.0"),
            Scope:         domain.InstanceScope(req.Scope),
            NamespaceID:   req.NamespaceID,
            GroupID:       req.GroupID,
            Config:        req.Config,
            IsActive:      true,
        }

        // Compute hash
        instance.Hash = computeScriptHash(instance.ScriptContent)

        err = h.service.Create(c.Request.Context(), instance)
    }

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, instance)
}

func getOrDefault[T any](ptr *T, defaultVal T) T {
    if ptr != nil {
        return *ptr
    }
    return defaultVal
}
```

### Phase 6: API Documentation Updates

#### 6.1 Update API.md

**파일**: `docs/en/API.md`, `docs/kr/API.md`

변경사항:

```markdown
### Create Check Instance

**엔드포인트:** `POST /api/v1/check-instances`

**방법 1: Template에서 생성 (권장)**

```json
{
  "template_id": "template-uuid",
  "scope": "group",
  "namespace_id": "namespace-uuid",
  "group_id": "group-uuid",
  "config": {
    "mount_points": ["/mnt/data"]
  }
}
```

**방법 2: 직접 생성 (고급)**

```json
{
  "name": "custom-disk-check",
  "check_type": "disk",
  "script_content": "#!/bin/bash\necho 'custom check'",
  "language": "bash",
  "default_config": {},
  "description": "Custom disk check",
  "version": "1.0.0",
  "scope": "group",
  "group_id": "group-uuid",
  "config": {}
}
```

**주의사항**:
- Instance 생성 시 Template의 내용이 복사됩니다
- Template 수정은 기존 Instance에 영향을 주지 않습니다
- Instance는 독립적으로 수정/삭제 가능합니다
```

### Phase 7: Testing Updates

#### 7.1 Unit Tests

Update unit tests to reflect new structure:

```go
func TestCreateCheckInstanceFromTemplate(t *testing.T) {
    template := &domain.CheckTemplate{
        ID:            "template-1",
        Name:          "disk-check",
        CheckType:     "disk",
        ScriptContent: "#!/bin/bash\necho 'test'",
        Language:      "bash",
        DefaultConfig: map[string]interface{}{"threshold": 80},
        Version:       "1.0.0",
    }

    instance := domain.NewCheckInstanceFromTemplate(
        template,
        domain.ScopeGroup,
        stringPtr("ns-1"),
        stringPtr("group-1"),
        map[string]interface{}{"threshold": 90},
    )

    // Verify deep copy
    assert.Equal(t, template.Name, instance.Name)
    assert.Equal(t, template.CheckType, instance.CheckType)
    assert.Equal(t, template.ScriptContent, instance.ScriptContent)

    // Verify independence (modifying template doesn't affect instance)
    template.Name = "modified-name"
    assert.NotEqual(t, template.Name, instance.Name)

    // Verify config merge
    merged := instance.MergeConfig()
    assert.Equal(t, 90, merged["threshold"]) // Override takes precedence
}
```

## Migration Guide

### For Existing Deployments

1. **Backup Database**:
   ```bash
   pg_dump -U postgres config_server > backup_before_migration.sql
   ```

2. **Run Migration**:
   ```bash
   psql -U postgres -d config_server -f migrations/XXX_decouple_check_template_instance.sql
   psql -U postgres -d config_server -f migrations/XXX_decouple_alert_template_rule.sql
   ```

3. **Verify Data Migration**:
   ```sql
   -- Check that all instances have copied template data
   SELECT COUNT(*) FROM check_instances WHERE script_content = '';
   -- Should return 0

   SELECT COUNT(*) FROM alert_rules WHERE query_template = '';
   -- Should return 0
   ```

4. **Deploy New Code**:
   ```bash
   # Build and deploy
   make build
   ./config-server
   ```

5. **Verify API**:
   ```bash
   # Test creating instance from template
   curl -X POST http://localhost:8080/api/v1/check-instances \
     -H "Content-Type: application/json" \
     -d '{
       "template_id": "existing-template-id",
       "scope": "global",
       "config": {}
     }'
   ```

## Benefits After Refactoring

### 1. Independence
- Template 수정이 기존 Instance에 영향 없음
- Template 삭제 가능 (Instance는 계속 작동)
- Instance별로 완전히 다른 스크립트 사용 가능

### 2. Performance
- N+1 query 문제 해결
- Template join 불필요
- 단일 테이블 쿼리로 모든 정보 획득

### 3. Flexibility
- Instance를 독립적으로 수정 가능
- Template 없이도 Instance 직접 생성 가능
- 버전 관리가 명확 (생성 시점 스냅샷)

### 4. Traceability
- `created_from_template_id`로 출처 추적 가능
- Template 삭제 후에도 이력 확인 가능
- 감사(audit) 목적으로 유용

## Rollback Plan

If issues occur:

1. **Rollback Database**:
   ```bash
   psql -U postgres config_server < backup_before_migration.sql
   ```

2. **Deploy Old Code**:
   ```bash
   git checkout <previous-commit>
   make build
   ./config-server
   ```

## Timeline

- **Phase 1-2**: Domain & Schema (1 day)
- **Phase 3**: Repository Layer (1 day)
- **Phase 4**: Service Layer (1 day)
- **Phase 5-6**: API & Documentation (0.5 day)
- **Phase 7**: Testing (1 day)
- **Total**: ~4.5 days

## Approval Checklist

- [ ] Architecture review approved
- [ ] Database migration tested on staging
- [ ] Rollback plan verified
- [ ] API documentation updated
- [ ] Tests updated and passing
- [ ] Ready for production deployment
