# Implementation Plan: Refresh Session Use Case

**Branch**: `master` | **Date**: 2026-06-21 | **Spec**: [specs/014-refresh-session-usecase/spec.md](spec.md)

**Input**: Feature specification from `/specs/014-refresh-session-usecase/spec.md`

## Summary

Implement the `RefreshSessionUseCase` in the application layer. The use case accepts a refresh token string, fetches the session using `SessionTokenRepository.FindByToken`, validates its expiry and revocation status, generates new access/refresh tokens using `TokenProvider`, updates the session in the database via `SessionTokenRepository.UpdateToken`, and publishes the audit event.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/stretchr/testify` (for unit tests)

**Storage**: GORM Postgres (via repository interfaces)

**Testing**: `go test`, unit tests with mocks for dependencies (`SessionTokenRepository`, `TokenProvider`, `AuditLogger`).

**Target Platform**: Linux

**Project Type**: Backend Web Service (Application/UseCase Layer)

**Performance Goals**: Usecase logic execution time <1ms (excluding mock I/O).

**Constraints**: Clean Architecture. Do not access GORM or Redis directly from the usecase.

**Scale/Scope**: Focus strictly on the use case business logic, inputs, outputs, and unit testing.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: Changes are restricted to application use cases (`internal/application/usecase`), domain `AuditLogger` interface (`internal/domain/auth`), and infrastructure `SlogAuditLogger` implementation (`internal/infrastructure/auth`). (PASS)
2. **Documentation-First**: Spec is finalized. No REST HTTP handler or OpenAPI updates are required in this slice; they will be addressed in a subsequent handler slice. (PASS)
3. **Unit-Test-Per-File**: `refresh_session_usecase.go` will have its corresponding `refresh_session_usecase_test.go`. (PASS)
4. **Task-Driven**: Plan is strictly scoped to TSK-AUTH-014. (PASS)
5. **Observability**: Usecase logs `session.refreshed` through the injected `AuditLogger`. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/014-refresh-session-usecase/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── auth/
│       └── audit.go                        # Update domain interface to add LogSessionRefreshed
├── infrastructure/
│   └── auth/
│       ├── audit_logger.go                 # Implement LogSessionRefreshed using slog
│       └── audit_logger_test.go            # Update test to cover new log event
└── application/
    └── usecase/
        ├── refresh_session_usecase.go      # New usecase implementation
        └── refresh_session_usecase_test.go # New unit tests
```

**Structure Decision**: Placed `RefreshSessionUseCase` inside the application usecase package `internal/application/usecase` to separate orchestration from HTTP delivery/adapters.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
