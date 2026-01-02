// Package health provides GPU health score calculation and monitoring.
package health

import "time"

// ScoreWeights defines the weights for health score calculation.
type ScoreWeights struct {
	Temperature float64 `yaml:"temperature"` // Default: 0.20
	ECCErrors   float64 `yaml:"ecc_errors"`  // Default: 0.25
	XidErrors   float64 `yaml:"xid_errors"`  // Default: 0.25
	NVLink      float64 `yaml:"nvlink"`      // Default: 0.15
	Uptime      float64 `yaml:"uptime"`      // Default: 0.15
}

// DefaultWeights returns the default scoring weights.
func DefaultWeights() ScoreWeights {
	return ScoreWeights{
		Temperature: 0.20,
		ECCErrors:   0.25,
		XidErrors:   0.25,
		NVLink:      0.15,
		Uptime:      0.15,
	}
}

// ComponentScore represents the score for a single component.
type ComponentScore struct {
	Name       string  `json:"name"`
	Score      float64 `json:"score"`       // 0-100
	Weight     float64 `json:"weight"`      // 0-1
	Weighted   float64 `json:"weighted"`    // Score * Weight
	RawValue   float64 `json:"raw_value"`   // Original metric value
	Threshold  float64 `json:"threshold"`   // Threshold for scoring
	Status     string  `json:"status"`      // good, warning, critical
	Message    string  `json:"message"`     // Human readable status
}

// GPUHealth represents the health status of a single GPU.
type GPUHealth struct {
	Index        int              `json:"index"`
	UUID         string           `json:"uuid"`
	Name         string           `json:"name"`
	OverallScore float64          `json:"overall_score"` // 0-100
	Status       string           `json:"status"`        // healthy, warning, critical
	Components   []ComponentScore `json:"components"`
	CollectedAt  time.Time        `json:"collected_at"`
}

// NodeHealth represents the health status of all GPUs on a node.
type NodeHealth struct {
	NodeName     string      `json:"node_name"`
	NodeIP       string      `json:"node_ip"`
	GPUs         []GPUHealth `json:"gpus"`
	OverallScore float64     `json:"overall_score"` // Average of all GPU scores
	Status       string      `json:"status"`        // healthy, warning, critical
	HealthyGPUs  int         `json:"healthy_gpus"`
	WarningGPUs  int         `json:"warning_gpus"`
	CriticalGPUs int         `json:"critical_gpus"`
	CollectedAt  time.Time   `json:"collected_at"`
}

// ClusterHealth represents the health status of the entire cluster.
type ClusterHealth struct {
	Nodes        []NodeHealth `json:"nodes"`
	OverallScore float64      `json:"overall_score"`
	Status       string       `json:"status"`
	TotalGPUs    int          `json:"total_gpus"`
	HealthyGPUs  int          `json:"healthy_gpus"`
	WarningGPUs  int          `json:"warning_gpus"`
	CriticalGPUs int          `json:"critical_gpus"`
	CollectedAt  time.Time    `json:"collected_at"`
}

// HealthStatus constants
const (
	StatusHealthy  = "healthy"
	StatusWarning  = "warning"
	StatusCritical = "critical"
	StatusUnknown  = "unknown"
)

// GetStatusFromScore returns the status based on the score.
func GetStatusFromScore(score float64) string {
	switch {
	case score >= 80:
		return StatusHealthy
	case score >= 50:
		return StatusWarning
	default:
		return StatusCritical
	}
}

// TemperatureThresholds defines temperature thresholds for scoring.
type TemperatureThresholds struct {
	Optimal  float64 `yaml:"optimal"`  // Below this = 100 score
	Warning  float64 `yaml:"warning"`  // Above this = warning
	Critical float64 `yaml:"critical"` // Above this = critical
}

// DefaultTemperatureThresholds returns default temperature thresholds.
func DefaultTemperatureThresholds() TemperatureThresholds {
	return TemperatureThresholds{
		Optimal:  60.0,
		Warning:  75.0,
		Critical: 85.0,
	}
}

// ECCThresholds defines ECC error thresholds for scoring.
type ECCThresholds struct {
	SingleBitWarning  int `yaml:"single_bit_warning"`  // SBE count for warning
	SingleBitCritical int `yaml:"single_bit_critical"` // SBE count for critical
	DoubleBitWarning  int `yaml:"double_bit_warning"`  // DBE count for warning
	DoubleBitCritical int `yaml:"double_bit_critical"` // DBE count for critical
}

// DefaultECCThresholds returns default ECC thresholds.
func DefaultECCThresholds() ECCThresholds {
	return ECCThresholds{
		SingleBitWarning:  100,
		SingleBitCritical: 1000,
		DoubleBitWarning:  1,
		DoubleBitCritical: 10,
	}
}

// XidThresholds defines Xid error thresholds for scoring.
type XidThresholds struct {
	CountWarning  int `yaml:"count_warning"`  // Xid count for warning (24h)
	CountCritical int `yaml:"count_critical"` // Xid count for critical (24h)
}

// DefaultXidThresholds returns default Xid thresholds.
func DefaultXidThresholds() XidThresholds {
	return XidThresholds{
		CountWarning:  1,
		CountCritical: 5,
	}
}

// GPUMetrics contains raw metrics collected from Prometheus.
type GPUMetrics struct {
	GPU         string  `json:"gpu"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Instance    string  `json:"instance"`
	Temperature float64 `json:"temperature"`
	ECCSingleBit int64   `json:"ecc_single_bit"`
	ECCDoubleBit int64   `json:"ecc_double_bit"`
	XidErrors    int64   `json:"xid_errors"`
	NVLinkActive int     `json:"nvlink_active"`
	NVLinkTotal  int     `json:"nvlink_total"`
	Uptime       float64 `json:"uptime"` // seconds
}

// NodeMetrics contains metrics for all GPUs on a node.
type NodeMetrics struct {
	NodeName string       `json:"node_name"`
	NodeIP   string       `json:"node_ip"`
	GPUs     []GPUMetrics `json:"gpus"`
}
