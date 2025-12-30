# Sprint 6: Unified Error Handling

**Status**: 📋 Planned
**Duration**: 8-12 days
**Started**: TBD
**Completed**: TBD

## Goals
Consolidate all error definitions into a centralized package and ensure consistent error handling across all layers. Improve validation error messages to provide detailed field-level feedback for better developer experience.

## Tasks

### Phase 1: Create Unified Error Package (2 days)
- [ ] Create `internal/errors/errors.go`
- [ ] Define domain errors (ErrNotFound, ErrDuplicateKey, ErrAlreadyExists, etc.)
- [ ] Define structured errors (ValidationError, BindingError, DatabaseError)
- [ ] Implement GORM error converter (`FromGormError`)
- [ ] Implement helper functions (IsNotFound, IsConstraintViolation, etc.)
- [ ] Write unit tests for error package

### Phase 2: Update Repository Layer (2-3 days)
- [ ] Convert GORM errors in all 11 repository files
- [ ] ~100 error conversion points
- [ ] Test repository error handling

### Phase 3: Update Service Layer (2-3 days)
- [ ] Remove 62 GORM error conversions
- [ ] Simplify service error handling
- [ ] Test service layer with new errors

### Phase 4: Update Handler Layer (2-3 days)
- [ ] Update `respondError()` function in common.go
- [ ] Map all domain errors to HTTP status codes
- [ ] Replace ~70 direct JSON responses with `respondError()`
- [ ] **Improve validation error messages with field-level details**
  - [ ] Include missing required fields in error response
  - [ ] Show validation constraints (min/max length, allowed values)
  - [ ] Return structured error with field names and violation details
  - [ ] Example: `{"error": "validation failed", "fields": {"merge_strategy": "required, must be one of: override, merge, append"}}`
- [ ] Test all API endpoints

### Phase 5: Cleanup (1 day)
- [ ] Delete `internal/service/errors.go`
- [ ] Delete `internal/util/errors.go`
- [ ] Remove duplicate ValidationError definitions
- [ ] Update imports
- [ ] Final integration testing

## Deliverables
- Single source of truth for all errors
- Consistent error responses across APIs
- Better error messages with context
- **Detailed validation error messages with field-level information**
- Database-agnostic error handling
- ~232 error handling updates

## Benefits
1. Type-safe error checking with `errors.Is()`
2. No GORM errors leaking to service layer
3. Consistent API error codes
4. **Improved developer experience with detailed validation errors**
5. Easy to mock and test
6. Can replace GORM without changing services

## Reference
- Detailed implementation plan: `/Users/sh/.claude/plans/transient-painting-key.md`

## Notes
- Critical files: errors.go, target.go (repo & service), common.go (handler)
- Safe migration with rollback points after each phase
- Keep old error definitions until Phase 5 for safety

## Known Issues to Address
- **Validation error messages lack detail** (discovered 2024-12-29)
  - Current: `{"error":"failed to bind request","code":"BINDING_ERROR"}`
  - Desired: `{"error":"validation failed","code":"VALIDATION_ERROR","fields":{"merge_strategy":"required field missing, must be one of: override, merge, append"}}`
  - Impact: Users cannot determine which field is invalid or what the requirements are
  - Solution: Implement structured validation error responses in Phase 4
