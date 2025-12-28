## Config Server

The AAMI Config Server is a REST API service that manages the configuration for the monitoring infrastructure. It provides centralized management of targets, groups, exporters, alerts, and check settings.

## Architecture

The Config Server follows a Clean Architecture pattern with clear separation of concerns:

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
│  Database (PostgreSQL + Redis)      │
└─────────────────────────────────────┘
```

**Key Principles**:
- **Domain Independence**: Domain models contain no framework dependencies (no GORM tags)
- **Dependency Rule**: Dependencies flow downward (API → Service → Repository → Domain)
- **ORM Separation**: Repository layer uses ORM models that convert to/from domain models

📖 **For Developers**: See [AGENT.md](./.agent/docs/AGENT.md) for detailed architecture guidelines and coding patterns.

## Features

- **Hierarchical Group Management**: Three independent namespaces (infrastructure, logical, environment) with unlimited depth
- **Target Registration**: Register and manage monitored servers/nodes with multiple group memberships
- **Exporter Configuration**: Configure metric exporters (Node Exporter, DCGM Exporter, custom)
- **Alert Management**: Template-based alert rules with policy inheritance
- **Check Settings**: Hierarchical configuration settings with merge strategies
- **Bootstrap Tokens**: Secure auto-registration tokens for new nodes
- **Service Discovery**: Generate Prometheus SD configuration

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### Installation

```bash
# Clone the repository
git clone https://github.com/fregataa/aami.git
cd aami/services/config-server

# Install dependencies
go mod download

# Set up configuration
# Option 1: Using environment variables
cp .env.example .env
# Edit .env with your settings

# Option 2: Using config file
cp config.yaml.example config.yaml
# Edit config.yaml with your settings

# Run database migrations
psql -U postgres -f migrations/001_initial_schema.sql

# Build and run
go build -o config-server ./cmd/server
./config-server
```

### Development

```bash
# Run with hot reload (using air)
air

# Run all tests
go test ./...

# Run unit tests only
go test ./test/unit/...

# Run integration tests only
go test ./test/integration/...

# Run tests with coverage
go test -cover ./test/...

# Build
make build

# Run
make run
```

## Database Schema

### Core Tables

- **groups**: Hierarchical organization with three namespaces
- **targets**: Monitored servers with status tracking
- **exporters**: Metric collector configurations
- **alert_templates**: Reusable alert definitions
- **alert_rules**: Group-specific alert configurations
- **check_templates**: Reusable check script definitions
- **check_instances**: Scope-specific check template applications (Global/Namespace/Group)
- **bootstrap_tokens**: Auto-registration tokens

### Relationships

- Targets belong to one primary group and multiple secondary groups
- Exporters belong to targets
- Alert rules reference templates and belong to groups
- Check instances reference templates and can be scoped to Global/Namespace/Group
- Bootstrap tokens reference a default group

## API Endpoints

### Groups

```
GET    /api/v1/groups                  # List all groups
POST   /api/v1/groups                  # Create a group
GET    /api/v1/groups/:id              # Get group by ID
PUT    /api/v1/groups/:id              # Update group
DELETE /api/v1/groups/:id              # Delete group
GET    /api/v1/groups/:id/children     # Get child groups
GET    /api/v1/groups/:id/ancestors    # Get ancestor groups
```

### Targets

```
GET    /api/v1/targets                 # List all targets
POST   /api/v1/targets                 # Register a target
GET    /api/v1/targets/:id             # Get target by ID
PUT    /api/v1/targets/:id             # Update target
DELETE /api/v1/targets/:id             # Delete target
GET    /api/v1/targets/:id/effective-rules  # Get effective alert rules
```

### Alert Management

```
GET    /api/v1/alert-templates         # List alert templates
GET    /api/v1/alert-templates/:id     # Get template by ID
POST   /api/v1/groups/:id/alert-rules  # Apply alert rule to group
GET    /api/v1/groups/:id/alert-rules  # Get group's alert rules
DELETE /api/v1/alert-rules/:id         # Delete alert rule
```

### Check Management

```
# Check Templates
GET    /api/v1/check-templates         # List check templates
POST   /api/v1/check-templates         # Create check template
GET    /api/v1/check-templates/:id     # Get template by ID
PUT    /api/v1/check-templates/:id     # Update template
DELETE /api/v1/check-templates/:id     # Soft delete template

# Check Instances
GET    /api/v1/check-instances         # List check instances
POST   /api/v1/check-instances         # Create check instance
GET    /api/v1/check-instances/:id     # Get instance by ID
PUT    /api/v1/check-instances/:id     # Update instance
DELETE /api/v1/check-instances/:id     # Soft delete instance

# Node API
GET    /api/v1/checks/node/:hostname   # Get effective checks for node
```

### Bootstrap

```
POST   /api/v1/bootstrap/tokens        # Create bootstrap token
GET    /api/v1/bootstrap/tokens        # List bootstrap tokens
DELETE /api/v1/bootstrap/tokens/:id    # Delete bootstrap token
POST   /api/v1/bootstrap/register      # Auto-register using token
```

### Service Discovery

```
GET    /api/v1/sd/prometheus           # Prometheus HTTP SD endpoint
GET    /api/v1/sd/prometheus/file      # Generate Prometheus file SD
```

## Configuration

Configuration can be provided via:
1. Environment variables
2. Config file (`config.yaml`)
3. Command-line flags

### Environment Variables

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=config_server
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Config File Example

```yaml
server:
  port: 8080
  host: 0.0.0.0

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: config_server
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

## Current Implementation Status

### ✅ Completed

- Project structure and dependencies
- Domain models with business logic
- Database migrations and schema management
- Repository interfaces and GORM implementations
- Database connection (PostgreSQL + Redis)
- Configuration management
- Error handling utilities
- DTOs and validation
- Service layer implementation
- API handlers and routing
- Bootstrap functionality
- CheckTemplate/CheckInstance system
- Target-Group relationship with junction table
- Priority system (higher number = higher priority)

### 🚧 In Progress (Sprint 5)

- Service Discovery generation
- Health check enhancements
- Docker support
- Kubernetes deployment manifests
- Structured logging

### 📋 Backlog (Future Sprints)

- Unit tests and integration tests
- API documentation (OpenAPI/Swagger)
- Authentication and authorization
- Performance optimization
- Backup and recovery

📖 **Sprint Planning**: See [SPRINT_PLAN.md](./.agent/docs/SPRINT_PLAN.md) for detailed sprint roadmap.

## Development Guidelines

### Code Style

- Use `gofmt` for formatting
- Follow Go best practices
- Write meaningful commit messages
- Add comments for exported functions

### Testing

- Write unit tests for business logic
- Use testcontainers for integration tests
- Aim for >70% test coverage

### Database Migrations

- Never modify existing migrations
- Create new migrations for schema changes
- Test both up and down migrations

## Contributing

1. Create a feature branch
2. Make your changes
3. Write tests
4. Run linter: `golangci-lint run`
5. Submit a pull request

## License

MIT License - see [LICENSE](../../LICENSE) for details
