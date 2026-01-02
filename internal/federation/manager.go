package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/fregataa/aami/internal/config"
)

// Manager handles federation operations.
type Manager struct {
	config     *config.Config
	federation FederationConfig
	configDir  string
	dataDir    string
}

// NewManager creates a new federation manager.
func NewManager(cfg *config.Config, fed FederationConfig) *Manager {
	return &Manager{
		config:     cfg,
		federation: fed,
		configDir:  "/etc/aami",
		dataDir:    "/var/lib/aami",
	}
}

// CalculateShards automatically determines shard distribution based on strategy.
func (m *Manager) CalculateShards(strategy ShardingStrategy, shardCount int) []ShardConfig {
	nodeCount := len(m.config.Nodes)
	if nodeCount == 0 {
		return nil
	}

	switch strategy {
	case ShardingStrategyRack:
		return m.calculateShardsByRack()
	case ShardingStrategyCount, ShardingStrategyAuto:
		return m.calculateShardsByCount(shardCount)
	default:
		return m.calculateShardsByCount(shardCount)
	}
}

// calculateShardsByCount distributes nodes evenly across shards.
func (m *Manager) calculateShardsByCount(shardCount int) []ShardConfig {
	nodeCount := len(m.config.Nodes)
	if shardCount <= 0 {
		shardCount = m.recommendShardCount(nodeCount)
	}

	nodesPerShard := nodeCount / shardCount
	remainder := nodeCount % shardCount

	shards := make([]ShardConfig, shardCount)
	nodeIndex := 0

	for i := 0; i < shardCount; i++ {
		shards[i] = ShardConfig{
			Name: fmt.Sprintf("shard-%d", i+1),
		}
		shards[i].Prometheus.Port = 9091 + i
		shards[i].Prometheus.StoragePath = fmt.Sprintf("%s/prometheus-shard-%d", m.dataDir, i+1)
		shards[i].Prometheus.Retention = "7d"

		// Assign nodes to shard
		count := nodesPerShard
		if i < remainder {
			count++ // Distribute remainder nodes to first shards
		}

		for j := 0; j < count && nodeIndex < nodeCount; j++ {
			shards[i].Nodes = append(shards[i].Nodes, m.config.Nodes[nodeIndex].Name)
			nodeIndex++
		}
	}

	return shards
}

// calculateShardsByRack distributes nodes by rack label.
func (m *Manager) calculateShardsByRack() []ShardConfig {
	// Group nodes by rack
	rackNodes := make(map[string][]string)

	for _, node := range m.config.Nodes {
		rack := "default"
		if r, ok := node.Labels["rack"]; ok {
			rack = r
		}
		rackNodes[rack] = append(rackNodes[rack], node.Name)
	}

	// Create shard per rack
	var shards []ShardConfig
	i := 0
	for rack, nodes := range rackNodes {
		shard := ShardConfig{
			Name:  fmt.Sprintf("shard-%s", rack),
			Nodes: nodes,
			Racks: []string{rack},
		}
		shard.Prometheus.Port = 9091 + i
		shard.Prometheus.StoragePath = fmt.Sprintf("%s/prometheus-shard-%s", m.dataDir, rack)
		shard.Prometheus.Retention = "7d"
		shards = append(shards, shard)
		i++
	}

	// Sort by name for consistent ordering
	sort.Slice(shards, func(i, j int) bool {
		return shards[i].Name < shards[j].Name
	})

	return shards
}

// recommendShardCount returns recommended shard count based on node count.
func (m *Manager) recommendShardCount(nodeCount int) int {
	switch {
	case nodeCount < 100:
		return 1
	case nodeCount < 300:
		return 2
	case nodeCount < 500:
		return 3
	case nodeCount < 1000:
		return 5
	default:
		return nodeCount / 200 // ~200 nodes per shard
	}
}

// Deploy deploys federation configuration.
func (m *Manager) Deploy(ctx context.Context) error {
	// 1. Create directories
	if err := m.createDirectories(); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	// 2. Deploy each shard
	for _, shard := range m.federation.Shards {
		if err := m.deployShard(ctx, shard); err != nil {
			return fmt.Errorf("deploy shard %s: %w", shard.Name, err)
		}
	}

	// 3. Deploy central Prometheus
	if err := m.deployCentral(ctx); err != nil {
		return fmt.Errorf("deploy central: %w", err)
	}

	return nil
}

func (m *Manager) createDirectories() error {
	dirs := []string{
		filepath.Join(m.configDir, "federation"),
		filepath.Join(m.dataDir, "targets"),
	}

	for _, shard := range m.federation.Shards {
		dirs = append(dirs, shard.Prometheus.StoragePath)
	}

	if m.federation.Central.StoragePath != "" {
		dirs = append(dirs, m.federation.Central.StoragePath)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) deployShard(ctx context.Context, shard ShardConfig) error {
	// 1. Generate Prometheus config for shard
	configPath := filepath.Join(m.configDir, "federation", fmt.Sprintf("prometheus-%s.yaml", shard.Name))
	if err := m.generateShardConfig(shard, configPath); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	// 2. Generate targets file for shard
	targetsPath := filepath.Join(m.dataDir, "targets", fmt.Sprintf("%s-nodes.json", shard.Name))
	if err := m.generateShardTargets(shard, targetsPath); err != nil {
		return fmt.Errorf("generate targets: %w", err)
	}

	// 3. Create systemd service
	servicePath := fmt.Sprintf("/etc/systemd/system/aami-prometheus-%s.service", shard.Name)
	if err := m.createShardService(shard, servicePath); err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	return nil
}

const shardConfigTemplate = `# Prometheus configuration for shard: {{ .Name }}
# Generated by AAMI - Do not edit manually

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    shard: "{{ .Name }}"
    cluster: "{{ .ClusterName }}"

scrape_configs:
  - job_name: 'node-exporter'
    file_sd_configs:
      - files:
          - '{{ .TargetsDir }}/{{ .Name }}-nodes.json'
        refresh_interval: 30s

  - job_name: 'dcgm-exporter'
    file_sd_configs:
      - files:
          - '{{ .TargetsDir }}/{{ .Name }}-dcgm.json'
        refresh_interval: 30s

storage:
  tsdb:
    path: {{ .StoragePath }}
    retention.time: {{ .Retention }}
`

type shardTemplateData struct {
	Name        string
	ClusterName string
	TargetsDir  string
	StoragePath string
	Retention   string
}

func (m *Manager) generateShardConfig(shard ShardConfig, outputPath string) error {
	tmpl, err := template.New("shard").Parse(shardConfigTemplate)
	if err != nil {
		return err
	}

	data := shardTemplateData{
		Name:        shard.Name,
		ClusterName: m.config.Cluster.Name,
		TargetsDir:  filepath.Join(m.dataDir, "targets"),
		StoragePath: shard.Prometheus.StoragePath,
		Retention:   shard.Prometheus.Retention,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func (m *Manager) generateShardTargets(shard ShardConfig, outputPath string) error {
	// Build targets list
	type target struct {
		Targets []string          `json:"targets"`
		Labels  map[string]string `json:"labels"`
	}

	var targets []target

	// Create node-to-IP mapping
	nodeIPs := make(map[string]string)
	for _, node := range m.config.Nodes {
		nodeIPs[node.Name] = node.IP
	}

	for _, nodeName := range shard.Nodes {
		ip, ok := nodeIPs[nodeName]
		if !ok {
			continue
		}

		targets = append(targets, target{
			Targets: []string{fmt.Sprintf("%s:9100", ip)},
			Labels: map[string]string{
				"node":  nodeName,
				"shard": shard.Name,
			},
		})
	}

	data, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

const shardServiceTemplate = `[Unit]
Description=AAMI Prometheus Shard - {{ .Name }}
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/bin/prometheus \
    --config.file={{ .ConfigPath }} \
    --storage.tsdb.path={{ .StoragePath }} \
    --storage.tsdb.retention.time={{ .Retention }} \
    --web.listen-address=:{{ .Port }} \
    --web.enable-lifecycle \
    --web.enable-admin-api

SyslogIdentifier=prometheus-{{ .Name }}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`

type serviceTemplateData struct {
	Name        string
	ConfigPath  string
	StoragePath string
	Retention   string
	Port        int
}

func (m *Manager) createShardService(shard ShardConfig, servicePath string) error {
	tmpl, err := template.New("service").Parse(shardServiceTemplate)
	if err != nil {
		return err
	}

	data := serviceTemplateData{
		Name:        shard.Name,
		ConfigPath:  filepath.Join(m.configDir, "federation", fmt.Sprintf("prometheus-%s.yaml", shard.Name)),
		StoragePath: shard.Prometheus.StoragePath,
		Retention:   shard.Prometheus.Retention,
		Port:        shard.Prometheus.Port,
	}

	f, err := os.Create(servicePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

const centralConfigTemplate = `# Prometheus Central - Federation Aggregator
# Generated by AAMI - Do not edit manually

global:
  scrape_interval: 60s
  evaluation_interval: 60s
  external_labels:
    role: "central"
    cluster: "{{ .ClusterName }}"

scrape_configs:
  - job_name: 'federation'
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{__name__=~"DCGM.*"}'
        - '{__name__=~"node.*"}'
        - '{__name__=~"up"}'
        - '{job=~".+"}'
    static_configs:
{{ range .Shards }}
      - targets: ['localhost:{{ .Port }}']
        labels:
          shard: '{{ .Name }}'
{{ end }}

rule_files:
  - '/etc/aami/rules/*.yaml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

storage:
  tsdb:
    path: {{ .StoragePath }}
    retention.time: {{ .RetentionRaw }}
`

type centralTemplateData struct {
	ClusterName  string
	StoragePath  string
	RetentionRaw string
	Shards       []struct {
		Name string
		Port int
	}
}

func (m *Manager) deployCentral(ctx context.Context) error {
	configPath := filepath.Join(m.configDir, "federation", "prometheus-central.yaml")

	tmpl, err := template.New("central").Parse(centralConfigTemplate)
	if err != nil {
		return err
	}

	data := centralTemplateData{
		ClusterName:  m.config.Cluster.Name,
		StoragePath:  m.federation.Central.StoragePath,
		RetentionRaw: m.federation.Central.RetentionRaw,
	}

	for _, shard := range m.federation.Shards {
		data.Shards = append(data.Shards, struct {
			Name string
			Port int
		}{
			Name: shard.Name,
			Port: shard.Prometheus.Port,
		})
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	// Create central service
	return m.createCentralService()
}

func (m *Manager) createCentralService() error {
	servicePath := "/etc/systemd/system/aami-prometheus-central.service"

	content := fmt.Sprintf(`[Unit]
Description=AAMI Prometheus Central - Federation Aggregator
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/bin/prometheus \
    --config.file=%s/federation/prometheus-central.yaml \
    --storage.tsdb.path=%s \
    --storage.tsdb.retention.time=%s \
    --web.listen-address=:%d \
    --web.enable-lifecycle \
    --web.enable-admin-api

SyslogIdentifier=prometheus-central
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, m.configDir, m.federation.Central.StoragePath, m.federation.Central.RetentionRaw, m.federation.Central.Port)

	return os.WriteFile(servicePath, []byte(content), 0644)
}

// GetStatus returns the current federation status.
func (m *Manager) GetStatus(ctx context.Context) (*FederationStatus, error) {
	status := &FederationStatus{
		Enabled:    m.federation.Enabled,
		Type:       string(m.federation.Type),
		ShardCount: len(m.federation.Shards),
	}

	for _, shard := range m.federation.Shards {
		shardStatus := m.checkShardStatus(ctx, shard)
		status.Shards = append(status.Shards, shardStatus)
		status.TotalNodes += shardStatus.NodeCount
		if shardStatus.Healthy {
			status.HealthyCount++
		}
	}

	// Check central
	status.Central = m.checkCentralStatus(ctx)

	return status, nil
}

func (m *Manager) checkShardStatus(ctx context.Context, shard ShardConfig) ShardStatus {
	status := ShardStatus{
		Name:      shard.Name,
		Endpoint:  fmt.Sprintf("localhost:%d", shard.Prometheus.Port),
		NodeCount: len(shard.Nodes),
	}

	// HTTP health check
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/-/healthy", shard.Prometheus.Port)

	resp, err := client.Get(url)
	if err != nil {
		status.Healthy = false
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	status.Healthy = resp.StatusCode == 200

	// Get metrics count
	metricsURL := fmt.Sprintf("http://localhost:%d/api/v1/status/tsdb", shard.Prometheus.Port)
	metricsResp, err := client.Get(metricsURL)
	if err == nil {
		defer metricsResp.Body.Close()
		var result struct {
			Data struct {
				HeadStats struct {
					NumSeries int64 `json:"numSeries"`
				} `json:"headStats"`
			} `json:"data"`
		}
		if json.NewDecoder(metricsResp.Body).Decode(&result) == nil {
			status.MetricCount = result.Data.HeadStats.NumSeries
		}
	}

	return status
}

func (m *Manager) checkCentralStatus(ctx context.Context) CentralStatus {
	port := m.federation.Central.Port
	if port == 0 {
		port = 9090
	}

	status := CentralStatus{
		Endpoint: fmt.Sprintf("localhost:%d", port),
	}

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/-/healthy", port)

	resp, err := client.Get(url)
	if err != nil {
		status.Healthy = false
		return status
	}
	defer resp.Body.Close()

	status.Healthy = resp.StatusCode == 200

	return status
}

// Disable disables federation and consolidates to single Prometheus.
func (m *Manager) Disable(ctx context.Context) error {
	// Stop shard services
	for _, shard := range m.federation.Shards {
		serviceName := fmt.Sprintf("aami-prometheus-%s", shard.Name)
		// In production, would use systemctl to stop and disable
		_ = serviceName
	}

	// Stop central service
	// In production, would use systemctl

	// Clean up config files
	fedDir := filepath.Join(m.configDir, "federation")
	if err := os.RemoveAll(fedDir); err != nil {
		return fmt.Errorf("cleanup federation dir: %w", err)
	}

	return nil
}

// GetConfig returns the current federation configuration.
func (m *Manager) GetConfig() FederationConfig {
	return m.federation
}

// SetShards sets the shard configuration.
func (m *Manager) SetShards(shards []ShardConfig) {
	m.federation.Shards = shards
}
