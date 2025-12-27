# AAMI Development Guide

Development environment setup and development guide for the AAMI project.

## Prerequisites

### Required Tools

- **Go 1.21+**: Config Server backend development
- **Node.js 20+**: Config Server UI development (optional)
- **Docker 20.10+**: Container build and execution
- **Docker Compose v2.0+**: Local development environment
- **PostgreSQL 15+**: Database (or run via Docker)
- **Redis 7+**: Caching (or run via Docker)

### Optional Tools

- **golangci-lint**: Go code linting
- **pnpm**: Node.js package manager (for UI development)
- **kubectl**: Kubernetes deployment (optional)
- **terraform**: Infrastructure provisioning (optional)

## Environment Setup

### 1. Install and Verify Go

```bash
# Check Go version
go version
# go version go1.21.x

# Check GOPATH
echo $GOPATH
# /Users/yourname/go

# Verify Go modules are enabled (default)
go env GO111MODULE
# on
```

### 2. Install and Verify Docker

```bash
# Check Docker version
docker --version
# Docker version 24.0.x

# Check Docker Compose version
docker-compose --version
# Docker Compose version v2.x.x

# Verify Docker is running
docker ps
```

### 3. Install Node.js (for UI development)

```bash
# Check Node.js version
node --version
# v20.x.x

# Install pnpm
npm install -g pnpm

# Check pnpm version
pnpm --version
# 8.x.x
```

### 4. Install golangci-lint

```bash
# macOS (Homebrew)
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Check version
golangci-lint --version
```

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/aami.git
cd aami
```

### 2. Start Local Development Environment

Start the complete monitoring stack (Prometheus, Grafana, PostgreSQL, Redis):

```bash
cd deploy/docker-compose

# Set up environment variables
cp .env.example .env
# Edit .env file (DB password, etc.)

# Start stack
docker-compose up -d

# Check logs
docker-compose logs -f

# Check status
docker-compose ps
```

**Access URLs:**
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Alertmanager: http://localhost:9093
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### 3. Config Server Development

#### Project Initialization

```bash
cd services/config-server

# Initialize Go modules
go mod init github.com/fregataa/aami/config-server

# Install dependencies
go mod tidy
```

#### Database Migration

```bash
# Run migration
go run cmd/migrate/main.go up

# Rollback migration
go run cmd/migrate/main.go down
```

#### Run Config Server

```bash
# Set environment variables
export DATABASE_URL="postgres://admin:changeme@localhost:5432/config_server?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export PORT="8080"

# Run server
go run cmd/server/main.go

# Or build and run
go build -o bin/config-server cmd/server/main.go
./bin/config-server
```

#### API Testing

```bash
# Health check
curl http://localhost:8080/api/v1/health

# List targets
curl http://localhost:8080/api/v1/targets

# Prometheus HTTP SD
curl http://localhost:8080/api/v1/sd/prometheus
```

### 4. Config Server UI Development (Optional)

```bash
cd services/config-server-ui

# Install pnpm globally
npm install -g pnpm

# Install dependencies
pnpm install

# Run development server
pnpm dev
# http://localhost:3000

# Build
pnpm build

# Check static build
ls -la out/
```

## Code Quality

### Linting

```bash
cd services/config-server

# Run golangci-lint
golangci-lint run

# Auto-fix issues
golangci-lint run --fix

# Lint specific directory
golangci-lint run ./internal/api/...
```

### Formatting

```bash
# Go formatting
go fmt ./...

# goimports (auto-organize imports)
goimports -w .

# Check formatting for all files
gofmt -l .
```

### Testing

```bash
cd services/config-server

# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Measure coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race condition detection
go test -race ./...

# Test specific package
go test ./internal/api/...

# Run specific test
go test -run TestCreateTarget ./internal/api/...
```

## Debugging

### VS Code Debugging

Create `.vscode/launch.json` file:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Config Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/services/config-server/cmd/server",
      "env": {
        "DATABASE_URL": "postgres://admin:changeme@localhost:5432/config_server?sslmode=disable",
        "REDIS_URL": "redis://localhost:6379",
        "PORT": "8080"
      },
      "args": []
    }
  ]
}
```

### Using Delve (dlv)

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run in debug mode
cd services/config-server
dlv debug cmd/server/main.go

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

## Branch Strategy

```
main
  ├── develop
  │   ├── feature/bootstrap-script
  │   ├── feature/ssh-agent
  │   └── feature/fleet-management
  ├── bugfix/fix-login-validation
  └── hotfix/critical-security-fix
```

### Branch Naming Convention

- `feature/*`: New features
- `bugfix/*`: Bug fixes
- `hotfix/*`: Urgent fixes
- `refactor/*`: Refactoring
- `docs/*`: Documentation updates
- `test/*`: Test additions

## Commit Message Convention

```bash
# Format
<type>: <subject>

<body>

# Example
feat: Add bootstrap script auto registration

- Implement bootstrap token management API
- Add hardware auto-detection logic
- Create bootstrap.sh script

# Types
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Code formatting
- refactor: Refactoring
- test: Test additions
- chore: Build/tool changes
```

## Build and Deployment

### Local Build

```bash
cd services/config-server

# Build binary
go build -o bin/config-server cmd/server/main.go

# Build static binary (disable CGO)
CGO_ENABLED=0 go build -o bin/config-server cmd/server/main.go

# Release build (optimized)
go build -ldflags="-s -w" -o bin/config-server cmd/server/main.go

# Cross-compile (Linux)
GOOS=linux GOARCH=amd64 go build -o bin/config-server-linux cmd/server/main.go
```

### Docker Build

```bash
cd services/config-server

# Build Docker image
docker build -t aami/config-server:latest .

# Build for specific platforms
docker buildx build --platform linux/amd64,linux/arm64 -t aami/config-server:latest .

# Run image
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://admin:changeme@host.docker.internal:5432/config_server?sslmode=disable" \
  -e REDIS_URL="redis://host.docker.internal:6379" \
  aami/config-server:latest
```

## Troubleshooting

### Reset Go Module Cache

```bash
go clean -modcache
go mod download
```

### Restart Docker Containers

```bash
cd deploy/docker-compose

# Stop and remove all containers
docker-compose down

# Remove volumes (WARNING: data loss!)
docker-compose down -v

# Restart
docker-compose up -d
```

### PostgreSQL Connection Errors

```bash
# Check PostgreSQL container logs
docker-compose logs postgres

# Test PostgreSQL connection
psql -h localhost -U admin -d config_server

# Re-run migrations
cd services/config-server
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up
```

## Additional Resources

- [PLAN.md](../PLAN.md) - Full architecture and requirements
- [sprint-plan.md](../sprint-plan.md) - Detailed sprint plan
- [Go Official Documentation](https://go.dev/doc/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## Support

If you encounter issues, please check:

1. [GitHub Issues](https://github.com/your-org/aami/issues)
2. [Troubleshooting Guide](./TROUBLESHOOTING.md)
3. Slack channel: #aami-dev

---

**Happy Coding! 🚀**
