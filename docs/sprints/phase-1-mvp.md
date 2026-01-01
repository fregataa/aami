# Phase 1: MVP

## Overview

- **Duration**: 4-6 weeks
- **Goal**: Core functionality for 30-minute GPU cluster monitoring setup
- **Output**: Working `aami` CLI binary

## Project Structure

```
aami/
├── cmd/
│   └── aami/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/                     # CLI commands (Cobra)
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── nodes.go
│   │   ├── alerts.go
│   │   ├── status.go
│   │   └── explain.go
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Config struct and loader
│   │   ├── validator.go         # Validation logic
│   │   └── types.go             # Type definitions
│   ├── ssh/                     # SSH executor
│   │   ├── executor.go          # Main executor
│   │   ├── pool.go              # Connection pool
│   │   └── retry.go             # Retry logic
│   ├── prometheus/              # Prometheus config generator
│   │   ├── generator.go
│   │   └── templates/
│   ├── alertmanager/            # Alertmanager config generator
│   │   └── generator.go
│   ├── grafana/                 # Grafana provisioning
│   │   └── provisioner.go
│   ├── installer/               # Component installer
│   │   ├── installer.go
│   │   └── exporter.go
│   └── xid/                     # Xid error database
│       ├── database.go
│       └── data.go
├── pkg/                         # Public packages (if needed)
├── configs/                     # Default configs and templates
│   ├── prometheus/
│   │   └── gpu-alerts.yaml
│   ├── grafana/
│   │   └── dashboards/
│   └── defaults.yaml
├── scripts/
│   └── install-exporter.sh
├── go.mod
└── go.sum
```

## Dependencies

```go
// go.mod
module github.com/fregataa/aami

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    golang.org/x/crypto v0.18.0      // SSH
    gopkg.in/yaml.v3 v3.0.1
    github.com/rs/zerolog v1.31.0    // Logging
    github.com/olekukonko/tablewriter v0.0.5  // CLI tables
    github.com/fatih/color v1.16.0   // Colored output
)
```

---

## Task 1: Project Initialization

### 1.1 Create Go Module

**Files to create:**
- `go.mod`
- `cmd/aami/main.go`

**Implementation:**

```go
// cmd/aami/main.go
package main

import (
    "os"
    "github.com/fregataa/aami/internal/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**Test:**
```bash
go build -o aami ./cmd/aami
./aami --help
```

### 1.2 CLI Root Command

**File:** `internal/cli/root.go`

**Implementation:**

```go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "aami",
    Short: "AI Accelerator Monitoring Infrastructure",
    Long:  "GPU cluster monitoring tool with Prometheus stack",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
        "config file (default: /etc/aami/config.yaml)")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigFile("/etc/aami/config.yaml")
    }
    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

### 1.3 Version Command

**File:** `internal/cli/version.go`

**Variables:** Set via `-ldflags` at build time.

```go
package cli

var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("AAMI %s (commit: %s, built: %s)\n",
            Version, Commit, BuildDate)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

**Build:**
```bash
go build -ldflags "-X github.com/fregataa/aami/internal/cli.Version=v0.1.0 \
  -X github.com/fregataa/aami/internal/cli.Commit=$(git rev-parse --short HEAD)" \
  -o aami ./cmd/aami
```

---

## Task 2: Configuration Management

### 2.1 Config Types

**File:** `internal/config/types.go`

```go
package config

type Config struct {
    Cluster       ClusterConfig       `yaml:"cluster"`
    Nodes         []NodeConfig        `yaml:"nodes"`
    SSH           SSHConfig           `yaml:"ssh"`
    Alerts        AlertsConfig        `yaml:"alerts"`
    Notifications NotificationsConfig `yaml:"notifications"`
    Prometheus    PrometheusConfig    `yaml:"prometheus"`
    Grafana       GrafanaConfig       `yaml:"grafana"`
}

type ClusterConfig struct {
    Name string `yaml:"name"`
}

type NodeConfig struct {
    Name     string            `yaml:"name"`
    IP       string            `yaml:"ip"`
    SSHUser  string            `yaml:"ssh_user"`
    SSHKey   string            `yaml:"ssh_key"`
    SSHPort  int               `yaml:"ssh_port"`
    Labels   map[string]string `yaml:"labels"`
}

type SSHConfig struct {
    MaxParallel    int `yaml:"max_parallel"`     // default: 50
    ConnectTimeout int `yaml:"connect_timeout"`  // seconds, default: 10
    CommandTimeout int `yaml:"command_timeout"`  // seconds, default: 300
    Retry          RetryConfig `yaml:"retry"`
}

type RetryConfig struct {
    MaxAttempts int `yaml:"max_attempts"`  // default: 3
    BackoffBase int `yaml:"backoff_base"`  // seconds, default: 2
    BackoffMax  int `yaml:"backoff_max"`   // seconds, default: 30
}

type AlertsConfig struct {
    Presets []string          `yaml:"presets"`
    Custom  []CustomAlertRule `yaml:"custom"`
}

type CustomAlertRule struct {
    Name     string `yaml:"name"`
    Expr     string `yaml:"expr"`
    For      string `yaml:"for"`
    Severity string `yaml:"severity"`
}

type NotificationsConfig struct {
    Slack   *SlackConfig   `yaml:"slack"`
    Email   *EmailConfig   `yaml:"email"`
    Webhook *WebhookConfig `yaml:"webhook"`
}

type SlackConfig struct {
    Enabled    bool   `yaml:"enabled"`
    WebhookURL string `yaml:"webhook_url"`
    Channel    string `yaml:"channel"`
}

type EmailConfig struct {
    Enabled  bool     `yaml:"enabled"`
    SMTPHost string   `yaml:"smtp_host"`
    SMTPPort int      `yaml:"smtp_port"`
    From     string   `yaml:"from"`
    To       []string `yaml:"to"`
}

type WebhookConfig struct {
    Enabled bool   `yaml:"enabled"`
    URL     string `yaml:"url"`
}

type PrometheusConfig struct {
    Retention   string `yaml:"retention"`     // default: "15d"
    StoragePath string `yaml:"storage_path"`  // default: "/var/lib/aami/prometheus"
    Port        int    `yaml:"port"`          // default: 9090
}

type GrafanaConfig struct {
    Port          int    `yaml:"port"`           // default: 3000
    AdminPassword string `yaml:"admin_password"` // supports ${ENV_VAR}
}
```

### 2.2 Config Loader

**File:** `internal/config/config.go`

```go
package config

import (
    "os"
    "regexp"
    "gopkg.in/yaml.v3"
)

const DefaultConfigPath = "/etc/aami/config.yaml"

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

func expandEnvVars(content string) string {
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    return re.ReplaceAllStringFunc(content, func(match string) string {
        varName := match[2 : len(match)-1]
        return os.Getenv(varName)
    })
}

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
    if cfg.Prometheus.Retention == "" {
        cfg.Prometheus.Retention = "15d"
    }
    if cfg.Prometheus.Port == 0 {
        cfg.Prometheus.Port = 9090
    }
    if cfg.Grafana.Port == 0 {
        cfg.Grafana.Port = 3000
    }
}
```

### 2.3 Config Validator

**File:** `internal/config/validator.go`

```go
package config

import (
    "fmt"
    "net"
    "os"
)

type ValidationError struct {
    Field   string
    Message string
}

func (c *Config) Validate() []ValidationError {
    var errors []ValidationError

    if c.Cluster.Name == "" {
        errors = append(errors, ValidationError{
            Field: "cluster.name", Message: "required"})
    }

    for i, node := range c.Nodes {
        if node.Name == "" {
            errors = append(errors, ValidationError{
                Field: fmt.Sprintf("nodes[%d].name", i), Message: "required"})
        }
        if net.ParseIP(node.IP) == nil {
            errors = append(errors, ValidationError{
                Field: fmt.Sprintf("nodes[%d].ip", i), Message: "invalid IP"})
        }
        if node.SSHKey != "" {
            if _, err := os.Stat(node.SSHKey); os.IsNotExist(err) {
                errors = append(errors, ValidationError{
                    Field: fmt.Sprintf("nodes[%d].ssh_key", i),
                    Message: "file not found"})
            }
        }
    }

    return errors
}
```

**Test:**
```bash
go test ./internal/config/...
```

---

## Task 3: SSH Executor

### 3.1 SSH Client

**File:** `internal/ssh/executor.go`

```go
package ssh

import (
    "context"
    "fmt"
    "os"
    "time"

    "golang.org/x/crypto/ssh"
)

type Executor struct {
    config ExecutorConfig
}

type ExecutorConfig struct {
    MaxParallel    int
    ConnectTimeout time.Duration
    CommandTimeout time.Duration
    MaxRetries     int
    BackoffBase    time.Duration
    BackoffMax     time.Duration
}

type Node struct {
    Name    string
    Host    string
    Port    int
    User    string
    KeyPath string
}

type Result struct {
    Node   string
    Output string
    Error  error
}

func NewExecutor(cfg ExecutorConfig) *Executor {
    return &Executor{config: cfg}
}

func (e *Executor) Run(ctx context.Context, node Node, command string) Result {
    client, err := e.connect(node)
    if err != nil {
        return Result{Node: node.Name, Error: err}
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return Result{Node: node.Name, Error: err}
    }
    defer session.Close()

    output, err := session.CombinedOutput(command)
    return Result{
        Node:   node.Name,
        Output: string(output),
        Error:  err,
    }
}

func (e *Executor) connect(node Node) (*ssh.Client, error) {
    key, err := os.ReadFile(node.KeyPath)
    if err != nil {
        return nil, fmt.Errorf("read key: %w", err)
    }

    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return nil, fmt.Errorf("parse key: %w", err)
    }

    config := &ssh.ClientConfig{
        User:            node.User,
        Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: proper host key
        Timeout:         e.config.ConnectTimeout,
    }

    addr := fmt.Sprintf("%s:%d", node.Host, node.Port)
    return ssh.Dial("tcp", addr, config)
}
```

### 3.2 Parallel Executor

**File:** `internal/ssh/pool.go`

```go
package ssh

import (
    "context"
    "sync"
)

func (e *Executor) RunParallel(ctx context.Context, nodes []Node, command string) []Result {
    results := make([]Result, len(nodes))
    sem := make(chan struct{}, e.config.MaxParallel)
    var wg sync.WaitGroup

    for i, node := range nodes {
        wg.Add(1)
        go func(idx int, n Node) {
            defer wg.Done()

            sem <- struct{}{}
            defer func() { <-sem }()

            results[idx] = e.RunWithRetry(ctx, n, command)
        }(i, node)
    }

    wg.Wait()
    return results
}
```

### 3.3 Retry Logic

**File:** `internal/ssh/retry.go`

```go
package ssh

import (
    "context"
    "time"
)

func (e *Executor) RunWithRetry(ctx context.Context, node Node, command string) Result {
    var lastResult Result

    for attempt := 0; attempt < e.config.MaxRetries; attempt++ {
        if attempt > 0 {
            backoff := e.calculateBackoff(attempt)
            select {
            case <-ctx.Done():
                return Result{Node: node.Name, Error: ctx.Err()}
            case <-time.After(backoff):
            }
        }

        lastResult = e.Run(ctx, node, command)
        if lastResult.Error == nil {
            return lastResult
        }

        // Don't retry auth failures
        if isAuthError(lastResult.Error) {
            return lastResult
        }
    }

    return lastResult
}

func (e *Executor) calculateBackoff(attempt int) time.Duration {
    backoff := e.config.BackoffBase * time.Duration(1<<uint(attempt))
    if backoff > e.config.BackoffMax {
        backoff = e.config.BackoffMax
    }
    return backoff
}

func isAuthError(err error) bool {
    // Check if error is authentication related
    return false // TODO: implement
}
```

---

## Task 4: CLI Commands

### 4.1 Init Command

**File:** `internal/cli/init.go`

```go
package cli

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize AAMI configuration",
    RunE:  runInit,
}

var offlineBundle string

func init() {
    initCmd.Flags().StringVar(&offlineBundle, "offline", "",
        "Path to offline bundle for air-gap installation")
    rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
    // 1. Create directories
    dirs := []string{
        "/etc/aami",
        "/var/lib/aami/prometheus",
        "/var/lib/aami/grafana",
        "/var/lib/aami/targets",
    }
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("create directory %s: %w", dir, err)
        }
    }

    // 2. Create default config if not exists
    configPath := "/etc/aami/config.yaml"
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        if err := createDefaultConfig(configPath); err != nil {
            return err
        }
        fmt.Printf("Created %s\n", configPath)
    }

    // 3. Install components
    if offlineBundle != "" {
        return installOffline(offlineBundle)
    }
    return installOnline()
}

func createDefaultConfig(path string) error {
    defaultConfig := `cluster:
  name: my-gpu-cluster

nodes: []

ssh:
  max_parallel: 50
  connect_timeout: 10
  command_timeout: 300

alerts:
  presets:
    - gpu-production

notifications:
  slack:
    enabled: false
    webhook_url: ""

prometheus:
  retention: 15d
  storage_path: /var/lib/aami/prometheus
`
    return os.WriteFile(path, []byte(defaultConfig), 0644)
}
```

### 4.2 Nodes Command

**File:** `internal/cli/nodes.go`

```go
package cli

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
)

var nodesCmd = &cobra.Command{
    Use:   "nodes",
    Short: "Manage GPU nodes",
}

var nodesAddCmd = &cobra.Command{
    Use:   "add [name]",
    Short: "Add a node to the cluster",
    RunE:  runNodesAdd,
}

var nodesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all nodes",
    RunE:  runNodesList,
}

var nodesInstallCmd = &cobra.Command{
    Use:   "install [name]",
    Short: "Install exporters on node(s)",
    RunE:  runNodesInstall,
}

var (
    nodeIP     string
    nodeUser   string
    nodeKey    string
    nodeLabels string
    nodesFile  string
    installAll bool
)

func init() {
    nodesAddCmd.Flags().StringVar(&nodeIP, "ip", "", "Node IP address")
    nodesAddCmd.Flags().StringVar(&nodeUser, "user", "root", "SSH user")
    nodesAddCmd.Flags().StringVar(&nodeKey, "key", "", "SSH key path")
    nodesAddCmd.Flags().StringVar(&nodeLabels, "labels", "", "Labels (k=v,k2=v2)")
    nodesAddCmd.Flags().StringVar(&nodesFile, "file", "", "File with nodes list")

    nodesInstallCmd.Flags().BoolVar(&installAll, "all", false, "Install on all nodes")

    nodesCmd.AddCommand(nodesAddCmd)
    nodesCmd.AddCommand(nodesListCmd)
    nodesCmd.AddCommand(nodesInstallCmd)
    rootCmd.AddCommand(nodesCmd)
}

func runNodesAdd(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    if nodesFile != "" {
        return addNodesFromFile(cfg, nodesFile)
    }

    if len(args) == 0 {
        return fmt.Errorf("node name required")
    }

    node := config.NodeConfig{
        Name:    args[0],
        IP:      nodeIP,
        SSHUser: nodeUser,
        SSHKey:  nodeKey,
        SSHPort: 22,
        Labels:  parseLabels(nodeLabels),
    }

    cfg.Nodes = append(cfg.Nodes, node)
    if err := saveConfig(cfg); err != nil {
        return err
    }

    fmt.Printf("✅ Node %s added\n", node.Name)
    return nil
}

func runNodesList(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Name", "IP", "User", "Labels"})

    for _, node := range cfg.Nodes {
        table.Append([]string{
            node.Name,
            node.IP,
            node.SSHUser,
            formatLabels(node.Labels),
        })
    }

    table.Render()
    return nil
}

func parseLabels(s string) map[string]string {
    labels := make(map[string]string)
    if s == "" {
        return labels
    }
    for _, pair := range strings.Split(s, ",") {
        parts := strings.SplitN(pair, "=", 2)
        if len(parts) == 2 {
            labels[parts[0]] = parts[1]
        }
    }
    return labels
}
```

### 4.3 Status Command

**File:** `internal/cli/status.go`

```go
package cli

import (
    "fmt"
    "net/http"
    "time"

    "github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show cluster status",
    RunE:  runStatus,
}

func init() {
    rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    fmt.Printf("Cluster: %s\n", cfg.Cluster.Name)
    fmt.Printf("Nodes:   %d\n\n", len(cfg.Nodes))

    fmt.Println("Components:")
    checkComponent("Prometheus", fmt.Sprintf("http://localhost:%d/-/ready",
        cfg.Prometheus.Port))
    checkComponent("Alertmanager", "http://localhost:9093/-/ready")
    checkComponent("Grafana", fmt.Sprintf("http://localhost:%d/api/health",
        cfg.Grafana.Port))

    return nil
}

func checkComponent(name, url string) {
    client := &http.Client{Timeout: 2 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        fmt.Printf("  %s: ❌ not running\n", name)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == 200 {
        fmt.Printf("  %s: ✅ running\n", name)
    } else {
        fmt.Printf("  %s: ⚠️ unhealthy (%d)\n", name, resp.StatusCode)
    }
}
```

### 4.4 Alerts Command

**File:** `internal/cli/alerts.go`

```go
package cli

import (
    "fmt"

    "github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
    Use:   "alerts",
    Short: "Manage alert rules",
}

var alertsListPresetsCmd = &cobra.Command{
    Use:   "list-presets",
    Short: "List available alert presets",
    RunE:  runAlertsListPresets,
}

var alertsApplyPresetCmd = &cobra.Command{
    Use:   "apply-preset [name]",
    Short: "Apply an alert preset",
    Args:  cobra.ExactArgs(1),
    RunE:  runAlertsApplyPreset,
}

var presets = map[string][]string{
    "gpu-basic": {
        "gpu_temperature_critical",
        "gpu_memory_high",
        "node_down",
    },
    "gpu-production": {
        "gpu_temperature_critical",
        "gpu_temperature_warning",
        "gpu_memory_high",
        "gpu_memory_leak",
        "gpu_ecc_errors",
        "gpu_xid_error",
        "gpu_nvlink_error",
        "node_down",
    },
}

func init() {
    alertsCmd.AddCommand(alertsListPresetsCmd)
    alertsCmd.AddCommand(alertsApplyPresetCmd)
    rootCmd.AddCommand(alertsCmd)
}

func runAlertsListPresets(cmd *cobra.Command, args []string) error {
    fmt.Println("Available presets:")
    for name, rules := range presets {
        fmt.Printf("  %s (%d rules)\n", name, len(rules))
    }
    return nil
}

func runAlertsApplyPreset(cmd *cobra.Command, args []string) error {
    presetName := args[0]
    rules, ok := presets[presetName]
    if !ok {
        return fmt.Errorf("unknown preset: %s", presetName)
    }

    // TODO: Generate Prometheus rules file
    // TODO: Reload Prometheus

    fmt.Printf("✅ Applied preset %s (%d rules)\n", presetName, len(rules))
    return nil
}
```

### 4.5 Explain Xid Command

**File:** `internal/cli/explain.go`

```go
package cli

import (
    "fmt"
    "strconv"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/fregataa/aami/internal/xid"
)

var explainCmd = &cobra.Command{
    Use:   "explain",
    Short: "Explain error codes",
}

var explainXidCmd = &cobra.Command{
    Use:   "xid [code]",
    Short: "Explain NVIDIA Xid error code",
    Args:  cobra.ExactArgs(1),
    RunE:  runExplainXid,
}

func init() {
    explainCmd.AddCommand(explainXidCmd)
    rootCmd.AddCommand(explainCmd)
}

func runExplainXid(cmd *cobra.Command, args []string) error {
    code, err := strconv.Atoi(args[0])
    if err != nil {
        return fmt.Errorf("invalid Xid code: %s", args[0])
    }

    info, ok := xid.Database[code]
    if !ok {
        return fmt.Errorf("unknown Xid code: %d", code)
    }

    red := color.New(color.FgRed).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()

    fmt.Printf("Xid %d: %s\n\n", code, info.Name)

    severity := info.Severity
    if severity == "Critical" {
        severity = red(severity)
    } else if severity == "Warning" {
        severity = yellow(severity)
    }
    fmt.Printf("Severity: %s\n\n", severity)

    fmt.Printf("Meaning:\n  %s\n\n", info.Description)

    fmt.Println("Common Causes:")
    for i, cause := range info.Causes {
        fmt.Printf("  %d. %s\n", i+1, cause)
    }
    fmt.Println()

    fmt.Println("Recommended Actions:")
    for i, action := range info.Actions {
        fmt.Printf("  %d. %s\n", i+1, action)
    }
    fmt.Println()

    fmt.Printf("Reference:\n  %s\n", info.Reference)

    return nil
}
```

### 4.6 Xid Database

**File:** `internal/xid/data.go`

```go
package xid

type XidInfo struct {
    Name        string
    Severity    string   // "Critical", "Warning", "Info"
    Description string
    Causes      []string
    Actions     []string
    Reference   string
}

var Database = map[int]XidInfo{
    13: {
        Name:        "Graphics Engine Exception",
        Severity:    "Critical",
        Description: "GPU encountered an unrecoverable graphics engine error.",
        Causes: []string{
            "Driver bug",
            "GPU hardware failure",
            "Overclocking instability",
        },
        Actions: []string{
            "Check GPU temperature",
            "Update NVIDIA driver",
            "Run GPU diagnostics",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
    31: {
        Name:        "GPU memory page fault",
        Severity:    "Critical",
        Description: "GPU detected invalid memory access.",
        Causes: []string{
            "Application bug (illegal memory access)",
            "GPU memory corruption",
            "Driver issue",
        },
        Actions: []string{
            "Check application code for memory errors",
            "Run memtest on GPU",
            "Update driver",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
    48: {
        Name:        "Double Bit ECC Error",
        Severity:    "Critical",
        Description: "Uncorrectable ECC memory error detected.",
        Causes: []string{
            "GPU memory hardware failure",
            "Memory aging",
        },
        Actions: []string{
            "Drain workloads from node",
            "Schedule GPU replacement",
            "Check ECC error trends",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
    63: {
        Name:        "ECC page retirement: Row remapping failure",
        Severity:    "Warning",
        Description: "GPU retired memory pages due to ECC errors.",
        Causes: []string{
            "Memory cell wear",
            "Manufacturing defect",
        },
        Actions: []string{
            "Monitor ECC error trends",
            "Plan for GPU replacement if errors increase",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
    79: {
        Name:        "GPU has fallen off the bus",
        Severity:    "Critical",
        Description: "GPU disconnected from PCIe bus. System cannot communicate with GPU.",
        Causes: []string{
            "PCIe slot contact issue",
            "Power supply instability",
            "GPU hardware failure",
            "Thermal shutdown",
            "Driver/firmware bug",
        },
        Actions: []string{
            "Immediately drain node from workloads",
            "Check BMC/IPMI hardware event logs",
            "Reseat GPU",
            "If recurring, replace GPU",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
    94: {
        Name:        "Contained ECC error",
        Severity:    "Warning",
        Description: "ECC error was contained and corrected.",
        Causes: []string{
            "Memory cell degradation",
            "Cosmic ray bit flip",
        },
        Actions: []string{
            "Monitor for increasing frequency",
            "No immediate action needed if isolated",
        },
        Reference: "https://docs.nvidia.com/deploy/xid-errors/",
    },
}
```

---

## Task 5: Prometheus Stack Installation

### 5.1 Component Installer

**File:** `internal/installer/installer.go`

```go
package installer

import (
    "fmt"
    "os"
    "os/exec"
    "runtime"
)

type Component struct {
    Name    string
    Version string
    URL     string
    Binary  string
}

var Components = map[string]Component{
    "prometheus": {
        Name:    "prometheus",
        Version: "2.48.0",
        Binary:  "prometheus",
    },
    "alertmanager": {
        Name:    "alertmanager",
        Version: "0.26.0",
        Binary:  "alertmanager",
    },
    "grafana": {
        Name:    "grafana",
        Version: "10.2.3",
        Binary:  "grafana-server",
    },
    "node_exporter": {
        Name:    "node_exporter",
        Version: "1.7.0",
        Binary:  "node_exporter",
    },
}

func (c Component) DownloadURL() string {
    arch := runtime.GOARCH
    if arch == "amd64" {
        arch = "amd64"
    }
    // Return download URL based on component
    return fmt.Sprintf(
        "https://github.com/prometheus/%s/releases/download/v%s/%s-%s.linux-%s.tar.gz",
        c.Name, c.Version, c.Name, c.Version, arch)
}

func Install(name string, installDir string) error {
    component, ok := Components[name]
    if !ok {
        return fmt.Errorf("unknown component: %s", name)
    }

    // Download
    url := component.DownloadURL()
    fmt.Printf("Downloading %s from %s\n", name, url)

    // Extract
    // Create systemd service
    // Start service

    return nil
}
```

### 5.2 Prometheus Config Generator

**File:** `internal/prometheus/generator.go`

```go
package prometheus

import (
    "os"
    "text/template"

    "github.com/fregataa/aami/internal/config"
)

const prometheusConfigTemplate = `
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

rule_files:
  - '/etc/aami/rules/*.yaml'

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node'
    file_sd_configs:
      - files:
          - '/var/lib/aami/targets/nodes.json'

  - job_name: 'dcgm'
    file_sd_configs:
      - files:
          - '/var/lib/aami/targets/dcgm.json'
`

func GenerateConfig(cfg *config.Config, outputPath string) error {
    tmpl, err := template.New("prometheus").Parse(prometheusConfigTemplate)
    if err != nil {
        return err
    }

    f, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    return tmpl.Execute(f, cfg)
}

// GenerateTargets creates the file_sd JSON files
func GenerateTargets(nodes []config.NodeConfig, outputDir string) error {
    // Generate /var/lib/aami/targets/nodes.json
    // Generate /var/lib/aami/targets/dcgm.json
    return nil
}
```

---

## Task 6: Alert Rules

### 6.1 GPU Alert Rules Template

**File:** `configs/prometheus/gpu-alerts.yaml`

```yaml
groups:
  - name: gpu_alerts
    rules:
      - alert: GPUTemperatureCritical
        expr: DCGM_FI_DEV_GPU_TEMP > 85
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "GPU temperature critical on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} temperature is {{ $value }}°C"

      - alert: GPUTemperatureWarning
        expr: DCGM_FI_DEV_GPU_TEMP > 75
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "GPU temperature warning on {{ $labels.instance }}"

      - alert: GPUMemoryLeak
        expr: |
          DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100 > 95
          and DCGM_FI_DEV_GPU_UTIL < 5
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Possible GPU memory leak on {{ $labels.instance }}"

      - alert: GPUECCErrors
        expr: increase(DCGM_FI_DEV_ECC_DBE_VOL_TOTAL[24h]) > 100
        labels:
          severity: critical
        annotations:
          summary: "High ECC error count on {{ $labels.instance }}"

      - alert: GPUXidError
        expr: increase(DCGM_FI_DEV_XID_ERRORS[5m]) > 0
        labels:
          severity: critical
        annotations:
          summary: "Xid error on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} reported Xid error"

      - alert: NodeDown
        expr: up{job="node"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Node {{ $labels.instance }} is down"
```

---

## Task 7: Air-gap Bundle

### 7.1 Bundle Command

**File:** `internal/cli/bundle.go`

```go
package cli

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
    Use:   "bundle",
    Short: "Manage offline bundles",
}

var bundleCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "Create offline bundle",
    RunE:  runBundleCreate,
}

var bundleOutput string

func init() {
    bundleCreateCmd.Flags().StringVar(&bundleOutput, "output",
        "aami-offline.tar.gz", "Output file path")
    bundleCmd.AddCommand(bundleCreateCmd)
    rootCmd.AddCommand(bundleCmd)
}

func runBundleCreate(cmd *cobra.Command, args []string) error {
    fmt.Println("Creating offline bundle...")

    // 1. Download all components
    // 2. Include dashboards and alert rules
    // 3. Create tar.gz

    fmt.Printf("Bundle created: %s\n", bundleOutput)
    return nil
}
```

---

## Test Commands

```bash
# Build
go build -o aami ./cmd/aami

# Unit tests
go test ./... -v

# Integration test
./aami init
./aami nodes add test-node --ip 192.168.1.100 --user root --key ~/.ssh/id_rsa
./aami nodes list
./aami alerts apply-preset gpu-production
./aami status
./aami explain xid 79
```

---

## Acceptance Criteria

| Feature | Test Command | Expected Output |
|---------|--------------|-----------------|
| Version | `aami version` | `AAMI v0.1.0 (commit: xxx)` |
| Init | `aami init` | Creates `/etc/aami/config.yaml` |
| Add node | `aami nodes add test --ip 1.2.3.4` | `✅ Node test added` |
| List nodes | `aami nodes list` | Table with node info |
| Apply preset | `aami alerts apply-preset gpu-production` | `✅ Applied preset (8 rules)` |
| Status | `aami status` | Shows component status |
| Explain Xid | `aami explain xid 79` | Shows Xid 79 details |
