# Sprint 16: Bootstrap Automation & One-Line Deployment

**Status**: 📋 Planned
**Priority**: High (Critical for User Experience)
**Duration**: 3-4 weeks
**Started**: TBD
**Completed**: TBD

## Goals

Implement fully automated installation experience for both Config Server and monitoring nodes. Enable one-line deployment commands with comprehensive pre-installation validation.

### Key Objectives
1. **Preflight Validation**: Verify system requirements before installation
2. **Config Server One-Click Install**: Single command to deploy entire monitoring stack
3. **Node Bootstrap Automation**: Single command to register and configure monitoring nodes
4. **Progress Feedback**: Clear visual feedback during installation process

---

## Problem Statement

### Current Process (Manual - 10+ steps total)

#### Config Server Installation (3-4 steps)
```bash
# 1. Clone repository
git clone https://github.com/fregataa/aami.git
cd aami

# 2. Configure environment
cd deploy/docker-compose
cp .env.example .env
vim .env  # Manual editing

# 3. Start services
docker-compose up -d

# 4. Verify installation
curl http://localhost:8080/api/v1/health
```

#### Node Registration (5-6 steps)
```bash
# 1. SSH to target node
ssh user@target-node

# 2. Download installation script
wget https://raw.githubusercontent.com/.../install-node-exporter.sh

# 3. Run installation script
sudo ./install-node-exporter.sh

# 4. Register target in Config Server (from admin machine)
curl -X POST http://config-server:8080/api/v1/targets -d '{...}'

# 5. Register exporter in Config Server
curl -X POST http://config-server:8080/api/v1/exporters -d '{...}'

# 6. Wait for Prometheus to discover (30s)
```

**Pain Points:**
- No pre-installation validation (failures occur mid-installation)
- 10+ manual steps for complete setup
- Requires knowledge of Docker, Config Server API
- Easy to make mistakes (wrong IP, port, missing dependencies)
- No progress feedback during long operations
- Doesn't scale (imagine 100 nodes)
- System info must be collected manually

### Target Process (Automated)

#### Config Server (1 step)
```bash
curl -fsSL https://aami.io/install.sh | bash
```

#### Node Registration (1 step)
```bash
curl -fsSL https://example.com/bootstrap.sh | \
  sudo bash -s -- --token aami_xxx --server https://config-server.example.com
```

**Result:**
- Pre-installation checks prevent mid-install failures
- Clear progress indicators show installation status
- Automatic error recovery where possible
- Exporter installed and running
- Node registered in Config Server
- Prometheus automatically discovers and scrapes
- System info auto-detected (hostname, IP, labels)
- Works for bulk deployment (100+ nodes)

---

## Tasks

### Phase 0: Preflight Check Script (2 days) - COMPLETED

#### Create `preflight-check.sh`
**Location**: `scripts/preflight-check.sh`

**Purpose**: Validate system requirements BEFORE installation to prevent mid-install failures.

**Functionality:**
- [x] Check system requirements (CPU, RAM, disk space)
- [x] Verify required software (Docker, curl, systemctl)
- [x] Test network connectivity (Config Server, registries)
- [x] Check port availability (8080, 9090, 9100, 3000, 5432)
- [x] Verify permissions (root/sudo, Docker socket)
- [x] Detect hardware (GPU for node mode)
- [x] Output JSON for automation or human-readable summary
- [x] Provide `--fix` option for automatic remediation

**Command-line Interface:**
```bash
preflight-check.sh [OPTIONS]

Options:
  --mode MODE            Check mode: 'server' or 'node' (default: auto-detect)
  --server URL           Config Server URL (for node mode connectivity check)
  --fix                  Attempt automatic fixes for issues
  --json                 Output results in JSON format
  --quiet                Only show errors
  --verbose              Show detailed check information

Examples:
  # Basic server check
  ./preflight-check.sh --mode server

  # Node check with connectivity test
  ./preflight-check.sh --mode node --server https://config.example.com

  # Auto-fix issues
  ./preflight-check.sh --mode server --fix

  # JSON output for CI/CD
  ./preflight-check.sh --mode node --json
```

**Check Categories:**

##### System Requirements
```bash
check_system_requirements() {
    # CPU cores (minimum 2 for server, 1 for node)
    # RAM (minimum 4GB for server, 1GB for node)
    # Disk space (minimum 20GB for server, 5GB for node)
    # OS compatibility (Ubuntu 20.04+, CentOS 8+, Debian 11+)
}
```

##### Software Dependencies
```bash
check_software_dependencies() {
    # Server mode: docker, docker-compose, curl
    # Node mode: curl, systemctl
    # Optional: jq, promtool
}
```

##### Network Connectivity
```bash
check_network_connectivity() {
    # DNS resolution
    # Registry access (docker.io, ghcr.io)
    # Config Server reachability (node mode)
    # Outbound HTTPS (443)
}
```

##### Port Availability
```bash
check_port_availability() {
    # Server mode: 8080, 9090, 3000, 5432, 6379
    # Node mode: 9100, 9400
    # Report which process is using conflicting ports
}
```

##### Permissions
```bash
check_permissions() {
    # root or sudo access
    # Docker socket access
    # /etc write access (systemd)
    # /var/lib write access (data)
}
```

##### Hardware Detection (Node Mode)
```bash
check_hardware() {
    # NVIDIA GPU (nvidia-smi)
    # AMD GPU (rocm-smi)
    # InfiniBand (ibstat)
    # NVMe devices
}
```

**Output Format:**
```
AAMI Preflight Check v1.0
=========================

Mode: Server Installation

System Requirements
  [✓] OS: Ubuntu 22.04 LTS (supported)
  [✓] CPU: 8 cores (minimum: 2)
  [✓] RAM: 32GB (minimum: 4GB)
  [✓] Disk: 150GB free (minimum: 20GB)

Software Dependencies
  [✓] Docker: 24.0.5 (minimum: 20.10)
  [✓] Docker Compose: v2.20.2 (minimum: v2.0)
  [✓] curl: 7.81.0

Network Connectivity
  [✓] DNS: working
  [✓] docker.io: reachable
  [✓] ghcr.io: reachable

Port Availability
  [✓] Port 8080: available
  [✓] Port 9090: available
  [✗] Port 3000: in use by 'grafana' (PID: 1234)
  [✓] Port 5432: available

Permissions
  [✓] sudo: available
  [✓] Docker socket: accessible

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Result: 1 issue found

ERRORS:
  [✗] Port 3000 is already in use by process 'grafana' (PID: 1234)
      Fix: Stop the service with 'sudo systemctl stop grafana-server'
      Or: Change GRAFANA_PORT in .env to use a different port

Run with --fix to attempt automatic fixes.
Exit code: 1
```

**JSON Output (for automation):**
```json
{
  "version": "1.0",
  "mode": "server",
  "timestamp": "2025-01-15T10:30:00Z",
  "passed": false,
  "checks": {
    "system": {
      "passed": true,
      "os": {"value": "Ubuntu 22.04", "passed": true},
      "cpu": {"value": 8, "minimum": 2, "passed": true},
      "ram_gb": {"value": 32, "minimum": 4, "passed": true},
      "disk_gb": {"value": 150, "minimum": 20, "passed": true}
    },
    "software": {
      "passed": true,
      "docker": {"value": "24.0.5", "minimum": "20.10", "passed": true},
      "docker_compose": {"value": "2.20.2", "minimum": "2.0", "passed": true}
    },
    "network": {
      "passed": true
    },
    "ports": {
      "passed": false,
      "8080": {"available": true},
      "3000": {"available": false, "process": "grafana", "pid": 1234}
    },
    "permissions": {
      "passed": true
    }
  },
  "errors": [
    {
      "category": "ports",
      "code": "PORT_IN_USE",
      "message": "Port 3000 is in use by grafana (PID: 1234)",
      "fix_command": "sudo systemctl stop grafana-server"
    }
  ]
}
```

---

### Phase 1: Config Server Install Script (2 days) - COMPLETED

#### Create `install-server.sh`
**Location**: `scripts/install-server.sh`

**Purpose**: One-command installation of complete AAMI monitoring stack.

**Functionality:**
- [x] Run preflight checks automatically
- [x] Clone or download AAMI repository
- [x] Generate secure default credentials
- [x] Configure environment variables interactively or via flags
- [x] Start Docker Compose stack
- [x] Wait for services to be healthy
- [x] Create initial bootstrap token
- [x] Display success summary with next steps

**Command-line Interface:**
```bash
install-server.sh [OPTIONS]

Options:
  --version VERSION      AAMI version to install (default: latest)
  --install-dir PATH     Installation directory (default: /opt/aami)
  --data-dir PATH        Data directory (default: /var/lib/aami)
  --domain DOMAIN        Domain for Config Server (default: localhost)
  --https                Enable HTTPS with Let's Encrypt
  --postgres-password PW PostgreSQL password (auto-generated if not set)
  --grafana-password PW  Grafana admin password (auto-generated if not set)
  --skip-preflight       Skip preflight checks (not recommended)
  --unattended           Non-interactive mode
  --verbose              Show detailed output

Examples:
  # Interactive installation
  curl -fsSL https://aami.io/install.sh | bash

  # Unattended installation with custom settings
  curl -fsSL https://aami.io/install.sh | bash -s -- \
    --unattended \
    --domain config.example.com \
    --https \
    --install-dir /opt/aami

  # Specify version
  curl -fsSL https://aami.io/install.sh | bash -s -- --version v1.0.0
```

**Installation Flow:**
```
┌─────────────────────────────────────────────────────────────┐
│                 AAMI Server Installation                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [1/7] Running preflight checks...              [====    ]  │
│        ✓ System requirements met                            │
│        ✓ Docker 24.0.5 detected                             │
│        ✓ All ports available                                │
│                                                              │
│  [2/7] Downloading AAMI v1.2.0...               [======  ]  │
│        ✓ Downloaded to /opt/aami                            │
│                                                              │
│  [3/7] Configuring environment...               [======= ]  │
│        ✓ Generated secure credentials                       │
│        ✓ Created .env file                                  │
│                                                              │
│  [4/7] Starting services...                     [========]  │
│        ✓ PostgreSQL started                                 │
│        ✓ Redis started                                      │
│        ✓ Config Server started                              │
│        ✓ Prometheus started                                 │
│        ✓ Grafana started                                    │
│                                                              │
│  [5/7] Waiting for health checks...             [========]  │
│        ✓ All services healthy                               │
│                                                              │
│  [6/7] Creating bootstrap token...              [========]  │
│        ✓ Token created: aami_bootstrap_xxx...               │
│                                                              │
│  [7/7] Finalizing installation...               [========]  │
│        ✓ Installation complete!                             │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ AAMI installed successfully!                            │
│                                                              │
│  Access URLs:                                                │
│    Config Server: http://localhost:8080                     │
│    Grafana:       http://localhost:3000 (admin/Xyz123...)   │
│    Prometheus:    http://localhost:9090                     │
│                                                              │
│  Bootstrap Token (save this!):                              │
│    aami_bootstrap_abc123def456...                           │
│                                                              │
│  Next Steps:                                                │
│    1. Register nodes using the bootstrap token              │
│    2. Access Grafana to view dashboards                     │
│    3. Configure alert rules in Config Server                │
│                                                              │
│  Documentation: https://aami.io/docs                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### Phase 2: Bootstrap Script Core (3 days) - COMPLETED

#### Create `bootstrap.sh`
**Location**: `scripts/node/bootstrap.sh`

**Functionality:**
- [x] Run preflight checks for node mode
- [x] Parse command-line arguments (--token, --server, --port, etc.)
- [x] Detect system information automatically
  - Hostname
  - Primary IP address
  - OS type and version
  - Architecture (amd64/arm64)
  - Available exporters (GPU detection)
- [x] Validate bootstrap token with Config Server
- [x] Install Node Exporter (call existing script)
- [x] Install dynamic-check.sh and register cron
- [x] Register node using Bootstrap Register API
- [x] Verify registration success
- [x] Show progress with visual feedback
- [x] Print summary and next steps

**Command-line Interface:**
```bash
bootstrap.sh [OPTIONS]

Required:
  --token TOKEN          Bootstrap token from Config Server
  --server URL           Config Server URL (e.g., https://config.example.com)

Optional:
  --port PORT            Node Exporter port (default: 9100)
  --labels KEY=VALUE     Additional labels (can be repeated)
  --dry-run              Show what would be done without executing
  --verbose              Enable verbose output
  --skip-preflight       Skip preflight checks (not recommended)
  --skip-firewall        Skip firewall configuration
  --skip-gpu             Skip GPU detection and exporter installation
  --unattended           Non-interactive mode (for automation)

Examples:
  # Basic usage
  ./bootstrap.sh --token aami_xxx --server https://config.example.com

  # With custom labels
  ./bootstrap.sh --token aami_xxx --server https://config.example.com \
    --labels env=production --labels rack=A1

  # Dry run to preview
  ./bootstrap.sh --token aami_xxx --server https://config.example.com --dry-run
```

**Progress Display:**
```
┌─────────────────────────────────────────────────────────────┐
│                  AAMI Node Bootstrap                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [1/8] Running preflight checks...              [====    ]  │
│        ✓ System: Ubuntu 22.04, 32GB RAM, 8 cores           │
│        ✓ Network: Config Server reachable                   │
│        ✓ Ports: 9100 available                              │
│                                                              │
│  [2/8] Validating bootstrap token...            [=====   ]  │
│        ✓ Token valid, 47 uses remaining                     │
│        ✓ Default group: production                          │
│                                                              │
│  [3/8] Detecting hardware...                    [======  ]  │
│        ✓ NVIDIA GPU detected: 8x A100-SXM4-80GB            │
│        ✓ InfiniBand: mlx5_0 (HDR)                          │
│                                                              │
│  [4/8] Installing Node Exporter...              [======= ]  │
│        ✓ Downloaded node_exporter 1.7.0                     │
│        ✓ Service configured and started                     │
│                                                              │
│  [5/8] Installing DCGM Exporter...              [======= ]  │
│        ✓ DCGM Exporter installed                            │
│        ✓ Service configured and started                     │
│                                                              │
│  [6/8] Installing Dynamic Check...              [========]  │
│        ✓ dynamic-check.sh installed                         │
│        ✓ Cron job registered (1 min interval)              │
│                                                              │
│  [7/8] Registering with Config Server...        [========]  │
│        ✓ Target registered: target-abc123                   │
│        ✓ Exporters registered: node_exporter, dcgm_exporter│
│                                                              │
│  [8/8] Verifying registration...                [========]  │
│        ✓ Metrics accessible                                 │
│        ✓ Config Server confirmed registration               │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ Node bootstrap complete!                                │
│                                                              │
│  Node Information:                                           │
│    Hostname:  gpu-node-01.example.com                       │
│    Target ID: target-abc123def456                           │
│    Group:     production                                    │
│                                                              │
│  Installed Exporters:                                        │
│    - Node Exporter:    http://localhost:9100/metrics        │
│    - DCGM Exporter:    http://localhost:9400/metrics        │
│    - Dynamic Checks:   /var/lib/node_exporter/textfile/     │
│                                                              │
│  The node will appear in Prometheus within 30 seconds.      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### Script Flow
```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Parse arguments and validate
# 2. Run preflight checks (unless --skip-preflight)
# 3. Detect system information
#    - hostname=$(hostname -f)
#    - ip=$(hostname -I | awk '{print $1}')
#    - os=$(. /etc/os-release && echo "$ID $VERSION_ID")
#    - arch=$(uname -m)
# 4. Validate bootstrap token with Config Server
#    GET /api/v1/bootstrap-tokens/validate?token=xxx
# 5. Detect hardware (GPU, InfiniBand)
# 6. Install Node Exporter
#    Call install-node-exporter.sh
# 7. Install GPU Exporter if GPU detected
#    Call install-dcgm-exporter.sh
# 8. Install dynamic-check.sh and cron
# 9. Wait for exporters to be healthy
#    curl http://localhost:9100/metrics
# 10. Register node with Config Server
#    POST /api/v1/bootstrap/register
#    {
#      "token": "aami_xxx",
#      "hostname": "server-01",
#      "ip_address": "192.168.1.100",
#      "exporters": [{"type": "node_exporter", "port": 9100}],
#      "labels": {"env": "production", "os": "linux"}
#    }
# 11. Verify registration
#    GET /api/v1/targets?hostname=server-01
# 12. Print success message and instructions
```

---

### Phase 3: Heartbeat API Integration (1 day) - NOT PLANNED

> **Decision**: Node agent is not planned. Heartbeat requires a persistent agent process on nodes, which adds operational complexity without sufficient value for most use cases. Prometheus scrape failures already indicate node status.

~~**Background**: Heartbeat endpoint already exists (`POST /api/v1/targets/:id/heartbeat`), needs integration with bootstrap workflow.~~

#### ~~Bootstrap Script Enhancement~~
- [ ] ~~Install heartbeat service during bootstrap~~
- [ ] ~~Send periodic heartbeat after successful registration~~
- [ ] ~~Handle heartbeat failures gracefully~~
- [ ] ~~Log heartbeat status~~

**Implementation:**
```bash
# In bootstrap.sh, after successful registration
echo "Installing heartbeat service..."

cat > /etc/systemd/system/aami-heartbeat.service <<EOF
[Unit]
Description=AAMI Config Server Heartbeat
After=network.target

[Service]
Type=simple
Environment="TARGET_ID=${TARGET_ID}"
Environment="CONFIG_SERVER=${CONFIG_SERVER}"
ExecStart=/usr/local/bin/aami-heartbeat.sh
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF

# Create heartbeat script
cat > /usr/local/bin/aami-heartbeat.sh <<'HEARTBEAT_EOF'
#!/bin/bash
set -e

while true; do
  response=$(curl -X POST "$CONFIG_SERVER/api/v1/targets/$TARGET_ID/heartbeat" \
    -H "Content-Type: application/json" \
    -w "%{http_code}" \
    -o /dev/null \
    -s) || true

  if [ "$response" != "200" ]; then
    logger -t aami-heartbeat "Heartbeat failed with status $response"
  fi

  sleep 30
done
HEARTBEAT_EOF

chmod +x /usr/local/bin/aami-heartbeat.sh
systemctl daemon-reload
systemctl enable aami-heartbeat
systemctl start aami-heartbeat
```

#### Verification
- [ ] Test heartbeat service installation
- [ ] Verify `last_seen` updates in database
- [ ] Test heartbeat service restart after failure

---

### Phase 4: Automatic Status Management (2 days) - NOT PLANNED

> **Decision**: Depends on Phase 3 (Heartbeat). Without node agent heartbeat, automatic status management is not feasible. Use Prometheus alerting for node down detection instead.

#### ~~Background Task Implementation~~
- [ ] ~~Create background job to detect stale targets~~
- [ ] ~~Mark targets as `inactive` if `last_seen` > 5 minutes~~
- [ ] ~~Mark targets as `down` if `last_seen` > 15 minutes~~
- [ ] ~~Add configuration for timeout thresholds~~

**Service Layer:**
```go
// internal/service/target_monitor.go
type TargetMonitor struct {
    targetService *TargetService
    ticker        *time.Ticker
    config        TargetMonitorConfig
}

type TargetMonitorConfig struct {
    CheckInterval     time.Duration // default: 1 minute
    InactiveThreshold time.Duration // default: 5 minutes
    DownThreshold     time.Duration // default: 15 minutes
}

func (m *TargetMonitor) Start() {
    m.ticker = time.NewTicker(m.config.CheckInterval)
    go func() {
        for range m.ticker.C {
            m.checkStaleTargets()
        }
    }()
}

func (m *TargetMonitor) checkStaleTargets() {
    ctx := context.Background()

    // Mark inactive (no heartbeat for 5 minutes)
    inactiveThreshold := time.Now().Add(-m.config.InactiveThreshold)
    m.targetService.MarkStaleTargets(ctx, inactiveThreshold, domain.StatusInactive)

    // Mark down (no heartbeat for 15 minutes)
    downThreshold := time.Now().Add(-m.config.DownThreshold)
    m.targetService.MarkStaleTargets(ctx, downThreshold, domain.StatusDown)
}
```

#### Repository Layer
- [ ] Add `MarkStaleTargets()` method to TargetRepository
- [ ] Batch update targets based on `last_seen` threshold

#### Service Discovery Integration
- [ ] Ensure `/api/v1/sd/prometheus/active` excludes inactive targets
- [ ] Verify Prometheus stops scraping inactive targets automatically

#### Testing
- [ ] Test status transitions (active → inactive → down)
- [ ] Test recovery (down → active after heartbeat)
- [ ] Test Prometheus target discovery updates

---

### Phase 5: GPU Support (2 days)

#### GPU Detection
- [ ] Detect NVIDIA GPUs (`nvidia-smi`)
- [ ] Detect AMD GPUs (`rocm-smi`)
- [ ] Detect Intel GPUs (future)

#### Auto-install GPU Exporters
- [ ] Install DCGM Exporter for NVIDIA GPUs
- [ ] Register GPU exporter in Config Server
- [ ] Add GPU-related labels automatically

**Example:**
```bash
# If NVIDIA GPU detected, automatically:
# 1. Install dcgm-exporter
# 2. Register with type=dcgm_exporter, port=9400
# 3. Add labels: gpu_vendor=nvidia, gpu_count=8, gpu_model=A100
```

---

### Phase 6: Bootstrap API Enhancement (2 days)

#### Update Bootstrap Register API
**Current**: Accepts only basic target info
**Enhancement**: Accept exporters array

**Updated DTO** (`internal/api/dto/bootstrap.go`):
```go
type BootstrapRegisterRequest struct {
    Token     string                 `json:"token" binding:"required"`
    Hostname  string                 `json:"hostname" binding:"required"`
    IPAddress string                 `json:"ip_address" binding:"required,ip"`
    Labels    map[string]string      `json:"labels,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`

    // Allow bootstrap script to register exporters in one call
    Exporters []ExporterInfo         `json:"exporters,omitempty"`

    // System information for diagnostics
    SystemInfo *SystemInfo           `json:"system_info,omitempty"`
}

type ExporterInfo struct {
    Type           string                 `json:"type" binding:"required,oneof=node_exporter dcgm_exporter custom"`
    Port           int                    `json:"port" binding:"required"`
    MetricsPath    string                 `json:"metrics_path,omitempty"`
    ScrapeInterval string                 `json:"scrape_interval,omitempty"`
    Config         map[string]interface{} `json:"config,omitempty"`
}

type SystemInfo struct {
    OS           string `json:"os"`
    OSVersion    string `json:"os_version"`
    Architecture string `json:"architecture"`
    CPUCores     int    `json:"cpu_cores"`
    MemoryGB     int    `json:"memory_gb"`
    GPUVendor    string `json:"gpu_vendor,omitempty"`
    GPUModel     string `json:"gpu_model,omitempty"`
    GPUCount     int    `json:"gpu_count,omitempty"`
}
```

#### Service Layer Update
- [ ] Update `BootstrapService.RegisterNode()` to handle exporters
- [ ] Create target and exporters in a single transaction
- [ ] Return complete registration info (target + exporters)

---

### Phase 7: Additional Exporters (2 days)

#### Create Installation Scripts
- [ ] `install-dcgm-exporter.sh` - NVIDIA GPU metrics
- [ ] `install-all-smi.sh` - Multi-vendor GPU support
- [ ] Update `bootstrap.sh` to call appropriate scripts based on detection

---

### Phase 8: Uninstall Support (1 day)

#### Create `uninstall.sh`
- [ ] Stop and disable all exporter services
- [ ] Remove exporter binaries
- [ ] Remove systemd service files
- [ ] Remove cron jobs
- [ ] Optionally unregister from Config Server
- [ ] Clean up configuration files

**Usage:**
```bash
# Uninstall all exporters
sudo ./uninstall.sh

# Uninstall and unregister from Config Server
sudo ./uninstall.sh --unregister --server https://config.example.com --token aami_xxx
```

---

### Phase 9: Testing (2 days)

#### Unit Tests
- [ ] Test argument parsing
- [ ] Test system detection functions
- [ ] Test preflight check functions
- [ ] Test token validation
- [ ] Test progress display
- [ ] Test API calls (mock server)

#### Integration Tests
- [ ] Test full server installation flow in Docker
- [ ] Test full bootstrap flow in Docker
- [ ] Test with various OS (Ubuntu, CentOS, Debian)
- [ ] Test with GPU detection (mock nvidia-smi)
- [ ] Test error scenarios (invalid token, network failure, port conflict)
- [ ] Test preflight --fix functionality

#### Manual Testing
- [ ] Test on real Ubuntu 22.04 server
- [ ] Test on real CentOS 8 server
- [ ] Test with real GPU node
- [ ] Test bulk deployment (10+ nodes)

---

### Phase 10: Documentation (1 day)

#### User Documentation
- [ ] Create `docs/INSTALLATION.md` - Complete installation guide
- [ ] Create `docs/BOOTSTRAP.md` - Node bootstrap guide
- [ ] Update `docs/QUICKSTART.md` - Add one-line install methods
- [ ] Update `docs/NODE-REGISTRATION.md` - Add automated section
- [ ] Add troubleshooting section with common preflight failures

#### Operator Documentation
- [ ] Token management best practices
- [ ] Security considerations
- [ ] Bulk deployment guide
- [ ] CI/CD integration examples
- [ ] Offline installation guide

---

## Deliverables

### Scripts
| Script | Purpose | Status |
|--------|---------|--------|
| `scripts/preflight-check.sh` | Pre-installation validation | ✅ Completed |
| `scripts/install-server.sh` | One-click server installation | ✅ Completed |
| `scripts/node/bootstrap.sh` | Node automation script | ✅ Completed |
| `scripts/node/install-dcgm-exporter.sh` | GPU exporter | 📋 Planned |
| `scripts/node/install-all-smi.sh` | Multi-vendor GPU | 📋 Planned |
| `scripts/node/uninstall.sh` | Cleanup script | 📋 Planned |

### API Updates
- [ ] Enhanced Bootstrap Register API with exporters support
- [ ] Transaction support for atomic registration
- [ ] System info collection endpoint

### Documentation
- [ ] `docs/INSTALLATION.md` - Complete installation guide
- [ ] `docs/BOOTSTRAP.md` - Node bootstrap guide
- [ ] Updated QUICKSTART.md
- [ ] Updated NODE-REGISTRATION.md
- [ ] Troubleshooting guide

### Tests
- [ ] Unit tests for script functions
- [ ] Integration tests for full flow
- [ ] Multi-OS compatibility tests
- [ ] Preflight check tests

---

## Success Criteria

### Installation Experience
- [ ] Config Server installable with single command
- [ ] Node bootstrap with single command
- [ ] Preflight checks catch 95%+ of potential issues before installation
- [ ] Clear progress feedback during all installations
- [ ] Less than 5 minutes for complete server installation
- [ ] Less than 2 minutes for node bootstrap

### Compatibility
- [ ] Works on Ubuntu 20.04, 22.04
- [ ] Works on CentOS 8, Rocky Linux 8
- [ ] Works on Debian 11, 12
- [ ] Automatically detects and installs GPU exporters

### Reliability
- [ ] Handles errors gracefully with clear messages
- [ ] Preflight --fix resolves common issues automatically
- [ ] Idempotent (safe to run multiple times)

### Documentation
- [ ] Documentation is clear and complete
- [ ] Troubleshooting covers common issues
- [ ] Examples for all major use cases

### Scale
- [ ] Tested with 10+ nodes in bulk deployment
- [ ] Supports unattended/automation mode

---

## Benefits

### For Operators
- **Time Savings**: 10+ steps → 2 steps (90% reduction)
- **Reliability**: Preflight checks prevent failed installations
- **Visibility**: Progress indicators show what's happening
- **Scalability**: Easy to deploy 100+ nodes
- **Security**: Token-based authentication
- **Consistency**: Same process for all nodes
- **Less Errors**: Auto-detection reduces mistakes

### For Users
- **Quick Onboarding**: New environment online in minutes
- **Simple**: No need to learn Docker or Config Server API
- **Confidence**: Preflight checks confirm readiness
- **Feedback**: Know exactly what's happening during install
- **Repeatable**: Works the same every time
- **Flexible**: Custom labels and configurations

### For Development
- **Easy Testing**: Spin up test environments quickly
- **CI/CD Ready**: Integrate into automation pipelines
- **Monitoring**: Track deployment success metrics
- **Debugging**: Preflight JSON output for automation

---

## Dependencies

### External
- Config Server with Bootstrap Register API
- Bootstrap tokens feature
- Service Discovery working

### Internal
- `install-node-exporter.sh` (existing)
- Bootstrap token management in Config Server

---

## Security Considerations

- Bootstrap tokens have expiration
- Bootstrap tokens have usage limits
- HTTPS required for Config Server (recommended)
- Token transmitted securely
- No hardcoded credentials
- Generated passwords are cryptographically secure
- Script must run as root (exporter installation)
- Validate script source before execution
- Use HTTPS for script downloads

---

## Future Enhancements (Post-Sprint)

- [ ] TUI (Text User Interface) installation wizard
- [ ] Web-based installation wizard
- [ ] Windows support
- [ ] macOS support
- [ ] Docker/Kubernetes deployment mode
- [ ] Agent auto-update mechanism
- [ ] Centralized configuration push
- [ ] Custom exporter auto-discovery
- [ ] Certificate-based authentication
- [ ] Installation telemetry (opt-in)

---

## Notes

- Prioritize Linux support (Ubuntu, CentOS, Debian)
- Keep scripts POSIX-compliant where possible
- Provide clear error messages for debugging
- Log all actions for audit trail
- Follow existing script conventions in `scripts/README.md`
- Progress display should work in both interactive and non-interactive modes
- JSON output mode for all scripts (automation support)

---

## Related Sprints

- **Sprint 06**: Error Handling (needed for better API errors)
- **Sprint 09**: Authentication (future: API key instead of tokens)
- **Sprint 11**: Alert Rule Generation (benefits from easy deployment)
- **Sprint 12-15**: Custom Exporters (all benefit from bootstrap automation)

---

**Estimated Effort**: 18-24 days
**Priority**: High (Critical for User Experience)
**Complexity**: Medium-High
