package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/fregataa/aami/config-server/internal/database"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Database   database.Config
	Prometheus PrometheusConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
}

// PrometheusConfig holds Prometheus integration configuration
type PrometheusConfig struct {
	// URL is the base URL of Prometheus server (e.g., "http://localhost:9090")
	URL string
	// RulePath is the directory path where rule files will be generated
	RulePath string
	// ReloadEnabled determines if automatic Prometheus reload is enabled
	ReloadEnabled bool
	// ReloadTimeout is the timeout for reload operations
	ReloadTimeout time.Duration
	// ValidateRules enables promtool validation before writing rule files
	ValidateRules bool
	// PromtoolPath is the path to promtool binary (optional, will search PATH if empty)
	PromtoolPath string
	// BackupEnabled enables backup of rule files before modification
	BackupEnabled bool
	// BackupPath is the directory for rule file backups (default: RulePath/.backup)
	BackupPath string
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "config_server")
	viper.SetDefault("database.sslmode", "disable")

	// Prometheus defaults
	viper.SetDefault("prometheus.url", "http://localhost:9090")
	viper.SetDefault("prometheus.rulepath", "/etc/prometheus/rules/generated")
	viper.SetDefault("prometheus.reloadenabled", true)
	viper.SetDefault("prometheus.reloadtimeout", "30s")
	viper.SetDefault("prometheus.validaterules", false)
	viper.SetDefault("prometheus.promtoolpath", "")
	viper.SetDefault("prometheus.backupenabled", true)
	viper.SetDefault("prometheus.backuppath", "")

	// Environment variables
	viper.AutomaticEnv()

	// Prometheus environment variable bindings
	_ = viper.BindEnv("prometheus.url", "PROMETHEUS_URL")
	_ = viper.BindEnv("prometheus.rulepath", "PROMETHEUS_RULE_PATH")
	_ = viper.BindEnv("prometheus.reloadenabled", "PROMETHEUS_RELOAD_ENABLED")
	_ = viper.BindEnv("prometheus.reloadtimeout", "PROMETHEUS_RELOAD_TIMEOUT")
	_ = viper.BindEnv("prometheus.validaterules", "PROMETHEUS_VALIDATE_RULES")
	_ = viper.BindEnv("prometheus.promtoolpath", "PROMETHEUS_PROMTOOL_PATH")
	_ = viper.BindEnv("prometheus.backupenabled", "PROMETHEUS_BACKUP_ENABLED")
	_ = viper.BindEnv("prometheus.backuppath", "PROMETHEUS_BACKUP_PATH")

	// Read from config file if exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
