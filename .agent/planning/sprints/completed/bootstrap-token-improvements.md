# Bootstrap Token System - Future Improvements

## Current State (Sprint 11)

### Implemented Design

**Bootstrap Token Structure**:
```go
type BootstrapToken struct {
    ID             string
    Token          string
    Name           string
    DefaultGroupID string  // Single group assignment
    MaxUses        int
    Uses           int
    ExpiresAt      time.Time
}
```

**Registration Flow**:
```
POST /api/v1/bootstrap/register
{
  "token": "aami_bootstrap_xxx",
  "hostname": "node-01",
  "ip_address": "10.0.1.100"
}

→ Node automatically assigned to token.DefaultGroupID
```

### Current Limitations

1. **One Token Per Group**: Each cluster/group requires a separate bootstrap token
   - Multi-region deployment: Need separate tokens for Seoul, Tokyo, US-West
   - Multi-environment: Need separate tokens for Dev, Staging, Production
   - Token management overhead for large deployments

2. **Fixed Group Assignment**: Nodes cannot choose which group to join
   - Same token cannot be used for different purposes
   - Less flexible for dynamic infrastructure

3. **No Group Constraints**: If we allow request-level group selection, no way to restrict which groups a token can access

## Improvement Proposals

### Proposal 1: Request-Level Group Selection with Constraints

**Goal**: Allow nodes to specify group during registration while maintaining security controls

#### Design

**Token Structure**:
```go
type BootstrapToken struct {
    ID              string
    Token           string
    Name            string
    AllowedGroupIDs []string  // List of groups this token can register to
                              // Empty = token holder chooses any group
                              // Single item = behaves like current DefaultGroupID
    MaxUses         int
    Uses            int
    ExpiresAt       time.Time
}
```

**Registration Request**:
```go
type BootstrapRegisterRequest struct {
    Token     string
    Hostname  string
    IPAddress string
    GroupID   string                 // NEW: Explicit group selection
    Labels    map[string]string
    Metadata  map[string]interface{}
}
```

**Validation Logic**:
```go
func (s *BootstrapTokenService) RegisterNode(req BootstrapRegisterRequest) error {
    token := s.GetToken(req.Token)

    // Validate group selection
    if len(token.AllowedGroupIDs) > 0 {
        // Token has constraints - verify group is allowed
        if !contains(token.AllowedGroupIDs, req.GroupID) {
            return ErrGroupNotAllowed
        }
    }
    // else: No constraints, any group allowed

    // Create target with selected group
    s.targetService.Create(req.Hostname, req.IPAddress, req.GroupID)
}
```

#### Use Cases Enabled

**Multi-Region Deployment**:
```bash
# Create single token for all regions
curl -X POST /api/v1/bootstrap-tokens \
  -d '{
    "name": "global-web-servers",
    "allowed_group_ids": [
      "seoul-datacenter",
      "tokyo-datacenter",
      "us-west-datacenter"
    ],
    "max_uses": 300
  }'

# Deploy to Seoul
terraform apply -var region=seoul -var group_id=seoul-datacenter

# Deploy to Tokyo
terraform apply -var region=tokyo -var group_id=tokyo-datacenter
```

**Multi-Environment Deployment**:
```bash
# Create single token for all environments
curl -X POST /api/v1/bootstrap-tokens \
  -d '{
    "name": "app-servers",
    "allowed_group_ids": [
      "development",
      "staging",
      "production"
    ],
    "max_uses": 500
  }'

# Deploy to production
docker run -e GROUP_ID=production app:latest
```

**Flexible Token (No Constraints)**:
```bash
# Create unrestricted token for testing
curl -X POST /api/v1/bootstrap-tokens \
  -d '{
    "name": "dev-testing-token",
    "allowed_group_ids": [],  // Empty = all groups allowed
    "max_uses": 10
  }'
```

#### Challenges with Own Group

**Problem**: How to handle automatic own group creation?

**Option A: Special sentinel value**
```go
// Request with empty GroupID = create own group
POST /api/v1/bootstrap/register
{
  "token": "xxx",
  "group_id": ""  // Empty = auto-create own group
}

// But: Token needs to allow empty group_id?
AllowedGroupIDs: []  // Allow any group, including auto-created own groups
```

**Option B: Explicit flag**
```go
type BootstrapRegisterRequest struct {
    Token           string
    Hostname        string
    IPAddress       string
    GroupID         string   // Optional if CreateOwnGroup = true
    CreateOwnGroup  bool     // NEW: Explicit flag
}

// Validation:
if req.CreateOwnGroup && token.AllowsOwnGroupCreation {
    // Create "target-{hostname}" group
} else if req.GroupID != "" {
    // Use specified group (must be in AllowedGroupIDs)
}
```

**Option C: Separate endpoint**
```go
// For constrained group assignment
POST /api/v1/bootstrap/register
{
  "token": "xxx",
  "group_id": "seoul-datacenter"
}

// For own group creation (original manual registration)
POST /api/v1/targets
{
  "hostname": "special-server",
  "ip_address": "10.0.1.99"
  // No group_ids = auto-create own group
}
```

**Recommendation**: Option C (keep separate endpoints)
- Bootstrap registration = constrained, group-based
- Manual registration = flexible, supports own groups
- Clearer separation of concerns

#### Migration Path

**Phase 1: Add AllowedGroupIDs (Backward Compatible)**
```go
// Migration: Convert existing tokens
DefaultGroupID → AllowedGroupIDs: [DefaultGroupID]

// Old tokens work exactly the same
token.AllowedGroupIDs = [token.DefaultGroupID]
```

**Phase 2: Add GroupID to Request (Optional)**
```go
// If GroupID not provided, use token.AllowedGroupIDs[0]
if req.GroupID == "" && len(token.AllowedGroupIDs) == 1 {
    req.GroupID = token.AllowedGroupIDs[0]
}
```

**Phase 3: Deprecate DefaultGroupID**
```go
// Remove DefaultGroupID field
// All tokens use AllowedGroupIDs
```

#### API Changes

**Create Token**:
```json
POST /api/v1/bootstrap-tokens
{
  "name": "global-token",
  "allowed_group_ids": ["group1", "group2", "group3"],  // NEW
  "max_uses": 100,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Register Node**:
```json
POST /api/v1/bootstrap/register
{
  "token": "aami_bootstrap_xxx",
  "hostname": "node-01",
  "ip_address": "10.0.1.100",
  "group_id": "group2",  // NEW: Must be in token.allowed_group_ids
  "labels": {...}
}
```

**Error Responses**:
```json
// If group_id not in allowed_group_ids
{
  "error": "Group not allowed for this token",
  "code": "GROUP_NOT_ALLOWED",
  "allowed_groups": ["group1", "group2", "group3"],
  "requested_group": "unauthorized-group"
}
```

### Proposal 2: Token Scopes and Permissions

**Goal**: More granular control over what a token can do

```go
type BootstrapToken struct {
    ID              string
    Token           string
    Scopes          []string  // e.g., ["target:create", "target:update"]
    AllowedGroupIDs []string
    AllowedLabels   map[string][]string  // e.g., {"region": ["seoul", "tokyo"]}
}
```

**Use Case**: Restrict token to specific operations and attributes
```bash
# Token can only create targets in Seoul with specific labels
{
  "scopes": ["target:create"],
  "allowed_group_ids": ["seoul-datacenter"],
  "allowed_labels": {
    "region": ["seoul"],
    "environment": ["production", "staging"]
  }
}
```

### Proposal 3: Dynamic Group Assignment Rules

**Goal**: Automatically assign groups based on node attributes

```go
type BootstrapToken struct {
    ID              string
    Token           string
    GroupAssignmentRules []GroupAssignmentRule
}

type GroupAssignmentRule struct {
    Condition  string  // JSONPath expression: "$.labels.region == 'seoul'"
    GroupID    string  // Group to assign if condition matches
    Priority   int     // Higher priority rules override lower
}
```

**Use Case**: Auto-assign based on instance metadata
```json
POST /api/v1/bootstrap/register
{
  "token": "xxx",
  "hostname": "node-01",
  "labels": {
    "region": "seoul",
    "gpu_count": "8"
  }
}

// Token rules:
[
  {
    "condition": "$.labels.region == 'seoul'",
    "group_id": "seoul-datacenter",
    "priority": 10
  },
  {
    "condition": "$.labels.gpu_count > 0",
    "group_id": "gpu-nodes",
    "priority": 5
  }
]

// Result: Assigned to both "seoul-datacenter" and "gpu-nodes"
```

## Implementation Priority

### High Priority (Next Sprint)
- [ ] **Proposal 1**: Request-level group selection with AllowedGroupIDs
  - Solves immediate multi-region/environment pain points
  - Backward compatible migration path
  - Estimated effort: 3-5 days

### Medium Priority (Sprint +2)
- [ ] **Proposal 2**: Token scopes and permissions
  - Adds security layer
  - Useful for multi-tenant environments
  - Estimated effort: 5-7 days

### Low Priority (Future)
- [ ] **Proposal 3**: Dynamic group assignment rules
  - Advanced feature for complex deployments
  - Requires rule engine implementation
  - Estimated effort: 10-14 days

## Security Considerations

### Token Reuse Attack
**Risk**: Stolen token can register malicious nodes to allowed groups

**Mitigation**:
- Short expiration times (24 hours for deployment windows)
- IP allow-lists (token only valid from specific CIDR blocks)
- Rate limiting (max N registrations per minute)
- Audit logging (track all token usage)

### Group Privilege Escalation
**Risk**: Node uses token to join privileged group

**Mitigation**:
- AllowedGroupIDs restricts which groups token can access
- Critical groups should use separate tokens with strict max_uses
- Regular token rotation

### Credential Leakage
**Risk**: Token exposed in cloud-init logs, Terraform state

**Mitigation**:
- Encrypt Terraform state (backend encryption)
- Use secrets managers (AWS Secrets Manager, HashiCorp Vault)
- Clear tokens from logs after use
- Use cloud-provider native secret injection

## Testing Requirements

### Unit Tests
- [ ] AllowedGroupIDs validation logic
- [ ] Empty AllowedGroupIDs = all groups allowed
- [ ] Single AllowedGroupID = behaves like DefaultGroupID
- [ ] Multiple AllowedGroupIDs with request selection

### Integration Tests
- [ ] Register with valid group (in AllowedGroupIDs)
- [ ] Register with invalid group (not in AllowedGroupIDs) → 400 error
- [ ] Register with empty AllowedGroupIDs (unrestricted) → success
- [ ] Backward compatibility with DefaultGroupID tokens

### E2E Tests
- [ ] Multi-region deployment with single token
- [ ] Multi-environment deployment with single token
- [ ] Token expiration and max_uses enforcement

## Documentation Updates

### User Documentation
- [ ] Update `docs/en/NODE-REGISTRATION.md` with new request format
- [ ] Update `docs/en/API.md` with AllowedGroupIDs field
- [ ] Update `docs/en/CLOUD-INIT.md` with multi-region examples
- [ ] Add security best practices guide

### Agent Documentation
- [ ] Update `.agent/planning/bootstrap-register-implementation.md`
- [ ] Add migration guide for existing tokens
- [ ] Document group selection patterns

## Alternative Approaches Considered

### Approach A: Multiple Tokens Per Deployment
**Current approach**: Use separate tokens for each region/environment

**Pros**:
- Simple implementation (already done)
- Clear security boundaries

**Cons**:
- Token management overhead
- More complex deployment scripts
- Harder to rotate tokens globally

### Approach B: No Token Constraints
**Proposal**: Remove AllowedGroupIDs, let nodes choose any group

**Pros**:
- Maximum flexibility
- Simplest implementation

**Cons**:
- Security risk (stolen token can register to any group)
- No control over node placement
- Not suitable for production

### Approach C: Hybrid (Recommended)
**Proposal 1**: AllowedGroupIDs with request-level selection

**Pros**:
- Balance between flexibility and security
- Backward compatible
- Handles most use cases

**Cons**:
- More complex than current design
- Requires migration effort

## Decision Log

### 2025-12-29: Defer AllowedGroupIDs to Future Sprint

**Decision**: Keep current `DefaultGroupID` design, add group flexibility in future sprint

**Rationale**:
1. Current design is simple and handles primary use case (single-purpose clusters)
2. AllowedGroupIDs adds complexity, especially for own group creation
3. Better to ship working basic feature, iterate based on real usage
4. Can add AllowedGroupIDs later without breaking existing tokens

**Trade-offs Accepted**:
- Multi-region deployments need multiple tokens (acceptable for v1)
- Less flexible than desired (can improve in v2)
- Simpler is better for initial release

**Next Steps**:
- Complete current implementation with DefaultGroupID
- Gather user feedback on token management pain points
- Revisit group flexibility in Sprint 12 or 13
- Design migration path that doesn't break existing deployments

## References

- Implementation Plan: `.agent/planning/bootstrap-register-implementation.md`
- Test Plan: `.agent/planning/bootstrap-register-tests.md`
- Current Code: `services/config-server/internal/service/bootstrap.go`
