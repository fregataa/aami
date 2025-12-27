# 빠른 시작 가이드

이 가이드는 AAMI를 처음부터 설치하고 첫 모니터링 대상을 등록하는 과정을 안내합니다.

## 목차

1. [사전 요구사항](#사전-요구사항)
2. [설치](#설치)
3. [초기 구성](#초기-구성)
4. [첫 번째 그룹 생성하기](#첫-번째-그룹-생성하기)
5. [대상 등록하기](#대상-등록하기)
6. [알림 설정하기](#알림-설정하기)
7. [메트릭 확인하기](#메트릭-확인하기)
8. [다음 단계](#다음-단계)

## 사전 요구사항

시작하기 전에 다음 사항을 확인하세요:

- Docker 20.10+ 및 Docker Compose v2.0+
- 최소 4GB RAM 및 20GB 디스크 공간
- 대상 노드에 대한 네트워크 액세스 (모니터링용)
- Prometheus 및 Grafana에 대한 기본 이해

## 설치

### 1단계: 저장소 클론

```bash
git clone https://github.com/your-org/aami.git
cd aami
```

### 2단계: 환경 구성

```bash
cd deploy/docker-compose
cp .env.example .env
```

`.env` 파일을 원하는 설정으로 편집하세요:

```env
# PostgreSQL 구성
POSTGRES_USER=admin
POSTGRES_PASSWORD=changeme
POSTGRES_DB=config_server

# Redis 구성
REDIS_PASSWORD=

# Config Server 구성
CONFIG_SERVER_PORT=8080

# Grafana 구성
GRAFANA_ADMIN_PASSWORD=admin
```

### 3단계: 스택 시작

```bash
docker-compose up -d
```

모든 서비스가 시작될 때까지 대기하세요 (1-2분 소요):

```bash
docker-compose ps
```

다음 서비스들이 모두 "Up" 상태여야 합니다:
- prometheus
- grafana
- alertmanager
- config-server
- postgres
- redis

### 4단계: 설치 확인

모든 서비스에 접근 가능한지 확인하세요:

```bash
# Config Server 상태 확인
curl http://localhost:8080/api/v1/health

# 예상 출력:
# {"status":"ok","timestamp":"2024-01-01T00:00:00Z"}

# Prometheus
curl http://localhost:9090/-/healthy

# Grafana (HTML 반환 예상)
curl -I http://localhost:3000
```

## 초기 구성

### 1단계: Grafana 접속

1. 브라우저에서 http://localhost:3000 접속
2. 기본 자격증명으로 로그인:
   - 사용자명: `admin`
   - 비밀번호: `admin` (또는 .env에 설정한 비밀번호)
3. 비밀번호 변경 메시지가 표시됩니다 (개발 환경에서는 선택사항)

### 2단계: Prometheus 데이터 소스 확인

1. **Configuration** → **Data Sources**로 이동
2. Prometheus 데이터 소스가 이미 구성되어 있어야 합니다
3. **Test**를 클릭하여 연결을 확인하세요

## 첫 번째 그룹 생성하기

그룹은 인프라를 계층적으로 구성합니다. 기본 구조를 만들어 봅시다.

### 1단계: 인프라 그룹 생성

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

다음 단계를 위해 반환된 `group_id`를 저장하세요.

### 2단계: 하위 그룹 생성

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpu-cluster",
    "namespace": "infrastructure",
    "parent_id": "PARENT_GROUP_ID",
    "description": "GPU 컴퓨팅 클러스터"
  }'
```

### 3단계: 그룹 확인

```bash
curl http://localhost:8080/api/v1/groups
```

## 대상 등록하기

이제 모니터링 대상(모니터링할 서버)을 등록해 봅시다.

### 방법 1: API를 통한 수동 등록

```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "gpu-node-01.example.com",
    "ip_address": "10.0.1.10",
    "primary_group_id": "GROUP_ID_HERE",
    "exporters": [
      {
        "type": "node_exporter",
        "port": 9100,
        "enabled": true,
        "scrape_interval": "15s",
        "scrape_timeout": "10s"
      },
      {
        "type": "dcgm_exporter",
        "port": 9400,
        "enabled": true,
        "scrape_interval": "30s"
      }
    ],
    "labels": {
      "datacenter": "dc1",
      "rack": "r1",
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

### 방법 2: 부트스트랩 스크립트 (권장)

대상 노드에서 다음을 실행하세요:

```bash
curl -fsSL https://your-config-server:8080/bootstrap.sh | \
  bash -s -- \
    --token YOUR_BOOTSTRAP_TOKEN \
    --server https://your-config-server:8080
```

이 스크립트는 다음 작업을 수행합니다:
1. 하드웨어 자동 감지 (CPU, GPU, 메모리)
2. 적절한 exporter 설치 (node_exporter, dcgm_exporter)
3. Config Server에 자체 등록
4. 메트릭 내보내기 시작

### 3단계: 대상 등록 확인

```bash
# 모든 대상 목록 조회
curl http://localhost:8080/api/v1/targets

# 특정 대상 확인
curl http://localhost:8080/api/v1/targets/TARGET_ID
```

### 4단계: Prometheus 서비스 디스커버리 확인

```bash
# Prometheus 대상 확인
curl http://localhost:9090/api/v1/targets

# 또는 브라우저에서 확인
open http://localhost:9090/targets
```

등록된 대상이 30초 이내에 나타나야 합니다.

## 알림 설정하기

### 1단계: 사용 가능한 알림 규칙 템플릿 목록 조회

```bash
curl http://localhost:8080/api/v1/alert-templates
```

일반적인 템플릿:
- `HighCPUUsage` - 높은 CPU 사용률
- `HighMemoryUsage` - 높은 메모리 사용률
- `DiskSpaceLow` - 디스크 공간 부족
- `NodeDown` - 노드 다운
- `GPUHighTemperature` - GPU 고온

### 2단계: 그룹에 알림 규칙 적용

```bash
curl -X POST http://localhost:8080/api/v1/groups/GROUP_ID/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "rule_template_id": "HighCPUUsage",
    "enabled": true,
    "config": {
      "threshold": 80,
      "duration": "5m"
    },
    "merge_strategy": "override"
  }'
```

### 3단계: 알림 규칙 확인

```bash
# 대상에 대한 유효한 규칙 확인
curl http://localhost:8080/api/v1/targets/TARGET_ID/alert-rules/effective

# 정책 상속 추적
curl http://localhost:8080/api/v1/targets/TARGET_ID/alert-rules/trace
```

### 4단계: Alertmanager 구성

`config/alertmanager/alertmanager.yml` 편집:

```yaml
route:
  receiver: 'default-receiver'
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

receivers:
  - name: 'default-receiver'
    email_configs:
      - to: 'alerts@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alerts@example.com'
        auth_password: 'your-password'
```

Alertmanager 재시작:

```bash
docker-compose restart alertmanager
```

## 메트릭 확인하기

### 1단계: Grafana 대시보드 접속

1. http://localhost:3000 접속
2. **Dashboards** → **Browse**로 이동
3. `config/grafana/dashboards/`에서 미리 구축된 대시보드 가져오기

### 2단계: Prometheus 메트릭 탐색

http://localhost:9090을 방문하여 다음 쿼리를 시도해보세요:

**노드 메트릭:**
```promql
# CPU 사용률
100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# 메모리 사용률
100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))

# 디스크 사용률
100 - ((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes)
```

**GPU 메트릭 (DCGM):**
```promql
# GPU 사용률
DCGM_FI_DEV_GPU_UTIL

# GPU 온도
DCGM_FI_DEV_GPU_TEMP

# GPU 메모리 사용률
DCGM_FI_DEV_FB_USED / DCGM_FI_DEV_FB_TOTAL * 100
```

### 3단계: 커스텀 대시보드 생성

1. Grafana에서 **+ → Dashboard** 클릭
2. 새 패널 추가
3. Prometheus를 데이터 소스로 선택
4. PromQL 쿼리 입력
5. 시각화 구성
6. 대시보드 저장

## 다음 단계

축하합니다! 이제 AAMI 설치가 완료되었습니다. 다음에 수행할 작업입니다:

### 모니터링 확장

1. **더 많은 대상 추가**: 추가 노드 등록
2. **그룹 계층 구조 생성**: 데이터센터, 환경 또는 기능별로 구성
3. **알림 커스터마이징**: 그룹별 임계값 미세 조정
4. **커스텀 Exporter 배포**: 특수 하드웨어 모니터링

### 고급 구성

- [API 문서](./API.md) - 전체 REST API 레퍼런스
- [배포 가이드](../../deploy/README.md) - 프로덕션 배포
- [알림 규칙 가이드](./ALERT_RULES.md) - 고급 알림 구성
- [대시보드 가이드](./DASHBOARDS.md) - 커스텀 대시보드 생성

### 자동화

- [부트스트랩 스크립트](../../scripts/node/README.md) - 자동화된 에이전트 배포
- [Terraform 예제](../../examples/terraform/) - Infrastructure as Code
- [Ansible Playbook](../../deploy/ansible/) - 구성 관리

### 문제 해결

문제가 발생하면:

1. 로그 확인: `docker-compose logs -f SERVICE_NAME`
2. 연결 확인: `docker-compose ps`
3. Config Server 확인: `curl http://localhost:8080/api/v1/health`
4. [문제 해결 가이드](./TROUBLESHOOTING.md) 참조

## 일반적인 문제

### Prometheus에 대상이 나타나지 않음

**문제**: API를 통해 등록한 대상이 Prometheus에 표시되지 않음

**해결 방법**:
```bash
# 서비스 디스커버리 파일 확인
curl http://localhost:8080/api/v1/sd/prometheus

# Prometheus 재시작하여 구성 다시 로드
docker-compose restart prometheus
```

### Exporter가 응답하지 않음

**문제**: 대상 노드의 node_exporter에 연결할 수 없음

**해결 방법**:
```bash
# 대상 노드에서 exporter 실행 상태 확인
systemctl status node_exporter

# 방화벽 확인
sudo ufw status
sudo ufw allow 9100/tcp

# 대상에서 로컬 테스트
curl http://localhost:9100/metrics
```

### 알림 규칙이 작동하지 않음

**문제**: 조건이 충족되었는데도 알림이 발생하지 않음

**해결 방법**:
```bash
# Prometheus 규칙 확인
curl http://localhost:9090/api/v1/rules

# 규칙 평가 확인
# http://localhost:9090/alerts 열기

# Alertmanager 확인
curl http://localhost:9093/api/v2/alerts
```

## 정리

전체 스택을 제거하려면:

```bash
cd deploy/docker-compose

# 모든 서비스 중지
docker-compose down

# 모든 데이터 제거 (데이터베이스 포함)
docker-compose down -v
```

---

**도움이 필요하신가요?**
- [GitHub Issues](https://github.com/your-org/aami/issues)
- [문서](../../README.md)
- [커뮤니티 토론](https://github.com/your-org/aami/discussions)
