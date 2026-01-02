# Phase 4: AMD GPU Support

## Overview

- **Duration**: On-demand (1-2 weeks)
- **Goal**: Add AMD GPU (ROCm) monitoring support alongside NVIDIA
- **Prerequisites**: Phase 3 completed, AMD GPU hardware available for testing

## Entry Criteria

Start Phase 4 when:
- Organization has AMD GPU nodes (MI100, MI200, MI300 series)
- Need to monitor mixed NVIDIA/AMD GPU environments
- ROCm driver is installed on target nodes

## New Files

```
aami/
├── internal/
│   ├── cli/
│   │   └── amd.go              # AMD GPU commands
│   └── amd/                    # AMD GPU support
│       ├── types.go            # ROCm types and metric mapping
│       ├── errors.go           # ROCm error codes database
│       ├── installer.go        # ROCm exporter installer
│       ├── collector.go        # ROCm metrics collector
│       └── health.go           # AMD GPU health scoring
├── configs/
│   └── prometheus/
│       └── amd-alerts.yaml     # AMD GPU alert rules
└── scripts/
    └── install-rocm-exporter.sh
```

---

## Epic 1: AMD GPU Types and Metrics

### 1.1 ROCm Types

**File:** `internal/amd/types.go`

```go
package amd

// GPUType identifies GPU vendor
type GPUType string

const (
    GPUTypeNVIDIA GPUType = "nvidia"
    GPUTypeAMD    GPUType = "amd"
    GPUTypeUnknown GPUType = "unknown"
)

// ROCmMetric maps DCGM metrics to ROCm equivalents
type ROCmMetric struct {
    Name           string
    DCGMEquivalent string
    Description    string
    Unit           string
}

// MetricMapping provides DCGM to ROCm metric translation
var MetricMapping = map[string]ROCmMetric{
    "DCGM_FI_DEV_GPU_TEMP": {
        Name:           "rocm_gpu_temperature",
        DCGMEquivalent: "DCGM_FI_DEV_GPU_TEMP",
        Description:    "GPU temperature in Celsius",
        Unit:           "celsius",
    },
    "DCGM_FI_DEV_GPU_UTIL": {
        Name:           "rocm_gpu_utilization",
        DCGMEquivalent: "DCGM_FI_DEV_GPU_UTIL",
        Description:    "GPU utilization percentage",
        Unit:           "percent",
    },
    "DCGM_FI_DEV_MEM_COPY_UTIL": {
        Name:           "rocm_memory_utilization",
        DCGMEquivalent: "DCGM_FI_DEV_MEM_COPY_UTIL",
        Description:    "Memory controller utilization",
        Unit:           "percent",
    },
    "DCGM_FI_DEV_FB_USED": {
        Name:           "rocm_memory_used",
        DCGMEquivalent: "DCGM_FI_DEV_FB_USED",
        Description:    "GPU memory used",
        Unit:           "bytes",
    },
    "DCGM_FI_DEV_FB_TOTAL": {
        Name:           "rocm_memory_total",
        DCGMEquivalent: "DCGM_FI_DEV_FB_TOTAL",
        Description:    "GPU memory total",
        Unit:           "bytes",
    },
    "DCGM_FI_DEV_POWER_USAGE": {
        Name:           "rocm_power_usage",
        DCGMEquivalent: "DCGM_FI_DEV_POWER_USAGE",
        Description:    "GPU power usage",
        Unit:           "watts",
    },
    "DCGM_FI_DEV_PCIE_TX_THROUGHPUT": {
        Name:           "rocm_pcie_tx",
        DCGMEquivalent: "DCGM_FI_DEV_PCIE_TX_THROUGHPUT",
        Description:    "PCIe transmit throughput",
        Unit:           "bytes/sec",
    },
    "DCGM_FI_DEV_PCIE_RX_THROUGHPUT": {
        Name:           "rocm_pcie_rx",
        DCGMEquivalent: "DCGM_FI_DEV_PCIE_RX_THROUGHPUT",
        Description:    "PCIe receive throughput",
        Unit:           "bytes/sec",
    },
}

// AMDGPUInfo contains AMD GPU information
type AMDGPUInfo struct {
    Index       int
    Name        string
    UUID        string
    VBIOS       string
    Driver      string
    Temperature int
    PowerUsage  float64
    MemoryUsed  uint64
    MemoryTotal uint64
    Utilization int
}
```

### 1.2 ROCm Error Database

**File:** `internal/amd/errors.go`

```go
package amd

// ROCmError represents an AMD GPU error
type ROCmError struct {
    Code        int
    Name        string
    Severity    string
    Description string
    Causes      []string
    Actions     []string
    DocURL      string
}

// ROCmErrors is the database of known ROCm errors
var ROCmErrors = map[int]ROCmError{
    1: {
        Code:        1,
        Name:        "GPU Memory Error",
        Severity:    "Critical",
        Description: "Uncorrectable memory error detected in GPU HBM",
        Causes: []string{
            "HBM hardware failure",
            "Memory aging or degradation",
            "Overheating causing memory instability",
        },
        Actions: []string{
            "Run rocm-smi --showmeminfo to check memory status",
            "Check GPU temperature history",
            "Schedule GPU replacement if errors persist",
        },
        DocURL: "https://rocm.docs.amd.com/en/latest/troubleshooting.html",
    },
    2: {
        Code:        2,
        Name:        "GPU Hang",
        Severity:    "Critical",
        Description: "GPU is not responding to commands",
        Causes: []string{
            "Driver bug or incompatibility",
            "Hardware failure",
            "Infinite loop in GPU kernel",
        },
        Actions: []string{
            "Reset GPU with rocm-smi --gpureset",
            "Update ROCm driver to latest version",
            "Check dmesg for additional error information",
        },
        DocURL: "https://rocm.docs.amd.com/en/latest/troubleshooting.html#gpu-hang",
    },
    3: {
        Code:        3,
        Name:        "PCIe Error",
        Severity:    "Warning",
        Description: "PCIe communication error detected",
        Causes: []string{
            "Loose PCIe connection",
            "PCIe lane degradation",
            "Power supply instability",
        },
        Actions: []string{
            "Reseat GPU in PCIe slot",
            "Check PCIe link status with lspci",
            "Verify power connections",
        },
    },
    4: {
        Code:        4,
        Name:        "Thermal Throttling",
        Severity:    "Warning",
        Description: "GPU is reducing performance due to high temperature",
        Causes: []string{
            "Inadequate cooling",
            "Blocked airflow",
            "Failed cooling fans",
            "High ambient temperature",
        },
        Actions: []string{
            "Check fan status with rocm-smi --showfan",
            "Improve datacenter cooling",
            "Clean dust from heatsinks",
        },
    },
    5: {
        Code:        5,
        Name:        "Power Limit Exceeded",
        Severity:    "Warning",
        Description: "GPU power consumption exceeded configured limit",
        Causes: []string{
            "High workload demand",
            "Power limit set too low",
            "Power supply limitations",
        },
        Actions: []string{
            "Check power settings with rocm-smi --showpower",
            "Adjust power limit if possible",
            "Reduce workload intensity",
        },
    },
    6: {
        Code:        6,
        Name:        "ECC Error - Correctable",
        Severity:    "Info",
        Description: "Single-bit memory error corrected by ECC",
        Causes: []string{
            "Normal operation (occasional errors)",
            "Cosmic ray strikes",
            "Early sign of memory degradation",
        },
        Actions: []string{
            "Monitor error rate over time",
            "No immediate action if rate is low",
            "Consider replacement if rate increases",
        },
    },
    7: {
        Code:        7,
        Name:        "ECC Error - Uncorrectable",
        Severity:    "Critical",
        Description: "Multi-bit memory error that could not be corrected",
        Causes: []string{
            "HBM hardware failure",
            "Memory cell degradation",
        },
        Actions: []string{
            "Drain node from scheduler immediately",
            "Run memory diagnostics",
            "Schedule GPU replacement",
        },
    },
    8: {
        Code:        8,
        Name:        "XGMI Link Error",
        Severity:    "Critical",
        Description: "Error on GPU-to-GPU interconnect link",
        Causes: []string{
            "XGMI cable or connection issue",
            "GPU Infinity Fabric hardware failure",
        },
        Actions: []string{
            "Check XGMI link status with rocm-smi --showtopo",
            "Reseat XGMI cables if applicable",
            "Test with reduced GPU configuration",
        },
    },
}

// GetError returns error information for a given code
func GetError(code int) (ROCmError, bool) {
    err, ok := ROCmErrors[code]
    return err, ok
}

// GetAllErrors returns all known errors
func GetAllErrors() map[int]ROCmError {
    return ROCmErrors
}
```

---

## Epic 2: ROCm Exporter Installation

### 2.1 Installer

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

    return GPUTypeUnknown, fmt.Errorf("no GPU detected on node")
}

// Install installs the ROCm exporter on a node
func (i *Installer) Install(ctx context.Context, node ssh.Node) error {
    result := i.executor.Run(ctx, node, rocmExporterInstallScript)
    if result.Error != nil {
        return fmt.Errorf("install failed: %s", result.Output)
    }
    return nil
}

// Uninstall removes the ROCm exporter from a node
func (i *Installer) Uninstall(ctx context.Context, node ssh.Node) error {
    script := `
systemctl stop rocm-exporter 2>/dev/null || true
systemctl disable rocm-exporter 2>/dev/null || true
rm -f /etc/systemd/system/rocm-exporter.service
rm -f /usr/local/bin/rocm_smi_exporter
systemctl daemon-reload
echo "ROCm exporter uninstalled"
`
    result := i.executor.Run(ctx, node, script)
    if result.Error != nil {
        return fmt.Errorf("uninstall failed: %s", result.Output)
    }
    return nil
}

// CheckInstalled checks if ROCm exporter is installed
func (i *Installer) CheckInstalled(ctx context.Context, node ssh.Node) bool {
    result := i.executor.Run(ctx, node, "systemctl is-active rocm-exporter")
    return result.Error == nil && result.Output == "active"
}

const rocmExporterInstallScript = `#!/bin/bash
set -e

# Check if ROCm is installed
if ! command -v rocm-smi &> /dev/null; then
    echo "ERROR: ROCm is not installed" >&2
    exit 1
fi

echo "ROCm detected, installing exporter..."

# Create temp directory
cd /tmp
rm -rf rocm_smi_exporter 2>/dev/null || true

# Clone and build rocm_smi_exporter
git clone https://github.com/amd/prometheus_rocm_exporter.git rocm_smi_exporter
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
Environment=LD_LIBRARY_PATH=/opt/rocm/lib

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable rocm-exporter
systemctl start rocm-exporter

# Cleanup
cd /
rm -rf /tmp/rocm_smi_exporter

echo "ROCm exporter installed and running on port 9401"
`
```

### 2.2 Installation Script

**File:** `scripts/install-rocm-exporter.sh`

```bash
#!/bin/bash
# AAMI ROCm Exporter Installation Script
# Usage: ./install-rocm-exporter.sh

set -euo pipefail

EXPORTER_PORT=9401
EXPORTER_USER="prometheus"

echo "=== AAMI ROCm Exporter Installation ==="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "ERROR: This script must be run as root"
   exit 1
fi

# Check ROCm installation
if ! command -v rocm-smi &> /dev/null; then
    echo "ERROR: ROCm is not installed"
    echo "Please install ROCm first: https://rocm.docs.amd.com/"
    exit 1
fi

echo "ROCm version:"
rocm-smi --version

# Check for existing installation
if systemctl is-active --quiet rocm-exporter; then
    echo "ROCm exporter is already running"
    read -p "Reinstall? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
    systemctl stop rocm-exporter
fi

# Install dependencies
echo "Installing dependencies..."
apt-get update -qq
apt-get install -y -qq git make golang-go

# Build exporter
echo "Building ROCm exporter..."
cd /tmp
rm -rf prometheus_rocm_exporter
git clone https://github.com/amd/prometheus_rocm_exporter.git
cd prometheus_rocm_exporter
make
cp rocm_smi_exporter /usr/local/bin/
chmod +x /usr/local/bin/rocm_smi_exporter

# Create systemd service
cat > /etc/systemd/system/rocm-exporter.service << EOF
[Unit]
Description=ROCm SMI Prometheus Exporter
Documentation=https://github.com/amd/prometheus_rocm_exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rocm_smi_exporter --web.listen-address=:${EXPORTER_PORT}
Restart=always
RestartSec=5
Environment=LD_LIBRARY_PATH=/opt/rocm/lib

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable rocm-exporter
systemctl start rocm-exporter

# Verify
sleep 2
if systemctl is-active --quiet rocm-exporter; then
    echo ""
    echo "=== Installation Complete ==="
    echo "ROCm exporter running on port ${EXPORTER_PORT}"
    echo "Test with: curl http://localhost:${EXPORTER_PORT}/metrics"
else
    echo "ERROR: Failed to start ROCm exporter"
    journalctl -u rocm-exporter --no-pager -n 20
    exit 1
fi

# Cleanup
rm -rf /tmp/prometheus_rocm_exporter
```

---

## Epic 3: AMD CLI Commands

### 3.1 CLI Implementation

**File:** `internal/cli/amd.go`

```go
package cli

import (
    "context"
    "fmt"
    "os"
    "strconv"
    "time"

    "github.com/fatih/color"
    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"

    "github.com/fregataa/aami/internal/amd"
    "github.com/fregataa/aami/internal/ssh"
)

var amdCmd = &cobra.Command{
    Use:   "amd",
    Short: "AMD GPU specific commands",
    Long: `Commands for managing AMD GPUs with ROCm.

Includes error explanation, exporter installation, and status monitoring.

Examples:
  aami amd explain 1              # Explain error code
  aami amd install node01         # Install ROCm exporter
  aami amd status                 # Show AMD GPU status
  aami amd detect node01          # Detect GPU type on node`,
}

var amdExplainCmd = &cobra.Command{
    Use:   "explain <error-code>",
    Short: "Explain AMD ROCm error code",
    Args:  cobra.ExactArgs(1),
    RunE:  runAmdExplain,
}

var amdInstallCmd = &cobra.Command{
    Use:   "install <node>",
    Short: "Install ROCm exporter on node",
    Args:  cobra.ExactArgs(1),
    RunE:  runAmdInstall,
}

var amdDetectCmd = &cobra.Command{
    Use:   "detect <node>",
    Short: "Detect GPU type on node",
    Args:  cobra.ExactArgs(1),
    RunE:  runAmdDetect,
}

var amdListErrorsCmd = &cobra.Command{
    Use:   "list-errors",
    Short: "List all known ROCm error codes",
    RunE:  runAmdListErrors,
}

func init() {
    amdCmd.AddCommand(amdExplainCmd)
    amdCmd.AddCommand(amdInstallCmd)
    amdCmd.AddCommand(amdDetectCmd)
    amdCmd.AddCommand(amdListErrorsCmd)
    rootCmd.AddCommand(amdCmd)
}

func runAmdExplain(cmd *cobra.Command, args []string) error {
    code, err := strconv.Atoi(args[0])
    if err != nil {
        return fmt.Errorf("invalid error code: %s", args[0])
    }

    errInfo, ok := amd.GetError(code)
    if !ok {
        return fmt.Errorf("unknown ROCm error code: %d", code)
    }

    red := color.New(color.FgRed).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    cyan := color.New(color.FgCyan).SprintFunc()

    fmt.Printf("ROCm Error %d: %s\n", code, cyan(errInfo.Name))
    fmt.Println(strings.Repeat("━", 50))

    severity := errInfo.Severity
    switch severity {
    case "Critical":
        severity = red(severity)
    case "Warning":
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

    if errInfo.DocURL != "" {
        fmt.Printf("\nDocumentation: %s\n", errInfo.DocURL)
    }

    return nil
}

func runAmdInstall(cmd *cobra.Command, args []string) error {
    nodeName := args[0]

    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    node := findNode(cfg, nodeName)
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

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := installer.Install(ctx, *node); err != nil {
        return err
    }

    green := color.New(color.FgGreen).SprintFunc()
    fmt.Printf("%s ROCm exporter installed successfully\n", green("✓"))
    fmt.Println("  Metrics available at: http://" + node.Host + ":9401/metrics")

    return nil
}

func runAmdDetect(cmd *cobra.Command, args []string) error {
    nodeName := args[0]

    cfg, err := loadConfig()
    if err != nil {
        return err
    }

    node := findNode(cfg, nodeName)
    if node == nil {
        return fmt.Errorf("node not found: %s", nodeName)
    }

    executor := ssh.NewExecutor(ssh.ExecutorConfig{
        ConnectTimeout: 10 * time.Second,
        CommandTimeout: 30 * time.Second,
    })

    installer := amd.NewInstaller(executor)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    gpuType, err := installer.DetectGPUType(ctx, *node)
    if err != nil {
        return err
    }

    switch gpuType {
    case amd.GPUTypeNVIDIA:
        fmt.Printf("Node %s: NVIDIA GPU detected\n", nodeName)
    case amd.GPUTypeAMD:
        fmt.Printf("Node %s: AMD GPU detected\n", nodeName)
    default:
        fmt.Printf("Node %s: No GPU detected\n", nodeName)
    }

    return nil
}

func runAmdListErrors(cmd *cobra.Command, args []string) error {
    errors := amd.GetAllErrors()

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Code", "Name", "Severity", "Description"})
    table.SetBorder(false)
    table.SetColumnAlignment([]int{
        tablewriter.ALIGN_RIGHT,
        tablewriter.ALIGN_LEFT,
        tablewriter.ALIGN_LEFT,
        tablewriter.ALIGN_LEFT,
    })

    for code := 1; code <= 8; code++ {
        if err, ok := errors[code]; ok {
            table.Append([]string{
                strconv.Itoa(code),
                err.Name,
                err.Severity,
                truncate(err.Description, 40),
            })
        }
    }

    table.Render()
    return nil
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}
```

---

## Epic 4: AMD Alert Rules

### 4.1 Alert Configuration

**File:** `configs/prometheus/amd-alerts.yaml`

```yaml
groups:
  - name: amd_gpu_critical
    interval: 30s
    rules:
      - alert: AMDGPUMemoryError
        expr: rocm_gpu_ecc_errors_uncorrectable > 0
        for: 0m
        labels:
          severity: critical
          category: amd
        annotations:
          summary: "AMD GPU uncorrectable memory error on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} has uncorrectable ECC errors. Data corruption possible."
          action: "aami slurm drain {{ $labels.instance | reReplaceAll \":.*\" \"\" }}"

      - alert: AMDGPUHang
        expr: rocm_gpu_hang == 1
        for: 0m
        labels:
          severity: critical
          category: amd
        annotations:
          summary: "AMD GPU hang detected on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} is not responding."
          action: "Reset GPU or reboot node"

  - name: amd_gpu_warning
    interval: 60s
    rules:
      - alert: AMDGPUHighTemperature
        expr: rocm_gpu_temperature > 85
        for: 5m
        labels:
          severity: warning
          category: amd
        annotations:
          summary: "AMD GPU high temperature on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} temperature is {{ $value }}°C"

      - alert: AMDGPUThermalThrottling
        expr: rocm_gpu_throttle_status > 0
        for: 5m
        labels:
          severity: warning
          category: amd
        annotations:
          summary: "AMD GPU thermal throttling on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} is throttling due to high temperature"

      - alert: AMDGPUMemoryUsageHigh
        expr: (rocm_memory_used / rocm_memory_total) > 0.95
        for: 10m
        labels:
          severity: warning
          category: amd
        annotations:
          summary: "AMD GPU memory nearly full on {{ $labels.instance }}"
          description: "GPU {{ $labels.gpu }} memory usage is {{ $value | humanizePercentage }}"
```

---

## Test Commands

```bash
# Build
make build

# AMD commands
./bin/aami amd explain 1
./bin/aami amd explain 7
./bin/aami amd list-errors
./bin/aami amd detect amd-node-01
./bin/aami amd install amd-node-01
```

---

## Acceptance Criteria

| Feature | Test Command | Expected Output |
|---------|--------------|-----------------|
| Explain error | `aami amd explain 1` | Memory error details with actions |
| List errors | `aami amd list-errors` | Table of all ROCm error codes |
| Detect GPU | `aami amd detect node01` | GPU type (NVIDIA/AMD/Unknown) |
| Install exporter | `aami amd install node01` | Exporter running on port 9401 |

---

## Dependencies

- ROCm 5.0+ installed on target nodes
- prometheus_rocm_exporter or compatible exporter
- Go 1.21+ for building exporter from source

## Notes

- AMD GPU metrics use different naming conventions than NVIDIA/DCGM
- XGMI topology (GPU interconnect) is AMD-specific
- Some features require specific GPU models (MI100+)
