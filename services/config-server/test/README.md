# Config Server Tests

This directory contains all tests for the AAMI Config Server.

## Directory Structure

```
test/
├── unit/              # Unit tests - fast, no external dependencies
│   ├── domain/        # Domain model business logic tests
│   └── repository/    # Repository interface tests (with mocks)
├── integration/       # Integration tests - use real database
│   └── repository/    # Repository implementation tests with PostgreSQL
├── testutil/          # Shared test utilities
│   ├── fixtures.go    # Test data fixtures
│   └── containers.go  # Testcontainers helpers
└── README.md          # This file
```

## Test Types

### Unit Tests (`test/unit/`)

Unit tests verify individual components in isolation without external dependencies.

**Characteristics:**
- Fast execution (< 1ms per test)
- No database or network calls
- Use mocks/stubs for dependencies
- Test business logic and validation

**Examples:**
- Domain model validation methods
- Business logic in domain models
- Repository interface contracts

**Run:**
```bash
make test-unit
# or
go test ./test/unit/...
```

### Integration Tests (`test/integration/`)

Integration tests verify components working together with real external dependencies.

**Characteristics:**
- Slower execution (using testcontainers)
- Real PostgreSQL database
- Test actual SQL queries and transactions
- Verify database constraints

**Examples:**
- Repository CRUD operations
- Complex queries (recursive CTEs)
- Transaction handling
- Foreign key constraints

**Run:**
```bash
make test-integration
# or
go test ./test/integration/...
```

## Test Utilities (`test/testutil/`)

Shared utilities and helpers for all tests.

### fixtures.go

Provides functions to create test data:

```go
// Example usage
func TestExample(t *testing.T) {
    group := testutil.NewTestGroup("production", domain.NamespaceEnvironment)
    target := testutil.NewTestTarget("server-01", "192.168.1.100", group.ID)
    // ... test code
}
```

### containers.go

Provides testcontainers setup helpers:

```go
// Example usage
func TestExample(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()
    // ... test code
}
```

## Running Tests

### All Tests
```bash
make test
# or
go test ./test/...
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests Only
```bash
make test-integration
```

### With Coverage
```bash
make test-coverage
```

### Specific Package
```bash
go test ./test/unit/domain -v
go test ./test/integration/repository -v -run TestGroupRepository
```

## Writing Tests

### Unit Test Example

**File:** `test/unit/domain/group_test.go`

```go
package domain_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/fregataa/aami/config-server/internal/domain"
)

func TestGroup_IsRoot(t *testing.T) {
    tests := []struct {
        name     string
        parentID *string
        want     bool
    }{
        {
            name:     "root group has no parent",
            parentID: nil,
            want:     true,
        },
        {
            name:     "child group has parent",
            parentID: stringPtr("parent-id"),
            want:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            group := &domain.Group{
                ParentID: tt.parentID,
            }
            assert.Equal(t, tt.want, group.IsRoot())
        })
    }
}

func stringPtr(s string) *string {
    return &s
}
```

### Integration Test Example

**File:** `test/integration/repository/group_test.go`

```go
package repository_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/fregataa/aami/config-server/internal/domain"
    "github.com/fregataa/aami/config-server/internal/repository"
    "github.com/fregataa/aami/config-server/test/testutil"
)

func TestGroupRepository_Create(t *testing.T) {
    db, cleanup := testutil.SetupTestDB(t)
    defer cleanup()

    repo := repository.NewGroupRepository(db)

    group := &domain.Group{
        Name:      "production",
        Namespace: domain.NamespaceEnvironment,
    }

    err := repo.Create(context.Background(), group)
    require.NoError(t, err)
    assert.NotEmpty(t, group.ID)

    // Verify in database
    retrieved, err := repo.GetByID(context.Background(), group.ID)
    require.NoError(t, err)
    assert.Equal(t, group.Name, retrieved.Name)
}
```

## Best Practices

1. **Test Naming**: Use descriptive names that explain what is being tested
   - Good: `TestGroup_IsRoot_WhenParentIsNil_ReturnsTrue`
   - Bad: `TestGroup1`

2. **Table-Driven Tests**: Use table-driven tests for multiple scenarios
   ```go
   tests := []struct {
       name string
       input string
       want bool
   }{
       // test cases
   }
   ```

3. **Arrange-Act-Assert**: Structure tests clearly
   ```go
   // Arrange
   group := &domain.Group{...}

   // Act
   result := group.IsRoot()

   // Assert
   assert.True(t, result)
   ```

4. **Test Independence**: Each test should be independent and repeatable

5. **Cleanup**: Always clean up resources (defer cleanup functions)

6. **Coverage Goal**: Aim for >60% coverage for Sprint 2

## CI/CD Integration

Tests run automatically on:
- Pull request creation
- Push to main branch
- Manual workflow trigger

See `.github/workflows/test.yml` for CI configuration.

## Troubleshooting

### Integration Tests Fail

If integration tests fail with database connection errors:

1. Check Docker is running (testcontainers requires Docker)
2. Check Docker resources (memory, disk space)
3. Check network connectivity

### Tests Are Slow

If tests are running slowly:

1. Run only unit tests during development: `make test-unit`
2. Use `-short` flag to skip slow tests: `go test -short ./test/...`
3. Run specific tests: `go test -run TestSpecificTest ./test/...`

### Coverage Not Generated

If coverage report is not generated:

```bash
# Clean old coverage files
rm coverage.out coverage.html

# Regenerate
make test-coverage
```
