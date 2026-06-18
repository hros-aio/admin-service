# Implementation Plan: Session Token Repository (GORM)

**Branch**: `master` | **Date**: 2026-06-18 | **Spec**: [specs/005-session-token-repository/spec.md](spec.md)

**Input**: Feature specification from `/specs/005-session-token-repository/spec.md`

## Summary

Implement the `SessionTokenRepository` interface using GORM to provide persistent storage for session tokens (refresh tokens). The implementation will be housed within the existing `auth` repository package and will support saving new tokens and deleting tokens by their value. This is a critical component for the authentication service's login and logout flows.

## Technical Context

**Language/Version**: Go 1.23

**Primary Dependencies**: GORM v1.25+, PostgreSQL driver, Uber Fx, sqlmock (for testing)

**Storage**: PostgreSQL

**Testing**: `go test`, table-driven unit tests with `sqlmock`, integration tests (if required by infra changes)

**Target Platform**: Linux (Dockerized environment)

**Project Type**: Backend Web Service (Infrastructure Layer)

**Performance Goals**: <50ms p95 for token operations (indexed relational queries)

**Constraints**: Strict Clean Architecture boundaries; GORM models must not leak into domain/application layers.

**Scale/Scope**: 100% unit test coverage for new repository methods.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: Implementation will be confined to `internal/infrastructure/repository/auth/`. No Echo or GORM imports in domain/application. (PASS)
2. **Documentation-First**: `spec.md` is complete and clarified. No new REST endpoints in this task. (PASS)
3. **Unit-Test-Per-File**: `session_token_repository_test.go` will be created alongside the implementation. (PASS)
4. **Task-Driven**: Plan is scoped strictly to TSK-AUTH-005. (PASS)
5. **Observability**: Structured logging using `slog` will be used for unexpected database errors. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/005-session-token-repository/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (generated later)
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── session_token.go           # Update with DeleteByToken interface
└── infrastructure/
    └── repository/
        └── auth/
            ├── mapper.go          # Add session token mappers
            ├── model.go           # Add session token GORM model
            ├── session_token_repository.go      # Implementation
            └── session_token_repository_test.go # Unit tests
```

**Structure Decision**: Following the established `internal/infrastructure/repository/auth/` pattern used by the `AdminUserRepository`. This ensures all auth-related persistence logic is co-located and shares models/mappers.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
