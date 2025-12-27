# Tools

This directory contains development tools, utilities, and helper scripts.

## Directory Structure

```
tools/
├── generators/          # Code generators
├── validators/          # Configuration validators
├── migration/          # Migration utilities
└── dev-tools/          # Development utilities
```

## Available Tools

### Code Generators
- **Location**: `generators/`
- **Purpose**: Generate boilerplate code and configurations

Available generators:
- `gen-exporter.sh` - Generate custom exporter template
- `gen-alert-rule.sh` - Generate alert rule from template
- `gen-dashboard.sh` - Generate Grafana dashboard skeleton
- `gen-migration.sh` - Generate database migration files

Example usage:
```bash
# Generate custom exporter
./tools/generators/gen-exporter.sh my-custom-exporter

# Generate alert rule
./tools/generators/gen-alert-rule.sh HighDiskUsage
```

### Configuration Validators
- **Location**: `validators/`
- **Purpose**: Validate configuration files

Available validators:
- `validate-prometheus.sh` - Validate Prometheus config
- `validate-alertmanager.sh` - Validate Alertmanager config
- `validate-grafana.sh` - Validate Grafana dashboards
- `validate-env.sh` - Validate environment variables

Example usage:
```bash
# Validate Prometheus configuration
./tools/validators/validate-prometheus.sh ../config/prometheus/prometheus.yml

# Validate all configs
make validate-configs
```

### Migration Utilities
- **Location**: `migration/`
- **Purpose**: Database and configuration migration tools

Available utilities:
- `migrate-db.go` - Database migration tool
- `migrate-config.sh` - Configuration migration helper
- `rollback.sh` - Rollback to previous version
- `version-check.sh` - Check compatibility between versions

Example usage:
```bash
# Run database migration
go run ./tools/migration/migrate-db.go up

# Check version compatibility
./tools/migration/version-check.sh 1.0.0 2.0.0
```

### Development Utilities
- **Location**: `dev-tools/`
- **Purpose**: Development and debugging tools

Available utilities:
- `mock-server.go` - Mock Config Server for testing
- `load-test-data.sh` - Load test data into database
- `export-metrics.sh` - Export metrics for testing
- `debug-sd.sh` - Debug service discovery issues

Example usage:
```bash
# Start mock server
go run ./tools/dev-tools/mock-server.go

# Load test data
./tools/dev-tools/load-test-data.sh --env development
```

## Tool Installation

Some tools require additional dependencies:

```bash
# Install Go tools
go install github.com/golang/mock/mockgen@latest

# Install shell utilities
brew install jq yq shellcheck

# Python tools (if needed)
pip install pyyaml prometheus-client
```

## Using Tools in Development

### 1. Code Generation Workflow

```bash
# Generate new exporter
./tools/generators/gen-exporter.sh my-exporter
cd services/exporters/my-exporter

# Implement exporter logic
vim main.go

# Build and test
go build
./my-exporter
```

### 2. Pre-commit Validation

```bash
# Add to .git/hooks/pre-commit
#!/bin/bash
./tools/validators/validate-prometheus.sh config/prometheus/prometheus.yml
./tools/validators/validate-alertmanager.sh config/alertmanager/alertmanager.yml
```

### 3. Database Migration

```bash
# Create new migration
./tools/generators/gen-migration.sh add_bootstrap_tokens

# Edit migration files
vim scripts/db/migrations/000X_add_bootstrap_tokens.up.sql
vim scripts/db/migrations/000X_add_bootstrap_tokens.down.sql

# Apply migration
./tools/migration/migrate-db.go up
```

## Tool Development Guidelines

When creating new tools:

1. **Single Responsibility**: Each tool should do one thing well
2. **Documentation**: Include help text (`--help` flag)
3. **Error Handling**: Provide clear error messages
4. **Exit Codes**: Follow standard exit code conventions
5. **Idempotency**: Tools should be safe to run multiple times
6. **Comments**: Write all comments in English
7. **Testing**: Include unit tests for Go tools

Example tool template:
```go
// tools/example/main.go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    var help bool
    flag.BoolVar(&help, "help", false, "Show help message")
    flag.Parse()

    if help {
        printHelp()
        os.Exit(0)
    }

    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func printHelp() {
    fmt.Println("Usage: tool [OPTIONS]")
    fmt.Println("\nOptions:")
    flag.PrintDefaults()
}

func run() error {
    // Tool implementation
    return nil
}
```

## Makefile Targets

Common make targets that use tools:

```bash
# Generate all code
make generate

# Validate all configs
make validate

# Run migrations
make migrate-up

# Load test data
make load-test-data

# Clean generated files
make clean-generated
```

## Quick Links

- [Development Guide](../docs/DEVELOPMENT.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Makefile](../Makefile)
