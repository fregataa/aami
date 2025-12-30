package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/fregataa/aami/config-server/internal/config"
	"github.com/fregataa/aami/config-server/internal/repository"
	"github.com/fregataa/aami/config-server/internal/service"
	"gorm.io/gorm/logger"
)

func main() {
	// Parse command line flags
	force := flag.Bool("force", false, "Overwrite existing builtin templates")
	dryRun := flag.Bool("dry-run", false, "Preview changes without inserting")
	alertTemplatesFile := flag.String("alert-templates", "", "Path to alert templates YAML file")
	scriptTemplatesFile := flag.String("script-templates", "", "Path to script templates YAML file")
	scriptsDir := flag.String("scripts-dir", "", "Directory containing script files")
	flag.Parse()

	// Initialize logger
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogger := slog.New(logHandler)

	// Load database config from environment
	dbConfig := loadDBConfig()

	// Create repository manager
	rm, err := repository.NewManager(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer rm.Close()

	// Build defaults config from flags or defaults
	defaultsConfig := &config.DefaultsConfig{
		AlertTemplatesFile:  getConfigValue(*alertTemplatesFile, "SEED_ALERT_TEMPLATES_FILE", "configs/defaults/alert-templates.yaml"),
		ScriptTemplatesFile: getConfigValue(*scriptTemplatesFile, "SEED_SCRIPT_TEMPLATES_FILE", "configs/defaults/script-templates.yaml"),
		ScriptsDir:          getConfigValue(*scriptsDir, "SEED_SCRIPTS_DIR", "configs/defaults/scripts"),
	}

	// Create defaults loader service
	loader := service.NewDefaultsLoaderService(
		defaultsConfig,
		rm.AlertTemplate,
		rm.ScriptTemplate,
		slogger,
	)

	// Load all templates
	ctx := context.Background()
	result, err := loader.LoadAll(ctx, *force, *dryRun)
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Print summary
	fmt.Println()
	fmt.Println("=== Seed Templates Summary ===")
	if *dryRun {
		fmt.Println("Mode: DRY RUN (no changes made)")
	} else if *force {
		fmt.Println("Mode: FORCE (overwrote existing templates)")
	} else {
		fmt.Println("Mode: NORMAL (skipped existing templates)")
	}
	fmt.Println()
	fmt.Printf("Alert Templates:  %d created, %d updated, %d skipped\n",
		result.AlertTemplatesCreated, result.AlertTemplatesUpdated, result.AlertTemplatesSkipped)
	fmt.Printf("Script Templates: %d created, %d updated, %d skipped\n",
		result.ScriptTemplatesCreated, result.ScriptTemplatesUpdated, result.ScriptTemplatesSkipped)

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Printf("Errors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Seed completed successfully!")
}

func loadDBConfig() repository.Config {
	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	return repository.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "aami_config"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
		LogLevel: logger.Silent,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getConfigValue(flagValue, envKey, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}
