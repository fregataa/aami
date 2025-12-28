# Check Management System

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [CheckTemplate vs CheckInstance](#checktemplate-vs-checkinstance)
4. [Scope-based Management](#scope-based-management)
5. [Script Output Format](#script-output-format)
6. [Workflow](#workflow)
7. [API Reference](#api-reference)
8. [Examples](#examples)

## Overview

AAMI's dynamic check system enables centralized management and deployment of custom monitoring checks across your infrastructure.

### Key Concepts

- **CheckTemplate**: Reusable check definition (script code + default parameters)
- **CheckInstance**: Group-specific check application (Template reference + Override parameters)
- **Scope-based Management**: Organize checks in Global → Namespace → Group hierarchy
- **Local Caching**: Nodes cache scripts as local files
- **Auto-update**: Hash-based version detection for automatic script updates
- **JSON Output**: Scripts output JSON, system converts to Prometheus format

---

## Architecture

### Consistency with Alert System

```
Alert System:
├─ AlertTemplate (Reusable alert rule definition)
└─ AlertRule (Group-specific alert application, references Template)

Check System (Same Pattern):
├─ CheckTemplate (Reusable check definition)
└─ CheckInstance (Group-specific check application, references Template)
```

### Data Flow

```
┌──────────────────┐
│ CheckTemplate    │  Defined by admin
│ (Reusable)       │  - Script code
└────────┬─────────┘  - Default parameters
         │
         │ Reference
         ↓
┌──────────────────┐
│ CheckInstance    │  Applied per group
│ (Group-specific) │  - Template reference
└────────┬─────────┘  - Override parameters
         │
         │ Node queries
         ↓
┌──────────────────┐
│ Node             │  Script execution
│ (dynamic-check)  │  - Local cache
└──────────────────┘  - Periodic execution
```

---

## CheckTemplate vs CheckInstance

### CheckTemplate (Check Definition)

**Purpose**: Define reusable check scripts

**Structure**:
```go
type CheckTemplate struct {
    ID            string
    Name          string                  // "disk-usage-check"
    CheckType     string                  // "disk"
    ScriptContent string                  // Script code
    Language      string                  // "bash", "python"
    DefaultConfig map[string]interface{}  // Default parameters
    Description   string
    Version       string
    Hash          string                  // SHA256 hash
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Features**:
- Define once, reuse everywhere
- Version management and history
- Centralized script logic management

**Example**:
```json
{
  "name": "disk-usage-check",
  "check_type": "disk",
  "script_content": "#!/bin/bash\nTHRESHOLD=${1:-90}\ndf -h...",
  "language": "bash",
  "default_config": {
    "threshold": 90,
    "for": "5m"
  }
}
```

---

### CheckInstance (Check Application)

**Purpose**: Apply Template to specific Scope (Global/Namespace/Group)

**Structure**:
```go
type CheckInstance struct {
    ID          string
    TemplateID  string                  // CheckTemplate reference
    Scope       string                  // "global", "namespace", "group"
    NamespaceID *string                 // When namespace-level
    GroupID     *string                 // When group-level
    Config      map[string]interface{}  // Override parameters
    Priority    int                     // Priority (lower = higher priority)
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Features**:
- References Template
- Can override parameters
- Scope-based priority resolution

**Example**:
```json
// ML Training Group's Disk Check
{
  "template_id": "disk-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "threshold": 70,  // Override: Template's 90 → 70
    "for": "3m"       // Override: Template's 5m → 3m
  },
  "priority": 100
}

// API Server Group's Disk Check
{
  "template_id": "disk-check-template-id",
  "scope": "group",
  "group_id": "api-server-group",
  "config": {
    "threshold": 95   // Override: More relaxed
  },
  "priority": 100
}
```

---

## Scope-based Management

### Scope Priority

CheckInstance supports 3 scopes, with more specific scopes taking precedence:

```
Group (Most specific, highest priority)
  ↑
Namespace (Medium)
  ↑
Global (Most general, lowest priority)
```

### Scope Resolution Logic

When a node queries for check scripts:

1. **Check Target's PrimaryGroup**
2. **Query Group-level CheckInstance**
   - Use if exists
3. **Query Namespace-level CheckInstance** (if not in Group)
   - Use if exists
4. **Query Global-level CheckInstance** (if not in Namespace)
   - Use if exists
5. **Error if none found**

### Example: Disk Check Application

```
Global CheckInstance:
  template: disk-check-template
  config: { threshold: 90 }

Namespace "production" CheckInstance:
  template: disk-check-template
  config: { threshold: 80 }  # Stricter

Group "ml-training" CheckInstance:
  template: disk-check-template
  config: { threshold: 70 }  # Strictest

Result:
├─ ml-training group nodes → threshold: 70 (Group)
├─ Other production group nodes → threshold: 80 (Namespace)
└─ development group nodes → threshold: 90 (Global)
```

---

## Script Output Format

### JSON Output (Recommended)

Check scripts output simple JSON format:

```json
{
  "metrics": [
    {
      "name": "disk_usage_percent",
      "value": 85.5,
      "labels": {
        "path": "/data",
        "fstype": "ext4"
      }
    },
    {
      "name": "disk_available_bytes",
      "value": 50000000000,
      "labels": {
        "path": "/data"
      }
    }
  ]
}
```

### Using Helper Libraries

#### Bash Helper (`/opt/aami/lib/prom-helper.sh`)

```bash
#!/bin/bash
source /opt/aami/lib/prom-helper.sh

# JSON output
echo '{"metrics": [{"name": "my_metric", "value": 42}]}'
```

#### Python Helper (`/opt/aami/lib/prom_helper.py`)

```python
from prom_helper import output_metrics

output_metrics([
    {"name": "my_metric", "value": 42, "labels": {"type": "example"}}
])
```

### Conversion Process

```
Check Script → JSON → dynamic-check.sh → Prometheus Format → Node Exporter
```

---

## Workflow

### 1. Create Template (Admin)

```bash
POST /api/v1/check-templates
{
  "name": "mount-check",
  "check_type": "mount",
  "script_content": "#!/bin/bash\n...",
  "language": "bash",
  "default_config": {
    "paths": ["/data"]
  }
}
```

### 2. Create Instance (Admin)

```bash
# Apply to ML Training Group
POST /api/v1/check-instances
{
  "template_id": "mount-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "paths": ["/data", "/mnt/models"]  # Override
  }
}
```

### 3. Execute on Node

```bash
# dynamic-check.sh runs periodically
/opt/aami/scripts/dynamic-check.sh

# Internal operation:
# 1. Query Config Server for this node's CheckInstances
# 2. Execute Template script + Instance parameters
# 3. Convert JSON output to Prometheus format
# 4. Save to /var/lib/node_exporter/textfile/*.prom
```

### 4. Prometheus Collection

```
Node Exporter reads textfile/*.prom
  ↓
Prometheus scrapes
  ↓
Metrics stored and AlertRules evaluated
```

---

## API Reference

### CheckTemplate API

#### Create Template
```http
POST /api/v1/check-templates
Content-Type: application/json

{
  "name": "disk-usage-check",
  "check_type": "disk",
  "script_content": "#!/bin/bash\n...",
  "language": "bash",
  "default_config": {
    "threshold": 90
  },
  "description": "Disk usage monitoring"
}
```

#### Get Template
```http
GET /api/v1/check-templates/:id
```

#### List Templates
```http
GET /api/v1/check-templates?page=1&limit=20
```

#### Update Template
```http
PUT /api/v1/check-templates/:id
Content-Type: application/json

{
  "script_content": "#!/bin/bash\n# Updated...",
  "version": "2.0.0"
}
```

#### Delete Template
```http
POST /api/v1/check-templates/delete
Content-Type: application/json

{"id": "template-id"}
```

---

### CheckInstance API

#### Create Instance
```http
POST /api/v1/check-instances
Content-Type: application/json

{
  "template_id": "disk-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "threshold": 70
  },
  "priority": 100
}
```

#### Get Instance
```http
GET /api/v1/check-instances/:id
```

#### List Instances by Scope
```http
GET /api/v1/check-instances/global
GET /api/v1/check-instances/namespace/:namespace_id
GET /api/v1/check-instances/group/:group_id
```

#### Update Instance
```http
PUT /api/v1/check-instances/:id
Content-Type: application/json

{
  "config": {
    "threshold": 75
  }
}
```

---

### Node API (For Nodes Only)

#### Get Effective Checks
Query all checks that a node should execute:

```http
GET /api/v1/checks/node?hostname=ml-node-01

Response:
[
  {
    "check_type": "disk",
    "template": {
      "script_content": "#!/bin/bash\n...",
      "language": "bash"
    },
    "config": {
      "threshold": 70
    },
    "hash": "abc123...",
    "version": "1.0.0"
  },
  {
    "check_type": "mount",
    ...
  }
]
```

#### Check Script Version
Verify script updates (compare hash):

```http
GET /api/v1/checks/node/hash?hostname=ml-node-01&check_type=disk

Response:
{
  "check_type": "disk",
  "hash": "abc123...",
  "version": "1.0.0"
}
```

---

## Examples

### Example 1: Mount Point Check

#### 1. Create Template
```bash
curl -X POST http://config-server:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mount-check",
    "check_type": "mount",
    "script_content": "#!/bin/bash\nsource /opt/aami/lib/prom-helper.sh\nPATHS=\"$1\"\nmetrics=[]\nfor path in ${PATHS//,/ }; do\n  if mountpoint -q \"$path\"; then\n    metrics+=('{\"name\":\"mount_status\",\"value\":1,\"labels\":{\"path\":\"'$path'\"}}')\n  else\n    metrics+=('{\"name\":\"mount_status\",\"value\":0,\"labels\":{\"path\":\"'$path'\"}}')\n  fi\ndone\necho \"{\\\"metrics\\\":[$metrics]}\"",
    "language": "bash",
    "default_config": {
      "paths": "/data"
    }
  }'
```

#### 2. Create Instance (ML Training Group)
```bash
curl -X POST http://config-server:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "mount-check-template-id",
    "scope": "group",
    "group_id": "ml-training-group",
    "config": {
      "paths": "/data,/mnt/models,/mnt/datasets"
    }
  }'
```

---

### Example 2: Disk Usage Check (Different Thresholds per Group)

#### 1. Create Template (Once)
```bash
curl -X POST http://config-server:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "disk-usage-check",
    "check_type": "disk",
    "script_content": "#!/bin/bash\nTHRESHOLD=${1:-90}\ndf -BG / | tail -1 | awk -v threshold=$THRESHOLD '\''{usage=int($5); echo \"{\\\"metrics\\\":[{\\\"name\\\":\\\"disk_usage_percent\\\",\\\"value\\\":\"usage\"}]}\"}'\'",
    "language": "bash",
    "default_config": {
      "threshold": 90
    }
  }'
```

#### 2. Create Instances (Multiple Groups)
```bash
# Critical Services: 70%
curl -X POST http://config-server:8080/api/v1/check-instances \
  -d '{
    "template_id": "disk-check-template-id",
    "scope": "group",
    "group_id": "critical-services",
    "config": {"threshold": 70}
  }'

# Standard Services: 85%
curl -X POST http://config-server:8080/api/v1/check-instances \
  -d '{
    "template_id": "disk-check-template-id",
    "scope": "group",
    "group_id": "standard-services",
    "config": {"threshold": 85}
  }'

# Development: 95%
curl -X POST http://config-server:8080/api/v1/check-instances \
  -d '{
    "template_id": "disk-check-template-id",
    "scope": "group",
    "group_id": "development",
    "config": {"threshold": 95}
  }'
```

Result: Same script, different thresholds per group!

---

## Best Practices

### Template Design
1. **Reusability**: Write parameterizable, generic scripts
2. **Version Management**: Update version on changes
3. **Documentation**: Use Description field
4. **Testing**: Test in dev environment before deployment

### Instance Management
1. **Minimize Scope**: Use Global/Namespace when possible, Group only for exceptions
2. **Minimize Overrides**: Override only necessary parameters
3. **Priority Management**: Clear priorities for conflicts
4. **Deactivation**: Use IsActive=false instead of deletion

### Node Configuration
1. **Caching**: Maintain local cache for network failures
2. **Auto-update**: Automatic updates via hash comparison
3. **Error Handling**: Keep previous results on script failure

---

## Troubleshooting

### Template Updates Not Reflected on Node
```bash
# Check hash on node
curl "http://config-server:8080/api/v1/checks/node/hash?hostname=$(hostname)&check_type=disk"

# Remove local cache and re-execute
rm -f /opt/aami/cache/check-*.sh
/opt/aami/scripts/dynamic-check.sh
```

### Instance Priority Conflict
```bash
# Query instances
curl "http://config-server:8080/api/v1/check-instances/group/my-group"

# Update priority
curl -X PUT "http://config-server:8080/api/v1/check-instances/:id" \
  -d '{"priority": 50}'
```

---

## References

- [Quick Start Guide](./QUICKSTART.md)
- [Node Registration](./NODE-REGISTRATION.md)
- [Alert Rules Guide](./ALERT-RULES.md)
- [API Documentation](./API.md)
