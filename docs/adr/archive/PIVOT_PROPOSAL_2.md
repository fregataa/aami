# AAMI (AI Accelerator Monitoring Infrastructure) 기획서

**문서 버전**: 0.2.1
**최종 수정일**: 2026-01-02
**상태**: Superseded by ADR-001

> **Note**: 이 문서는 [ADR-001](../001-pivot-to-prometheus-wrapper.md)로 대체되었습니다.
> 상세 기획 및 기술 설계 참고용으로 보관합니다.
> 최신 정보는 [README.md](../../../README.md)를 참조하세요.

---

## 프로젝트 개요

### 한 줄 정의

> **K8s 없는 GPU 클러스터를 위한 올인원 모니터링 도구**
>
> Prometheus 스택의 설치, 설정, 운영을 단일 CLI/UI로 단순화하고, GPU 특화 진단 기능을 제공

### 배경 및 문제 정의

현재 K8s 없이 GPU 클러스터를 운영하는 환경에서 모니터링을 구축하려면:

| 작업 | 소요 시간 | 어려움 |
|------|----------|--------|
| Prometheus + Grafana + Alertmanager 설치 | 반나절 | Ansible playbook 직접 작성 |
| DCGM exporter 배포 | 2-3시간 | 각 노드에 개별 설치 |
| 알람 룰 작성 | 반나절 | PromQL 문법 학습 필요 |
| Slack/Email 알람 연동 | 2-3시간 | Alertmanager YAML 설정 |
| Air-gap 환경 대응 | 1-2일 | 오프라인 패키지 직접 구성 |
| **총 소요 시간** | **2-3일** | |

---

*[상세 내용은 원본 파일 참조]*

*이 문서의 핵심 결정 사항은 ADR-001에 정리되어 있습니다.*
