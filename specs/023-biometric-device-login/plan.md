# Implementation Plan: Biometric Device Login (WebAuthn) - Repository Layer (TSK-BIO-004)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Update the `AdminUserRepository` interface and its GORM implementation `GormAdminUserRepository` with the `UpdateWebAuthnSignCount(ctx, adminID, newCount)` method.
This method performs an atomic PostgreSQL JSONB update to set the `sign_count` value inside the `webauthn_credentials` JSONB column of the `admin_users` table for the specified user. This is crucial for verifying that authenticators are not cloned during subsequent WebAuthn log in handshakes.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `gorm.io/gorm`, `gorm.io/driver/postgres`

**Storage**: PostgreSQL (GORM JSONB update)

**Testing**: `github.com/DATA-DOG/go-sqlmock` for GORM SQL mock tests

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: Sub-millisecond database updates for sign count updates

**Constraints**: Clean Architecture. Interface in `internal/domain/admin_user.go`, implementation in `internal/infrastructure/repository/auth/repository.go` (existing architectural structure), and unit tests in `internal/infrastructure/repository/auth/repository_test.go` using `sqlmock`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Clean Architecture: Database interfaces remain in `internal/domain/`, GORM implementation in `internal/infrastructure/repository/auth/`.
- [x] Documentation-First: Requirements and outcomes documented in `specs/023-biometric-device-login/spec.md`.
- [x] Unit-Test-Per-File: Unit tests for repository methods implemented in `repository_test.go`.
- [x] Task-Driven: Focus strictly on TSK-BIO-004.
- [x] Observability: Structured logging is NOT required at the raw repository level, but any database errors will be propagated.

## Project Structure

### Documentation (this feature)

```text
specs/023-biometric-device-login/
├── plan.md              # This file
├── research.md          # Research (Phase 0)
├── data-model.md        # Data model (Phase 1)
├── quickstart.md        # Quickstart validation guide (Phase 1)
└── tasks.md             # Tasks (Phase 2)
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── admin_user.go                        # AdminUserRepository interface
└── infrastructure/
    └── repository/
        └── auth/
            ├── repository.go                # GormAdminUserRepository implementation
            └── repository_test.go           # Repository Unit Tests (sqlmock)
```

**Structure Decision**: Single project layout matching existing repository structure. Although the task prompt specified `internal/infrastructure/database/admin_user_repository.go`, we will adhere to the existing architecture where `GormAdminUserRepository` is located in `internal/infrastructure/repository/auth/repository.go` to avoid codebase pollution and duplicate code.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
