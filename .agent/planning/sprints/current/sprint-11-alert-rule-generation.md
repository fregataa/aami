# Sprint 11: Alert Rule Generation & Prometheus Integration

**Status**: 📋 Planned → 🚧 Current
**Duration**: 13-17 days (~2.5-3 weeks)
**Prerequisites**: Sprint 1-6 (Error Handling completed)
**Phase**: Phase 1.5 - Alert System Completion

---

## Executive Summary

Complete the Alert system by implementing Prometheus rule generation, enabling dynamic group-based alert customization that's currently only available in the database but not deployed to Prometheus.

**Current State**:
- ✅ AlertTemplate/AlertRule API fully functional
- ✅ Domain logic complete (RenderQuery, config merge)
- ✅ Database schema implemented
- ❌ **No Prometheus rule file generation**
- ❌ **No automatic deployment to Prometheus**

**Target State**:
- ✅ Automatic Prometheus YAML generation from AlertRules
- ✅ Zero-downtime rule deployment
- ✅ Group-based alert customization in production
- ✅ Debugging tools (effective rules, policy trace)

---

## Background

The Alert system exists only in the database. AlertTemplate and AlertRule APIs work perfectly, but Prometheus continues using static YAML files. This sprint bridges that gap.

**Problem**:
- Users can create AlertRules via API
- Rules are stored in database
- Prometheus never sees them
- Group-based customization is unusable

**Solution**:
- Generate Prometheus YAML from database AlertRules
- Deploy to Prometheus rules directory
- Trigger Prometheus reload
- Provide effective rules API for debugging

---

## Goals

1. **Enable production alert customization** - Deploy AlertRules to Prometheus automatically
2. **Maintain zero downtime** - Seamless rule updates without Prometheus restart
3. **Ensure correctness** - Validate rules before deployment
4. **Provide visibility** - Effective rules and policy tracing APIs

---

## Implementation Plan

### **Task 1: Prometheus Rule Generator Core** (3-4 days)

**Goal**: AlertRule → Prometheus YAML transformation

**Files to Create**:
- `internal/service/prometheus_rule_generator.go`
- `internal/domain/prometheus_rule.go`

**Core Structures**:
```go
// Prometheus Rule File Structure
type PrometheusRuleFile struct {
    Groups []PrometheusRuleGroup `yaml:"groups"`
}

type PrometheusRuleGroup struct {
    Name     string            `yaml:"name"`
    Interval string            `yaml:"interval,omitempty"`
    Rules    []PrometheusRule  `yaml:"rules"`
}

type PrometheusRule struct {
    Alert       string            `yaml:"alert"`
    Expr        string            `yaml:"expr"`
    For         string            `yaml:"for,omitempty"`
    Labels      map[string]string `yaml:"labels"`
    Annotations map[string]string `yaml:"annotations"`
}

// Generator Service
type PrometheusRuleGenerator struct {
    alertRuleRepo repository.AlertRuleRepository
    groupRepo     repository.GroupRepository
    ruleFilePath  string
    logger        *slog.Logger
}
```

**Key Methods**:
```go
func (g *PrometheusRuleGenerator) GenerateRulesForGroup(ctx context.Context, groupID string) error
func (g *PrometheusRuleGenerator) GenerateAllRules(ctx context.Context) error
func (g *PrometheusRuleGenerator) DeleteRulesForGroup(ctx context.Context, groupID string) error
func (g *PrometheusRuleGenerator) convertToPrometheusRule(rule *domain.AlertRule) (*PrometheusRule, error)
```

**Logic Flow**:
1. Query enabled AlertRules for group (not deleted, enabled=true)
2. Call `rule.RenderQuery()` to get PromQL with config values
3. Extract `for_duration`, `labels`, `annotations` from config
4. Build PrometheusRule struct
5. Group by group_id
6. Marshal to YAML
7. Write to file: `/rules/generated/group-{groupID}.yml`

**Validation**:
- Unit tests for each transformation step
- YAML syntax validation
- PromQL rendering correctness

**Deliverables**:
- [ ] Generator service implementation
- [ ] Unit tests (>80% coverage)
- [ ] YAML marshaling logic
- [ ] Error handling for invalid queries

---

### **Task 2: File System Manager** (1-2 days)

**Goal**: Safe file operations with validation

**File to Create**:
- `internal/pkg/prometheus/file_manager.go`

**Structure**:
```go
type RuleFileManager struct {
    basePath string
    logger   *slog.Logger
}

// Core methods
func (m *RuleFileManager) WriteRuleFile(groupID string, content []byte) error
func (m *RuleFileManager) DeleteRuleFile(groupID string) error
func (m *RuleFileManager) ListRuleFiles() ([]string, error)
func (m *RuleFileManager) ValidateRuleFile(filePath string) error
func (m *RuleFileManager) BackupRuleFile(groupID string) error
```

**Features**:
- **Atomic Write**: Write to temp file → validate → rename
- **Validation**: Use `promtool check rules` before deployment
- **Backup**: Keep previous version for rollback
- **Permissions**: Set 0644, ensure correct ownership
- **Directory Creation**: Auto-create if missing

**Error Handling**:
```go
// Specific errors for better debugging
var (
    ErrDirectoryNotFound = errors.New("rules directory not found")
    ErrPermissionDenied  = errors.New("insufficient permissions")
    ErrValidationFailed  = errors.New("rule validation failed")
    ErrAtomicWriteFailed = errors.New("atomic write failed")
)
```

**Deliverables**:
- [ ] File manager implementation
- [ ] Atomic write with rollback
- [ ] promtool integration
- [ ] Unit tests with temp directories
- [ ] Error handling tests

---

### **Task 3: Prometheus Reload Integration** (2 days)

**Goal**: Trigger Prometheus to load new rules

**File to Create**:
- `internal/pkg/prometheus/client.go`

**Structure**:
```go
type PrometheusClient struct {
    baseURL string
    client  *http.Client
    timeout time.Duration
    logger  *slog.Logger
}

// Core methods
func (c *PrometheusClient) Reload(ctx context.Context) error
func (c *PrometheusClient) HealthCheck(ctx context.Context) error
func (c *PrometheusClient) ValidateConfig(ctx context.Context) error
```

**Reload Implementation**:
```go
func (c *PrometheusClient) Reload(ctx context.Context) error {
    // Option 1: HTTP POST (Recommended)
    req, _ := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/-/reload", nil)

    resp, err := c.client.Do(req)
    // ... handle response

    // Option 2: SIGHUP (Fallback)
    if httpFailed {
        return c.sendSIGHUP()
    }
}
```

**Retry Logic**:
- Max 3 attempts
- Exponential backoff: 1s, 2s, 4s
- Context cancellation support
- Health check after reload

**Configuration**:
```yaml
prometheus:
  url: "http://prometheus:9090"
  reload_timeout: "30s"
  reload_enabled: true
  health_check_enabled: true
```

**Deliverables**:
- [ ] HTTP client implementation
- [ ] Retry logic with backoff
- [ ] Health check integration
- [ ] Mock tests for HTTP calls
- [ ] Integration test with real Prometheus

---

### **Task 4: Service Layer Integration** (2 days)

**Goal**: Trigger rule generation on AlertRule changes

**Files to Modify**:
- `internal/service/alert.go`

**Integration Points**:
```go
type AlertRuleService struct {
    ruleRepo      repository.AlertRuleRepository
    templateRepo  repository.AlertTemplateRepository
    groupRepo     repository.GroupRepository
    ruleGenerator *PrometheusRuleGenerator  // NEW
}

// Modify existing methods
func (s *AlertRuleService) Create(ctx context.Context, req dto.CreateAlertRuleRequest) (*domain.AlertRule, error) {
    // ... existing DB logic ...

    // Generate Prometheus rules for this group
    if err := s.ruleGenerator.GenerateRulesForGroup(ctx, rule.GroupID); err != nil {
        s.logger.Error("Failed to generate Prometheus rules",
            "group_id", rule.GroupID, "error", err)
        // Don't fail the request - rule is in DB, can sync later
    }

    return rule, nil
}

func (s *AlertRuleService) Update(ctx context.Context, id string, req dto.UpdateAlertRuleRequest) (*domain.AlertRule, error) {
    // ... existing logic ...

    // Regenerate rules for affected group
    if err := s.ruleGenerator.GenerateRulesForGroup(ctx, rule.GroupID); err != nil {
        s.logger.Error("Failed to regenerate Prometheus rules", "error", err)
    }

    return rule, nil
}

func (s *AlertRuleService) Delete(ctx context.Context, id string) error {
    rule, _ := s.ruleRepo.GetByID(ctx, id)

    // ... existing DB delete ...

    // Check if group has other rules
    rules, _ := s.ruleRepo.GetByGroupID(ctx, rule.GroupID)
    if len(rules) == 0 {
        // No more rules for this group - delete file
        s.ruleGenerator.DeleteRulesForGroup(ctx, rule.GroupID)
    } else {
        // Regenerate remaining rules
        s.ruleGenerator.GenerateRulesForGroup(ctx, rule.GroupID)
    }

    return nil
}
```

**Error Handling Strategy**:
- **DB save succeeds, rule generation fails**: Log error, continue (manual sync available)
- **DB save fails**: Transaction rollback, no rule generation attempted
- **Validation fails**: Prevent DB save, return error to user

**Deliverables**:
- [ ] Service integration
- [ ] Error handling logic
- [ ] Logging for debugging
- [ ] Integration tests for create/update/delete flows

---

### **Task 5: API Endpoints** (1-2 days)

**Goal**: Manual control and debugging APIs

**File to Create**:
- `internal/api/handler/prometheus_rule.go`

**New Endpoints**:
```go
type PrometheusRuleHandler struct {
    ruleGenerator *service.PrometheusRuleGenerator
    prometheusClient *prometheus.Client
}

// POST /api/v1/prometheus/rules/regenerate
// Regenerate all rule files (full sync)
func (h *PrometheusRuleHandler) RegenerateAll(c *gin.Context)

// POST /api/v1/prometheus/rules/regenerate/:group_id
// Regenerate rules for specific group
func (h *PrometheusRuleHandler) RegenerateGroup(c *gin.Context)

// GET /api/v1/prometheus/rules/files
// List generated rule files
func (h *PrometheusRuleHandler) ListRuleFiles(c *gin.Context)

// POST /api/v1/prometheus/reload
// Manually trigger Prometheus reload
func (h *PrometheusRuleHandler) ReloadPrometheus(c *gin.Context)

// GET /api/v1/prometheus/rules/effective/:target_id
// Get effective rules for a target (considering group hierarchy)
func (h *PrometheusRuleHandler) GetEffectiveRulesByTarget(c *gin.Context)
```

**Router Configuration**:
```go
// internal/api/router.go
prometheusRules := v1.Group("/prometheus")
{
    prometheusRules.POST("/rules/regenerate", prometheusRuleHandler.RegenerateAll)
    prometheusRules.POST("/rules/regenerate/:group_id", prometheusRuleHandler.RegenerateGroup)
    prometheusRules.GET("/rules/files", prometheusRuleHandler.ListRuleFiles)
    prometheusRules.POST("/reload", prometheusRuleHandler.ReloadPrometheus)
    prometheusRules.GET("/rules/effective/:target_id", prometheusRuleHandler.GetEffectiveRulesByTarget)
}
```

**Response DTOs**:
```go
type RegenerateResponse struct {
    GroupsAffected int      `json:"groups_affected"`
    FilesGenerated int      `json:"files_generated"`
    Errors         []string `json:"errors,omitempty"`
    Duration       string   `json:"duration"`
}

type RuleFileInfo struct {
    GroupID   string `json:"group_id"`
    FileName  string `json:"file_name"`
    RuleCount int    `json:"rule_count"`
    Size      int64  `json:"size_bytes"`
    ModTime   string `json:"modified_at"`
}

type EffectiveRulesResponse struct {
    TargetID string            `json:"target_id"`
    Rules    []EffectiveRule   `json:"rules"`
}

type EffectiveRule struct {
    Name        string                 `json:"name"`
    Severity    string                 `json:"severity"`
    Query       string                 `json:"query"`
    Config      map[string]interface{} `json:"config"`
    Source      string                 `json:"source"` // "global", "namespace", "group"
    SourceID    string                 `json:"source_id"`
}
```

**Deliverables**:
- [ ] Handler implementation
- [ ] DTO definitions
- [ ] Router integration
- [ ] API tests
- [ ] Error response handling

---

### **Task 6: Configuration** (1 day)

**Goal**: Configurable Prometheus integration

**Files to Modify**:
- `internal/config/config.go`
- `config/config.yaml`

**Configuration Structure**:
```go
type Config struct {
    // ... existing fields ...
    Prometheus PrometheusConfig `yaml:"prometheus"`
}

type PrometheusConfig struct {
    URL           string        `yaml:"url" env:"PROMETHEUS_URL" default:"http://localhost:9090"`
    RulePath      string        `yaml:"rule_path" env:"PROMETHEUS_RULE_PATH" default:"/etc/prometheus/rules/generated"`
    ReloadEnabled bool          `yaml:"reload_enabled" env:"PROMETHEUS_RELOAD_ENABLED" default:"true"`
    ReloadTimeout time.Duration `yaml:"reload_timeout" env:"PROMETHEUS_RELOAD_TIMEOUT" default:"30s"`
    ValidateRules bool          `yaml:"validate_rules" env:"PROMETHEUS_VALIDATE_RULES" default:"true"`
    PromtoolPath  string        `yaml:"promtool_path" env:"PROMTOOL_PATH" default:"promtool"`
}
```

**Environment Variables**:
```bash
# Prometheus Integration
PROMETHEUS_URL=http://prometheus:9090
PROMETHEUS_RULE_PATH=/etc/prometheus/rules/generated
PROMETHEUS_RELOAD_ENABLED=true
PROMETHEUS_RELOAD_TIMEOUT=30s
PROMETHEUS_VALIDATE_RULES=true
PROMTOOL_PATH=/usr/local/bin/promtool
```

**Docker Compose Volume Setup**:
```yaml
volumes:
  prometheus-rules:

services:
  prometheus:
    volumes:
      - prometheus-rules:/etc/prometheus/rules/generated:ro
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--web.enable-lifecycle'  # Enable reload API

  config-server:
    volumes:
      - prometheus-rules:/app/rules
    environment:
      PROMETHEUS_URL: http://prometheus:9090
      PROMETHEUS_RULE_PATH: /app/rules
```

**Prometheus Config Update**:
```yaml
# config/prometheus/prometheus.yml
rule_files:
  - /etc/prometheus/rules/*.yml
  - /etc/prometheus/rules/generated/*.yml  # ADD THIS
```

**Deliverables**:
- [ ] Config struct definition
- [ ] Environment variable mapping
- [ ] Docker Compose configuration
- [ ] Prometheus config update
- [ ] Configuration validation

---

### **Task 7: Testing** (2-3 days)

**Goal**: Comprehensive test coverage

**Unit Tests** (`test/unit/service/`):
```go
// prometheus_rule_generator_test.go
func TestGenerateRulesForGroup(t *testing.T)
func TestConvertToPrometheusRule(t *testing.T)
func TestYAMLMarshaling(t *testing.T)
func TestConfigMerge(t *testing.T)
func TestEmptyGroup(t *testing.T)

// file_manager_test.go
func TestAtomicWrite(t *testing.T)
func TestValidation(t *testing.T)
func TestRollback(t *testing.T)
func TestPermissions(t *testing.T)

// prometheus_client_test.go
func TestReload(t *testing.T)
func TestRetryLogic(t *testing.T)
func TestTimeout(t *testing.T)
func TestHealthCheck(t *testing.T)
```

**Integration Tests** (`test/integration/`):
```go
// prometheus_integration_test.go
func TestEndToEndRuleGeneration(t *testing.T) {
    // Setup: Start testcontainers Prometheus
    // Create AlertRule via API
    // Verify: Rule file created
    // Verify: Prometheus reloaded
    // Verify: Rule visible in Prometheus
}

func TestRuleUpdateFlow(t *testing.T)
func TestRuleDeletionFlow(t *testing.T)
func TestInvalidRuleHandling(t *testing.T)
func TestPrometheusDowntime(t *testing.T)
```

**Test Scenarios**:
1. **Happy Path**: Create rule → File generated → Prometheus reload → Rule active
2. **Invalid PromQL**: Create rule with bad query → Validation fails → No file created
3. **Prometheus Down**: Create rule → File generated → Reload fails → Retry succeeds
4. **File Permissions**: Test with read-only directory → Proper error handling
5. **Concurrent Updates**: Multiple rules for same group → Atomic writes succeed

**Coverage Target**: >80% for new code

**Deliverables**:
- [ ] Unit tests for all core functions
- [ ] Integration tests with testcontainers
- [ ] Error scenario tests
- [ ] Performance tests (100 rules/group)
- [ ] CI/CD integration

---

### **Task 8: Documentation** (1 day)

**Goal**: Complete operational documentation

**Files to Update/Create**:

1. **`docs/kr/ALERTING-SYSTEM.md`** (Update lines 480-545)
   - Change status from ❌ to ✅
   - Add implementation details
   - Document new APIs

2. **`docs/en/ALERTING-SYSTEM.md`** (Same updates)

3. **`docs/kr/API.md`** (Add new section)
   ```markdown
   ## Prometheus Rule Management

   ### Regenerate All Rules
   POST /api/v1/prometheus/rules/regenerate

   ### Regenerate Group Rules
   POST /api/v1/prometheus/rules/regenerate/:group_id

   ### List Rule Files
   GET /api/v1/prometheus/rules/files

   ### Trigger Prometheus Reload
   POST /api/v1/prometheus/reload

   ### Get Effective Rules for Target
   GET /api/v1/prometheus/rules/effective/:target_id
   ```

4. **Create `docs/en/PROMETHEUS-INTEGRATION.md`**
   - Setup guide
   - Troubleshooting
   - Common issues and solutions
   - Best practices

5. **Update `README.md`**
   - Remove Sprint references (already done)
   - Keep only link to TRACKER.md

**Deliverables**:
- [ ] Updated alerting system docs
- [ ] API documentation for new endpoints
- [ ] Operational guide
- [ ] Troubleshooting guide
- [ ] README cleanup complete

---

## Work Schedule

| Task | Duration | Dependencies | Start | End |
|------|----------|--------------|-------|-----|
| 1. Rule Generator Core | 3-4 days | - | Day 1 | Day 4 |
| 2. File Manager | 1-2 days | Task 1 | Day 3 | Day 5 |
| 3. Prometheus Client | 2 days | - | Day 4 | Day 6 |
| 4. Service Integration | 2 days | Task 1,2,3 | Day 6 | Day 8 |
| 5. API Endpoints | 1-2 days | Task 4 | Day 8 | Day 10 |
| 6. Configuration | 1 day | - | Day 5 | Day 6 |
| 7. Testing | 2-3 days | Task 1-5 | Day 10 | Day 13 |
| 8. Documentation | 1 day | Task 7 | Day 13 | Day 14 |

**Total**: 13-17 days (~2.5-3 weeks)

**Critical Path**: Task 1 → Task 2 → Task 4 → Task 5 → Task 7

---

## Technical Considerations

### 1. Prometheus Rule File Location

**Chosen Approach**: Shared Docker volume

```yaml
# docker-compose.yml
volumes:
  prometheus-rules:

services:
  prometheus:
    volumes:
      - prometheus-rules:/etc/prometheus/rules/generated:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--web.enable-lifecycle'  # Required for reload API

  config-server:
    volumes:
      - prometheus-rules:/app/rules
    environment:
      PROMETHEUS_RULE_PATH: /app/rules
```

**Benefits**:
- Clean separation
- Easier permission management
- Works in both Docker and K8s

### 2. Reload Strategy

**Primary**: HTTP POST to `/-/reload`
- Requires `--web.enable-lifecycle` flag
- Zero downtime
- Immediate error feedback

**Fallback**: SIGHUP signal
- Works without lifecycle flag
- No direct error feedback
- Requires process access

### 3. Validation Strategy

**Pre-deployment**:
1. PromQL syntax check (in-app)
2. YAML format validation (go-yaml)
3. promtool validation (external)

**Post-deployment**:
1. Health check Prometheus
2. Verify rule loaded (query API)
3. Monitor for errors

### 4. Error Recovery

| Scenario | Impact | Recovery |
|----------|--------|----------|
| Rule generation fails | DB has rule, Prometheus doesn't | Manual sync API |
| Prometheus down | Rule file created, not loaded | Automatic reload on startup |
| Invalid YAML | No file created | User sees validation error |
| Validation fails | Previous file kept | No impact on running rules |

### 5. Performance Optimization

**For groups with many rules**:
- Batch processing: Generate all rules at once
- Incremental updates: Only affected groups
- Async processing: Background jobs (optional)
- Rate limiting: Prevent rapid successive reloads

**Benchmarks**:
- Target: <5s for 100 rules per group
- Target: <10s for Prometheus reload
- Target: <30s total for create→deploy→activate

---

## Success Criteria

### Functional Requirements
- ✅ Create AlertRule → Prometheus YAML generated automatically
- ✅ Update AlertRule → Prometheus reloads with new rules
- ✅ Delete AlertRule → Rule removed from Prometheus
- ✅ Group with no rules → Rule file deleted
- ✅ Manual sync API → Regenerate all rules
- ✅ Effective rules API → Shows merged configuration

### Technical Requirements
- ✅ Rule generation <5 seconds (100 rules/group)
- ✅ Prometheus reload <10 seconds
- ✅ Zero downtime during updates
- ✅ Validation before deployment
- ✅ Rollback on failure
- ✅ >80% test coverage

### Quality Requirements
- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ No regressions in existing APIs
- ✅ Documentation complete
- ✅ Code review approved

---

## Risk Management

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Prometheus permissions | Medium | High | Docker volume with correct ownership |
| Invalid PromQL generated | Low | High | Multi-layer validation, unit tests |
| File write conflicts | Low | Medium | Atomic writes, file locking |
| Reload failures | Medium | Medium | Retry logic, fallback mechanism |
| Performance issues | Low | Low | Batch processing, async jobs |

---

## Dependencies

### External Dependencies
- Prometheus 2.x+ with reload API enabled
- promtool binary (for validation)
- Shared volume or file system access
- Docker/K8s volume management

### Internal Dependencies
- AlertRule API (completed in Sprint 1-5) ✅
- Group management (completed) ✅
- Service Discovery (completed) ✅
- Error handling (completed in Sprint 6) ✅

---

## Future Enhancements

**Not in scope for Sprint 11, but consider for future**:

1. **Alertmanager Integration** (Sprint 15+)
   - Route configuration management
   - Receiver configuration
   - Inhibition rules

2. **Advanced Features**
   - Rule versioning and history
   - Diff preview before deployment
   - Dry-run mode
   - Rule templates library

3. **Web UI** (Phase 3)
   - Visual rule builder
   - Template editor with preview
   - Deployment dashboard

4. **Monitoring**
   - Metrics for rule generation success rate
   - Prometheus reload metrics
   - Alert on sync failures

---

## References

- [Alerting System Architecture](../../../docs/kr/ALERTING-SYSTEM.md)
- [Prometheus Reload API](https://prometheus.io/docs/prometheus/latest/management_api/)
- [Check Management System](../../../docs/kr/CHECK-MANAGEMENT.md) - Reference implementation
- [Sprint Tracker](../TRACKER.md)

---

## Notes

- This sprint moves from "planned" to "current" status
- Sprint 6 (Error Handling) marked as completed
- Prioritized over Exporter development (Sprint 12-15) due to critical functionality gap
- Group-based alert customization becomes production-ready after completion
- Completes alert system to feature parity with check system
