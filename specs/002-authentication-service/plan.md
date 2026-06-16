# Implementation Plan: Authentication Service

**Branch**: `002-authentication-service` | **Date**: 2026-06-16 | **Spec**: [specs/002-authentication-service/spec.md](spec.md)

**Input**: Feature specification from `/specs/002-authentication-service/spec.md`

## Summary
Implement a secure authentication system for the HROS Admin Portal using JWT and RBAC. This feature includes user management, session handling, and permission-based access control.

## Technical Context

**Language/Version**: Go 1.23

**Primary Dependencies**: Echo, GORM, bcrypt, golang-jwt, Uber Fx

**Storage**: PostgreSQL (Primary), Redis (Optional for token blacklist/cache)

**Testing**: Unit tests for all layers, integration tests for DB and Token logic.

**Constraints**: Clean Architecture; Standard error mapping; Unit-test-per-file.

## Constitution Check

- **I. Clean Architecture**: Separating DB models from domain entities.
- **II. Documentation-First**: OpenAPI contract will be updated for Auth.
- **III. Unit-Test-Per-File**: Every file will have a matching `_test.go`.
- **IV. Task-Driven**: Implementation follows the task list.

## Project Structure (New Files)

```text
internal/
  domain/
    auth/
      entity.go          # AdminUser, Role, SessionToken
      repository.go      # Interfaces for storage
      errors.go          # Auth-specific domain errors
  application/
    auth/
      service.go         # Login, Refresh, Logout use cases
      input.go           # LoginInput, etc.
      output.go          # TokenOutput, etc.
  adapter/
    http/
      auth/
        handler.go       # Login/Refresh/Logout endpoints
        request.go       # DTOs
        response.go      # DTOs
      middleware/
        auth_middleware.go # JWT validation
  infrastructure/
    repository/
      auth/
        model.go         # GORM models
        repository.go    # GORM implementation
```

## Implementation Phases

### Phase 1: Database Foundation
- [ ] Create up/down migrations for `roles`, `role_permissions`, `admin_users`, and `session_tokens`.
- [ ] Seed initial system roles and permissions.

### Phase 2: Domain Layer
- [ ] Define entities and repository interfaces.
- [ ] Implement password hashing logic.

### Phase 3: Infrastructure Layer
- [ ] Implement GORM repositories for AdminUser and SessionToken.
- [ ] Add integration tests for persistence.

### Phase 4: Application Layer
- [ ] Implement AuthService (Login, Refresh, Logout).
- [ ] Implement JWT signing and validation logic.

### Phase 5: Adapter Layer
- [ ] Implement HTTP handlers and DTOs.
- [ ] Update OpenAPI contract.
- [ ] Implement Auth middleware for Echo.

## Complexity Tracking

*No violations detected.*
