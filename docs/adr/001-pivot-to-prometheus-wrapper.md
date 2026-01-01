# ADR-001: Pivot to Prometheus Stack Wrapper for K8s-less GPU Clusters

## Status

**Accepted** (2026-01-02)

## Context

### 기존 방향

AAMI는 원래 PostgreSQL 기반의 마이크로서비스 아키텍처로, 다양한 AI 가속기(GPU, NPU, TPU)와 고속 네트워크(InfiniBand), 병렬 파일시스템(Lustre, GPFS)을 모니터링하는 범용 플랫폼을 목표로 했다.

### 문제점

1. **범위가 너무 넓음**: GPU, NPU, TPU, InfiniBand, Lustre 등 모든 것을 지원하려다 보니 핵심 가치가 불명확
2. **복잡한 아키텍처**: PostgreSQL, 마이크로서비스, Kubernetes 배포 등 설치 복잡도가 높음
3. **차별화 부족**: 기존 도구(kube-prometheus-stack, Zabbix 등) 대비 명확한 이점이 없음
4. **타겟 시장 불명확**: K8s 환경에서는 kube-prometheus-stack이 이미 표준

### 시장 기회

K8s 없이 GPU 클러스터를 운영하는 환경에서 모니터링 구축은 여전히 어려움:

| 작업 | 소요 시간 |
|------|----------|
| Prometheus + Grafana + Alertmanager 설치 | 반나절 |
| DCGM exporter 배포 | 2-3시간 |
| 알람 룰 작성 (PromQL 학습) | 반나절 |
| Slack/Email 연동 | 2-3시간 |
| Air-gap 환경 대응 | 1-2일 |
| **총 소요 시간** | **2-3일** |

## Decision

AAMI를 **"K8s 없는 GPU 클러스터를 위한 Prometheus 스택 설치/운영 래퍼"**로 피벗한다.

### 핵심 변경사항

| 영역 | 이전 | 이후 |
|------|------|------|
| **포지셔닝** | 범용 AI 인프라 모니터링 | K8s-less GPU 클러스터 전용 |
| **데이터 저장** | PostgreSQL | YAML 파일 (DB 없음) |
| **노드 관리** | Agent 기반 | SSH 기반 (Agentless) |
| **설치 방식** | K8s/Docker Compose | 단일 CLI (`aami init`) |
| **차별화** | 없음 | Xid 에러 해석, 30분 설치 |

### 주요 설계 결정

#### 1. SSH 기반 (Agentless) 채택

| 고려 사항 | SSH 기반 | Agent 기반 (Salt 등) |
|-----------|----------|---------------------|
| 초기 설치 부담 | 없음 | Agent 먼저 배포 필요 |
| Air-gap 친화성 | ✅ 좋음 | ⚠️ Agent 배포가 추가 문제 |
| 1k+ 노드 확장 | ⚠️ 병렬화로 해결 | ✅ 기본 지원 |
| 복잡도 | 낮음 | 높음 |

**결정**: Air-gap 환경에서 Agent 배포가 "닭과 달걀" 문제가 되므로 SSH 기반 채택.
1k 노드 병렬 SSH는 Go goroutine으로 충분히 해결 가능.

#### 2. 파일 기반 설정 (DB 없음)

| 고려 사항 | PostgreSQL | 파일 기반 (YAML) |
|-----------|------------|------------------|
| 설치 복잡도 | DB 추가 설치 필요 | 없음 |
| Air-gap | 추가 패키징 필요 | 깔끔함 |
| 백업 | pg_dump | 파일 복사 |
| Git 버전 관리 | 어려움 | ✅ 가능 |

**결정**: 설치 단순화와 Air-gap 용이성을 위해 파일 기반 채택.

#### 3. Xid 해석을 킬러 피처로

NVIDIA GPU의 Xid 에러 코드를 자동 해석하고 권장 조치를 제공하는 기능은:
- 기존 도구에 없는 기능
- GPU 운영자의 실제 pain point 해결
- 차별화 요소로 충분

Xid 이력은 별도 저장소 없이 Prometheus 메트릭(`DCGM_FI_DEV_XID_ERRORS`)을 활용.

## Consequences

### 장점

1. **명확한 타겟 시장**: K8s 없는 GPU 클러스터 운영자
2. **단순한 아키텍처**: DB 없음, Agent 없음, 단일 바이너리
3. **빠른 설치**: 2-3일 → 30분
4. **Air-gap 친화적**: 단일 번들로 오프라인 설치 가능
5. **차별화된 가치**: Xid 해석, GPU 특화 알람 프리셋

### 단점

1. **축소된 범위**: NPU, TPU, InfiniBand 등 비지원
2. **NVIDIA 종속**: AMD GPU 미지원 (Phase 3에서 검토)
3. **대규모 한계**: 1k+ 노드는 Federation 필요 (Phase 3)
4. **기존 코드 폐기**: PostgreSQL 기반 코드 사용 불가

### 마이그레이션

기존 PostgreSQL 기반 코드는 더 이상 사용하지 않음. 새로운 아키텍처로 처음부터 구현.

## References

- [PIVOT_PROPOSAL.md](./archive/PIVOT_PROPOSAL.md) - 초기 피벗 제안 (v0.1)
- [PIVOT_PROPOSAL_2.md](./archive/PIVOT_PROPOSAL_2.md) - 상세 기획서 (v0.2)
- [NVIDIA Xid Errors](https://docs.nvidia.com/deploy/xid-errors/)
- [Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/)
