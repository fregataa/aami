package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fregataa/aami/config-server/internal/api"
	"github.com/fregataa/aami/config-server/internal/repository"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration from environment
	config := loadConfig()

	// Create repository manager with database connection
	rm, err := repository.NewManager(config.DBConfig)
	if err != nil {
		log.Fatalf("Failed to create repository manager: %v", err)
	}
	defer rm.Close()

	// Validate database schema
	if err := validateSchema(rm); err != nil {
		log.Fatalf("Database schema validation failed: %v", err)
	}

	// Create and setup API server
	server := api.NewServer(rm)
	router := server.SetupRouter()

	// Start server
	addr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Starting config-server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Config holds application configuration
type Config struct {
	DBConfig repository.Config
	Port     string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return Config{
		DBConfig: repository.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "aami_config"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			LogLevel: logger.Info,
		},
		Port: getEnv("PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// validateSchema validates that all required database tables exist
// The server does NOT run migrations automatically - migrations must be run manually
// before starting the server using psql or a migration tool like goose
func validateSchema(rm *repository.Manager) error {
	log.Println("Validating database schema...")

	db := rm.GetDB()

	// List of required tables
	requiredTables := []string{
		"namespaces",
		"groups",
		"targets",
		"target_groups",
		"exporters",
		"alert_templates",
		"alert_rules",
		"script_templates",
		"script_policies",
		"bootstrap_tokens",
	}

	var missingTables []string

	// Check each required table exists
	for _, table := range requiredTables {
		var exists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = ?
			)
		`, table).Scan(&exists).Error

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			missingTables = append(missingTables, table)
		}
	}

	// If any tables are missing, fail with clear error message
	if len(missingTables) > 0 {
		log.Printf("ERROR: Missing required database tables: %v", missingTables)
		log.Println("Please run database migrations before starting the server:")
		log.Println("  psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql")
		return fmt.Errorf("missing required tables: %v", missingTables)
	}

	log.Println("✓ Database schema validation successful - all required tables exist")
	return nil
}
