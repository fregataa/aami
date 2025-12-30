package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/fregataa/aami/config-server/internal/api"
	"github.com/fregataa/aami/config-server/internal/pkg/prometheus"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/fregataa/aami/config-server/internal/service"
	"gorm.io/gorm/logger"
)

func main() {
	// Initialize structured logger
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogger := slog.New(logHandler)

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

	// Initialize Prometheus components
	var fileManager *prometheus.RuleFileManager
	var prometheusClient *prometheus.PrometheusClient
	var ruleGenerator *service.PrometheusRuleGenerator

	if config.PrometheusConfig.RulePath != "" {
		// Initialize file manager
		fileManagerConfig := prometheus.RuleFileManagerConfig{
			BasePath:         config.PrometheusConfig.RulePath,
			BackupPath:       config.PrometheusConfig.BackupPath,
			EnableValidation: config.PrometheusConfig.ValidateRules,
			EnableBackup:     config.PrometheusConfig.BackupEnabled,
			PromtoolPath:     config.PrometheusConfig.PromtoolPath,
		}
		fileManager, err = prometheus.NewRuleFileManager(fileManagerConfig, slogger)
		if err != nil {
			log.Printf("Warning: Failed to initialize RuleFileManager: %v", err)
		}
	}

	if config.PrometheusConfig.URL != "" && config.PrometheusConfig.ReloadEnabled {
		// Initialize Prometheus client
		clientConfig := prometheus.PrometheusClientConfig{
			BaseURL: config.PrometheusConfig.URL,
			Timeout: config.PrometheusConfig.ReloadTimeout,
		}
		prometheusClient = prometheus.NewPrometheusClient(clientConfig, slogger)
	}

	if fileManager != nil {
		// Initialize rule generator
		ruleGenerator = service.NewPrometheusRuleGenerator(
			rm.AlertRule,
			rm.Group,
			fileManager,
			slogger,
		)
	}

	// Create and setup API server with Prometheus components
	server := api.NewServerWithPrometheus(rm, ruleGenerator, fileManager, prometheusClient)
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
	DBConfig         repository.Config
	Port             string
	PrometheusConfig PrometheusConfig
}

// PrometheusConfig holds Prometheus integration configuration
type PrometheusConfig struct {
	URL           string
	RulePath      string
	ReloadEnabled bool
	ReloadTimeout time.Duration
	ValidateRules bool
	PromtoolPath  string
	BackupEnabled bool
	BackupPath    string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	reloadTimeout, _ := time.ParseDuration(getEnv("PROMETHEUS_RELOAD_TIMEOUT", "30s"))

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
		PrometheusConfig: PrometheusConfig{
			URL:           getEnv("PROMETHEUS_URL", "http://localhost:9090"),
			RulePath:      getEnv("PROMETHEUS_RULE_PATH", "/etc/prometheus/rules/generated"),
			ReloadEnabled: getEnvBool("PROMETHEUS_RELOAD_ENABLED", true),
			ReloadTimeout: reloadTimeout,
			ValidateRules: getEnvBool("PROMETHEUS_VALIDATE_RULES", false),
			PromtoolPath:  getEnv("PROMETHEUS_PROMTOOL_PATH", ""),
			BackupEnabled: getEnvBool("PROMETHEUS_BACKUP_ENABLED", true),
			BackupPath:    getEnv("PROMETHEUS_BACKUP_PATH", ""),
		},
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

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
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
