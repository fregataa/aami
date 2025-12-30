# AGENT.md - Config Server Architecture Guide for Coding Agents

**Last Updated**: 2025-12-28
**Purpose**: Technical reference for coding agents working on the Config Server codebase

## 🎯 Project Overview

The Config Server is a REST API service that manages monitoring infrastructure configuration. It implements a hierarchical group-based policy inheritance system for targets, alerts, and dynamic checks (CheckTemplate/CheckInstance pattern).

**Tech Stack**:
- Go 1.25.5
- PostgreSQL 15+ (primary database)
- GORM (ORM)
- Gin (HTTP framework)

---

## 🏗️ Architecture Principles

### 1. Clean Architecture with Layered Design

```
┌─────────────────────────────────────┐
│  API Layer (HTTP)                   │  ← Handlers, Middlewares, DTOs
├─────────────────────────────────────┤
│  Service Layer                      │  ← Business Logic
├─────────────────────────────────────┤
│  Repository Layer                   │  ← Data Access (ORM Models)
├─────────────────────────────────────┤
│  Domain Layer                       │  ← Pure Business Entities
└─────────────────────────────────────┘
         ↓
┌─────────────────────────────────────┐
│  Database (PostgreSQL)              │
└─────────────────────────────────────┘
```

### 2. Dependency Rule

**CRITICAL**: Dependencies must flow downward only:
- API → Service → Repository → Domain
- Domain has NO dependencies on any other layer
- Repository knows about Domain but NOT Service
- Service knows about Repository and Domain but NOT API
- API knows about all layers

### 3. Domain-Driven Design

- **Domain models** are pure business entities with NO framework dependencies
- **NO GORM tags** in domain models
- **NO database types** (e.g., `gorm.DeletedAt`) in domain models
- Domain models contain only business logic and validation

---

## 📁 Layer Responsibilities

### Layer 1: Domain (`internal/domain/`)

**Purpose**: Pure business entities and logic

**Rules**:
- ✅ Business logic methods
- ✅ Validation functions
- ✅ Pure Go types only (`*time.Time`, not `gorm.DeletedAt`)
- ✅ JSON tags for serialization
- ❌ NO GORM tags
- ❌ NO database imports
- ❌ NO ORM types

**Example**:
```go
// domain/target.go
type Target struct {
    ID              string                 `json:"id"`
    Hostname        string                 `json:"hostname"`
    Status          TargetStatus           `json:"status"`
    DeletedAt       *time.Time             `json:"deleted_at,omitempty"` // NOT gorm.DeletedAt
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
}

// Business logic methods
func (t *Target) IsHealthy() bool {
    return t.Status == TargetStatusActive && t.LastSeen != nil
}
```

### Layer 2: Repository (`internal/repository/`)

**Purpose**: Data access and persistence

**Components**:
1. **Interface**: Defines data operations (in same file)
2. **ORM Model**: Database-specific struct with GORM tags
3. **Implementation**: Uses ORM model for DB operations
4. **Converters**: Transform between Domain ↔ ORM Model

**Rules**:
- ✅ GORM operations use ORM models, NOT domain models
- ✅ Convert domain → ORM before DB write
- ✅ Convert ORM → domain after DB read
- ✅ Handle soft deletes using `gorm.DeletedAt` in ORM models
- ❌ NEVER pass domain models directly to GORM

**Pattern**:
```go
// repository/target.go

// 1. ORM Model (DB layer)
type TargetModel struct {
    ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Hostname  string         `gorm:"not null;unique;index"`
    Status    string         `gorm:"type:varchar(20);not null;default:'active'"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (TargetModel) TableName() string {
    return "targets"
}

// 2. Converters
func ToTargetModel(d *domain.Target) *TargetModel {
    m := &TargetModel{
        ID:        d.ID,
        Hostname:  d.Hostname,
        Status:    string(d.Status),
        CreatedAt: d.CreatedAt,
        UpdatedAt: d.UpdatedAt,
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
        Status:    domain.TargetStatus(m.Status),
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
    if m.DeletedAt.Valid {
        deletedAt := m.DeletedAt.Time
        d.DeletedAt = &deletedAt
    }
    return d
}

// 3. Interface
type TargetRepository interface {
    Create(ctx context.Context, target *domain.Target) error
    GetByID(ctx context.Context, id string) (*domain.Target, error)
    Delete(ctx context.Context, id string) error   // Soft delete
    Purge(ctx context.Context, id string) error    // Hard delete (Unscoped)
    Restore(ctx context.Context, id string) error  // Restore soft-deleted
}

// 4. Implementation
type targetRepository struct {
    db *gorm.DB
}

func (r *targetRepository) Create(ctx context.Context, target *domain.Target) error {
    model := ToTargetModel(target)
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        return err
    }
    *target = *model.ToDomain() // Update with DB-generated fields
    return nil
}

func (r *targetRepository) GetByID(ctx context.Context, id string) (*domain.Target, error) {
    var model TargetModel
    err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return model.ToDomain(), nil
}

func (r *targetRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&TargetModel{}, "id = ?", id).Error
}

func (r *targetRepository) Purge(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Unscoped().Delete(&TargetModel{}, "id = ?", id).Error
}

func (r *targetRepository) Restore(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Unscoped().
        Model(&TargetModel{}).
        Where("id = ?", id).
        Update("deleted_at", nil).Error
}
```

### Layer 3: Service (`internal/service/`)

**Purpose**: Business logic orchestration

**Rules**:
- ✅ Coordinate multiple repositories
- ✅ Implement business workflows
- ✅ Handle transactions
- ✅ Work with domain models (NOT DTOs, NOT ORM models)
- ✅ Return domain models or errors
- ❌ NO HTTP concerns (status codes, headers)
- ❌ NO direct database access (use repositories)

**Pattern**:
```go
// service/target.go
type TargetService struct {
    targetRepo repository.TargetRepository
    groupRepo  repository.GroupRepository
}

func (s *TargetService) Create(ctx context.Context, req dto.CreateTargetRequest) (*domain.Target, error) {
    // 1. Validation
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 2. Business logic
    group, err := s.groupRepo.GetByID(ctx, req.PrimaryGroupID)
    if err != nil {
        return nil, ErrGroupNotFound
    }

    // 3. Create domain model
    target := &domain.Target{
        Hostname:       req.Hostname,
        IPAddress:      req.IPAddress,
        PrimaryGroupID: req.PrimaryGroupID,
        Status:         domain.TargetStatusActive,
    }

    // 4. Persist
    if err := s.targetRepo.Create(ctx, target); err != nil {
        return nil, err
    }

    return target, nil
}
```

### Layer 4: API (`internal/api/`)

**Purpose**: HTTP interface

**Components**:
- **Handlers** (`api/handler/`): HTTP request/response handling
- **DTOs** (`api/dto/`): API data structures with validation
- **Middleware** (`api/middleware/`): Cross-cutting concerns
- **Router** (`api/router.go`): Route definitions

**Rules**:
- ✅ Handle HTTP-specific concerns (status codes, headers)
- ✅ Validate input using DTO validation tags
- ✅ Convert DTOs ↔ Domain models
- ✅ Return appropriate HTTP status codes
- ❌ NO business logic (delegate to services)
- ❌ NO direct repository access

**Pattern**:
```go
// api/handler/target.go
type TargetHandler struct {
    targetService *service.TargetService
}

func (h *TargetHandler) Create(c *gin.Context) {
    // 1. Parse and validate DTO
    var req dto.CreateTargetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: err.Error(),
            Code:  "INVALID_REQUEST",
        })
        return
    }

    // 2. Call service
    target, err := h.targetService.Create(c.Request.Context(), req)
    if err != nil {
        respondError(c, err) // Maps service errors to HTTP status codes
        return
    }

    // 3. Convert to response DTO
    c.JSON(http.StatusCreated, dto.ToTargetResponse(target))
}
```

---

## 🔑 Key Patterns

### 1. Soft Delete Implementation

**Database**: Use `deleted_at` column with partial indexes
```sql
ALTER TABLE targets ADD COLUMN deleted_at TIMESTAMP;
CREATE INDEX idx_targets_deleted_at ON targets(deleted_at);
CREATE UNIQUE INDEX targets_hostname_deleted_at_key ON targets(hostname) WHERE deleted_at IS NULL;
```

**Domain**: Use `*time.Time`
```go
type Target struct {
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

**Repository ORM Model**: Use `gorm.DeletedAt`
```go
type TargetModel struct {
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

**Operations**:
- `Delete()`: Soft delete (sets `deleted_at`)
- `Purge()`: Hard delete using `Unscoped()`
- `Restore()`: Clears `deleted_at` using `Unscoped()`

### 2. REST API Design

**Endpoint Pattern**:
```
POST   /api/v1/targets          # Create
GET    /api/v1/targets          # List (paginated)
GET    /api/v1/targets/:id      # Get by ID
PUT    /api/v1/targets/:id      # Update
POST   /api/v1/targets/delete   # Soft delete (body: {"id": "uuid"})
POST   /api/v1/targets/purge    # Hard delete (body: {"id": "uuid"})
POST   /api/v1/targets/restore  # Restore (body: {"id": "uuid"})
```

**Why POST for delete/purge/restore?**
- Allows ID in request body (not URL)
- More secure for sensitive operations
- Supports additional parameters if needed

### 3. Error Handling

**Service Layer** - Use domain errors:
```go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrInvalidInput  = errors.New("invalid input")
)
```

**API Layer** - Map to HTTP status codes:
```go
func respondError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, service.ErrNotFound):
        c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error(), Code: "NOT_FOUND"})
    case errors.Is(err, service.ErrAlreadyExists):
        c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error(), Code: "CONFLICT"})
    case errors.Is(err, gorm.ErrRecordNotFound):
        c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Not found", Code: "NOT_FOUND"})
    default:
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error(), Code: "INTERNAL_ERROR"})
    }
}
```

### 4. DTO Pattern

**Request DTOs** - Validation tags:
```go
type CreateTargetRequest struct {
    Hostname      string   `json:"hostname" binding:"required,min=1,max=255"`
    IPAddress     string   `json:"ip_address" binding:"required,ip"`
    PrimaryGroupID string  `json:"primary_group_id" binding:"required,uuid"`
}
```

**Response DTOs** - Include related data:
```go
type TargetResponse struct {
    ID           string          `json:"id"`
    Hostname     string          `json:"hostname"`
    PrimaryGroup *GroupResponse  `json:"primary_group,omitempty"`
    CreatedAt    time.Time       `json:"created_at"`
}

func ToTargetResponse(t *domain.Target) TargetResponse {
    resp := TargetResponse{
        ID:        t.ID,
        Hostname:  t.Hostname,
        CreatedAt: t.CreatedAt,
    }
    if t.PrimaryGroup.ID != "" {
        group := ToGroupResponse(&t.PrimaryGroup)
        resp.PrimaryGroup = &group
    }
    return resp
}
```

---

## 📋 Implementation Checklist

When adding a new entity:

### 1. Domain Model (`internal/domain/`)
- [ ] Create domain struct with pure Go types
- [ ] Add JSON tags only (NO GORM tags)
- [ ] Use `*time.Time` for `DeletedAt`
- [ ] Add business logic methods
- [ ] Add validation functions

### 2. Repository (`internal/repository/`)
- [ ] Create ORM model with GORM tags
- [ ] Add `TableName()` method
- [ ] Create `ToModel()` converter (domain → ORM)
- [ ] Create `ToDomain()` method (ORM → domain)
- [ ] Define repository interface
- [ ] Implement CRUD operations
- [ ] Implement soft delete operations (Delete, Purge, Restore)
- [ ] Use ORM model in all GORM operations

### 3. Service (`internal/service/`)
- [ ] Create service struct with repository dependencies
- [ ] Implement business logic methods
- [ ] Handle validation and error cases
- [ ] Work with domain models
- [ ] Add transaction handling if needed

### 4. DTOs (`internal/api/dto/`)
- [ ] Create request DTOs with validation tags
- [ ] Create response DTOs
- [ ] Add converter functions (domain ↔ DTO)

### 5. Handler (`internal/api/handler/`)
- [ ] Create handler struct with service dependency
- [ ] Implement HTTP handlers
- [ ] Parse and validate DTOs
- [ ] Call service methods
- [ ] Return appropriate HTTP responses

### 6. Router (`internal/api/router.go`)
- [ ] Register routes
- [ ] Use POST for delete/purge/restore

### 7. Migration (`migrations/`)
- [ ] Create migration SQL file
- [ ] Add table with `deleted_at` column and index
- [ ] Add partial unique indexes (`WHERE deleted_at IS NULL`)

---

## 🚫 Common Mistakes to Avoid

### ❌ DON'T: Use GORM in domain models
```go
// domain/target.go - WRONG
type Target struct {
    ID string `gorm:"primaryKey" json:"id"` // ❌ NO GORM tags
}
```

### ✅ DO: Keep domain pure
```go
// domain/target.go - CORRECT
type Target struct {
    ID string `json:"id"` // ✅ Only JSON tags
}
```

### ❌ DON'T: Pass domain models to GORM
```go
// repository/target.go - WRONG
func (r *targetRepository) Create(ctx context.Context, target *domain.Target) error {
    return r.db.Create(target).Error // ❌ Domain model to GORM
}
```

### ✅ DO: Use ORM models with converters
```go
// repository/target.go - CORRECT
func (r *targetRepository) Create(ctx context.Context, target *domain.Target) error {
    model := ToTargetModel(target) // ✅ Convert first
    if err := r.db.Create(model).Error; err != nil {
        return err
    }
    *target = *model.ToDomain() // ✅ Convert back
    return nil
}
```

### ❌ DON'T: Put business logic in handlers
```go
// handler/target.go - WRONG
func (h *TargetHandler) Create(c *gin.Context) {
    // ❌ Business logic in handler
    if target.Status == "inactive" && target.Priority > 100 {
        // complex business rules...
    }
}
```

### ✅ DO: Delegate to service layer
```go
// handler/target.go - CORRECT
func (h *TargetHandler) Create(c *gin.Context) {
    target, err := h.targetService.Create(c.Request.Context(), req) // ✅ Service handles logic
    if err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, dto.ToTargetResponse(target))
}
```

---

## 🧪 Testing Strategy

### Unit Tests
- Test domain business logic
- Mock repositories in service tests
- Mock services in handler tests

### Integration Tests
- Test repository with real database (testcontainers)
- Verify ORM model ↔ domain conversions
- Test soft delete behavior

### API Tests
- Test full HTTP request/response cycle
- Verify status codes and error responses
- Test authentication and authorization

---

## 🔧 Current Implementation Status

**Entities**:
- [x] Namespace - Domain, Repository (ORM), Service, Handler, Router
- [ ] Group - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅
- [ ] Target - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅
- [ ] Exporter - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅
- [ ] AlertTemplate - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅
- [ ] AlertRule - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅
- [x] CheckTemplate - Domain ✅, Repository ✅, Service ✅, Handler ✅, Router ✅
- [x] CheckInstance - Domain ✅, Repository ✅, Service ✅, Handler ✅, Router ✅
- [ ] BootstrapToken - Domain ✅, Repository ⚠️ (needs ORM model), Service ✅, Handler ✅, Router ✅

**Next Steps**:
1. Complete ORM models for remaining 7 repositories
2. Add converter functions (ToModel, ToDomain)
3. Update repository implementations to use ORM models
4. Verify build and tests

---

## 📝 Documentation Standards

### Bilingual Documentation Requirement

**CRITICAL**: When adding documentation to the `docs/` directory, you MUST provide BOTH English and Korean versions.

**Directory Structure**:
```
docs/
├── en/           # English documentation
│   ├── API.md
│   ├── QUICKSTART.md
│   └── CHECK-SCRIPT-MANAGEMENT.md
└── kr/           # Korean documentation (한글 문서)
    ├── API.md
    ├── QUICKSTART.md
    └── CHECK-SCRIPT-MANAGEMENT.md
```

**Rules**:
- ✅ Create both `docs/en/FILENAME.md` and `docs/kr/FILENAME.md`
- ✅ Keep content synchronized between versions
- ✅ Translate all sections, including examples and code comments
- ❌ NEVER create only one language version

**When Creating New Documentation**:
1. Write the document in one language first (typically English)
2. Immediately create the translated version
3. Verify both files are complete before considering the task done

**Example Commit**:
```bash
git add docs/en/NEW-FEATURE.md docs/kr/NEW-FEATURE.md
git commit -m "docs: add NEW-FEATURE documentation (en + kr)"
```

---

## 📚 References

- **Clean Architecture**: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- **Domain-Driven Design**: Separate business logic from infrastructure
- **Repository Pattern**: Isolate data access logic
- **GORM Documentation**: https://gorm.io/docs/

---

## 🎨 Planning Process Guidelines

### High-Level Planning First, Technical Details Later

**CRITICAL**: When planning new features or architectural changes, ALWAYS follow this two-phase approach:

#### Phase 1: High-Level Overview
Present the overall architectural direction WITHOUT technical implementation details:

**Include**:
- ✅ What will change at a conceptual level
- ✅ How data flows will be modified
- ✅ Major design decisions and their rationale
- ✅ User-facing behavior changes
- ✅ Migration path overview
- ✅ Expected benefits and trade-offs

**Exclude**:
- ❌ Specific code structures or types
- ❌ Database column types and constraints
- ❌ Implementation patterns (e.g., converter functions, constructors)
- ❌ Detailed step-by-step code changes
- ❌ Line-by-line migration scripts

**Purpose**: Allow stakeholders to understand and approve the direction before investing time in technical details.

#### Phase 2: Technical Implementation (Only After Approval)
After the high-level plan is approved, document:
- Domain model changes
- Repository ORM model updates
- Service layer modifications
- API endpoint changes
- Detailed migration scripts
- Test cases and validation

**Example Flow**:
```
User Request → High-Level Plan → User Approval → Technical Details → Implementation
```

### Orphan Prevention Pattern

When implementing relationships that prevent orphan records:

**Pattern**: Junction Table with Minimum Constraint

**Rules**:
- ✅ Use junction table to manage all relationships (e.g., `target_groups`)
- ✅ Enforce "minimum 1 relationship" constraint at application level
- ✅ Check junction table record count before allowing deletion
- ❌ NO warning dialogs for default/special relationships
- ❌ NO soft validation (must enforce at database transaction level)

**Implementation**:
```go
// Service layer - Delete group mapping
func (s *TargetService) RemoveGroupMapping(ctx context.Context, targetID, groupID string) error {
    // Check remaining relationships
    count, err := s.targetGroupRepo.CountByTarget(ctx, targetID)
    if err != nil {
        return err
    }

    // Prevent deletion of last relationship
    if count <= 1 {
        return ErrCannotRemoveLastGroup // Clear error, no warning dialog
    }

    return s.targetGroupRepo.Delete(ctx, targetID, groupID)
}
```

**Database Validation**:
```sql
-- Trigger to prevent orphan targets (PostgreSQL)
CREATE OR REPLACE FUNCTION prevent_target_orphan()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM target_groups WHERE target_id = OLD.target_id) <= 1 THEN
        RAISE EXCEPTION 'Cannot remove last group mapping for target';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_prevent_target_orphan
    BEFORE DELETE ON target_groups
    FOR EACH ROW
    EXECUTE FUNCTION prevent_target_orphan();
```

**Why This Pattern**:
- Simple and predictable behavior
- Clear error messages instead of modal dialogs
- Enforced at both application and database levels
- Works consistently across all clients (API, CLI, UI)

---

**For Questions**: Refer to this document when uncertain about architecture decisions.
