# API Reference

Complete REST API documentation for AAMI Config Server.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. For production deployments, implement API key or OAuth authentication.

## Table of Contents

1. [Health Check](#health-check)
2. [Groups API](#groups-api)
3. [Targets API](#targets-api)
4. [Exporters API](#exporters-api)
5. [Alert Templates API](#alert-templates-api)
6. [Alert Rules API](#alert-rules-api)
7. [Active Alerts API](#active-alerts-api)
8. [Script Templates API](#script-templates-api)
9. [Script Policies API](#script-policies-api)
10. [Bootstrap Tokens API](#bootstrap-tokens-api)
11. [Service Discovery API](#service-discovery-api)
12. [Prometheus Management API](#prometheus-management-api)
13. [Error Responses](#error-responses)

---

## Health Check

### Check API Health

**Endpoint:** `GET /health`

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "v1.0.0",
  "database": "connected"
}
```

### Readiness Check

**Endpoint:** `GET /health/ready`

### Liveness Check

**Endpoint:** `GET /health/live`

---

## Groups API

Manage monitoring groups. Groups are flat (no hierarchy) and targets can belong to multiple groups.

### List All Groups

**Endpoint:** `GET /api/v1/groups`

```bash
curl http://localhost:8080/api/v1/groups
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "gpu-servers",
    "description": "GPU compute servers",
    "priority": 10,
    "is_default_own": false,
    "metadata": {},
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

### Get Group by ID

**Endpoint:** `GET /api/v1/groups/:id`

```bash
curl http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

### Create Group

**Endpoint:** `POST /api/v1/groups`

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-servers",
    "description": "Web application servers",
    "priority": 20,
    "metadata": {
      "environment": "production"
    }
  }'
```

### Update Group

**Endpoint:** `PUT /api/v1/groups/:id`

```bash
curl -X PUT http://localhost:8080/api/v1/groups/GROUP_ID \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description",
    "priority": 15
  }'
```

### Delete Group (Soft Delete)

**Endpoint:** `POST /api/v1/groups/delete`

```bash
curl -X POST http://localhost:8080/api/v1/groups/delete \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

### Restore Group

**Endpoint:** `POST /api/v1/groups/restore`

```bash
curl -X POST http://localhost:8080/api/v1/groups/restore \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

### Purge Group (Hard Delete)

**Endpoint:** `POST /api/v1/groups/purge`

```bash
curl -X POST http://localhost:8080/api/v1/groups/purge \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

---

## Targets API

Manage monitoring targets (nodes/servers).

### List All Targets

**Endpoint:** `GET /api/v1/targets`

```bash
curl http://localhost:8080/api/v1/targets
```

**Response:**
```json
[
  {
    "id": "target-uuid",
    "hostname": "gpu-node-01",
    "ip_address": "192.168.1.100",
    "port": 9100,
    "status": "active",
    "labels": {
      "rack": "A1",
      "gpu": "nvidia"
    },
    "groups": [
      {
        "id": "group-uuid",
        "name": "gpu-servers"
      }
    ],
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

### Get Target by ID

**Endpoint:** `GET /api/v1/targets/:id`

### Get Target by Hostname

**Endpoint:** `GET /api/v1/targets/hostname/:hostname`

```bash
curl http://localhost:8080/api/v1/targets/hostname/gpu-node-01
```

### Get Targets by Group

**Endpoint:** `GET /api/v1/targets/group/:group_id`

```bash
curl http://localhost:8080/api/v1/targets/group/GROUP_ID
```

### Create Target

**Endpoint:** `POST /api/v1/targets`

```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "gpu-node-02",
    "ip_address": "192.168.1.101",
    "port": 9100,
    "labels": {
      "rack": "A2"
    },
    "group_ids": ["group-uuid-1", "group-uuid-2"]
  }'
```

### Update Target

**Endpoint:** `PUT /api/v1/targets/:id`

### Update Target Status

**Endpoint:** `POST /api/v1/targets/:id/status`

```bash
curl -X POST http://localhost:8080/api/v1/targets/TARGET_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "inactive"}'
```

### Target Heartbeat

**Endpoint:** `POST /api/v1/targets/:id/heartbeat`

### Delete/Restore/Purge Target

- `POST /api/v1/targets/delete`
- `POST /api/v1/targets/restore`
- `POST /api/v1/targets/purge`

---

## Exporters API

Manage Prometheus exporters associated with targets.

### List All Exporters

**Endpoint:** `GET /api/v1/exporters`

### Get Exporter by ID

**Endpoint:** `GET /api/v1/exporters/:id`

### Get Exporters by Target

**Endpoint:** `GET /api/v1/exporters/target/:target_id`

### Get Exporters by Type

**Endpoint:** `GET /api/v1/exporters/type/:type`

```bash
curl http://localhost:8080/api/v1/exporters/type/node_exporter
```

### Create Exporter

**Endpoint:** `POST /api/v1/exporters`

```bash
curl -X POST http://localhost:8080/api/v1/exporters \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "target-uuid",
    "type": "node_exporter",
    "port": 9100,
    "path": "/metrics",
    "enabled": true
  }'
```

### Update/Delete/Restore/Purge Exporter

- `PUT /api/v1/exporters/:id`
- `POST /api/v1/exporters/delete`
- `POST /api/v1/exporters/restore`
- `POST /api/v1/exporters/purge`

---

## Alert Templates API

Manage reusable alert rule templates.

### List All Alert Templates

**Endpoint:** `GET /api/v1/alert-templates`

### Get Alert Template by ID

**Endpoint:** `GET /api/v1/alert-templates/:id`

### Get Alert Templates by Severity

**Endpoint:** `GET /api/v1/alert-templates/severity/:severity`

```bash
curl http://localhost:8080/api/v1/alert-templates/severity/critical
```

### Create Alert Template

**Endpoint:** `POST /api/v1/alert-templates`

```bash
curl -X POST http://localhost:8080/api/v1/alert-templates \
  -H "Content-Type: application/json" \
  -d '{
    "id": "high-cpu",
    "name": "High CPU Usage",
    "description": "Alert when CPU usage exceeds threshold",
    "severity": "warning",
    "query_template": "100 - (avg by(instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > {{.threshold}}",
    "default_config": {
      "threshold": 80
    }
  }'
```

### Update/Delete/Restore/Purge Alert Template

- `PUT /api/v1/alert-templates/:id`
- `POST /api/v1/alert-templates/delete`
- `POST /api/v1/alert-templates/restore`
- `POST /api/v1/alert-templates/purge`

---

## Alert Rules API

Manage alert rules assigned to groups.

### List All Alert Rules

**Endpoint:** `GET /api/v1/alert-rules`

### Get Alert Rule by ID

**Endpoint:** `GET /api/v1/alert-rules/:id`

### Get Alert Rules by Group

**Endpoint:** `GET /api/v1/alert-rules/group/:group_id`

### Get Alert Rules by Template

**Endpoint:** `GET /api/v1/alert-rules/template/:template_id`

### Create Alert Rule (from Template)

**Endpoint:** `POST /api/v1/alert-rules`

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "group-uuid",
    "template_id": "high-cpu",
    "enabled": true,
    "config": {
      "threshold": 90
    },
    "priority": 100
  }'
```

### Create Alert Rule (Direct)

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "group-uuid",
    "name": "Custom Alert",
    "description": "Custom alert rule",
    "severity": "critical",
    "query_template": "up == 0",
    "enabled": true,
    "priority": 100
  }'
```

### Update/Delete/Restore/Purge Alert Rule

- `PUT /api/v1/alert-rules/:id`
- `POST /api/v1/alert-rules/delete`
- `POST /api/v1/alert-rules/restore`
- `POST /api/v1/alert-rules/purge`

---

## Active Alerts API

Retrieve currently firing alerts from Alertmanager.

### Get Active Alerts

**Endpoint:** `GET /api/v1/alerts/active`

```bash
curl http://localhost:8080/api/v1/alerts/active
```

**Response:**
```json
{
  "alerts": [
    {
      "fingerprint": "abc123",
      "status": "firing",
      "labels": {
        "alertname": "HighCPU",
        "instance": "gpu-node-01:9100",
        "severity": "warning"
      },
      "annotations": {
        "summary": "High CPU usage detected",
        "description": "CPU usage is above 90%"
      },
      "starts_at": "2024-01-15T10:30:00Z",
      "generator_url": "http://prometheus:9090/graph?..."
    }
  ],
  "total": 1
}
```

---

## Script Templates API

Manage check script templates.

### List All Script Templates

**Endpoint:** `GET /api/v1/script-templates`

### List Active Script Templates

**Endpoint:** `GET /api/v1/script-templates/active`

### Get Script Template by ID

**Endpoint:** `GET /api/v1/script-templates/:id`

### Get Script Template by Name

**Endpoint:** `GET /api/v1/script-templates/name/:name`

### Get Script Templates by Type

**Endpoint:** `GET /api/v1/script-templates/type/:scriptType`

```bash
curl http://localhost:8080/api/v1/script-templates/type/check
```

### Verify Script Hash

**Endpoint:** `GET /api/v1/script-templates/:id/verify-hash`

### Create Script Template

**Endpoint:** `POST /api/v1/script-templates`

```bash
curl -X POST http://localhost:8080/api/v1/script-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "disk-check",
    "description": "Check disk usage",
    "script_type": "check",
    "script_content": "#!/bin/bash\ndf -h / | awk '\''NR==2 {print $5}'\''",
    "config_schema": {
      "threshold": {"type": "number", "default": 80}
    },
    "enabled": true
  }'
```

### Update/Delete/Restore/Purge Script Template

- `PUT /api/v1/script-templates/:id`
- `POST /api/v1/script-templates/delete`
- `POST /api/v1/script-templates/restore`
- `POST /api/v1/script-templates/purge`

---

## Script Policies API

Manage script policy assignments to groups.

### List All Script Policies

**Endpoint:** `GET /api/v1/script-policies`

### List Active Script Policies

**Endpoint:** `GET /api/v1/script-policies/active`

### Get Script Policy by ID

**Endpoint:** `GET /api/v1/script-policies/:id`

### Get Script Policies by Template

**Endpoint:** `GET /api/v1/script-policies/template/:templateId`

### Get Global Script Policies

**Endpoint:** `GET /api/v1/script-policies/global`

### Get Script Policies by Group

**Endpoint:** `GET /api/v1/script-policies/group/:groupId`

### Get Effective Checks by Group

**Endpoint:** `GET /api/v1/script-policies/effective/group/:groupId`

### Get Effective Checks by Target

**Endpoint:** `GET /api/v1/checks/target/:targetId`

This endpoint is used by nodes to fetch their assigned checks.

```bash
curl http://localhost:8080/api/v1/checks/target/TARGET_ID
```

### Create Script Policy

**Endpoint:** `POST /api/v1/script-policies`

```bash
curl -X POST http://localhost:8080/api/v1/script-policies \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "disk-check-template-id",
    "group_id": "group-uuid",
    "config": {
      "threshold": 85
    },
    "priority": 100,
    "enabled": true
  }'
```

### Update/Delete/Restore/Purge Script Policy

- `PUT /api/v1/script-policies/:id`
- `POST /api/v1/script-policies/delete`
- `POST /api/v1/script-policies/restore`
- `POST /api/v1/script-policies/purge`

---

## Bootstrap Tokens API

Manage bootstrap tokens for node auto-registration.

### List All Bootstrap Tokens

**Endpoint:** `GET /api/v1/bootstrap-tokens`

### Get Bootstrap Token by ID

**Endpoint:** `GET /api/v1/bootstrap-tokens/:id`

### Get Bootstrap Token by Token String

**Endpoint:** `GET /api/v1/bootstrap-tokens/token/:token`

### Create Bootstrap Token

**Endpoint:** `POST /api/v1/bootstrap-tokens`

```bash
curl -X POST http://localhost:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpu-cluster-token",
    "description": "Token for GPU cluster nodes",
    "group_id": "gpu-servers-group-id",
    "expires_at": "2024-12-31T23:59:59Z",
    "max_uses": 100
  }'
```

**Response:**
```json
{
  "id": "token-uuid",
  "name": "gpu-cluster-token",
  "token": "aami_bootstrap_abc123xyz...",
  "group_id": "gpu-servers-group-id",
  "expires_at": "2024-12-31T23:59:59Z",
  "max_uses": 100,
  "use_count": 0,
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Validate and Use Token

**Endpoint:** `POST /api/v1/bootstrap-tokens/validate`

```bash
curl -X POST http://localhost:8080/api/v1/bootstrap-tokens/validate \
  -H "Content-Type: application/json" \
  -d '{"token": "aami_bootstrap_abc123xyz..."}'
```

### Register Node with Token

**Endpoint:** `POST /api/v1/bootstrap-tokens/register`

```bash
curl -X POST http://localhost:8080/api/v1/bootstrap-tokens/register \
  -H "Content-Type: application/json" \
  -d '{
    "token": "aami_bootstrap_abc123xyz...",
    "hostname": "gpu-node-03",
    "ip_address": "192.168.1.103",
    "port": 9100,
    "labels": {
      "rack": "B1",
      "gpu_count": "8"
    }
  }'
```

### Update/Delete/Restore/Purge Bootstrap Token

- `PUT /api/v1/bootstrap-tokens/:id`
- `POST /api/v1/bootstrap-tokens/delete`
- `POST /api/v1/bootstrap-tokens/restore`
- `POST /api/v1/bootstrap-tokens/purge`

---

## Service Discovery API

Prometheus service discovery endpoints.

### HTTP Service Discovery

**Get All Prometheus Targets:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus
```

**Get Active Prometheus Targets:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus/active
```

**Get Prometheus Targets by Group:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus/group/GROUP_ID
```

**Response (Prometheus HTTP SD format):**
```json
[
  {
    "targets": ["192.168.1.100:9100"],
    "labels": {
      "__meta_aami_hostname": "gpu-node-01",
      "__meta_aami_group": "gpu-servers"
    }
  }
]
```

### File Service Discovery

**Generate File SD (All Targets):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file
```

**Generate File SD (Active Only):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file/active
```

**Generate File SD (By Group):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file/group/GROUP_ID
```

---

## Prometheus Management API

Manage Prometheus rule files and configuration.

### Get Prometheus Status

**Endpoint:** `GET /api/v1/prometheus/status`

```bash
curl http://localhost:8080/api/v1/prometheus/status
```

### List Rule Files

**Endpoint:** `GET /api/v1/prometheus/rules/files`

```bash
curl http://localhost:8080/api/v1/prometheus/rules/files
```

### Get Effective Rules for Target

**Endpoint:** `GET /api/v1/prometheus/rules/effective/:target_id`

```bash
curl http://localhost:8080/api/v1/prometheus/rules/effective/TARGET_ID
```

### Regenerate All Rules

**Endpoint:** `POST /api/v1/prometheus/rules/regenerate`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/rules/regenerate
```

### Regenerate Group Rules

**Endpoint:** `POST /api/v1/prometheus/rules/regenerate/:group_id`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/rules/regenerate/GROUP_ID
```

### Reload Prometheus

**Endpoint:** `POST /api/v1/prometheus/reload`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/reload
```

---

## Error Responses

All API errors follow a consistent format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "details": "Group with ID 'xxx' not found"
  }
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `BAD_REQUEST` | Invalid request body or parameters |
| 400 | `VALIDATION_ERROR` | Request validation failed |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists |
| 500 | `INTERNAL_ERROR` | Internal server error |

---

## Common Patterns

### Soft Delete

All resources support soft delete. Deleted resources are marked with `deleted_at` timestamp and can be restored.

```bash
# Soft delete
POST /api/v1/{resource}/delete
{"id": "resource-id"}

# Restore
POST /api/v1/{resource}/restore
{"id": "resource-id"}

# Hard delete (permanent)
POST /api/v1/{resource}/purge
{"id": "resource-id"}
```

### Pagination

List endpoints support pagination:

```bash
GET /api/v1/targets?page=1&limit=20
```

Response includes pagination metadata in headers or response body.
