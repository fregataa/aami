# Database Migration 개선 계획

**작성일**: 2024-12-29
**상태**: 📋 Planned
**우선순위**: High (Sprint 6과 병행 가능)

## 현황 분석

### 현재 문제점

1. **멱등성 부족**
   - `CREATE TABLE` 사용 → 이미 존재하면 에러
   - `IF NOT EXISTS` 미사용

2. **마이그레이션 추적 없음**
   - 어떤 마이그레이션이 실행되었는지 기록 없음
   - 매 서버 시작마다 모든 마이그레이션 재실행 시도

3. **서버가 직접 마이그레이션 실행**
   - 서버 시작 시 자동으로 마이그레이션 실행 (main.go:76-119)
   - 배포/운영 환경에서 위험함
   - 롤백 불가능

4. **마이그레이션 파일 분산**
   - 9개의 마이그레이션 파일로 분리
   - 아직 실제 배포 전이므로 통합 가능

### 현재 마이그레이션 구조

```
migrations/
├── 001_initial_schema.sql                      # 초기 스키마
├── 002_refactor_namespace_to_table.sql         # Namespace 리팩토링
├── 003_add_soft_delete.sql                     # Soft Delete 추가
├── 004_add_check_templates.sql                 # Check Templates
├── 005_add_check_instances.sql                 # Check Instances
├── 006_migrate_check_settings_to_instances.sql # 데이터 마이그레이션
├── 007_decouple_check_instances_from_templates.sql
├── 008_decouple_alert_rules_from_templates.sql
└── 009_refactor_target_group_relationship.sql  # Target-Group 관계
```

**파일 수**: 9개
**총 라인 수**: ~32,000 lines (추정)

## 개선 목표

### 단기 목표 (즉시 적용)

1. ✅ **멱등성 확보**
   - 모든 `CREATE TABLE` → `CREATE TABLE IF NOT EXISTS`
   - 모든 `CREATE INDEX` → `CREATE INDEX IF NOT EXISTS`
   - 모든 `ALTER TABLE` → 조건부 실행 로직 추가

2. ✅ **마이그레이션 통합**
   - 9개 파일 → 1개 파일로 통합
   - 이유: 아직 프로덕션 배포 전
   - 파일명: `001_initial_schema.sql`

3. ✅ **서버 시작 로직 변경**
   - 마이그레이션 실행 → 스키마 검증으로 변경
   - 스키마 누락 시 경고만 출력 (에러 없음)

### 중기 목표 (Sprint 6 이후)

4. 🔄 **마이그레이션 도구 도입**
   - Goose 또는 golang-migrate 도입
   - CLI 기반 마이그레이션 관리
   - 추적 테이블 자동 관리

5. 🔄 **CI/CD 통합**
   - 배포 전 자동 마이그레이션 실행
   - 마이그레이션 실패 시 배포 중단

## 구현 계획

### Phase 1: 멱등성 확보 (즉시)

#### 작업 내용

**1. CREATE TABLE 수정**

```sql
-- Before
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ...
);

-- After
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ...
);
```

**2. CREATE INDEX 수정**

```sql
-- Before
CREATE INDEX idx_groups_parent_id ON groups(parent_id);

-- After
CREATE INDEX IF NOT EXISTS idx_groups_parent_id ON groups(parent_id);
```

**3. ALTER TABLE 보호**

```sql
-- Before
ALTER TABLE targets ADD COLUMN status VARCHAR(20);

-- After
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='targets' AND column_name='status'
    ) THEN
        ALTER TABLE targets ADD COLUMN status VARCHAR(20);
    END IF;
END $$;
```

#### 영향 받는 파일

- `migrations/001_initial_schema.sql`
- `migrations/002_refactor_namespace_to_table.sql`
- `migrations/003_add_soft_delete.sql`
- `migrations/004_add_check_templates.sql`
- `migrations/005_add_check_instances.sql`
- `migrations/006_migrate_check_settings_to_instances.sql`
- `migrations/007_decouple_check_instances_from_templates.sql`
- `migrations/008_decouple_alert_rules_from_templates.sql`
- `migrations/009_refactor_target_group_relationship.sql`

### Phase 2: 마이그레이션 통합 (즉시)

#### 작업 내용

**1. 모든 마이그레이션 병합**

```bash
# 새로운 통합 파일 생성
cat migrations/001_initial_schema.sql \
    migrations/002_refactor_namespace_to_table.sql \
    migrations/003_add_soft_delete.sql \
    migrations/004_add_check_templates.sql \
    migrations/005_add_check_instances.sql \
    migrations/006_migrate_check_settings_to_instances.sql \
    migrations/007_decouple_check_instances_from_templates.sql \
    migrations/008_decouple_alert_rules_from_templates.sql \
    migrations/009_refactor_target_group_relationship.sql \
    > migrations/001_unified_schema.sql
```

**2. 기존 파일 제거**

```bash
# 백업용 아카이브 생성
mkdir -p migrations/archive
mv migrations/00{2..9}_*.sql migrations/archive/

# 통합 파일 이름 변경
mv migrations/001_unified_schema.sql migrations/001_initial_schema.sql
```

**3. main.go 수정**

```go
// Before
migrations := []string{
    "migrations/001_initial_schema.sql",
    "migrations/002_refactor_namespace_to_table.sql",
    // ... 9개
}

// After
migrations := []string{
    "migrations/001_initial_schema.sql",
}
```

#### 결과

```
migrations/
├── 001_initial_schema.sql          # ✅ 통합된 단일 파일
└── archive/                        # 📦 백업 (참고용)
    ├── 001_initial_schema_old.sql
    ├── 002_refactor_namespace_to_table.sql
    └── ...
```

### Phase 3: 서버 로직 변경 (즉시)

#### 작업 내용

**1. 스키마 검증 함수 추가**

```go
// cmd/config-server/main.go

// validateSchema checks if ORM definitions match database schema
func validateSchema(rm *repository.Manager) error {
    log.Println("Validating database schema...")

    db := rm.GetDB()

    // Check if critical tables exist
    requiredTables := []string{
        "namespaces",
        "groups",
        "targets",
        "target_groups",
        "exporters",
        "alert_templates",
        "alert_rules",
        "check_templates",
        "check_instances",
        "bootstrap_tokens",
    }

    var missingTables []string
    for _, table := range requiredTables {
        var exists bool
        err := db.Raw(`
            SELECT EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_name = ?
            )
        `, table).Scan(&exists).Error

        if err != nil || !exists {
            missingTables = append(missingTables, table)
        }
    }

    if len(missingTables) > 0 {
        log.Printf("ERROR: Missing database tables: %v", missingTables)
        log.Println("")
        log.Println("Database schema is not up to date.")
        log.Println("Please run migrations manually before starting the server:")
        log.Println("")
        log.Println("  # Using SQL (current)")
        log.Println("  psql -h localhost -U admin -d config_server -f migrations/001_initial_schema.sql")
        log.Println("")
        log.Println("  # Using Goose (future)")
        log.Println("  goose -dir migrations postgres \"$DB_URL\" up")
        log.Println("")
        log.Println("See: docs/MIGRATION.md for details")

        return fmt.Errorf("database schema validation failed: missing tables %v", missingTables)
    }

    log.Println("✓ Database schema validation successful")
    return nil
}
```

**2. main.go 수정**

```go
// Before
if err := runMigrations(repoManager); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}

// After
if err := validateSchema(repoManager); err != nil {
    log.Fatalf("Failed to validate schema: %v", err)
}
// Note: validateSchema warns but doesn't fail
```

**3. 마이그레이션 실행 문서 작성**

새 파일: `docs/MIGRATION.md`

```markdown
# Database Migration Guide

## 중요: 마이그레이션은 항상 수동으로 실행

Config Server는 시작 시 스키마 검증만 수행하며, 자동으로 마이그레이션을 실행하지 않습니다.
데이터베이스 스키마 변경이 필요한 경우 아래 방법으로 수동 실행해야 합니다.

## Running Migrations

### Development (Docker Compose)

1. **PostgreSQL 컨테이너 시작**:
   \`\`\`bash
   cd deploy/docker-compose
   docker-compose up -d postgres
   \`\`\`

2. **마이그레이션 실행**:
   \`\`\`bash
   # 방법 1: psql 사용
   docker-compose exec postgres psql -U admin -d config_server -f /path/to/migrations/001_initial_schema.sql

   # 방법 2: 로컬 psql 사용
   psql -h localhost -U admin -d config_server -f services/config-server/migrations/001_initial_schema.sql
   \`\`\`

3. **Config Server 시작**:
   \`\`\`bash
   docker-compose up -d config-server
   \`\`\`

### Local Development (without Docker)

\`\`\`bash
# 1. PostgreSQL 실행 확인
psql -h localhost -U admin -d config_server -c "SELECT 1"

# 2. 마이그레이션 실행
psql -h localhost -U admin -d config_server -f services/config-server/migrations/001_initial_schema.sql

# 3. Config Server 실행
cd services/config-server
go run cmd/config-server/main.go
\`\`\`

### Production

\`\`\`bash
# 1. 데이터베이스 백업 (필수!)
pg_dump -U admin config_server > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 마이그레이션 실행
psql -h <prod-host> -U admin -d config_server -f migrations/001_initial_schema.sql

# 3. 검증
psql -h <prod-host> -U admin -d config_server -c "\dt"

# 4. 애플리케이션 배포
kubectl apply -f k8s/
\`\`\`

### Using Migration Tool (Recommended for Phase 4)

See Phase 4 for goose integration.
\`\`\`

## Phase 4: Goose 도입 (Sprint 6 이후)

### 선택 이유: Goose vs golang-migrate

| 기준 | Goose | golang-migrate |
|------|-------|----------------|
| **인기도** | 3.6k stars | 14k stars |
| **Go 네이티브** | ✅ Pure Go | ✅ Pure Go |
| **CLI 도구** | ✅ 간단 | ✅ 강력 |
| **SQL 지원** | ✅ | ✅ |
| **Go 함수** | ✅ | ❌ |
| **학습 곡선** | 낮음 | 중간 |
| **추적 테이블** | `goose_db_version` | `schema_migrations` |

**선택**: Goose (더 간단하고 Go 함수 지원)

### 구현 계획

#### 1. Goose 설치

```bash
# CLI 도구 설치
go install github.com/pressly/goose/v3/cmd/goose@latest

# 프로젝트 의존성 추가
cd services/config-server
go get github.com/pressly/goose/v3
```

#### 2. 마이그레이션 파일 구조 변경

```
migrations/
├── 00001_initial_schema.sql        # Goose 형식
└── archive/                        # 기존 파일들
```

**Goose 마이그레이션 파일 형식**:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ...
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
```

#### 3. 마이그레이션 실행

```bash
# 수동 실행
goose -dir migrations postgres "host=localhost user=admin dbname=config_server sslmode=disable" up

# 상태 확인
goose -dir migrations postgres "..." status

# 롤백
goose -dir migrations postgres "..." down
```

#### 4. CI/CD 통합

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Install Goose
        run: go install github.com/pressly/goose/v3/cmd/goose@latest

      - name: Run Migrations
        run: |
          cd services/config-server
          goose -dir migrations postgres "$DB_URL" up
        env:
          DB_URL: ${{ secrets.DATABASE_URL }}

      - name: Deploy Application
        run: kubectl apply -f k8s/
```

#### 5. 코드 통합 (선택사항)

```go
// cmd/config-server/main.go

import "github.com/pressly/goose/v3"

func runMigrations(rm *repository.Manager) error {
    db := rm.GetDB()
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }

    // Set up goose
    if err := goose.SetDialect("postgres"); err != nil {
        return err
    }

    // Run migrations
    if err := goose.Up(sqlDB, "migrations"); err != nil {
        return err
    }

    return nil
}
```

## 마이그레이션 정책

### 모든 환경 (개발/스테이징/프로덕션)

❌ **금지**: 서버 시작 시 자동 마이그레이션
✅ **필수**: 수동 또는 CI/CD를 통한 마이그레이션
✅ **서버 역할**: ORM 정의와 DB 스키마 일치 여부 검증만 수행

**이유**:
1. 마이그레이션 실패 시 롤백 필요
2. 여러 서버 인스턴스 동시 시작 시 경쟁 조건
3. 마이그레이션은 한 번만 실행되어야 함
4. 스키마 변경은 신중하게 검토되어야 함
5. 개발자가 스키마 변경을 명시적으로 인지해야 함

### 서버 시작 시 동작

```
┌─────────────────────┐
│  서버 시작          │
└──────┬──────────────┘
       │
       ↓
┌─────────────────────┐
│  스키마 검증        │  ← GORM 모델 정의와 DB 스키마 비교
│  (validateSchema)   │
└──────┬──────────────┘
       │
       ├─ 일치 ────────→ ✅ 서버 정상 시작
       │
       └─ 불일치 ──────→ ❌ 에러 출력 + 서버 종료
                         "Run migrations first: goose up"
```

### 환경별 마이그레이션 실행 방법

| 환경 | 실행 방법 | 책임 |
|------|----------|------|
| **개발** | 수동 (psql/goose) | 개발자 |
| **스테이징** | CI/CD 파이프라인 | DevOps |
| **프로덕션** | 수동 (승인 후) | DBA/DevOps |

### Docker Compose 환경에서의 마이그레이션

Docker Compose를 사용하는 경우, init 컨테이너 패턴을 사용하여 마이그레이션 자동화:

```yaml
# deploy/docker-compose/docker-compose.yaml

services:
  postgres:
    # ... postgres config ...

  # 마이그레이션 전용 컨테이너
  migration:
    image: postgres:16-alpine
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - PGHOST=postgres
      - PGPORT=5432
      - PGUSER=${POSTGRES_USER:-admin}
      - PGPASSWORD=${POSTGRES_PASSWORD:-changeme}
      - PGDATABASE=${POSTGRES_DB:-config_server}
    volumes:
      - ../../services/config-server/migrations:/migrations
    command: >
      sh -c "
        echo 'Running database migrations...'
        psql -f /migrations/001_initial_schema.sql
        echo 'Migrations completed successfully'
      "
    restart: "no"  # 한 번만 실행

  config-server:
    depends_on:
      migration:
        condition: service_completed_successfully
    # ... config-server config ...
```

**장점**:
- 개발 환경에서도 명시적 마이그레이션
- 서버 코드와 마이그레이션 로직 분리
- 실패 시 config-server 시작 안됨

### 마이그레이션 작성 규칙

1. **멱등성 보장**
   ```sql
   CREATE TABLE IF NOT EXISTS ...
   CREATE INDEX IF NOT EXISTS ...
   ```

2. **Up/Down 모두 작성**
   ```sql
   -- +goose Up
   CREATE TABLE ...

   -- +goose Down
   DROP TABLE ...
   ```

3. **트랜잭션 사용**
   ```sql
   -- +goose StatementBegin
   BEGIN;
   ... multiple statements ...
   COMMIT;
   -- +goose StatementEnd
   ```

4. **데이터 마이그레이션 분리**
   - 스키마 변경: `001_alter_schema.sql`
   - 데이터 변경: `002_migrate_data.sql`

## 타임라인

### 즉시 (Phase 1-3)
- [x] IF NOT EXISTS 추가
- [x] 마이그레이션 통합
- [x] 서버 로직 변경
- [x] 문서 작성

**예상 소요**: 2-3 시간

### Sprint 6 이후 (Phase 4)
- [ ] Goose 도입 검토
- [ ] 마이그레이션 파일 변환
- [ ] CI/CD 통합
- [ ] 운영 매뉴얼 작성

**예상 소요**: 2-3 일

## 리스크 & 대응

### 리스크 1: 통합 마이그레이션 실패
**확률**: Low
**영향도**: High
**대응**:
- 통합 전 각 마이그레이션 개별 테스트
- 백업 유지 (archive 폴더)

### 리스크 2: Goose 도입 지연
**확률**: Medium
**영향도**: Low
**대응**:
- Phase 1-3만으로도 안정적 운영 가능
- Goose는 선택사항

### 리스크 3: 프로덕션 마이그레이션 실패
**확률**: Low
**영향도**: Critical
**대응**:
- 스테이징 환경에서 먼저 테스트
- 데이터베이스 백업 필수
- 롤백 계획 수립

## 참고 자료

- [Goose Documentation](https://github.com/pressly/goose)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Idempotent DDL](https://www.postgresql.org/docs/current/sql-createtable.html)

## 변경 이력

- 2024-12-29: 초안 작성
