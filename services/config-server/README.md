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

📖 **For Developers**: See [ARCHITECTURE.md](./.agent/ARCHITECTURE.md) for detailed architecture guidelines and coding patterns.

## Features

- **Flat Group Management**: Organize targets with flat groups (no hierarchy)
- **Target Registration**: Register and manage monitored servers/nodes with multiple group memberships
- **Exporter Configuration**: Configure metric exporters (Node Exporter, DCGM Exporter, custom)
- **Alert Management**: Template-based alert rules with group-level policies
- **Script Policies**: Configurable script policies with merge strategies
- **Bootstrap Tokens**: Secure auto-registration tokens for new nodes
- **Service Discovery**: Generate Prometheus SD configuration

## Getting Started

📚 **Quick Start Guide**: [QUICKSTART.md](./QUICKSTART.md) - Detailed setup instructions (local/cloud)

### Fastest Way - Docker Compose (Recommended)

```bash
# 1. Clone the repository
git clone https://github.com/fregataa/aami.git
cd aami/deploy/docker-compose

# 2. Start all services (PostgreSQL + Config Server + Monitoring Stack)
docker-compose up -d

# 3. Check health
curl http://localhost:8080/health

# 4. Test API
curl http://localhost:8080/api/v1/namespaces
```

### Alternative - Direct Installation

**Prerequisites**: Go 1.25+, PostgreSQL 16+

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

For detailed setup instructions, please refer to [QUICKSTART.md](./QUICKSTART.md).

## CLI Tool

AAMI provides a command-line interface for easy management of the monitoring infrastructure.

### Install CLI

```bash
# Build from source
cd ../../cli
go build -o aami cmd/aami/main.go

# Install to system path (optional)
sudo cp aami /usr/local/bin/
```

### Quick Start

```bash
# Initialize configuration
aami config init
aami config set server http://localhost:8080

# Create namespace and group
aami namespace create --name=production --priority=100
aami group create --name=web-tier --namespace=<ns-id>

# Register a target
aami target create --hostname=web-01 --ip=10.0.1.100 --group=<group-id>

# Create bootstrap token for automated registration
aami bootstrap-token create --name=prod-token --max-uses=50 --expires=30d
```

📖 **Full CLI Documentation**: See [CLI directory](../../cli/) for complete documentation

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

- **groups**: Flat organizational units for targets
- **targets**: Monitored servers with status tracking
- **exporters**: Metric collector configurations
- **alert_templates**: Reusable alert definitions
- **alert_rules**: Group-specific alert configurations
- **script_templates**: Reusable script definitions
- **script_policies**: Group-level script configurations
- **bootstrap_tokens**: Auto-registration tokens

### Relationships

- Targets can belong to multiple groups (M:N via target_groups)
- Exporters belong to targets
- Alert rules reference templates and belong to groups
- Script policies reference templates and can be scoped to global or group level
- Bootstrap tokens are used for auto-registration

## API Endpoints

### Groups

```
GET    /api/v1/groups                  # List all groups
POST   /api/v1/groups                  # Create a group
GET    /api/v1/groups/:id              # Get group by ID
PUT    /api/v1/groups/:id              # Update group
DELETE /api/v1/groups/:id              # Delete group
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
# Navigate to deployment directory
cd ../../deploy/docker-compose

# Start all services (full stack)
docker-compose up -d

# View Config Server logs
docker-compose logs -f config-server

# Stop services
docker-compose down
```

**Note**: Docker Compose files are located in `deploy/docker-compose/` to manage the complete AAMI stack.

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

For more details, please refer to [k8s/README.md](./k8s/README.md).

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

## Project Status

For detailed project status and sprint planning, see [Sprint Tracker](../../.agent/planning/TRACKER.md).

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
