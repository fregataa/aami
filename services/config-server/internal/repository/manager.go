package repository

import (
	"fmt"

	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Manager manages all repositories and database connections
// This ensures DB connections are only accessible within the repository layer
type Manager struct {
	db *gorm.DB

	// Repository instances
	Group          GroupRepository
	Target         TargetRepository
	TargetGroup    TargetGroupRepository
	Exporter       ExporterRepository
	ScriptTemplate ScriptTemplateRepository
	ScriptPolicy   ScriptPolicyRepository
	AlertTemplate  AlertTemplateRepository
	AlertRule      AlertRuleRepository
	BootstrapToken BootstrapTokenRepository
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	LogLevel logger.LogLevel
}

// NewManager creates a new repository manager with database connection
// This is the ONLY entry point for repository layer initialization
func NewManager(config Config) (*Manager, error) {
	// 1. Create database connection (internal to repository layer)
	db, err := connectDatabase(config)
	if err != nil {
		return nil, domainerrors.Wrap(err, "failed to connect database")
	}

	// 2. Initialize all repositories
	return newManagerWithDB(db), nil
}

// NewManagerWithDB creates a repository manager with an existing database connection
// WARNING: This is intended for TESTING ONLY. Production code should use NewManager()
func NewManagerWithDB(db *gorm.DB) *Manager {
	return newManagerWithDB(db)
}

// newManagerWithDB is the internal function that creates a manager with a given DB
func newManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		db:             db,
		Group:          NewGroupRepository(db),
		Target:         NewTargetRepository(db),
		TargetGroup:    NewTargetGroupRepository(db),
		Exporter:       NewExporterRepository(db),
		ScriptTemplate: NewScriptTemplateRepository(db),
		ScriptPolicy:   NewScriptPolicyRepository(db),
		AlertTemplate:  NewAlertTemplateRepository(db),
		AlertRule:      NewAlertRuleRepository(db),
		BootstrapToken: NewBootstrapTokenRepository(db),
	}
}

// connectDatabase creates a PostgreSQL database connection
// This function is private to the repository package
func connectDatabase(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Set default log level if not specified
	logLevel := config.LogLevel
	if logLevel == 0 {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, domainerrors.Wrap(err, "failed to open database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, domainerrors.Wrap(err, "failed to get database instance")
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0) // Reuse connections indefinitely

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, domainerrors.Wrap(err, "failed to ping database")
	}

	return db, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return domainerrors.Wrap(err, "failed to get database instance")
	}
	return sqlDB.Close()
}

// Health checks the database connection health
func (m *Manager) Health() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return domainerrors.Wrap(err, "failed to get database instance")
	}
	return sqlDB.Ping()
}

// GetDB returns the database instance (for migrations only)
// WARNING: This should only be used for database migrations
// Regular application code should NEVER call this method
func (m *Manager) GetDB() *gorm.DB {
	return m.db
}
