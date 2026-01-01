# Phase 3: Scale

## Overview

- **Duration**: On-demand (2 weeks per Epic)
- **Goal**: Large-scale environment support, external system integration
- **Prerequisites**: Phase 2 completed, demand confirmed

## Entry Criteria

Start Phase 3 when any of the following conditions are met:
- 500+ nodes with performance issues
- Slurm/scheduler integration request
- Multi-cluster management requirement
- AMD GPU support request

## New Files

```
aami/
├── internal/
│   ├── cli/
│   │   ├── federation.go       # Federation commands
│   │   ├── slurm.go            # Slurm integration commands
│   │   ├── clusters.go         # Multi-cluster commands
│   │   └── amd.go              # AMD GPU commands
│   ├── federation/             # Prometheus federation
│   │   ├── manager.go          # Federation manager
│   │   ├── shard.go            # Shard configuration
│   │   └── thanos.go           # Thanos integration (optional)
│   ├── slurm/                  # Slurm integration
│   │   ├── client.go           # Slurm API client
│   │   ├── analyzer.go         # Job-GPU correlation
│   │   └── hooks.go            # Pre/post job hooks
│   ├── multicluster/           # Multi-cluster management
│   │   ├── registry.go         # Cluster registry
│   │   ├── client.go           # Remote AAMI client
│   │   └── aggregator.go       # Metric aggregation
│   └── amd/                    # AMD GPU support
│       ├── rocm.go             # ROCm metrics
│       ├── mapping.go          # DCGM to ROCm mapping
│       └── errors.go           # ROCm error codes
├── configs/
│   ├── prometheus/
│   │   ├── federation.yaml     # Federation config template
│   │   └── slurm-alerts.yaml   # Slurm alert rules
│   └── grafana/
│       └── dashboards/
│           ├── slurm.json      # Slurm dashboard
│           └── multi-cluster.json
└── scripts/
    ├── install-rocm-exporter.sh
    └── install-slurm-exporter.sh
```

---

## Epic 1: Federation Support

### 1.1 Federation Types

**File:** `internal/federation/types.go`

```go
package federation

import "time"

type FederationConfig struct {
    Enabled     bool          `yaml:"enabled"`
    Type        FederationType `yaml:"type"` // "prometheus" or "thanos"
    Shards      []ShardConfig  `yaml:"shards"`
    CentralNode string         `yaml:"central_node"`
}

type FederationType string

const (
    FederationTypePrometheus FederationType = "prometheus"
    FederationTypeThanos     FederationType = "thanos"
)

type ShardConfig struct {
    Name       string   `yaml:"name"`
    Nodes      []string `yaml:"nodes"`      // Node names or patterns
    Racks      []string `yaml:"racks"`      // Rack identifiers
    Prometheus struct {
        Port        int    `yaml:"port"`
        StoragePath string `yaml:"storage_path"`
    } `yaml:"prometheus"`
}

type ShardStatus struct {
    Name        string
    Endpoint    string
    NodeCount   int
    Healthy     bool
    LastScrape  time.Time
    MetricCount int64
}

type CentralConfig struct {
    RetentionRaw     string `yaml:"retention_raw"`      // Short retention for raw metrics
    RetentionDownsampled string `yaml:"retention_downsampled"` // Long retention for aggregated
    FederateInterval string `yaml:"federate_interval"`  // How often to pull from shards
}
```

### 1.2 Shard Manager

**File:** `internal/federation/manager.go`

```go
package federation

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "text/template"

    "github.com/fregataa/aami/internal/config"
)

type Manager struct {
    config     *config.Config
    federation FederationConfig
}

func NewManager(cfg *config.Config, fed FederationConfig) *Manager {
    return &Manager{config: cfg, federation: fed}
}

// CalculateShards automatically determines shard distribution
func (m *Manager) CalculateShards(nodeCount, shardCount int) []ShardConfig {
    nodesPerShard := nodeCount / shardCount
    shards := make([]ShardConfig, shardCount)

    for i := 0; i < shardCount; i++ {
        shards[i] = ShardConfig{
            Name: fmt.Sprintf("shard-%d", i+1),
            Prometheus: struct {
                Port        int    `yaml:"port"`
                StoragePath string `yaml:"storage_path"`
            }{
                Port:        9090 + i,
                StoragePath: fmt.Sprintf("/var/lib/aami/prometheus-shard-%d", i+1),
            },
        }

        // Assign nodes to shard
        start := i * nodesPerShard
        end := start + nodesPerShard
        if i == shardCount-1 {
            end = nodeCount // Last shard takes remaining
        }

        for j := start; j < end; j++ {
            if j < len(m.config.Nodes) {
                shards[i].Nodes = append(shards[i].Nodes, m.config.Nodes[j].Name)
            }
        }
    }

    return shards
}

// Deploy deploys federation configuration
func (m *Manager) Deploy(ctx context.Context) error {
    for _, shard := range m.federation.Shards {
        if err := m.deployShard(ctx, shard); err != nil {
            return fmt.Errorf("deploy shard %s: %w", shard.Name, err)
        }
    }

    if err := m.deployCentral(ctx); err != nil {
        return fmt.Errorf("deploy central: %w", err)
    }

    return nil
}

func (m *Manager) deployShard(ctx context.Context, shard ShardConfig) error {
    // 1. Create shard directory
    if err := os.MkdirAll(shard.Prometheus.StoragePath, 0755); err != nil {
        return err
    }

    // 2. Generate Prometheus config for shard
    configPath := filepath.Join("/etc/aami", fmt.Sprintf("prometheus-%s.yaml", shard.Name))
    if err := m.generateShardConfig(shard, configPath); err != nil {
        return err
    }

    // 3. Create systemd service
    servicePath := fmt.Sprintf("/etc/systemd/system/aami-prometheus-%s.service", shard.Name)
    if err := m.createShardService(shard, servicePath); err != nil {
        return err
    }

    // 4. Start service
    return m.startService(fmt.Sprintf("aami-prometheus-%s", shard.Name))
}

const shardConfigTemplate = `
global:
  scrape_interval: 15s
  external_labels:
    shard: "{{ .Name }}"

scrape_configs:
  - job_name: 'node'
    file_sd_configs:
      - files:
          - '/var/lib/aami/targets/{{ .Name }}-nodes.json'

  - job_name: 'dcgm'
    file_sd_configs:
      - files:
          - '/var/lib/aami/targets/{{ .Name }}-dcgm.json'
`

func (m *Manager) generateShardConfig(shard ShardConfig, outputPath string) error {
    tmpl, err := template.New("shard").Parse(shardConfigTemplate)
    if err != nil {
        return err
    }

    f, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    return tmpl.Execute(f, shard)
}

const centralConfigTemplate = `
global:
  scrape_interval: 60s
  evaluation_interval: 60s

scrape_configs:
  - job_name: 'federation'
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{__name__=~"DCGM.*"}'
        - '{__name__=~"node.*"}'
        - 'up'
    static_configs:
{{ range .Shards }}
      - targets: ['{{ .Endpoint }}']
        labels:
          shard: '{{ .Name }}'
{{ end }}

rule_files:
  - '/etc/aami/rules/*.yaml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
`

func (m *Manager) deployCentral(ctx context.Context) error {
    // Generate central Prometheus config
    configPath := "/etc/aami/prometheus-central.yaml"

    tmpl, err := template.New("central").Parse(centralConfigTemplate)
    if err != nil {
        return err
    }

    data := struct {
        Shards []struct {
            Name     string
            Endpoint string
        }
    }{}

    for _, shard := range m.federation.Shards {
        data.Shards = append(data.Shards, struct {
            Name     string
            Endpoint string
        }{
            Name:     shard.Name,
            Endpoint: fmt.Sprintf("localhost:%d", shard.Prometheus.Port),
        })
    }

    f, err := os.Create(configPath)
    if err != nil {
        return err
    }
    defer f.Close()

    return tmpl.Execute(f, data)
}

func (m *Manager) createShardService(shard ShardConfig, servicePath string) error {
    // Create systemd service file for shard
    return nil
}

func (m *Manager) startService(name string) error {
    // Start systemd service
    return nil
}

// GetStatus returns status of all shards
func (m *Manager) GetStatus(ctx context.Context) ([]ShardStatus, error) {
    var statuses []ShardStatus

    for _, shard := range m.federation.Shards {
        status := ShardStatus{
            Name:      shard.Name,
            Endpoint:  fmt.Sprintf("localhost:%d", shard.Prometheus.Port),
            NodeCount: len(shard.Nodes),
        }

        // Check health
        status.Healthy = m.checkShardHealth(shard)

        statuses = append(statuses, status)
    }

    return statuses, nil
}

func (m *Manager) checkShardHealth(shard ShardConfig) bool {
    // HTTP check to shard endpoint
    return true
}
```

### 1.3 Federation CLI

**File:** `internal/cli/federation.go`

```go
package cli

import (
    "context"
    "fmt"
    "os"

    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/federation"
)

var federationCmd = &cobra.Command{
    Use:   "federation",
    Short: "Manage Prometheus federation",
}

var federationEnableCmd = &cobra.Command{
    Use:   "enable",
    Short: "Enable federation mode",
    RunE:  runFederationEnable,
}

var federationStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show federation status",
    RunE:  runFederationStatus,
}

var federationDisableCmd = &cobra.Command{
    Use:   "disable",
    Short: "Disable federation mode",
    RunE:  runFederationDisable,
}

var (
    shardCount int
    shardBy    string // "auto", "rack", "count"
)

func init() {
    federationEnableCmd.Flags().IntVar(&shardCount, "shards", 3,
        "Number of shards")
    federationEnableCmd.Flags().StringVar(&shardBy, "by", "auto",
        "Sharding strategy: auto, rack, count")

    federationCmd.AddCommand(federationEnableCmd)
    federationCmd.AddCommand(federationStatusCmd)
    federationCmd.AddCommand(federationDisableCmd)
    rootCmd.AddCommand(federationCmd)
}

func runFederationEnable(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    nodeCount := len(cfg.Nodes)
    if nodeCount < 100 {
        fmt.Printf("Warning: Federation is recommended for 500+ nodes. You have %d nodes.\n", nodeCount)
        fmt.Println("Continue anyway? [y/N]")
        var answer string
        fmt.Scanln(&answer)
        if answer != "y" && answer != "Y" {
            return nil
        }
    }

    fedConfig := federation.FederationConfig{
        Enabled: true,
        Type:    federation.FederationTypePrometheus,
    }

    manager := federation.NewManager(cfg, fedConfig)

    // Calculate shards
    shards := manager.CalculateShards(nodeCount, shardCount)
    fedConfig.Shards = shards

    fmt.Println("Creating Prometheus shards...")
    for _, shard := range shards {
        fmt.Printf("  %s: %d nodes\n", shard.Name, len(shard.Nodes))
    }

    // Deploy
    ctx := context.Background()
    if err := manager.Deploy(ctx); err != nil {
        return err
    }

    fmt.Println("\nCentral Prometheus configured")
    fmt.Println("Federation enabled")

    return nil
}

func runFederationStatus(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    // Load federation config
    fedConfig := federation.FederationConfig{} // Load from config

    manager := federation.NewManager(cfg, fedConfig)

    statuses, err := manager.GetStatus(context.Background())
    if err != nil {
        return err
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Shard", "Endpoint", "Nodes", "Status"})

    for _, s := range statuses {
        status := "Healthy"
        if !s.Healthy {
            status = "Unhealthy"
        }
        table.Append([]string{
            s.Name,
            s.Endpoint,
            fmt.Sprintf("%d", s.NodeCount),
            status,
        })
    }

    table.Render()
    return nil
}

func runFederationDisable(cmd *cobra.Command, args []string) error {
    fmt.Println("Disabling federation...")
    // Stop shard Prometheus instances
    // Consolidate to single Prometheus
    fmt.Println("Federation disabled")
    return nil
}
```

---

## Epic 2: Slurm Integration

### 2.1 Slurm Types

**File:** `internal/slurm/types.go`

```go
package slurm

import "time"

type Job struct {
    ID         int64
    Name       string
    User       string
    Partition  string
    State      JobState
    Nodes      []string
    GPUs       []GPUAllocation
    StartTime  time.Time
    EndTime    time.Time
    ExitCode   int
}

type JobState string

const (
    JobStatePending   JobState = "PENDING"
    JobStateRunning   JobState = "RUNNING"
    JobStateCompleted JobState = "COMPLETED"
    JobStateFailed    JobState = "FAILED"
    JobStateCancelled JobState = "CANCELLED"
    JobStateTimeout   JobState = "TIMEOUT"
)

type GPUAllocation struct {
    Node     string
    GPUIndex int
    UUID     string
}

type JobGPUCorrelation struct {
    Job          Job
    GPUEvents    []GPUEvent
    Correlation  CorrelationType
    Confidence   float64
    Recommendation string
}

type GPUEvent struct {
    Timestamp time.Time
    Node      string
    GPU       int
    Type      string  // "xid", "temperature", "ecc"
    Value     string
    Severity  string
}

type CorrelationType string

const (
    CorrelationNone      CorrelationType = "none"
    CorrelationLikely    CorrelationType = "likely"
    CorrelationConfirmed CorrelationType = "confirmed"
)

type SlurmConfig struct {
    Enabled      bool   `yaml:"enabled"`
    Endpoint     string `yaml:"endpoint"`       // slurmrestd endpoint or scontrol
    PreJobCheck  bool   `yaml:"pre_job_check"`  // Check GPU before job starts
    PostJobCheck bool   `yaml:"post_job_check"` // Check GPU after job ends
    AutoDrain    bool   `yaml:"auto_drain"`     // Drain node on GPU issue
}
```

### 2.2 Slurm Client

**File:** `internal/slurm/client.go`

```go
package slurm

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os/exec"
    "strconv"
    "strings"
    "time"
)

type Client struct {
    config     SlurmConfig
    httpClient *http.Client
}

func NewClient(cfg SlurmConfig) *Client {
    return &Client{
        config:     cfg,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

// GetJob retrieves job information
func (c *Client) GetJob(ctx context.Context, jobID int64) (*Job, error) {
    if c.config.Endpoint != "" {
        return c.getJobREST(ctx, jobID)
    }
    return c.getJobCLI(ctx, jobID)
}

func (c *Client) getJobREST(ctx context.Context, jobID int64) (*Job, error) {
    url := fmt.Sprintf("%s/slurm/v0.0.38/job/%d", c.config.Endpoint, jobID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Jobs []struct {
            JobID     int64  `json:"job_id"`
            Name      string `json:"name"`
            UserName  string `json:"user_name"`
            Partition string `json:"partition"`
            JobState  string `json:"job_state"`
            Nodes     string `json:"nodes"`
        } `json:"jobs"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if len(result.Jobs) == 0 {
        return nil, fmt.Errorf("job not found: %d", jobID)
    }

    j := result.Jobs[0]
    return &Job{
        ID:        j.JobID,
        Name:      j.Name,
        User:      j.UserName,
        Partition: j.Partition,
        State:     JobState(j.JobState),
        Nodes:     c.expandNodeList(j.Nodes),
    }, nil
}

func (c *Client) getJobCLI(ctx context.Context, jobID int64) (*Job, error) {
    cmd := exec.CommandContext(ctx, "scontrol", "show", "job", strconv.FormatInt(jobID, 10))
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    return c.parseScontrolOutput(string(output))
}

func (c *Client) parseScontrolOutput(output string) (*Job, error) {
    job := &Job{}

    for _, line := range strings.Split(output, "\n") {
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        switch key {
        case "JobId":
            job.ID, _ = strconv.ParseInt(value, 10, 64)
        case "JobName":
            job.Name = value
        case "UserId":
            job.User = strings.Split(value, "(")[0]
        case "Partition":
            job.Partition = value
        case "JobState":
            job.State = JobState(value)
        case "NodeList":
            job.Nodes = c.expandNodeList(value)
        }
    }

    return job, nil
}

// expandNodeList expands Slurm node list notation
// e.g., "gpu-node-[01-04]" -> ["gpu-node-01", "gpu-node-02", "gpu-node-03", "gpu-node-04"]
func (c *Client) expandNodeList(nodeList string) []string {
    // Use scontrol to expand
    cmd := exec.Command("scontrol", "show", "hostnames", nodeList)
    output, err := cmd.Output()
    if err != nil {
        return []string{nodeList}
    }

    var nodes []string
    for _, line := range strings.Split(string(output), "\n") {
        line = strings.TrimSpace(line)
        if line != "" {
            nodes = append(nodes, line)
        }
    }
    return nodes
}

// GetRunningJobs returns all currently running jobs
func (c *Client) GetRunningJobs(ctx context.Context) ([]Job, error) {
    cmd := exec.CommandContext(ctx, "squeue", "-h", "-o", "%i %j %u %P %T %N")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var jobs []Job
    for _, line := range strings.Split(string(output), "\n") {
        parts := strings.Fields(line)
        if len(parts) < 6 {
            continue
        }

        id, _ := strconv.ParseInt(parts[0], 10, 64)
        jobs = append(jobs, Job{
            ID:        id,
            Name:      parts[1],
            User:      parts[2],
            Partition: parts[3],
            State:     JobState(parts[4]),
            Nodes:     c.expandNodeList(parts[5]),
        })
    }

    return jobs, nil
}

// DrainNode drains a node from Slurm scheduling
func (c *Client) DrainNode(ctx context.Context, nodeName, reason string) error {
    cmd := exec.CommandContext(ctx, "scontrol", "update",
        fmt.Sprintf("NodeName=%s", nodeName),
        "State=DRAIN",
        fmt.Sprintf("Reason=\"%s\"", reason))
    return cmd.Run()
}
```

### 2.3 Job-GPU Analyzer

**File:** `internal/slurm/analyzer.go`

```go
package slurm

import (
    "context"
    "fmt"
    "time"

    "github.com/fregataa/aami/internal/health"
)

type Analyzer struct {
    slurmClient *Client
    prometheus  *health.PrometheusClient
}

func NewAnalyzer(slurm *Client, prom *health.PrometheusClient) *Analyzer {
    return &Analyzer{
        slurmClient: slurm,
        prometheus:  prom,
    }
}

// AnalyzeJob correlates job with GPU events
func (a *Analyzer) AnalyzeJob(ctx context.Context, jobID int64) (*JobGPUCorrelation, error) {
    job, err := a.slurmClient.GetJob(ctx, jobID)
    if err != nil {
        return nil, err
    }

    result := &JobGPUCorrelation{
        Job:         *job,
        Correlation: CorrelationNone,
    }

    // Query GPU events during job execution
    events, err := a.queryGPUEvents(ctx, job)
    if err != nil {
        return nil, err
    }
    result.GPUEvents = events

    // Analyze correlation
    a.analyzeCorrelation(result)

    return result, nil
}

func (a *Analyzer) queryGPUEvents(ctx context.Context, job *Job) ([]GPUEvent, error) {
    var events []GPUEvent

    // Query Xid errors
    for _, node := range job.Nodes {
        query := fmt.Sprintf(
            `DCGM_FI_DEV_XID_ERRORS{instance=~"%s.*"} > 0`, node)

        // Query Prometheus for events during job time range
        xidEvents, _ := a.queryPrometheusEvents(ctx, query, job.StartTime, job.EndTime)
        events = append(events, xidEvents...)

        // Query temperature spikes
        tempQuery := fmt.Sprintf(
            `DCGM_FI_DEV_GPU_TEMP{instance=~"%s.*"} > 85`, node)
        tempEvents, _ := a.queryPrometheusEvents(ctx, tempQuery, job.StartTime, job.EndTime)
        events = append(events, tempEvents...)

        // Query ECC errors
        eccQuery := fmt.Sprintf(
            `increase(DCGM_FI_DEV_ECC_DBE_VOL_TOTAL{instance=~"%s.*"}[5m]) > 0`, node)
        eccEvents, _ := a.queryPrometheusEvents(ctx, eccQuery, job.StartTime, job.EndTime)
        events = append(events, eccEvents...)
    }

    return events, nil
}

func (a *Analyzer) queryPrometheusEvents(ctx context.Context, query string, start, end time.Time) ([]GPUEvent, error) {
    // Query Prometheus range
    return nil, nil
}

func (a *Analyzer) analyzeCorrelation(result *JobGPUCorrelation) {
    if len(result.GPUEvents) == 0 {
        result.Correlation = CorrelationNone
        result.Recommendation = "No GPU issues detected during job execution"
        return
    }

    // Check for critical events
    hasXid := false
    hasHighTemp := false

    for _, event := range result.GPUEvents {
        if event.Type == "xid" {
            hasXid = true
        }
        if event.Type == "temperature" {
            hasHighTemp = true
        }
    }

    if hasXid {
        result.Correlation = CorrelationConfirmed
        result.Confidence = 0.9

        // Find the specific Xid and node
        for _, event := range result.GPUEvents {
            if event.Type == "xid" {
                result.Recommendation = fmt.Sprintf(
                    "Job failure likely caused by GPU hardware issue on %s GPU %d. "+
                        "Recommend: Exclude node from scheduling, inspect GPU.",
                    event.Node, event.GPU)
                break
            }
        }
    } else if hasHighTemp {
        result.Correlation = CorrelationLikely
        result.Confidence = 0.7
        result.Recommendation = "High GPU temperature detected. Consider throttling or cooling issues."
    } else {
        result.Correlation = CorrelationLikely
        result.Confidence = 0.5
        result.Recommendation = "GPU events detected but correlation uncertain."
    }
}

// FindAffectedJobs finds jobs affected by GPU issues on a node
func (a *Analyzer) FindAffectedJobs(ctx context.Context, node string, startTime time.Time) ([]Job, error) {
    allJobs, err := a.slurmClient.GetRunningJobs(ctx)
    if err != nil {
        return nil, err
    }

    var affected []Job
    for _, job := range allJobs {
        for _, jobNode := range job.Nodes {
            if jobNode == node {
                affected = append(affected, job)
                break
            }
        }
    }

    return affected, nil
}
```

### 2.4 Pre/Post Job Hooks

**File:** `internal/slurm/hooks.go`

```go
package slurm

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

type HookManager struct {
    config     SlurmConfig
    slurmClient *Client
}

func NewHookManager(cfg SlurmConfig, client *Client) *HookManager {
    return &HookManager{config: cfg, slurmClient: client}
}

// InstallHooks installs Slurm prolog/epilog scripts
func (h *HookManager) InstallHooks(prologPath, epilogPath string) error {
    if h.config.PreJobCheck {
        if err := h.installProlog(prologPath); err != nil {
            return err
        }
    }

    if h.config.PostJobCheck {
        if err := h.installEpilog(epilogPath); err != nil {
            return err
        }
    }

    return nil
}

const prologScript = `#!/bin/bash
# AAMI Pre-Job GPU Check
# Installed by aami slurm install-hooks

AAMI_BIN="/usr/local/bin/aami"
NODE=$(hostname)
JOB_ID=$SLURM_JOB_ID

# Check GPU health before job starts
result=$($AAMI_BIN health $NODE --json 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "Warning: Could not check GPU health on $NODE" >&2
    exit 0  # Don't block job on check failure
fi

# Parse health score
score=$(echo "$result" | jq -r '.overall')
if [ "$score" -lt 50 ]; then
    echo "ERROR: GPU health critical on $NODE (score: $score)" >&2
    echo "Job $JOB_ID rejected due to unhealthy GPU" >&2
    $AAMI_BIN slurm drain $NODE --reason "GPU health critical (score: $score)"
    exit 1
fi

exit 0
`

const epilogScript = `#!/bin/bash
# AAMI Post-Job GPU Check
# Installed by aami slurm install-hooks

AAMI_BIN="/usr/local/bin/aami"
NODE=$(hostname)
JOB_ID=$SLURM_JOB_ID
EXIT_CODE=$SLURM_JOB_EXIT_CODE

# Check GPU health after job completes
result=$($AAMI_BIN health $NODE --json 2>/dev/null)
if [ $? -ne 0 ]; then
    exit 0
fi

score=$(echo "$result" | jq -r '.overall')

# If job failed and GPU health is low, correlate
if [ "$EXIT_CODE" -ne 0 ] && [ "$score" -lt 70 ]; then
    $AAMI_BIN slurm log-correlation \
        --job $JOB_ID \
        --node $NODE \
        --score $score \
        --exit-code $EXIT_CODE
fi

# Auto-drain if health is critical
if [ "$score" -lt 50 ]; then
    $AAMI_BIN slurm drain $NODE --reason "GPU health degraded after job $JOB_ID"
fi

exit 0
`

func (h *HookManager) installProlog(path string) error {
    return os.WriteFile(path, []byte(prologScript), 0755)
}

func (h *HookManager) installEpilog(path string) error {
    return os.WriteFile(path, []byte(epilogScript), 0755)
}

// GenerateSlurmConf generates slurm.conf snippet for hooks
func (h *HookManager) GenerateSlurmConf(prologPath, epilogPath string) string {
    return fmt.Sprintf(`# AAMI GPU Health Hooks
Prolog=%s
Epilog=%s
`, prologPath, epilogPath)
}
```

### 2.5 Slurm CLI

**File:** `internal/cli/slurm.go`

```go
package cli

import (
    "context"
    "fmt"
    "os"
    "strconv"

    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/health"
    "github.com/fregataa/aami/internal/slurm"
)

var slurmCmd = &cobra.Command{
    Use:   "slurm",
    Short: "Slurm integration commands",
}

var slurmAnalyzeCmd = &cobra.Command{
    Use:   "job-analyze [job-id]",
    Short: "Analyze job-GPU correlation",
    Args:  cobra.ExactArgs(1),
    RunE:  runSlurmAnalyze,
}

var slurmDrainCmd = &cobra.Command{
    Use:   "drain [node]",
    Short: "Drain node from Slurm",
    Args:  cobra.ExactArgs(1),
    RunE:  runSlurmDrain,
}

var slurmInstallHooksCmd = &cobra.Command{
    Use:   "install-hooks",
    Short: "Install Slurm prolog/epilog hooks",
    RunE:  runSlurmInstallHooks,
}

var drainReason string

func init() {
    slurmDrainCmd.Flags().StringVar(&drainReason, "reason", "",
        "Reason for draining node")

    slurmCmd.AddCommand(slurmAnalyzeCmd)
    slurmCmd.AddCommand(slurmDrainCmd)
    slurmCmd.AddCommand(slurmInstallHooksCmd)
    rootCmd.AddCommand(slurmCmd)
}

func runSlurmAnalyze(cmd *cobra.Command, args []string) error {
    jobID, err := strconv.ParseInt(args[0], 10, 64)
    if err != nil {
        return fmt.Errorf("invalid job ID: %s", args[0])
    }

    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    slurmClient := slurm.NewClient(slurm.SlurmConfig{})
    promClient, _ := health.NewPrometheusClient(
        fmt.Sprintf("http://localhost:%d", cfg.Prometheus.Port))

    analyzer := slurm.NewAnalyzer(slurmClient, promClient)

    result, err := analyzer.AnalyzeJob(context.Background(), jobID)
    if err != nil {
        return err
    }

    // Print analysis
    fmt.Printf("Job %d Analysis\n", jobID)
    fmt.Println(strings.Repeat("━", 55))
    fmt.Printf("Job: %d\n", result.Job.ID)
    fmt.Printf("User: %s\n", result.Job.User)
    fmt.Printf("Nodes: %s\n", strings.Join(result.Job.Nodes, ", "))
    fmt.Printf("State: %s\n", result.Job.State)
    fmt.Println()

    if len(result.GPUEvents) > 0 {
        fmt.Println("GPU Events During Job:")
        for _, event := range result.GPUEvents {
            fmt.Printf("  [%s] %s GPU %d: %s\n",
                event.Timestamp.Format("15:04:05"),
                event.Node, event.GPU, event.Value)
        }
        fmt.Println()
    }

    fmt.Printf("Correlation: %s (confidence: %.0f%%)\n",
        result.Correlation, result.Confidence*100)
    fmt.Printf("Recommendation: %s\n", result.Recommendation)

    return nil
}

func runSlurmDrain(cmd *cobra.Command, args []string) error {
    node := args[0]
    reason := drainReason
    if reason == "" {
        reason = "Drained by AAMI"
    }

    slurmClient := slurm.NewClient(slurm.SlurmConfig{})
    if err := slurmClient.DrainNode(context.Background(), node, reason); err != nil {
        return err
    }

    fmt.Printf("Node %s drained: %s\n", node, reason)
    return nil
}

func runSlurmInstallHooks(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    slurmClient := slurm.NewClient(cfg.Slurm)
    hookMgr := slurm.NewHookManager(cfg.Slurm, slurmClient)

    prologPath := "/etc/slurm/aami-prolog.sh"
    epilogPath := "/etc/slurm/aami-epilog.sh"

    if err := hookMgr.InstallHooks(prologPath, epilogPath); err != nil {
        return err
    }

    fmt.Println("Hooks installed:")
    fmt.Printf("  Prolog: %s\n", prologPath)
    fmt.Printf("  Epilog: %s\n", epilogPath)
    fmt.Println("\nAdd the following to slurm.conf:")
    fmt.Println(hookMgr.GenerateSlurmConf(prologPath, epilogPath))

    return nil
}
```

---

## Epic 3: Multi-Cluster Support

### 3.1 Multi-Cluster Types

**File:** `internal/multicluster/types.go`

```go
package multicluster

import "time"

type ClusterConfig struct {
    Name      string `yaml:"name"`
    Endpoint  string `yaml:"endpoint"`
    APIKey    string `yaml:"api_key"`
    TLSCert   string `yaml:"tls_cert"`
    TLSKey    string `yaml:"tls_key"`
    TLSCACert string `yaml:"tls_ca_cert"`
}

type ClusterStatus struct {
    Name           string
    Endpoint       string
    Nodes          int
    HealthyNodes   int
    HealthScore    float64
    Connected      bool
    LastSync       time.Time
    AlertsActive   int
}

type GlobalAlert struct {
    Cluster     string
    AlertName   string
    Severity    string
    Node        string
    Description string
    FiredAt     time.Time
}

type AggregatedMetrics struct {
    TotalNodes       int
    TotalGPUs        int
    AverageHealth    float64
    ActiveAlerts     int
    ClusterBreakdown map[string]ClusterMetrics
}

type ClusterMetrics struct {
    Nodes        int
    GPUs         int
    HealthScore  float64
    AlertCount   int
}
```

### 3.2 Cluster Registry

**File:** `internal/multicluster/registry.go`

```go
package multicluster

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Registry struct {
    path     string
    clusters map[string]ClusterConfig
}

func NewRegistry(path string) *Registry {
    return &Registry{
        path:     path,
        clusters: make(map[string]ClusterConfig),
    }
}

func (r *Registry) Load() error {
    data, err := os.ReadFile(r.path)
    if os.IsNotExist(err) {
        return nil // No clusters configured yet
    }
    if err != nil {
        return err
    }

    var config struct {
        Clusters []ClusterConfig `yaml:"clusters"`
    }
    if err := yaml.Unmarshal(data, &config); err != nil {
        return err
    }

    for _, c := range config.Clusters {
        r.clusters[c.Name] = c
    }
    return nil
}

func (r *Registry) Save() error {
    var clusters []ClusterConfig
    for _, c := range r.clusters {
        clusters = append(clusters, c)
    }

    config := struct {
        Clusters []ClusterConfig `yaml:"clusters"`
    }{Clusters: clusters}

    data, err := yaml.Marshal(config)
    if err != nil {
        return err
    }

    return os.WriteFile(r.path, data, 0600)
}

func (r *Registry) Add(cluster ClusterConfig) error {
    if _, exists := r.clusters[cluster.Name]; exists {
        return fmt.Errorf("cluster already exists: %s", cluster.Name)
    }
    r.clusters[cluster.Name] = cluster
    return r.Save()
}

func (r *Registry) Remove(name string) error {
    if _, exists := r.clusters[name]; !exists {
        return fmt.Errorf("cluster not found: %s", name)
    }
    delete(r.clusters, name)
    return r.Save()
}

func (r *Registry) Get(name string) (ClusterConfig, bool) {
    c, ok := r.clusters[name]
    return c, ok
}

func (r *Registry) List() []ClusterConfig {
    var list []ClusterConfig
    for _, c := range r.clusters {
        list = append(list, c)
    }
    return list
}
```

### 3.3 Remote AAMI Client

**File:** `internal/multicluster/client.go`

```go
package multicluster

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

type Client struct {
    config     ClusterConfig
    httpClient *http.Client
}

func NewClient(cfg ClusterConfig) (*Client, error) {
    client := &Client{config: cfg}

    transport := &http.Transport{}

    // Configure TLS if provided
    if cfg.TLSCert != "" {
        cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
        if err != nil {
            return nil, err
        }

        caCert, err := os.ReadFile(cfg.TLSCACert)
        if err != nil {
            return nil, err
        }

        caCertPool := x509.NewCertPool()
        caCertPool.AppendCertsFromPEM(caCert)

        transport.TLSClientConfig = &tls.Config{
            Certificates: []tls.Certificate{cert},
            RootCAs:      caCertPool,
        }
    }

    client.httpClient = &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }

    return client, nil
}

// GetStatus retrieves cluster status
func (c *Client) GetStatus(ctx context.Context) (*ClusterStatus, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        c.config.Endpoint+"/api/v1/status", nil)
    if err != nil {
        return nil, err
    }

    if c.config.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return &ClusterStatus{
            Name:      c.config.Name,
            Endpoint:  c.config.Endpoint,
            Connected: false,
        }, nil
    }
    defer resp.Body.Close()

    var status ClusterStatus
    if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
        return nil, err
    }

    status.Connected = true
    return &status, nil
}

// GetHealth retrieves cluster health
func (c *Client) GetHealth(ctx context.Context) (*ClusterMetrics, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        c.config.Endpoint+"/api/v1/health", nil)
    if err != nil {
        return nil, err
    }

    if c.config.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var metrics ClusterMetrics
    if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
        return nil, err
    }

    return &metrics, nil
}

// GetAlerts retrieves active alerts
func (c *Client) GetAlerts(ctx context.Context) ([]GlobalAlert, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        c.config.Endpoint+"/api/v1/alerts", nil)
    if err != nil {
        return nil, err
    }

    if c.config.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var alerts []GlobalAlert
    if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
        return nil, err
    }

    // Tag with cluster name
    for i := range alerts {
        alerts[i].Cluster = c.config.Name
    }

    return alerts, nil
}
```

### 3.4 Multi-Cluster CLI

**File:** `internal/cli/clusters.go`

```go
package cli

import (
    "context"
    "fmt"
    "os"

    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/multicluster"
)

var clustersCmd = &cobra.Command{
    Use:   "clusters",
    Short: "Manage multiple AAMI clusters",
}

var clustersAddCmd = &cobra.Command{
    Use:   "add [name]",
    Short: "Add a remote cluster",
    Args:  cobra.ExactArgs(1),
    RunE:  runClustersAdd,
}

var clustersListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all clusters",
    RunE:  runClustersList,
}

var clustersRemoveCmd = &cobra.Command{
    Use:   "remove [name]",
    Short: "Remove a cluster",
    Args:  cobra.ExactArgs(1),
    RunE:  runClustersRemove,
}

var clustersStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show status of all clusters",
    RunE:  runClustersStatus,
}

var (
    clusterEndpoint string
    clusterAPIKey   string
)

func init() {
    clustersAddCmd.Flags().StringVar(&clusterEndpoint, "endpoint", "",
        "Cluster API endpoint (required)")
    clustersAddCmd.Flags().StringVar(&clusterAPIKey, "api-key", "",
        "API key for authentication")
    clustersAddCmd.MarkFlagRequired("endpoint")

    clustersCmd.AddCommand(clustersAddCmd)
    clustersCmd.AddCommand(clustersListCmd)
    clustersCmd.AddCommand(clustersRemoveCmd)
    clustersCmd.AddCommand(clustersStatusCmd)
    rootCmd.AddCommand(clustersCmd)
}

func runClustersAdd(cmd *cobra.Command, args []string) error {
    name := args[0]

    registry := multicluster.NewRegistry("/etc/aami/clusters.yaml")
    if err := registry.Load(); err != nil {
        return err
    }

    cluster := multicluster.ClusterConfig{
        Name:     name,
        Endpoint: clusterEndpoint,
        APIKey:   clusterAPIKey,
    }

    // Test connection
    client, err := multicluster.NewClient(cluster)
    if err != nil {
        return err
    }

    status, err := client.GetStatus(context.Background())
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }

    if !status.Connected {
        return fmt.Errorf("cannot connect to cluster")
    }

    if err := registry.Add(cluster); err != nil {
        return err
    }

    fmt.Printf("Cluster %s added (%d nodes)\n", name, status.Nodes)
    return nil
}

func runClustersList(cmd *cobra.Command, args []string) error {
    registry := multicluster.NewRegistry("/etc/aami/clusters.yaml")
    if err := registry.Load(); err != nil {
        return err
    }

    clusters := registry.List()

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Name", "Endpoint", "Nodes", "Status"})

    for _, c := range clusters {
        client, _ := multicluster.NewClient(c)
        status, _ := client.GetStatus(context.Background())

        nodes := "?"
        statusStr := "Unknown"
        if status != nil {
            nodes = fmt.Sprintf("%d", status.Nodes)
            if status.Connected {
                statusStr = "Connected"
            } else {
                statusStr = "Disconnected"
            }
        }

        table.Append([]string{c.Name, c.Endpoint, nodes, statusStr})
    }

    // Add local cluster
    cfg, _ := loadConfig()
    if cfg != nil {
        table.Append([]string{"(local)", "localhost", fmt.Sprintf("%d", len(cfg.Nodes)), "Connected"})
    }

    table.Render()
    return nil
}

func runClustersRemove(cmd *cobra.Command, args []string) error {
    name := args[0]

    registry := multicluster.NewRegistry("/etc/aami/clusters.yaml")
    if err := registry.Load(); err != nil {
        return err
    }

    if err := registry.Remove(name); err != nil {
        return err
    }

    fmt.Printf("Cluster %s removed\n", name)
    return nil
}

func runClustersStatus(cmd *cobra.Command, args []string) error {
    registry := multicluster.NewRegistry("/etc/aami/clusters.yaml")
    if err := registry.Load(); err != nil {
        return err
    }

    clusters := registry.List()

    fmt.Println("Multi-Cluster Status")
    fmt.Println(strings.Repeat("━", 60))

    var totalNodes, totalAlerts int
    var totalHealth float64

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Cluster", "Nodes", "Health", "Alerts", "Status"})

    for _, c := range clusters {
        client, _ := multicluster.NewClient(c)
        status, _ := client.GetStatus(context.Background())

        if status != nil && status.Connected {
            totalNodes += status.Nodes
            totalAlerts += status.AlertsActive
            totalHealth += status.HealthScore

            table.Append([]string{
                c.Name,
                fmt.Sprintf("%d", status.Nodes),
                fmt.Sprintf("%.0f", status.HealthScore),
                fmt.Sprintf("%d", status.AlertsActive),
                "Connected",
            })
        } else {
            table.Append([]string{c.Name, "?", "?", "?", "Disconnected"})
        }
    }

    table.Render()

    if len(clusters) > 0 {
        fmt.Printf("\nTotal: %d nodes, %.0f avg health, %d active alerts\n",
            totalNodes, totalHealth/float64(len(clusters)), totalAlerts)
    }

    return nil
}
```

---

## Epic 4: AMD GPU Support

### 4.1 AMD GPU Types

**File:** `internal/amd/types.go`

```go
package amd

type ROCmMetric struct {
    Name          string
    DCGMEquivalent string
    Description   string
}

// Metric mapping from DCGM to ROCm
var MetricMapping = map[string]ROCmMetric{
    "DCGM_FI_DEV_GPU_TEMP": {
        Name:          "rocm_gpu_temperature",
        DCGMEquivalent: "DCGM_FI_DEV_GPU_TEMP",
        Description:   "GPU temperature in Celsius",
    },
    "DCGM_FI_DEV_GPU_UTIL": {
        Name:          "rocm_gpu_utilization",
        DCGMEquivalent: "DCGM_FI_DEV_GPU_UTIL",
        Description:   "GPU utilization percentage",
    },
    "DCGM_FI_DEV_FB_USED": {
        Name:          "rocm_memory_used",
        DCGMEquivalent: "DCGM_FI_DEV_FB_USED",
        Description:   "GPU memory used in bytes",
    },
    "DCGM_FI_DEV_FB_TOTAL": {
        Name:          "rocm_memory_total",
        DCGMEquivalent: "DCGM_FI_DEV_FB_TOTAL",
        Description:   "GPU memory total in bytes",
    },
    "DCGM_FI_DEV_POWER_USAGE": {
        Name:          "rocm_power_usage",
        DCGMEquivalent: "DCGM_FI_DEV_POWER_USAGE",
        Description:   "GPU power usage in watts",
    },
}

type ROCmError struct {
    Code        int
    Name        string
    Severity    string
    Description string
    Causes      []string
    Actions     []string
}

var ROCmErrors = map[int]ROCmError{
    1: {
        Code:        1,
        Name:        "GPU Memory Error",
        Severity:    "Critical",
        Description: "Uncorrectable memory error detected",
        Causes:      []string{"Memory hardware failure", "Memory aging"},
        Actions:     []string{"Check memory with rocm-smi", "Schedule GPU replacement"},
    },
    2: {
        Code:        2,
        Name:        "GPU Hang",
        Severity:    "Critical",
        Description: "GPU is not responding",
        Causes:      []string{"Driver bug", "Hardware failure"},
        Actions:     []string{"Reset GPU", "Update ROCm driver"},
    },
    // Add more error codes as needed
}

type GPUType string

const (
    GPUTypeNVIDIA GPUType = "nvidia"
    GPUTypeAMD    GPUType = "amd"
)
```

### 4.2 ROCm Exporter Installer

**File:** `internal/amd/installer.go`

```go
package amd

import (
    "context"
    "fmt"

    "github.com/fregataa/aami/internal/ssh"
)

type Installer struct {
    executor *ssh.Executor
}

func NewInstaller(executor *ssh.Executor) *Installer {
    return &Installer{executor: executor}
}

const rocmExporterInstallScript = `#!/bin/bash
set -e

# Check if ROCm is installed
if ! command -v rocm-smi &> /dev/null; then
    echo "ERROR: ROCm is not installed" >&2
    exit 1
fi

# Install prometheus-amd-gpu-exporter
# This is a placeholder - actual installation depends on the exporter used
cd /tmp

# Option 1: Use amd_exporter (if available)
# wget https://github.com/amd/amd_exporter/releases/download/v1.0.0/amd_exporter_linux_amd64.tar.gz
# tar -xzf amd_exporter_linux_amd64.tar.gz
# mv amd_exporter /usr/local/bin/

# Option 2: Use rocm_smi_exporter
git clone https://github.com/amd/rocm_smi_exporter.git
cd rocm_smi_exporter
make
cp rocm_smi_exporter /usr/local/bin/

# Create systemd service
cat > /etc/systemd/system/rocm-exporter.service << 'EOF'
[Unit]
Description=ROCm SMI Prometheus Exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rocm_smi_exporter --web.listen-address=:9401
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable rocm-exporter
systemctl start rocm-exporter

echo "ROCm exporter installed and running on port 9401"
`

func (i *Installer) Install(ctx context.Context, node ssh.Node) error {
    result := i.executor.Run(ctx, node, rocmExporterInstallScript)
    if result.Error != nil {
        return fmt.Errorf("install failed: %s", result.Output)
    }
    return nil
}

// DetectGPUType detects whether the node has NVIDIA or AMD GPUs
func (i *Installer) DetectGPUType(ctx context.Context, node ssh.Node) (GPUType, error) {
    // Try nvidia-smi first
    nvidiaResult := i.executor.Run(ctx, node, "nvidia-smi -L 2>/dev/null")
    if nvidiaResult.Error == nil && nvidiaResult.Output != "" {
        return GPUTypeNVIDIA, nil
    }

    // Try rocm-smi
    amdResult := i.executor.Run(ctx, node, "rocm-smi -i 2>/dev/null")
    if amdResult.Error == nil && amdResult.Output != "" {
        return GPUTypeAMD, nil
    }

    return "", fmt.Errorf("no GPU detected on node")
}
```

### 4.3 AMD CLI Commands

**File:** `internal/cli/amd.go`

```go
package cli

import (
    "context"
    "fmt"
    "strconv"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/amd"
    "github.com/fregataa/aami/internal/ssh"
)

var amdCmd = &cobra.Command{
    Use:   "amd",
    Short: "AMD GPU specific commands",
}

var amdExplainCmd = &cobra.Command{
    Use:   "explain [error-code]",
    Short: "Explain AMD ROCm error code",
    Args:  cobra.ExactArgs(1),
    RunE:  runAmdExplain,
}

var amdInstallCmd = &cobra.Command{
    Use:   "install [node]",
    Short: "Install ROCm exporter on node",
    Args:  cobra.ExactArgs(1),
    RunE:  runAmdInstall,
}

func init() {
    amdCmd.AddCommand(amdExplainCmd)
    amdCmd.AddCommand(amdInstallCmd)
    rootCmd.AddCommand(amdCmd)
}

func runAmdExplain(cmd *cobra.Command, args []string) error {
    code, err := strconv.Atoi(args[0])
    if err != nil {
        return fmt.Errorf("invalid error code: %s", args[0])
    }

    errInfo, ok := amd.ROCmErrors[code]
    if !ok {
        return fmt.Errorf("unknown ROCm error code: %d", code)
    }

    red := color.New(color.FgRed).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()

    fmt.Printf("ROCm Error %d: %s\n\n", code, errInfo.Name)

    severity := errInfo.Severity
    if severity == "Critical" {
        severity = red(severity)
    } else if severity == "Warning" {
        severity = yellow(severity)
    }
    fmt.Printf("Severity: %s\n\n", severity)

    fmt.Printf("Description:\n  %s\n\n", errInfo.Description)

    fmt.Println("Common Causes:")
    for i, cause := range errInfo.Causes {
        fmt.Printf("  %d. %s\n", i+1, cause)
    }
    fmt.Println()

    fmt.Println("Recommended Actions:")
    for i, action := range errInfo.Actions {
        fmt.Printf("  %d. %s\n", i+1, action)
    }

    return nil
}

func runAmdInstall(cmd *cobra.Command, args []string) error {
    nodeName := args[0]

    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    var node *ssh.Node
    for _, n := range cfg.Nodes {
        if n.Name == nodeName {
            node = &ssh.Node{
                Name:    n.Name,
                Host:    n.IP,
                Port:    n.SSHPort,
                User:    n.SSHUser,
                KeyPath: n.SSHKey,
            }
            break
        }
    }
    if node == nil {
        return fmt.Errorf("node not found: %s", nodeName)
    }

    executor := ssh.NewExecutor(ssh.ExecutorConfig{
        ConnectTimeout: 10 * time.Second,
        CommandTimeout: 300 * time.Second,
        MaxRetries:     3,
    })

    installer := amd.NewInstaller(executor)

    fmt.Printf("Installing ROCm exporter on %s...\n", nodeName)
    if err := installer.Install(context.Background(), *node); err != nil {
        return err
    }

    fmt.Println("ROCm exporter installed successfully")
    return nil
}
```

---

## Test Commands

```bash
# Build
go build -o aami ./cmd/aami

# Federation tests
./aami federation enable --shards 3
./aami federation status
./aami federation disable

# Slurm tests
./aami slurm job-analyze 12345
./aami slurm drain gpu-node-01 --reason "GPU maintenance"
./aami slurm install-hooks

# Multi-cluster tests
./aami clusters add cluster-east --endpoint https://aami-east.example.com
./aami clusters list
./aami clusters status
./aami clusters remove cluster-east

# AMD tests
./aami amd explain 1
./aami amd install amd-node-01
```

---

## Acceptance Criteria

| Feature | Test Command | Expected Output |
|---------|--------------|-----------------|
| Federation enable | `aami federation enable --shards 3` | Creates 3 shard Prometheus instances |
| Federation status | `aami federation status` | Table of shard health |
| Slurm analyze | `aami slurm job-analyze 12345` | Job-GPU correlation report |
| Slurm drain | `aami slurm drain node01` | Node drained from Slurm |
| Cluster add | `aami clusters add name --endpoint url` | Cluster registered |
| Cluster list | `aami clusters list` | Table of all clusters |
| Cluster status | `aami clusters status` | Aggregated status of all clusters |
| AMD explain | `aami amd explain 1` | ROCm error explanation |
| AMD install | `aami amd install node01` | ROCm exporter installed |
