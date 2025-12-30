# AI Accelerator Monitoring Infrastructure - Project Plan

## 📋 프로젝트 개요

### 목적
다양한 AI 가속기(GPU, NPU, TPU 등)를 탑재한 서버들을 효율적으로 모니터링하고 관리하기 위한 중앙 집중식 모니터링 인프라를 구축합니다. Prometheus, Grafana, Alertmanager 기반의 확장 가능하고 동적으로 설정 변경이 가능한 모니터링 시스템을 제공합니다.

### 핵심 요구사항
1. **동적 설정 관리**: 마운트 포인트, 디바이스, 알림 설정을 중앙에서 동적으로 관리
2. **다양한 가속기 지원**: NVIDIA GPU, AMD GPU, Intel Gaudi, Rebellions NPU, Furiosa NPU, Tenstorrent, Google TPU 등 다양한 AI 가속기 통합 모니터링 (all-smi 활용)
3. **통합 모니터링 접근**: all-smi를 주력 도구로 사용하여 단일 Exporter로 다중 가속기 모니터링, 필요시 벤더별 심화 메트릭 추가
4. **하이브리드 인프라 지원**: On-Premise 물리 서버 및 Cloud VM (AWS EC2, GCP Compute Engine, Azure VM 등) 통합 관리
5. **고속 인프라 모니터링**: InfiniBand/RoCE 등 고속 네트워크, NVMe/병렬 파일시스템 등 고속 스토리지 접근성 체크
6. **커스텀 체크**: 사용자 정의 명령어 및 스크립트를 통한 유연한 모니터링
7. **다양한 알림 채널**: SMTP, Webhook, Slack 등 다양한 프로토콜 지원
8. **간편한 배포**: Docker Compose 또는 Kubernetes 기반 배포, 폐쇄망 환경 지원

### 주요 기술 스택
- **모니터링**: Prometheus, Grafana
- **알림**: Alertmanager
- **AI 가속기 메트릭** (하이브리드 접근):
  - **all-smi** (통합 가속기 모니터링 - 주력)
    - NVIDIA GPU, AMD GPU, Intel Gaudi, Rebellions NPU, Furiosa NPU, Tenstorrent, Google TPU 지원
    - Prometheus 메트릭 API 내장
  - **DCGM Exporter** (NVIDIA GPU - 상세 메트릭용)
    - all-smi와 병행하여 NVIDIA GPU의 심화 메트릭 수집
- **시스템 메트릭**: Node Exporter
- **서비스 디스커버리**: HTTP SD, File SD, Consul (선택)
- **설정 관리**:
  - Config Server (Go, REST API)
  - Config Server UI (Next.js, 선택)
  - GitOps (Ansible/Flux)
- **컨테이너**: Docker, Docker Compose, Kubernetes

---

## 🏗️ 저장소 구조

```
ai-accelerator-monitoring/
│
├── README.md                          # 프로젝트 소개 및 빠른 시작 가이드
├── LICENSE                            # 라이선스 정보
├── CONTRIBUTING.md                    # 기여 가이드라인
├── .gitignore                         # Git 제외 파일 목록
│
├── docs/                              # 문서화
│   ├── architecture.md                # 아키텍처 설계 문서
│   ├── installation.md                # 상세 설치 가이드
│   ├── configuration.md               # 설정 가이드
│   ├── troubleshooting.md             # 문제 해결 가이드
│   ├── api-reference.md               # Config Server API 문서
│   └── diagrams/                      # 아키텍처 다이어그램
│       ├── system-overview.png
│       ├── data-flow.png
│       └── deployment-architecture.png
│
├── deploy/                            # 배포 관련 파일
│   ├── docker-compose/                # Docker Compose 배포
│   │   ├── docker-compose.yml         # 메인 compose 파일
│   │   ├── docker-compose.dev.yml     # 개발 환경용
│   │   ├── docker-compose.prod.yml    # 프로덕션 환경용
│   │   ├── .env.example               # 환경변수 예시
│   │   └── README.md                  # Docker Compose 배포 가이드
│   │
│   ├── offline/                       # 폐쇄망 환경 배포
│   │   ├── create-bundle.sh           # 오프라인 패키지 번들 생성
│   │   ├── install-offline.sh         # 폐쇄망 환경 설치 스크립트
│   │   ├── setup-local-registry.sh    # 로컬 Docker 레지스트리 구축
│   │   ├── packages/                  # 사전 다운로드 패키지 저장
│   │   │   ├── binaries/              # Go 컴파일 바이너리
│   │   │   │   ├── config-server
│   │   │   │   └── custom-exporter
│   │   │   ├── ui-static/             # Config Server UI 정적 빌드 (선택)
│   │   │   │   └── config-server-ui.tar.gz
│   │   │   ├── docker-images/         # Docker 이미지 tar 파일
│   │   │   ├── debs/                  # Debian 패키지
│   │   │   ├── rpms/                  # RPM 패키지
│   │   │   └── python-wheels/         # Python wheel (스크립트용, 선택)
│   │   └── README.md                  # 폐쇄망 설치 가이드
│   │
│   ├── kubernetes/                    # Kubernetes 배포
│   │   ├── namespace.yaml
│   │   ├── prometheus/                # Prometheus 관련 리소스
│   │   │   ├── deployment.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── service.yaml
│   │   │   └── pvc.yaml
│   │   ├── grafana/                   # Grafana 관련 리소스
│   │   ├── alertmanager/              # Alertmanager 관련 리소스
│   │   ├── config-server/             # Config Server 관련 리소스
│   │   ├── config-server-ui/          # Config Server UI 리소스 (선택)
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   └── ingress.yaml
│   │   └── kustomization.yaml
│   │
│   └── ansible/                       # Ansible 배포 스크립트
│       ├── inventory/
│       │   ├── hosts.example
│       │   └── group_vars/
│       ├── playbooks/
│       │   ├── deploy-monitoring.yml
│       │   ├── deploy-nodes.yml
│       │   ├── install-all-smi.yml
│       │   └── update-config.yml
│       └── roles/
│           ├── prometheus/
│           ├── node-exporter/
│           ├── all-smi/              # all-smi 설치 및 설정
│           ├── dcgm-exporter/        # 선택적 DCGM 설치
│           ├── vendor-drivers/       # 벤더 드라이버 설치
│           └── custom-checks/
│
├── config/                            # 설정 파일 템플릿
│   ├── prometheus/
│   │   ├── prometheus.yml.j2          # Prometheus 설정 템플릿
│   │   ├── rules/                     # 알림 규칙
│   │   │   ├── nvidia-gpu-alerts.yml
│   │   │   ├── amd-gpu-alerts.yml
│   │   │   ├── intel-gaudi-alerts.yml
│   │   │   ├── rebellions-npu-alerts.yml
│   │   │   ├── furiosa-npu-alerts.yml
│   │   │   ├── tenstorrent-alerts.yml
│   │   │   ├── storage-alerts.yml
│   │   │   ├── system-alerts.yml
│   │   │   └── custom-alerts.yml.example
│   │   └── sd/                        # Service Discovery 설정
│   │       ├── file-sd.json.example
│   │       └── http-sd-config.yml
│   │
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/
│   │   │   │   └── prometheus.yml
│   │   │   └── dashboards/
│   │   │       └── default.yml
│   │   └── dashboards/                # 대시보드 JSON 파일
│   │       ├── accelerator-overview.json      # 모든 가속기 통합 뷰
│   │       ├── nvidia-gpu-detailed.json
│   │       ├── amd-gpu-detailed.json
│   │       ├── intel-gaudi-detailed.json
│   │       ├── rebellions-npu-detailed.json
│   │       ├── furiosa-npu-detailed.json
│   │       ├── tenstorrent-detailed.json
│   │       ├── system-metrics.json
│   │       ├── storage-monitoring.json
│   │       └── alert-overview.json
│   │
│   ├── alertmanager/
│   │   ├── alertmanager.yml.j2        # Alertmanager 설정 템플릿
│   │   └── templates/
│   │       ├── email.tmpl
│   │       └── slack.tmpl
│   │
│   └── node-exporter/
│       └── textfile-collector/
│           └── README.md              # textfile collector 사용법
│
├── services/                          # 커스텀 서비스
│   ├── config-server/                 # 중앙 설정 관리 API 서버 (Go)
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── main.go                    # 메인 진입점
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go            # 서버 실행
│   │   ├── internal/
│   │   │   ├── api/                   # API 핸들러
│   │   │   │   ├── targets.go         # 타겟 관리
│   │   │   │   ├── checks.go          # 체크 설정
│   │   │   │   └── alerts.go          # 알림 설정
│   │   │   ├── models/                # 데이터 모델
│   │   │   │   └── models.go
│   │   │   ├── database/              # DB 연결
│   │   │   │   └── postgres.go
│   │   │   ├── service/               # 비즈니스 로직
│   │   │   │   └── config.go
│   │   │   └── utils/
│   │   │       ├── prometheus.go      # Prometheus 연동
│   │   │       └── validation.go      # 설정 검증
│   │   ├── pkg/                       # 공용 패키지
│   │   ├── tests/
│   │   ├── Makefile                   # 빌드 스크립트
│   │   └── README.md
│   │
│   ├── config-server-ui/              # Config Server 관리 UI (선택)
│   │   ├── Dockerfile                 # nginx + 정적 파일
│   │   ├── package.json
│   │   ├── next.config.ts             # Static export 설정
│   │   ├── tsconfig.json
│   │   ├── tailwind.config.ts
│   │   ├── components.json            # shadcn/ui 설정
│   │   ├── src/
│   │   │   ├── app/                   # Next.js App Router
│   │   │   │   ├── layout.tsx
│   │   │   │   ├── page.tsx           # 대시보드
│   │   │   │   ├── groups/            # 그룹 관리 페이지
│   │   │   │   ├── targets/           # 타겟 관리 페이지
│   │   │   │   ├── checks/            # 체크 설정 페이지
│   │   │   │   └── alerts/            # 알림 설정 페이지
│   │   │   ├── components/            # UI 컴포넌트
│   │   │   │   └── ui/                # shadcn/ui 컴포넌트
│   │   │   ├── lib/
│   │   │   │   ├── api-client.ts      # Config Server API 클라이언트
│   │   │   │   └── utils.ts
│   │   │   └── types/
│   │   │       └── config.ts          # 타입 정의
│   │   ├── public/                    # 정적 파일
│   │   ├── Makefile                   # 빌드 스크립트
│   │   └── README.md
│   │
│   └── exporters/                     # Exporter 설정 및 커스텀
│       ├── README.md                  # all-smi 및 Exporter 통합 가이드
│       │
│       ├── all-smi/                   # all-smi 설정
│       │   ├── config.yaml.example    # all-smi 설정 파일
│       │   ├── deployment.yaml        # K8s DaemonSet
│       │   ├── docker-compose.yml     # Docker Compose 설정
│       │   ├── systemd/               # Systemd 서비스
│       │   │   └── all-smi.service
│       │   └── README.md              # all-smi 배포 가이드
│       │
│       ├── dcgm/                      # NVIDIA DCGM Exporter (선택)
│       │   ├── README.md
│       │   ├── deployment.yaml
│       │   └── dcgm-config.yaml
│       │
│       ├── vendor-drivers/            # 벤더 드라이버/SDK 설치 가이드
│       │   ├── nvidia-driver.md
│       │   ├── amd-rocm.md
│       │   ├── intel-gaudi.md
│       │   ├── rebellions-sdk.md
│       │   ├── furiosa-sdk.md
│       │   └── tenstorrent-sdk.md
│       │
│       ├── custom-exporter-template/  # 커스텀 Exporter 템플릿 (필요시)
│       │   ├── Dockerfile
│       │   ├── go.mod
│       │   ├── main.go                # 메인 진입점
│       │   ├── collector/
│       │   │   └── device_collector.go  # 디바이스 수집기
│       │   ├── Makefile
│       │   └── README.md              # 개발 가이드
│       │
│       └── mount-checker/             # 마운트 포인트 체커
│           ├── mount-checker.sh
│           └── README.md
│
├── scripts/                           # 유틸리티 스크립트
│   ├── node/                          # 각 서버에 배포되는 스크립트
│   │   ├── install-all-smi.sh         # all-smi 자동 설치
│   │   ├── install-vendor-driver.sh   # 벤더 드라이버 자동 설치
│   │   ├── setup-all-smi-api.sh       # all-smi API 모드 설정
│   │   ├── dynamic-check.sh           # 동적 체크 스크립트
│   │   └── mount-check.sh             # 마운트 포인트 체크
│   │
│   ├── management/                    # 관리 스크립트
│   │   ├── add-server.sh              # 신규 서버 추가
│   │   ├── remove-server.sh           # 서버 제거
│   │   ├── update-checks.sh           # 체크 설정 업데이트
│   │   ├── backup-config.sh           # 설정 백업
│   │   └── restore-config.sh          # 설정 복원
│   │
│   └── testing/                       # 테스트 스크립트
│       ├── test-connectivity.sh       # 연결성 테스트
│       ├── test-metrics.sh            # 메트릭 수집 테스트
│       └── simulate-alerts.sh         # 알림 시뮬레이션
│
├── examples/                          # 예제 및 튜토리얼
│   ├── basic-setup/                   # 기본 설정 예제
│   │   ├── all-smi-quickstart.md      # all-smi 빠른 시작
│   │   └── README.md
│   ├── config-server-ui/              # UI 사용 예제 (선택)
│   │   ├── screenshots/               # UI 스크린샷
│   │   ├── usage-guide.md             # UI 사용 가이드
│   │   └── README.md
│   ├── custom-checks/                 # 커스텀 체크 예제
│   │   ├── check-accelerator-health.sh    # 가속기 헬스체크
│   │   ├── check-infiniband.sh            # InfiniBand 링크 체크
│   │   ├── check-nvme-health.sh           # NVMe 상태 체크
│   │   ├── check-parallel-fs.sh           # 병렬 파일시스템 체크
│   │   ├── check-disk-smart.sh
│   │   ├── check-mount-points.sh
│   │   └── README.md
│   ├── all-smi-configs/               # all-smi 설정 예제
│   │   ├── nvidia-server.yaml         # NVIDIA 서버 설정
│   │   ├── mixed-accelerators.yaml    # 혼합 가속기 설정
│   │   ├── kubernetes-deployment.yaml # K8s 배포 예제
│   │   └── README.md
│   ├── offline-deployment/            # 폐쇄망 배포 예제
│   │   ├── bundle-creation.md         # 번들 생성 가이드
│   │   ├── airgap-install.md          # 완전 폐쇄망 설치
│   │   ├── local-registry.md          # 로컬 레지스트리 구축
│   │   └── README.md
│   ├── alert-configs/                 # 알림 설정 예제
│   │   ├── unified-accelerator-alerts.yml  # all-smi 통합 알림
│   │   ├── infra-health-alerts.yml    # 인프라 헬스 알림
│   │   ├── critical-alerts.yml
│   │   ├── team-routing.yml
│   │   └── README.md
│   └── dashboards/                    # 대시보드 예제
│       ├── all-smi-overview.json      # all-smi 통합 대시보드
│       ├── high-speed-infra.json      # 고속 인프라 대시보드
│       ├── per-accelerator-detail.json
│       └── README.md
│
├── tests/                             # 통합 테스트
│   ├── integration/
│   │   ├── test_prometheus.py
│   │   ├── test_alerting.py
│   │   └── test_config_server.py
│   ├── e2e/
│   │   └── test_full_workflow.py
│   └── fixtures/
│       └── sample-data.json
│
└── tools/                             # 개발 도구
    ├── dev-setup.sh                   # 개발 환경 설정
    ├── lint.sh                        # 코드 린팅
    ├── format.sh                      # 코드 포맷팅
    └── generate-docs.sh               # 문서 자동 생성
```

---

## 📐 아키텍처 개요

### 시스템 구성도

```
┌─────────────────────────────────────────────────────────────┐
│                    관리자 / 사용자                          │
└───────────┬─────────────────────────────────────────────────┘
            │
            │ (웹 UI, API 호출)
            │
┌───────────▼─────────────────────────────────────────────────┐
│                   중앙 모니터링 시스템                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Grafana    │  │  Prometheus  │  │ Alertmanager │      │
│  │   (UI)       │◄─┤  (TSDB)      │◄─┤  (알림)      │      │
│  └──────────────┘  └──────┬───────┘  └──────────────┘      │
│                            │                                 │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │Config Server │◄─┤Config UI     │ (선택)                  │
│  │(설정 관리 API)│  │(관리 웹 UI)  │                         │
│  └──────┬───────┘  └──────────────┘                         │
│         │                                                    │
└─────────┼──────────────────────────────────────────────────┘
          │
          │ (메트릭 수집)
          │
         ┌───────────────────┼───────────────────┬──────────────┐
         │                   │                   │              │
┌────────▼────────┐ ┌────────▼────────┐ ┌───────▼─────────┐ ┌─▼──────────┐
│ NVIDIA GPU 서버 │ │  Intel Gaudi    │ │ Rebellions NPU  │ │ 혼합 서버  │
│                 │ │     서버        │ │     서버        │ │            │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ Multiple   │
│ │Node Exporter│ │ │ │Node Exporter│ │ │ │Node Exporter│ │ │ Devices    │
│ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ │ │            │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌────────┐ │
│ │ all-smi API │ │ │ │ all-smi API │ │ │ │ all-smi API │ │ │ │all-smi │ │
│ │ (통합)      │ │ │ │ (Gaudi)     │ │ │ │ (Rebellions)│ │ │ │        │ │
│ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ │ │ └────────┘ │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌────────┐ │
│ │DCGM Exporter│ │ │ │Custom Checks│ │ │ │Custom Checks│ │ │ │DCGM    │ │
│ │(상세 메트릭)│ │ │ └─────────────┘ │ │ └─────────────┘ │ │ │(선택)  │ │
│ └─────────────┘ │ │                 │ │                 │ │ └────────┘ │
│ ┌─────────────┐ │ │                 │ │                 │ │            │
│ │Custom Checks│ │ │                 │ │                 │ │            │
│ └─────────────┘ │ │                 │ │                 │ │            │
└─────────────────┘ └─────────────────┘ └─────────────────┘ └────────────┘

         ┌──────────────────┐ ┌──────────────────┐
         │ Furiosa NPU 서버 │ │ Tenstorrent 서버 │
         │                  │ │                  │
         │ ┌──────────────┐ │ │ ┌──────────────┐ │
         │ │Node Exporter │ │ │ │Node Exporter │ │
         │ └──────────────┘ │ │ └──────────────┘ │
         │ ┌──────────────┐ │ │ ┌──────────────┐ │
         │ │ all-smi API  │ │ │ │ all-smi API  │ │
         │ │ (Furiosa)    │ │ │ │(Tenstorrent) │ │
         │ └──────────────┘ │ │ └──────────────┘ │
         │ ┌──────────────┐ │ │ ┌──────────────┐ │
         │ │Custom Checks │ │ │ │Custom Checks │ │
         │ └──────────────┘ │ │ └──────────────┘ │
         └──────────────────┘ └──────────────────┘
```

### 데이터 흐름

1. **메트릭 수집**: Prometheus가 각 서버의 Exporter로부터 메트릭 스크랩
2. **동적 타겟 관리**: Config Server가 Prometheus에 HTTP SD 또는 파일 기반 타겟 정보 제공
3. **알림 평가**: Prometheus가 알림 규칙 평가 후 Alertmanager로 전송
4. **알림 라우팅**: Alertmanager가 설정된 채널(SMTP, Slack, Webhook)로 알림 발송
5. **시각화**: Grafana가 Prometheus를 데이터 소스로 사용하여 대시보드 제공

---

## 🚀 핵심 기능

### 1. 동적 설정 관리 (Config Server)

#### 개요
중앙 집중식 API 서버를 통해 모니터링 대상, 체크 항목, 알림 설정을 동적으로 관리합니다.

#### 핵심 기능
- **타겟 관리**: 모니터링 대상 서버/디바이스 추가/제거/수정
- **그룹 관리**: 타겟을 계층적 그룹으로 조직화, 그룹 단위 설정 적용
- **체크 설정**: 서버별/그룹별 커스텀 체크(마운트, 디스크 등) 동적 설정
- **설정 상속**: 그룹 설정을 하위 타겟에 자동 적용, 개별 오버라이드 가능
- **알림 설정**: SMTP, Slack, Webhook 등 알림 채널 동적 변경
- **서비스 디스커버리**: Prometheus HTTP SD를 통한 자동 타겟 발견
- **설정 검증**: 변경 전 설정 유효성 검사
- **설정 리로드**: 서비스 재시작 없이 설정 적용

#### Config Server API 예시
```http
# 타겟 관리
GET    /api/v1/targets                # 모든 타겟 조회
POST   /api/v1/targets                # 신규 타겟 추가
PUT    /api/v1/targets/{id}           # 타겟 업데이트
DELETE /api/v1/targets/{id}           # 타겟 제거

# 그룹 관리
GET    /api/v1/groups                 # 모든 그룹 조회
POST   /api/v1/groups                 # 그룹 생성
PUT    /api/v1/groups/{id}            # 그룹 수정
DELETE /api/v1/groups/{id}            # 그룹 삭제
GET    /api/v1/groups/{id}/targets    # 그룹 내 타겟 조회
POST   /api/v1/groups/{id}/targets    # 그룹에 타겟 추가

# 체크 설정
GET    /api/v1/checks/{server_id}     # 체크 항목 조회
PUT    /api/v1/checks/{server_id}     # 체크 항목 업데이트
GET    /api/v1/groups/{id}/checks     # 그룹 체크 설정 조회
POST   /api/v1/groups/{id}/checks     # 그룹 체크 설정 추가
GET    /api/v1/targets/{id}/checks/effective  # 최종 체크 (상속 포함)

# 알림 설정
POST   /api/v1/alerts/config          # 알림 채널 설정 변경

# 알림 규칙 커스터마이징
GET    /api/v1/alert-rules/templates           # 알림 규칙 템플릿 조회
GET    /api/v1/alert-rules/templates/{name}    # 특정 템플릿 조회
GET    /api/v1/groups/{id}/alert-rules         # 그룹 알림 규칙 조회
POST   /api/v1/groups/{id}/alert-rules         # 그룹 알림 규칙 설정
PUT    /api/v1/groups/{id}/alert-rules/{rule_id}  # 그룹 알림 규칙 수정
DELETE /api/v1/groups/{id}/alert-rules/{rule_id}  # 그룹 알림 규칙 삭제
GET    /api/v1/targets/{id}/alert-rules/effective  # 타겟 최종 알림 규칙
POST   /api/v1/targets/{id}/alert-rules        # 타겟 알림 규칙 오버라이드
GET    /api/v1/targets/{id}/alert-rules/trace  # 알림 규칙 정책 추적

# Bootstrap & Auto Registration
POST   /api/v1/bootstrap/register     # 새 서버 자동 등록
POST   /api/v1/bootstrap/complete     # Bootstrap 완료 보고
GET    /api/v1/bootstrap/token        # Bootstrap 토큰 생성
GET    /api/v1/bootstrap/script       # Bootstrap 스크립트 다운로드

# SSH Agent & Remote Deploy
POST   /api/v1/targets/{id}/deploy    # 원격 배포 트리거
GET    /api/v1/targets/{id}/deploy/status  # 배포 상태 조회
POST   /api/v1/ssh-keys               # SSH 키 등록
GET    /api/v1/ssh-keys               # SSH 키 목록 조회
DELETE /api/v1/ssh-keys/{id}          # SSH 키 삭제

# Fleet Management (일괄 배포)
POST   /api/v1/fleet/deploy           # 그룹 일괄 배포
GET    /api/v1/fleet/deploy/{job_id}  # 배포 작업 상태 조회
POST   /api/v1/fleet/deploy/{job_id}/cancel  # 배포 작업 취소
GET    /api/v1/fleet/jobs             # 배포 작업 이력 조회

# 시스템
POST   /api/v1/reload                 # 설정 리로드 트리거
```

#### 기술 스택
- **Go (Gin/Fiber)**: REST API 프레임워크
- **PostgreSQL**: 설정 데이터 영구 저장
- **Redis**: 캐싱 및 세션 관리
- **pgx**: Go PostgreSQL 드라이버
- **go-redis**: Go Redis 클라이언트

#### 워크플로우
```
1. 사용자 → Config Server API 호출 (타겟/체크/알림 변경)
2. Config Server → 설정 검증 및 DB 저장
3. Config Server → Prometheus 설정 파일 생성 또는 HTTP SD 응답
4. Config Server → Prometheus/Alertmanager 리로드 API 호출
5. 변경사항 즉시 반영 (서비스 재시작 불필요)
```

상세 API 문서는 `docs/api-reference.md` 참조

#### Namespace 기반 그룹 관리

##### 개요
타겟을 Namespace와 Group으로 조직화하여 하이브리드 인프라(On-Premise + Cloud)를 통합 관리합니다. Namespace는 인프라, 논리적 분류, 환경 등 서로 다른 도메인을 분리하고, 각 Namespace 내에서 계층적 그룹 구조를 가집니다.

##### 핵심 개념

**Namespace**: 서로 다른 관점의 분류 도메인
- **infrastructure**: 물리적/클라우드 인프라 위치
- **logical**: 프로젝트, 클러스터, 워크로드 등 논리적 그룹
- **environment**: production, staging, development 등 환경

**Group**: 각 Namespace 내의 계층적 구조
- 각 Namespace는 독립적인 계층 구조
- Full path로 식별: `namespace:path/to/group`
- 예: `infrastructure:aws/us-east-1/us-east-1a`, `logical:ml-training/gpu-workers`

##### 주요 기능
- **하이브리드 인프라 통합**: On-Premise 물리 서버와 Cloud VM을 동일한 방식으로 관리
- **도메인 분리**: 물리적, 논리적, 환경별 분류를 Namespace로 명확히 분리
- **명확한 정책 우선순위**: Namespace 레벨에서 정책 우선순위 정의 (environment > logical > infrastructure)
- **유연한 다중 분류**: 하나의 타겟이 여러 Namespace의 그룹에 소속 가능
- **Provider 메타데이터**: AWS, GCP, Azure 등 클라우드 Provider별 상세 정보 저장

##### 인프라 계층 구조 예시

**On-Premise**
```
infrastructure:onprem/datacenter-01/rack-a/chassis-1
└── target: gpu-server-01 (NVIDIA A100 x8)

infrastructure:onprem/datacenter-01/rack-b/chassis-1
└── target: storage-server-01 (Lustre)
```

**AWS**
```
infrastructure:aws/us-east-1/us-east-1a
├── target: i-1234567890 (p3.8xlarge, V100 x4)
└── target: i-abcdef1234 (g4dn.xlarge, T4 x1)

infrastructure:aws/us-west-2/us-west-2a
└── target: i-fedcba0987 (p4d.24xlarge, A100 x8)
```

**GCP**
```
infrastructure:gcp/asia-northeast3/asia-northeast3-a
└── target: vm-gpu-worker-001 (n1-highmem-8 + T4)

infrastructure:gcp/us-central1/us-central1-a
└── target: vm-gpu-worker-002 (a2-highgpu-1g + A100)
```

**논리적 그룹**
```
logical:ml-training/gpu-workers
├── gpu-server-01 (onprem)
├── i-1234567890 (aws)
└── vm-gpu-worker-001 (gcp)

logical:inference/serving-cluster
├── i-abcdef1234 (aws)
└── vm-gpu-worker-002 (gcp)
```

**환경**
```
environment:production
├── gpu-server-01 (onprem)
├── i-1234567890 (aws)
└── vm-gpu-worker-001 (gcp)

environment:staging
├── i-fedcba0987 (aws)
└── vm-gpu-worker-002 (gcp)
```

##### 타겟의 다차원 분류

```
gpu-server-01 (On-Premise 물리 서버)
├── infrastructure:onprem/datacenter-01/rack-a/chassis-1 (primary)
├── logical:ml-training/gpu-workers
├── logical:gpu-cluster
└── environment:production

i-1234567890 (AWS EC2)
├── infrastructure:aws/us-east-1/us-east-1a (primary)
├── logical:ml-training/gpu-workers
└── environment:production

vm-gpu-worker-001 (GCP Compute Engine)
├── infrastructure:gcp/asia-northeast3/asia-northeast3-a (primary)
├── logical:inference/serving-cluster
└── environment:production
```

##### 정책 적용 우선순위

```
Namespace 우선순위:
environment (10) ← 최우선
  ↓
logical (50)
  ↓
infrastructure (100)

설정 병합 예시:
1. 타겟 개별 설정 (override)
2. environment 그룹 설정 (Alert: PagerDuty)
3. logical 그룹 설정 (Check: mount, merge)
4. infrastructure 그룹 설정 (Check: infiniband, nvme)
```

##### Provider 메타데이터

각 infrastructure 그룹은 Provider별 메타데이터를 포함:

**AWS**
```json
{
  "provider": "aws",
  "region": "us-east-1",
  "availability_zone": "us-east-1a",
  "instance_type": "p3.8xlarge",
  "gpus": [{"type": "Tesla V100", "count": 4, "memory_gb": 16}],
  "spot_instance": false
}
```

**GCP**
```json
{
  "provider": "gcp",
  "project_id": "my-ml-project",
  "zone": "asia-northeast3-a",
  "machine_type": "n1-highmem-8",
  "accelerators": [{"type": "nvidia-tesla-t4", "count": 1}],
  "preemptible": false
}
```

**On-Premise**
```json
{
  "provider": "onprem",
  "datacenter": "datacenter-01",
  "rack": "rack-a",
  "chassis": "chassis-1",
  "position": "U10-U12",
  "hardware": {
    "vendor": "Supermicro",
    "model": "SYS-420GP-TNR",
    "gpus": [{"type": "NVIDIA A100", "count": 8, "memory_gb": 80}]
  }
}
```

##### API 워크플로우

```bash
# 1. Namespace 조회
GET /api/v1/namespaces
→ [infrastructure, logical, environment]

# 2. Infrastructure 그룹 생성 (AWS)
POST /api/v1/groups
{
  "namespace": "infrastructure",
  "name": "us-east-1a",
  "full_path": "infrastructure:aws/us-east-1/us-east-1a",
  "provider_type": "aws",
  "provider_metadata": {...}
}

# 3. 타겟 등록 및 Primary 그룹 설정
POST /api/v1/targets
{"name": "i-1234567890", "address": "10.0.1.50", ...}

PUT /api/v1/targets/{target_id}/primary-group
{"group_id": "<us-east-1a-group-id>"}

# 4. 논리 그룹 추가
POST /api/v1/targets/{target_id}/groups
{"namespace": "logical", "group_id": "<ml-training-group-id>"}

# 5. 환경 그룹 추가
POST /api/v1/targets/{target_id}/groups
{"namespace": "environment", "group_id": "<production-group-id>"}

# 6. 그룹 체크 설정
POST /api/v1/groups/{group_id}/checks
{"check_type": "mount", "config": {...}, "merge_strategy": "merge"}

# 7. 최종 설정 확인 (Namespace 우선순위 적용)
GET /api/v1/targets/{target_id}/checks/effective

# 8. 정책 추적 (디버깅)
GET /api/v1/targets/{target_id}/policies/trace
```

상세 가이드는 `docs/namespace-group-management.md` 참조

#### Config Server UI (선택)

간단한 웹 UI를 통해 Config Server API를 보다 쉽게 사용할 수 있습니다.

##### 주요 화면
- **대시보드**: 전체 모니터링 상태 요약
- **그룹 관리**: 계층적 그룹 트리 뷰, 그룹별 체크 설정
- **타겟 관리**: 서버/디바이스 추가, 수정, 삭제, 그룹 할당
- **체크 설정**: 서버별/그룹별 커스텀 체크 설정
- **알림 설정**: 알림 채널 및 라우팅 규칙 설정

##### 기술 스택
- **Next.js 15**: React 프레임워크 (Static Export)
- **TypeScript**: 타입 안정성
- **shadcn/ui**: 경량 UI 컴포넌트
- **Tailwind CSS**: 스타일링
- **nginx**: 정적 파일 서빙

##### 폐쇄망 배포
- `next build && next export`로 정적 HTML/CSS/JS 생성
- nginx Docker 이미지에 정적 파일 포함
- 런타임 의존성 없음 (모든 API 호출은 클라이언트 사이드)

상세 가이드는 `services/config-server-ui/README.md` 참조

### 2. 커스텀 체크 시스템

#### 개요
사용자 정의 스크립트를 통해 벤더 도구가 제공하지 않는 메트릭을 수집합니다.

#### 체크 유형
- **마운트 포인트**: NFS, CIFS, Lustre, GPFS 등 스토리지 마운트 상태 및 접근성
- **고속 스토리지**: NVMe 디바이스 상태, 병렬 파일시스템 응답 시간, I/O 대역폭 체크
- **고속 네트워크**: InfiniBand/RoCE 링크 상태, 대역폭 테스트, 스위치 연결성
- **디스크 헬스**: SMART 상태, RAID 컨트롤러 상태
- **가속기 헬스**: all-smi/DCGM 외 추가 체크 (온도 임계값, 프로세스 수 등)
- **네트워크 연결성**: 특정 서비스 포트 체크
- **라이선스 상태**: 소프트웨어 라이선스 만료 체크

#### 동적 체크 에이전트
```bash
# /usr/local/bin/dynamic-check.sh (cron 1분마다 실행)
# 1. Config Server에서 체크 설정 가져오기
CONFIG=$(curl -s http://config-server:8000/api/v1/checks/$HOSTNAME)

# 2. 설정에 따라 체크 수행
#    - 마운트 포인트 확인
#    - 디스크 SMART 상태
#    - 기타 커스텀 스크립트 실행

# 3. 결과를 Node Exporter textfile collector로 노출
echo "mount_check{path=\"/data\"} 1" > /var/lib/node_exporter/mount.prom
```

#### textfile Collector
Node Exporter의 textfile collector를 통해 커스텀 메트릭 노출:
- 위치: `/var/lib/node_exporter/textfile_collector/*.prom`
- 형식: Prometheus exposition format
- 자동 수집: Node Exporter가 주기적으로 파일 읽기

상세 가이드는 `config/node-exporter/textfile-collector/README.md` 참조

### 3. AI 가속기 모니터링 (하이브리드 접근)

#### 모니터링 전략
- **all-smi**: 다양한 AI 가속기(NVIDIA, AMD, Intel Gaudi, Rebellions, Furiosa, Tenstorrent, Google TPU 등)를 단일 도구로 통합 모니터링하는 오픈소스 도구. Prometheus API 내장
- **DCGM Exporter** (선택): NVIDIA GPU 심화 메트릭 수집 (ECC 오류, NVLink 등)
- **커스텀 Exporter** (필요시): all-smi 미지원 가속기 추가

#### 구현 방향
1. 각 서버에 all-smi API 모드로 배포 (포트 9400)
2. Prometheus에서 all-smi 메트릭 스크랩
3. NVIDIA GPU 서버는 DCGM Exporter 추가 배포 (선택)
4. 통합 Grafana 대시보드 및 알림 규칙 작성

상세 구현 가이드는 `services/exporters/all-smi/README.md` 참조

### 4. 알림 시스템 (Alertmanager)

#### 개요
Prometheus 알림 규칙 평가 → Alertmanager 라우팅 → 다양한 채널로 알림 전송

#### 지원 알림 채널
- **Email (SMTP)**: 일반 알림, 보고서 전송
- **Slack**: 팀 채널별 실시간 알림
- **Webhook**: 사용자 정의 시스템 통합
- **PagerDuty**: 온콜 관리, 에스컬레이션
- **OpsGenie**: 인시던트 관리, 자동 티켓 생성
- **Microsoft Teams**: 기업 협업 도구 통합

#### 고급 라우팅 기능

##### 1. 심각도별 라우팅
```
critical   → PagerDuty (온콜 즉시 호출)
warning    → Slack (팀 채널)
info       → Email (일일 요약)
```

##### 2. 팀별 라우팅
레이블 기반 팀 분리:
```
team=infra     → #infra-alerts
team=ml        → #ml-ops-alerts
team=platform  → #platform-team
```

##### 3. 시간대별 라우팅
```
업무 시간 (9-18시)  → Slack
야간/주말          → PagerDuty (긴급만)
```

##### 4. 알림 그룹화 및 억제
- **그룹화**: 동일 서버의 여러 알림을 하나로 묶음
- **억제**: 상위 알림 발생 시 하위 알림 억제
  - 예: 노드 다운(`up{job="node-exporter"} == 0`) 시 해당 노드의 모든 디바이스/리소스 알림 자동 억제
  - 다수 노드 동시 다운 시 인프라 전체 장애로 에스컬레이션
- **음소거**: 유지보수 기간 동안 특정 알림 일시 중단

#### 알림 템플릿
커스텀 메시지 템플릿:
- 이메일: HTML 포맷, 그래프 포함
- Slack: 컬러 코드, 버튼 액션
- PagerDuty: 심각도별 자동 에스컬레이션

#### 알림 규칙 그룹별 커스터마이징
그룹 및 Namespace별로 알림 규칙 threshold를 다르게 설정 가능:

**기능:**
- **규칙 템플릿**: 기본 알림 규칙 정의 (CPU 80%, 메모리 90% 등)
- **그룹별 커스터마이징**: 각 그룹마다 다른 threshold 설정
  - AWS: CPU 90% (클라우드 버스팅 고려)
  - On-Premise: CPU 75% (더 엄격)
  - Production: CPU 70% (최우선, environment namespace)
- **Namespace 우선순위**: environment (10) > logical (50) > infrastructure (100)
- **정책 병합**: override, merge 전략으로 설정 상속
- **타겟 오버라이드**: 특정 서버는 개별 threshold 설정
- **정책 추적 API**: 어느 그룹에서 threshold가 적용되었는지 디버깅

**예시:**
```
infrastructure:aws/us-east-1        → CPU 90%, Memory 95%
logical:ml-training                 → CPU 95% (merge)
environment:production              → CPU 70% (override, 최우선)

→ 최종: CPU 70% (production이 최우선 적용)
```

**API:**
```bash
# 그룹별 알림 규칙 설정
POST /api/v1/groups/{id}/alert-rules
{
  "rule_template_id": "HighCPUUsage",
  "config": {"cpu_percent": 90},
  "merge_strategy": "override"
}

# 최종 규칙 확인 (정책 병합 적용)
GET /api/v1/targets/{id}/alert-rules/effective

# 정책 추적 (디버깅)
GET /api/v1/targets/{id}/alert-rules/trace
```

상세 설정은 `config/alertmanager/` 및 `docs/alert-rules-customization.md` 참조

#### Bootstrap Script & Auto Registration

새 서버 추가 시 한 줄 명령으로 자동 등록 및 Exporter 설치:

**기능:**
- **자동 등록**: 서버 하드웨어 정보 자동 수집 및 Config Server 등록
- **하드웨어 감지**: GPU/NPU 자동 감지 및 적절한 Exporter 설치
- **토큰 기반 인증**: Bootstrap 토큰으로 안전한 등록
- **설치 자동화**: Node Exporter, DCGM Exporter 등 자동 설치
- **설정 자동 적용**: Config Server에서 그룹 설정 가져와 적용
- **멀티 환경 지원**: 온프레미스, 클라우드, Kubernetes 모두 지원

**사용 예시:**
```bash
# 새 GPU 서버에서 실행
curl -fsSL https://config.example.com/bootstrap.sh | \
  bash -s -- --token=SECRET_TOKEN

# 실행 과정:
# 1. 서버 정보 수집 (hostname, IP, GPU 타입, 메모리 등)
# 2. Config Server에 자동 등록
# 3. 설치할 exporter 목록 조회
# 4. Node Exporter 설치
# 5. DCGM Exporter 설치 (GPU 감지 시)
# 6. Systemd 서비스 등록 및 시작
# 7. 설치 완료 보고
```

**API:**
```bash
# Bootstrap 토큰 생성 (Config Server UI/API)
POST /api/v1/bootstrap/token
{
  "name": "aws-gpu-cluster",
  "expires_at": "2025-12-31T23:59:59Z",
  "group_id": "logical:ml-training"
}

# 새 서버 자동 등록 (Bootstrap 스크립트가 호출)
POST /api/v1/bootstrap/register
{
  "token": "secret-bootstrap-token",
  "hostname": "gpu-node-05.example.com",
  "ip_address": "10.0.1.15",
  "hardware": {
    "gpu_count": 8,
    "gpu_model": "A100",
    "cpu_cores": 128,
    "memory_gb": 1024
  }
}

# 응답: 설치할 exporter 목록
{
  "target_id": "uuid-xxx",
  "exporters_to_install": [
    {"type": "node_exporter", "version": "1.6.1", "port": 9100},
    {"type": "dcgm_exporter", "version": "3.1.0", "port": 9400}
  ]
}

# Bootstrap 완료 보고
POST /api/v1/targets/{id}/bootstrap/complete
{
  "installed_exporters": ["node_exporter", "dcgm_exporter"],
  "status": "success"
}
```

**Cloud-init 통합** (클라우드 환경):
```yaml
# Terraform user_data
user_data = <<-EOF
  #!/bin/bash
  curl -fsSL https://config.example.com/bootstrap.sh | \
    bash -s -- --token=SECRET_TOKEN
EOF
```

#### SSH Agent & Remote Deployment

Config Server에서 SSH를 통해 타겟 서버에 원격으로 Exporter 배포:

**기능:**
- **원격 배포**: Config Server UI에서 버튼 클릭으로 원격 설치
- **SSH 키 관리**: 안전한 SSH 키 저장 및 관리
- **실시간 진행 상황**: WebSocket으로 설치 과정 실시간 스트리밍
- **재배포 지원**: 기존 서버에 Exporter 재설치
- **하드웨어 자동 감지**: SSH로 연결 후 GPU/NPU 자동 탐지
- **설치 검증**: 설치 후 health check 자동 수행

**UI 예시:**
```
Config Server UI:
┌──────────────────────────────────────────────────────┐
│ Remote Deploy to gpu-node-06.example.com            │
├──────────────────────────────────────────────────────┤
│ SSH Configuration:                                   │
│   User: ubuntu                                       │
│   SSH Key: [prod-ssh-key]                           │
│   Port: 22                                           │
│                                                      │
│ Exporters to Install:                                │
│   [✓] Node Exporter (auto-detected: Ubuntu 22.04)  │
│   [✓] DCGM Exporter (auto-detected: 8x A100 GPUs)  │
│                                                      │
│ Deployment Progress:                                 │
│   [████████████████░░░░] 80% - Installing DCGM...   │
│                                                      │
│ [Cancel] [View Logs]                                │
└──────────────────────────────────────────────────────┘
```

**API:**
```bash
# SSH 키 등록
POST /api/v1/ssh-keys
{
  "name": "prod-ssh-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----...",
  "passphrase": "optional"
}

# 원격 배포 트리거
POST /api/v1/targets/{id}/deploy
{
  "ssh_config": {
    "user": "ubuntu",
    "key_id": "ssh-key-uuid",
    "port": 22
  },
  "exporters": ["node_exporter", "dcgm_exporter"],
  "auto_start": true
}

# 배포 상태 조회 (WebSocket)
WS /api/v1/targets/{id}/deploy/status
→ {"stage": "connecting", "progress": 10}
→ {"stage": "detecting_hardware", "progress": 20}
→ {"stage": "installing_node_exporter", "progress": 50}
→ {"stage": "installing_dcgm_exporter", "progress": 80}
→ {"stage": "completed", "progress": 100}
```

**보안:**
- SSH 키는 암호화되어 PostgreSQL에 저장
- Vault 통합 지원 (선택)
- SSH 접근 로그 기록
- Role-based access control (RBAC)

#### Fleet Management (일괄 배포)

여러 서버에 동시에 Exporter 배포 및 관리:

**기능:**
- **그룹 단위 배포**: 특정 그룹의 모든 서버에 일괄 배포
- **병렬 실행**: 최대 N개 서버에 동시 배포 (concurrency limit)
- **롤링 업데이트**: 순차 배포로 서비스 중단 최소화
- **배포 전략**: all-at-once, rolling, canary
- **실패 처리**: 실패 시 자동 롤백 또는 계속 진행
- **배포 이력**: 모든 배포 작업 기록 및 추적
- **드라이런 모드**: 실제 배포 전 시뮬레이션

**사용 예시:**
```bash
# ml-training 그룹의 모든 서버에 DCGM Exporter 업그레이드
POST /api/v1/fleet/deploy
{
  "target_group_id": "logical:ml-training",
  "action": "upgrade",
  "exporter": "dcgm_exporter",
  "version": "3.2.0",
  "strategy": "rolling",
  "concurrency": 5,
  "rollback_on_failure": true,
  "dry_run": false
}

# 응답
{
  "job_id": "deploy-job-uuid",
  "targets_count": 50,
  "status": "in_progress",
  "estimated_duration": "10m"
}

# 배포 작업 상태 조회
GET /api/v1/fleet/deploy/{job_id}
{
  "job_id": "deploy-job-uuid",
  "status": "in_progress",
  "progress": {
    "total": 50,
    "completed": 30,
    "failed": 2,
    "in_progress": 5,
    "pending": 13
  },
  "results": [
    {"target": "gpu-01", "status": "success", "duration": "45s"},
    {"target": "gpu-02", "status": "success", "duration": "42s"},
    {"target": "gpu-03", "status": "failed", "error": "SSH timeout"},
    ...
  ]
}

# 배포 작업 취소
POST /api/v1/fleet/deploy/{job_id}/cancel

# 배포 이력 조회
GET /api/v1/fleet/jobs?limit=20&offset=0
```

**배포 전략:**
- **all-at-once**: 모든 서버에 동시 배포 (빠름, 위험)
- **rolling**: 순차 배포 (안전, 느림)
  - `concurrency: 5` → 5대씩 배포
  - 실패 시 자동 중단
- **canary**: 소수 서버에 먼저 배포 후 검증
  - 10% 배포 → 검증 → 나머지 배포

**UI 예시:**
```
Fleet Management Dashboard:
┌────────────────────────────────────────────────────┐
│ Deploy to: logical:ml-training (50 servers)        │
├────────────────────────────────────────────────────┤
│ Exporter: DCGM Exporter → 3.2.0                   │
│ Strategy: Rolling (5 concurrent)                   │
│ Rollback on failure: [✓]                          │
│                                                    │
│ Progress: [████████████░░░░] 60% (30/50)          │
│   ✓ Completed: 30                                 │
│   ✗ Failed: 2                                     │
│   ⟳ In Progress: 5                                │
│   ⋯ Pending: 13                                   │
│                                                    │
│ [Cancel Deployment] [View Failed Servers]         │
└────────────────────────────────────────────────────┘
```

상세 가이드: `docs/bootstrap-and-deployment.md`, `docs/fleet-management.md`

---

## 📝 추가 필요 문서

### 1. 운영 가이드
- **백업 및 복구**: Prometheus 데이터, 설정 백업 방법
- **스케일링**: 서버 증가 시 대응 방법
- **성능 튜닝**: 대규모 환경에서의 최적화
- **보안**: 인증, 암호화, 접근 제어
- **폐쇄망 운영**: 오프라인 업데이트, 로컬 저장소 유지보수

### 2. 개발 가이드
- **Config Server 개발** (Go): REST API 서버 개발
  - Gin/Fiber 프레임워크 사용
  - PostgreSQL 연동 (pgx)
  - Redis 캐싱 (go-redis)
  - API 엔드포인트 구현
- **Config Server UI 개발** (선택, Next.js): 관리 웹 UI 개발
  - Next.js 15 App Router 사용
  - Static Export 설정 (`output: 'export'`)
  - shadcn/ui 컴포넌트 활용
  - API 클라이언트 구현
  - 정적 빌드 및 nginx 배포
- **커스텀 Exporter 개발** (Go, 필요시): all-smi가 지원하지 않는 가속기 추가
  - prometheus/client_golang 활용
  - Exporter 템플릿 활용
  - 메트릭 네이밍 컨벤션
  - 벤더 SDK/CLI 통합 방법
- **all-smi 통합**: all-smi를 사용한 가속기 모니터링 설정
  - API 모드 설정 및 Prometheus 통합
  - 벤더 드라이버/SDK 설치 가이드
  - 메트릭 수집 검증 방법
- **체크 스크립트 개발** (Python/Bash): 커스텀 체크 로직 구현
  - textfile collector 사용법
  - 동적 체크 스크립트 패턴
- **대시보드 개발**: Grafana 대시보드 작성 가이드
- **알림 규칙 작성**: PromQL을 사용한 통합 알림 규칙 작성

#### 코드 주석 규칙

**IMPORTANT: All code comments must be written in English**

- **Go 코드**: 모든 주석은 영어로 작성
  ```go
  // Good: English comment
  // CreateTarget creates a new monitoring target
  func CreateTarget(target *Target) error {
      // Validate target configuration
      if err := target.Validate(); err != nil {
          return err
      }
      return db.Insert(target)
  }

  // Bad: Korean comment (DO NOT USE)
  // CreateTarget는 새로운 모니터링 타겟을 생성합니다
  func CreateTarget(target *Target) error {
      // 타겟 설정 검증
      if err := target.Validate(); err != nil {
          return err
      }
      return db.Insert(target)
  }
  ```

- **TypeScript/JavaScript 코드**: 모든 주석은 영어로 작성
  ```typescript
  // Good: English comment
  // Fetch all targets from the API
  async function fetchTargets(): Promise<Target[]> {
      const response = await fetch('/api/v1/targets')
      return response.json()
  }

  // Bad: Korean comment (DO NOT USE)
  // API에서 모든 타겟을 가져옵니다
  async function fetchTargets(): Promise<Target[]> {
      const response = await fetch('/api/v1/targets')
      return response.json()
  }
  ```

- **Bash/Shell 스크립트**: 모든 주석은 영어로 작성
  ```bash
  # Good: English comment
  # Install Node Exporter on the target server
  install_node_exporter() {
      # Download the latest version
      wget https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz
  }

  # Bad: Korean comment (DO NOT USE)
  # 타겟 서버에 Node Exporter를 설치합니다
  install_node_exporter() {
      # 최신 버전 다운로드
      wget https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz
  }
  ```

- **설정 파일 (YAML, JSON 등)**: 모든 주석은 영어로 작성
  ```yaml
  # Good: English comment
  # Prometheus scrape configuration
  scrape_configs:
    - job_name: 'node-exporter'
      # Scrape interval: 15 seconds
      scrape_interval: 15s

  # Bad: Korean comment (DO NOT USE)
  # Prometheus 스크랩 설정
  scrape_configs:
    - job_name: 'node-exporter'
      # 스크랩 간격: 15초
      scrape_interval: 15s
  ```

**이유:**
- 국제 협업 용이성
- 오픈소스 프로젝트로 전환 가능성
- 코드 가독성 및 유지보수성 향상
- 한글 인코딩 문제 회피

### 3. API 문서
- **Config Server REST API 명세**: OpenAPI/Swagger 문서
- **인증 방법**: API 키, OAuth 등
- **Rate Limiting**: API 사용 제한

### 4. FAQ 및 트러블슈팅
- **자주 묻는 질문**: 일반적인 질문과 답변
- **알려진 이슈**: 현재 제한사항 및 해결 방법
- **디버깅 가이드**: 문제 진단 방법

---

## 🎯 마일스톤

### Phase 1: 기본 인프라 (2주)
- [ ] Docker Compose 기반 기본 스택 구성
- [ ] Prometheus, Grafana, Alertmanager 설정
- [ ] Node Exporter, DCGM Exporter 통합
- [ ] 기본 대시보드 및 알림 규칙 작성

### Phase 2: 동적 설정 시스템 (2주)
- [ ] Config Server API 개발 (Go)
- [ ] HTTP SD 또는 File SD 구현
- [ ] 에이전트 스크립트 개발 및 배포
- [ ] 설정 변경 자동화
- [ ] Bootstrap Script & Auto Registration 개발
  - [ ] Bootstrap 토큰 관리 API
  - [ ] 서버 자동 등록 API
  - [ ] 하드웨어 자동 감지 로직
  - [ ] Bootstrap 스크립트 작성 (bash)
  - [ ] Cloud-init 템플릿 제공
- [ ] Config Server UI 개발 (선택, +3일)
  - [ ] Next.js 프로젝트 설정 (Static Export)
  - [ ] 타겟/체크/알림 관리 화면
  - [ ] Bootstrap 토큰 관리 UI
  - [ ] 정적 빌드 및 nginx Docker 이미지

### Phase 3: all-smi 통합 및 고속 인프라 모니터링 (2주)
- [ ] all-smi 설치 및 설정 자동화 (Ansible role 작성)
- [ ] all-smi API 모드 Prometheus 통합
- [ ] DCGM Exporter 선택적 배포 설정
- [ ] 각 가속기별 벤더 드라이버/SDK 설치 가이드 작성
- [ ] all-smi 메트릭 기반 통합 대시보드 작성
- [ ] 통합 알림 규칙 작성 (all-smi 메트릭 기반)
- [ ] 고속 네트워크 체크 구현 (InfiniBand/RoCE)
- [ ] 고속 스토리지 체크 구현 (NVMe, 병렬 FS)
- [ ] 마운트 포인트 체크 구현
- [ ] all-smi가 지원하지 않는 가속기용 커스텀 Exporter 프레임워크 (필요시)
- [ ] all-smi 사용 가이드 문서화

### Phase 4: 폐쇄망 지원 및 프로덕션 준비 (1.5주)
- [ ] 오프라인 패키지 번들 생성 스크립트 작성
- [ ] 로컬 Docker 레지스트리 설정 자동화
- [ ] 로컬 APT/YUM 저장소 구축 스크립트
- [ ] 의존성 사전 다운로드 도구 개발
- [ ] 오프라인 설치 스크립트 작성
- [ ] Air-gapped 환경 테스트
- [ ] Kubernetes 배포 매니페스트 작성
- [ ] Ansible Playbook 작성 (온라인/오프라인 모드)
- [ ] SSH Agent & Remote Deployment 개발
  - [ ] SSH 키 관리 API 및 암호화 저장
  - [ ] 원격 배포 엔진 개발 (SSH 연결, 스크립트 실행)
  - [ ] WebSocket 기반 실시간 배포 상태 스트리밍
  - [ ] 하드웨어 자동 감지 (SSH 원격)
  - [ ] 배포 실패 처리 및 롤백
  - [ ] UI 통합 (원격 배포 화면)
- [ ] 백업/복구 스크립트
- [ ] 보안 강화 (TLS, 인증)

### Phase 5: 알림 규칙 그룹별 커스터마이징 (1주)
- [ ] 알림 규칙 템플릿 DB 스키마 설계
- [ ] 그룹별 알림 규칙 커스터마이징 기능 개발
- [ ] Namespace 우선순위 기반 정책 병합 로직 구현
- [ ] 알림 규칙 API 개발 (템플릿, 그룹 규칙, 타겟 오버라이드)
- [ ] Prometheus 규칙 파일 자동 생성기 구현
- [ ] 정책 추적 API 개발 (디버깅용)
- [ ] 그룹별 threshold 설정 테스트 (AWS 90%, On-Premise 75% 등)
- [ ] Config Server UI에 알림 규칙 관리 화면 추가 (선택)

### Phase 6: Fleet Management (1주)
- [ ] Fleet Management 데이터 모델 설계
  - [ ] 배포 작업(Job) 테이블
  - [ ] 배포 작업 상세(Job Details) 테이블
  - [ ] 배포 이력 저장소
- [ ] Fleet Deployment 엔진 개발
  - [ ] 그룹 단위 배포 로직
  - [ ] 병렬 실행 제어 (concurrency limit)
  - [ ] 배포 전략 구현 (all-at-once, rolling, canary)
  - [ ] 실패 처리 및 롤백 로직
- [ ] Fleet Management API 개발
  - [ ] 배포 작업 생성/조회/취소 API
  - [ ] WebSocket 기반 실시간 진행 상황 스트리밍
  - [ ] 배포 이력 조회 API
- [ ] Fleet Management UI 개발
  - [ ] 그룹 선택 및 배포 설정 화면
  - [ ] 실시간 진행 상황 대시보드
  - [ ] 배포 이력 및 통계 화면
- [ ] 드라이런 모드 및 배포 검증
- [ ] 대규모 배포 테스트 (50+ 서버)

### Phase 7: 문서화 및 테스트 (1주)
- [ ] 전체 문서 작성
- [ ] 통합 테스트 작성
- [ ] 예제 및 튜토리얼 작성
- [ ] 사용자 가이드 작성

---

## 👥 팀 역할

### 필요한 역할
- **DevOps 엔지니어**: 인프라 구축, 자동화
- **백엔드 개발자**: Config Server API 개발
- **SRE**: 모니터링 규칙, 대시보드 설계
- **문서 작성자**: 사용자 가이드, API 문서 작성

### 기여 방법
- Pull Request를 통한 코드 기여
- Issue를 통한 버그 리포트 및 기능 제안
- 문서 개선 및 번역

---

## 🔧 Exporter 통합 개요

### AI 가속기 모니터링

#### all-smi 기반 통합
- **도구**: [all-smi](https://github.com/lablup/all-smi) - 범용 AI 가속기 모니터링 오픈소스
- **배포 방식**: 각 서버에 all-smi를 API 모드로 실행 (포트 9400)
- **지원 가속기**: NVIDIA, AMD, Intel Gaudi, Rebellions, Furiosa, Tenstorrent, Google TPU
- **전제 조건**: 각 벤더의 드라이버/SDK 설치 필요

#### 하이브리드 접근
- **기본**: all-smi로 통합 모니터링
- **선택**: NVIDIA GPU 환경에서 DCGM Exporter 추가 (심화 메트릭)
- **확장**: all-smi 미지원 가속기는 커스텀 Exporter 개발

#### 구현 위치
상세 설치, 설정, 배포 가이드는 다음 문서 참조:
- `services/exporters/all-smi/README.md` - all-smi 통합 가이드
- `services/exporters/vendor-drivers/*.md` - 벤더별 드라이버 설치
- `deploy/ansible/roles/all-smi/` - Ansible 자동화
- `deploy/kubernetes/all-smi/` - Kubernetes 배포
- `examples/all-smi-configs/` - 설정 예제

### 기타 Exporter

#### Node Exporter
시스템 메트릭(CPU, 메모리, 디스크, 네트워크) 수집

#### 커스텀 체크
사용자 정의 스크립트를 통한 마운트 포인트, 디스크 헬스 등 체크

---

## 🔒 폐쇄망 환경 지원

### 개요
인터넷 접속이 제한된 폐쇄망(Air-gapped) 환경에서도 쉽게 설치하고 운영할 수 있는 오프라인 배포 방식을 제공합니다.

### 오프라인 배포 전략

#### 1. 패키지 번들 생성 (인터넷 접속 가능 환경)
```bash
# 모든 의존성 패키지 다운로드 및 번들 생성
cd deploy/offline
./create-bundle.sh

# 생성되는 번들 내용:
# - Go 컴파일 바이너리 (config-server, custom-exporter)
# - Config Server UI 정적 빌드 (선택)
# - Docker 이미지 (tar 파일)
# - all-smi 및 의존 패키지 (deb/rpm)
# - 벤더 드라이버/SDK
# - 스크립트용 Python wheel (선택)
# - Grafana 대시보드
# - Prometheus 알림 규칙
```

#### 2. 번들 전송
```
USB 드라이브 또는 내부 네트워크를 통해
패키지 번들을 폐쇄망 환경으로 전송
```

#### 3. 로컬 인프라 구축
```bash
# 로컬 Docker 레지스트리 설정
./setup-local-registry.sh

# 로컬 APT/YUM 저장소 설정 (선택)
./setup-local-repo.sh
```

#### 4. 오프라인 설치 실행
```bash
# 번들에서 패키지 설치
./install-offline.sh

# 또는 Ansible을 통한 다중 서버 배포
ansible-playbook -i inventory/hosts playbooks/offline-install.yml
```

### 주요 특징

#### 단일 번들 패키지
- 모든 필수 컴포넌트를 하나의 번들로 제공
- 버전 호환성 보장
- 체크섬 검증 포함

#### 로컬 레지스트리 지원
- Docker 이미지 로컬 레지스트리 자동 구축
- Private registry 설정 자동화
- TLS 인증서 생성 (선택)

#### 최소 외부 의존성
- Go 정적 바이너리 (런타임 의존성 없음)
- 스크립트용 Python wheel (선택적)
- 시스템 패키지 사전 다운로드
- 오프라인 문서 포함 (HTML)

#### 점진적 업데이트
- 델타 업데이트 지원
- 변경된 컴포넌트만 전송
- 롤백 기능

### 폐쇄망 설치 체크리스트

- [ ] 인터넷 접속 환경에서 번들 생성
- [ ] 번들 무결성 검증 (체크섬)
- [ ] 폐쇄망 환경으로 번들 전송
- [ ] 로컬 Docker 레지스트리 구축
- [ ] 로컬 패키지 저장소 구축 (선택)
- [ ] 중앙 모니터링 서버 설치
- [ ] 각 서버에 Exporter 설치
- [ ] 연결성 테스트 및 검증

상세 가이드는 `deploy/offline/README.md` 및 `examples/offline-deployment/` 참조

---

## 📊 성공 지표

### 기술적 지표
- **메트릭 수집 성공률**: > 99.9%
- **알림 지연 시간**: < 1분
- **대시보드 로딩 시간**: < 2초
- **API 응답 시간**: < 100ms

### 운영 지표
- **평균 장애 감지 시간 (MTTD)**: < 1분
- **평균 복구 시간 (MTTR)**: < 5분
- **거짓 양성 알림율**: < 5%

---

## 🔒 보안 고려사항

### 인증 및 권한
- Grafana: LDAP/OAuth 통합
- Prometheus: Basic Auth 또는 OAuth2 Proxy
- Config Server: API 키 기반 인증

### 네트워크 보안
- 내부 통신: TLS 암호화
- 외부 노출: Reverse Proxy (Nginx, Traefik)
- 방화벽: 필요한 포트만 개방

### 데이터 보호
- 민감한 설정: Kubernetes Secrets 또는 Vault 사용
- 메트릭 데이터: 보존 기간 설정 (예: 30일)
- 백업: 암호화된 백업 스토리지

---

## 📐 메트릭 수집 개요

### 가속기 메트릭

#### all-smi 통합 메트릭
```promql
# 공통 메트릭 (모든 가속기)
allsmi_device_utilization{device="0", type="nvidia|gaudi|rebellions|..."}
allsmi_device_temperature{device="0", type="..."}
allsmi_memory_used_bytes{device="0", type="..."}
allsmi_power_usage_watts{device="0", type="..."}
```

#### 벤더별 심화 메트릭
- **NVIDIA (DCGM)**: ECC 오류, NVLink 대역폭, Tensor Core 사용률
- **Intel Gaudi**: HBM 메모리, NIC 대역폭, AIP 사용률
- **Furiosa**: PE(Processing Element) 상태, PCIe AER 오류
- **Tenstorrent**: NOC 사용률, Ethernet 대역폭

상세 메트릭 명세는 `config/prometheus/rules/` 및 `config/grafana/dashboards/` 참조

### 시스템 메트릭
Node Exporter를 통한 CPU, 메모리, 디스크, 네트워크 수집

### 커스텀 메트릭
textfile collector를 통한 마운트 포인트, SMART 상태 등 수집

---

## 🎛️ 알림 규칙 개요

### 알림 계층 구조

#### 1. 범용 가속기 알림
모든 가속기 타입에 적용되는 공통 알림:
- 고온 경고 (온도 > 85°C)
- 고사용률 알림 (사용률 > 95%)
- 메모리 부족 경고 (메모리 > 90%)
- 디바이스 오프라인 알림

#### 2. 벤더별 특화 알림
- **NVIDIA**: XID 오류, ECC 오류, NVLink 장애
- **Intel Gaudi**: AIP 오류, NIC 대역폭 저하
- **Furiosa**: PE 다운, PCIe AER 오류
- **Tenstorrent**: NOC 장애, Ethernet 링크 다운

#### 3. 시스템 및 인프라 알림
- **노드 다운/연결 단절** (up 메트릭 기반 탐지)
- **다수 노드 동시 다운** (인프라 전체 장애 감지)
- 디스크 공간 부족
- 마운트 포인트 오류
- 높은 CPU/메모리 사용률
- **InfiniBand/RoCE 링크 다운**
- **NVMe 디바이스 오류**
- **병렬 파일시스템 응답 지연**

상세 알림 규칙은 `config/prometheus/rules/*.yml` 참조

---

## 📚 참고 자료

### 공식 문서
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [all-smi GitHub](https://github.com/lablup/all-smi) - 통합 AI 가속기 모니터링
- [DCGM Exporter](https://github.com/NVIDIA/dcgm-exporter)
- [Node Exporter](https://github.com/prometheus/node_exporter)

### AI 가속기 벤더 문서
- [Intel Gaudi Documentation](https://docs.habana.ai/)
- [Rebellions Documentation](https://www.rebellions.ai/) (벤더 제공 문서 참조)
- [Furiosa AI Documentation](https://furiosa-ai.github.io/docs/)
- [Tenstorrent Documentation](https://docs.tenstorrent.com/)
- [AMD ROCm Documentation](https://rocmdocs.amd.com/)

### 모범 사례
- [Google SRE Book - Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/best-practices/)
- [Writing Exporters - Prometheus](https://prometheus.io/docs/instrumenting/writing_exporters/)

---

## 🤝 라이선스 및 기여

### 라이선스
- MIT License (또는 프로젝트에 맞는 라이선스 선택)

### 기여 가이드라인
- 코드 스타일:
  - Go: Effective Go, golangci-lint 규칙
  - TypeScript/React (UI): Biome 또는 ESLint + Prettier
  - Python (스크립트): PEP 8
  - Bash: Google Style Guide
- Commit 메시지: Conventional Commits 형식
- PR 프로세스: 리뷰 필수, CI 통과 후 머지

---

## 📞 연락처 및 지원

### 프로젝트 관리자
- 이름: [담당자 이름]
- 이메일: [이메일 주소]
- Slack: [Slack 채널]

### 이슈 리포팅
- GitHub Issues: [저장소 URL]/issues
- 긴급 문제: [온콜 연락처]

---

## 🔄 버전 관리

### Semantic Versioning
- **MAJOR**: 하위 호환성 없는 변경
- **MINOR**: 하위 호환성 있는 기능 추가
- **PATCH**: 버그 수정

### 변경 로그
- CHANGELOG.md에 모든 변경사항 기록
- Release Notes 자동 생성

---

## 부록: 기술 스택 상세

### 모니터링 스택
| 컴포넌트 | 버전 | 목적 |
|---------|------|-----|
| Prometheus | 2.45+ | 메트릭 수집 및 저장 |
| Grafana | 10.0+ | 시각화 |
| Alertmanager | 0.26+ | 알림 관리 |
| Node Exporter | 1.6+ | 시스템 메트릭 |
| **all-smi** | **Latest** | **통합 AI 가속기 모니터링 (주력)** |
| DCGM Exporter | 3.1+ | NVIDIA GPU 심화 메트릭 (선택) |

### Config Server 스택
| 컴포넌트 | 버전 | 목적 |
|---------|------|-----|
| Go | 1.21+ | 서비스 개발 언어 |
| Gin/Fiber | Latest | REST API 프레임워크 |
| PostgreSQL | 15+ | 설정 데이터 저장 |
| pgx | 5+ | PostgreSQL 드라이버 |
| Redis | 7+ | 캐싱 |
| go-redis | 9+ | Redis 클라이언트 |

### Config Server UI 스택 (선택)
| 컴포넌트 | 버전 | 목적 |
|---------|------|-----|
| Next.js | 15+ | React 프레임워크 (Static Export) |
| TypeScript | 5+ | 타입 안정성 |
| shadcn/ui | Latest | UI 컴포넌트 라이브러리 |
| Tailwind CSS | 4+ | 스타일링 |
| nginx | 1.25+ | 정적 파일 서빙 |

### 인프라 도구
| 도구 | 버전 | 목적 |
|------|------|-----|
| Docker | 24+ | 컨테이너화 |
| Docker Compose | 2.20+ | 로컬/개발 환경 |
| Kubernetes | 1.27+ | 프로덕션 오케스트레이션 |
| Ansible | 2.15+ | 자동화 배포 |
