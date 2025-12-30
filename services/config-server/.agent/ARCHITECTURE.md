# Config Server Architecture Guide

This document describes the architecture, coding conventions, and development guidelines for the Config Server.

## Architecture Overview

### Clean Architecture Layers

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

### Key Principles

1. **Domain Independence**: Domain models contain no framework dependencies (no GORM tags)
2. **Dependency Rule**: Dependencies flow downward (API → Service → Repository → Domain)
3. **ORM Separation**: Repository layer uses ORM models that convert to/from domain models

## Directory Structure

```
services/config-server/
├── cmd/
│   ├── config-server/       # Server entry point
│   ├── migrate/             # Database migration tool
│   └── server/              # Alternative server entry
│
├── internal/
│   ├── api/
│   │   ├── dto/            # Data Transfer Objects
│   │   ├── handler/        # HTTP handlers
│   │   └── middleware/     # HTTP middlewares
│   │
│   ├── domain/             # Pure business entities
│   │   ├── namespace.go
│   │   ├── group.go
│   │   ├── target.go
│   │   └── ...
│   │
│   ├── repository/         # Data access layer
│   │   ├── models/         # GORM models
│   │   ├── namespace.go
│   │   └── ...
│   │
│   ├── service/            # Business logic
│   │   ├── namespace.go
│   │   └── ...
│   │
│   ├── config/             # Configuration
│   └── util/               # Utilities
│
├── migrations/             # Database migrations
├── test/                   # Tests
└── docs/                   # Documentation
```

## Coding Conventions

### Domain Layer

**Rules**:
- No framework dependencies
- Pure Go structs
- Business logic methods
- No database tags

**Example**:
```go
// domain/namespace.go
package domain

type Namespace struct {
    ID             string
    Name           string
    Description    string
    PolicyPriority int
    MergeStrategy  string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// Business logic
func (n *Namespace) Validate() error {
    if n.Name == "" {
        return errors.New("name is required")
    }
    return nil
}
```

### Repository Layer

**Rules**:
- GORM models in `models/` subdirectory
- Convert between domain and ORM models
- Return domain models to service layer
- Handle database errors

**Example**:
```go
// repository/models/namespace.go
package models

type Namespace struct {
    ID             string `gorm:"type:uuid;primaryKey"`
    Name           string `gorm:"type:varchar(100);uniqueIndex"`
    // ... GORM tags
}

func (m *Namespace) ToDomain() *domain.Namespace {
    return &domain.Namespace{
        ID:   m.ID,
        Name: m.Name,
        // ...
    }
}

// repository/namespace.go
func (r *namespaceRepo) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
    var model models.Namespace
    if err := r.db.First(&model, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return model.ToDomain(), nil
}
```

### Service Layer

**Rules**:
- Business logic orchestration
- Transaction management
- Input validation
- Error handling

**Example**:
```go
// service/namespace.go
func (s *namespaceService) Create(ctx context.Context, ns *domain.Namespace) error {
    // Validate
    if err := ns.Validate(); err != nil {
        return err
    }

    // Business logic
    // ...

    // Persist
    return s.repo.Create(ctx, ns)
}
```

### API Layer

**Rules**:
- DTOs for request/response
- Handler functions thin
- Delegate to service layer
- Consistent error responses

**Example**:
```go
// api/dto/namespace.go
type CreateNamespaceRequest struct {
    Name           string `json:"name" binding:"required"`
    Description    string `json:"description"`
    PolicyPriority int    `json:"policy_priority"`
}

// api/handler/namespace.go
func (h *NamespaceHandler) Create(c *gin.Context) {
    var req dto.CreateNamespaceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, err)
        return
    }

    ns := &domain.Namespace{
        Name:           req.Name,
        Description:    req.Description,
        PolicyPriority: req.PolicyPriority,
    }

    if err := h.service.Create(c.Request.Context(), ns); err != nil {
        respondError(c, err)
        return
    }

    c.JSON(http.StatusCreated, dto.FromDomain(ns))
}
```

## Error Handling

### Current Pattern (Sprint 1-5)
- Service layer defines errors
- Handlers map errors to HTTP status

### Planned Pattern (Sprint 6+)
- Unified `internal/errors` package
- Domain errors (ErrNotFound, ErrDuplicateKey)
- Structured errors (ValidationError, BindingError)

## Database Conventions

### Migrations
- Never modify existing migrations
- Create new migration for changes
- Use descriptive names: `YYYYMMDDHHMMSS_description.sql`

### Schema Design
- UUIDs for primary keys
- Soft delete with `deleted_at`
- Timestamps: `created_at`, `updated_at`
- Foreign keys with ON DELETE CASCADE/RESTRICT

### Naming
- Table names: plural, snake_case (`namespaces`, `target_groups`)
- Column names: snake_case
- Index names: `idx_table_column`

## Testing Conventions

### Unit Tests
- Test file: `*_test.go` in same package
- Test function: `TestFunctionName`
- Use table-driven tests

### Integration Tests
- Location: `test/integration/`
- Use testcontainers for database
- Clean up after tests

### Test Coverage
- Target: >70%
- Focus on service and domain layers

## Git Workflow

### Branch Naming
- Feature: `feature/description`
- Bugfix: `bugfix/description`
- Refactor: `refactor/description`

### Commit Messages
- Format: `<type>: <description>`
- Types: feat, fix, refactor, docs, test, chore
- Keep concise

## Performance Guidelines

### Database
- Use indexes for frequently queried columns
- Avoid N+1 queries (use preload/joins)
- Use connection pooling

### API
- Paginate list endpoints
- Use caching for frequently accessed data
- Implement rate limiting

## Security Guidelines

### Input Validation
- Validate all user input
- Use binding tags in DTOs
- Sanitize SQL inputs (use parameterized queries)

### Authentication (Sprint 9+)
- JWT tokens
- Password hashing with bcrypt
- API keys for CLI

## Development Workflow

1. Read this architecture guide
2. Understand the layer responsibilities
3. Follow coding conventions
4. Write tests
5. Update documentation
6. Submit PR

## References

- Clean Architecture: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- Go Project Layout: https://github.com/golang-standards/project-layout
- GORM Documentation: https://gorm.io/docs/
