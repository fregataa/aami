# Check System Migration Guide (CheckScript → CheckTemplate/Instance)

## 개요

이 가이드는 기존 CheckScript/CheckSetting 시스템을 새로운 CheckTemplate/CheckInstance 패턴으로 마이그레이션하는 방법을 설명합니다.

## 변경 사항 요약

### 기존 시스템 (Deprecated)
- **CheckScript**: 스코프별 스크립트 정의 (Global/Namespace/Group)
- **CheckSetting**: 그룹별 파라미터 설정

### 새로운 시스템
- **CheckTemplate**: 재사용 가능한 체크 정의 (스크립트 + 기본 파라미터)
- **CheckInstance**: 스코프별 템플릿 적용 (Global/Namespace/Group)

### 주요 개선점
1. **템플릿 재사용**: 동일한 스크립트를 여러 스코프에서 다른 파라미터로 재사용
2. **일관된 패턴**: Alert 시스템과 동일한 Template/Instance 패턴
3. **명확한 네이밍**: Template(정의) vs Instance(적용)
4. **자동 업데이트**: Hash 기반 버전 감지

## 마이그레이션 단계

### Phase 1: 준비 (완료)

✅ 코드 리팩토링 완료
- Domain 모델 생성
- Repository, Service, Handler, Router 구현
- 기존 CheckScript/CheckSetting 코드 제거

✅ DB 스키마 준비 완료
- 마이그레이션 파일 생성 (004, 005)
- main.go에 마이그레이션 등록

### Phase 2: DB 마이그레이션 실행

#### 2.1. 서버 재시작

새로운 테이블이 자동으로 생성됩니다:

```bash
# 환경 변수 설정
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=aami_config

# 서버 실행 (마이그레이션 자동 실행)
./config-server
```

로그에서 다음 메시지 확인:
```
Executing migration: migrations/004_add_check_templates.sql
Migration completed: migrations/004_add_check_templates.sql
Executing migration: migrations/005_add_check_instances.sql
Migration completed: migrations/005_add_check_instances.sql
```

#### 2.2. 테이블 생성 확인

```sql
-- 새 테이블 확인
SELECT * FROM check_templates LIMIT 1;
SELECT * FROM check_instances LIMIT 1;
```

### Phase 3: 데이터 마이그레이션 (선택)

기존 `check_settings` 데이터가 있는 경우에만 필요합니다.

#### 3.1. 체크 템플릿 생성

기존 체크 타입별로 템플릿을 생성해야 합니다.

**예시: Disk Usage Check 템플릿**

```bash
curl -X POST http://localhost:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "disk-usage-check",
    "check_type": "disk",
    "script_content": "#!/bin/bash\ndf -h / | tail -1 | awk '\''{print \"disk_usage_percent{mount=\\\"/\\\"} \"$5}'\'' | sed '\''s/%//'\''",
    "language": "bash",
    "default_config": {
      "threshold": 80,
      "mount": "/"
    },
    "description": "Disk usage check for root filesystem",
    "version": "1.0.0"
  }'
```

#### 3.2. 기존 체크 타입 확인

```sql
-- 마이그레이션이 필요한 체크 타입 확인
SELECT DISTINCT check_type
FROM check_settings
WHERE NOT EXISTS (
    SELECT 1 FROM check_templates
    WHERE check_templates.check_type = check_settings.check_type
);
```

#### 3.3. 데이터 마이그레이션 실행

모든 템플릿 생성 후:

```sql
-- migrations/006_migrate_check_settings_to_instances.sql 파일의
-- 주석을 해제하고 실행
```

또는 psql 명령어:

```bash
psql -h localhost -U postgres -d aami_config \
  -f migrations/006_migrate_check_settings_to_instances.sql
```

#### 3.4. 마이그레이션 검증

```sql
-- 변환 결과 확인
SELECT
    cs.id as old_setting_id,
    cs.check_type,
    cs.group_id,
    ci.id as new_instance_id,
    t.name as template_name
FROM check_settings cs
LEFT JOIN groups g ON cs.group_id = g.id
LEFT JOIN check_templates t ON t.check_type = cs.check_type
LEFT JOIN check_instances ci ON ci.template_id = t.id AND ci.group_id = cs.group_id
ORDER BY cs.check_type, cs.group_id;
```

모든 행에 `new_instance_id`가 있어야 합니다.

### Phase 4: API 전환

#### 4.1. 새로운 API 엔드포인트

**체크 템플릿 관리**
```
POST   /api/v1/check-templates          # 템플릿 생성
GET    /api/v1/check-templates           # 템플릿 목록
GET    /api/v1/check-templates/:id       # 템플릿 조회
PUT    /api/v1/check-templates/:id       # 템플릿 수정
DELETE /api/v1/check-templates/:id       # 템플릿 삭제 (soft)
```

**체크 인스턴스 관리**
```
POST   /api/v1/check-instances           # 인스턴스 생성
GET    /api/v1/check-instances           # 인스턴스 목록
GET    /api/v1/check-instances/:id       # 인스턴스 조회
PUT    /api/v1/check-instances/:id       # 인스턴스 수정
DELETE /api/v1/check-instances/:id       # 인스턴스 삭제 (soft)
```

**노드 API (에이전트용)**
```
GET    /api/v1/checks/node/:hostname     # 노드의 effective checks
```

#### 4.2. 기존 API (Deprecated)

다음 엔드포인트는 더 이상 사용되지 않습니다:
- `/api/v1/check-scripts/*`
- `/api/v1/check-settings/*`

### Phase 5: 정리 (선택)

마이그레이션이 완전히 검증된 후:

```sql
-- 백업 생성
CREATE TABLE check_settings_backup AS SELECT * FROM check_settings;

-- 기존 테이블 삭제
DROP TABLE IF EXISTS check_settings CASCADE;
```

## 새로운 워크플로우 예시

### 1. 템플릿 생성

```bash
curl -X POST http://localhost:8080/api/v1/check-templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory-usage-check",
    "check_type": "memory",
    "script_content": "#!/bin/bash\nfree -m | awk '\''NR==2{printf \"memory_usage_percent %.0f\", $3*100/$2}'\''",
    "language": "bash",
    "default_config": {
      "threshold": 90
    },
    "version": "1.0.0"
  }'
```

### 2. 글로벌 인스턴스 생성

```bash
curl -X POST http://localhost:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "uuid-of-memory-check-template",
    "scope": "global",
    "config": {
      "threshold": 85
    },
    "priority": 100,
    "is_active": true
  }'
```

### 3. 그룹별 오버라이드

```bash
curl -X POST http://localhost:8080/api/v1/check-instances \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "uuid-of-memory-check-template",
    "scope": "group",
    "namespace_id": "uuid-of-namespace",
    "group_id": "uuid-of-critical-group",
    "config": {
      "threshold": 70
    },
    "priority": 50,
    "is_active": true
  }'
```

우선순위: Group (50) > Global (100) → Critical 그룹은 threshold 70 사용

### 4. 노드에서 조회

```bash
curl http://localhost:8080/api/v1/checks/node/web-server-01
```

응답:
```json
[
  {
    "check_type": "memory",
    "script_content": "#!/bin/bash\n...",
    "language": "bash",
    "config": {
      "threshold": 70
    },
    "version": "1.0.0",
    "hash": "abc123...",
    "template_id": "uuid-of-template",
    "instance_id": "uuid-of-group-instance"
  }
]
```

## 롤백 절차

문제 발생 시:

1. **코드 롤백**: 이전 버전으로 되돌리기
2. **DB 롤백**:
   ```sql
   DROP TABLE IF EXISTS check_instances CASCADE;
   DROP TABLE IF EXISTS check_templates CASCADE;
   ```
3. 서버 재시작

## 문제 해결

### 마이그레이션 실패

**증상**: "Failed to execute migration" 오류

**해결**:
```sql
-- 트랜잭션 확인
SELECT * FROM pg_stat_activity WHERE datname = 'aami_config';

-- 필요시 연결 종료
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE datname = 'aami_config' AND pid <> pg_backend_pid();
```

### 중복 데이터

**증상**: "duplicate key value violates unique constraint"

**해결**:
```sql
-- 중복 확인
SELECT template_id, scope, namespace_id, group_id, COUNT(*)
FROM check_instances
GROUP BY template_id, scope, namespace_id, group_id
HAVING COUNT(*) > 1;

-- 중복 제거 (오래된 것 삭제)
DELETE FROM check_instances a USING check_instances b
WHERE a.id > b.id
AND a.template_id = b.template_id
AND a.scope = b.scope
AND (a.namespace_id = b.namespace_id OR (a.namespace_id IS NULL AND b.namespace_id IS NULL))
AND (a.group_id = b.group_id OR (a.group_id IS NULL AND b.group_id IS NULL));
```

## 참고 문서

- [Check Management API](./CHECK-MANAGEMENT.md)
- [Architecture Design](../ARCHITECTURE.md)
- [Alert Template/Rule Pattern](./ALERT-MANAGEMENT.md)
