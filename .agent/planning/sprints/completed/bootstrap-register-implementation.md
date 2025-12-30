# Bootstrap Register API Implementation Plan

## Problem Statement

### Current Conflict

**BootstrapToken Domain**:
- Has `DefaultGroupID` field (required, non-nullable)
- Purpose: Auto-assign nodes to a specific group during bootstrap registration

**Target Service Logic**:
- When `GroupIDs` is empty → Auto-creates "target-{hostname}" own group
- When `GroupIDs` is provided → Uses provided groups

**The Conflict**:
- If we use `token.DefaultGroupID` → Own group won't be created
- If we let own group be created → `token.DefaultGroupID` is ignored
- User feedback: "굉장히 헷갈릴 여지가 있다" (very confusing)

### Missing Implementation

`POST /api/v1/bootstrap/register` endpoint is documented but **not implemented**:
- Documentation: `docs/en/NODE-REGISTRATION.md:154` references this endpoint
- Reality: Only `POST /api/v1/bootstrap-tokens/validate` exists (doesn't register nodes)

## Design Decisions

### Decision 1: DefaultGroupID vs Own Group

**Recommendation: Use DefaultGroupID only (no auto own-group creation)**

**Rationale**:
1. **Semantic clarity**: Bootstrap tokens are for **bulk deployment** where all nodes should join the **same logical group** (e.g., "ml-training-cluster")
2. **Own groups are for manual registration**: Individual servers that need isolated configuration
3. **Admin control**: Token creator explicitly decides which group nodes should join
4. **Scalability**: Creating 100 own groups for 100 nodes in same cluster is wasteful

**Alternative Considered**: Make DefaultGroupID optional
- ❌ Rejected: Breaks existing token creation (field is currently required)
- ❌ Requires migration of existing tokens
- ❌ Adds complexity to decide "use DefaultGroupID or create own group?"

### Decision 2: Bootstrap Register Behavior

**Behavior**:
```
Bootstrap Register
    ├─ Validate token (check expiry, usage limit)
    ├─ Create target
    ├─ Assign to token.DefaultGroupID (single group only)
    └─ Increment token usage counter
```

**Differences from Manual Registration**:

| Aspect | Manual (POST /targets) | Bootstrap (POST /bootstrap/register) |
|--------|------------------------|--------------------------------------|
| **Auth** | None (future: API key) | Bootstrap token |
| **Groups** | Multiple or none (→ own group) | Single (token's DefaultGroupID) |
| **Use Case** | Existing servers, custom setup | New nodes, bulk deployment |
| **Auto-discovery** | No (admin provides info) | Yes (script collects info) |

### Decision 3: Request Structure

**New DTO**: `BootstrapRegisterRequest`

```go
type BootstrapRegisterRequest struct {
    Token     string                 `json:"token" binding:"required"`
    Hostname  string                 `json:"hostname" binding:"required,min=1,max=255"`
    IPAddress string                 `json:"ip_address" binding:"required,ip"`
    Labels    map[string]string      `json:"labels,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

**Note**: No `GroupIDs` field (uses token's DefaultGroupID internally)

## Implementation Plan

### Phase 1: DTO and Domain Changes

#### File: `internal/api/dto/bootstrap.go`

**Add new DTO**:

```go
// BootstrapRegisterRequest represents a request to register a new node using bootstrap token
type BootstrapRegisterRequest struct {
    Token     string                 `json:"token" binding:"required"`
    Hostname  string                 `json:"hostname" binding:"required,min=1,max=255"`
    IPAddress string                 `json:"ip_address" binding:"required,ip"`
    Labels    map[string]string      `json:"labels,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BootstrapRegisterResponse represents the response for node registration
type BootstrapRegisterResponse struct {
    Target       TargetResponse `json:"target"`
    TokenUsage   int            `json:"token_usage"`    // Current usage count
    RemainingUses int           `json:"remaining_uses"` // Remaining usage count
}
```

### Phase 2: Service Layer

#### Option A: Add method to BootstrapTokenService (Recommended)

**File**: `internal/service/bootstrap.go`

```go
// RegisterNode validates token and creates target with default group assignment
func (s *BootstrapTokenService) RegisterNode(
    ctx context.Context,
    req dto.BootstrapRegisterRequest,
) (*domain.Target, *domain.BootstrapToken, error) {
    // 1. Validate and use token
    token, err := s.ValidateAndUse(ctx, dto.ValidateTokenRequest{
        Token: req.Token,
    })
    if err != nil {
        return nil, nil, err
    }

    // 2. Create target with DefaultGroupID
    // IMPORTANT: Pass token.DefaultGroupID explicitly
    target, err := s.targetService.Create(ctx, dto.CreateTargetRequest{
        Hostname:  req.Hostname,
        IPAddress: req.IPAddress,
        GroupIDs:  []string{token.DefaultGroupID}, // ← Use token's group
        Labels:    req.Labels,
        Metadata:  req.Metadata,
    })
    if err != nil {
        // Rollback: Decrement token usage on failure
        token.Uses--
        _ = s.tokenRepo.Update(ctx, token)
        return nil, nil, err
    }

    return target, token, nil
}
```

**Dependencies**: Need to inject `TargetService`

```go
type BootstrapTokenService struct {
    tokenRepo     repository.BootstrapTokenRepository
    groupRepo     repository.GroupRepository
    targetService *TargetService // ← Add this
}

func NewBootstrapTokenService(
    tokenRepo repository.BootstrapTokenRepository,
    groupRepo repository.GroupRepository,
    targetService *TargetService, // ← Add parameter
) *BootstrapTokenService {
    return &BootstrapTokenService{
        tokenRepo:     tokenRepo,
        groupRepo:     groupRepo,
        targetService: targetService,
    }
}
```

#### Option B: Keep services separate (Alternative)

Handler calls both services directly (simpler, no circular dependency):

```go
// In handler:
token, err := h.bootstrapService.ValidateAndUse(ctx, ...)
target, err := h.targetService.Create(ctx, dto.CreateTargetRequest{
    GroupIDs: []string{token.DefaultGroupID},
    ...
})
```

**Recommendation**: Use Option A for better encapsulation and transaction handling

### Phase 3: Handler Layer

#### File: `internal/api/handler/bootstrap.go`

**Add new handler method**:

```go
// RegisterNode handles POST /api/v1/bootstrap/register
// Registers a new node using bootstrap token
func (h *BootstrapHandler) RegisterNode(c *gin.Context) {
    var req dto.BootstrapRegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, domainerrors.NewBindingError(err))
        return
    }

    ctx := c.Request.Context()

    // Register node with bootstrap token
    target, token, err := h.bootstrapService.RegisterNode(ctx, req)
    if err != nil {
        respondError(c, err)
        return
    }

    // Build response
    resp := dto.BootstrapRegisterResponse{
        Target:        dto.ToTargetResponse(target),
        TokenUsage:    token.Uses,
        RemainingUses: token.RemainingUses(),
    }

    c.JSON(http.StatusCreated, resp)
}
```

### Phase 4: Router Registration

#### File: `internal/api/router.go`

**Add route**:

```go
// Bootstrap tokens
bootstrapTokens := api.Group("/bootstrap-tokens")
{
    bootstrapTokens.GET("", bootstrapTokenHandler.List)
    bootstrapTokens.POST("", bootstrapTokenHandler.Create)
    bootstrapTokens.GET("/:id", bootstrapTokenHandler.GetByID)
    bootstrapTokens.PUT("/:id", bootstrapTokenHandler.Update)
    bootstrapTokens.DELETE("/:id", bootstrapTokenHandler.Delete)

    // Token validation
    bootstrapTokens.POST("/validate", bootstrapTokenHandler.ValidateAndUse)

    // Node registration (NEW)
    bootstrapTokens.POST("/register", bootstrapTokenHandler.RegisterNode)

    // Token management
    bootstrapTokens.POST("/delete", bootstrapTokenHandler.DeleteResource)
    bootstrapTokens.POST("/purge", bootstrapTokenHandler.PurgeResource)
    bootstrapTokens.POST("/restore", bootstrapTokenHandler.RestoreResource)
}
```

### Phase 5: Service Initialization

#### File: `internal/api/server.go` or `cmd/config-server/main.go`

**Update service initialization** (add circular dependency resolution):

```go
// Create repositories
targetRepo := repository.NewTargetRepository(db)
tokenRepo := repository.NewBootstrapTokenRepository(db)
groupRepo := repository.NewGroupRepository(db)

// Create target service first (no bootstrap dependency)
targetService := service.NewTargetService(targetRepo, targetGroupRepo, groupRepo, namespaceRepo)

// Create bootstrap service with target service
bootstrapService := service.NewBootstrapTokenService(tokenRepo, groupRepo, targetService)

// Create handlers
bootstrapHandler := handler.NewBootstrapHandler(bootstrapService)
```

### Phase 6: Testing

#### File: `test/integration/api/bootstrap_register_test.go` (New)

**Test cases**:

1. **Success**: Valid token, new hostname
   ```go
   func TestBootstrapRegister_Success(t *testing.T) {
       // Setup: Create group and token
       token := createBootstrapToken(t, group.ID, 10)

       // Act: Register node
       resp := POST("/api/v1/bootstrap/register", {
           "token": token.Token,
           "hostname": "test-node-01",
           "ip_address": "10.0.1.100",
       })

       // Assert
       assert.Equal(t, 201, resp.StatusCode)
       assert.Equal(t, "test-node-01", resp.Target.Hostname)
       assert.Equal(t, group.ID, resp.Target.Groups[0].ID)
       assert.Equal(t, 1, resp.TokenUsage)
       assert.Equal(t, 9, resp.RemainingUses)
   }
   ```

2. **Expired Token**: Token expired
3. **Exhausted Token**: Token max uses reached
4. **Duplicate Hostname**: Hostname already exists
5. **Invalid Group**: Token's DefaultGroupID doesn't exist (shouldn't happen but test for safety)
6. **Rollback on Failure**: Token usage should rollback if target creation fails

#### File: `test/integration/service/bootstrap_service_test.go`

**Add test for RegisterNode method**

## Error Handling

### Errors to Handle

1. **Token validation**:
   - `ErrNotFound`: Token doesn't exist
   - `ValidationError`: Token expired or exhausted

2. **Target creation**:
   - `ErrAlreadyExists`: Hostname already registered
   - `ErrForeignKeyViolation`: DefaultGroupID doesn't exist

3. **Rollback scenarios**:
   - If target creation fails after token usage increment → Decrement usage count

### Response Codes

| Error | HTTP Status | Code |
|-------|-------------|------|
| Token not found | 404 | `TOKEN_NOT_FOUND` |
| Token expired | 400 | `TOKEN_EXPIRED` |
| Token exhausted | 400 | `TOKEN_EXHAUSTED` |
| Hostname exists | 409 | `HOSTNAME_EXISTS` |
| Invalid group | 400 | `INVALID_GROUP` |
| Binding error | 400 | `INVALID_REQUEST` |

## API Documentation

### Endpoint Specification

```yaml
POST /api/v1/bootstrap/register
Summary: Register a new node using bootstrap token
Description: |
  Validates bootstrap token and automatically registers the node to the
  token's default group. This is the automated node registration endpoint
  used by cloud-init scripts and bootstrap agents.

Request:
  Content-Type: application/json
  Body:
    token: string (required) - Bootstrap token
    hostname: string (required, max 255) - Node hostname
    ip_address: string (required, valid IP) - Node IP address
    labels: object (optional) - Custom labels
    metadata: object (optional) - Custom metadata

Response:
  Status: 201 Created
  Body:
    target: object - Created target details with groups
    token_usage: int - Current token usage count
    remaining_uses: int - Remaining uses for the token

Errors:
  400 Bad Request - Invalid request, token expired/exhausted
  404 Not Found - Token doesn't exist
  409 Conflict - Hostname already exists
```

### Example Request/Response

**Request**:
```bash
curl -X POST http://config-server:8080/api/v1/bootstrap/register \
  -H "Content-Type: application/json" \
  -d '{
    "token": "aami_bootstrap_xxxxxxxxxxxxxxxxxxxxxxxx",
    "hostname": "ml-node-01",
    "ip_address": "10.0.1.100",
    "labels": {
      "datacenter": "us-east-1",
      "gpu_model": "A100",
      "gpu_count": "8"
    }
  }'
```

**Response** (201 Created):
```json
{
  "target": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "hostname": "ml-node-01",
    "ip_address": "10.0.1.100",
    "groups": [
      {
        "id": "group-123",
        "name": "ml-training-cluster",
        "namespace": "production"
      }
    ],
    "status": "active",
    "labels": {
      "datacenter": "us-east-1",
      "gpu_model": "A100",
      "gpu_count": "8"
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "token_usage": 1,
  "remaining_uses": 99
}
```

## Documentation Updates

### Files to Update

1. **`docs/en/NODE-REGISTRATION.md`**:
   - Update bootstrap flow to reflect actual implementation
   - Add API endpoint details
   - Update example scripts

2. **`docs/en/API.md`**:
   - Add `POST /api/v1/bootstrap/register` documentation
   - Add request/response examples
   - Add error codes

3. **`docs/en/CLOUD-INIT.md`**:
   - Update bootstrap script to use new endpoint
   - Update example cloud-init configurations

## Migration Considerations

### Existing Tokens

**No migration needed**: Existing tokens already have `DefaultGroupID` (required field)

### Existing Scripts

**Update required**: Bootstrap scripts in the wild may need updating to call the new endpoint

**Backward compatibility**: Keep `/validate` endpoint for scripts that only validate tokens

## Implementation Checklist

### Phase 1: Core Implementation
- [ ] Add `BootstrapRegisterRequest` DTO to `internal/api/dto/bootstrap.go`
- [ ] Add `BootstrapRegisterResponse` DTO to `internal/api/dto/bootstrap.go`
- [ ] Add `RegisterNode()` method to `BootstrapTokenService`
- [ ] Update `NewBootstrapTokenService` constructor to accept `TargetService`
- [ ] Add `RegisterNode()` handler method to `internal/api/handler/bootstrap.go`
- [ ] Register route in `internal/api/router.go`
- [ ] Update service initialization to resolve dependency

### Phase 2: Testing
- [ ] Write unit tests for `BootstrapTokenService.RegisterNode()`
- [ ] Write integration tests for API endpoint
- [ ] Test token expiration handling
- [ ] Test token exhaustion handling
- [ ] Test hostname conflict handling
- [ ] Test rollback on failure

### Phase 3: Documentation
- [ ] Update `docs/en/NODE-REGISTRATION.md`
- [ ] Update `docs/en/API.md`
- [ ] Update `docs/en/CLOUD-INIT.md`
- [ ] Add code examples to documentation

### Phase 4: Validation
- [ ] Manual testing with real bootstrap flow
- [ ] Test with cloud-init scripts
- [ ] Verify Prometheus service discovery integration
- [ ] Load test with 100 concurrent registrations

## Security Considerations

1. **Token reuse**: Tokens can be used multiple times up to `max_uses`
   - Consider: Should each token create only one node?
   - Current design: Allows bulk deployment (100 nodes with one token)

2. **Token exposure**: Tokens in cloud-init scripts may be visible in instance metadata
   - Mitigation: Short expiration times (e.g., 24 hours for deployment window)
   - Mitigation: One-time use tokens for sensitive environments

3. **Hostname spoofing**: Malicious node could register with fake hostname
   - Current: No hostname verification
   - Future: Consider DNS verification or certificate-based auth

## Future Enhancements

### High Priority (Next Sprint)

1. **Flexible Group Selection**: Allow nodes to choose group during registration
   - See: `.agent/planning/bootstrap-token-improvements.md` (Proposal 1)
   - Token with `AllowedGroupIDs` instead of single `DefaultGroupID`
   - Registration request includes `group_id` field
   - Solves multi-region and multi-environment deployment challenges

### Medium Priority

2. **Auto-detect hardware**: Service auto-populates labels from metadata
   - GPU model, count
   - CPU cores, memory
   - Disk capacity

3. **Bootstrap script endpoint**: Serve customized bootstrap scripts
   ```bash
   GET /api/v1/bootstrap/script?token=xxx
   → Returns: Customized bash script with token embedded
   ```

4. **Registration webhooks**: Notify external systems on new node registration
   - Slack notification
   - Service mesh integration
   - CMDB update

### Low Priority

5. **Dynamic group assignment rules**: Allow multiple groups based on labels
   - See: `.agent/planning/bootstrap-token-improvements.md` (Proposal 3)
   - Rule-based automatic group assignment
   - Complex deployment scenarios

6. **Token scopes and permissions**: Granular control over token capabilities
   - See: `.agent/planning/bootstrap-token-improvements.md` (Proposal 2)
   - Multi-tenant security
   - Advanced access control

## Summary

### What Changes

1. **New endpoint**: `POST /api/v1/bootstrap/register`
2. **New DTOs**: `BootstrapRegisterRequest`, `BootstrapRegisterResponse`
3. **New method**: `BootstrapTokenService.RegisterNode()`
4. **Behavior change**: Bootstrap registration uses `DefaultGroupID` only (no auto own-group)

### What Stays the Same

1. **Manual registration**: `POST /api/v1/targets` still creates own groups when no groups provided
2. **Token validation**: `POST /api/v1/bootstrap-tokens/validate` still works independently
3. **Token structure**: No changes to BootstrapToken domain model

### Clear Separation

| Registration Type | Endpoint | Groups | Use Case |
|------------------|----------|--------|----------|
| **Bootstrap** | `POST /bootstrap/register` | Single (token's DefaultGroupID) | New nodes, bulk deployment |
| **Manual** | `POST /targets` | Multiple or none (→ own group) | Existing servers, custom setup |

This design eliminates confusion by making the bootstrap flow explicit and separate from manual registration.
