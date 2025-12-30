# Prometheus Integration Guide

This guide covers the integration between AAMI Config Server and Prometheus for dynamic alert rule management.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Setup](#setup)
4. [Configuration](#configuration)
5. [API Reference](#api-reference)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

---

## Overview

AAMI Config Server integrates with Prometheus to provide dynamic alert rule management. This integration enables:

- **Dynamic Rule Generation**: AlertRules defined in the database are automatically converted to Prometheus rule files
- **Group-based Customization**: Different alert thresholds per group using label-based filtering
- **Zero-downtime Updates**: Prometheus rules are reloaded without service interruption
- **Centralized Management**: Manage all alert rules through a single API

### Key Components

| Component | Description |
|-----------|-------------|
| **Prometheus Rule Generator** | Converts AlertRules to Prometheus YAML format |
| **Rule File Manager** | Handles atomic file writes with validation and backup |
| **Prometheus Client** | Triggers configuration reload and health checks |
| **AlertRule API** | CRUD operations for group-specific alert rules |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     AAMI Config Server                          │
│  ┌─────────────────┐  ┌────────────────┐  ┌─────────────────┐  │
│  │  AlertRule API  │  │ Rule Generator │  │ Prometheus Client│  │
│  │  (CRUD)         │→ │ (YAML gen)     │→ │ (Reload)        │  │
│  └─────────────────┘  └────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                       │                    │
         ▼                       ▼                    ▼
┌─────────────┐         ┌─────────────────┐   ┌─────────────────┐
│  PostgreSQL │         │ Shared Volume   │   │   Prometheus    │
│  (AlertRule │         │ /rules/generated│   │   /-/reload     │
│   storage)  │         │ *.yml files     │   │   /-/ready      │
└─────────────┘         └─────────────────┘   └─────────────────┘
```

### Data Flow

1. User creates/updates AlertRule via API
2. Rule Generator converts AlertRule to Prometheus YAML
3. File Manager writes rule file atomically (temp → rename)
4. Prometheus Client triggers `/-/reload` endpoint
5. Prometheus loads new rules without restart

---

## Setup

### Prerequisites

- Prometheus 2.x+ with `--web.enable-lifecycle` flag
- Shared volume between Config Server and Prometheus
- Network access from Config Server to Prometheus API

### Docker Compose Setup

```yaml
# docker-compose.yml
version: '3.8'

volumes:
  prometheus-rules:

services:
  prometheus:
    image: prom/prometheus:v2.48.0
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--web.enable-lifecycle'  # Required for reload API
    volumes:
      - prometheus-rules:/etc/prometheus/rules/generated:ro
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    ports:
      - "9090:9090"

  config-server:
    image: aami/config-server:latest
    volumes:
      - prometheus-rules:/app/rules
    environment:
      PROMETHEUS_URL: http://prometheus:9090
      PROMETHEUS_RULE_PATH: /app/rules
      PROMETHEUS_RELOAD_ENABLED: "true"
    depends_on:
      - prometheus
```

### Prometheus Configuration

Add generated rules directory to Prometheus config:

```yaml
# config/prometheus/prometheus.yml
rule_files:
  - /etc/prometheus/rules/*.yml           # Static rules
  - /etc/prometheus/rules/generated/*.yml # Dynamic rules from Config Server
```

### Kubernetes Setup

```yaml
# ConfigMap for Prometheus
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    rule_files:
      - /etc/prometheus/rules/generated/*.yml

---
# PersistentVolumeClaim for shared rules
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: prometheus-rules-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi

---
# Mount in both Prometheus and Config Server deployments
# Prometheus: /etc/prometheus/rules/generated (read-only)
# Config Server: /app/rules (read-write)
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PROMETHEUS_URL` | `http://localhost:9090` | Prometheus server URL |
| `PROMETHEUS_RULE_PATH` | `/etc/prometheus/rules/generated` | Directory for generated rule files |
| `PROMETHEUS_RELOAD_ENABLED` | `true` | Enable automatic Prometheus reload |
| `PROMETHEUS_RELOAD_TIMEOUT` | `30s` | Timeout for reload API call |
| `PROMETHEUS_VALIDATE_RULES` | `false` | Enable promtool validation before write |
| `PROMETHEUS_BACKUP_ENABLED` | `true` | Keep backup of previous rule files |
| `PROMTOOL_PATH` | `promtool` | Path to promtool binary (if validation enabled) |

### Config File (config.yaml)

```yaml
prometheus:
  url: "http://prometheus:9090"
  rule_path: "/app/rules"
  reload_enabled: true
  reload_timeout: 30s
  validate_rules: false
  backup_enabled: true
```

---

## API Reference

### Prometheus Rule Management Endpoints

#### Regenerate All Rules

Regenerates Prometheus rule files for all groups.

```bash
POST /api/v1/prometheus/rules/regenerate
```

**Response:**
```json
{
  "groups_affected": 5,
  "files_generated": 5,
  "errors": [],
  "duration": "1.234s"
}
```

#### Regenerate Group Rules

Regenerates Prometheus rule file for a specific group.

```bash
POST /api/v1/prometheus/rules/regenerate/:group_id
```

**Response:**
```json
{
  "group_id": "grp-123",
  "file_name": "group-grp-123.yml",
  "rules_count": 3,
  "duration": "0.456s"
}
```

#### List Rule Files

Lists all generated Prometheus rule files.

```bash
GET /api/v1/prometheus/rules/files
```

**Response:**
```json
{
  "files": [
    {
      "group_id": "grp-123",
      "file_name": "group-grp-123.yml",
      "rule_count": 3,
      "size_bytes": 1024,
      "modified_at": "2025-01-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

#### Get Effective Rules for Target

Returns all effective alert rules for a specific target, considering group membership.

```bash
GET /api/v1/prometheus/rules/effective/:target_id
```

**Response:**
```json
{
  "target_id": "target-456",
  "hostname": "gpu-node-01",
  "rules": [
    {
      "id": "rule-789",
      "name": "HighCPUUsage",
      "severity": "warning",
      "query": "cpu_usage{group_id=\"grp-123\"} > 80",
      "for_duration": "5m",
      "labels": {
        "group_id": "grp-123"
      },
      "annotations": {
        "summary": "High CPU usage detected"
      },
      "config": {
        "threshold": 80
      },
      "source": "group",
      "source_id": "grp-123",
      "source_name": "production"
    }
  ],
  "total": 1
}
```

#### Trigger Prometheus Reload

Manually triggers Prometheus configuration reload.

```bash
POST /api/v1/prometheus/reload
```

**Response:**
```json
{
  "status": "success",
  "message": "Prometheus reload triggered successfully"
}
```

#### Get Prometheus Status

Checks Prometheus connection status.

```bash
GET /api/v1/prometheus/status
```

**Response:**
```json
{
  "status": "healthy",
  "url": "http://prometheus:9090",
  "ready": true,
  "healthy": true
}
```

---

## Troubleshooting

### Common Issues

#### 1. Rules Not Loading in Prometheus

**Symptoms:**
- AlertRules created in API but not visible in Prometheus
- `/api/v1/prometheus/rules/files` shows files but Prometheus doesn't see them

**Solutions:**

1. Check volume mounts:
```bash
# In Config Server container
ls -la /app/rules/

# In Prometheus container
ls -la /etc/prometheus/rules/generated/
```

2. Verify Prometheus config:
```bash
curl http://prometheus:9090/api/v1/status/config | jq '.data.yaml' | grep rule_files
```

3. Check Prometheus reload:
```bash
curl -X POST http://prometheus:9090/-/reload
```

4. Look for Prometheus errors:
```bash
docker logs prometheus 2>&1 | grep -i "rule\|error"
```

#### 2. Reload API Not Working

**Symptoms:**
- Error: "Lifecycle API is not enabled"

**Solution:**
Ensure Prometheus is started with `--web.enable-lifecycle` flag:

```yaml
# docker-compose.yml
prometheus:
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--web.enable-lifecycle'
```

#### 3. Rule Validation Failures

**Symptoms:**
- Rule generation succeeds but Prometheus ignores the file
- Syntax errors in generated YAML

**Solutions:**

1. Enable rule validation in Config Server:
```bash
PROMETHEUS_VALIDATE_RULES=true
PROMTOOL_PATH=/usr/local/bin/promtool
```

2. Manually validate rule file:
```bash
promtool check rules /path/to/rule-file.yml
```

3. Check for common issues:
   - Invalid PromQL syntax
   - Missing required fields (expr, alert)
   - YAML indentation errors

#### 4. Permission Issues

**Symptoms:**
- "Permission denied" when writing rule files
- Empty rules directory after generation

**Solutions:**

1. Check directory permissions:
```bash
ls -la /app/rules/
```

2. Ensure Config Server process has write access:
```yaml
# docker-compose.yml
config-server:
  user: "1000:1000"  # Match volume owner
```

3. For Kubernetes, use proper SecurityContext:
```yaml
securityContext:
  runAsUser: 1000
  fsGroup: 1000
```

#### 5. Prometheus Not Connecting

**Symptoms:**
- `/api/v1/prometheus/status` returns error
- Reload calls fail

**Solutions:**

1. Verify network connectivity:
```bash
curl http://prometheus:9090/-/ready
```

2. Check environment variables:
```bash
echo $PROMETHEUS_URL
```

3. Ensure Prometheus is healthy:
```bash
curl http://prometheus:9090/-/healthy
```

### Logs to Check

1. **Config Server logs:**
```bash
docker logs config-server 2>&1 | grep -i prometheus
```

2. **Prometheus logs:**
```bash
docker logs prometheus 2>&1 | grep -E "(rule|reload|error)"
```

---

## Best Practices

### 1. Use Atomic Rule Updates

Always use the regenerate API instead of manually editing files. This ensures:
- Atomic file writes (no partial writes)
- Automatic backup of previous version
- Proper Prometheus reload

### 2. Monitor Rule Generation

Set up alerting for rule generation failures:

```yaml
- alert: AlertRuleGenerationFailed
  expr: aami_rule_generation_errors_total > 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Alert rule generation failed"
```

### 3. Use Group-specific Thresholds

Leverage the group label for environment-specific rules:

```yaml
# AlertTemplate query_template
(100 - avg(rate(node_cpu_seconds_total{mode="idle",group_id="{{.group_id}}"}[5m])) * 100) > {{.threshold}}
```

### 4. Test Rules Before Production

1. Create rules in development group first
2. Verify in Prometheus UI (`/alerts`)
3. Check PromQL expression (`/graph`)
4. Promote to production group

### 5. Keep Backups

Enable backup in configuration:

```bash
PROMETHEUS_BACKUP_ENABLED=true
```

This creates `.bak` files before overwriting rule files.

### 6. Use Meaningful Alert Names

Follow naming conventions:
- `HighCPUUsage_Production` - Include group context
- Use consistent prefixes for related alerts
- Include severity in labels, not name

### 7. Separate Rule Files by Group

Config Server automatically generates one file per group:
```
/rules/generated/
  group-grp-123.yml    # Production alerts
  group-grp-456.yml    # Development alerts
```

This isolation prevents one group's issues from affecting others.

---

## References

- [AAMI Alerting System Architecture](./ALERTING-SYSTEM.md)
- [Prometheus Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [Prometheus Alerting Rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [Prometheus Management API](https://prometheus.io/docs/prometheus/latest/management_api/)
