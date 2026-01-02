package config

// Config represents the main AAMI configuration
type Config struct {
	Cluster       ClusterConfig       `yaml:"cluster"`
	Nodes         []NodeConfig        `yaml:"nodes"`
	SSH           SSHConfig           `yaml:"ssh"`
	Alerts        AlertsConfig        `yaml:"alerts"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Prometheus    PrometheusConfig    `yaml:"prometheus"`
	Grafana       GrafanaConfig       `yaml:"grafana"`
}

// ClusterConfig contains cluster-wide settings
type ClusterConfig struct {
	Name string `yaml:"name"`
}

// NodeConfig represents a monitored node
type NodeConfig struct {
	Name    string            `yaml:"name"`
	IP      string            `yaml:"ip"`
	SSHUser string            `yaml:"ssh_user"`
	SSHKey  string            `yaml:"ssh_key"`
	SSHPort int               `yaml:"ssh_port"`
	Labels  map[string]string `yaml:"labels"`
}

// SSHConfig contains SSH connection settings
type SSHConfig struct {
	MaxParallel    int         `yaml:"max_parallel"`    // default: 50
	ConnectTimeout int         `yaml:"connect_timeout"` // seconds, default: 10
	CommandTimeout int         `yaml:"command_timeout"` // seconds, default: 300
	Retry          RetryConfig `yaml:"retry"`
}

// RetryConfig contains retry settings for SSH connections
type RetryConfig struct {
	MaxAttempts int `yaml:"max_attempts"` // default: 3
	BackoffBase int `yaml:"backoff_base"` // seconds, default: 2
	BackoffMax  int `yaml:"backoff_max"`  // seconds, default: 30
}

// AlertsConfig contains alert settings
type AlertsConfig struct {
	Presets []string          `yaml:"presets"`
	Custom  []CustomAlertRule `yaml:"custom"`
}

// CustomAlertRule represents a custom alert rule
type CustomAlertRule struct {
	Name     string `yaml:"name"`
	Expr     string `yaml:"expr"`
	For      string `yaml:"for"`
	Severity string `yaml:"severity"`
}

// NotificationsConfig contains notification settings
type NotificationsConfig struct {
	Slack   *SlackConfig   `yaml:"slack"`
	Email   *EmailConfig   `yaml:"email"`
	Webhook *WebhookConfig `yaml:"webhook"`
}

// SlackConfig contains Slack notification settings
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	Enabled  bool     `yaml:"enabled"`
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

// WebhookConfig contains webhook notification settings
type WebhookConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// PrometheusConfig contains Prometheus settings
type PrometheusConfig struct {
	Retention   string `yaml:"retention"`     // default: "15d"
	StoragePath string `yaml:"storage_path"`  // default: "/var/lib/aami/prometheus"
	Port        int    `yaml:"port"`          // default: 9090
}

// GrafanaConfig contains Grafana settings
type GrafanaConfig struct {
	Port          int    `yaml:"port"`           // default: 3000
	AdminPassword string `yaml:"admin_password"` // supports ${ENV_VAR}
}
