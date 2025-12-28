# Node Exporter Textfile Collector

This directory contains configuration and documentation for Node Exporter's textfile collector feature, which enables custom metrics collection through the AAMI dynamic check system.

## Overview

The textfile collector allows external scripts to expose custom metrics by writing Prometheus-formatted metrics to `.prom` files in a designated directory. Node Exporter reads these files and includes them in its metrics endpoint.

**Key Features**:
- **Dynamic Checks**: Checks are managed centrally in AAMI Config Server
- **Automatic Updates**: Scripts update automatically based on hash versioning
- **Scope-based Configuration**: Different checks for different groups/namespaces
- **Flexible Scheduling**: Run checks via cron or systemd timers

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  AAMI Config Server                                         │
│  - CheckTemplates (reusable check definitions)             │
│  - CheckInstances (scope-specific applications)            │
│  - API: /api/v1/checks/node/:hostname                      │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ HTTP GET (every 1 min)
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Node: dynamic-check.sh                                     │
│  1. Fetch effective checks for this node                   │
│  2. Save/update check scripts (hash-based versioning)      │
│  3. Execute checks with node-specific config               │
│  4. Write results to textfile collector directory          │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ Write .prom files
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  /var/lib/node_exporter/textfile_collector/*.prom          │
│  - mount_check.prom                                         │
│  - disk_smart.prom                                          │
│  - infiniband.prom                                          │
│  - nvme_health.prom                                         │
│  - parallel_fs.prom                                         │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ Read every scrape
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Node Exporter                                              │
│  --collector.textfile.directory=/var/lib/.../textfile...   │
│  Exposes metrics at :9100/metrics                           │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ Scrape
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Prometheus                                                 │
│  Stores and queries all metrics                             │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### 1. Install Node Exporter with Textfile Collector

Use the provided installation script:

```bash
# Download and run the installer
sudo bash /path/to/aami/scripts/node/install-node-exporter.sh

# Or with custom options
sudo bash /path/to/aami/scripts/node/install-node-exporter.sh \
    --version 1.6.1 \
    --port 9100
```

This will:
- Install Node Exporter as a systemd service
- Enable textfile collector at `/var/lib/node_exporter/textfile_collector`
- Create necessary directories with proper permissions

### 2. Install Dynamic Check Script

```bash
# Copy dynamic check script
sudo cp scripts/node/dynamic-check.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/dynamic-check.sh

# Create configuration file
sudo mkdir -p /etc/aami
sudo tee /etc/aami/config <<EOF
AAMI_CONFIG_SERVER_URL="http://config-server:8080"
AAMI_HOSTNAME="$(hostname)"
EOF
```

### 3. Set Up Scheduling

Choose either cron or systemd timer:

#### Option A: Cron (Simpler)

```bash
# Install cron job
sudo cp config/node-exporter/cron.d/aami-dynamic-check /etc/cron.d/
sudo chmod 0644 /etc/cron.d/aami-dynamic-check

# Verify
sudo crontab -l
```

#### Option B: Systemd Timer (Recommended)

```bash
# Copy service and timer units
sudo cp config/node-exporter/systemd/aami-dynamic-check.service /etc/systemd/system/
sudo cp config/node-exporter/systemd/aami-dynamic-check.timer /etc/systemd/system/

# Update Config Server URL in service file
sudo vi /etc/systemd/system/aami-dynamic-check.service
# Edit: Environment="AAMI_CONFIG_SERVER_URL=http://your-config-server:8080"

# Reload systemd and enable timer
sudo systemctl daemon-reload
sudo systemctl enable aami-dynamic-check.timer
sudo systemctl start aami-dynamic-check.timer

# Check status
sudo systemctl status aami-dynamic-check.timer
sudo systemctl list-timers aami-dynamic-check.timer

# View logs
sudo journalctl -u aami-dynamic-check.service -f
```

### 4. Verify Installation

```bash
# Test dynamic check script manually
sudo /usr/local/bin/dynamic-check.sh --debug

# Check textfile collector directory
ls -la /var/lib/node_exporter/textfile_collector/

# View metrics
curl http://localhost:9100/metrics | grep -E '(mount_check|disk_smart|infiniband|nvme|parallel_fs|aami_)'

# Check logs
tail -f /var/log/aami/dynamic-check.log
```

## Custom Check Development

### Check Script Format

Check scripts must follow this format:

1. **Input**: JSON configuration via stdin or first argument
2. **Output**: Prometheus-formatted metrics to stdout
3. **Error Handling**: Exit with non-zero status on failure

Example check script:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Read JSON config
if [ $# -eq 0 ]; then
    config=$(cat)
else
    config="$1"
fi

# Extract parameters
param=$(echo "$config" | jq -r '.param_name')

# Output Prometheus metrics
echo "# HELP my_custom_metric Description of metric"
echo "# TYPE my_custom_metric gauge"
echo "my_custom_metric{label=\"value\"} 1"
```

### Prometheus Metric Format

```
# HELP metric_name Description of what this metric measures
# TYPE metric_name <gauge|counter|histogram|summary>
metric_name{label1="value1",label2="value2"} <numeric_value>
```

**Metric Types**:
- `gauge`: Value that can go up or down (e.g., temperature, count)
- `counter`: Value that only increases (e.g., bytes transferred, requests)
- `histogram`: Observations bucketed by value range
- `summary`: Observations with quantiles

**Best Practices**:
- Use lowercase names with underscores: `my_metric_name`
- Include units in names: `_seconds`, `_bytes`, `_total`
- Use labels for dimensions: `{device="sda", status="healthy"}`
- Keep label cardinality low (avoid unique IDs in labels)

### Creating a CheckTemplate

Via Config Server API:

```bash
curl -X POST http://config-server:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my_custom_check",
    "description": "My custom health check",
    "script_content": "#!/bin/bash\necho \"# HELP my_metric My metric\"\necho \"# TYPE my_metric gauge\"\necho \"my_metric 1\"",
    "default_config": {
      "param1": "value1",
      "timeout_seconds": 30
    }
  }'
```

### Applying a CheckInstance

```bash
# Apply to a specific group
curl -X POST http://config-server:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 123,
    "scope": "group",
    "group_id": 456,
    "config": {
      "param1": "custom_value"
    }
  }'
```

**Scope Priority**: Group > Namespace > Global

## Built-in Checks

AAMI provides several built-in check scripts:

### 1. Mount Point Check

**Script**: `check-mount-points.sh`
**Purpose**: Verify mount point accessibility and writability
**Metrics**: `mount_check{path}`

**Configuration**:
```json
{
  "mount_points": ["/mnt/data", "/mnt/backup", "/mnt/scratch"]
}
```

### 2. Disk SMART Check

**Script**: `check-disk-smart.sh`
**Purpose**: Monitor disk health via SMART
**Metrics**:
- `disk_smart_health{device}`
- `disk_smart_temperature_celsius{device}`
- `disk_smart_reallocated_sectors{device}`

**Requirements**: `smartmontools` package

**Configuration**:
```json
{
  "devices": ["/dev/sda", "/dev/sdb"]
}
```

### 3. InfiniBand Check

**Script**: `check-infiniband.sh`
**Purpose**: Monitor InfiniBand link state and throughput
**Metrics**:
- `infiniband_link_state{device,port}`
- `infiniband_link_rate_gbps{device,port}`
- `infiniband_port_rcv_data_bytes{device,port}`
- `infiniband_port_xmit_data_bytes{device,port}`

**Configuration**:
```json
{
  "devices": ["mlx5_0", "mlx5_1"]
}
```

### 4. NVMe Health Check

**Script**: `check-nvme-health.sh`
**Purpose**: Monitor NVMe SSD health
**Metrics**:
- `nvme_health{device}`
- `nvme_temperature_celsius{device}`
- `nvme_available_spare_percent{device}`
- `nvme_percentage_used{device}`

**Requirements**: `nvme-cli` package

**Configuration**:
```json
{
  "devices": ["/dev/nvme0n1", "/dev/nvme1n1"]
}
```

### 5. Parallel Filesystem Check

**Script**: `check-parallel-fs.sh`
**Purpose**: Monitor parallel filesystem latency and throughput
**Metrics**:
- `parallel_fs_accessible{path,type}`
- `parallel_fs_latency_seconds{path,type,operation}`
- `parallel_fs_throughput_mbps{path,type,operation}`

**Configuration**:
```json
{
  "filesystems": [
    {"path": "/mnt/lustre", "type": "lustre"},
    {"path": "/mnt/gpfs", "type": "gpfs"},
    {"path": "/mnt/beegfs", "type": "beegfs"}
  ],
  "timeout_seconds": 5,
  "test_size_mb": 1
}
```

## Troubleshooting

### Check Script Not Running

```bash
# Check timer status
sudo systemctl status aami-dynamic-check.timer
sudo systemctl list-timers

# Check service status
sudo systemctl status aami-dynamic-check.service

# View logs
sudo journalctl -u aami-dynamic-check.service -n 50

# Test manually
sudo /usr/local/bin/dynamic-check.sh --debug
```

### No Metrics in Prometheus

```bash
# Check Node Exporter is running
sudo systemctl status node_exporter

# Check textfile collector directory
ls -la /var/lib/node_exporter/textfile_collector/

# View a metric file
cat /var/lib/node_exporter/textfile_collector/mount_check.prom

# Test Node Exporter endpoint
curl http://localhost:9100/metrics | grep mount_check
```

### Config Server Connection Failed

```bash
# Check config
cat /etc/aami/config

# Test connection
curl -v http://config-server:8080/api/v1/checks/node/$(hostname)

# Check network
ping config-server
telnet config-server 8080
```

### Check Script Permissions

```bash
# Check script ownership
ls -la /usr/local/lib/aami/checks/

# Fix permissions
sudo chown -R root:root /usr/local/lib/aami/checks/
sudo chmod -R 755 /usr/local/lib/aami/checks/

# Check textfile directory permissions
ls -la /var/lib/node_exporter/textfile_collector/
sudo chown -R node_exporter:node_exporter /var/lib/node_exporter/
```

## Performance Considerations

- **Check Frequency**: Default 1 minute is suitable for most checks
- **Timeout**: Each check has a 30-second timeout to prevent hangs
- **Script Caching**: Scripts are cached locally with hash-based versioning
- **Atomic Updates**: Metrics are written to temp files then moved atomically
- **Log Rotation**: Configure logrotate for `/var/log/aami/dynamic-check.log`

Example logrotate configuration:

```
/var/log/aami/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
```

## Security Considerations

- Check scripts run as root (required for hardware access)
- Use systemd security features: `ProtectSystem`, `ProtectHome`, `PrivateTmp`
- Validate all external inputs in check scripts
- Use hash-based script versioning to detect tampering
- Restrict network access to Config Server only
- Audit check script contents before deployment

## References

- [Prometheus Textfile Collector](https://github.com/prometheus/node_exporter#textfile-collector)
- [Prometheus Metric Naming](https://prometheus.io/docs/practices/naming/)
- [Writing Exporters](https://prometheus.io/docs/instrumenting/writing_exporters/)
- [AAMI Config Server API](/docs/en/API.md)
