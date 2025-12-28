# 체크 관리 시스템

## 목차

1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [CheckTemplate vs CheckInstance](#checktemplate-vs-checkinstance)
4. [Scope 기반 관리](#scope-기반-관리)
5. [스크립트 출력 형식](#스크립트-출력-형식)
6. [워크플로우](#워크플로우)
7. [API 레퍼런스](#api-레퍼런스)
8. [예제](#예제)

## 개요

AAMI의 동적 체크 시스템은 인프라 전반에 걸쳐 커스텀 모니터링 체크를 중앙에서 관리하고 배포할 수 있게 해줍니다.

### 핵심 개념

- **CheckTemplate**: 재사용 가능한 체크 정의 (스크립트 코드 + 기본 파라미터)
- **CheckInstance**: 그룹별 체크 적용 (Template 참조 + Override 파라미터)
- **Scope 기반 관리**: Global → Namespace → Group 계층으로 체크 적용
- **로컬 캐싱**: 노드는 스크립트를 로컬 파일로 캐시
- **자동 업데이트**: 해시 기반 버전 감지로 스크립트 자동 갱신
- **JSON 출력**: 스크립트는 JSON으로 출력, 시스템이 Prometheus 형식으로 변환

---

## 아키텍처

### Alert 시스템과의 일관성

```
Alert 시스템:
├─ AlertTemplate (재사용 가능한 알림 규칙 정의)
└─ AlertRule (그룹별 알림 규칙 적용, Template 참조)

Check 시스템 (동일 패턴):
├─ CheckTemplate (재사용 가능한 체크 정의)
└─ CheckInstance (그룹별 체크 적용, Template 참조)
```

### 데이터 흐름

```
┌──────────────────┐
│ CheckTemplate    │  관리자가 정의
│ (재사용 가능)     │  - 스크립트 코드
└────────┬─────────┘  - 기본 파라미터
         │
         │ 참조
         ↓
┌──────────────────┐
│ CheckInstance    │  그룹별 적용
│ (그룹별 커스터마이징) │  - Template 참조
└────────┬─────────┘  - Override 파라미터
         │
         │ 노드가 조회
         ↓
┌──────────────────┐
│ Node             │  스크립트 실행
│ (dynamic-check)  │  - 로컬 캐시
└──────────────────┘  - 주기적 실행
```

---

## CheckTemplate vs CheckInstance

### CheckTemplate (체크 정의)

**목적**: 재사용 가능한 체크 스크립트 정의

**구조**:
```go
type CheckTemplate struct {
    ID            string
    Name          string                  // "disk-usage-check"
    CheckType     string                  // "disk"
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
- 한 번 정의, 여러 곳에서 재사용
- 버전 관리 및 히스토리
- 스크립트 로직 중앙 관리

**예시**:
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

### CheckInstance (체크 적용)

**목적**: 특정 Scope(Global/Namespace/Group)에 Template 적용

**구조**:
```go
type CheckInstance struct {
    ID          string
    TemplateID  string                  // CheckTemplate 참조
    Scope       string                  // "global", "namespace", "group"
    NamespaceID *string                 // Namespace 레벨일 때
    GroupID     *string                 // Group 레벨일 때
    Config      map[string]interface{}  // Override 파라미터
    Priority    int                     // 우선순위 (낮을수록 우선)
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**특징**:
- Template을 참조하여 적용
- 파라미터를 Override 가능
- Scope별 우선순위 해석

**예시**:
```json
// ML Training Group의 Disk Check
{
  "template_id": "disk-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "threshold": 70,  // Override: Template의 90 → 70
    "for": "3m"       // Override: Template의 5m → 3m
  },
  "priority": 100
}

// API Server Group의 Disk Check
{
  "template_id": "disk-check-template-id",
  "scope": "group",
  "group_id": "api-server-group",
  "config": {
    "threshold": 95   // Override: 더 여유롭게
  },
  "priority": 100
}
```

---

## Scope 기반 관리

### Scope 우선순위

CheckInstance는 3가지 Scope를 지원하며, 더 구체적인 Scope가 우선합니다:

```
Group (가장 구체적, 최우선)
  ↑
Namespace (중간)
  ↑
Global (가장 일반적, 최후순위)
```

### Scope 해석 로직

노드가 체크 스크립트를 조회할 때:

1. **Target의 PrimaryGroup 확인**
2. **Group 레벨 CheckInstance 조회**
   - 존재하면 사용
3. **Namespace 레벨 CheckInstance 조회** (Group에 없으면)
   - 존재하면 사용
4. **Global 레벨 CheckInstance 조회** (Namespace에 없으면)
   - 존재하면 사용
5. **없으면 에러**

### 예시: Disk Check 적용

```
Global CheckInstance:
  template: disk-check-template
  config: { threshold: 90 }

Namespace "production" CheckInstance:
  template: disk-check-template
  config: { threshold: 80 }  # 더 엄격

Group "ml-training" CheckInstance:
  template: disk-check-template
  config: { threshold: 70 }  # 가장 엄격

결과:
├─ ml-training 그룹 노드 → threshold: 70 (Group)
├─ production의 다른 그룹 노드 → threshold: 80 (Namespace)
└─ development 그룹 노드 → threshold: 90 (Global)
```

---

## 스크립트 출력 형식

### JSON 출력 (권장)

체크 스크립트는 간단한 JSON 형식으로 출력합니다:

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

### Helper Library 사용

#### Bash Helper (`/opt/aami/lib/prom-helper.sh`)

```bash
#!/bin/bash
source /opt/aami/lib/prom-helper.sh

# JSON 출력
echo '{"metrics": [{"name": "my_metric", "value": 42}]}'
```

#### Python Helper (`/opt/aami/lib/prom_helper.py`)

```python
from prom_helper import output_metrics

output_metrics([
    {"name": "my_metric", "value": 42, "labels": {"type": "example"}}
])
```

### 변환 프로세스

```
Check Script → JSON → dynamic-check.sh → Prometheus Format → Node Exporter
```

---

## 워크플로우

### 1. Template 생성 (관리자)

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

### 2. Instance 생성 (관리자)

```bash
# ML Training Group에 적용
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

### 3. 노드에서 실행

```bash
# dynamic-check.sh가 주기적으로 실행
/opt/aami/scripts/dynamic-check.sh

# 내부 동작:
# 1. Config Server에서 이 노드의 CheckInstance 조회
# 2. Template 스크립트 + Instance 파라미터로 실행
# 3. JSON 출력을 Prometheus 형식으로 변환
# 4. /var/lib/node_exporter/textfile/*.prom 저장
```

### 4. Prometheus 수집

```
Node Exporter가 textfile/*.prom 읽음
  ↓
Prometheus가 scrape
  ↓
메트릭 저장 및 AlertRule 평가
```

---

## API 레퍼런스

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

### Node API (노드 전용)

#### Get Effective Checks
노드가 실행해야 할 모든 체크를 조회:

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
스크립트 업데이트 확인 (hash 비교):

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

## 예제

### 예제 1: Mount Point 체크

#### 1. Template 생성
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

#### 2. Instance 생성 (ML Training Group)
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

#### 3. 노드에서 실행
```bash
# dynamic-check.sh 내부
hostname=$(hostname)
checks=$(curl "http://config-server:8080/api/v1/checks/node?hostname=$hostname")

for check in $(echo "$checks" | jq -c '.[]'); do
  check_type=$(echo "$check" | jq -r '.check_type')
  script=$(echo "$check" | jq -r '.template.script_content')
  config=$(echo "$check" | jq -r '.config')

  # 스크립트 저장 및 실행
  echo "$script" > "/tmp/check-$check_type.sh"
  chmod +x "/tmp/check-$check_type.sh"

  # 파라미터 추출 및 실행
  paths=$(echo "$config" | jq -r '.paths')
  result=$("/tmp/check-$check_type.sh" "$paths")

  # JSON → Prometheus 변환
  convert_to_prometheus "$result" > "/var/lib/node_exporter/textfile/check_$check_type.prom"
done
```

---

### 예제 2: Disk 사용량 체크 (Group별 다른 임계값)

#### 1. Template 생성 (한 번만)
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

#### 2. Instance 생성 (여러 그룹)
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

결과: 동일한 스크립트를 그룹별로 다른 임계값으로 실행!

---

## 모범 사례

### Template 설계
1. **재사용성**: 파라미터화 가능한 범용 스크립트 작성
2. **버전 관리**: 변경 시 버전 업데이트
3. **문서화**: Description 필드 활용
4. **테스트**: 배포 전 개발 환경에서 테스트

### Instance 관리
1. **Scope 최소화**: 가능한 Global/Namespace 사용, 예외만 Group
2. **Override 최소화**: 필요한 파라미터만 Override
3. **Priority 관리**: 충돌 시 우선순위 명확히
4. **비활성화**: 삭제 대신 IsActive=false 사용

### 노드 설정
1. **캐싱**: 네트워크 장애 대비 로컬 캐시 유지
2. **Auto-update**: Hash 비교로 자동 업데이트
3. **에러 처리**: 스크립트 실패 시 이전 결과 유지

---

## 문제 해결

### Template 업데이트가 노드에 반영 안 됨
```bash
# 노드에서 hash 확인
curl "http://config-server:8080/api/v1/checks/node/hash?hostname=$(hostname)&check_type=disk"

# 로컬 캐시 제거 후 재실행
rm -f /opt/aami/cache/check-*.sh
/opt/aami/scripts/dynamic-check.sh
```

### Instance 우선순위 충돌
```bash
# Instance 조회
curl "http://config-server:8080/api/v1/check-instances/group/my-group"

# Priority 수정
curl -X PUT "http://config-server:8080/api/v1/check-instances/:id" \
  -d '{"priority": 50}'
```

---

## 참고 자료

- [Quick Start Guide](./QUICKSTART.md)
- [Node Registration](./NODE-REGISTRATION.md)
- [Alert Rules Guide](./ALERT-RULES.md)
- [API Documentation](./API.md)
