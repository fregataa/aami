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
│  Database (PostgreSQL)              │
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

📚 **Quick Start Guide**: [QUICKSTART.md](./QUICKSTART.md) - 상세한 구동 방법 (로컬/클라우드)

### Fastest Way - Docker Compose (권장)

```bash
# 1. Clone the repository
git clone https://github.com/fregataa/aami.git
cd aami/services/config-server

# 2. Start all services (PostgreSQL + Config Server)
docker-compose up -d

# 3. Check health
curl http://localhost:8080/health

# 4. Test API
curl http://localhost:8080/api/v1/namespaces
```

### Alternative - Direct Installation

**Prerequisites**: Go 1.21+, PostgreSQL 15+

```bash
# 1. Install dependencies
go mod download

# 2. Set up environment
cp .env.example .env
# Edit .env with your database settings

# 3. Build and run
go build -o config-server ./cmd/config-server
./config-server
```

자세한 구동 방법은 [QUICKSTART.md](./QUICKSTART.md)를 참조하세요.

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
# Prometheus HTTP Service Discovery
GET    /api/v1/sd/prometheus                      # All targets (with filters)
GET    /api/v1/sd/prometheus/active               # Active targets only
GET    /api/v1/sd/prometheus/group/:groupId       # Group-specific targets
GET    /api/v1/sd/prometheus/namespace/:nsId      # Namespace-specific targets

# Prometheus File Service Discovery
POST   /api/v1/sd/prometheus/file                 # Generate file SD (custom filter)
POST   /api/v1/sd/prometheus/file/active          # Generate file SD (active only)
POST   /api/v1/sd/prometheus/file/group/:groupId  # Generate file SD (by group)
POST   /api/v1/sd/prometheus/file/namespace/:nsId # Generate file SD (by namespace)

# Health Check
GET    /health                                    # Complete health status
GET    /health/ready                              # Readiness probe (K8s)
GET    /health/live                               # Liveness probe (K8s)
```

## Docker & Kubernetes

### Docker Compose

```bash
# Production environment
docker-compose up -d

# Development environment (with hot reload)
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose logs -f config-server

# Stop services
docker-compose down
```

### Kubernetes

```bash
# Deploy to Kubernetes
kubectl apply -k k8s/

# Check status
kubectl get pods -n aami
kubectl get svc -n aami

# View logs
kubectl logs -f deployment/config-server -n aami

# Port forward for testing
kubectl port-forward svc/config-server 8080:80 -n aami
```

자세한 내용은 [k8s/README.md](./k8s/README.md)를 참조하세요.

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
```

## Current Implementation Status

### ✅ Completed (Sprint 1-5)

**Core Features**:
- Project structure and dependencies
- Domain models with business logic
- Database migrations and schema management
- Repository interfaces and GORM implementations
- Database connection (PostgreSQL)
- Configuration management
- Error handling utilities
- DTOs and validation
- Service layer implementation
- API handlers and routing
- Bootstrap functionality
- CheckTemplate/CheckInstance system
- Target-Group relationship with junction table
- Priority system (higher number = higher priority)

**Sprint 5 - Operations Ready**:
- ✅ **Service Discovery**: Prometheus HTTP SD & File SD
- ✅ **Health Check**: Readiness/Liveness probes with detailed component status
- ✅ **Containerization**: Optimized Dockerfile with security best practices
- ✅ **Docker Compose**: Development and production environments
- ✅ **Kubernetes**: Complete manifests (Deployment, Service, Ingress, HPA)

### 📋 Backlog (Sprint 6+)

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
