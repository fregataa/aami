package database

import (
	"fmt"

	"github.com/fregataa/aami/config-server/internal/domain"
	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, domainerrors.Wrap(err, "failed to connect to database")
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, domainerrors.Wrap(err, "failed to get database instance")
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, domainerrors.Wrap(err, "failed to ping database")
	}

	return db, nil
}

// AutoMigrate runs auto-migration for all domain models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Group{},
		&domain.Target{},
		&domain.Exporter{},
		&domain.AlertTemplate{},
		&domain.AlertRule{},
		&domain.BootstrapToken{},
	)
}
