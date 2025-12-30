# 알림 시스템 아키텍처

## 목차

1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [구성 요소](#구성-요소)
4. [알림 유형](#알림-유형)
5. [데이터 흐름](#데이터-흐름)
6. [그룹 기반 커스터마이징](#그룹-기반-커스터마이징)
7. [Alert Rule 생성](#alert-rule-생성)
8. [통합 지점](#통합-지점)
9. [예시](#예시)
10. [FAQ](#faq)

---

## 개요

AAMI의 알림 시스템은 AI 가속기 인프라를 위한 포괄적인 모니터링 및 알림 기능을 제공합니다. 이 시스템은 Prometheus와 Alertmanager를 기반으로 구축되어 있으며, 표준 메트릭 기반 알림과 커스텀 체크 기반 알림을 모두 처리하는 통합된 알림 경로를 제공합니다.

### 주요 기능

- **통합 알림 경로**: 모든 알림이 Prometheus → Alertmanager 경로를 따름
- **그룹 기반 커스터마이징**: 그룹별로 다른 알림 임계값 설정
- **Label 기반 필터링**: 특정 인프라에 대한 정밀한 알림 타겟팅
- **동적 체크 시스템**: 커스텀 요구사항을 위한 스크립트 기반 모니터링
- **템플릿 기반 관리**: 재사용 가능한 AlertTemplate 및 ScriptTemplate
- **정책 상속**: 그룹 기반 스마트 설정 병합

### 설계 철학

AAMI는 여러 개의 독립적인 알림 시스템 대신 **단일하고 일관된 알림 파이프라인**을 유지합니다. 이 접근 방식은 다음을 제공합니다:

- 중앙 집중식 알림 관리
- 일관된 라우팅 및 그룹화 정책
- 통합된 알림 채널
- 더 쉬운 문제 해결 및 디버깅
- 예측 가능한 알림 동작

---

## 아키텍처

### 시스템 개요

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI 가속기 클러스터                             │
│              (GPU 서버, 스토리지, 네트워크)                       │
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
                 │  - 메트릭 수집      │
                 │  - Rule 평가        │
                 │  - TSDB 저장        │
                 └──────────┬──────────┘
                            │ Firing Alerts
                 ┌──────────▼──────────┐
                 │   Alertmanager      │
                 │  - Alert 라우팅     │
                 │  - 그룹화/억제      │
                 │  - 중복 제거        │
                 └──────────┬──────────┘
                            │ Notifications
         ┌──────────────────┼──────────────────┐
         │                  │                  │
    ┌────▼────┐      ┌─────▼──────┐     ┌────▼─────┐
    │  Email  │      │   Slack    │     │ Webhook  │
    └─────────┘      └────────────┘     └──────────┘
```

### 통합 알림 경로

**핵심 설계 결정**: 모든 알림은 소스에 관계없이 동일한 경로를 따릅니다:

```
소스 → 메트릭 → Prometheus → Alert Rules → Alertmanager → 알림
```

이것의 의미:
- ❌ 체크 스크립트에서 직접 이메일/Slack 전송 안 함
- ❌ 독립적인 알림 시스템 없음
- ✅ 모든 알림이 Prometheus/Alertmanager를 거침
- ✅ 일관된 라우팅 및 그룹화
- ✅ 중앙 집중식 설정

---

## 구성 요소

### Prometheus

**역할**: 메트릭 수집, 저장, alert rule 평가

**책임**:
- 15초마다 exporter에서 메트릭 수집 (설정 가능)
- TSDB에 시계열 데이터 저장
- 15초마다 alert rules 평가 (설정 가능)
- Firing된 alert를 Alertmanager로 전송
- PromQL 쿼리 인터페이스 제공

**설정**:
- `config/prometheus/prometheus.yml`: 메인 설정
- `config/prometheus/rules/*.yml`: Alert rules
- Config Server의 HTTP SD를 통한 Service Discovery

### Alertmanager

**역할**: Alert 관리 및 라우팅

**책임**:
- **라우팅**: Label 기반으로 적절한 수신자에게 alert 전달
- **그룹화**: 유사한 alert를 결합하여 알림 볼륨 감소
- **억제**: 상위 우선순위 alert 발생 시 하위 우선순위 alert 억제
- **중복 제거**: 중복 알림 방지
- **침묵**: 특정 alert를 일시적으로 음소거

**설정**: `config/alertmanager/alertmanager.yml`

**주요 기능**:
- 심각도 기반 라우팅 (critical, warning, info)
- 네임스페이스 기반 라우팅 (infrastructure, logical, environment)
- 시간 기반 그룹화 (group_wait, group_interval, repeat_interval)

### Alert Rules

**역할**: Alert를 트리거하는 조건 정의

**구조**:
```yaml
- alert: AlertName
  expr: PromQL 표현식
  for: 지속 시간
  labels:
    severity: critical
    group_id: grp-123
  annotations:
    summary: Alert 요약
    description: 상세 설명
```

**저장 위치**: `config/prometheus/rules/*.yml`

**현재 상태**:
- ✅ 정적 rule 파일 (수동 생성)
- ✅ 동적 생성 (AlertRule API → Prometheus YAML 자동 변환)

---

## 알림 유형

### 1. Prometheus 기반 Alert

**정의**: Exporter의 표준 Prometheus 메트릭에 의해 트리거되는 alert

**데이터 흐름**:
```
Exporter → Prometheus → Alert Rules → Alertmanager
```

**예시**:
- Node Exporter 메트릭: CPU, 메모리, 디스크, 네트워크
- DCGM Exporter 메트릭: GPU 사용률, 온도, 전력
- Custom Exporter 메트릭: 애플리케이션 특정 메트릭

**Rule 예시**:
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
    summary: "{{ $labels.instance }}에서 높은 CPU 사용률"
```

### 2. 커스텀 체크 시스템

**정의**: 표준 exporter로 커버되지 않는 인프라 구성 요소를 위한 스크립트 기반 모니터링

**데이터 흐름**:
```
Config Server (ScriptTemplate/ScriptPolicy)
  ↓
노드가 유효한 체크 조회
  ↓
dynamic-check.sh가 스크립트 실행
  ↓
Prometheus 텍스트 형식으로 출력
  ↓
/var/lib/node_exporter/textfile/*.prom에 저장
  ↓
Node Exporter textfile collector
  ↓
Prometheus가 수집
  ↓
Alert Rules 평가
  ↓
Alertmanager
```

**사용 사례**:
- 마운트 포인트 가용성
- 디바이스 연결 상태
- 네트워크 인터페이스 체크
- 커스텀 애플리케이션 헬스 체크
- 파일시스템 특정 모니터링

**핵심 구성 요소**:
- **ScriptTemplate**: 재사용 가능한 스크립트 정의 (services/config-server/internal/domain/script_template.go)
- **ScriptPolicy**: 그룹별 적용 (services/config-server/internal/domain/script_policy.go)
- **Scope 기반 관리**: Global → Group 계층

**예시**: 마운트 포인트 체크

```bash
# ScriptTemplate 스크립트 (Prometheus 텍스트 형식 직접 출력)
#!/bin/bash
PATHS="$1"
for path in ${PATHS//,/ }; do
  if mountpoint -q "$path"; then
    echo "mount_status{path=\"$path\"} 1"
  else
    echo "mount_status{path=\"$path\"} 0"
  fi
done
```

Textfile 출력 (동일):
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
    summary: "마운트 포인트 {{ $labels.path }} 사용 불가"
```

**중요**: 커스텀 체크도 직접 알림이 아닌 Prometheus/Alertmanager를 거칩니다.

---

## 데이터 흐름

### 표준 메트릭 경로

```
┌─────────────────┐
│ Node Exporter   │  Port 9100, metrics 엔드포인트
│ DCGM Exporter   │  Port 9400, metrics 엔드포인트
│ Custom Exporter │  Port 9xxx, metrics 엔드포인트
└────────┬────────┘
         │ HTTP GET /metrics (15초마다)
         ↓
┌────────────────────────────────┐
│ Prometheus                     │
│ - 메트릭 수집                  │
│ - TSDB에 저장                  │
│ - Rules 평가 (15초마다)        │
└────────┬───────────────────────┘
         │ Firing alerts
         ↓
┌────────────────────────────────┐
│ Alertmanager                   │
│ - 심각도/네임스페이스로 라우팅 │
│ - 유사 alert 그룹화            │
│ - 억제 규칙 적용               │
└────────┬───────────────────────┘
         │ 알림 전송
         ↓
┌────────────────────────────────┐
│ 알림 채널                      │
│ - Email (SMTP)                 │
│ - Slack (Webhook)              │
│ - PagerDuty (Webhook)          │
└────────────────────────────────┘
```

### 커스텀 체크 경로

```
┌─────────────────────────────────┐
│ Config Server                   │
│ - ScriptTemplate 저장           │
│ - ScriptPolicy 관리             │
│ - Scope 해석 (Global, Group)    │
└────────┬────────────────────────┘
         │ GET /api/v1/checks/target/{targetId}
         ↓
┌─────────────────────────────────┐
│ 노드: dynamic-check.sh          │
│ 1. 유효한 체크 조회             │
│ 2. 스크립트 실행                │
│ 3. Prometheus 텍스트 출력       │
└────────┬────────────────────────┘
         │ Prometheus 메트릭
         ↓
┌─────────────────────────────────┐
│ 파일에 직접 저장                │
│ mount_status{path="/data"} 1    │
└────────┬────────────────────────┘
         │
         ↓
┌─────────────────────────────────┐
│ /var/lib/node_exporter/         │
│   textfile/*.prom               │
└────────┬────────────────────────┘
         │ textfile collector가 읽음
         ↓
┌─────────────────────────────────┐
│ Node Exporter                   │
│ - 메트릭으로 노출               │
└────────┬────────────────────────┘
         │ 수집
         ↓
┌─────────────────────────────────┐
│ Prometheus                      │
│ (표준 메트릭과 동일한 경로)    │
└─────────────────────────────────┘
```

---

## 그룹 기반 커스터마이징

### 문제 정의

**질문**: Prometheus alert rule은 글로벌합니다. 어떻게 그룹별로 다른 alert 임계값을 지원할 수 있을까요?

**예시**:
- Production 그룹: CPU alert 80%에서
- Development 그룹: CPU alert 95%에서

### 해결책: Label 기반 필터링 + 동적 Rule 생성

#### 단계 1: Service Discovery에서 그룹 Label 추가

**코드**: `services/config-server/internal/domain/service_discovery.go`

```go
// 타겟 등록 시, 그룹 정보를 label로 추가
labels["group"] = target.Groups[0].Name           // "gpu-cluster-a"
labels["group_id"] = target.Groups[0].ID          // "grp-123"
```

**결과**: 이 타겟의 모든 메트릭에 그룹 label 포함

```promql
node_cpu_seconds_total{
  instance="gpu-node-01",
  group="gpu-cluster-a",
  group_id="grp-123"
}
```

#### 단계 2: 그룹별 Alert Rule 생성

각 그룹은 다음을 포함하는 자체 alert rule을 가집니다:
- 그룹별 PromQL 필터 (`group_id="grp-123"`)
- 그룹별 임계값 (80% vs 95%)
- 그룹별 지속 시간 (5m vs 10m)

**Production 그룹** (임계값: 80%):
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
              group_id="grp-123"  # 이 그룹으로 필터링
            }[5m])
          ) * 100)) > 80  # Production 임계값
        for: 5m
        labels:
          severity: warning
          group_id: grp-123
          namespace: production
```

**Development 그룹** (임계값: 95%):
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
              group_id="grp-456"  # 이 그룹으로 필터링
            }[5m])
          ) * 100)) > 95  # Development 임계값
        for: 10m
        labels:
          severity: info
          group_id: grp-456
          namespace: development
```

#### 단계 3: 동적 생성을 위한 AlertRule.RenderQuery()

**코드**: `services/config-server/internal/domain/alert.go:102-125`

```go
// 쿼리 템플릿을 포함한 AlertTemplate
QueryTemplate: `(100 - avg(rate(node_cpu_seconds_total{
  mode="idle",
  group_id="{{.group_id}}"
}[5m])) * 100) > {{.threshold}}`

// Production 그룹 Config
Config: {
  "group_id": "grp-123",
  "threshold": 80,
  "for_duration": "5m"
}

// 렌더링 결과:
"(100 - avg(rate(node_cpu_seconds_total{
  mode=\"idle\",
  group_id=\"grp-123\"
}[5m])) * 100) > 80"
```

### 장점

- ✅ 동일한 메트릭, 그룹별로 다른 임계값
- ✅ 그룹별로 깔끔한 rule 분리
- ✅ 디버깅 용이 (label에 group_id)
- ✅ 확장 가능 (자동 생성)
- ✅ 유연함 (템플릿 + config 접근)

### 타겟별 커스터마이징

더 세밀한 제어를 위해 타겟 label 사용:

```yaml
- alert: HighCPUUsage_GPU_Servers
  expr: |
    (100 - avg by(instance) (
      rate(node_cpu_seconds_total{
        mode="idle",
        group_id="grp-123",
        target_label_type="gpu"  # 타겟별 필터
      }[5m])
    ) * 100) > 70  # GPU 서버는 다른 임계값
```

---

## Alert Rule 생성

### 아키텍처

Alert Rule 생성 시스템은 다음 컴포넌트로 구성됩니다:

- **AlertTemplate API**: 재사용 가능한 alert 템플릿 관리
- **AlertRule API**: 그룹별 alert rule 설정
- **Prometheus Rule Generator**: AlertRule을 Prometheus rule 파일로 변환
- **Rule File Manager**: Atomic write, 검증, 백업 기능 제공
- **Prometheus Client**: Prometheus reload 및 헬스체크

### Prometheus Rule Management API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/prometheus/rules/regenerate` | 모든 Prometheus rule 파일 재생성 |
| POST | `/api/v1/prometheus/rules/regenerate/:group_id` | 특정 그룹의 rule 파일 재생성 |
| GET | `/api/v1/prometheus/rules/files` | 생성된 rule 파일 목록 조회 |
| GET | `/api/v1/prometheus/rules/effective/:target_id` | 특정 Target에 적용되는 effective rules 조회 |
| POST | `/api/v1/prometheus/reload` | Prometheus 설정 reload 트리거 |
| GET | `/api/v1/prometheus/status` | Prometheus 연결 상태 확인 |

### 구현 세부사항

**Rule Generator** (`prometheus_rule_generator.go`):
- `GenerateRulesForGroup()`: 특정 그룹의 AlertRule을 Prometheus YAML로 변환
- `GenerateAllRules()`: 모든 그룹의 rule 파일 일괄 생성
- `DeleteRulesForGroup()`: 그룹의 rule 파일 삭제

**File Manager** (`file_manager.go`):
- Atomic write (temp 파일 → rename)
- promtool 검증 지원
- 백업 및 롤백 기능

**Prometheus Client** (`client.go`):
- HTTP POST `/-/reload` 엔드포인트 호출
- 재시도 로직 (exponential backoff)
- 헬스체크 (`/-/ready`, `/-/healthy`)

### 환경 변수

```bash
PROMETHEUS_URL=http://localhost:9090
PROMETHEUS_RULE_PATH=/etc/prometheus/rules/generated
PROMETHEUS_RELOAD_ENABLED=true
PROMETHEUS_RELOAD_TIMEOUT=30s
PROMETHEUS_VALIDATE_RULES=false
PROMETHEUS_BACKUP_ENABLED=true
```

### 트리거 이벤트
- AlertRule 생성/수정/삭제 시 자동 재생성
- API를 통한 수동 재생성

---

## 통합 지점

### 1. Service Discovery → Labels

**파일**: `services/config-server/internal/domain/service_discovery.go`

타겟이 등록될 때 그룹 정보가 label로 추가됩니다:

```go
labels["group"] = target.Groups[0].Name
labels["group_id"] = target.Groups[0].ID
```

이 label들은 다음에서 사용됩니다:
- Alert rule 필터링 (`group_id="grp-123"`)
- Alertmanager 라우팅 (`group: production`)
- Grafana 대시보드 변수

### 2. Alert Rules → Alertmanager

**파일**: `config/prometheus/prometheus.yml:8-12`

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093
```

Prometheus는 모든 label이 보존된 상태로 firing된 alert를 Alertmanager로 전송합니다.

### 3. Alertmanager → 알림 채널

**파일**: `config/alertmanager/alertmanager.yml`

다음을 기반으로 alert 라우팅:
- **심각도**: critical, warning, info
- **그룹**: production, development, staging
- **커스텀 label**: team, service 등

라우팅 예시:
```yaml
routes:
  - match:
      severity: critical
    receiver: 'oncall-team'
    group_wait: 0s
    repeat_interval: 4h

  - match:
      group: production
    receiver: 'production-team'
    continue: true
```

### 4. ScriptPolicy → 노드 실행

**API 엔드포인트**: `GET /api/v1/checks/target/{targetId}`

노드는 Config Server를 조회하여 다음을 얻습니다:
- 유효한 ScriptPolicy (scope 해석 후)
- 스크립트 내용 및 hash
- 병합된 설정 (default_config + config)

응답:
```json
[
  {
    "name": "mount-check",
    "script_type": "mount",
    "script_content": "#!/bin/bash\n...",
    "language": "bash",
    "config": {
      "paths": "/data,/mnt/models"
    },
    "version": "1.0.0",
    "hash": "abc123..."
  }
]
```

---

## 예시

### 예시 1: 표준 메트릭 Alert (노드 다운)

**Rule 파일**: `config/prometheus/rules/system-alerts.yml`

```yaml
- alert: NodeDown
  expr: up{job="node-exporter"} == 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "노드 {{ $labels.instance }} 다운"
    description: |
      노드가 2분 이상 응답하지 않습니다.
      인스턴스: {{ $labels.instance }}
      그룹: {{ $labels.group }}
```

**흐름**:
1. Node Exporter 응답 중지
2. Prometheus가 `up{job="node-exporter"}`를 0으로 표시
3. Alert rule 조건이 2분간 충족
4. Prometheus가 Alertmanager로 alert 전송
5. Alertmanager가 'critical-alerts' 수신자로 라우팅 (이메일 + PagerDuty)

### 예시 2: 그룹별 디스크 Alert

**시나리오**: 환경별로 다른 디스크 임계값

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

**생성된 Prometheus Rules**:

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

### 예시 3: 커스텀 체크 (마운트 포인트)

**ScriptTemplate 생성**:
```bash
POST /api/v1/script-templates
{
  "name": "mount-check",
  "script_type": "mount",
  "script_content": "#!/bin/bash\nPATHS=\"$1\"\nfor path in ${PATHS//,/ }; do\n  if mountpoint -q \"$path\"; then\n    echo \"mount_status{path=\\\"$path\\\"} 1\"\n  else\n    echo \"mount_status{path=\\\"$path\\\"} 0\"\n  fi\ndone",
  "language": "bash",
  "default_config": {
    "paths": "/data"
  },
  "version": "1.0.0"
}
```

**ScriptPolicy (ML Training 그룹)**:
```bash
POST /api/v1/script-policies
{
  "template_id": "mount-check-template-id",
  "scope": "group",
  "group_id": "ml-training-group",
  "config": {
    "paths": "/data,/mnt/models,/mnt/datasets"
  },
  "is_active": true
}
```

**노드 실행**:
```bash
# dynamic-check.sh 주기적으로 실행
/opt/aami/scripts/dynamic-check.sh

# textfile에 출력:
# /var/lib/node_exporter/textfile/mount-check.prom
mount_status{path="/data"} 1
mount_status{path="/mnt/models"} 0  # 실패!
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
    summary: "{{ $labels.instance }}에서 마운트 포인트 {{ $labels.path }} 사용 불가"
```

**결과**: `/mnt/models` 마운트 실패 시, 2분 후 alert가 발생하고 알림이 전송됩니다.

---

## FAQ

### Q: Alert 시스템이 Alertmanager에 의존하나요?

**A**: 부분적으로 그렇습니다.

- **Alert 평가**: 의존하지 않음. Prometheus가 독립적으로 alert rule을 평가하고 내부 상태에서 alert를 "firing"으로 표시합니다.
- **Alert 알림**: 네, Alertmanager가 필요합니다. 없으면 alert가 Prometheus UI(`http://localhost:9090/alerts`)에서 보이지만 알림은 전송되지 않습니다.

**Alertmanager 있을 때**:
```
Prometheus → Rules 평가 → Alert 발생 → Alertmanager → Email/Slack
```

**Alertmanager 없을 때**:
```
Prometheus → Rules 평가 → Alert 발생 → [알림 없음]
                                      └→ Prometheus UI에서만 확인 가능
```

### Q: Prometheus alert rule이 글로벌한가요?

**A**: 네, 하지만 AAMI는 **label 기반 필터링**을 사용하여 그룹별 동작을 구현합니다.

- Prometheus rule 파일은 글로벌 (`config/prometheus/rules/*.yml`에서 로드)
- 각 rule은 label로 메트릭 필터링 가능 (`group_id="grp-123"`)
- AAMI는 각 그룹에 대해 다른 임계값으로 별도 rule 생성
- 결과: 그룹별로 보이지만, 여러 글로벌 rule로 구현됨

### Q: 커스텀 인프라 모니터링은 custom exporter를 통해 수행되나요?

**A**: 아니요, AAMI는 custom exporter가 아닌 **커스텀 체크 시스템**을 사용합니다.

**Custom Exporter** (전통적 접근):
- 별도의 Go/Python 프로세스
- HTTP 메트릭 엔드포인트 노출
- 바이너리 배포 필요
- 그룹별 커스터마이징 어려움

**AAMI 체크 시스템** (동적 접근):
- Shell/Python 스크립트
- Prometheus 텍스트 형식으로 직접 출력
- Node Exporter textfile collector
- ScriptPolicy를 통한 쉬운 그룹별 커스터마이징
- Config Server API를 통한 동적 배포

두 경로 모두 결국 Prometheus → Alertmanager를 거칩니다.

### Q: Alert가 Prometheus/Alertmanager를 우회하여 더 빠른 알림을 받을 수 있나요?

**A**: 아니요, 이것은 의도된 설계입니다.

**통합 경로를 사용하는 이유**:
- 일관된 alert 라우팅 및 그룹화
- Alert 상태에 대한 단일 정보원
- 더 쉬운 문제 해결 (한 곳에서 확인)
- Alertmanager 기능 (억제, 중복 제거, 침묵)
- 여러 소스에서 오는 alert 폭주 방지

**트레이드오프**:
- 약간의 지연 (scrape_interval + evaluation_interval + Alertmanager 처리)
- 일반적인 지연: 30-60초
- 인프라 모니터링에는 허용 가능
- 1초 미만 요구사항의 경우, 애플리케이션 코드에서 직접 모니터링 고려

### Q: 배포 전에 alert rule을 테스트하려면 어떻게 해야 하나요?

**A**: Prometheus UI와 promtool 사용:

```bash
# 문법 검증
promtool check rules config/prometheus/rules/system-alerts.yml

# Prometheus UI에서 쿼리 테스트
http://localhost:9090/graph

# PromQL 표현식 입력
(100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)) > 80

# 수동으로 alert 트리거 (임계값을 매우 낮게 설정)
# Alerts 페이지 확인
http://localhost:9090/alerts
```

### Q: 글로벌 및 그룹별 alert rule을 모두 가질 수 있나요?

**A**: 네, 이것은 일반적인 패턴입니다.

**글로벌 Rule** (모든 그룹의 기준):
```yaml
- alert: NodeDown
  expr: up{job="node-exporter"} == 0
  for: 5m  # 더 관대함
```

**그룹별 Rule** (production에 대해 더 엄격):
```yaml
- alert: NodeDown_Production
  expr: up{job="node-exporter",group="production"} == 0
  for: 1m  # production에 대해 더 빠른 alert
```

중복 alert를 방지하기 위해 억제 규칙을 사용하세요.

### Q: Alert rule 파일에 문법 오류가 있으면 어떻게 되나요?

**A**: Prometheus는:
1. 시작/reload 시 오류 로그
2. 잘못된 rule 파일 건너뜀
3. 유효한 rule 파일로 계속 진행
4. 유효한 rule에 대해 alert 평가 계속

배포 전에 항상 `promtool check rules`로 검증하세요.

### Q: 유지보수 중에 alert를 침묵시키려면 어떻게 해야 하나요?

**A**: Alertmanager silences 사용:

```bash
# UI를 통해
http://localhost:9093/#/silences

# API를 통해
curl -X POST http://localhost:9093/api/v2/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {"name": "instance", "value": "gpu-node-01", "isRegex": false}
    ],
    "startsAt": "2025-01-01T10:00:00Z",
    "endsAt": "2025-01-01T12:00:00Z",
    "comment": "예정된 유지보수",
    "createdBy": "admin@example.com"
  }'
```

Silence는 일시적이며 자동으로 만료됩니다.

---

## 참고 자료

- [체크 관리 시스템](./CHECK-MANAGEMENT.md) - 커스텀 체크 시스템 상세
- [빠른 시작 가이드](./QUICKSTART.md) - AAMI 시작하기
- [API 문서](./API.md) - Alert 및 check API 레퍼런스
- [Prometheus 문서](https://prometheus.io/docs/alerting/latest/overview/)
- [Alertmanager 문서](https://prometheus.io/docs/alerting/latest/alertmanager/)
