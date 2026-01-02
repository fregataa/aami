package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ShardOperations provides operations on individual shards.
type ShardOperations struct {
	shard      ShardConfig
	httpClient *http.Client
}

// NewShardOperations creates a new shard operations handler.
func NewShardOperations(shard ShardConfig) *ShardOperations {
	return &ShardOperations{
		shard: shard,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start starts the shard Prometheus service.
func (s *ShardOperations) Start(ctx context.Context) error {
	serviceName := fmt.Sprintf("aami-prometheus-%s", s.shard.Name)
	cmd := exec.CommandContext(ctx, "systemctl", "start", serviceName)
	return cmd.Run()
}

// Stop stops the shard Prometheus service.
func (s *ShardOperations) Stop(ctx context.Context) error {
	serviceName := fmt.Sprintf("aami-prometheus-%s", s.shard.Name)
	cmd := exec.CommandContext(ctx, "systemctl", "stop", serviceName)
	return cmd.Run()
}

// Restart restarts the shard Prometheus service.
func (s *ShardOperations) Restart(ctx context.Context) error {
	serviceName := fmt.Sprintf("aami-prometheus-%s", s.shard.Name)
	cmd := exec.CommandContext(ctx, "systemctl", "restart", serviceName)
	return cmd.Run()
}

// Reload reloads the shard configuration without restart.
func (s *ShardOperations) Reload(ctx context.Context) error {
	url := fmt.Sprintf("http://localhost:%d/-/reload", s.shard.Prometheus.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reload failed: %s", string(body))
	}

	return nil
}

// GetTargets returns the active targets for this shard.
func (s *ShardOperations) GetTargets(ctx context.Context) ([]TargetInfo, error) {
	url := fmt.Sprintf("http://localhost:%d/api/v1/targets", s.shard.Prometheus.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				Labels       map[string]string `json:"labels"`
				ScrapeURL    string            `json:"scrapeUrl"`
				Health       string            `json:"health"`
				LastScrape   time.Time         `json:"lastScrape"`
				LastError    string            `json:"lastError"`
				ScrapePool   string            `json:"scrapePool"`
			} `json:"activeTargets"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var targets []TargetInfo
	for _, t := range result.Data.ActiveTargets {
		targets = append(targets, TargetInfo{
			Labels:     t.Labels,
			ScrapeURL:  t.ScrapeURL,
			Health:     t.Health,
			LastScrape: t.LastScrape,
			LastError:  t.LastError,
			ScrapePool: t.ScrapePool,
		})
	}

	return targets, nil
}

// TargetInfo represents information about a scrape target.
type TargetInfo struct {
	Labels     map[string]string `json:"labels"`
	ScrapeURL  string            `json:"scrape_url"`
	Health     string            `json:"health"`
	LastScrape time.Time         `json:"last_scrape"`
	LastError  string            `json:"last_error,omitempty"`
	ScrapePool string            `json:"scrape_pool"`
}

// GetRuntimeInfo returns runtime information about the shard.
func (s *ShardOperations) GetRuntimeInfo(ctx context.Context) (*RuntimeInfo, error) {
	url := fmt.Sprintf("http://localhost:%d/api/v1/status/runtimeinfo", s.shard.Prometheus.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status string      `json:"status"`
		Data   RuntimeInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// RuntimeInfo contains Prometheus runtime information.
type RuntimeInfo struct {
	StartTime           time.Time `json:"startTime"`
	CWD                 string    `json:"CWD"`
	ReloadConfigSuccess bool      `json:"reloadConfigSuccess"`
	LastConfigTime      time.Time `json:"lastConfigTime"`
	CorruptionCount     int       `json:"corruptionCount"`
	GoroutineCount      int       `json:"goroutineCount"`
	StorageRetention    string    `json:"storageRetention"`
}

// GetTSDBStats returns TSDB statistics for the shard.
func (s *ShardOperations) GetTSDBStats(ctx context.Context) (*TSDBStats, error) {
	url := fmt.Sprintf("http://localhost:%d/api/v1/status/tsdb", s.shard.Prometheus.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status string    `json:"status"`
		Data   TSDBStats `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// TSDBStats contains TSDB statistics.
type TSDBStats struct {
	HeadStats struct {
		NumSeries     int64 `json:"numSeries"`
		NumLabelPairs int64 `json:"numLabelPairs"`
		ChunkCount    int64 `json:"chunkCount"`
		MinTime       int64 `json:"minTime"`
		MaxTime       int64 `json:"maxTime"`
	} `json:"headStats"`
	SeriesCountByMetricName []struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	} `json:"seriesCountByMetricName"`
}

// Snapshot creates a TSDB snapshot of the shard.
func (s *ShardOperations) Snapshot(ctx context.Context) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/api/v1/admin/tsdb/snapshot", s.shard.Prometheus.Port)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("snapshot failed: %s", string(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	snapshotPath := filepath.Join(s.shard.Prometheus.StoragePath, "snapshots", result.Data.Name)
	return snapshotPath, nil
}

// CleanTombstones triggers a cleanup of deleted data.
func (s *ShardOperations) CleanTombstones(ctx context.Context) error {
	url := fmt.Sprintf("http://localhost:%d/api/v1/admin/tsdb/clean_tombstones", s.shard.Prometheus.Port)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clean tombstones failed: %s", string(body))
	}

	return nil
}

// ShardValidator validates shard configuration.
type ShardValidator struct{}

// NewShardValidator creates a new validator.
func NewShardValidator() *ShardValidator {
	return &ShardValidator{}
}

// Validate validates a shard configuration.
func (v *ShardValidator) Validate(shard ShardConfig) []string {
	var errors []string

	if shard.Name == "" {
		errors = append(errors, "shard name is required")
	}

	if len(shard.Nodes) == 0 {
		errors = append(errors, fmt.Sprintf("shard %s has no nodes assigned", shard.Name))
	}

	if shard.Prometheus.Port <= 0 || shard.Prometheus.Port > 65535 {
		errors = append(errors, fmt.Sprintf("shard %s has invalid port: %d", shard.Name, shard.Prometheus.Port))
	}

	if shard.Prometheus.StoragePath == "" {
		errors = append(errors, fmt.Sprintf("shard %s has no storage path", shard.Name))
	}

	return errors
}

// ValidateAll validates all shards and checks for conflicts.
func (v *ShardValidator) ValidateAll(shards []ShardConfig) []string {
	var errors []string

	ports := make(map[int]string)
	nodes := make(map[string]string)
	names := make(map[string]bool)

	for _, shard := range shards {
		// Validate individual shard
		errors = append(errors, v.Validate(shard)...)

		// Check for duplicate names
		if names[shard.Name] {
			errors = append(errors, fmt.Sprintf("duplicate shard name: %s", shard.Name))
		}
		names[shard.Name] = true

		// Check for port conflicts
		if existing, ok := ports[shard.Prometheus.Port]; ok {
			errors = append(errors, fmt.Sprintf("port %d used by both %s and %s",
				shard.Prometheus.Port, existing, shard.Name))
		}
		ports[shard.Prometheus.Port] = shard.Name

		// Check for node assignment conflicts
		for _, node := range shard.Nodes {
			if existing, ok := nodes[node]; ok {
				errors = append(errors, fmt.Sprintf("node %s assigned to both %s and %s",
					node, existing, shard.Name))
			}
			nodes[node] = shard.Name
		}
	}

	return errors
}

// ShardRebalancer handles rebalancing nodes across shards.
type ShardRebalancer struct {
	shards []ShardConfig
}

// NewShardRebalancer creates a new rebalancer.
func NewShardRebalancer(shards []ShardConfig) *ShardRebalancer {
	return &ShardRebalancer{shards: shards}
}

// GetImbalance returns the imbalance ratio (0 = perfect balance, 1 = max imbalance).
func (r *ShardRebalancer) GetImbalance() float64 {
	if len(r.shards) == 0 {
		return 0
	}

	var total, min, max int
	for i, shard := range r.shards {
		count := len(shard.Nodes)
		total += count
		if i == 0 || count < min {
			min = count
		}
		if i == 0 || count > max {
			max = count
		}
	}

	if max == 0 {
		return 0
	}

	return float64(max-min) / float64(max)
}

// SuggestRebalance suggests node moves to balance shards.
func (r *ShardRebalancer) SuggestRebalance() []NodeMove {
	if len(r.shards) <= 1 {
		return nil
	}

	var moves []NodeMove
	var total int
	for _, shard := range r.shards {
		total += len(shard.Nodes)
	}

	target := total / len(r.shards)

	// Find shards with excess and deficit
	for i := range r.shards {
		excess := len(r.shards[i].Nodes) - target
		if excess > 0 {
			// Find shards that need nodes
			for j := range r.shards {
				if i == j {
					continue
				}
				deficit := target - len(r.shards[j].Nodes)
				if deficit > 0 {
					moveCount := min(excess, deficit)
					for k := 0; k < moveCount && k < len(r.shards[i].Nodes); k++ {
						moves = append(moves, NodeMove{
							Node:       r.shards[i].Nodes[k],
							FromShard:  r.shards[i].Name,
							ToShard:    r.shards[j].Name,
						})
					}
					excess -= moveCount
				}
			}
		}
	}

	return moves
}

// NodeMove represents a suggested node move between shards.
type NodeMove struct {
	Node      string `json:"node"`
	FromShard string `json:"from_shard"`
	ToShard   string `json:"to_shard"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GeneratePrometheusRules generates recording rules for federation.
func GeneratePrometheusRules(outputPath string) error {
	rules := `# Recording rules for federation
# Generated by AAMI

groups:
  - name: federation_aggregations
    interval: 60s
    rules:
      # Aggregate GPU utilization by shard
      - record: shard:DCGM_FI_DEV_GPU_UTIL:avg
        expr: avg by (shard) (DCGM_FI_DEV_GPU_UTIL)

      # Aggregate memory usage by shard
      - record: shard:DCGM_FI_DEV_FB_USED:sum
        expr: sum by (shard) (DCGM_FI_DEV_FB_USED)

      # Count GPUs per shard
      - record: shard:gpu:count
        expr: count by (shard) (DCGM_FI_DEV_GPU_UTIL)

      # Aggregate temperature max by shard
      - record: shard:DCGM_FI_DEV_GPU_TEMP:max
        expr: max by (shard) (DCGM_FI_DEV_GPU_TEMP)

      # Power consumption per shard
      - record: shard:DCGM_FI_DEV_POWER_USAGE:sum
        expr: sum by (shard) (DCGM_FI_DEV_POWER_USAGE)

      # ECC errors per shard
      - record: shard:DCGM_FI_DEV_ECC_DBE_VOL_TOTAL:sum
        expr: sum by (shard) (DCGM_FI_DEV_ECC_DBE_VOL_TOTAL)

  - name: cluster_aggregations
    interval: 60s
    rules:
      # Total GPU utilization
      - record: cluster:DCGM_FI_DEV_GPU_UTIL:avg
        expr: avg(DCGM_FI_DEV_GPU_UTIL)

      # Total memory
      - record: cluster:DCGM_FI_DEV_FB_USED:sum
        expr: sum(DCGM_FI_DEV_FB_USED)

      # Total GPU count
      - record: cluster:gpu:count
        expr: count(DCGM_FI_DEV_GPU_UTIL)

      # Max temperature in cluster
      - record: cluster:DCGM_FI_DEV_GPU_TEMP:max
        expr: max(DCGM_FI_DEV_GPU_TEMP)

      # Total power
      - record: cluster:DCGM_FI_DEV_POWER_USAGE:sum
        expr: sum(DCGM_FI_DEV_POWER_USAGE)
`

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(strings.TrimSpace(rules)), 0644)
}
