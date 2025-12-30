# API 레퍼런스

AAMI Config Server의 완전한 REST API 문서입니다.

## Base URL

```
http://localhost:8080/api/v1
```

## 인증

현재 API는 인증이 필요하지 않습니다. 프로덕션 배포 시 API 키 또는 OAuth 인증을 구현하세요.

## 목차

1. [헬스 체크](#헬스-체크)
2. [그룹 API](#그룹-api)
3. [타겟 API](#타겟-api)
4. [익스포터 API](#익스포터-api)
5. [알림 템플릿 API](#알림-템플릿-api)
6. [알림 규칙 API](#알림-규칙-api)
7. [활성 알림 API](#활성-알림-api)
8. [스크립트 템플릿 API](#스크립트-템플릿-api)
9. [스크립트 정책 API](#스크립트-정책-api)
10. [부트스트랩 토큰 API](#부트스트랩-토큰-api)
11. [서비스 디스커버리 API](#서비스-디스커버리-api)
12. [Prometheus 관리 API](#prometheus-관리-api)
13. [에러 응답](#에러-응답)

---

## 헬스 체크

### API 상태 확인

**엔드포인트:** `GET /health`

```bash
curl http://localhost:8080/health
```

**응답:**
```json
{
  "status": "healthy",
  "version": "v1.0.0",
  "database": "connected"
}
```

### 준비 상태 확인

**엔드포인트:** `GET /health/ready`

### 활성 상태 확인

**엔드포인트:** `GET /health/live`

---

## 그룹 API

모니터링 그룹을 관리합니다. 그룹은 플랫(비계층) 구조이며 타겟은 여러 그룹에 속할 수 있습니다.

### 전체 그룹 목록 조회

**엔드포인트:** `GET /api/v1/groups`

```bash
curl http://localhost:8080/api/v1/groups
```

**응답:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "gpu-servers",
    "description": "GPU 컴퓨팅 서버",
    "priority": 10,
    "is_default_own": false,
    "metadata": {},
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

### ID로 그룹 조회

**엔드포인트:** `GET /api/v1/groups/:id`

```bash
curl http://localhost:8080/api/v1/groups/550e8400-e29b-41d4-a716-446655440000
```

### 그룹 생성

**엔드포인트:** `POST /api/v1/groups`

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-servers",
    "description": "웹 애플리케이션 서버",
    "priority": 20,
    "metadata": {
      "environment": "production"
    }
  }'
```

### 그룹 수정

**엔드포인트:** `PUT /api/v1/groups/:id`

```bash
curl -X PUT http://localhost:8080/api/v1/groups/GROUP_ID \
  -H "Content-Type: application/json" \
  -d '{
    "description": "업데이트된 설명",
    "priority": 15
  }'
```

### 그룹 삭제 (소프트 삭제)

**엔드포인트:** `POST /api/v1/groups/delete`

```bash
curl -X POST http://localhost:8080/api/v1/groups/delete \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

### 그룹 복원

**엔드포인트:** `POST /api/v1/groups/restore`

```bash
curl -X POST http://localhost:8080/api/v1/groups/restore \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

### 그룹 영구 삭제 (하드 삭제)

**엔드포인트:** `POST /api/v1/groups/purge`

```bash
curl -X POST http://localhost:8080/api/v1/groups/purge \
  -H "Content-Type: application/json" \
  -d '{"id": "GROUP_ID"}'
```

---

## 타겟 API

모니터링 타겟(노드/서버)을 관리합니다.

### 전체 타겟 목록 조회

**엔드포인트:** `GET /api/v1/targets`

```bash
curl http://localhost:8080/api/v1/targets
```

**응답:**
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

### ID로 타겟 조회

**엔드포인트:** `GET /api/v1/targets/:id`

### 호스트네임으로 타겟 조회

**엔드포인트:** `GET /api/v1/targets/hostname/:hostname`

```bash
curl http://localhost:8080/api/v1/targets/hostname/gpu-node-01
```

### 그룹별 타겟 조회

**엔드포인트:** `GET /api/v1/targets/group/:group_id`

```bash
curl http://localhost:8080/api/v1/targets/group/GROUP_ID
```

### 타겟 생성

**엔드포인트:** `POST /api/v1/targets`

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

### 타겟 수정

**엔드포인트:** `PUT /api/v1/targets/:id`

### 타겟 상태 업데이트

**엔드포인트:** `POST /api/v1/targets/:id/status`

```bash
curl -X POST http://localhost:8080/api/v1/targets/TARGET_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "inactive"}'
```

### 타겟 하트비트

**엔드포인트:** `POST /api/v1/targets/:id/heartbeat`

### 타겟 삭제/복원/영구삭제

- `POST /api/v1/targets/delete`
- `POST /api/v1/targets/restore`
- `POST /api/v1/targets/purge`

---

## 익스포터 API

타겟에 연결된 Prometheus 익스포터를 관리합니다.

### 전체 익스포터 목록 조회

**엔드포인트:** `GET /api/v1/exporters`

### ID로 익스포터 조회

**엔드포인트:** `GET /api/v1/exporters/:id`

### 타겟별 익스포터 조회

**엔드포인트:** `GET /api/v1/exporters/target/:target_id`

### 타입별 익스포터 조회

**엔드포인트:** `GET /api/v1/exporters/type/:type`

```bash
curl http://localhost:8080/api/v1/exporters/type/node_exporter
```

### 익스포터 생성

**엔드포인트:** `POST /api/v1/exporters`

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

### 익스포터 수정/삭제/복원/영구삭제

- `PUT /api/v1/exporters/:id`
- `POST /api/v1/exporters/delete`
- `POST /api/v1/exporters/restore`
- `POST /api/v1/exporters/purge`

---

## 알림 템플릿 API

재사용 가능한 알림 규칙 템플릿을 관리합니다.

### 전체 알림 템플릿 목록 조회

**엔드포인트:** `GET /api/v1/alert-templates`

### ID로 알림 템플릿 조회

**엔드포인트:** `GET /api/v1/alert-templates/:id`

### 심각도별 알림 템플릿 조회

**엔드포인트:** `GET /api/v1/alert-templates/severity/:severity`

```bash
curl http://localhost:8080/api/v1/alert-templates/severity/critical
```

### 알림 템플릿 생성

**엔드포인트:** `POST /api/v1/alert-templates`

```bash
curl -X POST http://localhost:8080/api/v1/alert-templates \
  -H "Content-Type: application/json" \
  -d '{
    "id": "high-cpu",
    "name": "높은 CPU 사용률",
    "description": "CPU 사용률이 임계값을 초과할 때 알림",
    "severity": "warning",
    "query_template": "100 - (avg by(instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > {{.threshold}}",
    "default_config": {
      "threshold": 80
    }
  }'
```

### 알림 템플릿 수정/삭제/복원/영구삭제

- `PUT /api/v1/alert-templates/:id`
- `POST /api/v1/alert-templates/delete`
- `POST /api/v1/alert-templates/restore`
- `POST /api/v1/alert-templates/purge`

---

## 알림 규칙 API

그룹에 할당된 알림 규칙을 관리합니다.

### 전체 알림 규칙 목록 조회

**엔드포인트:** `GET /api/v1/alert-rules`

### ID로 알림 규칙 조회

**엔드포인트:** `GET /api/v1/alert-rules/:id`

### 그룹별 알림 규칙 조회

**엔드포인트:** `GET /api/v1/alert-rules/group/:group_id`

### 템플릿별 알림 규칙 조회

**엔드포인트:** `GET /api/v1/alert-rules/template/:template_id`

### 알림 규칙 생성 (템플릿 기반)

**엔드포인트:** `POST /api/v1/alert-rules`

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

### 알림 규칙 직접 생성

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "group-uuid",
    "name": "커스텀 알림",
    "description": "커스텀 알림 규칙",
    "severity": "critical",
    "query_template": "up == 0",
    "enabled": true,
    "priority": 100
  }'
```

### 알림 규칙 수정/삭제/복원/영구삭제

- `PUT /api/v1/alert-rules/:id`
- `POST /api/v1/alert-rules/delete`
- `POST /api/v1/alert-rules/restore`
- `POST /api/v1/alert-rules/purge`

---

## 활성 알림 API

Alertmanager에서 현재 발생 중인 알림을 조회합니다.

### 활성 알림 조회

**엔드포인트:** `GET /api/v1/alerts/active`

```bash
curl http://localhost:8080/api/v1/alerts/active
```

**응답:**
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
        "summary": "높은 CPU 사용률 감지",
        "description": "CPU 사용률이 90%를 초과했습니다"
      },
      "starts_at": "2024-01-15T10:30:00Z",
      "generator_url": "http://prometheus:9090/graph?..."
    }
  ],
  "total": 1
}
```

---

## 스크립트 템플릿 API

체크 스크립트 템플릿을 관리합니다.

### 전체 스크립트 템플릿 목록 조회

**엔드포인트:** `GET /api/v1/script-templates`

### 활성 스크립트 템플릿 목록 조회

**엔드포인트:** `GET /api/v1/script-templates/active`

### ID로 스크립트 템플릿 조회

**엔드포인트:** `GET /api/v1/script-templates/:id`

### 이름으로 스크립트 템플릿 조회

**엔드포인트:** `GET /api/v1/script-templates/name/:name`

### 타입별 스크립트 템플릿 조회

**엔드포인트:** `GET /api/v1/script-templates/type/:scriptType`

```bash
curl http://localhost:8080/api/v1/script-templates/type/check
```

### 스크립트 해시 검증

**엔드포인트:** `GET /api/v1/script-templates/:id/verify-hash`

### 스크립트 템플릿 생성

**엔드포인트:** `POST /api/v1/script-templates`

```bash
curl -X POST http://localhost:8080/api/v1/script-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "disk-check",
    "description": "디스크 사용률 체크",
    "script_type": "check",
    "script_content": "#!/bin/bash\ndf -h / | awk '\''NR==2 {print $5}'\''",
    "config_schema": {
      "threshold": {"type": "number", "default": 80}
    },
    "enabled": true
  }'
```

### 스크립트 템플릿 수정/삭제/복원/영구삭제

- `PUT /api/v1/script-templates/:id`
- `POST /api/v1/script-templates/delete`
- `POST /api/v1/script-templates/restore`
- `POST /api/v1/script-templates/purge`

---

## 스크립트 정책 API

그룹에 스크립트 정책 할당을 관리합니다.

### 전체 스크립트 정책 목록 조회

**엔드포인트:** `GET /api/v1/script-policies`

### 활성 스크립트 정책 목록 조회

**엔드포인트:** `GET /api/v1/script-policies/active`

### ID로 스크립트 정책 조회

**엔드포인트:** `GET /api/v1/script-policies/:id`

### 템플릿별 스크립트 정책 조회

**엔드포인트:** `GET /api/v1/script-policies/template/:templateId`

### 글로벌 스크립트 정책 조회

**엔드포인트:** `GET /api/v1/script-policies/global`

### 그룹별 스크립트 정책 조회

**엔드포인트:** `GET /api/v1/script-policies/group/:groupId`

### 그룹의 유효 체크 조회

**엔드포인트:** `GET /api/v1/script-policies/effective/group/:groupId`

### 타겟의 유효 체크 조회

**엔드포인트:** `GET /api/v1/checks/target/:targetId`

노드가 할당된 체크를 조회하는 데 사용됩니다.

```bash
curl http://localhost:8080/api/v1/checks/target/TARGET_ID
```

### 스크립트 정책 생성

**엔드포인트:** `POST /api/v1/script-policies`

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

### 스크립트 정책 수정/삭제/복원/영구삭제

- `PUT /api/v1/script-policies/:id`
- `POST /api/v1/script-policies/delete`
- `POST /api/v1/script-policies/restore`
- `POST /api/v1/script-policies/purge`

---

## 부트스트랩 토큰 API

노드 자동 등록을 위한 부트스트랩 토큰을 관리합니다.

### 전체 부트스트랩 토큰 목록 조회

**엔드포인트:** `GET /api/v1/bootstrap-tokens`

### ID로 부트스트랩 토큰 조회

**엔드포인트:** `GET /api/v1/bootstrap-tokens/:id`

### 토큰 문자열로 부트스트랩 토큰 조회

**엔드포인트:** `GET /api/v1/bootstrap-tokens/token/:token`

### 부트스트랩 토큰 생성

**엔드포인트:** `POST /api/v1/bootstrap-tokens`

```bash
curl -X POST http://localhost:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpu-cluster-token",
    "description": "GPU 클러스터 노드용 토큰",
    "group_id": "gpu-servers-group-id",
    "expires_at": "2024-12-31T23:59:59Z",
    "max_uses": 100
  }'
```

**응답:**
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

### 토큰 검증 및 사용

**엔드포인트:** `POST /api/v1/bootstrap-tokens/validate`

```bash
curl -X POST http://localhost:8080/api/v1/bootstrap-tokens/validate \
  -H "Content-Type: application/json" \
  -d '{"token": "aami_bootstrap_abc123xyz..."}'
```

### 토큰으로 노드 등록

**엔드포인트:** `POST /api/v1/bootstrap-tokens/register`

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

### 부트스트랩 토큰 수정/삭제/복원/영구삭제

- `PUT /api/v1/bootstrap-tokens/:id`
- `POST /api/v1/bootstrap-tokens/delete`
- `POST /api/v1/bootstrap-tokens/restore`
- `POST /api/v1/bootstrap-tokens/purge`

---

## 서비스 디스커버리 API

Prometheus 서비스 디스커버리 엔드포인트입니다.

### HTTP 서비스 디스커버리

**전체 Prometheus 타겟 조회:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus
```

**활성 Prometheus 타겟 조회:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus/active
```

**그룹별 Prometheus 타겟 조회:**
```bash
curl http://localhost:8080/api/v1/sd/prometheus/group/GROUP_ID
```

**응답 (Prometheus HTTP SD 형식):**
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

### 파일 서비스 디스커버리

**파일 SD 생성 (전체 타겟):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file
```

**파일 SD 생성 (활성 타겟만):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file/active
```

**파일 SD 생성 (그룹별):**
```bash
curl -X POST http://localhost:8080/api/v1/sd/prometheus/file/group/GROUP_ID
```

---

## Prometheus 관리 API

Prometheus 규칙 파일 및 구성을 관리합니다.

### Prometheus 상태 조회

**엔드포인트:** `GET /api/v1/prometheus/status`

```bash
curl http://localhost:8080/api/v1/prometheus/status
```

### 규칙 파일 목록 조회

**엔드포인트:** `GET /api/v1/prometheus/rules/files`

```bash
curl http://localhost:8080/api/v1/prometheus/rules/files
```

### 타겟의 유효 규칙 조회

**엔드포인트:** `GET /api/v1/prometheus/rules/effective/:target_id`

```bash
curl http://localhost:8080/api/v1/prometheus/rules/effective/TARGET_ID
```

### 전체 규칙 재생성

**엔드포인트:** `POST /api/v1/prometheus/rules/regenerate`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/rules/regenerate
```

### 그룹 규칙 재생성

**엔드포인트:** `POST /api/v1/prometheus/rules/regenerate/:group_id`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/rules/regenerate/GROUP_ID
```

### Prometheus 리로드

**엔드포인트:** `POST /api/v1/prometheus/reload`

```bash
curl -X POST http://localhost:8080/api/v1/prometheus/reload
```

---

## 에러 응답

모든 API 에러는 일관된 형식을 따릅니다:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "리소스를 찾을 수 없습니다",
    "details": "ID 'xxx'인 그룹을 찾을 수 없습니다"
  }
}
```

### 일반 에러 코드

| HTTP 상태 | 코드 | 설명 |
|-----------|------|------|
| 400 | `BAD_REQUEST` | 잘못된 요청 본문 또는 파라미터 |
| 400 | `VALIDATION_ERROR` | 요청 검증 실패 |
| 404 | `NOT_FOUND` | 리소스를 찾을 수 없음 |
| 409 | `CONFLICT` | 리소스가 이미 존재함 |
| 500 | `INTERNAL_ERROR` | 내부 서버 에러 |

---

## 공통 패턴

### 소프트 삭제

모든 리소스는 소프트 삭제를 지원합니다. 삭제된 리소스는 `deleted_at` 타임스탬프로 표시되며 복원할 수 있습니다.

```bash
# 소프트 삭제
POST /api/v1/{resource}/delete
{"id": "resource-id"}

# 복원
POST /api/v1/{resource}/restore
{"id": "resource-id"}

# 하드 삭제 (영구)
POST /api/v1/{resource}/purge
{"id": "resource-id"}
```

### 페이지네이션

목록 엔드포인트는 페이지네이션을 지원합니다:

```bash
GET /api/v1/targets?page=1&limit=20
```

응답에는 헤더 또는 응답 본문에 페이지네이션 메타데이터가 포함됩니다.
