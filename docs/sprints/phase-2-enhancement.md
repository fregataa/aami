# Phase 2: Enhancement

## Overview

- **Duration**: 3-4 weeks
- **Goal**: Differentiation features, operational convenience
- **Prerequisites**: Phase 1 MVP completed

## New Files

```
aami/
├── internal/
│   ├── cli/
│   │   ├── topology.go          # NVLink topology command
│   │   ├── health.go            # Health score command
│   │   ├── upgrade.go           # Upgrade command
│   │   ├── backup.go            # Backup/restore commands
│   │   ├── diagnose.go          # Diagnose command
│   │   └── diff.go              # Config diff command
│   ├── nvlink/                  # NVLink topology
│   │   ├── collector.go         # Data collection
│   │   ├── topology.go          # Topology parsing
│   │   └── renderer.go          # ASCII art rendering
│   ├── health/                  # GPU health score
│   │   ├── calculator.go        # Score calculation
│   │   ├── weights.go           # Weight configuration
│   │   └── prometheus.go        # Prometheus queries
│   ├── upgrade/                 # Component upgrade
│   │   ├── checker.go           # Version checker
│   │   ├── upgrader.go          # Upgrade logic
│   │   └── rollback.go          # Rollback support
│   └── backup/                  # Backup/restore
│       ├── backup.go            # Backup logic
│       ├── restore.go           # Restore logic
│       └── schedule.go          # Scheduled backups
├── configs/
│   └── prometheus/
│       └── nvlink-alerts.yaml   # NVLink alert rules
└── api/                         # Web UI API (optional)
    ├── server.go
    ├── handlers/
    │   ├── topology.go
    │   └── health.go
    └── types.go
```

## Additional Dependencies

```go
// go.mod additions
require (
    github.com/prometheus/client_golang v1.18.0  // Prometheus API client
    github.com/muesli/termenv v0.15.2            // Terminal styling
)
```

---

## Task 1: NVLink Topology Visualization

### 1.1 NVLink Types

**File:** `internal/nvlink/types.go`

```go
package nvlink

type Topology struct {
    Node     string
    GPUs     []GPU
    Links    []Link
    Errors   []LinkError
}

type GPU struct {
    Index       int
    UUID        string
    Name        string
    PCIBusID    string
}

type Link struct {
    SourceGPU   int
    TargetGPU   int
    Type        LinkType   // NVLink, PCIe, etc.
    Bandwidth   float64    // GB/s
    Status      LinkStatus
}

type LinkType string

const (
    LinkTypeNVLink LinkType = "NVLINK"
    LinkTypePCIe   LinkType = "PCIE"
    LinkTypeSXM    LinkType = "SXM"
)

type LinkStatus string

const (
    LinkStatusOK       LinkStatus = "ok"
    LinkStatusDegraded LinkStatus = "degraded"
    LinkStatusDown     LinkStatus = "down"
)

type LinkError struct {
    SourceGPU    int
    TargetGPU    int
    CRCErrors    int64
    RecoveryErrs int64
    Timestamp    time.Time
}
```

### 1.2 NVLink Data Collector

**File:** `internal/nvlink/collector.go`

```go
package nvlink

import (
    "context"
    "fmt"
    "strings"

    "github.com/fregataa/aami/internal/ssh"
)

type Collector struct {
    executor *ssh.Executor
}

func NewCollector(executor *ssh.Executor) *Collector {
    return &Collector{executor: executor}
}

// CollectTopology runs nvidia-smi topo and parses the output
func (c *Collector) CollectTopology(ctx context.Context, node ssh.Node) (*Topology, error) {
    // Get topology matrix
    topoResult := c.executor.Run(ctx, node, "nvidia-smi topo -m")
    if topoResult.Error != nil {
        return nil, fmt.Errorf("nvidia-smi topo: %w", topoResult.Error)
    }

    topo, err := c.parseTopologyMatrix(topoResult.Output)
    if err != nil {
        return nil, err
    }
    topo.Node = node.Name

    // Get NVLink errors from DCGM/nvidia-smi
    errResult := c.executor.Run(ctx, node,
        "nvidia-smi nvlink --status -i all")
    if errResult.Error == nil {
        errors := c.parseNVLinkErrors(errResult.Output)
        topo.Errors = errors
        c.updateLinkStatus(topo)
    }

    return topo, nil
}

// parseTopologyMatrix parses nvidia-smi topo -m output
// Example output:
//
//	GPU0    GPU1    GPU2    GPU3    CPU Affinity
//
// GPU0     X      NV12    NV12    NV12    0-31
// GPU1    NV12     X      NV12    NV12    0-31
func (c *Collector) parseTopologyMatrix(output string) (*Topology, error) {
    lines := strings.Split(output, "\n")
    topo := &Topology{}

    // Parse header for GPU count
    // Parse each row for link types
    for i, line := range lines {
        if strings.HasPrefix(line, "GPU") && i > 0 {
            // Parse GPU row
            parts := strings.Fields(line)
            if len(parts) < 2 {
                continue
            }

            gpuIdx := c.parseGPUIndex(parts[0])
            gpu := GPU{Index: gpuIdx}
            topo.GPUs = append(topo.GPUs, gpu)

            // Parse links to other GPUs
            for j := 1; j < len(parts)-1; j++ { // exclude CPU Affinity
                if parts[j] == "X" {
                    continue // self
                }
                link := Link{
                    SourceGPU: gpuIdx,
                    TargetGPU: j - 1,
                    Type:      c.parseLinkType(parts[j]),
                    Status:    LinkStatusOK,
                }
                topo.Links = append(topo.Links, link)
            }
        }
    }

    return topo, nil
}

func (c *Collector) parseLinkType(s string) LinkType {
    if strings.HasPrefix(s, "NV") {
        return LinkTypeNVLink
    }
    if strings.Contains(s, "PIX") || strings.Contains(s, "PXB") {
        return LinkTypePCIe
    }
    return LinkTypePCIe
}

func (c *Collector) parseGPUIndex(s string) int {
    // Parse "GPU0" -> 0
    var idx int
    fmt.Sscanf(s, "GPU%d", &idx)
    return idx
}

func (c *Collector) parseNVLinkErrors(output string) []LinkError {
    var errors []LinkError
    // Parse nvidia-smi nvlink --status output for CRC/recovery errors
    return errors
}

func (c *Collector) updateLinkStatus(topo *Topology) {
    // Update link status based on errors
    for i := range topo.Links {
        for _, err := range topo.Errors {
            if err.SourceGPU == topo.Links[i].SourceGPU &&
                err.TargetGPU == topo.Links[i].TargetGPU {
                if err.CRCErrors > 100 || err.RecoveryErrs > 10 {
                    topo.Links[i].Status = LinkStatusDegraded
                }
            }
        }
    }
}
```

### 1.3 ASCII Topology Renderer

**File:** `internal/nvlink/renderer.go`

```go
package nvlink

import (
    "fmt"
    "strings"

    "github.com/fatih/color"
)

type Renderer struct {
    green  func(...interface{}) string
    yellow func(...interface{}) string
    red    func(...interface{}) string
}

func NewRenderer() *Renderer {
    return &Renderer{
        green:  color.New(color.FgGreen).SprintFunc(),
        yellow: color.New(color.FgYellow).SprintFunc(),
        red:    color.New(color.FgRed).SprintFunc(),
    }
}

// Render outputs ASCII art topology for 8-GPU systems (2x4 grid)
func (r *Renderer) Render(topo *Topology) string {
    var sb strings.Builder

    sb.WriteString(fmt.Sprintf("GPU Topology (%s)\n", topo.Node))

    gpuCount := len(topo.GPUs)

    if gpuCount == 8 {
        r.render8GPU(&sb, topo)
    } else if gpuCount == 4 {
        r.render4GPU(&sb, topo)
    } else {
        r.renderGeneric(&sb, topo)
    }

    // Legend
    sb.WriteString("\nLegend: ")
    sb.WriteString(r.green("═") + " NVLink OK, ")
    sb.WriteString(r.yellow("⚠") + " Degraded, ")
    sb.WriteString(r.red("✗") + " Down\n")

    // Issues
    if len(topo.Errors) > 0 {
        sb.WriteString("\nIssues:\n")
        for _, err := range topo.Errors {
            sb.WriteString(fmt.Sprintf("  - GPU %d ↔ GPU %d: ",
                err.SourceGPU, err.TargetGPU))
            if err.CRCErrors > 0 {
                sb.WriteString(fmt.Sprintf("CRC errors: %d ", err.CRCErrors))
            }
            if err.RecoveryErrs > 0 {
                sb.WriteString(fmt.Sprintf("Recovery errors: %d", err.RecoveryErrs))
            }
            sb.WriteString("\n")
        }
    }

    return sb.String()
}

func (r *Renderer) render8GPU(sb *strings.Builder, topo *Topology) {
    // 8 GPU layout: 2 rows of 4
    //
    // ┌─────┐     ┌─────┐     ┌─────┐     ┌─────┐
    // │GPU 0│═════│GPU 1│═════│GPU 2│═════│GPU 3│
    // └──┬──┘     └──┬──┘     └──┬──┘     └──┬──┘
    //    ║           ║           ║           ║
    // ┌──┴──┐     ┌──┴──┐     ┌──┴──┐     ┌──┴──┐
    // │GPU 4│═════│GPU 5│═════│GPU 6│═════│GPU 7│
    // └─────┘     └─────┘     └─────┘     └─────┘

    // Row 1
    sb.WriteString("┌─────┐     ┌─────┐     ┌─────┐     ┌─────┐\n")
    sb.WriteString("│GPU 0│")
    sb.WriteString(r.linkChar(topo, 0, 1))
    sb.WriteString("│GPU 1│")
    sb.WriteString(r.linkChar(topo, 1, 2))
    sb.WriteString("│GPU 2│")
    sb.WriteString(r.linkChar(topo, 2, 3))
    sb.WriteString("│GPU 3│\n")
    sb.WriteString("└──┬──┘     └──┬──┘     └──┬──┘     └──┬──┘\n")

    // Vertical links
    sb.WriteString("   ")
    sb.WriteString(r.verticalLinkChar(topo, 0, 4))
    sb.WriteString("           ")
    sb.WriteString(r.verticalLinkChar(topo, 1, 5))
    sb.WriteString("           ")
    sb.WriteString(r.verticalLinkChar(topo, 2, 6))
    sb.WriteString("           ")
    sb.WriteString(r.verticalLinkChar(topo, 3, 7))
    sb.WriteString("\n")

    // Row 2
    sb.WriteString("┌──┴──┐     ┌──┴──┐     ┌──┴──┐     ┌──┴──┐\n")
    sb.WriteString("│GPU 4│")
    sb.WriteString(r.linkChar(topo, 4, 5))
    sb.WriteString("│GPU 5│")
    sb.WriteString(r.linkChar(topo, 5, 6))
    sb.WriteString("│GPU 6│")
    sb.WriteString(r.linkChar(topo, 6, 7))
    sb.WriteString("│GPU 7│\n")
    sb.WriteString("└─────┘     └─────┘     └─────┘     └─────┘\n")
}

func (r *Renderer) linkChar(topo *Topology, src, dst int) string {
    link := r.findLink(topo, src, dst)
    if link == nil {
        return "─────"
    }

    switch link.Status {
    case LinkStatusOK:
        return r.green("═════")
    case LinkStatusDegraded:
        return r.yellow("═⚠══")
    case LinkStatusDown:
        return r.red("──✗──")
    }
    return "─────"
}

func (r *Renderer) verticalLinkChar(topo *Topology, src, dst int) string {
    link := r.findLink(topo, src, dst)
    if link == nil {
        return "│"
    }

    switch link.Status {
    case LinkStatusOK:
        return r.green("║")
    case LinkStatusDegraded:
        return r.yellow("⚠")
    case LinkStatusDown:
        return r.red("✗")
    }
    return "│"
}

func (r *Renderer) findLink(topo *Topology, src, dst int) *Link {
    for i := range topo.Links {
        if (topo.Links[i].SourceGPU == src && topo.Links[i].TargetGPU == dst) ||
            (topo.Links[i].SourceGPU == dst && topo.Links[i].TargetGPU == src) {
            return &topo.Links[i]
        }
    }
    return nil
}

func (r *Renderer) render4GPU(sb *strings.Builder, topo *Topology) {
    // 4 GPU layout
    sb.WriteString("┌─────┐     ┌─────┐\n")
    sb.WriteString("│GPU 0│")
    sb.WriteString(r.linkChar(topo, 0, 1))
    sb.WriteString("│GPU 1│\n")
    sb.WriteString("└──┬──┘     └──┬──┘\n")
    sb.WriteString("   ")
    sb.WriteString(r.verticalLinkChar(topo, 0, 2))
    sb.WriteString("           ")
    sb.WriteString(r.verticalLinkChar(topo, 1, 3))
    sb.WriteString("\n")
    sb.WriteString("┌──┴──┐     ┌──┴──┐\n")
    sb.WriteString("│GPU 2│")
    sb.WriteString(r.linkChar(topo, 2, 3))
    sb.WriteString("│GPU 3│\n")
    sb.WriteString("└─────┘     └─────┘\n")
}

func (r *Renderer) renderGeneric(sb *strings.Builder, topo *Topology) {
    sb.WriteString(fmt.Sprintf("GPUs: %d\n", len(topo.GPUs)))
    sb.WriteString("Links:\n")
    for _, link := range topo.Links {
        status := r.green("OK")
        if link.Status == LinkStatusDegraded {
            status = r.yellow("DEGRADED")
        } else if link.Status == LinkStatusDown {
            status = r.red("DOWN")
        }
        sb.WriteString(fmt.Sprintf("  GPU %d ↔ GPU %d: %s [%s]\n",
            link.SourceGPU, link.TargetGPU, link.Type, status))
    }
}
```

### 1.4 Topology CLI Command

**File:** `internal/cli/topology.go`

```go
package cli

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/nvlink"
    "github.com/fregataa/aami/internal/ssh"
)

var topologyCmd = &cobra.Command{
    Use:   "topology [node]",
    Short: "Show NVLink topology of a node",
    Args:  cobra.ExactArgs(1),
    RunE:  runTopology,
}

var topologyAllCmd = &cobra.Command{
    Use:   "all",
    Short: "Show topology for all nodes",
    RunE:  runTopologyAll,
}

func init() {
    topologyCmd.AddCommand(topologyAllCmd)
    rootCmd.AddCommand(topologyCmd)
}

func runTopology(cmd *cobra.Command, args []string) error {
    nodeName := args[0]

    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    // Find node
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
        CommandTimeout: 60 * time.Second,
        MaxRetries:     3,
    })

    collector := nvlink.NewCollector(executor)
    topo, err := collector.CollectTopology(context.Background(), *node)
    if err != nil {
        return fmt.Errorf("collect topology: %w", err)
    }

    renderer := nvlink.NewRenderer()
    fmt.Println(renderer.Render(topo))

    return nil
}

func runTopologyAll(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    executor := ssh.NewExecutor(ssh.ExecutorConfig{
        ConnectTimeout: 10 * time.Second,
        CommandTimeout: 60 * time.Second,
        MaxRetries:     3,
    })
    collector := nvlink.NewCollector(executor)
    renderer := nvlink.NewRenderer()

    for _, n := range cfg.Nodes {
        node := ssh.Node{
            Name:    n.Name,
            Host:    n.IP,
            Port:    n.SSHPort,
            User:    n.SSHUser,
            KeyPath: n.SSHKey,
        }

        topo, err := collector.CollectTopology(context.Background(), node)
        if err != nil {
            fmt.Printf("⚠️  %s: %v\n\n", n.Name, err)
            continue
        }

        fmt.Println(renderer.Render(topo))
        fmt.Println()
    }

    return nil
}
```

### 1.5 NVLink Alert Rules

**File:** `configs/prometheus/nvlink-alerts.yaml`

```yaml
groups:
  - name: nvlink_alerts
    rules:
      - alert: NVLinkCRCErrors
        expr: increase(DCGM_FI_DEV_NVLINK_CRC_ERRORS[1h]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "NVLink CRC errors increasing on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} NVLink CRC errors: {{ $value }}"

      - alert: NVLinkRecoveryErrors
        expr: increase(DCGM_FI_DEV_NVLINK_RECOVERY_ERRORS[1h]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "NVLink recovery errors on {{ $labels.instance }}"

      - alert: NVLinkBandwidthDegraded
        expr: |
          DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL <
          (DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL offset 1d) * 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "NVLink bandwidth degraded on {{ $labels.instance }}"
```

---

## Task 2: GPU Health Score

### 2.1 Health Score Types

**File:** `internal/health/types.go`

```go
package health

import "time"

type HealthScore struct {
    Node        string
    Overall     float64    // 0-100
    Components  map[string]ComponentScore
    Status      HealthStatus
    LastUpdated time.Time
    TopIssues   []Issue
}

type ComponentScore struct {
    Name    string
    Score   float64
    Weight  float64
    Details string
}

type HealthStatus string

const (
    StatusHealthy  HealthStatus = "healthy"   // 90-100
    StatusWarning  HealthStatus = "warning"   // 70-89
    StatusDegraded HealthStatus = "degraded"  // 50-69
    StatusCritical HealthStatus = "critical"  // 0-49
)

type Issue struct {
    Severity    string  // critical, warning
    Description string
    Metric      string
    Value       float64
}

type ClusterHealth struct {
    Overall      float64
    Status       HealthStatus
    TotalNodes   int
    HealthyNodes int
    WarningNodes int
    CriticalNodes int
    NodeScores   []HealthScore
}

type WeightConfig struct {
    Temperature  float64 `yaml:"temperature"`   // default: 0.20
    ECCErrors    float64 `yaml:"ecc_errors"`    // default: 0.25
    XidErrors    float64 `yaml:"xid_errors"`    // default: 0.25
    NVLinkStatus float64 `yaml:"nvlink_status"` // default: 0.15
    Uptime       float64 `yaml:"uptime"`        // default: 0.15
}

func DefaultWeights() WeightConfig {
    return WeightConfig{
        Temperature:  0.20,
        ECCErrors:    0.25,
        XidErrors:    0.25,
        NVLinkStatus: 0.15,
        Uptime:       0.15,
    }
}
```

### 2.2 Prometheus Queries

**File:** `internal/health/prometheus.go`

```go
package health

import (
    "context"
    "fmt"
    "time"

    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

type PrometheusClient struct {
    api v1.API
}

func NewPrometheusClient(url string) (*PrometheusClient, error) {
    client, err := api.NewClient(api.Config{Address: url})
    if err != nil {
        return nil, err
    }
    return &PrometheusClient{api: v1.NewAPI(client)}, nil
}

// GetNodeMetrics retrieves health-related metrics for a node
func (p *PrometheusClient) GetNodeMetrics(ctx context.Context, instance string) (*NodeMetrics, error) {
    metrics := &NodeMetrics{Instance: instance}

    // Temperature (max GPU temp on the node)
    temp, err := p.queryScalar(ctx, fmt.Sprintf(
        `max(DCGM_FI_DEV_GPU_TEMP{instance="%s"})`, instance))
    if err == nil {
        metrics.MaxTemperature = temp
    }

    // ECC errors (sum over 24h)
    eccQuery := fmt.Sprintf(
        `sum(increase(DCGM_FI_DEV_ECC_DBE_VOL_TOTAL{instance="%s"}[24h]))`, instance)
    ecc, err := p.queryScalar(ctx, eccQuery)
    if err == nil {
        metrics.ECCErrors24h = int64(ecc)
    }

    // Xid errors (count in last 7 days)
    xidQuery := fmt.Sprintf(
        `count_over_time(DCGM_FI_DEV_XID_ERRORS{instance="%s"}[7d])`, instance)
    xid, err := p.queryScalar(ctx, xidQuery)
    if err == nil {
        metrics.XidErrors7d = int64(xid)
    }

    // NVLink errors
    nvlinkQuery := fmt.Sprintf(
        `sum(increase(DCGM_FI_DEV_NVLINK_CRC_ERRORS{instance="%s"}[24h]))`, instance)
    nvlink, err := p.queryScalar(ctx, nvlinkQuery)
    if err == nil {
        metrics.NVLinkErrors = int64(nvlink)
    }

    // Uptime (percentage in last 30 days)
    uptimeQuery := fmt.Sprintf(
        `avg_over_time(up{instance="%s",job="node"}[30d]) * 100`, instance)
    uptime, err := p.queryScalar(ctx, uptimeQuery)
    if err == nil {
        metrics.UptimePercent = uptime
    }

    return metrics, nil
}

func (p *PrometheusClient) queryScalar(ctx context.Context, query string) (float64, error) {
    result, _, err := p.api.Query(ctx, query, time.Now())
    if err != nil {
        return 0, err
    }

    switch v := result.(type) {
    case model.Vector:
        if len(v) > 0 {
            return float64(v[0].Value), nil
        }
    case *model.Scalar:
        return float64(v.Value), nil
    }

    return 0, fmt.Errorf("no data")
}

type NodeMetrics struct {
    Instance       string
    MaxTemperature float64
    ECCErrors24h   int64
    XidErrors7d    int64
    NVLinkErrors   int64
    UptimePercent  float64
}
```

### 2.3 Health Score Calculator

**File:** `internal/health/calculator.go`

```go
package health

import (
    "context"
    "sort"
    "time"
)

type Calculator struct {
    prometheus *PrometheusClient
    weights    WeightConfig
}

func NewCalculator(prometheusURL string, weights WeightConfig) (*Calculator, error) {
    client, err := NewPrometheusClient(prometheusURL)
    if err != nil {
        return nil, err
    }
    return &Calculator{prometheus: client, weights: weights}, nil
}

func (c *Calculator) CalculateNodeHealth(ctx context.Context, instance string) (*HealthScore, error) {
    metrics, err := c.prometheus.GetNodeMetrics(ctx, instance)
    if err != nil {
        return nil, err
    }

    score := &HealthScore{
        Node:        instance,
        Components:  make(map[string]ComponentScore),
        LastUpdated: time.Now(),
    }

    // Temperature score (100 if < 60, 0 if > 90)
    tempScore := c.calculateTemperatureScore(metrics.MaxTemperature)
    score.Components["temperature"] = ComponentScore{
        Name:    "Temperature",
        Score:   tempScore,
        Weight:  c.weights.Temperature,
        Details: fmt.Sprintf("Max: %.1f°C", metrics.MaxTemperature),
    }

    // ECC score (100 if 0 errors, decreases with more errors)
    eccScore := c.calculateECCScore(metrics.ECCErrors24h)
    score.Components["ecc"] = ComponentScore{
        Name:    "ECC Errors",
        Score:   eccScore,
        Weight:  c.weights.ECCErrors,
        Details: fmt.Sprintf("%d errors in 24h", metrics.ECCErrors24h),
    }

    // Xid score
    xidScore := c.calculateXidScore(metrics.XidErrors7d)
    score.Components["xid"] = ComponentScore{
        Name:    "Xid Errors",
        Score:   xidScore,
        Weight:  c.weights.XidErrors,
        Details: fmt.Sprintf("%d errors in 7d", metrics.XidErrors7d),
    }

    // NVLink score
    nvlinkScore := c.calculateNVLinkScore(metrics.NVLinkErrors)
    score.Components["nvlink"] = ComponentScore{
        Name:    "NVLink Status",
        Score:   nvlinkScore,
        Weight:  c.weights.NVLinkStatus,
        Details: fmt.Sprintf("%d CRC errors", metrics.NVLinkErrors),
    }

    // Uptime score
    uptimeScore := metrics.UptimePercent
    score.Components["uptime"] = ComponentScore{
        Name:    "Uptime",
        Score:   uptimeScore,
        Weight:  c.weights.Uptime,
        Details: fmt.Sprintf("%.1f%% in 30d", metrics.UptimePercent),
    }

    // Calculate weighted average
    var totalScore float64
    for _, comp := range score.Components {
        totalScore += comp.Score * comp.Weight
    }
    score.Overall = totalScore

    // Determine status
    score.Status = c.determineStatus(score.Overall)

    // Identify top issues
    score.TopIssues = c.identifyIssues(metrics, score)

    return score, nil
}

func (c *Calculator) calculateTemperatureScore(temp float64) float64 {
    if temp < 60 {
        return 100
    }
    if temp > 90 {
        return 0
    }
    // Linear interpolation between 60-90
    return 100 - ((temp - 60) / 30 * 100)
}

func (c *Calculator) calculateECCScore(errors int64) float64 {
    if errors == 0 {
        return 100
    }
    if errors > 1000 {
        return 0
    }
    // Logarithmic decay
    return 100 - (float64(errors) / 10)
}

func (c *Calculator) calculateXidScore(errors int64) float64 {
    if errors == 0 {
        return 100
    }
    if errors >= 5 {
        return 0
    }
    // Each Xid costs 25 points
    return 100 - float64(errors)*25
}

func (c *Calculator) calculateNVLinkScore(errors int64) float64 {
    if errors == 0 {
        return 100
    }
    if errors > 500 {
        return 0
    }
    return 100 - (float64(errors) / 5)
}

func (c *Calculator) determineStatus(score float64) HealthStatus {
    switch {
    case score >= 90:
        return StatusHealthy
    case score >= 70:
        return StatusWarning
    case score >= 50:
        return StatusDegraded
    default:
        return StatusCritical
    }
}

func (c *Calculator) identifyIssues(metrics *NodeMetrics, score *HealthScore) []Issue {
    var issues []Issue

    if metrics.MaxTemperature > 80 {
        issues = append(issues, Issue{
            Severity:    "warning",
            Description: fmt.Sprintf("High temperature: %.1f°C", metrics.MaxTemperature),
            Metric:      "temperature",
            Value:       metrics.MaxTemperature,
        })
    }

    if metrics.XidErrors7d > 0 {
        issues = append(issues, Issue{
            Severity:    "critical",
            Description: fmt.Sprintf("Xid errors detected: %d in 7d", metrics.XidErrors7d),
            Metric:      "xid_errors",
            Value:       float64(metrics.XidErrors7d),
        })
    }

    if metrics.ECCErrors24h > 100 {
        issues = append(issues, Issue{
            Severity:    "warning",
            Description: fmt.Sprintf("ECC errors: %d in 24h", metrics.ECCErrors24h),
            Metric:      "ecc_errors",
            Value:       float64(metrics.ECCErrors24h),
        })
    }

    return issues
}

func (c *Calculator) CalculateClusterHealth(ctx context.Context, instances []string) (*ClusterHealth, error) {
    cluster := &ClusterHealth{
        TotalNodes: len(instances),
        NodeScores: make([]HealthScore, 0, len(instances)),
    }

    var totalScore float64
    for _, instance := range instances {
        score, err := c.CalculateNodeHealth(ctx, instance)
        if err != nil {
            continue
        }

        cluster.NodeScores = append(cluster.NodeScores, *score)
        totalScore += score.Overall

        switch score.Status {
        case StatusHealthy:
            cluster.HealthyNodes++
        case StatusWarning, StatusDegraded:
            cluster.WarningNodes++
        case StatusCritical:
            cluster.CriticalNodes++
        }
    }

    if len(cluster.NodeScores) > 0 {
        cluster.Overall = totalScore / float64(len(cluster.NodeScores))
    }
    cluster.Status = c.determineStatus(cluster.Overall)

    // Sort by score (lowest first)
    sort.Slice(cluster.NodeScores, func(i, j int) bool {
        return cluster.NodeScores[i].Overall < cluster.NodeScores[j].Overall
    })

    return cluster, nil
}
```

### 2.4 Health CLI Command

**File:** `internal/cli/health.go`

```go
package cli

import (
    "context"
    "fmt"
    "os"

    "github.com/fatih/color"
    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/health"
)

var healthCmd = &cobra.Command{
    Use:   "health [node]",
    Short: "Show GPU health scores",
    RunE:  runHealth,
}

func init() {
    rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    prometheusURL := fmt.Sprintf("http://localhost:%d", cfg.Prometheus.Port)
    calculator, err := health.NewCalculator(prometheusURL, health.DefaultWeights())
    if err != nil {
        return err
    }

    ctx := context.Background()

    if len(args) > 0 {
        // Single node health
        return showNodeHealth(ctx, calculator, args[0])
    }

    // Cluster health
    return showClusterHealth(ctx, calculator, cfg)
}

func showClusterHealth(ctx context.Context, calc *health.Calculator, cfg *config.Config) error {
    var instances []string
    for _, node := range cfg.Nodes {
        instances = append(instances, node.IP)
    }

    cluster, err := calc.CalculateClusterHealth(ctx, instances)
    if err != nil {
        return err
    }

    fmt.Println("Cluster Health Summary")
    fmt.Println(strings.Repeat("━", 55))

    statusColor := getStatusColor(cluster.Status)
    fmt.Printf("Overall: %s (%.0f/100)\n",
        statusColor(string(cluster.Status)), cluster.Overall)
    fmt.Printf("Nodes: %d total, %d healthy, %d warning, %d critical\n\n",
        cluster.TotalNodes, cluster.HealthyNodes,
        cluster.WarningNodes, cluster.CriticalNodes)

    // Show top issues (worst nodes)
    if len(cluster.NodeScores) > 0 && cluster.NodeScores[0].Overall < 90 {
        fmt.Println("Top Issues:")

        table := tablewriter.NewWriter(os.Stdout)
        table.SetHeader([]string{"Node", "Score", "Reason"})
        table.SetBorder(true)

        for i := 0; i < min(5, len(cluster.NodeScores)); i++ {
            node := cluster.NodeScores[i]
            if node.Overall >= 90 {
                break
            }

            reason := ""
            if len(node.TopIssues) > 0 {
                reason = node.TopIssues[0].Description
            }

            table.Append([]string{
                node.Node,
                fmt.Sprintf("%.0f", node.Overall),
                reason,
            })
        }

        table.Render()

        // Recommendations
        fmt.Println("\nRecommendation:")
        for i := 0; i < min(3, len(cluster.NodeScores)); i++ {
            node := cluster.NodeScores[i]
            if node.Status == health.StatusCritical {
                fmt.Printf("  - %s: Schedule maintenance, consider GPU replacement\n", node.Node)
            }
        }
    }

    return nil
}

func showNodeHealth(ctx context.Context, calc *health.Calculator, nodeName string) error {
    score, err := calc.CalculateNodeHealth(ctx, nodeName)
    if err != nil {
        return err
    }

    fmt.Printf("Health Score: %s\n", nodeName)
    fmt.Println(strings.Repeat("━", 55))

    statusColor := getStatusColor(score.Status)
    fmt.Printf("Overall: %s (%.0f/100)\n\n", statusColor(string(score.Status)), score.Overall)

    fmt.Println("Components:")
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Component", "Score", "Weight", "Details"})

    for _, comp := range score.Components {
        table.Append([]string{
            comp.Name,
            fmt.Sprintf("%.0f", comp.Score),
            fmt.Sprintf("%.0f%%", comp.Weight*100),
            comp.Details,
        })
    }
    table.Render()

    if len(score.TopIssues) > 0 {
        fmt.Println("\nIssues:")
        for _, issue := range score.TopIssues {
            icon := "⚠️"
            if issue.Severity == "critical" {
                icon = "❌"
            }
            fmt.Printf("  %s %s\n", icon, issue.Description)
        }
    }

    return nil
}

func getStatusColor(status health.HealthStatus) func(...interface{}) string {
    switch status {
    case health.StatusHealthy:
        return color.New(color.FgGreen).SprintFunc()
    case health.StatusWarning:
        return color.New(color.FgYellow).SprintFunc()
    case health.StatusDegraded:
        return color.New(color.FgHiYellow).SprintFunc()
    case health.StatusCritical:
        return color.New(color.FgRed).SprintFunc()
    }
    return fmt.Sprint
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

---

## Task 3: Upgrade and Backup

### 3.1 Version Checker

**File:** `internal/upgrade/checker.go`

```go
package upgrade

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type VersionInfo struct {
    Component string
    Current   string
    Latest    string
    UpdateAvailable bool
}

type Checker struct {
    httpClient *http.Client
}

func NewChecker() *Checker {
    return &Checker{
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

type GitHubRelease struct {
    TagName string `json:"tag_name"`
}

func (c *Checker) CheckVersion(component, currentVersion string) (*VersionInfo, error) {
    info := &VersionInfo{
        Component: component,
        Current:   currentVersion,
    }

    latest, err := c.getLatestVersion(component)
    if err != nil {
        return nil, err
    }

    info.Latest = latest
    info.UpdateAvailable = latest != currentVersion

    return info, nil
}

func (c *Checker) getLatestVersion(component string) (string, error) {
    repos := map[string]string{
        "prometheus":   "prometheus/prometheus",
        "alertmanager": "prometheus/alertmanager",
        "grafana":      "grafana/grafana",
    }

    repo, ok := repos[component]
    if !ok {
        return "", fmt.Errorf("unknown component: %s", component)
    }

    url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var release GitHubRelease
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return "", err
    }

    return release.TagName, nil
}

func (c *Checker) CheckAll() ([]VersionInfo, error) {
    components := map[string]string{
        "prometheus":   "v2.48.0",  // Get actual versions from installed binaries
        "alertmanager": "v0.26.0",
        "grafana":      "v10.2.3",
    }

    var results []VersionInfo
    for comp, version := range components {
        info, err := c.CheckVersion(comp, version)
        if err != nil {
            continue
        }
        results = append(results, *info)
    }

    return results, nil
}
```

### 3.2 Upgrader

**File:** `internal/upgrade/upgrader.go`

```go
package upgrade

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

type Upgrader struct {
    backupDir string
}

func NewUpgrader(backupDir string) *Upgrader {
    return &Upgrader{backupDir: backupDir}
}

func (u *Upgrader) Upgrade(component, version string) error {
    // 1. Backup current version
    if err := u.backupCurrent(component); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }

    // 2. Stop service
    if err := u.stopService(component); err != nil {
        return fmt.Errorf("stop service: %w", err)
    }

    // 3. Download new version
    if err := u.download(component, version); err != nil {
        u.startService(component) // Try to restart old version
        return fmt.Errorf("download failed: %w", err)
    }

    // 4. Install
    if err := u.install(component); err != nil {
        u.Rollback(component)
        return fmt.Errorf("install failed: %w", err)
    }

    // 5. Start service
    if err := u.startService(component); err != nil {
        u.Rollback(component)
        return fmt.Errorf("start service: %w", err)
    }

    // 6. Health check
    if err := u.healthCheck(component); err != nil {
        u.Rollback(component)
        return fmt.Errorf("health check failed: %w", err)
    }

    return nil
}

func (u *Upgrader) Rollback(component string) error {
    backupPath := filepath.Join(u.backupDir, component+".backup")
    if _, err := os.Stat(backupPath); os.IsNotExist(err) {
        return fmt.Errorf("no backup found for %s", component)
    }

    u.stopService(component)

    // Restore from backup
    installPath := u.getInstallPath(component)
    if err := os.Rename(backupPath, installPath); err != nil {
        return err
    }

    return u.startService(component)
}

func (u *Upgrader) backupCurrent(component string) error {
    installPath := u.getInstallPath(component)
    backupPath := filepath.Join(u.backupDir, component+".backup")
    return os.Rename(installPath, backupPath)
}

func (u *Upgrader) stopService(component string) error {
    cmd := exec.Command("systemctl", "stop", "aami-"+component)
    return cmd.Run()
}

func (u *Upgrader) startService(component string) error {
    cmd := exec.Command("systemctl", "start", "aami-"+component)
    return cmd.Run()
}

func (u *Upgrader) download(component, version string) error {
    // Download from GitHub releases
    return nil
}

func (u *Upgrader) install(component string) error {
    // Extract and install binary
    return nil
}

func (u *Upgrader) healthCheck(component string) error {
    // Check if service is healthy
    return nil
}

func (u *Upgrader) getInstallPath(component string) string {
    return filepath.Join("/usr/local/bin", component)
}
```

### 3.3 Backup/Restore

**File:** `internal/backup/backup.go`

```go
package backup

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

type BackupConfig struct {
    OutputPath    string
    IncludeData   bool
    EncryptionKey string
}

type Backup struct {
    config BackupConfig
}

func NewBackup(cfg BackupConfig) *Backup {
    return &Backup{config: cfg}
}

// Create creates a backup of AAMI configuration
func (b *Backup) Create() error {
    outputPath := b.config.OutputPath
    if outputPath == "" {
        outputPath = fmt.Sprintf("aami-backup-%s.tar.gz",
            time.Now().Format("2006-01-02"))
    }

    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    gzWriter := gzip.NewWriter(file)
    defer gzWriter.Close()

    tarWriter := tar.NewWriter(gzWriter)
    defer tarWriter.Close()

    // Backup files
    filesToBackup := []string{
        "/etc/aami/config.yaml",
        "/etc/aami/rules",
        "/var/lib/aami/grafana/dashboards",
    }

    for _, path := range filesToBackup {
        if err := b.addToArchive(tarWriter, path); err != nil {
            fmt.Printf("Warning: could not backup %s: %v\n", path, err)
        }
    }

    // Optionally include Prometheus data
    if b.config.IncludeData {
        if err := b.addToArchive(tarWriter, "/var/lib/aami/prometheus/data"); err != nil {
            fmt.Printf("Warning: could not backup prometheus data: %v\n", err)
        }
    }

    fmt.Printf("Backup created: %s\n", outputPath)
    return nil
}

func (b *Backup) addToArchive(tw *tar.Writer, path string) error {
    return filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        header, err := tar.FileInfoHeader(fi, file)
        if err != nil {
            return err
        }

        header.Name = file

        if err := tw.WriteHeader(header); err != nil {
            return err
        }

        if !fi.IsDir() {
            data, err := os.Open(file)
            if err != nil {
                return err
            }
            defer data.Close()

            if _, err := io.Copy(tw, data); err != nil {
                return err
            }
        }

        return nil
    })
}
```

**File:** `internal/backup/restore.go`

```go
package backup

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

type RestoreConfig struct {
    BackupPath string
    ConfigOnly bool
}

func Restore(cfg RestoreConfig) error {
    file, err := os.Open(cfg.BackupPath)
    if err != nil {
        return err
    }
    defer file.Close()

    gzReader, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gzReader.Close()

    tarReader := tar.NewReader(gzReader)

    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        // Skip data if config-only restore
        if cfg.ConfigOnly && filepath.HasPrefix(header.Name, "/var/lib/aami/prometheus/data") {
            continue
        }

        target := header.Name

        switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(target, 0755); err != nil {
                return err
            }
        case tar.TypeReg:
            if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
                return err
            }

            outFile, err := os.Create(target)
            if err != nil {
                return err
            }

            if _, err := io.Copy(outFile, tarReader); err != nil {
                outFile.Close()
                return err
            }
            outFile.Close()

            if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
                return err
            }
        }
    }

    fmt.Println("Restore complete")
    return nil
}
```

### 3.4 Upgrade/Backup CLI Commands

**File:** `internal/cli/upgrade.go`

```go
package cli

import (
    "fmt"
    "os"

    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/upgrade"
)

var upgradeCmd = &cobra.Command{
    Use:   "upgrade",
    Short: "Upgrade AAMI components",
    RunE:  runUpgrade,
}

var (
    checkOnly bool
    rollback  bool
)

func init() {
    upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates")
    upgradeCmd.Flags().BoolVar(&rollback, "rollback", false, "Rollback to previous version")
    rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
    checker := upgrade.NewChecker()

    if checkOnly {
        versions, err := checker.CheckAll()
        if err != nil {
            return err
        }

        fmt.Println("Current versions:")
        table := tablewriter.NewWriter(os.Stdout)
        table.SetHeader([]string{"Component", "Current", "Latest", "Update"})

        for _, v := range versions {
            update := ""
            if v.UpdateAvailable {
                update = "available"
            } else {
                update = "(latest)"
            }
            table.Append([]string{v.Component, v.Current, v.Latest, update})
        }
        table.Render()
        return nil
    }

    upgrader := upgrade.NewUpgrader("/var/lib/aami/backup")

    if rollback {
        fmt.Println("Rolling back to previous version...")
        // Rollback logic
        return nil
    }

    // Upgrade all components with updates available
    versions, _ := checker.CheckAll()
    for _, v := range versions {
        if v.UpdateAvailable {
            fmt.Printf("Upgrading %s from %s to %s...\n",
                v.Component, v.Current, v.Latest)
            if err := upgrader.Upgrade(v.Component, v.Latest); err != nil {
                fmt.Printf("Failed to upgrade %s: %v\n", v.Component, err)
                continue
            }
            fmt.Printf("✅ %s upgraded\n", v.Component)
        }
    }

    return nil
}
```

**File:** `internal/cli/backup.go`

```go
package cli

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/backup"
)

var backupCmd = &cobra.Command{
    Use:   "backup",
    Short: "Manage backups",
}

var backupCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a backup",
    RunE:  runBackupCreate,
}

var backupRestoreCmd = &cobra.Command{
    Use:   "restore [file]",
    Short: "Restore from backup",
    Args:  cobra.ExactArgs(1),
    RunE:  runBackupRestore,
}

var (
    backupOutput      string
    backupIncludeData bool
    restoreConfigOnly bool
)

func init() {
    backupCreateCmd.Flags().StringVar(&backupOutput, "output", "",
        "Output file path")
    backupCreateCmd.Flags().BoolVar(&backupIncludeData, "include-data", false,
        "Include Prometheus data")

    backupRestoreCmd.Flags().BoolVar(&restoreConfigOnly, "config-only", false,
        "Restore configuration only")

    backupCmd.AddCommand(backupCreateCmd)
    backupCmd.AddCommand(backupRestoreCmd)
    rootCmd.AddCommand(backupCmd)
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
    b := backup.NewBackup(backup.BackupConfig{
        OutputPath:  backupOutput,
        IncludeData: backupIncludeData,
    })

    if err := b.Create(); err != nil {
        return err
    }

    fmt.Println("✅ Backup created")
    return nil
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
    if err := backup.Restore(backup.RestoreConfig{
        BackupPath: args[0],
        ConfigOnly: restoreConfigOnly,
    }); err != nil {
        return err
    }

    fmt.Println("✅ Restore complete")
    fmt.Println("Run 'aami apply' to reload services")
    return nil
}
```

---

## Task 4: Operations Tools

### 4.1 Diagnose Command

**File:** `internal/cli/diagnose.go`

```go
package cli

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/ssh"
)

var diagnoseCmd = &cobra.Command{
    Use:   "diagnose",
    Short: "Run system diagnostics",
    RunE:  runDiagnose,
}

func init() {
    rootCmd.AddCommand(diagnoseCmd)
}

type DiagnosticResult struct {
    Check   string
    Status  string
    Message string
}

func runDiagnose(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    fmt.Println("System Diagnosis")
    fmt.Println(strings.Repeat("━", 55))

    var results []DiagnosticResult

    // 1. Configuration validation
    validationErrors := cfg.Validate()
    if len(validationErrors) == 0 {
        results = append(results, DiagnosticResult{
            Check: "Configuration", Status: "ok", Message: "valid"})
    } else {
        results = append(results, DiagnosticResult{
            Check: "Configuration", Status: "error",
            Message: fmt.Sprintf("%d errors", len(validationErrors))})
    }

    // 2. Component health
    components := []struct {
        name string
        url  string
    }{
        {"Prometheus", fmt.Sprintf("http://localhost:%d/-/ready", cfg.Prometheus.Port)},
        {"Alertmanager", "http://localhost:9093/-/ready"},
        {"Grafana", fmt.Sprintf("http://localhost:%d/api/health", cfg.Grafana.Port)},
    }

    for _, c := range components {
        if checkEndpoint(c.url) {
            results = append(results, DiagnosticResult{
                Check: c.name, Status: "ok", Message: "running"})
        } else {
            results = append(results, DiagnosticResult{
                Check: c.name, Status: "error", Message: "not responding"})
        }
    }

    // 3. Node connectivity
    executor := ssh.NewExecutor(ssh.ExecutorConfig{
        ConnectTimeout: 5 * time.Second,
        CommandTimeout: 10 * time.Second,
        MaxRetries:     1,
    })

    reachable := 0
    unreachable := []string{}

    for _, n := range cfg.Nodes {
        node := ssh.Node{
            Name:    n.Name,
            Host:    n.IP,
            Port:    n.SSHPort,
            User:    n.SSHUser,
            KeyPath: n.SSHKey,
        }

        result := executor.Run(context.Background(), node, "echo ok")
        if result.Error == nil {
            reachable++
        } else {
            unreachable = append(unreachable, n.Name)
        }
    }

    if len(unreachable) == 0 {
        results = append(results, DiagnosticResult{
            Check: "Nodes", Status: "ok",
            Message: fmt.Sprintf("%d/%d reachable", reachable, len(cfg.Nodes))})
    } else {
        results = append(results, DiagnosticResult{
            Check: "Nodes", Status: "warning",
            Message: fmt.Sprintf("%d unreachable", len(unreachable))})
    }

    // Print results
    green := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    for _, r := range results {
        icon := green("✅")
        if r.Status == "warning" {
            icon = yellow("⚠️")
        } else if r.Status == "error" {
            icon = red("❌")
        }
        fmt.Printf("%s %s: %s\n", icon, r.Check, r.Message)
    }

    // Print issues with recommendations
    if len(unreachable) > 0 {
        fmt.Println("\nIssues Found:")
        for i, node := range unreachable {
            fmt.Printf("  %d. %s: SSH connection failed\n", i+1, node)
            fmt.Printf("     → Check if SSH service is running\n")
            fmt.Printf("     → Verify SSH key and user\n")
        }
    }

    return nil
}

func checkEndpoint(url string) bool {
    client := &http.Client{Timeout: 2 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}
```

### 4.2 Config Diff Command

**File:** `internal/cli/diff.go`

```go
package cli

import (
    "fmt"
    "strings"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/config"
)

var diffCmd = &cobra.Command{
    Use:   "diff",
    Short: "Show pending configuration changes",
    RunE:  runDiff,
}

var applyDryRun bool

func init() {
    rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load(config.DefaultConfigPath)
    if err != nil {
        return err
    }

    // Compare with running config
    changes := detectChanges(cfg)

    if len(changes) == 0 {
        fmt.Println("No changes detected")
        return nil
    }

    green := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    red := color.New(color.FgRed).SprintFunc()

    fmt.Println("Changes detected:")
    for _, c := range changes {
        var icon string
        switch c.Type {
        case "add":
            icon = green("[+]")
        case "modify":
            icon = yellow("[~]")
        case "remove":
            icon = red("[-]")
        }
        fmt.Printf("  %s %s: %s\n", icon, c.Path, c.Description)
    }

    return nil
}

type Change struct {
    Type        string // add, modify, remove
    Path        string
    Description string
}

func detectChanges(cfg *config.Config) []Change {
    var changes []Change

    // Compare current config with running state
    // This requires reading current Prometheus/Alertmanager configs
    // and comparing with what would be generated

    // Example changes:
    // - New nodes added
    // - Alert rules changed
    // - Notification settings changed

    return changes
}
```

---

## Test Commands

```bash
# Build
go build -o aami ./cmd/aami

# Unit tests
go test ./internal/nvlink/... -v
go test ./internal/health/... -v
go test ./internal/upgrade/... -v
go test ./internal/backup/... -v

# Integration tests
./aami topology gpu-node-01
./aami health
./aami health gpu-node-01
./aami upgrade --check
./aami backup create --output backup.tar.gz
./aami backup restore backup.tar.gz --config-only
./aami diagnose
./aami diff
```

---

## Acceptance Criteria

| Feature | Test Command | Expected Output |
|---------|--------------|-----------------|
| Topology | `aami topology gpu-node-01` | ASCII art with NVLink status |
| Topology issues | `aami topology gpu-node-01` (with errors) | Shows degraded links highlighted |
| Cluster health | `aami health` | Summary with overall score |
| Node health | `aami health gpu-node-01` | Detailed component scores |
| Upgrade check | `aami upgrade --check` | Table of versions |
| Backup create | `aami backup create` | `Backup created: aami-backup-YYYY-MM-DD.tar.gz` |
| Backup restore | `aami backup restore backup.tar.gz` | `Restore complete` |
| Diagnose | `aami diagnose` | Status of all components |
| Config diff | `aami diff` | List of pending changes |
