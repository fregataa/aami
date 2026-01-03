# AAMI (AI Accelerator Monitoring Infrastructure)

> **All-in-one monitoring tool for GPU clusters without Kubernetes**
>
> Simplifies the installation, configuration, and operation of the Prometheus stack through a single CLI/UI, with GPU-specific diagnostic features

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

## Why AAMI?

Setting up GPU cluster monitoring without K8s takes **2-3 days**:

| Task | Traditional Approach | AAMI |
|------|---------------------|------|
| Install Prometheus + Grafana + Alertmanager | Half day | Automatic |
| Deploy DCGM exporter | 2-3 hours | Automatic |
| Write alert rules | Half day (learning PromQL) | Presets provided |
| Slack/Email integration | 2-3 hours | CLI/UI configuration |
| Air-gap environment support | 1-2 days | Bundle provided |
| **Total Time** | **2-3 days** | **30 minutes** |

### Key Differentiators

```
"Setting up Prometheus + Grafana + Alertmanager + DCGM on a GPU cluster takes 2-3 days.
With AAMI, it takes 30 minutes. Air-gap is also supported.
And when Xid 79 occurs, it tells you what it is, why it happened, and what to do."
```

## Features

### 1. One-Click Installation

```bash
# Online installation
curl -fsSL https://get.aami.dev | bash
aami init

# Air-gap installation
aami bundle create --output aami-offline.tar.gz  # On internet-connected machine
aami init --offline ./aami-offline.tar.gz        # On air-gapped machine
```

### 2. Node Management CLI

```bash
# Add node
aami nodes add gpu-node-01 --ip 192.168.1.101 --user root --key ~/.ssh/id_rsa

# Bulk add
aami nodes add --file hosts.txt

# List nodes
aami nodes list

┌──────────────┬───────────────┬──────┬────────┬─────────┐
│ Name         │ IP            │ GPUs │ Status │ Alerts  │
├──────────────┼───────────────┼──────┼────────┼─────────┤
│ gpu-node-01  │ 192.168.1.101 │ 8    │ ✅     │ 0       │
│ gpu-node-02  │ 192.168.1.102 │ 8    │ ⚠️     │ 1       │
└──────────────┴───────────────┴──────┴────────┴─────────┘
```

### 3. GPU-Specific Alert Presets

```bash
aami alerts apply-preset gpu-production
# → 8 alert rules applied instantly
```

| Alert | Condition | Severity |
|-------|-----------|----------|
| GPU Temperature Overheat | temp > 85°C for 5 minutes | Critical |
| GPU Memory Leak | memory > 95% AND util < 5% | Warning |
| ECC Error Threshold | ECC errors > 100/24h | Critical |
| NVLink Error | NVLink error count increase | Warning |
| Xid Error Detected | Xid error detected | Critical |
| Node Down | node_exporter not responding | Critical |

### 4. Xid Error Interpretation (Differentiating Feature)

```bash
aami explain xid 79

┌─────────────────────────────────────────────────────────────────┐
│ Xid 79: GPU has fallen off the bus                             │
├─────────────────────────────────────────────────────────────────┤
│ Severity: Critical                                             │
│                                                                 │
│ Meaning:                                                       │
│   GPU disconnected from PCIe bus. System cannot communicate    │
│   with the GPU.                                                │
│                                                                 │
│ Common Causes:                                                 │
│   1. PCIe slot contact failure                                 │
│   2. Unstable power supply                                     │
│   3. GPU hardware defect                                       │
│                                                                 │
│ Recommended Actions:                                           │
│   1. Immediately remove the node from workload                 │
│   2. Attempt GPU reseat (reinstallation)                       │
│   3. Consider GPU replacement if issue recurs                  │
└─────────────────────────────────────────────────────────────────┘
```

### 5. Web UI Alert Configuration

Configure alerts with clicks instead of YAML editing:

```
┌─────────────────────────────────────────────────────────────────┐
│ 🔔 Alert Rules                                        [+ New]   │
├─────────────────────────────────────────────────────────────────┤
│ ☑ GPU Temperature Critical                                     │
│   Condition: GPU temp > [85]°C for [5] minutes                 │
│   Severity: [Critical ▼]                                       │
├─────────────────────────────────────────────────────────────────┤
│ 📬 Notification Channels                             [+ Add]   │
├─────────────────────────────────────────────────────────────────┤
│ ✅ Slack: #gpu-alerts                              [Test] [Edit]│
│ ✅ Email: infra-team@company.com                   [Test] [Edit]│
└─────────────────────────────────────────────────────────────────┘
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Control Node                               │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │  AAMI CLI   │  │  AAMI UI    │  │  SSH Executor (Go)      │ │
│  │             │  │  (Web)      │  │  - Parallel (100 conc.) │ │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬────────────┘ │
│         └────────┬───────┘                      │              │
│                  ▼                              │              │
│  ┌─────────────────────────────────┐            │              │
│  │        config.yaml              │            │              │
│  │   (File-based, no DB)           │            │              │
│  └─────────────────────────────────┘            │              │
│                  │                              │              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Prometheus + Alertmanager + Grafana (Container/Binary) │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────────────┬────────────────────────────┘
                                     │ SSH (Install) / HTTP (Metrics)
        ┌────────────────────────────┼────────────────────────────┐
        ▼                            ▼                            ▼
   ┌──────────┐                ┌──────────┐                ┌──────────┐
   │ GPU Node │                │ GPU Node │                │ GPU Node │
   │    01    │                │    02    │       ...      │    N     │
   ├──────────┤                ├──────────┤                ├──────────┤
   │• node    │                │• node    │                │• node    │
   │  exporter│                │  exporter│                │  exporter│
   │• dcgm    │                │• dcgm    │                │• dcgm    │
   │  exporter│                │  exporter│                │  exporter│
   └──────────┘                └──────────┘                └──────────┘
```

### Design Decisions

| Area | Choice | Reason |
|------|--------|--------|
| Node Access | SSH (Agentless) | Air-gap friendly, no additional agent required |
| Data Storage | YAML File | Simplified installation without DB, Git version control |
| GPU Metrics | DCGM Exporter | Official NVIDIA, detailed metrics |

## Quick Start

### Prerequisites

- Control Node: Linux (Ubuntu 20.04+, RHEL 8+)
- GPU Nodes: SSH accessible, NVIDIA Driver 450.80+
- Optional: Docker or Podman (for container deployment)

### Installation

```bash
# 1. Install AAMI
curl -fsSL https://get.aami.dev | bash

# 2. Initialize
aami init

# 3. Register nodes
cat << EOF > hosts.txt
gpu-node-01 192.168.1.101
gpu-node-02 192.168.1.102
gpu-node-03 192.168.1.103
EOF

aami nodes add --file hosts.txt --user root --key ~/.ssh/id_rsa

# 4. Apply alert preset
aami alerts apply-preset gpu-production

# 5. Configure notifications
aami config notifications slack --webhook https://hooks.slack.com/xxx

# 6. Check status
aami status
```

### Air-gap Installation

```bash
# Create bundle on internet-connected machine
aami bundle create --output aami-offline-v1.0.0.tar.gz

# Install on air-gapped machine
aami init --offline ./aami-offline-v1.0.0.tar.gz
```

## Configuration

```yaml
# /etc/aami/config.yaml

cluster:
  name: gpu-cluster-prod

nodes:
  - name: gpu-node-01
    ip: 192.168.1.101
    ssh_user: root
    ssh_key: /root/.ssh/id_rsa
    labels:
      gpu_type: a100

alerts:
  presets:
    - gpu-production

notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#gpu-alerts"

prometheus:
  retention: 15d
  storage_path: /var/lib/aami/prometheus
```

## Comparison

| Feature | AAMI | kube-prometheus-stack | Ansible + Prometheus | Zabbix |
|---------|------|----------------------|---------------------|--------|
| K8s Required | ❌ Not required | ✅ Required | ❌ Not required | ❌ Not required |
| Installation Time | 30 min | 10 min (with K8s) | 2-3 days | Half day |
| Air-gap | ✅ Bundle provided | ⚠️ Image mirroring | ⚠️ Manual setup | ⚠️ Manual setup |
| GPU Native | ✅ DCGM included | ⚠️ Separate install | ⚠️ Separate install | ❌ Custom required |
| Xid Interpretation | ✅ Built-in | ❌ None | ❌ None | ❌ None |
| Operations CLI | ✅ Built-in | ❌ kubectl | ❌ ansible-playbook | ❌ None |

## Technology Stack

| Area | Technology |
|------|------------|
| CLI | Go 1.21+ (single binary) |
| Monitoring | Prometheus, Grafana, Alertmanager |
| GPU Metrics | DCGM Exporter (NVIDIA), ROCm Exporter (AMD, planned) |
| Configuration Storage | YAML (No DB) |
| Node Communication | SSH (Agentless) |
| Large Scale | Prometheus Federation |
| Scheduler Integration | Slurm |

## Project Structure

```
aami/
├── cmd/                    # Application entrypoints
├── internal/               # Core packages
│   ├── cli/                # CLI commands
│   ├── config/             # Configuration management
│   ├── ssh/                # SSH executor
│   ├── installer/          # Component installers
│   ├── xid/                # Xid error interpretation
│   ├── health/             # GPU health scoring
│   ├── nvlink/             # NVLink topology
│   ├── federation/         # Prometheus federation
│   ├── slurm/              # Slurm integration
│   ├── multicluster/       # Multi-cluster management
│   ├── backup/             # Backup & restore
│   └── upgrade/            # Upgrade management
├── configs/                # Default configuration templates
├── docs/                   # Documentation
├── examples/               # Examples
├── scripts/                # Installation/utility scripts
└── deploy/
    └── offline/            # Air-gap bundles
```

## Roadmap

### Phase 1: MVP ✅
- [x] One-click installation (`aami init`)
- [x] Air-gap bundler (`aami bundle`)
- [x] Node management CLI (`aami nodes`)
- [x] Alert presets (`aami alerts`)
- [x] Xid interpretation (`aami explain xid`)

### Phase 2: Enhancement ✅
- [x] NVLink topology visualization
- [x] GPU Health Score
- [x] Upgrade/Backup
- [x] Operations tools

### Phase 3: Scale ✅
- [x] Prometheus Federation (1k+ nodes)
- [x] Slurm integration (Job-GPU correlation)
- [x] Multi-cluster management

### Phase 4: AMD GPU Support (Planned)
- [ ] ROCm exporter integration
- [ ] AMD error code interpretation
- [ ] Unified alert rules for NVIDIA/AMD

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/fregataa/aami/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fregataa/aami/discussions)
