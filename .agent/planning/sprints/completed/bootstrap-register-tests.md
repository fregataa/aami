# Bootstrap Register API - Test Plan

## Test Strategy

Tests will be implemented after the core functionality is complete. This document outlines all required tests for the bootstrap register feature.

## Unit Tests

### File: `test/unit/service/bootstrap_service_test.go`

#### Test: `TestBootstrapTokenService_RegisterNode_Success`
- Setup: Create group and valid token
- Act: Call RegisterNode with valid request
- Assert:
  - Target created with correct hostname, IP
  - Target assigned to token's DefaultGroupID
  - Token usage incremented
  - No errors

#### Test: `TestBootstrapTokenService_RegisterNode_TokenNotFound`
- Setup: Use non-existent token
- Act: Call RegisterNode
- Assert: Returns ErrNotFound

#### Test: `TestBootstrapTokenService_RegisterNode_TokenExpired`
- Setup: Create expired token (ExpiresAt in past)
- Act: Call RegisterNode
- Assert: Returns ValidationError with "expired" message

#### Test: `TestBootstrapTokenService_RegisterNode_TokenExhausted`
- Setup: Create token with Uses >= MaxUses
- Act: Call RegisterNode
- Assert: Returns ValidationError with "exhausted" message

#### Test: `TestBootstrapTokenService_RegisterNode_HostnameExists`
- Setup: Create token and existing target with same hostname
- Act: Call RegisterNode with duplicate hostname
- Assert: Returns ErrAlreadyExists

#### Test: `TestBootstrapTokenService_RegisterNode_RollbackOnFailure`
- Setup: Create token, mock target creation to fail
- Act: Call RegisterNode
- Assert:
  - Error returned
  - Token usage count NOT incremented (rollback)

#### Test: `TestBootstrapTokenService_RegisterNode_InvalidDefaultGroup`
- Setup: Create token with non-existent DefaultGroupID
- Act: Call RegisterNode
- Assert: Returns ErrForeignKeyViolation

## Integration Tests

### File: `test/integration/api/bootstrap_register_test.go`

#### Test: `TestBootstrapRegisterAPI_Success`
```go
func TestBootstrapRegisterAPI_Success(t *testing.T) {
    // Setup
    group := createGroup(t, "test-cluster", "production")
    token := createBootstrapToken(t, group.ID, 10)

    // Act
    resp := httpPost("/api/v1/bootstrap/register", map[string]interface{}{
        "token":      token.Token,
        "hostname":   "test-node-01",
        "ip_address": "10.0.1.100",
        "labels": map[string]string{
            "gpu_model": "A100",
            "gpu_count": "8",
        },
    })

    // Assert
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var result dto.BootstrapRegisterResponse
    json.Unmarshal(resp.Body, &result)

    assert.Equal(t, "test-node-01", result.Target.Hostname)
    assert.Equal(t, "10.0.1.100", result.Target.IPAddress)
    assert.Equal(t, 1, len(result.Target.Groups))
    assert.Equal(t, group.ID, result.Target.Groups[0].ID)
    assert.Equal(t, "A100", result.Target.Labels["gpu_model"])
    assert.Equal(t, 1, result.TokenUsage)
    assert.Equal(t, 9, result.RemainingUses)
}
```

#### Test: `TestBootstrapRegisterAPI_ExpiredToken`
- Setup: Create expired token
- Act: POST /api/v1/bootstrap/register
- Assert: 400 Bad Request with TOKEN_EXPIRED code

#### Test: `TestBootstrapRegisterAPI_ExhaustedToken`
- Setup: Create token and use it MaxUses times
- Act: POST /api/v1/bootstrap/register
- Assert: 400 Bad Request with TOKEN_EXHAUSTED code

#### Test: `TestBootstrapRegisterAPI_InvalidToken`
- Setup: Use random token string
- Act: POST /api/v1/bootstrap/register
- Assert: 404 Not Found with TOKEN_NOT_FOUND code

#### Test: `TestBootstrapRegisterAPI_DuplicateHostname`
- Setup: Create token and existing target
- Act: POST /api/v1/bootstrap/register with same hostname
- Assert: 409 Conflict with HOSTNAME_EXISTS code

#### Test: `TestBootstrapRegisterAPI_InvalidRequest_MissingHostname`
- Act: POST /api/v1/bootstrap/register without hostname
- Assert: 400 Bad Request with INVALID_REQUEST code

#### Test: `TestBootstrapRegisterAPI_InvalidRequest_InvalidIP`
- Act: POST /api/v1/bootstrap/register with invalid IP (e.g., "not-an-ip")
- Assert: 400 Bad Request with INVALID_REQUEST code

#### Test: `TestBootstrapRegisterAPI_MultipleNodes_SameToken`
- Setup: Create token with MaxUses=10
- Act: Register 10 different nodes with same token
- Assert:
  - All 10 succeed with 201 Created
  - Token usage increases correctly (1, 2, 3... 10)
  - 11th attempt fails with TOKEN_EXHAUSTED

#### Test: `TestBootstrapRegisterAPI_ConcurrentRegistration`
- Setup: Create token with MaxUses=100
- Act: Register 50 nodes concurrently
- Assert:
  - All 50 succeed
  - Token usage = 50
  - No race conditions or duplicate increments

## End-to-End Tests

### File: `test/e2e/bootstrap_flow_test.go`

#### Test: `TestBootstrapFlow_CompleteWorkflow`
```go
func TestBootstrapFlow_CompleteWorkflow(t *testing.T) {
    // 1. Create namespace
    ns := createNamespace(t, "production")

    // 2. Create group
    group := createGroup(t, "ml-cluster", ns.ID)

    // 3. Create bootstrap token
    token := createBootstrapToken(t, group.ID, 10)

    // 4. Register node via bootstrap
    target := registerNode(t, token.Token, "ml-node-01", "10.0.1.100")

    // 5. Verify target in database
    dbTarget := getTargetFromDB(t, target.ID)
    assert.Equal(t, "ml-node-01", dbTarget.Hostname)

    // 6. Verify group assignment
    groups := getTargetGroups(t, target.ID)
    assert.Equal(t, 1, len(groups))
    assert.Equal(t, group.ID, groups[0].ID)

    // 7. Verify Prometheus service discovery
    sdTargets := getPrometheusServiceDiscovery(t)
    found := false
    for _, sdTarget := range sdTargets {
        if sdTarget.Labels["instance"] == "ml-node-01:9100" {
            found = true
            break
        }
    }
    assert.True(t, found, "Target should appear in Prometheus SD")
}
```

#### Test: `TestBootstrapFlow_BulkDeployment`
- Setup: Create token with MaxUses=100
- Act: Register 100 nodes sequentially
- Assert:
  - All 100 succeed
  - All appear in Prometheus service discovery
  - Token exhausted after 100 uses

## Performance Tests

### File: `test/performance/bootstrap_load_test.go`

#### Test: `TestBootstrapRegister_LoadTest_1000Nodes`
- Setup: Create token with MaxUses=1000
- Act: Register 1000 nodes with 100 concurrent workers
- Measure:
  - Total time to register 1000 nodes
  - Average response time per registration
  - P50, P95, P99 latencies
- Assert:
  - Average < 100ms per registration
  - P99 < 500ms
  - No errors

## Error Handling Tests

### File: `test/integration/api/bootstrap_register_errors_test.go`

#### Test: `TestBootstrapRegister_DatabaseError`
- Setup: Mock database to return error during target creation
- Act: POST /api/v1/bootstrap/register
- Assert:
  - 500 Internal Server Error
  - Token usage NOT incremented (rollback)

#### Test: `TestBootstrapRegister_GroupDeletedDuringRegistration`
- Setup: Create token, delete the DefaultGroupID before registration
- Act: POST /api/v1/bootstrap/register
- Assert: 400 Bad Request with INVALID_GROUP code

## Security Tests

### File: `test/security/bootstrap_register_security_test.go`

#### Test: `TestBootstrapRegister_TokenReuse`
- Setup: Create token with MaxUses=1
- Act: Register 2 nodes with same token
- Assert:
  - First succeeds (201)
  - Second fails (400 TOKEN_EXHAUSTED)

#### Test: `TestBootstrapRegister_ExpiredTokenRejected`
- Setup: Create token that expires in 1 second, wait 2 seconds
- Act: POST /api/v1/bootstrap/register
- Assert: 400 Bad Request with TOKEN_EXPIRED code

#### Test: `TestBootstrapRegister_SQLInjection`
- Act: POST /api/v1/bootstrap/register with SQL injection in hostname
  - `"hostname": "test'; DROP TABLE targets; --"`
- Assert:
  - 400 Bad Request OR 201 Created with sanitized hostname
  - Database NOT corrupted

#### Test: `TestBootstrapRegister_XSS`
- Act: POST /api/v1/bootstrap/register with XSS in labels
  - `"labels": {"note": "<script>alert('xss')</script>"}`
- Assert:
  - 201 Created
  - Label stored as-is (backend doesn't render HTML)

## Regression Tests

### File: `test/regression/bootstrap_register_regression_test.go`

#### Test: `TestManualRegistration_StillCreatesOwnGroup`
- Setup: Create namespace
- Act: POST /api/v1/targets without group_ids
- Assert:
  - 201 Created
  - Target assigned to auto-created "target-{hostname}" group

#### Test: `TestValidateEndpoint_StillWorks`
- Setup: Create token
- Act: POST /api/v1/bootstrap-tokens/validate
- Assert:
  - 200 OK
  - Token usage incremented
  - No target created

## Test Data Fixtures

### File: `test/fixtures/bootstrap.go`

```go
// CreateTestBootstrapToken creates a bootstrap token for testing
func CreateTestBootstrapToken(
    t *testing.T,
    db *gorm.DB,
    groupID string,
    maxUses int,
) *domain.BootstrapToken {
    token := &domain.BootstrapToken{
        ID:             uuid.New().String(),
        Token:          "test_token_" + uuid.New().String(),
        Name:           "test-token",
        DefaultGroupID: groupID,
        MaxUses:        maxUses,
        Uses:           0,
        ExpiresAt:      time.Now().Add(24 * time.Hour),
        Labels:         make(map[string]string),
    }

    err := db.Create(token).Error
    require.NoError(t, err)

    return token
}

// CreateExpiredToken creates an expired token for testing
func CreateExpiredToken(t *testing.T, db *gorm.DB, groupID string) *domain.BootstrapToken {
    token := CreateTestBootstrapToken(t, db, groupID, 10)
    token.ExpiresAt = time.Now().Add(-24 * time.Hour) // Yesterday

    err := db.Save(token).Error
    require.NoError(t, err)

    return token
}

// CreateExhaustedToken creates an exhausted token for testing
func CreateExhaustedToken(t *testing.T, db *gorm.DB, groupID string) *domain.BootstrapToken {
    token := CreateTestBootstrapToken(t, db, groupID, 5)
    token.Uses = 5 // MaxUses reached

    err := db.Save(token).Error
    require.NoError(t, err)

    return token
}
```

## Test Coverage Goals

- **Unit Tests**: 90%+ coverage of service layer
- **Integration Tests**: All API endpoints and error paths
- **E2E Tests**: Complete user workflows
- **Performance Tests**: Baseline performance metrics

## Test Execution Order

1. **Unit Tests** - Fast, no external dependencies
2. **Integration Tests** - Database required
3. **E2E Tests** - Full system required
4. **Performance Tests** - Last, may take time

## CI/CD Integration

```yaml
# .github/workflows/test.yml
test:
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:15
  steps:
    - name: Run Unit Tests
      run: go test ./test/unit/...

    - name: Run Integration Tests
      run: go test ./test/integration/...

    - name: Run E2E Tests
      run: go test ./test/e2e/...

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
```

## Notes

- Tests should be independent (no shared state)
- Use database transactions and rollback for integration tests
- Mock external dependencies (Prometheus, etc.) in unit tests
- Use real database for integration tests
- Performance tests should run in isolated environment
