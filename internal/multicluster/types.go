package multicluster

import "time"

// ClusterConfig defines configuration for a remote AAMI cluster.
type ClusterConfig struct {
	Name      string `yaml:"name" json:"name"`
	Endpoint  string `yaml:"endpoint" json:"endpoint"`
	APIKey    string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	TLSCert   string `yaml:"tls_cert,omitempty" json:"tls_cert,omitempty"`
	TLSKey    string `yaml:"tls_key,omitempty" json:"tls_key,omitempty"`
	TLSCACert string `yaml:"tls_ca_cert,omitempty" json:"tls_ca_cert,omitempty"`
	SkipTLS   bool   `yaml:"skip_tls_verify,omitempty" json:"skip_tls_verify,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ClusterStatus represents the current status of a cluster.
type ClusterStatus struct {
	Name         string    `json:"name"`
	Endpoint     string    `json:"endpoint"`
	Nodes        int       `json:"nodes"`
	HealthyNodes int       `json:"healthy_nodes"`
	TotalGPUs    int       `json:"total_gpus"`
	HealthyGPUs  int       `json:"healthy_gpus"`
	HealthScore  float64   `json:"health_score"`
	Connected    bool      `json:"connected"`
	LastSync     time.Time `json:"last_sync"`
	AlertsActive int       `json:"alerts_active"`
	Version      string    `json:"version"`
	Error        string    `json:"error,omitempty"`
}

// GlobalAlert represents an alert from any cluster.
type GlobalAlert struct {
	Cluster     string            `json:"cluster"`
	AlertName   string            `json:"alert_name"`
	Severity    string            `json:"severity"`
	Node        string            `json:"node"`
	GPU         int               `json:"gpu,omitempty"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
	FiredAt     time.Time         `json:"fired_at"`
	State       AlertState        `json:"state"`
}

// AlertState represents the state of an alert.
type AlertState string

const (
	AlertStateFiring   AlertState = "firing"
	AlertStatePending  AlertState = "pending"
	AlertStateResolved AlertState = "resolved"
)

// AggregatedMetrics provides a unified view across all clusters.
type AggregatedMetrics struct {
	Timestamp        time.Time                  `json:"timestamp"`
	TotalNodes       int                        `json:"total_nodes"`
	HealthyNodes     int                        `json:"healthy_nodes"`
	TotalGPUs        int                        `json:"total_gpus"`
	HealthyGPUs      int                        `json:"healthy_gpus"`
	AverageHealth    float64                    `json:"average_health"`
	ActiveAlerts     int                        `json:"active_alerts"`
	CriticalAlerts   int                        `json:"critical_alerts"`
	ClusterCount     int                        `json:"cluster_count"`
	ConnectedCount   int                        `json:"connected_count"`
	ClusterBreakdown map[string]ClusterMetrics  `json:"cluster_breakdown"`
}

// ClusterMetrics contains metrics for a single cluster.
type ClusterMetrics struct {
	Nodes        int     `json:"nodes"`
	HealthyNodes int     `json:"healthy_nodes"`
	GPUs         int     `json:"gpus"`
	HealthyGPUs  int     `json:"healthy_gpus"`
	HealthScore  float64 `json:"health_score"`
	AlertCount   int     `json:"alert_count"`
	Connected    bool    `json:"connected"`
}

// MultiClusterConfig holds the multi-cluster configuration.
type MultiClusterConfig struct {
	Enabled       bool            `yaml:"enabled" json:"enabled"`
	Clusters      []ClusterConfig `yaml:"clusters" json:"clusters"`
	SyncInterval  string          `yaml:"sync_interval" json:"sync_interval"`   // e.g., "30s"
	AlertForward  bool            `yaml:"alert_forward" json:"alert_forward"`   // Forward alerts to central
	MetricsMerge  bool            `yaml:"metrics_merge" json:"metrics_merge"`   // Merge metrics from all clusters
}

// ClusterHealth provides detailed health information.
type ClusterHealth struct {
	Status       ClusterStatus   `json:"status"`
	TopIssues    []ClusterIssue  `json:"top_issues"`
	RecentAlerts []GlobalAlert   `json:"recent_alerts"`
	Trends       ClusterTrends   `json:"trends"`
}

// ClusterIssue represents a notable issue in a cluster.
type ClusterIssue struct {
	Severity    string    `json:"severity"`
	Category    string    `json:"category"`  // "gpu", "node", "network", etc.
	Description string    `json:"description"`
	AffectedNodes []string `json:"affected_nodes,omitempty"`
	Since       time.Time `json:"since"`
}

// ClusterTrends contains trend information for the cluster.
type ClusterTrends struct {
	HealthTrend  string  `json:"health_trend"`  // "improving", "stable", "degrading"
	AlertTrend   string  `json:"alert_trend"`   // "increasing", "stable", "decreasing"
	NodeChange   int     `json:"node_change"`   // Change in node count
	GPUChange    int     `json:"gpu_change"`    // Change in GPU count
}

// ClusterEvent represents an event in the cluster.
type ClusterEvent struct {
	Cluster   string    `json:"cluster"`
	Type      string    `json:"type"`      // "alert", "node_down", "gpu_error", etc.
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ClusterSummary provides a brief summary for display.
type ClusterSummary struct {
	Name        string  `json:"name"`
	Nodes       int     `json:"nodes"`
	GPUs        int     `json:"gpus"`
	HealthScore float64 `json:"health_score"`
	Alerts      int     `json:"alerts"`
	Status      string  `json:"status"`  // "healthy", "warning", "critical", "offline"
}

// GetStatusString returns a human-readable status string.
func (s *ClusterStatus) GetStatusString() string {
	if !s.Connected {
		return "offline"
	}
	if s.HealthScore >= 90 {
		return "healthy"
	}
	if s.HealthScore >= 70 {
		return "warning"
	}
	return "critical"
}

// ToSummary converts ClusterStatus to ClusterSummary.
func (s *ClusterStatus) ToSummary() ClusterSummary {
	return ClusterSummary{
		Name:        s.Name,
		Nodes:       s.Nodes,
		GPUs:        s.TotalGPUs,
		HealthScore: s.HealthScore,
		Alerts:      s.AlertsActive,
		Status:      s.GetStatusString(),
	}
}
