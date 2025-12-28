# Cloud-Init 통합 가이드

이 가이드는 다양한 클라우드 제공업체에서 cloud-init을 사용하여 AAMI 모니터링을 클라우드 인스턴스 초기화와 통합하는 방법을 설명합니다.

## 목차

1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [부트스트랩 프로세스](#부트스트랩-프로세스)
4. [클라우드 제공업체 통합](#클라우드-제공업체-통합)
   - [AWS EC2](#aws-ec2)
   - [Google Cloud Compute Engine](#google-cloud-compute-engine)
   - [Azure Virtual Machines](#azure-virtual-machines)
5. [Terraform 통합](#terraform-통합)
6. [커스터마이징](#커스터마이징)
7. [문제 해결](#문제-해결)
8. [모범 사례](#모범-사례)

## 개요

Cloud-init은 첫 부팅 시 클라우드 인스턴스의 자동 구성을 가능하게 합니다. AAMI는 다음 작업을 수행하는 cloud-init 스크립트를 제공합니다:

- AAMI Config Server에 **인스턴스 등록**
- **모니터링 에이전트 설치** (Node Exporter, DCGM Exporter)
- 하드웨어 상태 모니터링을 위한 **동적 체크 배포**
- systemd 타이머를 통한 **자동 업데이트 구성**

**장점**:
- 무접촉 모니터링 배포
- 환경 전반에 걸친 일관된 구성
- 그룹 기반 정책에 자동 등록
- Infrastructure as Code 호환

## 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│  클라우드 제공업체 콘솔 / Terraform / CLI                      │
│  - cloud-init 스크립트로 VM 시작                              │
│  - 부트스트랩 토큰 및 구성 전달                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ VM 부팅
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  VM 인스턴스 - Cloud-Init 실행                              │
│  1. 시스템 패키지 업데이트                                    │
│  2. 의존성 설치 (jq, curl, smartmontools 등)                │
│  3. AAMI 부트스트랩 API 호출                                  │
│     POST /api/v1/bootstrap/register                         │
│  4. textfile collector와 함께 Node Exporter 설치            │
│  5. dynamic-check.sh 스크립트 설치                          │
│  6. systemd 타이머 구성 (1분 간격)                           │
│  7. 설치 검증                                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ 등록
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  AAMI Config Server                                         │
│  - 타겟 레코드 생성                                           │
│  - 주 그룹 + 보조 그룹에 할당                                 │
│  - 유효한 체크 구성 반환                                      │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ 폴링 (1분마다)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  dynamic-check.sh                                           │
│  - 유효한 체크 가져오기                                       │
│  - 체크 스크립트 실행                                         │
│  - textfile collector에 메트릭 작성                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ 스크래핑
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Prometheus                                                 │
│  - 모든 메트릭 수집 (시스템 + 커스텀)                          │
└─────────────────────────────────────────────────────────────┘
```

## 부트스트랩 프로세스

### 1. 부트스트랩 토큰 생성

먼저 AAMI Config Server에서 부트스트랩 토큰을 생성합니다:

```bash
curl -X POST http://config-server:8080/api/v1/bootstrap/tokens \
  -H "Content-Type: application/json" \
  -d '{
    "token": "my-secure-bootstrap-token-12345",
    "description": "AWS 프로덕션 GPU 노드용 토큰",
    "default_group_id": 123,
    "expires_at": "2024-12-31T23:59:59Z",
    "max_uses": 100
  }'
```

**토큰 속성**:
- **Token**: 인증에 사용되는 비밀 문자열
- **Default Group**: 등록된 노드의 주 그룹
- **Expiration**: 토큰 유효 기간
- **Max Uses**: 등록 횟수 제한

### 2. 인스턴스 부트스트랩

cloud-init으로 인스턴스가 부팅되면:

1. 클라우드 제공업체로부터 **메타데이터 수집** (인스턴스 ID, 타입, 리전 등)
2. 토큰과 메타데이터로 **부트스트랩 API 호출**
3. **구성 수신** (노드 ID, 할당된 그룹, 체크)
4. **컴포넌트 설치** (Node Exporter, 체크 스크립트)
5. **모니터링 시작** (systemd 타이머가 체크 실행 시작)

### 3. 검증

부트스트랩 완료 후:

```bash
# 인스턴스에 SSH 접속
ssh ubuntu@<instance-ip>

# 부트스트랩 로그 확인
sudo cat /var/log/aami-bootstrap.log

# Node Exporter 확인
systemctl status node_exporter
curl http://localhost:9100/metrics

# 동적 체크 확인
systemctl status aami-dynamic-check.timer
sudo tail -f /var/log/aami/dynamic-check.log

# 메트릭 확인
curl http://localhost:9100/metrics | grep -E 'aami_|mount_check|disk_smart'
```

## 클라우드 제공업체 통합

### AWS EC2

#### User Data 스크립트

AWS EC2는 cloud-init을 위해 **user data**를 사용합니다. 스크립트는 첫 부팅 시 root로 실행됩니다.

**템플릿**: `examples/cloud-init/aws-ec2-userdata.sh`

#### 주요 기능
- EC2 메타데이터 가져오기 (인스턴스 ID, 타입, AZ, IP)
- 메타데이터 접근을 위한 `ec2-metadata` 명령 사용
- EBS 볼륨 연결 지원
- UFW 방화벽 규칙 구성

#### 수동 배포

1. **AWS 콘솔로 인스턴스 생성**:
   - 인스턴스 시작 마법사
   - 고급 세부 정보 → User data
   - `aws-ec2-userdata.sh` 내용 복사
   - `${bootstrap_token}`, `${config_server_url}`, `${primary_group}` 교체

2. **AWS CLI 사용**:

```bash
aws ec2 run-instances \
  --image-id ami-0c7217cdde317cfec \
  --instance-type p4d.24xlarge \
  --key-name my-ssh-key \
  --subnet-id subnet-xxxxx \
  --security-group-ids sg-xxxxx \
  --user-data file://aws-ec2-userdata.sh \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=aami-gpu-node}]'
```

#### Terraform 통합

[examples/terraform/aws-gpu-instance.tf](../examples/terraform/aws-gpu-instance.tf) 참조

```hcl
resource "aws_instance" "gpu_node" {
  ami           = "ami-0c7217cdde317cfec"
  instance_type = "p4d.24xlarge"

  user_data = templatefile("${path.module}/../cloud-init/aws-ec2-userdata.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:aws/us-east-1"
  })
}
```

### Google Cloud Compute Engine

#### Startup 스크립트

GCP는 초기화를 위해 **메타데이터 startup-script**를 사용합니다.

**템플릿**: `examples/cloud-init/gcp-startup-script.sh`

#### 주요 기능
- GCP 메타데이터 API 접근 (`http://metadata.google.internal`)
- 프로젝트 ID, 존, 머신 타입 가져오기
- gcloud를 통한 방화벽 규칙 구성
- 내부 및 외부 IP 모두 지원

#### 수동 배포

1. **GCP 콘솔로 인스턴스 생성**:
   - 인스턴스 만들기
   - 관리 → 자동화 → Startup 스크립트
   - `gcp-startup-script.sh` 내용 복사
   - 변수 교체

2. **gcloud CLI 사용**:

```bash
gcloud compute instances create aami-gpu-node-0 \
  --zone=us-central1-a \
  --machine-type=a2-highgpu-8g \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=100GB \
  --boot-disk-type=pd-ssd \
  --metadata-from-file startup-script=gcp-startup-script.sh \
  --tags=aami-gpu-node
```

#### Terraform 통합

[examples/terraform/gcp-gpu-instance.tf](../examples/terraform/gcp-gpu-instance.tf) 참조

```hcl
resource "google_compute_instance" "gpu_node" {
  name         = "aami-gpu-node-0"
  machine_type = "a2-highgpu-8g"

  metadata_startup_script = templatefile("${path.module}/../cloud-init/gcp-startup-script.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:gcp/us-central1"
  })
}
```

### Azure Virtual Machines

#### Custom Data

Azure는 cloud-init 구성을 위해 **custom data**를 사용합니다.

**템플릿**: `examples/cloud-init/azure-custom-data.yaml`

#### 주요 기능
- cloud-config YAML 형식 사용
- Azure Instance Metadata Service (IMDS) 접근
- VM ID, 리소스 그룹, 구독 가져오기
- cloud-init 지시문 지원 (packages, write_files, runcmd)

#### 수동 배포

1. **Azure Portal로 VM 생성**:
   - 가상 머신 만들기
   - 고급 → Custom data
   - `azure-custom-data.yaml` 내용 복사
   - 변수 교체

2. **Azure CLI 사용**:

```bash
az vm create \
  --resource-group my-resource-group \
  --name aami-gpu-node-0 \
  --location eastus \
  --size Standard_NC24ads_A100_v4 \
  --image Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest \
  --custom-data azure-custom-data.yaml \
  --admin-username azureuser \
  --ssh-key-values @~/.ssh/id_rsa.pub
```

#### Terraform 통합

[examples/terraform/azure-gpu-instance.tf](../examples/terraform/azure-gpu-instance.tf) 참조

```hcl
resource "azurerm_linux_virtual_machine" "gpu_node" {
  name                = "aami-gpu-node-0"
  size                = "Standard_NC24ads_A100_v4"

  custom_data = base64encode(templatefile("${path.module}/../cloud-init/azure-custom-data.yaml", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = "infrastructure:azure/eastus"
  }))
}
```

## Terraform 통합

### 완전한 워크플로우

1. **Terraform 변수 생성**:

```hcl
# variables.tf
variable "aami_config_server_url" {
  description = "AAMI Config Server URL"
  type        = string
}

variable "aami_bootstrap_token" {
  description = "자동 등록을 위한 부트스트랩 토큰"
  type        = string
  sensitive   = true
}

variable "aami_primary_group" {
  description = "모니터링을 위한 주 그룹"
  type        = string
  default     = "infrastructure:cloud/region"
}
```

2. **인스턴스 정의에 사용**:

```hcl
# main.tf
resource "aws_instance" "gpu" {
  # ... 인스턴스 구성 ...

  user_data = templatefile("${path.module}/cloud-init.sh", {
    bootstrap_token    = var.aami_bootstrap_token
    config_server_url  = var.aami_config_server_url
    primary_group      = var.aami_primary_group
  })
}
```

3. **배포**:

```bash
terraform init
terraform plan
terraform apply
```

### 멀티 클라우드 배포

여러 클라우드에 동일한 모니터링 배포:

```hcl
# main.tf
module "aws_gpu" {
  source = "./modules/aws-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:aws/us-east-1"
}

module "gcp_gpu" {
  source = "./modules/gcp-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:gcp/us-central1"
}

module "azure_gpu" {
  source = "./modules/azure-gpu"

  aami_config_server_url = var.aami_config_server_url
  aami_bootstrap_token   = var.aami_bootstrap_token
  aami_primary_group     = "infrastructure:azure/eastus"
}
```

## 커스터마이징

### 커스텀 초기화 추가

cloud-init 스크립트를 커스텀 단계로 확장:

```bash
#!/bin/bash
# ... 기존 AAMI 부트스트랩 코드 ...

# 커스텀 초기화
echo "[CUSTOM] 추가 패키지 설치 중..."
apt-get install -y \
    nvidia-driver-535 \
    cuda-toolkit-12-2

# GPU 설정 구성
echo "[CUSTOM] GPU 영속성 모드 구성 중..."
nvidia-smi -pm 1

# 추가 스토리지 마운트
echo "[CUSTOM] NFS 공유 마운트 중..."
mkdir -p /mnt/shared
mount -t nfs nfs-server.example.com:/export/shared /mnt/shared

# 커스텀 모니터링 체크 설치
echo "[CUSTOM] 커스텀 체크 설치 중..."
curl -o /usr/local/lib/aami/checks/my-custom-check.sh \
    https://my-repo.com/checks/my-custom-check.sh
chmod +x /usr/local/lib/aami/checks/my-custom-check.sh
```

### 환경별 구성

환경에 따라 다른 그룹 사용:

```hcl
# Terraform
locals {
  environment = terraform.workspace

  aami_group_mapping = {
    production  = "infrastructure:aws/prod/us-east-1"
    staging     = "infrastructure:aws/staging/us-east-1"
    development = "infrastructure:aws/dev/us-east-1"
  }
}

resource "aws_instance" "node" {
  user_data = templatefile("${path.module}/cloud-init.sh", {
    primary_group = local.aami_group_mapping[local.environment]
  })
}
```

### 조건부 체크 설치

특정 인스턴스 타입에만 체크 설치:

```bash
#!/bin/bash
INSTANCE_TYPE=$(ec2-metadata --instance-type | cut -d ' ' -f 2)

# GPU 인스턴스에만 GPU 체크 설치
if [[ "$INSTANCE_TYPE" =~ ^(p3|p4|p5|g4|g5)\. ]]; then
    echo "GPU 모니터링을 위한 DCGM Exporter 설치 중..."
    curl -L https://github.com/NVIDIA/dcgm-exporter/releases/download/3.1.7-3.1.4/dcgm-exporter_3.1.7-3.1.4_amd64.deb \
        -o /tmp/dcgm-exporter.deb
    dpkg -i /tmp/dcgm-exporter.deb
fi

# IB가 있는 인스턴스에만 InfiniBand 체크 설치
if [ -d "/sys/class/infiniband" ]; then
    echo "InfiniBand 모니터링 설치 중..."
    apt-get install -y infiniband-diags
fi
```

## 문제 해결

### Cloud-Init이 실행되지 않음

```bash
# cloud-init 상태 확인
cloud-init status

# cloud-init 로그 보기
sudo cat /var/log/cloud-init.log
sudo cat /var/log/cloud-init-output.log

# cloud-init 재실행 (테스트 전용)
sudo cloud-init clean --logs
sudo cloud-init init
sudo cloud-init modules --mode=config
sudo cloud-init modules --mode=final
```

### 부트스트랩 API 실패

```bash
# 부트스트랩 로그 확인
sudo cat /var/log/aami-bootstrap.log

# Config Server 연결 테스트
curl -v http://config-server:8080/api/v1/health

# 부트스트랩 API 수동 테스트
curl -X POST http://config-server:8080/api/v1/bootstrap/register \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your-bootstrap-token",
    "hostname": "test-node",
    "primary_group": "infrastructure:test"
  }'
```

### 컴포넌트 설치 실패

```bash
# 설치 오류 확인
sudo journalctl -xe

# 개별 컴포넌트 테스트
sudo systemctl status node_exporter
sudo systemctl status aami-dynamic-check.timer

# 컴포넌트 수동 설치
sudo /path/to/install-node-exporter.sh
sudo /usr/local/bin/dynamic-check.sh --debug
```

### 네트워크 문제

```bash
# DNS 해결 테스트
nslookup config-server.example.com

# HTTP 연결 테스트
curl -v http://config-server:8080/api/v1/health

# 보안 그룹/방화벽 확인
# AWS
aws ec2 describe-security-groups --group-ids sg-xxxxx

# GCP
gcloud compute firewall-rules list --filter="name~aami"

# Azure
az network nsg show --resource-group xxx --name xxx
```

## 모범 사례

### 보안

1. **부트스트랩 토큰 보호**:
   - 안전한 비밀 관리에 저장 (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)
   - 짧은 만료 기간 사용
   - 정기적으로 순환
   - 토큰당 최대 사용 횟수 제한

2. **프라이빗 네트워크 사용**:
   - Config Server를 프라이빗 서브넷에 배포
   - 연결을 위해 VPC 피어링 또는 VPN 사용
   - 보안 그룹을 내부 트래픽만으로 제한

3. **암호화 활성화**:
   - Config Server에 HTTPS 사용
   - 인스턴스 디스크 암호화
   - 가능한 경우 secure boot 사용

### 멱등성

스크립트가 여러 번 안전하게 실행될 수 있도록 보장:

```bash
# 이미 설치되었는지 확인
if [ -f "/usr/local/bin/node_exporter" ]; then
    echo "Node Exporter가 이미 설치되어 있습니다. 건너뜁니다..."
    exit 0
fi

# 조건부 체크 사용
if ! systemctl is-active --quiet node_exporter; then
    echo "Node Exporter 시작 중..."
    systemctl start node_exporter
fi
```

### 오류 처리

견고한 오류 처리 추가:

```bash
set -euo pipefail  # 오류, 정의되지 않은 변수, 파이프 실패 시 종료

# 오류 트랩
trap 'echo "라인 $LINENO에서 오류"' ERR

# 네트워크 작업을 위한 재시도 로직
retry_count=0
max_retries=5
while [ $retry_count -lt $max_retries ]; do
    if curl -f http://config-server:8080/api/v1/health; then
        break
    fi
    retry_count=$((retry_count + 1))
    sleep 10
done
```

### 로깅

디버깅을 위한 포괄적인 로깅:

```bash
# 파일과 stdout에 로깅
exec > >(tee /var/log/aami-bootstrap.log)
exec 2>&1

# 모든 출력에 타임스탬프 추가
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log "AAMI 부트스트랩 시작 중..."
```

### 테스트

프로덕션 전에 cloud-init 스크립트 테스트:

```bash
# Docker로 로컬 테스트
docker run -it ubuntu:22.04 bash
# 스크립트를 붙여넣고 수동으로 실행

# Vagrant로 테스트
vagrant up
vagrant ssh
sudo cat /var/log/aami-bootstrap.log

# cloud-init 구문 테스트
cloud-init schema --config-file azure-custom-data.yaml
```

## 참고 자료

- [Cloud-Init 문서](https://cloudinit.readthedocs.io/)
- [AWS EC2 User Data](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html)
- [GCP Startup Scripts](https://cloud.google.com/compute/docs/instances/startup-scripts)
- [Azure Custom Data](https://docs.microsoft.com/en-us/azure/virtual-machines/custom-data)
- [AAMI Bootstrap API](/docs/kr/API.md#bootstrap-api)
