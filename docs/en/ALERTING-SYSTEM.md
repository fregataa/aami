# Alerting System Architecture

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Components](#components)
4. [Alert Types](#alert-types)
5. [Data Flow](#data-flow)
6. [Group-based Customization](#group-based-customization)
7. [Alert Rule Generation](#alert-rule-generation)
8. [Integration Points](#integration-points)
9. [Examples](#examples)
10. [FAQ](#faq)

---

## Overview

AAMI's alerting system provides comprehensive monitoring and notification capabilities for AI accelerator infrastructure. The system is built on Prometheus and Alertmanager, offering a unified alert path that handles both standard metric-based alerts and custom check-based alerts.

### Key Features

- **Unified Alert Path**: All alerts flow through Prometheus → Alertmanager
- **Group-based Customization**: Different alert thresholds per group/namespace
- **Label-based Filtering**: Precise targeting of alerts to specific infrastructure
- **Dynamic Check System**: Script-based monitoring for custom requirements
- **Template-based Management**: Reusable alert and check templates
- **Policy Inheritance**: Smart configuration merging across group hierarchy

### Design Philosophy

AAMI maintains a **single, consistent alerting pipeline** rather than multiple independent notification systems. This approach provides:

- Centralized alert management
- Consistent routing and grouping policies
- Unified notification channels
- Easier troubleshooting and debugging
- Predictable alert behavior

---

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Accelerator Cluster                        │
│              (GPU Servers, Storage, Network)                     │
└────────────────┬──────────────────────┬─────────────────────────┘
                 │                      │
       ┌─────────▼─────────┐  ┌────────▼────────┐
       │  Node Exporter    │  │ Custom Checks   │
       │  DCGM Exporter    │  │ (dynamic-check) │
       │  Custom Exporters │  │                 │
       └─────────┬─────────┘  └────────┬────────┘
                 │                      │
                 └──────────┬───────────┘
                            │ Metrics
                 ┌──────────▼──────────┐
                 │    Prometheus       │
                 │  - Scrape metrics   │
                 │  - Evaluate rules   │
                 │  - Store TSDB       │
                 └──────────┬──────────┘
                            │ Firing Alerts
                 ┌──────────▼──────────┐
                 │   Alertmanager      │
                 │  - Route alerts     │
                 │  - Group/Inhibit    │
                 │  - Deduplicate      │
                 └──────────┬──────────┘
                            │ Notifications
         ┌──────────────────┼──────────────────┐
         │                  │                  │
    ┌────▼────┐      ┌─────▼──────┐     ┌────▼─────┐
    │  Email  │      │   Slack    │     │ Webhook  │
    └─────────┘      └────────────┘     └──────────┘
```

### Unified Alert Path

**Critical Design Decision**: All alerts, regardless of source, follow the same path:

```
Source → Metrics → Prometheus → Alert Rules → Alertmanager → Notification
```

This means:
- ❌ No direct email sending from check scripts
- ❌ No independent notification systems
- ✅ All alerts through Prometheus/Alertmanager
- ✅ Consistent routing and grouping
- ✅ Centralized configuration

---

## Components

### Prometheus

**Role**: Metrics collection, storage, and alert rule evaluation

**Responsibilities**:
- Scrape metrics from exporters every 15 seconds (configurable)
- Store time-series data in TSDB
- Evaluate alert rules every 15 seconds (configurable)
- Send firing alerts to Alertmanager
- Provide PromQL query interface

**Configuration**:
- `config/prometheus/prometheus.yml`: Main configuration
- `config/prometheus/rules/*.yml`: Alert rules
- Service Discovery via HTTP SD from Config Server

### Alertmanager

**Role**: Alert management and routing

**Responsibilities**:
- **Routing**: Direct alerts to appropriate receivers based on labels
- **Grouping**: Combine similar alerts to reduce notification volume
- **Inhibition**: Suppress lower-priority alerts when higher-priority ones fire
- **Deduplication**: Prevent duplicate notifications
- **Silencing**: Temporarily mute specific alerts

**Configuration**: `config/alertmanager/alertmanager.yml`

**Key Features**:
- Severity-based routing (critical, warning, info)
- Namespace-based routing (infrastructure, logical, environment)
- Time-based grouping (group_wait, group_interval, repeat_interval)

### Alert Rules

**Role**: Define conditions that trigger alerts

**Structure**:
```yaml
- alert: AlertName
  expr: PromQL expression
  for: duration
  labels:
    severity: critical
    group_id: grp-123
  annotations:
    summary: Alert summary
    description: Detailed description
```

**Storage**: `config/prometheus/rules/*.yml`

**Current State**:
- ✅ Static rule files (manually created)
- 📋 Dynamic generation (planned for Phase 3)

---

## Alert Types

### 1. Prometheus-based Alerts

**Definition**: Alerts triggered by standard Prometheus metrics from exporters

**Data Flow**:
```
Exporter → Prometheus → Alert Rules → Alertmanager
```

**Examples**:
- Node Exporter metrics: CPU, memory, disk, network
- DCGM Exporter metrics: GPU utilization, temperature, power
- Custom Exporter metrics: Application-specific

**Rule Example**:
```yaml
- alert: HighCPUUsage
  expr: |
    (100 - (avg by(instance) (
      rate(node_cpu_seconds_total{mode="idle"}[5m])
    ) * 100)) > 80
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High CPU usage on {{ $labels.instance }}"
```

### 2. Custom Check System

**Definition**: Script-based monitoring for infrastructure components not covered by standard exporters

**Data Flow**:
```
Config Server (CheckTemplate/Instance)
  ↓
Node queries effective checks
  ↓
dynamic-check.sh executes scripts
  ↓
JSON output
  ↓
Convert to Prometheus format
  ↓
Save to /var/lib/node_exporter/textfile/*.prom
  ↓
Node Exporter textfile collector
  ↓
Prometheus scrapes
  ↓
Alert Rules evaluate
  ↓
Alertmanager
```

**Use Cases**:
- Mount point availability
- Device connection status
- Network interface checks
- Custom application health checks
- Filesystem-specific monitoring

**Key Components**:
- **CheckTemplate**: Reusable script definition (services/config-server/internal/domain/check_template.go)
- **CheckInstance**: Group-specific application (services/config-server/internal/domain/check_instance.go)
- **Scope-based Management**: Global → Namespace → Group hierarchy

**Example**: Mount Point Check

```bash
# CheckTemplate script
#!/bin/bash
PATHS="$1"
for path in ${PATHS//,/ }; do
  if mountpoint -q "$path"; then
    echo '{"name":"mount_status","value":1,"labels":{"path":"'$path'"}}'
  else
    echo '{"name":"mount_status","value":0,"labels":{"path":"'$path'"}}'
  fi
done
```

Output to textfile:
```
mount_status{path="/data"} 1
mount_status{path="/mnt/models"} 0
```

Alert rule:
```yaml
- alert: MountPointUnavailable
  expr: mount_status == 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Mount point {{ $labels.path }} unavailable"
```

**Important**: Custom checks still go through Prometheus/Alertmanager, not direct notification.

---

## Data Flow

### Standard Metrics Path

```
┌─────────────────┐
│ Node Exporter   │  Port 9100, metrics endpoint
│ DCGM Exporter   │  Port 9400, metrics endpoint
│ Custom Exporter │  Port 9xxx, metrics endpoint
└────────┬────────┘
         │ HTTP GET /metrics (every 15s)
         ↓
┌────────────────────────────────┐
│ Prometheus                     │
│ - Scrape metrics               │
│ - Store in TSDB                │
│ - Evaluate rules (every 15s)   │
└────────┬───────────────────────┘
         │ Firing alerts
         ↓
┌────────────────────────────────┐
│ Alertmanager                   │
│ - Route by severity/namespace  │
│ - Group similar alerts         │
│ - Apply inhibition rules       │
└────────┬───────────────────────┘
         │ Send notifications
         ↓
┌────────────────────────────────┐
│ Notification Channels          │
│ - Email (SMTP)                 │
│ - Slack (Webhook)              │
│ - PagerDuty (Webhook)          │
└────────────────────────────────┘
```

### Custom Check Path

```
┌─────────────────────────────────┐
│ Config Server                   │
│ - CheckTemplate storage         │
│ - CheckInstance management      │
│ - Scope resolution              │
└────────┬────────────────────────┘
         │ GET /api/v1/checks/node?hostname=gpu-node-01
         ↓
┌─────────────────────────────────┐
│ Node: dynamic-check.sh          │
│ 1. Query effective checks       │
│ 2. Execute scripts              │
│ 3. Collect JSON output          │
└────────┬────────────────────────┘
         │ JSON metrics
         ↓
┌─────────────────────────────────┐
│ Convert to Prometheus format    │
│ mount_status{path="/data"} 1    │
└────────┬────────────────────────┘
         │ Write to file
         ↓
┌─────────────────────────────────┐
│ /var/lib/node_exporter/         │
│   textfile/*.prom               │
└────────┬────────────────────────┘
         │ Read by textfile collector
         ↓
┌─────────────────────────────────┐
│ Node Exporter                   │
│ - Expose as metrics             │
└────────┬────────────────────────┘
         │ Scrape
         ↓
┌─────────────────────────────────┐
│ Prometheus                      │
│ (Same path as standard metrics) │
└─────────────────────────────────┘
```

---

## Group-based Customization

### Problem Statement

**Question**: Prometheus alert rules are global. How can we support different alert thresholds per group?

**Example**:
- Production group: CPU alert at 80%
- Development group: CPU alert at 95%

### Solution: Label-based Filtering + Dynamic Rule Generation

#### Step 1: Add Group Labels in Service Discovery

**Code**: `services/config-server/internal/domain/service_discovery.go:38-54`

```go
// When registering targets, add group information as labels
labels["group"] = target.Groups[0].Name           // "gpu-cluster-a"
labels["group_id"] = target.Groups[0].ID          // "grp-123"
labels["namespace"] = target.Groups[0].Namespace.Name  // "production"
```

**Result**: All metrics from this target include group labels

```promql
node_cpu_seconds_total{
  instance="gpu-node-01",
  group="gpu-cluster-a",
  group_id="grp-123",
  namespace="production"
}
```

#### Step 2: Generate Group-specific Alert Rules

Each group gets its own alert rule with:
- Group-specific PromQL filter (`group_id="grp-123"`)
- Group-specific threshold (80% vs 95%)
- Group-specific duration (5m vs 10m)

**Production Group** (threshold: 80%):
```yaml
# /etc/prometheus/rules/generated/production-group-grp-123.yml
groups:
  - name: production_cpu_alerts
    rules:
      - alert: HighCPUUsage_Production
        expr: |
          (100 - (avg by(instance) (
            rate(node_cpu_seconds_total{
              mode="idle",
              group_id="grp-123"  # Filter to this group
            }[5m])
          ) * 100)) > 80  # Production threshold
        for: 5m
        labels:
          severity: warning
          group_id: grp-123
          namespace: production
```

**Development Group** (threshold: 95%):
```yaml
# /etc/prometheus/rules/generated/development-group-grp-456.yml
groups:
  - name: development_cpu_alerts
    rules:
      - alert: HighCPUUsage_Development
        expr: |
          (100 - (avg by(instance) (
            rate(node_cpu_seconds_total{
              mode="idle",
              group_id="grp-456"  # Filter to this group
            }[5m])
          ) * 100)) > 95  # Development threshold
        for: 10m
        labels:
          severity: info
          group_id: grp-456
          namespace: development
```

#### Step 3: AlertRule.RenderQuery() for Dynamic Generation

**Code**: `services/config-server/internal/domain/alert.go:102-125`

```go
// AlertTemplate with query template
QueryTemplate: `(100 - avg(rate(node_cpu_seconds_total{
  mode="idle",
  group_id="{{.group_id}}"
}[5m])) * 100) > {{.threshold}}`

// Production Group Config
Config: {
  "group_id": "grp-123",
  "threshold": 80,
  "for_duration": "5m"
}

// Rendered PromQL:
"(100 - avg(rate(node_cpu_seconds_total{
  mode=\"idle\",
  group_id=\"grp-123\"
}[5m])) * 100) > 80"
```

### Benefits

- ✅ Same metric, different thresholds per group
- ✅ Clean separation of rules by group
- ✅ Easy to debug (group_id in labels)
- ✅ Scalable (automatic generation)
- ✅ Flexible (template + config approach)

### Target-specific Customization

For even finer control, use target labels:

```yaml
- alert: HighCPUUsage_GPU_Servers
  expr: |
    (100 - avg by(instance) (
      rate(node_cpu_seconds_total{
        mode="idle",
        group_id="grp-123",
        target_label_type="gpu"  # Target-specific filter
      }[5m])
    ) * 100) > 70  # Different threshold for GPU servers
```

---

## Alert Rule Generation

### Current State

**Implemented**:
- ✅ AlertTemplate API (services/config-server/internal/service/alert.go)
- ✅ AlertRule API (group-specific configuration)
- ✅ AlertRule.RenderQuery() (template rendering)
- ✅ Group hierarchy and policy inheritance
- ✅ Database schema (alert_templates, alert_rules)

**Not Implemented**:
- ❌ Prometheus rule file generation
- ❌ Dynamic rule deployment to Prometheus
- ❌ Automatic Prometheus reload

### Planned Implementation (Phase 3)

**Location**: `services/config-server/internal/service/prometheus_rule_generator.go` (future)

**Workflow**:
```go
func GeneratePrometheusRules(ctx context.Context) error {
    // 1. Query all enabled alert rules from database
    rules := alertRuleRepo.ListEnabled(ctx)

    // 2. Group by group_id
    rulesByGroup := groupRules(rules)

    // 3. For each group, generate rule file
    for groupID, groupRules := range rulesByGroup {
        prometheusRules := []PrometheusRule{}

        for _, rule := range groupRules {
            // 4. Render PromQL query (already implemented!)
            query := rule.RenderQuery()

            // 5. Convert to Prometheus YAML format
            prometheusRules = append(prometheusRules, PrometheusRule{
                Alert: fmt.Sprintf("%s_Group_%s", rule.Name, groupID),
                Expr:  query,
                For:   rule.Config["for_duration"],
                Labels: map[string]string{
                    "group_id": groupID,
                    "severity": string(rule.Severity),
                },
            })
        }

        // 6. Write to file
        filename := fmt.Sprintf("/etc/prometheus/rules/generated/group-%s.yml", groupID)
        writeYAML(filename, prometheusRules)
    }

    // 7. Reload Prometheus
    reloadPrometheus()
}
```

**Trigger Events**:
- Alert rule created/updated/deleted
- Group configuration changed
- Manual refresh via API

**Expected Timeline**: Q2 2025 (Phase 3: Integration & Advanced Features)

---

## Integration Points

### 1. Service Discovery → Labels

**File**: `services/config-server/internal/domain/service_discovery.go`

When targets are registered, group information is added as labels:

```go
labels["group"] = target.Groups[0].Name
labels["group_id"] = target.Groups[0].ID
labels["namespace"] = target.Groups[0].Namespace.Name
```

These labels are used in:
- Alert rule filtering (`group_id="grp-123"`)
- Alertmanager routing (`namespace: production`)
- Grafana dashboard variables

### 2. Alert Rules → Alertmanager

**File**: `config/prometheus/prometheus.yml:8-12`

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093
```

Prometheus sends firing alerts to Alertmanager with all labels preserved.

### 3. Alertmanager → Notification Channels

**File**: `config/alertmanager/alertmanager.yml`

Routes alerts based on:
- **Severity**: critical, warning, info
- **Namespace**: infrastructure, logical, environment
- **Custom labels**: team, service, etc.

Example routing:
```yaml
routes:
  - match:
      severity: critical
    receiver: 'oncall-team'
    group_wait: 0s
    repeat_interval: 4h

  - match:
      namespace: infrastructure
    receiver: 'infrastructure-team'
    continue: true
```

### 4. CheckInstance → Node Execution

**API Endpoint**: `GET /api/v1/checks/node?hostname={hostname}`

Nodes query Config Server to get:
- Effective CheckInstances (after scope resolution)
- Script content and hash
- Merged configuration

Response:
```json
[
  {
    "check_type": "mount",
    "script_content": "#!/bin/bash\n...",
    "config": {
      "paths": "/data,/mnt/models"
    },
    "hash": "abc123..."
  }
]
```

---

## Examples

### Example 1: Standard Metric Alert (Node Down)

**Rule File**: `config/prometheus/rules/system-alerts.yml`

```yaml
- alert: NodeDown
  expr: up{job="node-exporter"} == 0
  for: 2m
  labels:
    severity: critical
    namespace: infrastructure
  annotations:
    summary: "Node {{ $labels.instance }} is down"
    description: |
      Node has not responded for more than 2 minutes.
      Instance: {{ $labels.instance }}
      Primary Group: {{ $labels.group }}
```

**Flow**:
1. Node Exporter stops responding
2. Prometheus marks `up{job="node-exporter"}` as 0
3. Alert rule condition met for 2 minutes
4. Prometheus sends alert to Alertmanager
5. Alertmanager routes to 'critical-alerts' receiver (email + PagerDuty)

### Example 2: Group-specific Disk Alert

**Scenario**: Different disk thresholds for different environments

**AlertTemplate**:
```json
{
  "name": "HighDiskUsage",
  "query_template": "((node_filesystem_avail_bytes{group_id=\"{{.group_id}}\"} / node_filesystem_size_bytes) * 100) < {{.threshold}}",
  "default_config": {
    "threshold": 20
  }
}
```

**AlertRule (Production)**:
```json
{
  "group_id": "production-grp-123",
  "template_id": "HighDiskUsage",
  "config": {
    "threshold": 20,
    "for_duration": "5m"
  }
}
```

**AlertRule (Development)**:
```json
{
  "group_id": "development-grp-456",
  "template_id": "HighDiskUsage",
  "config": {
    "threshold": 10,
    "for_duration": "10m"
  }
}
```

**Generated Prometheus Rules**:

Production:
```yaml
- alert: HighDiskUsage_Production
  expr: ((node_filesystem_avail_bytes{group_id="production-grp-123"} / node_filesystem_size_bytes) * 100) < 20
  for: 5m
```

Development:
```yaml
- alert: HighDiskUsage_Development
  expr: ((node_filesystem_avail_bytes{group_id="development-grp-456"} / node_filesystem_size_bytes) * 100) < 10
  for: 10m
```

### Example 3: Custom Check (Mount Point)

**CheckTemplate**:
```bash
POST /api/v1/check-templates
{
  "name": "mount-check",
  "check_type": "mount",
  "script_content": "#!/bin/bash\nPATHS=\"$1\"\nfor path in ${PATHS//,/ }; do\n  if mountpoint -q \"$path\"; then\n    echo '{\"name\":\"mount_status\",\"value\":1,\"labels\":{\"path\":\"'$path'\"}}'\n  else\n    echo '{\"name\":\"mount_status\",\"value\":0,\"labels\":{\"path\":\"'$path'\"}}'\n  fi\ndone",
  "language": "bash",
  "default_config": {
    "paths": "/data"
  }
}
```

**CheckInstance (ML Training Group)**:
```bash
POST /api/v1/check-instances
{
  "template_id": "mount-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "paths": "/data,/mnt/models,/mnt/datasets"
  }
}
```

**Node Execution**:
```bash
# dynamic-check.sh runs periodically
/opt/aami/scripts/dynamic-check.sh

# Outputs to textfile:
# /var/lib/node_exporter/textfile/mount-check.prom
mount_status{path="/data"} 1
mount_status{path="/mnt/models"} 0  # Failed!
mount_status{path="/mnt/datasets"} 1
```

**Alert Rule**:
```yaml
- alert: MountPointUnavailable
  expr: mount_status == 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Mount point {{ $labels.path }} unavailable on {{ $labels.instance }}"
```

**Result**: When `/mnt/models` fails to mount, alert fires after 2 minutes and sends notification.

---

## FAQ

### Q: Does the alert system depend on Alertmanager?

**A**: Partially.

- **Alert Evaluation**: No dependency. Prometheus evaluates alert rules independently and marks alerts as "firing" in its internal state.
- **Alert Notification**: Yes, requires Alertmanager. Without it, alerts are visible in Prometheus UI (`http://localhost:9090/alerts`) but no notifications are sent.

**With Alertmanager**:
```
Prometheus → Evaluates rules → Fires alerts → Alertmanager → Email/Slack
```

**Without Alertmanager**:
```
Prometheus → Evaluates rules → Fires alerts → [No notifications]
                                             └→ Visible in Prometheus UI only
```

### Q: Are Prometheus alert rules global?

**A**: Yes, but AAMI uses **label-based filtering** to achieve group-specific behavior.

- Prometheus rule files are global (loaded from `config/prometheus/rules/*.yml`)
- Each rule can filter metrics by labels (`group_id="grp-123"`)
- AAMI generates separate rules for each group with different thresholds
- Result: Appears group-specific, implemented as multiple global rules

### Q: Is custom infrastructure monitoring done via custom exporters?

**A**: No, AAMI uses the **Custom Check System**, not custom exporters.

**Custom Exporter** (traditional approach):
- Separate Go/Python process
- Exposes HTTP metrics endpoint
- Requires binary deployment
- Hard to customize per group

**AAMI Check System** (dynamic approach):
- Shell/Python scripts
- JSON output → Prometheus format conversion
- Node Exporter textfile collector
- Easy per-group customization via CheckInstance
- Dynamic deployment via Config Server API

Both paths eventually go through Prometheus → Alertmanager.

### Q: Can alerts bypass Prometheus/Alertmanager for faster notification?

**A**: No, and this is by design.

**Why unified path**:
- Consistent alert routing and grouping
- Single source of truth for alert state
- Easier troubleshooting (one place to check)
- Alertmanager features (inhibition, deduplication, silencing)
- Prevents alert storms from multiple sources

**Trade-off**:
- Slight delay (scrape_interval + evaluation_interval + Alertmanager processing)
- Typical latency: 30-60 seconds
- Acceptable for infrastructure monitoring
- For sub-second requirements, consider direct monitoring in application code

### Q: How do I test alert rules before deployment?

**A**: Use Prometheus UI and promtool:

```bash
# Validate syntax
promtool check rules config/prometheus/rules/system-alerts.yml

# Test query in Prometheus UI
http://localhost:9090/graph

# Enter PromQL expression
(100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)) > 80

# Manually trigger alert (set threshold very low)
# Watch Alerts page
http://localhost:9090/alerts
```

### Q: Can I have both global and group-specific alert rules?

**A**: Yes, this is a common pattern.

**Global Rule** (baseline for all groups):
```yaml
- alert: NodeDown
  expr: up{job="node-exporter"} == 0
  for: 5m  # More lenient
```

**Group-specific Rule** (stricter for production):
```yaml
- alert: NodeDown_Production
  expr: up{job="node-exporter",namespace="production"} == 0
  for: 1m  # Faster alert for production
```

Use inhibition rules to prevent duplicate alerts.

### Q: What happens when an alert rule file has syntax errors?

**A**: Prometheus will:
1. Log error on startup/reload
2. Skip the invalid rule file
3. Continue with valid rule files
4. Alert evaluation continues for valid rules

Always validate with `promtool check rules` before deployment.

### Q: How do I silence alerts during maintenance?

**A**: Use Alertmanager silences:

```bash
# Via UI
http://localhost:9093/#/silences

# Via API
curl -X POST http://localhost:9093/api/v2/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {"name": "instance", "value": "gpu-node-01", "isRegex": false}
    ],
    "startsAt": "2025-01-01T10:00:00Z",
    "endsAt": "2025-01-01T12:00:00Z",
    "comment": "Scheduled maintenance",
    "createdBy": "admin@example.com"
  }'
```

Silences are temporary and automatically expire.

---

## References

- [Check Management System](./CHECK-MANAGEMENT.md) - Custom check system details
- [Quick Start Guide](./QUICKSTART.md) - Getting started with AAMI
- [API Documentation](./API.md) - Alert and check API reference
- [Prometheus Documentation](https://prometheus.io/docs/alerting/latest/overview/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
