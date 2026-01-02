package config

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath is the default path for the AAMI configuration file
const DefaultConfigPath = "/etc/aami/config.yaml"

// Load loads the configuration from the specified path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables: ${VAR_NAME}
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

// Save saves the configuration to the specified path
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// expandEnvVars expands environment variables in the format ${VAR_NAME}
func expandEnvVars(content string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[2 : len(match)-1]
		return os.Getenv(varName)
	})
}

// setDefaults sets default values for unspecified configuration fields
func setDefaults(cfg *Config) {
	if cfg.SSH.MaxParallel == 0 {
		cfg.SSH.MaxParallel = 50
	}
	if cfg.SSH.ConnectTimeout == 0 {
		cfg.SSH.ConnectTimeout = 10
	}
	if cfg.SSH.CommandTimeout == 0 {
		cfg.SSH.CommandTimeout = 300
	}
	if cfg.SSH.Retry.MaxAttempts == 0 {
		cfg.SSH.Retry.MaxAttempts = 3
	}
	if cfg.SSH.Retry.BackoffBase == 0 {
		cfg.SSH.Retry.BackoffBase = 2
	}
	if cfg.SSH.Retry.BackoffMax == 0 {
		cfg.SSH.Retry.BackoffMax = 30
	}
	if cfg.Prometheus.Retention == "" {
		cfg.Prometheus.Retention = "15d"
	}
	if cfg.Prometheus.StoragePath == "" {
		cfg.Prometheus.StoragePath = "/var/lib/aami/prometheus"
	}
	if cfg.Prometheus.Port == 0 {
		cfg.Prometheus.Port = 9090
	}
	if cfg.Grafana.Port == 0 {
		cfg.Grafana.Port = 3000
	}
}

// NewDefault creates a new configuration with default values
func NewDefault() *Config {
	cfg := &Config{
		Cluster: ClusterConfig{
			Name: "my-gpu-cluster",
		},
		Nodes: []NodeConfig{},
		Alerts: AlertsConfig{
			Presets: []string{"gpu-production"},
		},
		Notifications: NotificationsConfig{
			Slack: &SlackConfig{Enabled: false},
		},
	}
	setDefaults(cfg)
	return cfg
}
