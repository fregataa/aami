# Quick Start Guide

This guide will walk you through setting up AAMI from scratch and registering your first monitoring targets.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Initial Configuration](#initial-configuration)
4. [Creating Your First Group](#creating-your-first-group)
5. [Registering Targets](#registering-targets)
6. [Setting Up Alerts](#setting-up-alerts)
7. [Viewing Metrics](#viewing-metrics)
8. [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have:

- Docker 20.10+ and Docker Compose v2.0+
- At least 4GB RAM and 20GB disk space
- Network access to target nodes (for monitoring)
- Basic understanding of Prometheus and Grafana

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/your-org/aami.git
cd aami
```

### Step 2: Configure Environment

```bash
cd deploy/docker-compose
cp .env.example .env
```

Edit the `.env` file with your settings:

```env
# PostgreSQL Configuration
POSTGRES_USER=admin
POSTGRES_PASSWORD=changeme
POSTGRES_DB=config_server

# Redis Configuration
REDIS_PASSWORD=

# Config Server Configuration
CONFIG_SERVER_PORT=8080

# Grafana Configuration
GRAFANA_ADMIN_PASSWORD=admin
```

### Step 3: Start the Stack

```bash
docker-compose up -d
```

Wait for all services to start (this may take 1-2 minutes):

```bash
docker-compose ps
```

You should see all services in "Up" state:
- prometheus
- grafana
- alertmanager
- config-server
- postgres
- redis

### Step 4: Verify Installation

Check that all services are accessible:

```bash
# Config Server health check
curl http://localhost:8080/api/v1/health

# Expected output:
# {"status":"ok","timestamp":"2024-01-01T00:00:00Z"}

# Prometheus
curl http://localhost:9090/-/healthy

# Grafana (should return HTML)
curl -I http://localhost:3000
```

## Initial Configuration

### Step 1: Access Grafana

1. Open your browser and navigate to http://localhost:3000
2. Login with default credentials:
   - Username: `admin`
   - Password: `admin` (or the password you set in .env)
3. You'll be prompted to change the password (optional in dev environment)

### Step 2: Verify Prometheus Data Source

1. Go to **Configuration** → **Data Sources**
2. You should see a Prometheus data source already configured
3. Click **Test** to verify connectivity

## Creating Your First Group

Groups organize your infrastructure hierarchically. Let's create a basic structure.

### Step 1: Create Infrastructure Group

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "namespace": "environment",
    "parent_id": null,
    "description": "Production environment"
  }'
```

Save the returned `group_id` for the next steps.

### Step 2: Create Subgroup

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpu-cluster",
    "namespace": "infrastructure",
    "parent_id": "PARENT_GROUP_ID",
    "description": "GPU compute cluster"
  }'
```

### Step 3: Verify Groups

```bash
curl http://localhost:8080/api/v1/groups
```

## Registering Targets

Now let's register monitoring targets (servers to monitor).

### Method 1: Manual Registration via API

```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "gpu-node-01.example.com",
    "ip_address": "10.0.1.10",
    "primary_group_id": "GROUP_ID_HERE",
    "exporters": [
      {
        "type": "node_exporter",
        "port": 9100,
        "enabled": true,
        "scrape_interval": "15s",
        "scrape_timeout": "10s"
      },
      {
        "type": "dcgm_exporter",
        "port": 9400,
        "enabled": true,
        "scrape_interval": "30s"
      }
    ],
    "labels": {
      "datacenter": "dc1",
      "rack": "r1",
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

### Method 2: Bootstrap Script (Recommended)

On your target node, run:

```bash
curl -fsSL https://your-config-server:8080/bootstrap.sh | \
  bash -s -- \
    --token YOUR_BOOTSTRAP_TOKEN \
    --server https://your-config-server:8080
```

This will:
1. Auto-detect hardware (CPU, GPU, memory)
2. Install appropriate exporters (node_exporter, dcgm_exporter)
3. Register itself with the Config Server
4. Start exporting metrics

### Step 3: Verify Target Registration

```bash
# List all targets
curl http://localhost:8080/api/v1/targets

# Check specific target
curl http://localhost:8080/api/v1/targets/TARGET_ID
```

### Step 4: Verify Prometheus Service Discovery

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Or visit in browser
open http://localhost:9090/targets
```

You should see your registered targets appear within 30 seconds.

## Setting Up Alerts

### Step 1: List Available Alert Rule Templates

```bash
curl http://localhost:8080/api/v1/alert-templates
```

Common templates include:
- `HighCPUUsage`
- `HighMemoryUsage`
- `DiskSpaceLow`
- `NodeDown`
- `GPUHighTemperature`

### Step 2: Apply Alert Rule to Group

```bash
curl -X POST http://localhost:8080/api/v1/groups/GROUP_ID/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "rule_template_id": "HighCPUUsage",
    "enabled": true,
    "config": {
      "threshold": 80,
      "duration": "5m"
    },
    "merge_strategy": "override"
  }'
```

### Step 3: Verify Alert Rules

```bash
# Check effective rules for a target
curl http://localhost:8080/api/v1/targets/TARGET_ID/alert-rules/effective

# Trace policy inheritance
curl http://localhost:8080/api/v1/targets/TARGET_ID/alert-rules/trace
```

### Step 4: Configure Alertmanager

Edit `config/alertmanager/alertmanager.yml`:

```yaml
route:
  receiver: 'default-receiver'
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

receivers:
  - name: 'default-receiver'
    email_configs:
      - to: 'alerts@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alerts@example.com'
        auth_password: 'your-password'
```

Reload Alertmanager:

```bash
docker-compose restart alertmanager
```

## Viewing Metrics

### Step 1: Access Grafana Dashboards

1. Go to http://localhost:3000
2. Navigate to **Dashboards** → **Browse**
3. Import pre-built dashboards from `config/grafana/dashboards/`

### Step 2: Explore Prometheus Metrics

Visit http://localhost:9090 and try these queries:

**Node Metrics:**
```promql
# CPU usage
100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory usage
100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))

# Disk usage
100 - ((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes)
```

**GPU Metrics (DCGM):**
```promql
# GPU utilization
DCGM_FI_DEV_GPU_UTIL

# GPU temperature
DCGM_FI_DEV_GPU_TEMP

# GPU memory usage
DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100
```

### Step 3: Create Custom Dashboard

1. In Grafana, click **+ → Dashboard**
2. Add a new panel
3. Select Prometheus as data source
4. Enter a PromQL query
5. Configure visualization
6. Save dashboard

## Next Steps

Congratulations! You now have a working AAMI installation. Here's what to do next:

### Expand Your Monitoring

1. **Add More Targets**: Register additional nodes
2. **Create Group Hierarchy**: Organize by datacenter, environment, or function
3. **Customize Alerts**: Fine-tune thresholds per group
4. **Deploy Custom Exporters**: Monitor specialized hardware

### Advanced Configuration

- [API Documentation](./API.md) - Full REST API reference
- [Deployment Guide](../../deploy/README.md) - Production deployment
- [Alert Rules Guide](./ALERT_RULES.md) - Advanced alert configuration
- [Dashboard Guide](./DASHBOARDS.md) - Creating custom dashboards

### Automation

- [Bootstrap Script](../../scripts/node/README.md) - Automated agent deployment
- [Terraform Examples](../../examples/terraform/) - Infrastructure as Code
- [Ansible Playbooks](../../deploy/ansible/) - Configuration management

### Troubleshooting

If you encounter issues:

1. Check logs: `docker-compose logs -f SERVICE_NAME`
2. Verify connectivity: `docker-compose ps`
3. Check Config Server: `curl http://localhost:8080/api/v1/health`
4. See [Troubleshooting Guide](./TROUBLESHOOTING.md)

## Common Issues

### Targets Not Appearing in Prometheus

**Problem**: Targets registered via API but not showing in Prometheus

**Solution**:
```bash
# Check service discovery file
curl http://localhost:8080/api/v1/sd/prometheus

# Restart Prometheus to reload config
docker-compose restart prometheus
```

### Exporters Not Responding

**Problem**: Cannot reach node_exporter on target node

**Solution**:
```bash
# On target node, check if exporter is running
systemctl status node_exporter

# Check firewall
sudo ufw status
sudo ufw allow 9100/tcp

# Test locally on target
curl http://localhost:9100/metrics
```

### Alert Rules Not Working

**Problem**: Alerts not firing despite conditions being met

**Solution**:
```bash
# Check Prometheus rules
curl http://localhost:9090/api/v1/rules

# Verify rule evaluation
# Open http://localhost:9090/alerts

# Check Alertmanager
curl http://localhost:9093/api/v2/alerts
```

## Clean Up

To remove the entire stack:

```bash
cd deploy/docker-compose

# Stop all services
docker-compose down

# Remove all data (including databases)
docker-compose down -v
```

---

**Need Help?**
- [GitHub Issues](https://github.com/your-org/aami/issues)
- [Documentation](../../README.md)
- [Community Discussions](https://github.com/your-org/aami/discussions)
