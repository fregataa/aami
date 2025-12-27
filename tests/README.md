# Tests

This directory contains integration tests, end-to-end tests, and test utilities.

## Directory Structure

```
tests/
├── integration/         # Integration tests
├── e2e/                # End-to-end tests
├── performance/        # Performance and load tests
└── fixtures/           # Test data and fixtures
```

## Test Categories

### Integration Tests
- **Location**: `integration/`
- **Purpose**: Test interactions between components
- **Framework**: Go testing, testcontainers

Test scenarios:
- Config Server API integration
- PostgreSQL data persistence
- Redis caching behavior
- Prometheus service discovery
- Alert rule generation

Example:
```bash
cd tests/integration
go test -v ./...
```

### End-to-End Tests
- **Location**: `e2e/`
- **Purpose**: Test complete workflows
- **Framework**: Go testing, Docker Compose

Test scenarios:
- Complete deployment workflow
- Target registration and monitoring
- Alert rule application and evaluation
- Bootstrap script execution
- Fleet deployment

Example:
```bash
cd tests/e2e
docker-compose -f docker-compose.test.yml up -d
go test -v -tags=e2e ./...
docker-compose -f docker-compose.test.yml down
```

### Performance Tests
- **Location**: `performance/`
- **Purpose**: Load testing and benchmarking
- **Tools**: k6, Apache Bench, Go benchmarks

Test scenarios:
- API endpoint performance
- Database query performance
- Redis cache hit rates
- Concurrent target updates
- Service discovery scalability

Example:
```bash
cd tests/performance
k6 run api-load-test.js
```

### Test Fixtures
- **Location**: `fixtures/`
- **Purpose**: Test data and mock configurations

Contents:
- Sample target configurations
- Mock Prometheus responses
- Example alert rules
- Test database seeds
- Mock exporter metrics

## Running Tests

### Prerequisites

```bash
# Install dependencies
go install gotest.tools/gotestsum@latest
docker pull postgres:15
docker pull redis:7
```

### Run All Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test suite
make test-integration
make test-e2e
make test-performance
```

### Run Individual Tests

```bash
# Run specific integration test
go test -v ./tests/integration -run TestTargetCreation

# Run with race detector
go test -race ./tests/integration/...

# Run with timeout
go test -timeout 5m ./tests/e2e/...
```

## Writing Tests

### Test Structure

```go
// tests/integration/target_test.go
package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestTargetCreation(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // Act
    target := createTarget(db, "test-host")

    // Assert
    assert.NotNil(t, target)
    assert.Equal(t, "test-host", target.Hostname)
}
```

### Test Conventions

1. **Naming**: `TestFunctionName` format
2. **Subtests**: Use `t.Run()` for test cases
3. **Cleanup**: Use `defer` or `t.Cleanup()`
4. **Isolation**: Each test should be independent
5. **Comments**: Explain complex test logic in English

### Test Data Management

```go
// Use fixtures for consistent test data
func loadFixture(t *testing.T, name string) []byte {
    data, err := os.ReadFile(filepath.Join("fixtures", name))
    require.NoError(t, err)
    return data
}
```

## Continuous Integration

Tests are automatically run on:
- Pull requests
- Pushes to main/develop branches
- Nightly builds

CI configuration: [../.github/workflows/ci.yml](../.github/workflows/ci.yml)

## Test Coverage

View coverage reports:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# Coverage summary
go tool cover -func=coverage.out
```

Target coverage:
- Unit tests: > 80%
- Integration tests: > 70%
- Overall: > 75%

## Debugging Tests

```bash
# Verbose output
go test -v ./tests/...

# Print test logs
go test -v ./tests/... 2>&1 | tee test.log

# Run with debugger (delve)
dlv test ./tests/integration -- -test.run TestTargetCreation
```

## Quick Links

- [Development Guide](../docs/DEVELOPMENT.md)
- [CI/CD Configuration](../.github/workflows/)
- [Contributing Guide](../CONTRIBUTING.md)
