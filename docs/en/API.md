# API Reference

Complete REST API documentation for AAMI Config Server.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication in development mode. For production deployments, implement API key or OAuth authentication.

**Future Authentication Header:**
```http
Authorization: Bearer YOUR_API_KEY
```

## Table of Contents

1. [Health Check](#health-check)
2. [Groups API](#groups-api)
3. [Targets API](#targets-api)
4. [Alert Rules API](#alert-rules-api)
5. [Check Management API](#check-management-api)
6. [Service Discovery API](#service-discovery-api)
7. [Bootstrap API](#bootstrap-api)
8. [Error Responses](#error-responses)

---

## Health Check

### Check API Health

Get the health status of the Config Server.

**Endpoint:** `GET /health`

**Example:**
```bash
curl http://localhost:8080/api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "database": "connected",
  "redis": "connected"
}
```

---

## Groups API

Manage monitoring groups and hierarchies.

### List All Groups

**Endpoint:** `GET /groups`

**Query Parameters:**
- `namespace` (optional): Filter by namespace (infrastructure, logical, environment)
- `parent_id` (optional): Filter by parent group
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)

**Example:**
```bash
# List all groups
curl http://localhost:8080/api/v1/groups

# Filter by namespace
curl http://localhost:8080/api/v1/groups?namespace=environment

# Filter by parent
curl http://localhost:8080/api/v1/groups?parent_id=GROUP_ID
```

**Response:**
```json
{
  "groups": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "production",
      "namespace": "environment",
      "parent_id": null,
      "description": "Production environment",
      "priority": 10,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "target_count": 15
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 50
}
```

### Get Group by ID

**Endpoint:** `GET /groups/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "Production environment",
  "priority": 10,
  "metadata": {
    "owner": "devops-team",
    "contact": "devops@example.com"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "children": [],
  "targets": []
}
```

### Create Group

**Endpoint:** `POST /groups`

**Request Body:**
```json
{
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "Production environment",
  "metadata": {
    "owner": "devops-team",
    "contact": "devops@example.com"
  }
}
```

**Example:**
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

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "Production environment",
  "priority": 10,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Update Group

**Endpoint:** `PUT /groups/:id`

**Request Body:**
```json
{
  "name": "production-updated",
  "description": "Updated production environment"
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated production environment"
  }'
```

### Delete Group

**Endpoint:** `DELETE /groups/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "message": "Group deleted successfully"
}
```

---

## Targets API

Manage monitoring targets (servers, nodes).

### List All Targets

**Endpoint:** `GET /targets`

**Query Parameters:**
- `group_id` (optional): Filter by group
- `status` (optional): Filter by status (active, inactive, down)
- `page` (optional): Page number
- `limit` (optional): Items per page

**Example:**
```bash
# List all targets
curl http://localhost:8080/api/v1/targets

# Filter by group
curl http://localhost:8080/api/v1/targets?group_id=GROUP_ID

# Filter by status
curl http://localhost:8080/api/v1/targets?status=active
```

**Response:**
```json
{
  "targets": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "hostname": "gpu-node-01.example.com",
      "ip_address": "10.0.1.10",
      "primary_group_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "active",
      "last_seen": "2024-01-01T12:00:00Z",
      "labels": {
        "gpu_model": "A100",
        "gpu_count": "8"
      },
      "created_at": "2024-01-01T11:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 50
}
```

### Get Target by ID

**Endpoint:** `GET /targets/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001
```

### Create Target

**Endpoint:** `POST /targets`

**Request Body:**
```json
{
  "hostname": "gpu-node-01.example.com",
  "ip_address": "10.0.1.10",
  "primary_group_id": "550e8400-e29b-41d4-a716-446655440000",
  "secondary_group_ids": [],
  "exporters": [
    {
      "type": "node_exporter",
      "port": 9100,
      "enabled": true,
      "scrape_interval": "15s",
      "scrape_timeout": "10s",
      "metrics_path": "/metrics"
    },
    {
      "type": "dcgm_exporter",
      "port": 9400,
      "enabled": true,
      "scrape_interval": "30s",
      "scrape_timeout": "10s"
    }
  ],
  "labels": {
    "datacenter": "dc1",
    "rack": "r1",
    "gpu_model": "A100",
    "gpu_count": "8",
    "instance_type": "p4d.24xlarge"
  },
  "metadata": {
    "provisioned_by": "terraform",
    "owner": "ml-team"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "gpu-node-01.example.com",
    "ip_address": "10.0.1.10",
    "primary_group_id": "550e8400-e29b-41d4-a716-446655440000",
    "exporters": [
      {
        "type": "node_exporter",
        "port": 9100,
        "enabled": true
      },
      {
        "type": "dcgm_exporter",
        "port": 9400,
        "enabled": true
      }
    ],
    "labels": {
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

### Update Target

**Endpoint:** `PUT /targets/:id`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -d '{
    "labels": {
      "gpu_model": "A100",
      "gpu_count": "8",
      "maintenance": "false"
    }
  }'
```

### Delete Target

**Endpoint:** `DELETE /targets/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001
```

---

## Alert Rules API

Manage alert rules and thresholds.

### List Alert Rule Templates

**Endpoint:** `GET /alert-templates`

**Example:**
```bash
curl http://localhost:8080/api/v1/alert-templates
```

**Response:**
```json
{
  "templates": [
    {
      "id": "HighCPUUsage",
      "name": "High CPU Usage",
      "description": "Alert when CPU usage exceeds threshold",
      "severity": "warning",
      "default_config": {
        "threshold": 80,
        "duration": "5m"
      },
      "query_template": "100 - (avg by(instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > {{.threshold}}"
    }
  ]
}
```

### Apply Alert Rule to Group

**Endpoint:** `POST /groups/:id/alert-rules`

**Request Body:**
```json
{
  "rule_template_id": "HighCPUUsage",
  "enabled": true,
  "config": {
    "threshold": 70,
    "duration": "5m"
  },
  "merge_strategy": "override",
  "annotations": {
    "summary": "High CPU usage detected",
    "description": "CPU usage is above 70% for more than 5 minutes"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "rule_template_id": "HighCPUUsage",
    "enabled": true,
    "config": {
      "threshold": 70,
      "duration": "5m"
    },
    "merge_strategy": "override"
  }'
```

### Get Effective Alert Rules for Target

**Endpoint:** `GET /targets/:id/alert-rules/effective`

**Example:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001/alert-rules/effective
```

**Response:**
```json
{
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "rules": [
    {
      "rule_id": "HighCPUUsage",
      "enabled": true,
      "config": {
        "threshold": 70,
        "duration": "5m"
      },
      "source_group": "production",
      "priority": 10
    }
  ]
}
```

### Trace Alert Rule Policy

**Endpoint:** `GET /targets/:id/alert-rules/trace`

Shows which groups contributed to the final alert configuration.

**Example:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001/alert-rules/trace
```

**Response:**
```json
{
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "trace": [
    {
      "rule_id": "HighCPUUsage",
      "inheritance_chain": [
        {
          "group_name": "infrastructure",
          "group_id": "...",
          "config": {"threshold": 80},
          "priority": 100
        },
        {
          "group_name": "production",
          "group_id": "...",
          "config": {"threshold": 70},
          "priority": 10,
          "override": true
        }
      ],
      "final_config": {"threshold": 70, "duration": "5m"}
    }
  ]
}
```

---

## Check Management API

Manages the dynamic check system. Use CheckTemplate (reusable check definitions) and CheckInstance (scope-specific template applications) to manage node-level checks.

### Create Check Template

**Endpoint:** `POST /check-templates`

Creates a reusable check script template.

**Request Body:**
```json
{
  "name": "disk-usage-check",
  "check_type": "disk",
  "script_content": "#!/bin/bash\ndf -h / | tail -1 | awk '{print \"disk_usage_percent{mount=\\\"/\\\"} \"$5}' | sed 's/%//'",
  "language": "bash",
  "default_config": {
    "threshold": 80,
    "mount": "/"
  },
  "description": "Disk usage check",
  "version": "1.0.0"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "disk-usage-check",
    "check_type": "disk",
    "script_content": "#!/bin/bash\ndf -h / | tail -1",
    "language": "bash",
    "default_config": {"threshold": 80},
    "version": "1.0.0"
  }'
```

**Response:**
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "name": "disk-usage-check",
  "check_type": "disk",
  "script_content": "...",
  "language": "bash",
  "default_config": {"threshold": 80},
  "hash": "a1b2c3d4...",
  "version": "1.0.0",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### List Check Templates

**Endpoint:** `GET /check-templates`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Example:**
```bash
curl http://localhost:8080/api/v1/check-templates
```

### Get Check Template

**Endpoint:** `GET /check-templates/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/check-templates/880e8400-e29b-41d4-a716-446655440003
```

### Create Check Instance

**Endpoint:** `POST /check-instances`

Applies a check template to a specific scope (Global/Namespace/Group).

**Request Body:**
```json
{
  "template_id": "880e8400-e29b-41d4-a716-446655440003",
  "scope": "global",
  "config": {
    "threshold": 85
  },
  "priority": 100,
  "is_active": true
}
```

**Scope-Specific Parameters:**
- **Global**: `scope: "global"` (no namespace_id or group_id required)
- **Namespace**: `scope: "namespace"`, requires `namespace_id`
- **Group**: `scope: "group"`, requires both `namespace_id` and `group_id`

**Example (Global):**
```bash
curl -X POST http://localhost:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "880e8400-e29b-41d4-a716-446655440003",
    "scope": "global",
    "config": {"threshold": 85},
    "priority": 100,
    "is_active": true
  }'
```

**Example (Group - Override):**
```bash
curl -X POST http://localhost:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "880e8400-e29b-41d4-a716-446655440003",
    "scope": "group",
    "namespace_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_id": "660e8400-e29b-41d4-a716-446655440001",
    "config": {"threshold": 70},
    "priority": 50,
    "is_active": true
  }'
```

### Get Node Effective Checks

**Endpoint:** `GET /checks/node/:hostname`

Retrieves all checks applied to a node based on priority (Group > Namespace > Global).

**Example:**
```bash
curl http://localhost:8080/api/v1/checks/node/gpu-node-01.example.com
```

**Response:**
```json
[
  {
    "check_type": "disk",
    "script_content": "#!/bin/bash\n...",
    "language": "bash",
    "config": {
      "threshold": 70,
      "mount": "/"
    },
    "version": "1.0.0",
    "hash": "a1b2c3d4...",
    "template_id": "880e8400-e29b-41d4-a716-446655440003",
    "instance_id": "990e8400-e29b-41d4-a716-446655440004"
  }
]
```

**Note:** The config field is the merged result of the template's default_config and the instance's config.

### List Check Instances

**Endpoint:** `GET /check-instances`

**Query Parameters:**
- `scope` (optional): Filter by scope (global, namespace, group)
- `namespace_id` (optional): Filter by namespace
- `group_id` (optional): Filter by group

**Example:**
```bash
# List all
curl http://localhost:8080/api/v1/check-instances

# Global instances only
curl http://localhost:8080/api/v1/check-instances?scope=global

# Specific group instances
curl http://localhost:8080/api/v1/check-instances?group_id=660e8400-e29b-41d4-a716-446655440001
```

### Update Check Template

**Endpoint:** `PUT /check-templates/:id`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/check-templates/880e8400-e29b-41d4-a716-446655440003 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated disk check",
    "version": "1.1.0"
  }'
```

### Update Check Instance

**Endpoint:** `PUT /check-instances/:id`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/check-instances/990e8400-e29b-41d4-a716-446655440004 \
  -H "Content-Type: application/json" \
  -d '{
    "config": {"threshold": 75},
    "is_active": true
  }'
```

### Delete Check Template (Soft Delete)

**Endpoint:** `DELETE /check-templates/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/check-templates/880e8400-e29b-41d4-a716-446655440003
```

### Delete Check Instance (Soft Delete)

**Endpoint:** `DELETE /check-instances/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/check-instances/990e8400-e29b-41d4-a716-446655440004
```

**Note:** For more detailed check management guidance, refer to the [Check Management Documentation](CHECK-MANAGEMENT.md).

---

## Service Discovery API

Endpoints for Prometheus service discovery integration.

### Get Prometheus SD Targets

**Endpoint:** `GET /sd/prometheus`

Returns targets in Prometheus file-based service discovery format.

**Example:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus
```

**Response:**
```json
[
  {
    "targets": ["10.0.1.10:9100"],
    "labels": {
      "__meta_aami_target_id": "660e8400-e29b-41d4-a716-446655440001",
      "__meta_aami_group": "production",
      "hostname": "gpu-node-01.example.com",
      "gpu_model": "A100",
      "gpu_count": "8",
      "job": "node-exporter"
    }
  },
  {
    "targets": ["10.0.1.10:9400"],
    "labels": {
      "__meta_aami_target_id": "660e8400-e29b-41d4-a716-446655440001",
      "__meta_aami_group": "production",
      "hostname": "gpu-node-01.example.com",
      "gpu_model": "A100",
      "job": "dcgm-exporter"
    }
  }
]
```

### Get Alert Rules for Prometheus

**Endpoint:** `GET /sd/alert-rules`

Returns alert rules in Prometheus rule format.

**Example:**
```bash
curl http://localhost:8080/api/v1/sd/alert-rules
```

---

## Bootstrap API

Endpoints for automated node registration.

### Create Bootstrap Token

**Endpoint:** `POST /bootstrap/tokens`

**Request Body:**
```json
{
  "name": "datacenter-1-token",
  "expires_at": "2024-12-31T23:59:59Z",
  "max_uses": 100,
  "default_group_id": "550e8400-e29b-41d4-a716-446655440000",
  "labels": {
    "datacenter": "dc1"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/bootstrap/tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "datacenter-1-token",
    "expires_at": "2024-12-31T23:59:59Z",
    "max_uses": 100,
    "default_group_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response:**
```json
{
  "token": "aami_1234567890abcdef",
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "datacenter-1-token",
  "expires_at": "2024-12-31T23:59:59Z",
  "max_uses": 100,
  "uses": 0,
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Bootstrap Registration

**Endpoint:** `POST /bootstrap/register`

Used by bootstrap script to auto-register nodes.

**Request Body:**
```json
{
  "token": "aami_1234567890abcdef",
  "hostname": "auto-gpu-node-01",
  "ip_address": "10.0.1.15",
  "hardware_info": {
    "cpu_cores": 96,
    "memory_gb": 768,
    "gpu_count": 8,
    "gpu_model": "NVIDIA A100",
    "disk_size_gb": 2048
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Invalid request body or parameters |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists |
| `VALIDATION_ERROR` | 422 | Validation failed |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `DATABASE_ERROR` | 500 | Database operation failed |

**Example Error Response:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Group not found",
    "details": {
      "group_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

---

## Rate Limiting

Currently, no rate limiting is enforced in development mode. For production deployments, implement rate limiting at the API gateway or load balancer level.

**Recommended Headers:**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1609459200
```

---

## API Versioning

The API uses URL-based versioning (`/api/v1`). Breaking changes will result in a new version (`/api/v2`).

---

For more examples, see [API Usage Examples](../../examples/api-usage/).
