# 체크 관리 시스템

## 목차

1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [ScriptTemplate vs ScriptPolicy](#scripttemplate-vs-scriptpolicy)
4. [Scope 기반 관리](#scope-기반-관리)
5. [스크립트 출력 형식](#스크립트-출력-형식)
6. [워크플로우](#워크플로우)
7. [API 레퍼런스](#api-레퍼런스)
8. [예제](#예제)

## 개요

AAMI의 동적 체크 시스템은 인프라 전반에 걸쳐 커스텀 모니터링 체크를 중앙에서 관리하고 배포할 수 있게 해줍니다.

### 핵심 개념

- **ScriptTemplate**: 재사용 가능한 스크립트 정의 (스크립트 코드 + 기본 파라미터)
- **ScriptPolicy**: 그룹별 스크립트 적용 (Template 참조 + Override 파라미터)
- **Scope 기반 관리**: Global → Group 계층으로 체크 적용
- **로컬 캐싱**: 노드는 스크립트를 로컬 파일로 캐시
- **자동 업데이트**: 해시 기반 버전 감지로 스크립트 자동 갱신
- **Prometheus 메트릭 출력**: 스크립트는 Prometheus 형식으로 출력

---

## 아키텍처

### Alert 시스템과의 일관성

```
Alert 시스템:
├─ AlertTemplate (재사용 가능한 알림 규칙 정의)
└─ AlertRule (그룹별 알림 규칙 적용, Template 참조)

Script 시스템 (동일 패턴):
├─ ScriptTemplate (재사용 가능한 스크립트 정의)
└─ ScriptPolicy (그룹별 스크립트 적용, Template 참조)
```

### 전체 데이터 흐름

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         Config Server                                     │
│  ┌─────────────────┐        ┌─────────────────┐                          │
│  │ ScriptTemplate  │        │  AlertTemplate  │                          │
│  │ (스크립트 정의)  │        │  (알림 정의)     │                          │
│  └────────┬────────┘        └────────┬────────┘                          │
│           │                          │                                    │
│           ▼                          ▼                                    │
│  ┌─────────────────┐        ┌─────────────────┐                          │
│  │  ScriptPolicy   │        │   AlertRule     │                          │
│  │  (그룹에 적용)   │        │  (그룹에 적용)   │                          │
│  └────────┬────────┘        └────────┬────────┘                          │
│           │                          │                                    │
│           ▼                          ▼                                    │
│  ┌─────────────────┐        ┌─────────────────┐                          │
│  │ EffectiveCheck  │        │ PrometheusRule  │                          │
│  │  (노드가 수신)   │        │   (YAML 생성)   │                          │
│  └────────┬────────┘        └────────┬────────┘                          │
└───────────┼──────────────────────────┼───────────────────────────────────┘
            │                          │
            ▼                          ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                              Target Node                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ dynamic-check.sh                                                     │  │
│  │ 1. EffectiveCheck 조회 (GET /api/v1/checks/target/:targetId)        │  │
│  │ 2. 스크립트 로컬 캐시 저장                                           │  │
│  │ 3. 스크립트 실행 (config 전달)                                       │  │
│  │ 4. Prometheus 메트릭 출력                                            │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                       │
│                                    ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ /var/lib/node_exporter/textfile/*.prom                              │  │
│  │ mount_check{path="/mnt/data"} 1                                     │  │
│  │ mount_check{path="/mnt/backup"} 0                                   │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                       │
│                                    ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ Node Exporter (textfile collector)                                   │  │
│  │ - 메트릭을 Prometheus 형식으로 노출                                  │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼ scrape
┌───────────────────────────────────────────────────────────────────────────┐
│                            Prometheus                                      │
│  - 메트릭 수집                                                             │
│  - AlertRule 평가 (mount_check == 0 → 알림)                               │
│  - Alertmanager로 알림 전송                                               │
└───────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                           Alertmanager                                     │
│  - 알림 라우팅                                                             │
│  - 그룹화/억제/중복제거                                                    │
│  - Email, Slack, PagerDuty 전송                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

---

## ScriptTemplate vs ScriptPolicy

### ScriptTemplate (스크립트 정의)

**목적**: 재사용 가능한 모니터링 스크립트 정의

**구조**:
```go
type ScriptTemplate struct {
    ID            string
    Name          string                  // "check-mount-points"
    ScriptType    string                  // "mount"
    ScriptContent string                  // 스크립트 코드
    Language      string                  // "bash", "python"
    DefaultConfig map[string]interface{}  // 기본 파라미터
    Description   string
    Version       string
    Hash          string                  // SHA256 hash
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**특징**:
- 한 번 정의, 여러 그룹에서 재사용
- 버전 관리 및 히스토리
- 스크립트 로직 중앙 관리
- Hash로 변경 감지

---

### ScriptPolicy (스크립트 적용)

**목적**: 특정 그룹에 Template 적용

**구조**:
```go
type ScriptPolicy struct {
    ID          string

    // Template 필드 (생성 시 복사됨)
    Name          string
    ScriptType    string
    ScriptContent string
    Language      string
    DefaultConfig map[string]interface{}
    Version       string
    Hash          string

    // Policy 필드
    Scope       string                  // "global", "group"
    GroupID     *string                 // Group 레벨일 때
    Config      map[string]interface{}  // Override 파라미터
    Priority    int                     // 우선순위 (높을수록 우선)
    IsActive    bool

    // 메타데이터
    CreatedFromTemplateID   *string
    CreatedFromTemplateName *string

    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**특징**:
- Template을 참조하여 적용
- 파라미터를 Override 가능
- Scope별 우선순위 해석
- Template과 독립적 (생성 시 복사)

---

## Scope 기반 관리

### Scope 우선순위

ScriptPolicy는 2가지 Scope를 지원하며, 더 구체적인 Scope가 우선합니다:

```
Group (가장 구체적, 최우선)
  ↑
Global (가장 일반적, 최후순위)
```

### Scope 해석 로직

노드가 체크 스크립트를 조회할 때:

1. **Target의 Group 확인**
2. **Group 레벨 ScriptPolicy 조회**
   - 존재하면 사용
3. **Global 레벨 ScriptPolicy 조회** (Group에 없으면)
   - 존재하면 사용
4. **없으면 해당 체크 없음**

### 예시: Mount Check 적용

```
Global ScriptPolicy:
  template: check-mount-points
  config: { mount_points: ["/data"] }

Group "gpu-cluster" ScriptPolicy:
  template: check-mount-points
  config: { mount_points: ["/mnt/models", "/mnt/scratch"] }  # Override

Group "storage-cluster" ScriptPolicy:
  (없음)

결과:
├─ gpu-cluster 그룹 노드 → mount_points: ["/mnt/models", "/mnt/scratch"] (Group)
└─ storage-cluster 그룹 노드 → mount_points: ["/data"] (Global fallback)
```

---

## 스크립트 출력 형식

### Prometheus 메트릭 형식 (권장)

체크 스크립트는 Prometheus text format으로 출력합니다:

```
# HELP mount_check Mount point accessibility (1=ok, 0=fail)
# TYPE mount_check gauge
mount_check{path="/mnt/data"} 1
mount_check{path="/mnt/backup"} 0
```

### 스크립트 작성 가이드

#### Bash 스크립트 예시

```bash
#!/usr/bin/env bash
# Mount Point Check Script
# Config는 환경변수 CONFIG로 전달됨

set -euo pipefail

# JSON config에서 mount_points 추출
mount_points=$(echo "$CONFIG" | jq -r '.mount_points[]' 2>/dev/null || echo "")

# Prometheus 메트릭 헤더
echo "# HELP mount_check Mount point accessibility (1=ok, 0=fail)"
echo "# TYPE mount_check gauge"

# 각 마운트 포인트 체크
for mp in $mount_points; do
    if mountpoint -q "$mp" 2>/dev/null && [ -w "$mp" ]; then
        echo "mount_check{path=\"$mp\"} 1"
    else
        echo "mount_check{path=\"$mp\"} 0"
    fi
done
```

#### Python 스크립트 예시

```python
#!/usr/bin/env python3
import os
import json

config = json.loads(os.environ.get('CONFIG', '{}'))
mount_points = config.get('mount_points', [])

print('# HELP mount_check Mount point accessibility (1=ok, 0=fail)')
print('# TYPE mount_check gauge')

for mp in mount_points:
    if os.path.ismount(mp) and os.access(mp, os.W_OK):
        print(f'mount_check{{path="{mp}"}} 1')
    else:
        print(f'mount_check{{path="{mp}"}} 0')
```

---

## 워크플로우

### 1. ScriptTemplate 생성 (관리자)

```bash
POST /api/v1/script-templates
{
  "name": "check-mount-points",
  "script_type": "mount",
  "script_content": "#!/bin/bash\n...",
  "language": "bash",
  "version": "1.0.0",
  "default_config": {
    "mount_points": ["/data"]
  },
  "description": "마운트 포인트 접근성 체크"
}
```

### 2. ScriptPolicy 생성 (관리자)

```bash
# GPU 클러스터에 적용
POST /api/v1/script-policies
{
  "template_id": "TEMPLATE_ID",
  "scope": "group",
  "group_id": "GPU_GROUP_ID",
  "config": {
    "mount_points": ["/mnt/models", "/mnt/scratch"]
  },
  "priority": 100,
  "is_active": true
}
```

### 3. 노드에서 실행

```bash
# dynamic-check.sh가 주기적으로 실행
/opt/aami/scripts/dynamic-check.sh

# 내부 동작:
# 1. Config Server에서 이 노드의 EffectiveCheck 조회
# 2. 스크립트를 로컬에 캐시 (hash로 변경 감지)
# 3. CONFIG 환경변수와 함께 스크립트 실행
# 4. 출력을 /var/lib/node_exporter/textfile/*.prom에 저장
```

### 4. Prometheus 수집 및 Alert 평가

```
Node Exporter가 textfile/*.prom 읽음
  ↓
Prometheus가 scrape (15초마다)
  ↓
AlertRule 평가: mount_check == 0
  ↓
조건 충족 시 Alertmanager로 전송
```

---

## API 레퍼런스

### ScriptTemplate API

#### Create Template
```http
POST /api/v1/script-templates
Content-Type: application/json

{
  "name": "check-mount-points",
  "script_type": "mount",
  "script_content": "#!/bin/bash\n...",
  "language": "bash",
  "version": "1.0.0",
  "default_config": {
    "mount_points": ["/data"]
  },
  "description": "마운트 포인트 접근성 체크"
}
```

#### Get Template
```http
GET /api/v1/script-templates/:id
```

#### List Templates
```http
GET /api/v1/script-templates?page=1&limit=20
```

#### Update Template
```http
PUT /api/v1/script-templates/:id
Content-Type: application/json

{
  "script_content": "#!/bin/bash\n# Updated...",
  "version": "2.0.0"
}
```

#### Delete Template
```http
POST /api/v1/script-templates/delete
Content-Type: application/json

{"id": "template-id"}
```

---

### ScriptPolicy API

#### Create Policy (from Template)
```http
POST /api/v1/script-policies
Content-Type: application/json

{
  "template_id": "TEMPLATE_ID",
  "scope": "group",
  "group_id": "GROUP_ID",
  "config": {
    "mount_points": ["/mnt/models"]
  },
  "priority": 100,
  "is_active": true
}
```

#### Create Policy (Direct, without Template)
```http
POST /api/v1/script-policies/direct
Content-Type: application/json

{
  "name": "custom-check",
  "script_type": "custom",
  "script_content": "#!/bin/bash\n...",
  "language": "bash",
  "version": "1.0.0",
  "scope": "group",
  "group_id": "GROUP_ID",
  "config": {},
  "is_active": true
}
```

#### Get Policy
```http
GET /api/v1/script-policies/:id
```

#### List Policies by Scope
```http
GET /api/v1/script-policies/global
GET /api/v1/script-policies/group/:group_id
```

#### Update Policy
```http
PUT /api/v1/script-policies/:id
Content-Type: application/json

{
  "config": {
    "mount_points": ["/mnt/models", "/mnt/scratch"]
  },
  "is_active": true
}
```

---

### Node API (노드 전용)

#### Get Effective Checks
노드가 실행해야 할 모든 체크를 조회:

```http
GET /api/v1/checks/target/:targetId

Response:
[
  {
    "name": "check-mount-points",
    "script_type": "mount",
    "script_content": "#!/bin/bash\n...",
    "language": "bash",
    "config": {
      "mount_points": ["/mnt/models", "/mnt/scratch"]
    },
    "version": "1.0.0",
    "hash": "abc123...",
    "instance_id": "policy-id-123"
  }
]
```

---

## 예제

### 예제 1: Mount Point 모니터링 (End-to-End)

이 예제는 마운트 포인트 모니터링의 전체 흐름을 보여줍니다.

#### Step 1: ScriptTemplate 생성

```bash
curl -X POST http://localhost:8080/api/v1/script-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "check-mount-points",
    "script_type": "mount",
    "language": "bash",
    "version": "1.0.0",
    "description": "마운트 포인트 접근성 및 쓰기 가능 여부 체크",
    "script_content": "#!/usr/bin/env bash\nset -euo pipefail\n\nmount_points=$(echo \"$CONFIG\" | jq -r '\''.mount_points[]'\'' 2>/dev/null || echo \"\")\n\necho \"# HELP mount_check Mount point accessibility (1=ok, 0=fail)\"\necho \"# TYPE mount_check gauge\"\n\nfor mp in $mount_points; do\n    if mountpoint -q \"$mp\" 2>/dev/null && [ -w \"$mp\" ]; then\n        echo \"mount_check{path=\\\"$mp\\\"} 1\"\n    else\n        echo \"mount_check{path=\\\"$mp\\\"} 0\"\n    fi\ndone",
    "default_config": {
      "mount_points": ["/data"]
    }
  }'
```

#### Step 2: 그룹별 ScriptPolicy 생성

```bash
# GPU 클러스터 (모델/스크래치 스토리지)
curl -X POST http://localhost:8080/api/v1/script-policies \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "TEMPLATE_ID",
    "scope": "group",
    "group_id": "GPU_CLUSTER_GROUP_ID",
    "config": {
      "mount_points": ["/mnt/models", "/mnt/scratch", "/mnt/datasets"]
    },
    "priority": 100,
    "is_active": true
  }'

# Storage 클러스터 (NFS/백업 스토리지)
curl -X POST http://localhost:8080/api/v1/script-policies \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "TEMPLATE_ID",
    "scope": "group",
    "group_id": "STORAGE_CLUSTER_GROUP_ID",
    "config": {
      "mount_points": ["/mnt/nfs", "/mnt/backup", "/mnt/archive"]
    },
    "priority": 100,
    "is_active": true
  }'
```

#### Step 3: AlertTemplate 생성

```bash
curl -X POST http://localhost:8080/api/v1/alert-templates \
  -H "Content-Type: application/json" \
  -d '{
    "id": "mount-point-unavailable",
    "name": "MountPointUnavailable",
    "description": "마운트 포인트 접근 불가 알림",
    "severity": "critical",
    "query_template": "mount_check{path=\"{{.mount_path}}\", group_id=\"{{.group_id}}\"} == 0",
    "default_config": {
      "for_duration": "5m",
      "labels": {
        "team": "infra",
        "component": "storage"
      },
      "annotations": {
        "summary": "마운트 포인트 {{ $labels.path }} 접근 불가",
        "description": "{{ $labels.instance }}에서 {{ $labels.path }}가 5분 이상 접근 불가 상태입니다."
      }
    }
  }'
```

#### Step 4: 그룹별 AlertRule 생성

```bash
# GPU 클러스터용 (더 엄격: 2분)
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "GPU_CLUSTER_GROUP_ID",
    "template_id": "mount-point-unavailable",
    "enabled": true,
    "config": {
      "for_duration": "2m",
      "labels": {
        "severity": "critical",
        "escalation": "immediate"
      }
    }
  }'

# Storage 클러스터용 (덜 엄격: 10분)
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "STORAGE_CLUSTER_GROUP_ID",
    "template_id": "mount-point-unavailable",
    "enabled": true,
    "config": {
      "for_duration": "10m",
      "labels": {
        "severity": "warning"
      }
    }
  }'
```

#### Step 5: 노드에서 스크립트 실행 결과

**GPU 클러스터 노드의 출력:**
```
# HELP mount_check Mount point accessibility (1=ok, 0=fail)
# TYPE mount_check gauge
mount_check{path="/mnt/models"} 1
mount_check{path="/mnt/scratch"} 1
mount_check{path="/mnt/datasets"} 0    # 문제 발생!
```

#### Step 6: 생성되는 Prometheus Alert Rule

```yaml
groups:
  - name: group_gpu-cluster_GPU_CLUSTER_GROUP_ID
    rules:
      - alert: MountPointUnavailable_gpu-cluster
        expr: mount_check{group_id="GPU_CLUSTER_GROUP_ID"} == 0
        for: 2m
        labels:
          severity: critical
          group_id: GPU_CLUSTER_GROUP_ID
          team: infra
          component: storage
          escalation: immediate
        annotations:
          summary: "마운트 포인트 {{ $labels.path }} 접근 불가"
          description: "{{ $labels.instance }}에서 {{ $labels.path }}가 2분 이상 접근 불가 상태입니다."
```

#### Step 7: Alert 발생 흐름

```
1. /mnt/datasets 마운트 실패
   ↓
2. mount_check{path="/mnt/datasets"} 0 메트릭 출력
   ↓
3. Node Exporter가 textfile에서 읽어 노출
   ↓
4. Prometheus가 scrape
   ↓
5. AlertRule 조건 충족: mount_check == 0
   ↓
6. 2분간 지속 (for: 2m)
   ↓
7. Alert firing → Alertmanager
   ↓
8. severity: critical → oncall 팀에게 알림
```

---

### 예제 2: 그룹별 다른 디스크 임계값

```bash
# ScriptTemplate: 디스크 사용량 체크
curl -X POST http://localhost:8080/api/v1/script-templates \
  -d '{
    "name": "check-disk-usage",
    "script_type": "disk",
    "language": "bash",
    "version": "1.0.0",
    "script_content": "#!/bin/bash\necho \"# HELP disk_usage_percent Disk usage percentage\"\necho \"# TYPE disk_usage_percent gauge\"\ndf -h / | tail -1 | awk '\''{gsub(/%/,\"\"); print \"disk_usage_percent{mountpoint=\\\"/\\\"} \" $5}'\''",
    "default_config": {}
  }'

# Critical Services: 70% 임계값
curl -X POST http://localhost:8080/api/v1/script-policies \
  -d '{
    "template_id": "TEMPLATE_ID",
    "scope": "group",
    "group_id": "CRITICAL_SERVICES_GROUP",
    "config": {},
    "is_active": true
  }'

# AlertRule: Critical Services (70%)
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -d '{
    "group_id": "CRITICAL_SERVICES_GROUP",
    "name": "HighDiskUsage",
    "severity": "critical",
    "query_template": "disk_usage_percent{group_id=\"{{.group_id}}\"} > 70",
    "config": { "for_duration": "5m" }
  }'

# AlertRule: Development (90%)
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -d '{
    "group_id": "DEVELOPMENT_GROUP",
    "name": "HighDiskUsage",
    "severity": "warning",
    "query_template": "disk_usage_percent{group_id=\"{{.group_id}}\"} > 90",
    "config": { "for_duration": "15m" }
  }'
```

**결과:**
| 그룹 | 디스크 임계값 | Alert 조건 | Severity |
|------|-------------|-----------|----------|
| Critical Services | 70% | 5분 이상 | critical |
| Development | 90% | 15분 이상 | warning |

---

## 모범 사례

### ScriptTemplate 설계
1. **재사용성**: 파라미터화 가능한 범용 스크립트 작성
2. **버전 관리**: 변경 시 버전 업데이트
3. **에러 처리**: 실패 시에도 메트릭 출력 (value=0)
4. **문서화**: Description 필드 활용

### ScriptPolicy 관리
1. **Global fallback**: 기본값은 Global로 설정
2. **Group override**: 특별한 요구사항만 Group 레벨로
3. **Override 최소화**: 필요한 파라미터만 Override
4. **비활성화**: 삭제 대신 is_active=false 사용

### Alert 설계
1. **그룹별 임계값**: 중요도에 따라 다른 임계값
2. **for_duration**: 일시적 문제 무시
3. **Labels 활용**: 라우팅/필터링용 label 추가
4. **Runbook 링크**: annotations에 문제 해결 가이드 링크

---

## 문제 해결

### ScriptPolicy가 노드에 반영 안 됨
```bash
# 노드에서 EffectiveCheck 확인
curl "http://config-server:8080/api/v1/checks/target/TARGET_ID"

# 로컬 캐시 제거 후 재실행
rm -f /opt/aami/cache/check-*.sh
/opt/aami/scripts/dynamic-check.sh
```

### 스크립트 변경이 적용 안 됨
```bash
# Hash 확인 (변경되었는지)
curl "http://config-server:8080/api/v1/script-templates/TEMPLATE_ID" | jq '.hash'

# ScriptPolicy는 생성 시 Template을 복사함
# Template 변경 후 새 Policy 생성 필요
```

### Alert가 발생하지 않음
```bash
# Prometheus에서 메트릭 확인
curl "http://prometheus:9090/api/v1/query?query=mount_check"

# Alert 상태 확인
curl "http://prometheus:9090/api/v1/alerts"

# Alertmanager 확인
curl "http://alertmanager:9093/api/v2/alerts"
```

---

## 참고 자료

- [빠른 시작 가이드](./QUICKSTART.md)
- [알림 시스템 아키텍처](./ALERTING-SYSTEM.md)
- [노드 등록](./NODE-REGISTRATION.md)
- [API 문서](./API.md)
