# Check 시스템 네이밍 및 구조 리팩토링 계획

**작성일**: 2024-12-29
**목적**: 역할 중심 네이밍 적용 및 불필요한 CheckSetting 제거
**예상 기간**: 2-3일

---

## 변경 사항 요약

### 네이밍 변경

```
현재                    →  변경 후 (역할 중심)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CheckTemplate          →  MonitoringScript
CheckInstance          →  ScriptPolicy
CheckSetting           →  [삭제]
```

**대안 (더 간결)**:
```
CheckTemplate          →  Check
CheckInstance          →  CheckPolicy
CheckSetting           →  [삭제]
```

### CheckSetting 제거 근거

**현재 상태 분석**:
- ✅ 도메인 모델 존재: `internal/domain/check_setting.go`
- ✅ 유닛 테스트 존재: `test/unit/domain/check_setting_test.go`
- ✅ 통합 테스트 존재: `test/integration/repository/check_setting_test.go`
- ❌ **Repository 미구현**: `repository/manager.go`에 없음
- ❌ **API 미구현**: 엔드포인트 없음
- ❌ **Service 미구현**: 비즈니스 로직 없음
- ❌ **DB 테이블 없음**: `migrations/001_initial_schema.sql`에 없음
- ❌ **실제 사용처 없음**: CheckInstance와 기능 중복

**결론**: CheckSetting은 계획만 되고 실제 구현되지 않은 기능. CheckInstance와 역할이 중복되어 불필요함.

---

## 작업 계획

### Phase 1: CheckSetting 제거 (0.5일)

#### Task 1.1: 도메인 모델 삭제
```bash
# 삭제 대상 파일
rm services/config-server/internal/domain/check_setting.go
```

**영향받는 파일**:
- `internal/database/postgres.go` - AutoMigrate에서 제거
- `test/testutil/fixtures.go` - NewTestCheckSetting 제거
- `test/testutil/containers.go` - check_settings 테이블 truncate 제거

#### Task 1.2: 테스트 파일 삭제
```bash
# 삭제 대상 파일
rm test/unit/domain/check_setting_test.go
rm test/integration/repository/check_setting_test.go
```

#### Task 1.3: 참조 제거
- [ ] `internal/database/postgres.go` - `&domain.CheckSetting{}` 제거
- [ ] `test/testutil/fixtures.go` - `NewTestCheckSetting()` 함수 제거
- [ ] `test/testutil/containers.go` - `"check_settings"` 제이블 cleanup 제거

---

### Phase 2: CheckTemplate → MonitoringScript 리네이밍 (1일)

**네이밍 최종 결정 필요**: MonitoringScript vs Check

#### Task 2.1: 도메인 모델 리네이밍

**파일 이름 변경**:
```bash
cd services/config-server/internal/domain
mv check_template.go monitoring_script.go
```

**구조체 리네이밍**:
```go
// Before
type CheckTemplate struct {
    ID            string
    Name          string
    CheckType     string
    ScriptContent string
    // ...
}

// After
type MonitoringScript struct {
    ID            string
    Name          string
    ScriptType    string  // CheckType → ScriptType
    ScriptContent string
    // ...
}
```

**함수 리네이밍**:
```go
// Before
func (ct *CheckTemplate) ComputeHash() string
func (ct *CheckTemplate) UpdateHash()
func (ct *CheckTemplate) VerifyHash() bool
func (ct *CheckTemplate) Validate() error

// After
func (ms *MonitoringScript) ComputeHash() string
func (ms *MonitoringScript) UpdateHash()
func (ms *MonitoringScript) VerifyHash() bool
func (ms *MonitoringScript) Validate() error
```

#### Task 2.2: Repository 리네이밍

**파일 이름 변경**:
```bash
cd services/config-server/internal/repository
mv check_template.go monitoring_script.go
```

**인터페이스 및 구현 리네이밍**:
```go
// Before
type CheckTemplateRepository interface {
    Create(ctx context.Context, template *domain.CheckTemplate) error
    GetByID(ctx context.Context, id string) (*domain.CheckTemplate, error)
    // ...
}

type checkTemplateRepository struct {
    db *gorm.DB
}

func NewCheckTemplateRepository(db *gorm.DB) CheckTemplateRepository {
    return &checkTemplateRepository{db: db}
}

// After
type MonitoringScriptRepository interface {
    Create(ctx context.Context, script *domain.MonitoringScript) error
    GetByID(ctx context.Context, id string) (*domain.MonitoringScript, error)
    // ...
}

type monitoringScriptRepository struct {
    db *gorm.DB
}

func NewMonitoringScriptRepository(db *gorm.DB) MonitoringScriptRepository {
    return &monitoringScriptRepository{db: db}
}
```

#### Task 2.3: Repository Manager 업데이트

**파일**: `internal/repository/manager.go`

```go
// Before
type Manager struct {
    CheckTemplate   CheckTemplateRepository
    CheckInstance   CheckInstanceRepository
    // ...
}

// After
type Manager struct {
    MonitoringScript MonitoringScriptRepository
    ScriptPolicy     ScriptPolicyRepository  // CheckInstance도 같이 변경
    // ...
}

// Before
func newManagerWithDB(db *gorm.DB) *Manager {
    return &Manager{
        CheckTemplate: NewCheckTemplateRepository(db),
        CheckInstance: NewCheckInstanceRepository(db),
        // ...
    }
}

// After
func newManagerWithDB(db *gorm.DB) *Manager {
    return &Manager{
        MonitoringScript: NewMonitoringScriptRepository(db),
        ScriptPolicy:     NewScriptPolicyRepository(db),
        // ...
    }
}
```

#### Task 2.4: 데이터베이스 마이그레이션

**새 마이그레이션 파일**: `migrations/010_rename_check_tables.sql`

```sql
-- Rename check_templates table to monitoring_scripts
ALTER TABLE IF EXISTS check_templates RENAME TO monitoring_scripts;

-- Rename check_instances table to script_policies
ALTER TABLE IF EXISTS check_instances RENAME TO script_policies;

-- Update foreign key column names if needed
-- (check if check_instances has template_id → script_id)

-- Update indexes
ALTER INDEX IF EXISTS idx_check_templates_name
    RENAME TO idx_monitoring_scripts_name;
ALTER INDEX IF EXISTS idx_check_templates_check_type
    RENAME TO idx_monitoring_scripts_script_type;

ALTER INDEX IF EXISTS idx_check_instances_scope
    RENAME TO idx_script_policies_scope;
-- ... (rename all relevant indexes)
```

#### Task 2.5: API 업데이트

**예상 영향**:
- Handler 파일
- DTO 구조체
- 라우터 경로

**파일들**:
- `internal/api/handler/check_template.go` → `monitoring_script.go`
- `internal/api/dto/check_template.go` → `monitoring_script.go`
- `internal/api/router.go` - 라우트 경로 업데이트

**API 경로 변경**:
```go
// Before
r.POST("/api/v1/check-templates", handler.CreateCheckTemplate)
r.GET("/api/v1/check-templates", handler.ListCheckTemplates)
r.GET("/api/v1/check-templates/:id", handler.GetCheckTemplate)

// After
r.POST("/api/v1/monitoring-scripts", handler.CreateMonitoringScript)
r.GET("/api/v1/monitoring-scripts", handler.ListMonitoringScripts)
r.GET("/api/v1/monitoring-scripts/:id", handler.GetMonitoringScript)

// Or (간결한 버전)
r.POST("/api/v1/checks", handler.CreateCheck)
r.GET("/api/v1/checks", handler.ListChecks)
r.GET("/api/v1/checks/:id", handler.GetCheck)
```

#### Task 2.6: 테스트 업데이트

**영향받는 파일**:
- `test/unit/domain/check_template_test.go` → `monitoring_script_test.go`
- `test/integration/repository/check_template_test.go` → `monitoring_script_test.go`
- `test/testutil/fixtures.go` - `NewTestCheckTemplate()` → `NewTestMonitoringScript()`

---

### Phase 3: CheckInstance → ScriptPolicy 리네이밍 (1일)

동일한 패턴으로 진행:

#### Task 3.1: 도메인 모델
```bash
mv check_instance.go script_policy.go
```

```go
// Before
type CheckInstance struct {
    // Template fields (copied at creation time)
    Name          string
    CheckType     string
    ScriptContent string
    // ...
}

// After
type ScriptPolicy struct {
    // Script fields (copied at creation time)
    Name          string
    ScriptType    string  // CheckType → ScriptType
    ScriptContent string
    // ...
}
```

#### Task 3.2: Repository
```bash
mv check_instance.go script_policy.go
```

```go
type ScriptPolicyRepository interface {
    Create(ctx context.Context, policy *domain.ScriptPolicy) error
    GetByID(ctx context.Context, id string) (*domain.ScriptPolicy, error)
    // ...
}
```

#### Task 3.3: API & 테스트
- Handler, DTO, 라우터 업데이트
- 모든 테스트 업데이트

---

### Phase 4: 통합 테스트 및 검증 (0.5일)

#### Task 4.1: 컴파일 검증
```bash
cd services/config-server
go build ./...
```

#### Task 4.2: 테스트 실행
```bash
# Unit tests
go test ./test/unit/... -v

# Integration tests
go test ./test/integration/... -v

# All tests
go test ./... -v
```

#### Task 4.3: 마이그레이션 테스트
```bash
# Test migration on dev database
psql -U aami -d aami_dev < migrations/010_rename_check_tables.sql

# Verify table names
psql -U aami -d aami_dev -c "\dt"
```

#### Task 4.4: API 엔드포인트 테스트
```bash
# Start server
go run cmd/server/main.go

# Test new endpoints
curl http://localhost:8080/api/v1/monitoring-scripts
curl http://localhost:8080/api/v1/script-policies
```

---

### Phase 5: 문서 업데이트 (0.5일)

#### Task 5.1: API 문서
- `docs/en/API.md` - 엔드포인트 경로 업데이트
- `docs/kr/API.md` - 한글 문서 업데이트

#### Task 5.2: 아키텍처 문서
- `docs/en/CHECK-MANAGEMENT.md` - 용어 업데이트
  - CheckTemplate → MonitoringScript
  - CheckInstance → ScriptPolicy
- `docs/kr/CHECK-MANAGEMENT.md` - 한글 문서 업데이트

#### Task 5.3: README 업데이트
- 주요 컨셉 설명 부분 용어 변경
- 예제 코드 업데이트

---

## 파일 변경 체크리스트

### 삭제 (CheckSetting)
- [ ] `internal/domain/check_setting.go`
- [ ] `test/unit/domain/check_setting_test.go`
- [ ] `test/integration/repository/check_setting_test.go`
- [ ] `internal/database/postgres.go` - CheckSetting 참조 제거
- [ ] `test/testutil/fixtures.go` - NewTestCheckSetting 제거
- [ ] `test/testutil/containers.go` - check_settings truncate 제거

### 리네이밍 (CheckTemplate → MonitoringScript)
- [ ] `internal/domain/check_template.go` → `monitoring_script.go`
- [ ] `internal/repository/check_template.go` → `monitoring_script.go`
- [ ] `internal/repository/manager.go` - CheckTemplate 필드 업데이트
- [ ] `internal/api/handler/check_template.go` → `monitoring_script.go`
- [ ] `internal/api/dto/check_template.go` → `monitoring_script.go`
- [ ] `internal/api/router.go` - 라우트 업데이트
- [ ] `test/unit/domain/check_template_test.go` → `monitoring_script_test.go`
- [ ] `test/integration/repository/check_template_test.go` → `monitoring_script_test.go`
- [ ] `test/testutil/fixtures.go` - NewTestCheckTemplate 리네이밍

### 리네이밍 (CheckInstance → ScriptPolicy)
- [ ] `internal/domain/check_instance.go` → `script_policy.go`
- [ ] `internal/repository/check_instance.go` → `script_policy.go`
- [ ] `internal/repository/manager.go` - CheckInstance 필드 업데이트
- [ ] `internal/api/handler/check_instance.go` → `script_policy.go`
- [ ] `internal/api/dto/check_instance.go` → `script_policy.go`
- [ ] `internal/api/router.go` - 라우트 업데이트
- [ ] `test/unit/domain/check_instance_test.go` → `script_policy_test.go`
- [ ] `test/integration/repository/check_instance_test.go` → `script_policy_test.go`
- [ ] `test/testutil/fixtures.go` - NewTestCheckInstance 리네이밍

### 마이그레이션
- [ ] `migrations/010_rename_check_tables.sql` 생성
  - check_templates → monitoring_scripts
  - check_instances → script_policies
  - 인덱스 리네이밍
  - FK 컬럼 업데이트 (필요시)

### 문서
- [ ] `docs/en/CHECK-MANAGEMENT.md` - 용어 업데이트
- [ ] `docs/kr/CHECK-MANAGEMENT.md` - 용어 업데이트
- [ ] `docs/en/API.md` - API 경로 업데이트
- [ ] `README.md` - 예제 및 용어 업데이트
- [ ] `.agent/planning/refactoring/check-system-naming-refactor.md` (이 문서)

---

## 네이밍 최종 결정 필요

두 가지 옵션 중 선택:

### 옵션 A: 역할 중심 (명확성 우선)
```
CheckTemplate  →  MonitoringScript
CheckInstance  →  ScriptPolicy
```

**장점**:
- 역할이 명확함 (모니터링용 스크립트, 스크립트 정책)
- "Check"라는 모호한 용어 제거

**단점**:
- 이름이 다소 길어짐
- API 경로가 길어짐 (`/api/v1/monitoring-scripts`)

### 옵션 B: 간결성 우선
```
CheckTemplate  →  Check
CheckInstance  →  CheckPolicy
```

**장점**:
- 간결함
- Alert 시스템과 패턴 일치 (AlertTemplate/AlertRule vs Check/CheckPolicy)
- API 경로 간결 (`/api/v1/checks`)

**단점**:
- "Check"가 여전히 모호할 수 있음

### 추천: 옵션 B (Check / CheckPolicy)
- 간결하고 일관성 있음
- Alert 시스템과 네이밍 패턴 유사
- API 경로 깔끔

---

## 위험 요소 및 대응

### 위험 1: 프로덕션 데이터 손실
**완화책**:
- 마이그레이션 스크립트를 ALTER TABLE로 작성 (데이터 유지)
- 백업 후 마이그레이션 실행

### 위험 2: API 호환성 깨짐
**완화책**:
- API 버전 관리 고려 (v1 유지, v2 신규)
- 또는 모든 클라이언트가 내부 시스템이므로 동시 업데이트

### 위험 3: 누락된 참조
**완화책**:
- 전체 grep으로 "CheckTemplate", "CheckInstance", "CheckSetting" 검색
- 컴파일 에러로 누락 확인

---

## 작업 순서

1. **CheckSetting 제거** (가장 쉬움, 영향 적음)
2. **CheckTemplate 리네이밍** (독립적)
3. **CheckInstance 리네이밍** (CheckTemplate 참조 있을 수 있음)
4. **마이그레이션 작성 및 테스트**
5. **통합 테스트**
6. **문서 업데이트**

**예상 소요 시간**: 2-3일 (테스트 포함)

---

## 커밋 전략

### 커밋 1: Remove CheckSetting
```
refactor: remove unused CheckSetting domain model

- Delete domain/check_setting.go
- Delete check_setting tests
- Remove CheckSetting from database AutoMigrate
- Remove CheckSetting from test fixtures

Reason: CheckSetting was planned but never implemented.
It duplicates CheckInstance functionality and causes confusion.
```

### 커밋 2: Rename CheckTemplate to MonitoringScript (or Check)
```
refactor: rename CheckTemplate to MonitoringScript

- Rename domain model and all references
- Rename repository and interfaces
- Update API handlers and routes
- Update tests and fixtures
- Add migration script for table rename

Breaking change: API endpoints changed
- /api/v1/check-templates → /api/v1/monitoring-scripts
```

### 커밋 3: Rename CheckInstance to ScriptPolicy (or CheckPolicy)
```
refactor: rename CheckInstance to ScriptPolicy

- Rename domain model and all references
- Rename repository and interfaces
- Update API handlers and routes
- Update tests and fixtures
- Add migration script for table rename

Breaking change: API endpoints changed
- /api/v1/check-instances → /api/v1/script-policies
```

### 커밋 4: Update documentation
```
docs: update terminology after check system refactoring

- Update CHECK-MANAGEMENT.md
- Update API.md
- Update README.md
```

---

## 후속 작업 (선택사항)

### 향후 CheckOverride 추가 시
타겟 레벨 예외 처리가 필요하면:

```go
type CheckOverride struct {
    ID        string
    CheckID   string   // Check (MonitoringScript) 참조
    TargetID  string   // Target에 직접 매핑
    Config    map[string]interface{}
    Priority  int
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

이렇게 하면:
- Check (스크립트 정의)
- CheckPolicy (그룹 기본 설정)
- CheckOverride (타겟 예외)

3-tier 구조 완성.

---

## 참고 자료

- Alert 시스템 구조: `internal/domain/alert.go`
- 현재 Check 구조: `internal/domain/check_template.go`, `check_instance.go`
- 마이그레이션 예제: `migrations/001_initial_schema.sql`
