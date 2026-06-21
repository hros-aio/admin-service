# Implementation Plan: Session Token Repository Updates

**Branch**: `master` | **Date**: 2026-06-21 | **Spec**: [specs/013-session-token-repository/spec.md](spec.md)

**Input**: Feature specification from `/specs/013-session-token-repository/spec.md`

## Summary

Update the `SessionTokenRepository` interface and its GORM implementation to support `UpdateToken` functionality. This facilitates updating existing session properties (like rotated token values, updated expirations, and revocations) in GORM, and validates query execution via `sqlmock` unit tests.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `gorm.io/gorm`, `github.com/DATA-DOG/go-sqlmock` (for testing)

**Storage**: PostgreSQL with GORM

**Testing**: `go test`, unit tests with `sqlmock` to mock database queries.

**Target Platform**: Linux

**Project Type**: Backend Web Service (Database Repository)

**Performance Goals**: <5ms database query execution.

**Constraints**: Clean Architecture. Utilize transaction manager context when loading/saving database transactions.

**Scale/Scope**: Focus exclusively on adding and testing `UpdateToken` method on the GORM repository.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: Domain interface updated in `internal/domain/session_token.go`. GORM implementation updated in `internal/infrastructure/repository/auth/session_token_repository.go`. (PASS)
2. **Documentation-First**: Spec is created. No OpenAPI changes required since no HTTP handlers are added/modified. (PASS)
3. **Unit-Test-Per-File**: `session_token_repository.go` has a corresponding test file `session_token_repository_test.go` which will be updated. (PASS)
4. **Task-Driven**: Plan is strictly scoped to TSK-AUTH-013. (PASS)
5. **Observability**: GORM database error outcomes are returned cleanly to the calling layers. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/013-session-token-repository/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── session_token.go                       # Update interface
└── infrastructure/
    └── repository/
        └── auth/
            ├── session_token_repository.go      # Add UpdateToken implementation
            └── session_token_repository_test.go # Test FindByToken and UpdateToken
```

**Structure Decision**: Place GORM repository updates in `internal/infrastructure/repository/auth/` to align with the existing project structure and dependency injection mapping.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
