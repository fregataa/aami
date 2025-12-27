# Services

This directory contains microservice implementations for the AAMI platform.

## Directory Structure

```
services/
├── config-server/       # Config Server (Go) - Core API service
├── config-server-ui/    # Config Server UI (Next.js, optional)
└── exporters/          # Custom metric exporters
```

## Services Overview

### Config Server
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15, Redis 7
- **Purpose**: Central management API for targets, groups, and alert rules

Key features:
- Group and target management
- Dynamic service discovery (Prometheus SD)
- Alert rule customization
- Check configuration management
- Bootstrap token management
- Fleet deployment coordination

Quick start:
```bash
cd config-server
go mod download
go run cmd/server/main.go
```

See [config-server/README.md](config-server/README.md) for details.

### Config Server UI (Optional)
- **Framework**: Next.js 15
- **Build**: Static export
- **Purpose**: Web-based management interface

Features:
- Target and group management UI
- Dashboard overview
- Alert rule configuration
- Bootstrap token management
- Real-time deployment monitoring

Quick start:
```bash
cd config-server-ui
pnpm install
pnpm dev
```

See [config-server-ui/README.md](config-server-ui/README.md) for details.

### Custom Exporters
- **Purpose**: Metric exporters for specialized hardware
- **Language**: Go (recommended), Python (for scripts)

Example exporters:
- NPU exporters (vendor-specific)
- InfiniBand/RoCE network metrics
- Parallel filesystem metrics (Lustre, BeeGFS)
- Custom hardware sensors

See [exporters/README.md](exporters/README.md) for details.

## Development

### Prerequisites
- Go 1.21+ (for Go services)
- Node.js 20+ and pnpm (for UI)
- Docker (for local testing)
- PostgreSQL 15+ and Redis 7+ (or Docker)

### Running Locally

1. Start dependencies:
```bash
cd ../deploy/docker-compose
docker-compose up -d postgres redis
```

2. Run Config Server:
```bash
cd config-server
export DATABASE_URL="postgres://admin:changeme@localhost:5432/config_server?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
go run cmd/server/main.go
```

3. Run UI (optional):
```bash
cd config-server-ui
pnpm dev
```

### Testing

```bash
# Config Server tests
cd config-server
go test ./...

# UI tests
cd config-server-ui
pnpm test
```

## API Documentation

Config Server API documentation:
- Swagger/OpenAPI: http://localhost:8080/swagger
- API Reference: [../docs/api-reference.md](../docs/api-reference.md)

## Quick Links

- [Development Guide](../docs/DEVELOPMENT.md)
- [API Reference](../docs/api-reference.md)
- [Architecture](../docs/architecture.md)
