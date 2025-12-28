package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration
type Config struct {
	Server string `mapstructure:"server"`
	Default DefaultConfig `mapstructure:"default"`
	Output OutputConfig `mapstructure:"output"`
}

// DefaultConfig represents default settings
type DefaultConfig struct {
	Namespace string `mapstructure:"namespace"`
	Output    string `mapstructure:"output"`
}

// OutputConfig represents output settings
type OutputConfig struct {
	NoHeaders bool `mapstructure:"no-headers"`
	Color     bool `mapstructure:"color"`
}

var (
	defaultConfig = Config{
		Server: "http://localhost:8080",
		Default: DefaultConfig{
			Namespace: "",
			Output:    "table",
		},
		Output: OutputConfig{
			NoHeaders: false,
			Color:     true,
		},
	}
)

// Load loads the configuration
func Load() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &defaultConfig, nil
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration
func Save(cfg *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	viper.Set("server", cfg.Server)
	viper.Set("default", cfg.Default)
	viper.Set("output", cfg.Output)

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Init initializes the config file with default values
func Init() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	return Save(&defaultConfig)
}

// GetConfigDir returns the config directory path
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".aami"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// Set sets a config value
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "server":
		cfg.Server = value
	case "default-namespace", "default.namespace":
		cfg.Default.Namespace = value
	case "output", "default.output":
		cfg.Default.Output = value
	case "no-headers", "output.no-headers":
		cfg.Output.NoHeaders = value == "true"
	case "color", "output.color":
		cfg.Output.Color = value == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

// Get retrieves a config value
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "server":
		return cfg.Server, nil
	case "default-namespace", "default.namespace":
		return cfg.Default.Namespace, nil
	case "output", "default.output":
		return cfg.Default.Output, nil
	case "no-headers", "output.no-headers":
		if cfg.Output.NoHeaders {
			return "true", nil
		}
		return "false", nil
	case "color", "output.color":
		if cfg.Output.Color {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
