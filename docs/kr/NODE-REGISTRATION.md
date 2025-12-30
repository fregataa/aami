# 노드 등록 가이드

## 목차

1. [개요](#개요)
2. [등록 방법 비교](#등록-방법-비교)
3. [사전 준비사항](#사전-준비사항)
4. [온사이트 서버 등록](#온사이트-서버-등록)
5. [클라우드 서버 등록](#클라우드-서버-등록)
6. [등록 후 확인](#등록-후-확인)
7. [문제 해결](#문제-해결)

## 개요

AAMI는 온사이트(온프레미스) 서버와 클라우드 서버 모두를 모니터링 대상으로 등록할 수 있습니다. 두 환경의 등록 방식은 자동화 수준에서 차이가 있지만, 최종적으로 동일한 모니터링 기능을 제공합니다.

### 핵심 차이점

| 구분 | 온사이트 서버 | 클라우드 서버 |
|------|--------------|--------------|
| **등록 방식** | 수동 또는 반자동 | 완전 자동 |
| **초기 설정** | SSH 접속하여 스크립트 실행 | Cloud-init / User Data 활용 |
| **배포 속도** | 서버당 개별 작업 필요 | 대량 배포 간편 |
| **네트워크** | 기존 인프라 활용 | VPC/보안그룹 구성 필요 |
| **사용 사례** | GPU 클러스터, HPC, 기존 인프라 | 동적 확장, Auto Scaling |

## 등록 방법 비교

### 방법 1: Bootstrap Token 사용 (권장)

**특징**: 서버가 스스로 Config Server에 등록

**장점**:
- ✅ 완전 자동화 가능
- ✅ 하드웨어 정보 자동 감지
- ✅ 대량 배포에 적합
- ✅ 사람 실수 최소화

**적용 대상**:
- 클라우드 VM (AWS, GCP, Azure 등)
- 새로 프로비저닝되는 온사이트 서버
- 자동화된 배포 파이프라인

### 방법 2: API 직접 호출 (수동)

**특징**: 관리자가 수동으로 API 호출

**장점**:
- ✅ 세밀한 제어 가능
- ✅ 기존 서버 등록에 적합
- ✅ 특별한 설정 가능

**적용 대상**:
- 이미 운영 중인 온사이트 서버
- 특수한 설정이 필요한 서버
- 소규모 환경

## 사전 준비사항

### Config Server 설정

#### 1. 그룹 생성

노드를 등록하기 전에 그룹 구조를 먼저 설계합니다.

```bash
# 예시: 환경별 그룹
curl -X POST http://config-server:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "namespace": "environment",
    "description": "프로덕션 환경"
  }'

# 예시: 기능별 그룹
curl -X POST http://config-server:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml-training",
    "namespace": "logical",
    "description": "머신러닝 훈련용 GPU 클러스터"
  }'
```

#### 2. Bootstrap Token 생성 (자동 등록용)

```bash
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml-cluster-token",
    "default_group_id": "GROUP_ID",
    "max_uses": 100,
    "expires_at": "2024-12-31T23:59:59Z",
    "labels": {
      "environment": "production",
      "cluster": "ml-training"
    }
  }'

# 응답에서 token 값을 저장
# Response: {"token": "aami_bootstrap_xxxxx..."}
```

### 네트워크 요구사항

**Config Server 접근**:
- 포트 8080 (HTTP API)
- 노드 → Config Server 방향 통신 필요

**Prometheus 스크랩**:
- 포트 9100 (Node Exporter)
- 포트 9400 (DCGM Exporter, GPU 있는 경우)
- Prometheus → 노드 방향 통신 필요

### 사전 검증 (권장)

노드 등록 전에 시스템 요구사항을 검증할 수 있습니다:

```bash
# AAMI 저장소에서 스크립트 다운로드 후 실행
curl -fsSL https://raw.githubusercontent.com/fregataa/aami/main/scripts/preflight-check.sh -o preflight-check.sh
chmod +x preflight-check.sh

# 노드 모드로 검증 (Config Server 연결 테스트 포함)
./preflight-check.sh --mode node --server http://config-server:8080
```

이 스크립트는 다음을 검사합니다:
- 시스템 요구사항 (CPU, RAM, 디스크 공간)
- 소프트웨어 의존성 (curl, systemctl, tar)
- Config Server 연결 가능 여부
- 포트 가용성 (9100, 9400)
- GPU 감지 (NVIDIA, AMD)

## 온사이트 서버 등록

### 시나리오 1: Bootstrap 스크립트 사용 (반자동)

#### 단계 1: 서버 접속

```bash
ssh user@onsite-server-01
```

#### 단계 2: Bootstrap 스크립트 실행

```bash
# Bootstrap token 준비
BOOTSTRAP_TOKEN="aami_bootstrap_xxxxx..."
CONFIG_SERVER_URL="http://config-server.internal:8080"

# Bootstrap 실행
curl -fsSL ${CONFIG_SERVER_URL}/bootstrap.sh | \
  bash -s -- \
    --token ${BOOTSTRAP_TOKEN} \
    --server ${CONFIG_SERVER_URL}
```

#### 스크립트 동작 과정

1. **시스템 정보 수집**
   - 호스트명, IP 주소
   - CPU 코어 수, 메모리 용량
   - GPU 감지 (nvidia-smi)
   - 네트워크 인터페이스

2. **Exporter 설치**
   - Node Exporter (시스템 메트릭)
   - DCGM Exporter (GPU가 있는 경우)

3. **Config Server 등록**
   - Bootstrap token으로 인증
   - 수집한 정보 전송
   - 그룹 자동 할당 (token의 default_group_id)

4. **동적 체크 설정**
   - dynamic-check.sh 설치
   - Cron 작업 등록 (1분 간격)
   - 첫 체크 실행

#### 단계 3: 등록 확인

```bash
# Node Exporter 동작 확인
curl http://localhost:9100/metrics

# Config Server에서 등록 확인
curl http://config-server:8080/api/v1/targets?hostname=$(hostname)
```

### 시나리오 2: 수동 API 등록

기존에 이미 Node Exporter가 설치되어 있는 경우:

```bash
curl -X POST http://config-server:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "onsite-gpu-01",
    "ip_address": "192.168.1.100",
    "primary_group_id": "GROUP_ID",
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
      "datacenter": "seoul",
      "rack": "r1",
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

## 클라우드 서버 등록

### AWS EC2 예시

#### Terraform 코드

```hcl
# variables.tf
variable "config_server_url" {
  default = "http://config-server.internal:8080"
}

variable "bootstrap_token" {
  description = "AAMI Bootstrap Token"
  sensitive   = true
}

# main.tf
resource "aws_instance" "gpu_node" {
  ami           = "ami-xxxxx"  # Ubuntu 22.04 with GPU drivers
  instance_type = "p4d.24xlarge"

  vpc_security_group_ids = [aws_security_group.gpu_nodes.id]
  subnet_id              = aws_subnet.private.id

  user_data = templatefile("${path.module}/userdata.sh.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  })

  tags = {
    Name        = "ml-training-node-${count.index + 1}"
    Environment = "production"
    ManagedBy   = "terraform"
  }

  count = 10  # 10대의 GPU 노드 생성
}

# security_group.tf
resource "aws_security_group" "gpu_nodes" {
  name        = "aami-gpu-nodes"
  description = "AAMI monitored GPU nodes"

  # Node Exporter (Prometheus → Node)
  ingress {
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
    cidr_blocks = [var.prometheus_cidr]
  }

  # DCGM Exporter (Prometheus → Node)
  ingress {
    from_port   = 9400
    to_port     = 9400
    protocol    = "tcp"
    cidr_blocks = [var.prometheus_cidr]
  }

  # Outbound to Config Server
  egress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = [var.config_server_cidr]
  }

  # Outbound to internet (for package installation)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

#### User Data 스크립트

```bash
# userdata.sh.tpl
#!/bin/bash
set -e

# 로그 파일
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

echo "Starting AAMI bootstrap at $(date)"

# Config Server 설정
CONFIG_SERVER_URL="${config_server_url}"
BOOTSTRAP_TOKEN="${bootstrap_token}"

# 시스템 업데이트
apt-get update
apt-get install -y curl jq

# Bootstrap 실행
curl -fsSL $${CONFIG_SERVER_URL}/bootstrap.sh | \
  bash -s -- \
    --token $${BOOTSTRAP_TOKEN} \
    --server $${CONFIG_SERVER_URL}

echo "Bootstrap completed at $(date)"
```

### GCP Compute Engine 예시

```hcl
resource "google_compute_instance" "gpu_node" {
  name         = "ml-training-node-${count.index + 1}"
  machine_type = "a2-highgpu-8g"  # A100 x8
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 200
    }
  }

  network_interface {
    network = "default"
    access_config {
      # Ephemeral public IP
    }
  }

  metadata_startup_script = templatefile("${path.module}/startup.sh.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  })

  service_account {
    scopes = ["compute-ro", "storage-ro"]
  }

  count = 10
}
```

### Azure VM 예시

```hcl
resource "azurerm_linux_virtual_machine" "gpu_node" {
  name                = "ml-training-node-${count.index + 1}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_NC24ads_A100_v4"  # A100 x1

  admin_username = "azureuser"

  admin_ssh_key {
    username   = "azureuser"
    public_key = file("~/.ssh/id_rsa.pub")
  }

  network_interface_ids = [
    azurerm_network_interface.gpu_node[count.index].id,
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  custom_data = base64encode(templatefile("${path.module}/cloud-init.yaml.tpl", {
    config_server_url = var.config_server_url
    bootstrap_token   = var.bootstrap_token
  }))

  count = 10
}
```

## 등록 후 확인

### 1. Config Server 확인

```bash
# 등록된 모든 타겟 조회
curl http://config-server:8080/api/v1/targets

# 특정 타겟 상태 확인
curl http://config-server:8080/api/v1/targets/TARGET_ID
```

### 2. Prometheus 확인

**Web UI 접속**: http://prometheus:9090/targets

확인 사항:
- Target이 "UP" 상태인지
- Last Scrape 시간이 최근인지
- 에러 메시지가 없는지

### 3. Grafana 확인

**대시보드 접속**: http://grafana:3000

확인 사항:
- 노드 목록에 나타나는지
- 메트릭이 수집되는지 (CPU, 메모리, 디스크)
- GPU 메트릭이 보이는지 (GPU 노드인 경우)

### 4. 노드에서 직접 확인

```bash
# Node Exporter 메트릭
curl http://localhost:9100/metrics | grep node_

# DCGM Exporter 메트릭 (GPU 있는 경우)
curl http://localhost:9400/metrics | grep DCGM_

# 동적 체크 결과
ls -la /var/lib/node_exporter/textfile/
cat /var/lib/node_exporter/textfile/check_mount.prom
```

## 문제 해결

### 문제 1: Bootstrap 스크립트 실행 실패

**증상**: curl 명령어 실행 시 오류

**원인**:
- Config Server에 접근할 수 없음
- Bootstrap token이 만료되었거나 사용 횟수 초과

**해결 방법**:

```bash
# Config Server 연결 테스트
curl -I http://config-server:8080/api/v1/health

# Bootstrap token 상태 확인 (관리자)
curl http://config-server:8080/api/v1/bootstrap-tokens/TOKEN_ID

# 새 토큰 발급
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

### 문제 2: Prometheus에 타겟이 나타나지 않음

**증상**: Config Server에는 등록되었지만 Prometheus에서 보이지 않음

**원인**:
- Service Discovery 파일이 업데이트되지 않음
- Prometheus가 SD 파일을 읽지 못함

**해결 방법**:

```bash
# SD 파일 확인
curl http://config-server:8080/api/v1/sd/prometheus

# Prometheus 재시작
docker-compose restart prometheus

# Prometheus 로그 확인
docker-compose logs -f prometheus
```

### 문제 3: 노드에서 메트릭이 수집되지 않음

**증상**: Target은 UP 상태지만 메트릭 값이 없음

**원인**:
- Node Exporter가 실행되지 않음
- 방화벽이 포트를 차단함

**해결 방법**:

```bash
# 노드에서 확인
systemctl status node_exporter
systemctl status dcgm-exporter  # GPU 노드

# 방화벽 확인 및 오픈
sudo ufw status
sudo ufw allow 9100/tcp
sudo ufw allow 9400/tcp

# 로컬에서 테스트
curl http://localhost:9100/metrics
```

### 문제 4: 동적 체크가 실행되지 않음

**증상**: Check 메트릭이 나타나지 않음

**원인**:
- Cron이 실행되지 않음
- 스크립트 다운로드 실패

**해결 방법**:

```bash
# Cron 작업 확인
crontab -l
cat /etc/cron.d/aami-dynamic-check

# 수동으로 실행하여 에러 확인
/opt/aami/scripts/dynamic-check.sh

# 로그 확인
grep aami /var/log/syslog
```

## 대량 배포 가이드

### 시나리오: 100대의 GPU 노드 배포

#### 1. 준비 단계

```bash
# 1. 그룹 생성
curl -X POST http://config-server:8080/api/v1/groups \
  -d '{"name": "ml-cluster-batch-01", "namespace": "logical"}'

# 2. Bootstrap token 생성 (max_uses=100)
curl -X POST http://config-server:8080/api/v1/bootstrap-tokens \
  -d '{
    "name": "batch-01-token",
    "default_group_id": "GROUP_ID",
    "max_uses": 100,
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

#### 2. Terraform 배포

```bash
# Terraform 변수 설정
export TF_VAR_bootstrap_token="aami_bootstrap_xxxxx..."
export TF_VAR_node_count=100

# 배포 실행
terraform init
terraform plan
terraform apply
```

#### 3. 배포 진행 모니터링

```bash
# 등록된 노드 수 확인
watch -n 5 'curl -s http://config-server:8080/api/v1/targets | jq length'

# Prometheus target 수 확인
curl http://prometheus:9090/api/v1/targets | jq '.data.activeTargets | length'
```

#### 4. 배포 검증

```bash
# 모든 노드가 UP 상태인지 확인
curl http://prometheus:9090/api/v1/targets | \
  jq '.data.activeTargets[] | select(.health != "up") | .labels.instance'

# GPU 메트릭 수집 확인
curl -G http://prometheus:9090/api/v1/query \
  --data-urlencode 'query=count(DCGM_FI_DEV_GPU_TEMP)' | \
  jq '.data.result[0].value[1]'
```

## 모범 사례

### Bootstrap Token 관리

**DO**:
- ✅ 용도별로 별도의 토큰 생성 (개발/스테이징/프로덕션)
- ✅ 만료 기간 설정
- ✅ 사용 횟수 제한 설정
- ✅ 사용 후 비활성화

**DON'T**:
- ❌ 하나의 토큰을 모든 환경에서 재사용
- ❌ 만료 기간 없이 생성
- ❌ Public 저장소에 토큰 커밋
- ❌ 슬랙이나 이메일로 토큰 전송

### 그룹 설계

**권장 구조**:

```
environment (namespace)
├── production
│   ├── critical
│   └── standard
├── staging
└── development

infrastructure (namespace)
├── datacenter-seoul
│   ├── zone-a
│   └── zone-b
└── datacenter-tokyo

logical (namespace)
├── ml-training
├── ml-inference
├── api-servers
└── databases
```

### 라벨 전략

**유용한 라벨**:
- `environment`: production, staging, development
- `cluster`: 클러스터 이름
- `datacenter`: 데이터센터 위치
- `rack`: 랙 번호
- `gpu_model`: GPU 모델명
- `gpu_count`: GPU 개수
- `owner`: 소유 팀
- `cost_center`: 비용 센터

## 다음 단계

노드 등록 후:

1. **알림 규칙 설정**: [알림 규칙 가이드](./ALERT-RULES.md) 참조
2. **대시보드 생성**: [대시보드 가이드](./DASHBOARDS.md) 참조
3. **동적 체크 추가**: [체크 스크립트 관리](./CHECK-SCRIPT-MANAGEMENT.md) 참조
4. **용량 계획**: 메트릭 데이터 보관 정책 설정

## 참고 자료

- [빠른 시작 가이드](./QUICKSTART.md)
- [API 문서](./API.md)
- [체크 스크립트 관리](./CHECK-SCRIPT-MANAGEMENT.md)
- [문제 해결 가이드](./TROUBLESHOOTING.md)
