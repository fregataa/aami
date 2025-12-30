# Prometheus 통합 가이드

이 가이드는 동적 알림 규칙 관리를 위한 AAMI Config Server와 Prometheus 통합을 다룹니다.

## 목차

1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [설정](#설정)
4. [환경 설정](#환경-설정)
5. [API 레퍼런스](#api-레퍼런스)
6. [문제 해결](#문제-해결)
7. [모범 사례](#모범-사례)

---

## 개요

AAMI Config Server는 동적 알림 규칙 관리를 위해 Prometheus와 통합됩니다. 이 통합을 통해:

- **동적 규칙 생성**: 데이터베이스에 정의된 AlertRule이 자동으로 Prometheus 규칙 파일로 변환됨
- **그룹 기반 커스터마이징**: 레이블 기반 필터링을 통해 그룹별로 다른 알림 임계값 설정
- **무중단 업데이트**: 서비스 중단 없이 Prometheus 규칙 리로드
- **중앙 집중식 관리**: 단일 API를 통해 모든 알림 규칙 관리

### 주요 구성 요소

| 구성 요소 | 설명 |
|-----------|------|
| **Prometheus Rule Generator** | AlertRule을 Prometheus YAML 형식으로 변환 |
| **Rule File Manager** | 검증 및 백업과 함께 원자적 파일 쓰기 처리 |
| **Prometheus Client** | 설정 리로드 및 헬스체크 트리거 |
| **AlertRule API** | 그룹별 알림 규칙 CRUD 작업 |

---

## 아키텍처

```
┌─────────────────────────────────────────────────────────────────┐
│                     AAMI Config Server                          │
│  ┌─────────────────┐  ┌────────────────┐  ┌─────────────────┐  │
│  │  AlertRule API  │  │ Rule Generator │  │ Prometheus Client│  │
│  │  (CRUD)         │→ │ (YAML 생성)    │→ │ (리로드)        │  │
│  └─────────────────┘  └────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                       │                    │
         ▼                       ▼                    ▼
┌─────────────┐         ┌─────────────────┐   ┌─────────────────┐
│  PostgreSQL │         │ 공유 볼륨       │   │   Prometheus    │
│  (AlertRule │         │ /rules/generated│   │   /-/reload     │
│   저장소)   │         │ *.yml 파일      │   │   /-/ready      │
└─────────────┘         └─────────────────┘   └─────────────────┘
```

### 데이터 흐름

1. 사용자가 API를 통해 AlertRule 생성/수정
2. Rule Generator가 AlertRule을 Prometheus YAML로 변환
3. File Manager가 규칙 파일을 원자적으로 작성 (temp → rename)
4. Prometheus Client가 `/-/reload` 엔드포인트 트리거
5. Prometheus가 재시작 없이 새 규칙 로드

---

## 설정

### 사전 요구 사항

- `--web.enable-lifecycle` 플래그가 활성화된 Prometheus 2.x+
- Config Server와 Prometheus 간 공유 볼륨
- Config Server에서 Prometheus API로의 네트워크 접근

### Docker Compose 설정

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
      - '--web.enable-lifecycle'  # 리로드 API에 필요
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

### Prometheus 설정

생성된 규칙 디렉토리를 Prometheus 설정에 추가:

```yaml
# config/prometheus/prometheus.yml
rule_files:
  - /etc/prometheus/rules/*.yml           # 정적 규칙
  - /etc/prometheus/rules/generated/*.yml # Config Server의 동적 규칙
```

### Kubernetes 설정

```yaml
# Prometheus ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    rule_files:
      - /etc/prometheus/rules/generated/*.yml

---
# 공유 규칙용 PersistentVolumeClaim
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
# Prometheus와 Config Server 배포에서 마운트
# Prometheus: /etc/prometheus/rules/generated (읽기 전용)
# Config Server: /app/rules (읽기-쓰기)
```

---

## 환경 설정

### 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `PROMETHEUS_URL` | `http://localhost:9090` | Prometheus 서버 URL |
| `PROMETHEUS_RULE_PATH` | `/etc/prometheus/rules/generated` | 생성된 규칙 파일 디렉토리 |
| `PROMETHEUS_RELOAD_ENABLED` | `true` | 자동 Prometheus 리로드 활성화 |
| `PROMETHEUS_RELOAD_TIMEOUT` | `30s` | 리로드 API 호출 타임아웃 |
| `PROMETHEUS_VALIDATE_RULES` | `false` | 쓰기 전 promtool 검증 활성화 |
| `PROMETHEUS_BACKUP_ENABLED` | `true` | 이전 규칙 파일 백업 유지 |
| `PROMTOOL_PATH` | `promtool` | promtool 바이너리 경로 (검증 활성화 시) |

### 설정 파일 (config.yaml)

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

## API 레퍼런스

### Prometheus 규칙 관리 엔드포인트

#### 전체 규칙 재생성

모든 그룹의 Prometheus 규칙 파일을 재생성합니다.

```bash
POST /api/v1/prometheus/rules/regenerate
```

**응답:**
```json
{
  "groups_affected": 5,
  "files_generated": 5,
  "errors": [],
  "duration": "1.234s"
}
```

#### 그룹 규칙 재생성

특정 그룹의 Prometheus 규칙 파일을 재생성합니다.

```bash
POST /api/v1/prometheus/rules/regenerate/:group_id
```

**응답:**
```json
{
  "group_id": "grp-123",
  "file_name": "group-grp-123.yml",
  "rules_count": 3,
  "duration": "0.456s"
}
```

#### 규칙 파일 목록

생성된 모든 Prometheus 규칙 파일을 나열합니다.

```bash
GET /api/v1/prometheus/rules/files
```

**응답:**
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

#### 타겟의 유효 규칙 조회

그룹 멤버십을 고려하여 특정 타겟에 적용되는 모든 알림 규칙을 반환합니다.

```bash
GET /api/v1/prometheus/rules/effective/:target_id
```

**응답:**
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
        "summary": "높은 CPU 사용량 감지"
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

#### Prometheus 리로드 트리거

수동으로 Prometheus 설정 리로드를 트리거합니다.

```bash
POST /api/v1/prometheus/reload
```

**응답:**
```json
{
  "status": "success",
  "message": "Prometheus 리로드가 성공적으로 트리거되었습니다"
}
```

#### Prometheus 상태 확인

Prometheus 연결 상태를 확인합니다.

```bash
GET /api/v1/prometheus/status
```

**응답:**
```json
{
  "status": "healthy",
  "url": "http://prometheus:9090",
  "ready": true,
  "healthy": true
}
```

---

## 문제 해결

### 일반적인 문제

#### 1. Prometheus에서 규칙이 로드되지 않음

**증상:**
- API에서 AlertRule이 생성되었지만 Prometheus에서 보이지 않음
- `/api/v1/prometheus/rules/files`에서 파일이 표시되지만 Prometheus가 인식하지 못함

**해결 방법:**

1. 볼륨 마운트 확인:
```bash
# Config Server 컨테이너에서
ls -la /app/rules/

# Prometheus 컨테이너에서
ls -la /etc/prometheus/rules/generated/
```

2. Prometheus 설정 확인:
```bash
curl http://prometheus:9090/api/v1/status/config | jq '.data.yaml' | grep rule_files
```

3. Prometheus 리로드 확인:
```bash
curl -X POST http://prometheus:9090/-/reload
```

4. Prometheus 오류 확인:
```bash
docker logs prometheus 2>&1 | grep -i "rule\|error"
```

#### 2. 리로드 API가 작동하지 않음

**증상:**
- 오류: "Lifecycle API is not enabled"

**해결 방법:**
Prometheus가 `--web.enable-lifecycle` 플래그로 시작되었는지 확인:

```yaml
# docker-compose.yml
prometheus:
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--web.enable-lifecycle'
```

#### 3. 규칙 검증 실패

**증상:**
- 규칙 생성은 성공하지만 Prometheus가 파일을 무시함
- 생성된 YAML에 구문 오류

**해결 방법:**

1. Config Server에서 규칙 검증 활성화:
```bash
PROMETHEUS_VALIDATE_RULES=true
PROMTOOL_PATH=/usr/local/bin/promtool
```

2. 수동으로 규칙 파일 검증:
```bash
promtool check rules /path/to/rule-file.yml
```

3. 일반적인 문제 확인:
   - 잘못된 PromQL 구문
   - 필수 필드 누락 (expr, alert)
   - YAML 들여쓰기 오류

#### 4. 권한 문제

**증상:**
- 규칙 파일 작성 시 "Permission denied"
- 생성 후 빈 규칙 디렉토리

**해결 방법:**

1. 디렉토리 권한 확인:
```bash
ls -la /app/rules/
```

2. Config Server 프로세스에 쓰기 권한이 있는지 확인:
```yaml
# docker-compose.yml
config-server:
  user: "1000:1000"  # 볼륨 소유자와 일치
```

3. Kubernetes의 경우 적절한 SecurityContext 사용:
```yaml
securityContext:
  runAsUser: 1000
  fsGroup: 1000
```

#### 5. Prometheus 연결 불가

**증상:**
- `/api/v1/prometheus/status`가 오류 반환
- 리로드 호출 실패

**해결 방법:**

1. 네트워크 연결 확인:
```bash
curl http://prometheus:9090/-/ready
```

2. 환경 변수 확인:
```bash
echo $PROMETHEUS_URL
```

3. Prometheus 정상 상태 확인:
```bash
curl http://prometheus:9090/-/healthy
```

### 확인할 로그

1. **Config Server 로그:**
```bash
docker logs config-server 2>&1 | grep -i prometheus
```

2. **Prometheus 로그:**
```bash
docker logs prometheus 2>&1 | grep -E "(rule|reload|error)"
```

---

## 모범 사례

### 1. 원자적 규칙 업데이트 사용

파일을 수동으로 편집하는 대신 항상 regenerate API를 사용하세요. 이를 통해:
- 원자적 파일 쓰기 (부분 쓰기 없음)
- 이전 버전 자동 백업
- 적절한 Prometheus 리로드

### 2. 규칙 생성 모니터링

규칙 생성 실패에 대한 알림 설정:

```yaml
- alert: AlertRuleGenerationFailed
  expr: aami_rule_generation_errors_total > 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "알림 규칙 생성 실패"
```

### 3. 그룹별 임계값 사용

환경별 규칙에 그룹 레이블 활용:

```yaml
# AlertTemplate query_template
(100 - avg(rate(node_cpu_seconds_total{mode="idle",group_id="{{.group_id}}"}[5m])) * 100) > {{.threshold}}
```

### 4. 프로덕션 전 규칙 테스트

1. 먼저 개발 그룹에서 규칙 생성
2. Prometheus UI에서 확인 (`/alerts`)
3. PromQL 표현식 확인 (`/graph`)
4. 프로덕션 그룹으로 승격

### 5. 백업 유지

설정에서 백업 활성화:

```bash
PROMETHEUS_BACKUP_ENABLED=true
```

규칙 파일을 덮어쓰기 전에 `.bak` 파일이 생성됩니다.

### 6. 의미 있는 알림 이름 사용

명명 규칙 준수:
- `HighCPUUsage_Production` - 그룹 컨텍스트 포함
- 관련 알림에 일관된 접두사 사용
- 심각도는 이름이 아닌 레이블에 포함

### 7. 그룹별 규칙 파일 분리

Config Server는 그룹당 하나의 파일을 자동으로 생성합니다:
```
/rules/generated/
  group-grp-123.yml    # Production 알림
  group-grp-456.yml    # Development 알림
```

이 분리로 한 그룹의 문제가 다른 그룹에 영향을 주지 않습니다.

---

## 참고 자료

- [AAMI 알림 시스템 아키텍처](./ALERTING-SYSTEM.md)
- [Prometheus 설정](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [Prometheus 알림 규칙](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [Prometheus 관리 API](https://prometheus.io/docs/prometheus/latest/management_api/)
