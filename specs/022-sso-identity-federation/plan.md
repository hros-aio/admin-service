# Implementation Plan: SSO Identity Federation

**Branch**: `022-sso-identity-federation` | **Date**: 2026-06-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/022-sso-identity-federation/spec.md`

## Summary

This plan outlines the implementation of the Domain and Application Interface definitions, database schema updates, DTO/API contract design, Redis cache layer, and user repository lookup for the SSO Identity Federation.

**Phase 1 (TSK-SSO-001)**: Define `SSOStateCache` interface in `internal/application/interfaces/sso_state_cache.go`. Define domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState` in `internal/domain/errors/auth_errors.go`. Define event payload structs for the `login.sso_success` and `login.sso_failed` audit events in `internal/domain/events/auth_events.go`.

**Phase 2 (TSK-SSO-002)**: Create SQL migration scripts `migrations/000005_add_sso_to_admin_users.up.sql` and `migrations/000005_add_sso_to_admin_users.down.sql` to add SSO mapping fields (`sso_identity_id`, `sso_provider`) to `admin_users` table.

**Phase 3 (TSK-SSO-003)**: Define `SSOCallbackRequest` DTO in `internal/adapter/http/auth/dto/auth_dto.go`. Update `api/openapi.yaml` to document SSO initiation and callback endpoints.

**Phase 4 (TSK-SSO-004)**: Implement `SSOStateCache` using Redis in `internal/infrastructure/cache/sso_state_redis.go`. Implement unit tests with a mocked Redis client in `internal/infrastructure/cache/sso_state_redis_test.go`.

**Phase 5 (TSK-SSO-005)**: Define `FindByEmailOrSSO(ctx, email, ssoID)` in the `AdminUserRepository` domain interface. Implement the method in `GormAdminUserRepository` inside `internal/infrastructure/repository/auth/repository.go`. Add unit tests in `internal/infrastructure/repository/auth/repository_test.go`.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: go-redis (for cache implementation), GORM (for repository mapping).

**Storage**: PostgreSQL (admin_users mapping), Redis (temporary state cache).

## Constitution Check

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Repository interface is defined in `internal/domain/` and implemented in `internal/infrastructure/repository/auth/`. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Specs and plans updated prior to code modification. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Added repository unit tests utilizing `sqlmock`. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Focusing only on task TSK-SSO-005. |
| **V. Observability & Structured Logging** | ✅ PASS | SSO IDs are queried correctly to facilitate structured trace-linking. |

## Project Structure

### Documentation

```text
specs/022-sso-identity-federation/
├── checklists/
│   └── requirements.md  # Quality checklist
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Task definitions
```

### Source Code

```text
api/
└── openapi.yaml                 # OpenAPI specification
internal/
├── adapter/
│   └── http/
│       └── auth/
│           └── dto/
│               ├── auth_dto.go  # DTO structs including SSOCallbackRequest
│               └── auth_dto_test.go # DTO validation unit tests
├── application/
│   └── interfaces/
│       ├── sso_state_cache.go      # SSOStateCache interface
│       └── sso_state_cache_test.go # Unit tests/verifications for SSOStateCache interface
├── domain/
│   ├── admin_user.go               # AdminUserRepository domain interface update
│   ├── errors/
│   │   ├── auth_errors.go          # ErrNoAccountLinked and ErrInvalidSSOState
│   │   └── auth_errors_test.go     # Unit tests for domain errors
│   └── events/
│       ├── auth_events.go          # Event payload structs for SSO success/failure
│       └── auth_events_test.go     # Unit tests for event payloads
├── infrastructure/
│   ├── cache/
│   │   ├── sso_state_redis.go      # Redis implementation of SSOStateCache
│   │   └── sso_state_redis_test.go # Unit tests for Redis cache implementation
│   └── repository/
│       └── auth/
│           ├── repository.go       # GormAdminUserRepository implementation update
│           └── repository_test.go  # Unit tests for GormAdminUserRepository updates
migrations/
├── 000005_add_sso_to_admin_users.up.sql   # SQL migration up script
└── 000005_add_sso_to_admin_users.down.sql # SQL migration down script
test/
└── integration/
    └── sso_migration_test.go       # Integration test for the database migration
```

**Structure Decision**: Clean Architecture database repository and cache layers.
