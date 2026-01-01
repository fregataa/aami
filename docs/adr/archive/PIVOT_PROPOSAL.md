# AAMI 프로젝트 피봇 기획서

**문서 버전**: 1.0
**작성일**: 2026-01-01
**상태**: Superseded by ADR-001

> **Note**: 이 문서는 [ADR-001](../001-pivot-to-prometheus-wrapper.md)로 대체되었습니다.
> 의사결정 기록 목적으로 보관합니다.

---

## Executive Summary

AAMI(AI Accelerator Monitoring Infrastructure) 프로젝트의 방향을 재정립합니다.

**기존 방향**: 자체 모니터링 시스템 구축
**새로운 방향**: GPU 클러스터를 위한 Prometheus 스택 설치/관리 래퍼

핵심 가치는 **"GPU 클러스터용 Prometheus를 5분 만에 설치하고, 알람을 Web UI로 설정"** 입니다.

---

## 1. 배경: 왜 피봇하는가

### 1.1 기존 접근의 문제점

| 문제 | 설명 |
|------|------|
| 시장 포화 | Prometheus, Datadog, Zabbix 등 성숙한 도구가 이미 존재 |
| 설치 부담 | Config Server, PostgreSQL, Prometheus, Grafana 등 컴포넌트 다수 |
| 유지보수 | 혼자서 전체 모니터링 시스템을 장기 유지하기 어려움 |
| 채택 가능성 | Platform 팀 입장에서 "또 하나의 관리할 시스템" |

### 1.2 실제 Pain Point 재정의

현업에서 Prometheus + Grafana를 쓰고 싶어도 막히는 지점:

1. **설치가 번거롭다** — 컴포넌트별 개별 설치, 설정 파일 연결
2. **알람 설정이 어렵다** — YAML 문법 학습 필요, 실수 잦음
3. **GPU 모니터링 추가 작업** — DCGM exporter 별도 설치, 대시보드 직접 구성
4. **Air-gap 환경에서 더 힘듦** — 오프라인 패키지 직접 구성해야 함

### 1.3 새로운 포지셔닝

```
기존: "우리가 만든 모니터링 시스템을 써라"
     → 경쟁자: Datadog, Zabbix, Prometheus

신규: "Prometheus 스택을 GPU 클러스터에 쉽게 설치하게 해줄게"
     → 경쟁자: kube-prometheus-stack, 수동 설치
     → 차별화: GPU 특화 + Air-gap 지원 + 알람 Web UI
```

---

*[이하 원본 문서 내용 생략 - 전체 내용은 Git 히스토리에서 확인 가능]*
