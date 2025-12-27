# Configuration Files

This directory contains configuration files for monitoring components.

## Directory Structure

```
config/
├── prometheus/          # Prometheus configuration
├── grafana/            # Grafana dashboards and datasources
└── alertmanager/       # Alertmanager configuration
```

## Components

### Prometheus
- **Location**: `prometheus/`
- **Contents**:
  - `prometheus.yml` - Main Prometheus configuration
  - `rules/*.yml` - Alert rules
  - `sd/*.json` - Service discovery files (dynamic)

Example prometheus.yml:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'node-exporter'
    file_sd_configs:
      - files:
          - '/etc/prometheus/sd/node-exporter.json'
```

### Grafana
- **Location**: `grafana/`
- **Contents**:
  - `dashboards/*.json` - Dashboard definitions
  - `datasources.yml` - Data source configurations
  - `grafana.ini` - Grafana server configuration

Dashboard categories:
- System metrics (CPU, Memory, Disk, Network)
- GPU/NPU metrics (DCGM, custom exporters)
- Application metrics
- Alert overview

### Alertmanager
- **Location**: `alertmanager/`
- **Contents**:
  - `alertmanager.yml` - Alertmanager configuration
  - `templates/*.tmpl` - Notification templates

Supported notification channels:
- Email (SMTP)
- Slack
- Webhook
- PagerDuty

## Configuration Management

Configurations can be managed through:

1. **Static Files**: Edit files directly (requires restart)
2. **Config Server API**: Dynamic updates via REST API
3. **GitOps**: Version control and automated deployment

## Environment Variables

Common environment variables used in configurations:

- `PROMETHEUS_PORT` - Prometheus server port (default: 9090)
- `GRAFANA_PORT` - Grafana server port (default: 3000)
- `ALERTMANAGER_PORT` - Alertmanager port (default: 9093)
- `SMTP_HOST` - SMTP server for email alerts
- `SLACK_WEBHOOK` - Slack webhook URL for notifications

## Quick Links

- [Configuration Guide](../docs/configuration.md)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
