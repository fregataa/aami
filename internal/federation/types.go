package federation

import "time"

// FederationType defines the federation backend type.
type FederationType string

const (
	// FederationTypePrometheus uses native Prometheus federation.
	FederationTypePrometheus FederationType = "prometheus"
	// FederationTypeThanos uses Thanos for long-term storage and global view.
	FederationTypeThanos FederationType = "thanos"
)

// FederationConfig holds the federation configuration.
type FederationConfig struct {
	Enabled     bool           `yaml:"enabled"`
	Type        FederationType `yaml:"type"`
	Shards      []ShardConfig  `yaml:"shards"`
	CentralNode string         `yaml:"central_node"`
	Central     CentralConfig  `yaml:"central"`
}

// ShardConfig defines a single Prometheus shard configuration.
type ShardConfig struct {
	Name       string   `yaml:"name"`
	Nodes      []string `yaml:"nodes"`       // Node names assigned to this shard
	Racks      []string `yaml:"racks"`       // Rack identifiers (optional)
	Prometheus struct {
		Port        int    `yaml:"port"`
		StoragePath string `yaml:"storage_path"`
		Retention   string `yaml:"retention"`
	} `yaml:"prometheus"`
}

// CentralConfig defines the central Prometheus configuration.
type CentralConfig struct {
	Port                 int    `yaml:"port"`
	RetentionRaw         string `yaml:"retention_raw"`         // Short retention for raw metrics
	RetentionDownsampled string `yaml:"retention_downsampled"` // Long retention for aggregated
	FederateInterval     string `yaml:"federate_interval"`     // How often to pull from shards
	StoragePath          string `yaml:"storage_path"`
}

// ShardStatus represents the current status of a shard.
type ShardStatus struct {
	Name        string    `json:"name"`
	Endpoint    string    `json:"endpoint"`
	NodeCount   int       `json:"node_count"`
	Healthy     bool      `json:"healthy"`
	LastScrape  time.Time `json:"last_scrape"`
	MetricCount int64     `json:"metric_count"`
	Uptime      string    `json:"uptime"`
	Error       string    `json:"error,omitempty"`
}

// FederationStatus represents the overall federation status.
type FederationStatus struct {
	Enabled      bool          `json:"enabled"`
	Type         string        `json:"type"`
	ShardCount   int           `json:"shard_count"`
	TotalNodes   int           `json:"total_nodes"`
	HealthyCount int           `json:"healthy_count"`
	Shards       []ShardStatus `json:"shards"`
	Central      CentralStatus `json:"central"`
}

// CentralStatus represents the central Prometheus status.
type CentralStatus struct {
	Endpoint    string    `json:"endpoint"`
	Healthy     bool      `json:"healthy"`
	LastScrape  time.Time `json:"last_scrape"`
	MetricCount int64     `json:"metric_count"`
}

// ShardingStrategy defines how nodes are distributed across shards.
type ShardingStrategy string

const (
	// ShardingStrategyAuto automatically distributes nodes evenly.
	ShardingStrategyAuto ShardingStrategy = "auto"
	// ShardingStrategyRack distributes nodes by rack.
	ShardingStrategyRack ShardingStrategy = "rack"
	// ShardingStrategyCount distributes by fixed count per shard.
	ShardingStrategyCount ShardingStrategy = "count"
	// ShardingStrategyManual uses manually specified node lists.
	ShardingStrategyManual ShardingStrategy = "manual"
)

// DefaultFederationConfig returns default federation configuration.
func DefaultFederationConfig() FederationConfig {
	return FederationConfig{
		Enabled: false,
		Type:    FederationTypePrometheus,
		Central: CentralConfig{
			Port:                 9090,
			RetentionRaw:         "2d",
			RetentionDownsampled: "30d",
			FederateInterval:     "60s",
			StoragePath:          "/var/lib/aami/prometheus-central",
		},
	}
}

// DefaultShardConfig returns default shard configuration for a given index.
func DefaultShardConfig(index int) ShardConfig {
	cfg := ShardConfig{
		Name: "",
	}
	cfg.Prometheus.Port = 9091 + index
	cfg.Prometheus.StoragePath = ""
	cfg.Prometheus.Retention = "7d"
	return cfg
}
