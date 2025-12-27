# API 레퍼런스

AAMI Config Server의 완전한 REST API 문서입니다.

## Base URL

```
http://localhost:8080/api/v1
```

## 인증

현재 개발 모드에서는 API 인증이 필요하지 않습니다. 프로덕션 배포 시에는 API 키 또는 OAuth 인증을 구현하세요.

**향후 인증 헤더:**
```http
Authorization: Bearer YOUR_API_KEY
```

## 목차

1. [헬스 체크](#헬스-체크)
2. [그룹 API](#그룹-api)
3. [타겟 API](#타겟-api)
4. [알림 규칙 API](#알림-규칙-api)
5. [서비스 디스커버리 API](#서비스-디스커버리-api)
6. [부트스트랩 API](#부트스트랩-api)
7. [에러 응답](#에러-응답)

---

## 헬스 체크

### API 헬스 확인

Config Server의 헬스 상태를 조회합니다.

**엔드포인트:** `GET /health`

**예제:**
```bash
curl http://localhost:8080/api/v1/health
```

**응답:**
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

## 그룹 API

모니터링 그룹과 계층 구조를 관리합니다.

### 전체 그룹 목록 조회

**엔드포인트:** `GET /groups`

**쿼리 파라미터:**
- `namespace` (선택): 네임스페이스로 필터링 (infrastructure, logical, environment)
- `parent_id` (선택): 부모 그룹으로 필터링
- `page` (선택): 페이지 번호 (기본값: 1)
- `limit` (선택): 페이지당 항목 수 (기본값: 50)

**예제:**
```bash
# 전체 그룹 조회
curl http://localhost:8080/api/v1/groups

# 네임스페이스로 필터링
curl http://localhost:8080/api/v1/groups?namespace=environment

# 부모 그룹으로 필터링
curl http://localhost:8080/api/v1/groups?parent_id=GROUP_ID
```

**응답:**
```json
{
  "groups": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "production",
      "namespace": "environment",
      "parent_id": null,
      "description": "프로덕션 환경",
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

### ID로 그룹 조회

**엔드포인트:** `GET /groups/:id`

**예제:**
```bash
curl http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

**응답:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "프로덕션 환경",
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

### 그룹 생성

**엔드포인트:** `POST /groups`

**요청 본문:**
```json
{
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "프로덕션 환경",
  "metadata": {
    "owner": "devops-team",
    "contact": "devops@example.com"
  }
}
```

**예제:**
```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "namespace": "environment",
    "parent_id": null,
    "description": "프로덕션 환경"
  }'
```

**응답:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production",
  "namespace": "environment",
  "parent_id": null,
  "description": "프로덕션 환경",
  "priority": 10,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 그룹 수정

**엔드포인트:** `PUT /groups/:id`

**요청 본문:**
```json
{
  "name": "production-updated",
  "description": "업데이트된 프로덕션 환경"
}
```

**예제:**
```bash
curl -X PUT http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "업데이트된 프로덕션 환경"
  }'
```

### 그룹 삭제

**엔드포인트:** `DELETE /groups/:id`

**예제:**
```bash
curl -X DELETE http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

**응답:**
```json
{
  "message": "그룹이 성공적으로 삭제되었습니다"
}
```

---

## 타겟 API

모니터링 타겟(서버, 노드)을 관리합니다.

### 전체 타겟 목록 조회

**엔드포인트:** `GET /targets`

**쿼리 파라미터:**
- `group_id` (선택): 그룹으로 필터링
- `status` (선택): 상태로 필터링 (active, inactive, down)
- `page` (선택): 페이지 번호
- `limit` (선택): 페이지당 항목 수

**예제:**
```bash
# 전체 타겟 조회
curl http://localhost:8080/api/v1/targets

# 그룹으로 필터링
curl http://localhost:8080/api/v1/targets?group_id=GROUP_ID

# 상태로 필터링
curl http://localhost:8080/api/v1/targets?status=active
```

**응답:**
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

### ID로 타겟 조회

**엔드포인트:** `GET /targets/:id`

**예제:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001
```

### 타겟 생성

**엔드포인트:** `POST /targets`

**요청 본문:**
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

**예제:**
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

### 타겟 수정

**엔드포인트:** `PUT /targets/:id`

**예제:**
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

### 타겟 삭제

**엔드포인트:** `DELETE /targets/:id`

**예제:**
```bash
curl -X DELETE http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001
```

---

## 알림 규칙 API

알림 규칙과 임계값을 관리합니다.

### 알림 규칙 템플릿 목록 조회

**엔드포인트:** `GET /alert-templates`

**예제:**
```bash
curl http://localhost:8080/api/v1/alert-templates
```

**응답:**
```json
{
  "templates": [
    {
      "id": "HighCPUUsage",
      "name": "높은 CPU 사용률",
      "description": "CPU 사용률이 임계값을 초과할 때 알림",
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

### 그룹에 알림 규칙 적용

**엔드포인트:** `POST /groups/:id/alert-rules`

**요청 본문:**
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
    "summary": "높은 CPU 사용률 감지",
    "description": "CPU 사용률이 5분 이상 70%를 초과했습니다"
  }
}
```

**예제:**
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

### 타겟의 유효한 알림 규칙 조회

**엔드포인트:** `GET /targets/:id/alert-rules/effective`

**예제:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001/alert-rules/effective
```

**응답:**
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

### 알림 규칙 정책 추적

**엔드포인트:** `GET /targets/:id/alert-rules/trace`

어떤 그룹들이 최종 알림 설정에 기여했는지 보여줍니다.

**예제:**
```bash
curl http://localhost:8080/api/v1/targets/660e8400-e29b-41d4-a716-446655440001/alert-rules/trace
```

**응답:**
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

## 서비스 디스커버리 API

Prometheus 서비스 디스커버리 통합을 위한 엔드포인트입니다.

### Prometheus SD 타겟 조회

**엔드포인트:** `GET /sd/prometheus`

Prometheus 파일 기반 서비스 디스커버리 형식으로 타겟을 반환합니다.

**예제:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus
```

**응답:**
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

### Prometheus 알림 규칙 조회

**엔드포인트:** `GET /sd/alert-rules`

Prometheus 규칙 형식으로 알림 규칙을 반환합니다.

**예제:**
```bash
curl http://localhost:8080/api/v1/sd/alert-rules
```

---

## 부트스트랩 API

자동 노드 등록을 위한 엔드포인트입니다.

### 부트스트랩 토큰 생성

**엔드포인트:** `POST /bootstrap/tokens`

**요청 본문:**
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

**예제:**
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

**응답:**
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

### 부트스트랩 등록

**엔드포인트:** `POST /bootstrap/register`

부트스트랩 스크립트가 노드를 자동 등록하는 데 사용됩니다.

**요청 본문:**
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

## 에러 응답

모든 에러 응답은 다음 형식을 따릅니다:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "사람이 읽을 수 있는 에러 메시지",
    "details": {}
  }
}
```

### 일반적인 에러 코드

| 코드 | HTTP 상태 | 설명 |
|------|----------|------|
| `INVALID_REQUEST` | 400 | 잘못된 요청 본문 또는 파라미터 |
| `NOT_FOUND` | 404 | 리소스를 찾을 수 없음 |
| `CONFLICT` | 409 | 리소스가 이미 존재함 |
| `VALIDATION_ERROR` | 422 | 검증 실패 |
| `INTERNAL_ERROR` | 500 | 내부 서버 에러 |
| `DATABASE_ERROR` | 500 | 데이터베이스 작업 실패 |

**에러 응답 예제:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "그룹을 찾을 수 없습니다",
    "details": {
      "group_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

---

## 속도 제한

현재 개발 모드에서는 속도 제한이 적용되지 않습니다. 프로덕션 배포 시에는 API 게이트웨이 또는 로드 밸런서 레벨에서 속도 제한을 구현하세요.

**권장 헤더:**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1609459200
```

---

## API 버전 관리

API는 URL 기반 버전 관리를 사용합니다 (`/api/v1`). 호환성이 깨지는 변경사항이 있을 경우 새 버전이 생성됩니다 (`/api/v2`).

---

더 많은 예제는 [API 사용 예제](../../examples/api-usage/)를 참조하세요.
