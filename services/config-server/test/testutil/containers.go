package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresContainer represents a test PostgreSQL container
type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

// SetupTestDB creates a PostgreSQL testcontainer, runs migrations, and returns a Repository Manager
func SetupTestDB(t *testing.T) (*repository.Manager, func()) {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, err := createPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL container: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(pgContainer.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		pgContainer.Container.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		pgContainer.Container.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repository manager with the test database connection
	manager := repository.NewManagerWithDB(db)

	// Return cleanup function
	cleanup := func() {
		if err := pgContainer.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}

	return manager, cleanup
}

// createPostgresContainer creates and starts a PostgreSQL testcontainer
func createPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get mapped port
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Get host
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, mappedPort.Port())

	return &PostgresContainer{
		Container: container,
		DSN:       dsn,
	}, nil
}

// runMigrations runs the initial schema migration against the database
// For tests, we only need the initial schema (001_initial_schema.sql) which contains
// the complete current schema. Incremental migrations (002-005) are for existing
// deployments that need to upgrade from older schemas.
func runMigrations(db *gorm.DB) error {
	// Get the project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	// Only run the initial schema migration for fresh test databases
	migrationPath := filepath.Join(projectRoot, "migrations", "001_initial_schema.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if err := db.Exec(string(migrationSQL)).Error; err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// getProjectRoot finds the project root directory by looking for go.mod
func getProjectRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree until we find go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// TruncateAllTables removes all data from tables (useful for test isolation)
func TruncateAllTables(t *testing.T, manager *repository.Manager) {
	t.Helper()

	tables := []string{
		"target_groups",
		"exporters",
		"alert_rules",
		"bootstrap_tokens",
		"targets",
		"groups",
		"alert_templates",
	}

	// Get DB instance for direct SQL execution
	db := manager.GetDB()

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// AssertNoError is a test helper that fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: %v", fmt.Sprint(msgAndArgs...), err)
		} else {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

// AssertError is a test helper that fails the test if err is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected error but got nil", fmt.Sprint(msgAndArgs...))
		} else {
			t.Fatal("Expected error but got nil")
		}
	}
}
