# AAMI Web UI Specification

## Overview

AAMI Config Server 관리를 위한 웹 기반 사용자 인터페이스 명세서입니다.

---

## Technology Stack

| Category | Technology | Purpose |
|----------|------------|---------|
| Framework | Next.js 15 (App Router) | React 기반 프레임워크, Static Export |
| UI Components | shadcn/ui | 접근성 지원 컴포넌트 라이브러리 |
| Styling | Tailwind CSS | 유틸리티 기반 CSS |
| Data Fetching | SWR | 캐싱, 자동 갱신, 실시간 업데이트 |
| Forms | react-hook-form + zod | 폼 상태 관리 및 검증 |
| Code Editor | Monaco Editor | Script Template 편집용 |
| HTTP Client | fetch (native) | API 통신 |

---

## Page Structure

```
/                           → Dashboard (대시보드)
/targets                    → Target 목록
/targets/[id]               → Target 상세
/groups                     → Group 목록
/groups/[id]                → Group 상세
/exporters                  → Exporter 목록
/alerts/templates           → Alert Template 목록
/alerts/templates/[id]      → Alert Template 상세/편집
/alerts/rules               → Alert Rule 목록
/alerts/rules/[id]          → Alert Rule 상세/편집
/scripts/templates          → Script Template 목록
/scripts/templates/[id]     → Script Template 상세/편집 (코드 에디터)
/scripts/policies           → Script Policy 목록
/scripts/policies/[id]      → Script Policy 상세/편집
/bootstrap                  → Bootstrap Token 관리
/settings                   → 설정 (Prometheus 상태 등)
```

---

## Feature Specification by Page

### 1. Dashboard (`/`)

#### 기능
| Feature | Description | API |
|---------|-------------|-----|
| 시스템 상태 | Config Server, Prometheus 연결 상태 | `GET /health`, `GET /api/v1/prometheus/status` |
| 노드 요약 | 총 노드 수, 활성/비활성 비율 | `GET /api/v1/targets` |
| 그룹 요약 | 총 그룹 수 | `GET /api/v1/groups` |
| 활성 알림 | 현재 firing 상태인 알림 목록 | `GET /api/v1/alerts/active` **(NEW API)** |
| 외부 링크 | Grafana, Prometheus, Alertmanager 링크 | Config 기반 |

#### UI 컴포넌트
- Summary Cards (노드, 그룹, 알림 개수)
- Alert List (실시간 갱신, 10초 간격)
- Quick Links

---

### 2. Target Management (`/targets`)

#### 목록 페이지 (`/targets`)
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 페이지네이션, 정렬, 필터링 | `GET /api/v1/targets` |
| 검색 | hostname, IP로 검색 | `GET /api/v1/targets?search=xxx` **(NEW PARAM)** |
| 상태 필터 | active/inactive/all | `GET /api/v1/targets?status=xxx` **(NEW PARAM)** |
| 생성 | 수동 노드 등록 | `POST /api/v1/targets` |
| 삭제 | 소프트 삭제 | `POST /api/v1/targets/delete` |

#### 상세 페이지 (`/targets/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | hostname, IP, port, status, labels | `GET /api/v1/targets/:id` |
| 그룹 멤버십 | 속한 그룹 목록, 추가/제거 | `GET /api/v1/targets/:id` (groups 포함) |
| Exporter 목록 | 연결된 exporter 목록 | `GET /api/v1/exporters/target/:target_id` |
| 유효 Check | 이 노드에 적용되는 check 목록 | `GET /api/v1/checks/target/:targetId` |
| 유효 Alert Rule | 이 노드에 적용되는 alert rule | `GET /api/v1/prometheus/rules/effective/:target_id` |
| 수정 | 노드 정보 수정 | `PUT /api/v1/targets/:id` |
| Prometheus 링크 | 해당 노드의 Prometheus 쿼리 링크 | URL 생성 |
| Grafana 링크 | 해당 노드의 Grafana 대시보드 링크 | URL 생성 |

#### UI 컴포넌트
- Data Table (정렬, 필터, 페이지네이션)
- Create/Edit Dialog (react-hook-form)
- Group Multi-Select (그룹 할당)
- Status Badge

---

### 3. Group Management (`/groups`)

#### 목록 페이지 (`/groups`)
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 그룹 목록 | `GET /api/v1/groups` |
| 생성 | 새 그룹 생성 | `POST /api/v1/groups` |
| 삭제 | 소프트 삭제 | `POST /api/v1/groups/delete` |

#### 상세 페이지 (`/groups/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | name, description, priority | `GET /api/v1/groups/:id` |
| 소속 노드 | 이 그룹에 속한 노드 목록 | `GET /api/v1/targets/group/:group_id` |
| Alert Rules | 이 그룹에 할당된 alert rule | `GET /api/v1/alert-rules/group/:group_id` |
| Script Policies | 이 그룹에 할당된 script policy | `GET /api/v1/script-policies/group/:groupId` |
| 수정 | 그룹 정보 수정 | `PUT /api/v1/groups/:id` |

#### UI 컴포넌트
- Data Table
- Create/Edit Dialog
- Target Assignment Multi-Select

---

### 4. Exporter Management (`/exporters`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 exporter 목록 | `GET /api/v1/exporters` |
| 타입 필터 | node_exporter, dcgm_exporter 등 | `GET /api/v1/exporters/type/:type` |
| 생성 | 노드에 exporter 추가 | `POST /api/v1/exporters` |
| 수정 | exporter 설정 변경 | `PUT /api/v1/exporters/:id` |
| 삭제 | exporter 제거 | `POST /api/v1/exporters/delete` |

#### UI 컴포넌트
- Data Table
- Create/Edit Dialog (Target 선택 포함)

---

### 5. Alert Template Management (`/alerts/templates`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 alert template 목록 | `GET /api/v1/alert-templates` |
| 심각도 필터 | critical, warning, info | `GET /api/v1/alert-templates/severity/:severity` |
| 생성 | 새 템플릿 생성 | `POST /api/v1/alert-templates` |
| 삭제 | 소프트 삭제 | `POST /api/v1/alert-templates/delete` |

#### 상세/편집 페이지 (`/alerts/templates/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | name, description, severity | `GET /api/v1/alert-templates/:id` |
| PromQL 템플릿 | query_template 편집 | 포함 |
| 변수 정의 | config_schema (JSON) | 포함 |
| 사용 중인 Rule | 이 템플릿을 사용하는 alert rule 목록 | `GET /api/v1/alert-rules/template/:template_id` |
| 수정 | 템플릿 수정 | `PUT /api/v1/alert-templates/:id` |

#### UI 컴포넌트
- Data Table
- PromQL Editor (syntax highlighting)
- JSON Schema Editor

---

### 6. Alert Rule Management (`/alerts/rules`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 alert rule 목록 | `GET /api/v1/alert-rules` |
| 그룹 필터 | 특정 그룹의 rule만 | `GET /api/v1/alert-rules/group/:group_id` |
| 생성 | 새 rule 생성 (그룹에 할당) | `POST /api/v1/alert-rules` |
| 삭제 | 소프트 삭제 | `POST /api/v1/alert-rules/delete` |

#### 상세/편집 페이지 (`/alerts/rules/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | name, template, group, enabled | `GET /api/v1/alert-rules/:id` |
| 설정 값 | config (템플릿 변수 값) | 포함 |
| 렌더링된 PromQL | 미리보기 | Client-side 렌더링 |
| 수정 | rule 수정 | `PUT /api/v1/alert-rules/:id` |

#### UI 컴포넌트
- Data Table
- Template Select Dropdown
- Group Select Dropdown
- Config Form (템플릿 schema 기반 동적 생성)
- PromQL Preview

---

### 7. Script Template Management (`/scripts/templates`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 script template 목록 | `GET /api/v1/script-templates` |
| 활성만 조회 | enabled=true만 | `GET /api/v1/script-templates/active` |
| 타입 필터 | check, setup, cleanup | `GET /api/v1/script-templates/type/:scriptType` |
| 생성 | 새 템플릿 생성 | `POST /api/v1/script-templates` |
| 삭제 | 소프트 삭제 | `POST /api/v1/script-templates/delete` |

#### 상세/편집 페이지 (`/scripts/templates/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | name, description, script_type | `GET /api/v1/script-templates/:id` |
| 스크립트 편집 | **Monaco Editor** (Bash syntax) | 포함 |
| 변수 정의 | config_schema (JSON) | 포함 |
| 해시 검증 | 스크립트 무결성 확인 | `GET /api/v1/script-templates/:id/verify-hash` |
| 사용 중인 Policy | 이 템플릿을 사용하는 policy 목록 | `GET /api/v1/script-policies/template/:templateId` |
| 수정 | 템플릿 수정 | `PUT /api/v1/script-templates/:id` |

#### UI 컴포넌트
- Data Table
- **Monaco Editor** (Bash syntax highlighting, line numbers)
- JSON Schema Editor
- Hash Verification Button

---

### 8. Script Policy Management (`/scripts/policies`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 policy 목록 | `GET /api/v1/script-policies` |
| 그룹 필터 | 특정 그룹의 policy만 | `GET /api/v1/script-policies/group/:groupId` |
| 생성 | 새 policy 생성 | `POST /api/v1/script-policies` |
| 삭제 | 소프트 삭제 | `POST /api/v1/script-policies/delete` |

#### 상세/편집 페이지 (`/scripts/policies/[id]`)
| Feature | Description | API |
|---------|-------------|-----|
| 기본 정보 | template, group, priority, enabled | `GET /api/v1/script-policies/:id` |
| 설정 값 | config (템플릿 변수 값) | 포함 |
| 수정 | policy 수정 | `PUT /api/v1/script-policies/:id` |

#### UI 컴포넌트
- Data Table
- Template Select Dropdown
- Group Select Dropdown (또는 Global)
- Config Form (템플릿 schema 기반 동적 생성)

---

### 9. Bootstrap Token Management (`/bootstrap`)

#### 목록 페이지
| Feature | Description | API |
|---------|-------------|-----|
| 목록 조회 | 전체 토큰 목록 | `GET /api/v1/bootstrap-tokens` |
| 생성 | 새 토큰 발급 | `POST /api/v1/bootstrap-tokens` |
| 삭제 | 토큰 폐기 | `POST /api/v1/bootstrap-tokens/delete` |
| 명령 복사 | Bootstrap curl 명령 클립보드 복사 | Client-side |

#### UI 컴포넌트
- Data Table (토큰 마스킹, 사용 횟수, 만료일)
- Create Dialog (그룹 선택, 만료일 설정, 사용 제한)
- **Bootstrap Command Generator** (토큰 선택 시 curl 명령 자동 생성)
- Copy Button

---

### 10. Settings (`/settings`)

#### Prometheus 관리
| Feature | Description | API |
|---------|-------------|-----|
| 상태 조회 | Prometheus 연결 상태 | `GET /api/v1/prometheus/status` |
| 규칙 파일 목록 | 생성된 rule 파일 목록 | `GET /api/v1/prometheus/rules/files` |
| 규칙 재생성 | 모든 규칙 재생성 | `POST /api/v1/prometheus/rules/regenerate` |
| Prometheus 리로드 | 설정 리로드 | `POST /api/v1/prometheus/reload` |

#### 외부 링크
| Feature | Description |
|---------|-------------|
| Grafana 링크 | Grafana 대시보드 URL |
| Prometheus 링크 | Prometheus UI URL |
| Alertmanager 링크 | Alertmanager UI URL |

#### UI 컴포넌트
- Status Cards
- Action Buttons (Regenerate, Reload)
- External Link Cards

---

## New API Requirements

### 1. Active Alerts API (NEW)

Alertmanager에서 현재 firing 상태인 알림을 프록시하는 API.

```
GET /api/v1/alerts/active
```

**Response:**
```json
{
  "alerts": [
    {
      "fingerprint": "abc123",
      "status": "firing",
      "labels": {
        "alertname": "HighCPU",
        "instance": "node-01:9100",
        "severity": "warning"
      },
      "annotations": {
        "summary": "High CPU usage detected",
        "description": "CPU usage is above 90%"
      },
      "starts_at": "2024-01-15T10:30:00Z",
      "generator_url": "http://prometheus:9090/..."
    }
  ],
  "total": 1
}
```

**Implementation:**
- Config Server가 Alertmanager API (`GET /api/v2/alerts`)를 호출
- 결과를 정규화하여 반환

---

### 2. Target List Query Parameters (NEW)

기존 `GET /api/v1/targets`에 검색/필터 파라미터 추가.

```
GET /api/v1/targets?search=node-01&status=active&page=1&limit=20
```

| Parameter | Type | Description |
|-----------|------|-------------|
| search | string | hostname 또는 IP 검색 |
| status | string | `active`, `inactive`, `all` |
| group_id | string | 특정 그룹의 노드만 |
| page | int | 페이지 번호 (1-based) |
| limit | int | 페이지 크기 (default: 20) |
| sort | string | 정렬 필드 (예: `hostname`, `-created_at`) |

---

### 3. Pagination Response Format (NEW)

모든 목록 API에 일관된 페이지네이션 응답 형식 적용.

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

**Affected APIs:**
- `GET /api/v1/targets`
- `GET /api/v1/groups`
- `GET /api/v1/exporters`
- `GET /api/v1/alert-templates`
- `GET /api/v1/alert-rules`
- `GET /api/v1/script-templates`
- `GET /api/v1/script-policies`
- `GET /api/v1/bootstrap-tokens`

---

## Interactive UI Components

### 1. Real-time Alert Display

```
┌─────────────────────────────────────────────────────────┐
│ 🔴 Active Alerts (3)                         [Refresh] │
├─────────────────────────────────────────────────────────┤
│ CRITICAL  HighGPUTemp      gpu-node-01       10m ago   │
│ WARNING   HighCPU          web-server-03      5m ago   │
│ WARNING   DiskSpaceHigh    storage-01        30m ago   │
└─────────────────────────────────────────────────────────┘
```

- 10초 간격 자동 갱신 (SWR `refreshInterval`)
- 헤더에 알림 배지 표시
- 클릭 시 상세 정보 표시

---

### 2. Monaco Code Editor

```
┌─────────────────────────────────────────────────────────┐
│ Script Template Editor                    [Save] [Test] │
├─────────────────────────────────────────────────────────┤
│  1 │ #!/usr/bin/env bash                               │
│  2 │ set -euo pipefail                                 │
│  3 │                                                   │
│  4 │ # Check disk usage                                │
│  5 │ THRESHOLD="${THRESHOLD:-80}"                      │
│  6 │ USAGE=$(df -h / | awk 'NR==2 {print $5}' | tr -d '%') │
│  7 │                                                   │
│  8 │ if [ "$USAGE" -gt "$THRESHOLD" ]; then            │
│  9 │   echo "aami_disk_check{status=\"critical\"} 1"   │
│ 10 │ else                                              │
│ 11 │   echo "aami_disk_check{status=\"ok\"} 0"         │
│ 12 │ fi                                                │
└─────────────────────────────────────────────────────────┘
```

- Bash syntax highlighting
- Line numbers
- Auto-completion (선택사항)
- Validation feedback

---

### 3. Bootstrap Command Generator

```
┌─────────────────────────────────────────────────────────┐
│ Bootstrap Command                                       │
├─────────────────────────────────────────────────────────┤
│ Config Server URL: [https://config.example.com    ]     │
│ Token: [aami_bootstrap_abc123... ▼]                     │
│ Labels: [env=production, rack=A1              ]         │
├─────────────────────────────────────────────────────────┤
│ Generated Command:                                      │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ curl -fsSL https://config.example.com/api/v1/      │ │
│ │   bootstrap/script | sudo bash -s -- \             │ │
│ │   --token aami_bootstrap_abc123 \                  │ │
│ │   --server https://config.example.com \            │ │
│ │   --labels env=production --labels rack=A1         │ │
│ └─────────────────────────────────────────────────────┘ │
│                                              [📋 Copy]  │
└─────────────────────────────────────────────────────────┘
```

- 토큰 선택 시 자동 생성
- 클립보드 복사
- 추가 옵션 (labels, group-id)

---

## API Summary

### Existing APIs (No Changes)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/targets` | GET, POST | Target CRUD |
| `/api/v1/targets/:id` | GET, PUT | Target 상세/수정 |
| `/api/v1/targets/delete` | POST | Target 삭제 |
| `/api/v1/groups` | GET, POST | Group CRUD |
| `/api/v1/groups/:id` | GET, PUT | Group 상세/수정 |
| `/api/v1/exporters` | GET, POST | Exporter CRUD |
| `/api/v1/alert-templates` | GET, POST | Alert Template CRUD |
| `/api/v1/alert-rules` | GET, POST | Alert Rule CRUD |
| `/api/v1/script-templates` | GET, POST | Script Template CRUD |
| `/api/v1/script-policies` | GET, POST | Script Policy CRUD |
| `/api/v1/bootstrap-tokens` | GET, POST | Bootstrap Token CRUD |
| `/api/v1/prometheus/status` | GET | Prometheus 상태 |
| `/api/v1/prometheus/rules/regenerate` | POST | 규칙 재생성 |
| `/api/v1/prometheus/reload` | POST | Prometheus 리로드 |

### New APIs Required

| Endpoint | Method | Description | Priority |
|----------|--------|-------------|----------|
| `/api/v1/alerts/active` | GET | 현재 firing 알림 조회 | **High** |

### API Enhancements Required

| Endpoint | Enhancement | Priority |
|----------|-------------|----------|
| `GET /api/v1/targets` | search, status, pagination params | **High** |
| `GET /api/v1/groups` | pagination params | Medium |
| `GET /api/v1/alert-rules` | pagination params | Medium |
| All list endpoints | Consistent pagination response | Medium |

---

## Development Phases

### Phase 1: Foundation (MVP)
- 프로젝트 셋업 (Next.js, shadcn/ui, Tailwind)
- API 클라이언트
- 레이아웃, 네비게이션
- Dashboard (요약 카드, 외부 링크)
- Target 목록/상세/CRUD
- Group 목록/상세/CRUD

### Phase 2: Core Features
- Exporter 관리
- Alert Template 관리
- Alert Rule 관리
- Bootstrap Token 관리

### Phase 3: Advanced Features
- Script Template 관리 (Monaco Editor)
- Script Policy 관리
- 실시간 알림 표시
- Settings 페이지

### Phase 4: Polish
- 검색/필터/페이지네이션
- 에러 핸들링
- 로딩 상태
- 반응형 디자인
- Docker 빌드

---

## File Structure

```
services/config-server-ui/
├── app/
│   ├── layout.tsx              # Root layout
│   ├── page.tsx                # Dashboard
│   ├── targets/
│   │   ├── page.tsx            # Target list
│   │   └── [id]/page.tsx       # Target detail
│   ├── groups/
│   │   ├── page.tsx            # Group list
│   │   └── [id]/page.tsx       # Group detail
│   ├── exporters/
│   │   └── page.tsx            # Exporter list
│   ├── alerts/
│   │   ├── templates/
│   │   │   ├── page.tsx        # Template list
│   │   │   └── [id]/page.tsx   # Template detail
│   │   └── rules/
│   │       ├── page.tsx        # Rule list
│   │       └── [id]/page.tsx   # Rule detail
│   ├── scripts/
│   │   ├── templates/
│   │   │   ├── page.tsx        # Script template list
│   │   │   └── [id]/page.tsx   # Script template detail
│   │   └── policies/
│   │       ├── page.tsx        # Policy list
│   │       └── [id]/page.tsx   # Policy detail
│   ├── bootstrap/
│   │   └── page.tsx            # Bootstrap token management
│   └── settings/
│       └── page.tsx            # Settings
├── components/
│   ├── ui/                     # shadcn/ui components
│   ├── layout/
│   │   ├── header.tsx
│   │   ├── sidebar.tsx
│   │   └── footer.tsx
│   ├── targets/
│   │   ├── target-table.tsx
│   │   ├── target-form.tsx
│   │   └── target-card.tsx
│   ├── groups/
│   │   ├── group-table.tsx
│   │   └── group-form.tsx
│   ├── alerts/
│   │   ├── alert-badge.tsx
│   │   └── active-alerts.tsx
│   ├── scripts/
│   │   └── script-editor.tsx   # Monaco Editor wrapper
│   └── bootstrap/
│       └── command-generator.tsx
├── lib/
│   ├── api/
│   │   ├── client.ts           # API client
│   │   ├── targets.ts
│   │   ├── groups.ts
│   │   └── ...
│   ├── hooks/
│   │   ├── use-targets.ts
│   │   └── use-groups.ts
│   └── utils.ts
├── types/
│   └── api.ts                  # TypeScript types
├── next.config.ts
├── tailwind.config.ts
├── package.json
├── Dockerfile
└── nginx.conf
```

---

## Notes

- **Static Export**: `next build`로 정적 파일 생성, nginx로 서빙
- **CORS**: Config Server에 CORS 미들웨어 이미 포함
- **Authentication**: 현재 명세에 미포함 (향후 추가 가능)
- **i18n**: 초기 버전은 영어만 지원
