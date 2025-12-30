# Default Templates Config File Implementation

## Overview

Alert Template과 Script Template의 기본 데이터를 YAML 설정 파일로 관리하고, CLI 명령을 통해 DB에 삽입하는 기능을 구현합니다. 이 명령은 Config Server 설치 시 호출됩니다.

---

## Directory Structure

```
configs/
├── defaults/
│   ├── alert-templates.yaml      # 기본 Alert Template 정의
│   ├── script-templates.yaml     # Script Template 메타데이터
│   └── scripts/                  # 실제 스크립트 파일
│       ├── health-check.sh
│       ├── disk-usage.sh
│       └── process-monitor.sh
```

---

## YAML Schema

### alert-templates.yaml

```yaml
# configs/defaults/alert-templates.yaml
templates:
  - name: HighCPUUsage
    description: Alert when CPU usage exceeds threshold
    severity: warning
    query_template: |
      100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > {{ .threshold }}
    default_config:
      threshold: 80
      for_duration: "5m"
    labels:
      category: resource

  - name: HighMemoryUsage
    description: Alert when memory usage exceeds threshold
    severity: warning
    query_template: |
      (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > {{ .threshold }}
    default_config:
      threshold: 85
      for_duration: "5m"
    labels:
      category: resource

  - name: DiskSpaceLow
    description: Alert when disk space is running low
    severity: critical
    query_template: |
      (1 - (node_filesystem_avail_bytes{fstype!~"tmpfs|overlay"} / node_filesystem_size_bytes)) * 100 > {{ .threshold }}
    default_config:
      threshold: 90
      for_duration: "10m"
    labels:
      category: storage

  - name: InstanceDown
    description: Alert when instance is unreachable
    severity: critical
    query_template: |
      up{job="{{ .job }}"} == 0
    default_config:
      job: "node"
      for_duration: "1m"
    labels:
      category: availability

  - name: HighNetworkTraffic
    description: Alert when network traffic exceeds threshold
    severity: info
    query_template: |
      rate(node_network_receive_bytes_total[5m]) * 8 / 1000000 > {{ .threshold_mbps }}
    default_config:
      threshold_mbps: 100
      for_duration: "5m"
    labels:
      category: network
```

### script-templates.yaml

```yaml
# configs/defaults/script-templates.yaml
templates:
  - name: health-check
    description: Basic system health check script
    script_type: bash
    script_file: scripts/health-check.sh    # 상대 경로
    config_schema:
      type: object
      properties:
        timeout:
          type: integer
          default: 30
        check_disk:
          type: boolean
          default: true
        check_memory:
          type: boolean
          default: true
    enabled: true

  - name: disk-usage
    description: Detailed disk usage report
    script_type: bash
    script_file: scripts/disk-usage.sh
    config_schema:
      type: object
      properties:
        threshold_percent:
          type: integer
          default: 80
        include_mounts:
          type: array
          items:
            type: string
          default: ["/", "/home", "/var"]
    enabled: true

  - name: process-monitor
    description: Monitor specific processes
    script_type: bash
    script_file: scripts/process-monitor.sh
    config_schema:
      type: object
      properties:
        processes:
          type: array
          items:
            type: string
          default: ["nginx", "postgres"]
        restart_on_failure:
          type: boolean
          default: false
    enabled: true
```

---

## Implementation Steps

### Step 1: Define Config Types

**File: `internal/config/defaults.go`**

```go
type DefaultsConfig struct {
    AlertTemplatesFile  string `mapstructure:"alert_templates_file"`
    ScriptTemplatesFile string `mapstructure:"script_templates_file"`
    ScriptsDir          string `mapstructure:"scripts_dir"`
    SyncOnStartup       bool   `mapstructure:"sync_on_startup"`
}

type AlertTemplateYAML struct {
    Templates []AlertTemplateEntry `yaml:"templates"`
}

type AlertTemplateEntry struct {
    Name          string                 `yaml:"name"`
    Description   string                 `yaml:"description"`
    Severity      string                 `yaml:"severity"`
    QueryTemplate string                 `yaml:"query_template"`
    DefaultConfig map[string]interface{} `yaml:"default_config"`
    Labels        map[string]string      `yaml:"labels"`
}

type ScriptTemplateYAML struct {
    Templates []ScriptTemplateEntry `yaml:"templates"`
}

type ScriptTemplateEntry struct {
    Name         string                 `yaml:"name"`
    Description  string                 `yaml:"description"`
    ScriptType   string                 `yaml:"script_type"`
    ScriptFile   string                 `yaml:"script_file"`
    ConfigSchema map[string]interface{} `yaml:"config_schema"`
    Enabled      bool                   `yaml:"enabled"`
}
```

### Step 2: Create Loader Service

**File: `internal/service/defaults_loader.go`**

```go
type DefaultsLoaderService struct {
    config              *config.DefaultsConfig
    alertTemplateRepo   repository.AlertTemplateRepository
    scriptTemplateRepo  repository.ScriptTemplateRepository
    logger              *slog.Logger
}

func (s *DefaultsLoaderService) LoadAll(ctx context.Context) error {
    if err := s.loadAlertTemplates(ctx); err != nil {
        return fmt.Errorf("load alert templates: %w", err)
    }
    if err := s.loadScriptTemplates(ctx); err != nil {
        return fmt.Errorf("load script templates: %w", err)
    }
    return nil
}

func (s *DefaultsLoaderService) loadAlertTemplates(ctx context.Context) error {
    // 1. YAML 파일 파싱
    // 2. 각 템플릿에 대해 upsert (name 기준)
    // 3. is_builtin = true 설정
}

func (s *DefaultsLoaderService) loadScriptTemplates(ctx context.Context) error {
    // 1. YAML 메타데이터 파싱
    // 2. script_file에서 실제 스크립트 내용 로드
    // 3. hash 계산
    // 4. upsert (name 기준)
    // 5. is_builtin = true 설정
}
```

### Step 3: Add is_builtin Field

**Migration: Add `is_builtin` column**

```sql
ALTER TABLE alert_templates ADD COLUMN is_builtin BOOLEAN DEFAULT FALSE;
ALTER TABLE script_templates ADD COLUMN is_builtin BOOLEAN DEFAULT FALSE;
```

### Step 4: Update Domain & Repository

- `AlertTemplate` 도메인에 `IsBuiltin bool` 필드 추가
- `ScriptTemplate` 도메인에 `IsBuiltin bool` 필드 추가
- Repository에 `UpsertByName` 메서드 추가

### Step 5: Create CLI Command

**File: `cmd/aami/cmd/seed.go`**

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
    Use:   "seed",
    Short: "Seed database with default data",
}

var seedTemplatesCmd = &cobra.Command{
    Use:   "templates",
    Short: "Seed default alert and script templates",
    Long: `Load default templates from config files and insert into database.
This command is typically run once during initial setup.

Examples:
  aami seed templates
  aami seed templates --force    # Overwrite existing builtin templates
  aami seed templates --dry-run  # Preview without inserting`,
    RunE: runSeedTemplates,
}

func init() {
    rootCmd.AddCommand(seedCmd)
    seedCmd.AddCommand(seedTemplatesCmd)

    seedTemplatesCmd.Flags().Bool("force", false, "Overwrite existing builtin templates")
    seedTemplatesCmd.Flags().Bool("dry-run", false, "Preview changes without inserting")
}

func runSeedTemplates(cmd *cobra.Command, args []string) error {
    force, _ := cmd.Flags().GetBool("force")
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // 1. Load config
    // 2. Initialize DB connection
    // 3. Create DefaultsLoaderService
    // 4. Call loader.LoadAll(ctx, force, dryRun)
    // 5. Print summary

    return nil
}
```

**CLI Usage:**

```bash
# 기본 실행 (새 템플릿만 추가)
aami seed templates

# 기존 builtin 템플릿도 업데이트
aami seed templates --force

# 변경 사항 미리보기
aami seed templates --dry-run
```

### Step 6: Update Install Script

**File: `scripts/install-server.sh` (기존 스크립트에 추가)**

```bash
#!/bin/bash

# ... 기존 설치 로직 ...

# Run database migrations
echo "Running database migrations..."
./aami migrate up

# Seed default templates
echo "Seeding default templates..."
./aami seed templates

echo "Installation complete!"
```

### Step 7: Update Config File

**File: `configs/config.yaml`**

```yaml
defaults:
  alert_templates_file: "configs/defaults/alert-templates.yaml"
  script_templates_file: "configs/defaults/script-templates.yaml"
  scripts_dir: "configs/defaults/scripts"
```

---

## Default Templates 목록

### Alert Templates (5개)

| Name | Severity | Description |
|------|----------|-------------|
| HighCPUUsage | warning | CPU 사용률 임계치 초과 |
| HighMemoryUsage | warning | 메모리 사용률 임계치 초과 |
| DiskSpaceLow | critical | 디스크 공간 부족 |
| InstanceDown | critical | 인스턴스 접근 불가 |
| HighNetworkTraffic | info | 네트워크 트래픽 임계치 초과 |

### Script Templates (3개)

| Name | Type | Description |
|------|------|-------------|
| health-check | bash | 기본 시스템 상태 점검 |
| disk-usage | bash | 상세 디스크 사용량 리포트 |
| process-monitor | bash | 특정 프로세스 모니터링 |

---

## Files to Create/Modify

| Action | File | Description |
|--------|------|-------------|
| CREATE | `configs/defaults/alert-templates.yaml` | Alert Template 정의 |
| CREATE | `configs/defaults/script-templates.yaml` | Script Template 메타데이터 |
| CREATE | `configs/defaults/scripts/health-check.sh` | 상태 점검 스크립트 |
| CREATE | `configs/defaults/scripts/disk-usage.sh` | 디스크 사용량 스크립트 |
| CREATE | `configs/defaults/scripts/process-monitor.sh` | 프로세스 모니터링 스크립트 |
| CREATE | `internal/config/defaults.go` | Config 타입 정의 |
| CREATE | `internal/service/defaults_loader.go` | 로더 서비스 |
| CREATE | `cmd/aami/cmd/seed.go` | CLI seed 명령 |
| CREATE | `migrations/XXXXXX_add_is_builtin.sql` | Migration |
| MODIFY | `internal/domain/alert.go` | IsBuiltin 필드 추가 |
| MODIFY | `internal/domain/script.go` | IsBuiltin 필드 추가 |
| MODIFY | `internal/repository/alert_template.go` | UpsertByName 추가 |
| MODIFY | `internal/repository/script_template.go` | UpsertByName 추가 |
| MODIFY | `configs/config.yaml` | defaults 섹션 추가 |
| MODIFY | `scripts/install-server.sh` | seed 명령 호출 추가 |

---

## UI 표시

Web UI에서 builtin 템플릿 구분 표시:

```tsx
// Badge로 builtin 표시
{template.is_builtin && (
  <Badge variant="outline" className="text-blue-600">
    Built-in
  </Badge>
)}
```

Builtin 템플릿은 삭제 불가 (soft-delete만 가능하거나 삭제 버튼 숨김)

---

## Test Scenarios

1. **초기 실행**: `aami seed templates` - 빈 DB에 모든 기본 템플릿 생성
2. **재실행**: 이미 존재하는 템플릿은 스킵 (사용자 수정 보존)
3. **강제 업데이트**: `--force` 플래그로 builtin 템플릿 덮어쓰기
4. **Dry-run**: `--dry-run`으로 변경 사항 미리보기
5. **파일 누락**: 설정 파일 없으면 에러 메시지 출력
6. **잘못된 YAML**: 파싱 에러 시 명확한 에러 메시지와 함께 종료

---

## Future Enhancements (Out of Scope)

- [ ] API를 통한 템플릿 동기화 (`POST /api/v1/admin/sync-defaults`)
- [ ] 버전 관리 (템플릿 버전 필드)
- [ ] 템플릿 import/export 기능
- [ ] 커뮤니티 템플릿 저장소 연동
