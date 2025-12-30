package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete agent configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Executor ExecutorConfig `mapstructure:"executor"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig contains config-server connection settings
type ServerConfig struct {
	URL        string `mapstructure:"url"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	TLSCert    string `mapstructure:"tls_cert"`
	TLSKey     string `mapstructure:"tls_key"`
	TLSCA      string `mapstructure:"tls_ca"`
	SkipVerify bool   `mapstructure:"skip_verify"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	StateFile                 string            `mapstructure:"state_file"`
	Hostname                  string            `mapstructure:"hostname"`
	IPAddress                 string            `mapstructure:"ip_address"`
	GroupID                   string            `mapstructure:"group_id"`
	BootstrapToken            string            `mapstructure:"bootstrap_token"`
	Labels                    map[string]string `mapstructure:"labels"`
	Metadata                  map[string]string `mapstructure:"metadata"`
	HeartbeatInterval         time.Duration     `mapstructure:"heartbeat_interval"`
	CheckPollInterval         time.Duration     `mapstructure:"check_poll_interval"`
	RegistrationRetryInterval time.Duration     `mapstructure:"registration_retry_interval"`
	RegistrationMaxRetries    int               `mapstructure:"registration_max_retries"`
}

// ExecutorConfig contains script execution settings
type ExecutorConfig struct {
	WorkDir        string        `mapstructure:"work_dir"`
	ScriptDir      string        `mapstructure:"script_dir"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	MaxConcurrent  int           `mapstructure:"max_concurrent"`
	ShellPath      string        `mapstructure:"shell_path"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load reads configuration from the specified file path
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read config file
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		// Config file is optional for some use cases
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Allow environment variable overrides
	v.SetEnvPrefix("AAMI_AGENT")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Auto-detect hostname and IP if not specified
	if err := autoDetect(&cfg); err != nil {
		return nil, fmt.Errorf("failed to auto-detect settings: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.url", "http://localhost:8080")
	v.SetDefault("server.tls_enabled", false)
	v.SetDefault("server.skip_verify", false)
	v.SetDefault("server.timeout", "30s")

	// Agent defaults
	v.SetDefault("agent.state_file", "/var/lib/aami/state.json")
	v.SetDefault("agent.heartbeat_interval", "30s")
	v.SetDefault("agent.check_poll_interval", "60s")
	v.SetDefault("agent.registration_retry_interval", "10s")
	v.SetDefault("agent.registration_max_retries", 10)

	// Executor defaults
	v.SetDefault("executor.work_dir", "/var/lib/aami/work")
	v.SetDefault("executor.script_dir", "/var/lib/aami/scripts")
	v.SetDefault("executor.default_timeout", "60s")
	v.SetDefault("executor.max_concurrent", 5)
	v.SetDefault("executor.shell_path", "/bin/bash")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
}

func autoDetect(cfg *Config) error {
	// Auto-detect hostname
	if cfg.Agent.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("failed to detect hostname: %w", err)
		}
		cfg.Agent.Hostname = hostname
	}

	// Auto-detect IP address
	if cfg.Agent.IPAddress == "" {
		ip, err := getOutboundIP()
		if err != nil {
			// Not fatal, just log warning
			cfg.Agent.IPAddress = "127.0.0.1"
		} else {
			cfg.Agent.IPAddress = ip.String()
		}
	}

	return nil
}

func validate(cfg *Config) error {
	if cfg.Server.URL == "" {
		return fmt.Errorf("server.url is required")
	}

	if cfg.Agent.StateFile == "" {
		return fmt.Errorf("agent.state_file is required")
	}

	if cfg.Agent.HeartbeatInterval < time.Second {
		return fmt.Errorf("agent.heartbeat_interval must be at least 1 second")
	}

	if cfg.Agent.CheckPollInterval < time.Second {
		return fmt.Errorf("agent.check_poll_interval must be at least 1 second")
	}

	if cfg.Executor.MaxConcurrent < 1 {
		return fmt.Errorf("executor.max_concurrent must be at least 1")
	}

	return nil
}

// getOutboundIP gets the preferred outbound IP address
func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
