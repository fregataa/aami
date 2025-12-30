# AAMI Project Guide for Coding Agents

This document provides essential context and guidelines for AI coding agents working on the AAMI project.

## Project Overview

**AAMI** (AI Accelerator Monitoring Infrastructure) is a hybrid infrastructure monitoring system designed for large-scale AI accelerator clusters (GPUs, NPUs, etc.), built on the Prometheus ecosystem.

### Key Components
- **Config Server**: REST API for configuration management (Go)
- **CLI**: Command-line management tool (Go, independent module)
- **Exporters**: Custom metric exporters for AI infrastructure (Planned)
- **Web UI**: Web-based management interface (Planned)

## Project Structure

```
aami/
├── .agent/                    # Project-wide agent context
│   ├── planning/              # Sprint planning and tracking
│   │   ├── sprints/
│   │   │   ├── completed/    # Completed sprints
│   │   │   ├── current/      # Current sprint
│   │   │   ├── planned/      # Planned sprints
│   │   │   └── archived/     # Old sprint files
│   │   ├── TRACKER.md        # Sprint progress tracker
│   │   └── PLAN.md           # Overall project plan
│   └── AGENT.md              # This file
│
├── cli/                       # CLI tool (independent module)
│   ├── cmd/aami/             # CLI entry point
│   ├── internal/             # CLI implementation
│   ├── go.mod                # Independent Go module
│   └── README.md
│
├── services/
│   └── config-server/        # Config management server
│       ├── .agent/           # Config Server specific context
│       │   ├── AGENT.md      # Agent guide for this service
│       │   └── ARCHITECTURE.md  # Architecture documentation
│       ├── cmd/              # Server entry points
│       ├── internal/         # Service implementation
│       └── README.md
│
├── config/                    # Prometheus, Grafana configs
├── deploy/                    # Deployment configurations
└── docs/                      # Project documentation
```

## Critical .agent Directory Rules

### Principle: Clear Separation of Concerns

1. **Root `.agent/`** (Project-wide)
   - Purpose: Project-level planning, sprint management
   - Content: Sprint plans, overall roadmap, project decisions
   - Audience: Project managers, all developers

2. **Service `.agent/`** (e.g., `config-server/.agent/`)
   - Purpose: Service-specific architecture and conventions
   - Content: Architecture patterns, coding standards, development guides
   - Audience: Developers working on that specific service

### What Goes Where

| Content Type | Root .agent/ | Service .agent/ |
|--------------|--------------|-----------------|
| Sprint plans | ✅ Yes | ❌ No |
| Project roadmap | ✅ Yes | ❌ No |
| Architecture patterns | ❌ No | ✅ Yes |
| Coding conventions | ❌ No | ✅ Yes |
| Development setup | ❌ No | ✅ Yes |

### Example Violations to Avoid

❌ **Wrong**: Sprint plan in `config-server/.agent/docs/SPRINT_PLAN.md`
✅ **Right**: Sprint plan in `.agent/planning/sprints/`

❌ **Wrong**: Config Server architecture in `.agent/architecture.md`
✅ **Right**: Config Server architecture in `config-server/.agent/ARCHITECTURE.md`

## Sprint Management

### Current Sprint System

Sprints are tracked in `.agent/planning/` with clear state separation:

```
.agent/planning/sprints/
├── completed/        # Completed work (reference)
├── current/          # Active sprint (1 file)
├── planned/          # Upcoming sprints (N files)
└── archived/         # Old/outdated plans
```

### Sprint File Format

Each sprint has a dedicated markdown file:

```markdown
# Sprint N: [Title]

**Status**: 📋 Planned | 🚧 In Progress | ✅ Completed
**Duration**: [Estimated] / [Actual]
**Started**: YYYY-MM-DD
**Completed**: YYYY-MM-DD

## Goals
...

## Tasks
- [ ] Task 1
- [ ] Task 2

## Deliverables
...

## Notes
...
```

### Progress Tracking

Use `TRACKER.md` for quick overview of all sprints:
- Current sprint status
- Progress statistics
- Timeline and milestones
- Blockers and risks

## Current Development Phase

**Phase**: Config Server Stabilization (Sprint 6-10)

**Current Sprint**: Sprint 6 - Unified Error Handling
- Location: `.agent/planning/sprints/current/sprint-06-error-handling.md`
- Status: Planned
- Duration: 8-12 days

**Upcoming**:
- Sprint 7: Testing (2-3 weeks)
- Sprint 8: API Documentation (1-2 weeks)
- Sprint 9: Authentication (2-3 weeks)
- Sprint 10: Performance Optimization (2-3 weeks)

**Future Phase**: Exporter Development (Sprint 11+)
- Sprint 11: Exporter Foundation
- Sprint 12-14: Custom exporters (Lustre, InfiniBand, NVMe-oF)

## Architecture Overview

### Config Server Architecture

**Pattern**: Clean Architecture with clear layer separation

```
API Layer → Service Layer → Repository Layer → Domain Layer
                                                    ↓
                                               Database
```

**Key Principles**:
1. Domain independence (no framework dependencies in domain)
2. Dependency inversion (dependencies point inward)
3. ORM separation (GORM models separate from domain)

See: `services/config-server/.agent/ARCHITECTURE.md` for details

### CLI Architecture

**Pattern**: Independent Go module with client-server separation

```
CLI Commands → API Client → HTTP → Config Server API
```

**Key Principles**:
1. Standalone module (separate go.mod)
2. Configuration management (Viper)
3. Multiple output formats (table, JSON, YAML)

Location: `cli/`

## Technology Stack

### Backend
- **Language**: Go 1.25+
- **Web Framework**: Gin
- **ORM**: GORM v2
- **Database**: PostgreSQL 16+

### Monitoring
- **TSDB**: Prometheus 3.7+
- **Visualization**: Grafana 12.3+
- **Alerting**: Alertmanager 0.28+

### CLI
- **Framework**: Cobra
- **Config**: Viper
- **Output**: text/tabwriter, JSON, YAML

### Deployment
- **Containers**: Docker
- **Orchestration**: Kubernetes
- **Local Dev**: Docker Compose

## Development Workflow

### For New Features

1. **Check Current Sprint**: `.agent/planning/TRACKER.md`
2. **Read Service Guide**: `services/<service>/.agent/AGENT.md`
3. **Follow Architecture**: Clean Architecture principles
4. **Write Tests**: Unit + Integration tests
5. **Update Docs**: README and relevant guides
6. **Update Sprint**: Mark tasks complete

### For Bug Fixes

1. **Understand Architecture**: Read service ARCHITECTURE.md
2. **Locate Issue**: Use layer-appropriate tools
3. **Fix with Tests**: Add regression test
4. **Document**: Update relevant docs if needed

## Common Patterns

### Error Handling (Current)

**Service Layer**:
```go
// internal/service/errors.go
var (
    ErrNotFound = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
)
```

**Handler Layer**:
```go
// Map service errors to HTTP status
if errors.Is(err, service.ErrNotFound) {
    return http.StatusNotFound
}
```

### Error Handling (Sprint 6+)

**Unified Package**:
```go
// internal/errors/errors.go
var (
    ErrNotFound = errors.New("not found")
    ErrDuplicateKey = errors.New("duplicate key")
)

func FromGormError(err error) error {
    // Convert GORM errors to domain errors
}
```

### Repository Pattern

```go
// Always convert GORM models to domain models
func (r *repo) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
    var model models.Entity
    if err := r.db.First(&model, "id = ?", id).Error; err != nil {
        return nil, errors.FromGormError(err)  // Sprint 6+
    }
    return model.ToDomain(), nil
}
```

## Testing Strategy

### Current Status
- Testing infrastructure: Planned (Sprint 7)
- Target coverage: >70%
- Integration tests: Using testcontainers

### Test Organization
```
test/
├── unit/              # Unit tests
│   ├── domain/
│   ├── service/
│   └── ...
└── integration/       # Integration tests
    ├── repository/
    └── api/
```

## Key Decisions & Context

### Why CLI Instead of Web UI First?

**Decision**: Implement CLI before Web UI
**Reason**: Faster development, better for automation, essential for CI/CD
**Status**: CLI MVP completed (Sprint 5)
**Web UI**: Deferred to Phase 3 (Sprint 18+)

### Why Separate CLI Module?

**Decision**: Move CLI to `/cli/` with independent go.mod
**Reason**:
- Clear separation of concerns
- Independent versioning and releases
- CLI can be distributed without server code
- Easier maintenance

**Implementation**: Completed 2024-12-29

### Why Clean Architecture?

**Decision**: Use Clean Architecture pattern
**Reason**:
- Clear layer boundaries
- Testable business logic
- Framework-independent domain
- Easy to replace infrastructure (DB, API framework)

### Why Unified Error Handling (Sprint 6)?

**Decision**: Consolidate all errors in single package
**Problem**:
- GORM errors leak to service layer (62 places)
- Duplicate error definitions
- Inconsistent error responses

**Solution**: Create `internal/errors` package with domain errors

## Important Constraints

### Git Practices

1. **No AI Attribution**: Remove "Generated with Claude Code" and "Co-Authored-By: Claude" from commits
2. **Concise Messages**: Keep commit messages brief and clear
3. **Branch Naming**: Use `<type>/<description>` pattern
4. **Explicit Approval**: Always get user confirmation before committing

### Code Style

1. **Follow Go Standards**: Use `gofmt`, follow Go best practices
2. **Avoid Over-engineering**: Keep solutions simple and focused
3. **No Unnecessary Changes**: Don't refactor code outside the task scope
4. **Comments**: Only add where logic isn't self-evident

### .agent Files

1. **Never Commit**: All `.agent/` directories in `.gitignore`
2. **Agent-Only**: These files are for AI agent context
3. **Separate Concerns**: Follow root vs service .agent separation

## Quick Reference

### Find Sprint Information
```bash
# Current sprint
cat .agent/planning/TRACKER.md

# Sprint details
cat .agent/planning/sprints/current/sprint-06-error-handling.md

# All planned sprints
ls .agent/planning/sprints/planned/
```

### Find Architecture Information
```bash
# Config Server architecture
cat services/config-server/.agent/ARCHITECTURE.md

# Config Server agent guide
cat services/config-server/.agent/AGENT.md
```

### Build & Test
```bash
# Build Config Server
cd services/config-server
go build -o config-server cmd/config-server/main.go

# Build CLI
cd cli
go build -o aami cmd/aami/main.go

# Run tests (Sprint 7+)
go test ./...
```

## Getting Help

### When Starting a Task

1. Read `TRACKER.md` to understand current sprint
2. Read service-specific `AGENT.md` for architecture
3. Read current sprint file for task details
4. Ask user if requirements are unclear

### When Unsure About Structure

1. Check this file first
2. Check service-specific `ARCHITECTURE.md`
3. Follow existing patterns in codebase
4. Ask user for clarification

## Summary

- **Root .agent/**: Project planning, sprints, roadmap
- **Service .agent/**: Architecture, conventions, development guides
- **CLI**: Independent module at `/cli/`
- **Current Phase**: Config Server Stabilization (Sprint 6-10)
- **Next Phase**: Exporter Development (Sprint 11+)
- **Architecture**: Clean Architecture with layer separation
- **Testing**: Planned for Sprint 7 (>70% coverage)

Last Updated: 2024-12-29
