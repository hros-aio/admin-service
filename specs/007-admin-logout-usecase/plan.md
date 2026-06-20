# Implementation Plan: Admin Logout Use Case

**Branch**: `007-admin-logout-usecase` | **Date**: 2026-06-19 | **Spec**: [specs/007-admin-logout-usecase/spec.md](specs/007-admin-logout-usecase/spec.md)

**Input**: Feature specification from `/specs/007-admin-logout-usecase/spec.md`

## Summary

Implement `LogoutUseCase` within the application layer. The use case accepts a session token string, validates it, invokes `SessionTokenRepository.DeleteByToken` to delete it from persistence, and emits a `logout.success` event via the `AuditLogger` interface.

## Technical Context

**Language/Version**: Go 1.23

**Primary Dependencies**: Uber Fx (DI), log/slog

**Storage**: PostgreSQL via repository interface `SessionTokenRepository`

**Testing**: Unit tests with testify/mock, aiming for 100% coverage on new use case files.

**Target Platform**: Linux Server / Containerized

**Project Type**: Web Service / Clean Architecture

**Performance Goals**: < 100ms for logout execution.

**Constraints**: Clean Architecture; Standard error mapping; Unit-test-per-file.

## Constitution Check

- **I. Clean Architecture**: UseCase only interacts with repository and audit interfaces. No direct database, HTTP, or framework code.
- **II. Documentation-First**: The API DTO and contract changes are out of scope for this task (handled by adapter layer tasks).
- **III. Unit-Test-Per-File**: `logout_usecase_test.go` will be created with full unit tests.
- **IV. Task-Driven**: Implement only the logout use case logic as specified.

## Project Structure

### Documentation (this feature)

```text
specs/007-admin-logout-usecase/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── checklists/
│   └── requirements.md  # Quality checklist
└── tasks.md             # Task breakdown
```

### Source Code

```text
internal/
  domain/
    auth/
      audit.go           # [MODIFY] Add LogLogoutSuccess to AuditLogger interface
  application/
    usecase/
      logout_types.go    # [NEW] LogoutInput struct
      logout_usecase.go  # [NEW] LogoutUseCase implementation
      logout_usecase_test.go # [NEW] Test suite for LogoutUseCase
    module.go            # [MODIFY] Wire NewLogoutUseCase into Fx application module
  infrastructure/
    auth/
      audit_logger.go    # [MODIFY] Implement LogLogoutSuccess method
```

**Structure Decision**: Standard Clean Architecture structure. Use Case code is placed under `internal/application/usecase/` matching the existing structure of `login_usecase.go`.

## Complexity Tracking

*No violations detected.*
