package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/fregataa/aami/config-server/internal/database"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database database.Config
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
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

	// Environment variables
	viper.AutomaticEnv()

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
