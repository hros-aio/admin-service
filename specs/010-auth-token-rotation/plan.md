# Implementation Plan: Auth Token Rotation

**Branch**: `master` | **Date**: 2026-06-21 | **Spec**: [specs/010-auth-token-rotation/spec.md](spec.md)

**Input**: Feature specification from `/specs/010-auth-token-rotation/spec.md`

## Summary

Update the `SessionToken` domain entity in `internal/domain/session_token.go` to support explicit expiration tracking and add a `Rotate()` helper method. The `Rotate()` method will utilize standard library cryptography to generate new token strings and update the token state in-place. Additionally, define a `TokenBlacklist` cache interface in the new `internal/application/interfaces/cache.go` file for immediate revocation querying, and expose specific domain errors in `internal/domain/errors/auth_errors.go`.

## Technical Context

**Language/Version**: Go 1.23

**Primary Dependencies**: Go standard library (`crypto/rand`, `encoding/hex`), `github.com/stretchr/testify` (for unit testing)

**Storage**: N/A (this task is restricted to Domain models/errors and Application interface definitions. No database queries are performed.)

**Testing**: `go test`, table-driven unit tests

**Target Platform**: Linux (Dockerized environment)

**Project Type**: Backend Web Service (Domain and Application Layers)

**Performance Goals**: <1ms for token generation and state rotation (CPU-bound only, no I/O)

**Constraints**: Strict Clean Architecture layers. The domain and application layers must not import any third-party framework or database logic (GORM, Echo, etc.).

**Scale/Scope**: 100% unit test coverage for the modified domain code.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: Changes are restricted to domain models (`internal/domain`), application interfaces (`internal/application/interfaces`), and domain errors (`internal/domain/errors`). No Echo or GORM concrete models are imported. (PASS)
2. **Documentation-First**: Spec is created, validated, and finalized. No public API routes are modified in this slice, so OpenAPI changes are not required. (PASS)
3. **Unit-Test-Per-File**: Modified source files `session_token.go` and `auth_errors.go` have corresponding `_test.go` files which will be updated. The new `interfaces/cache.go` file contains only type interface definitions and is exempt. (PASS)
4. **Task-Driven**: Plan is strictly scoped to TSK-AUTH-010. (PASS)
5. **Observability**: Cryptographic rand errors are caught and propagated cleanly. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/010-auth-token-rotation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code (repository root)

```text
internal/
├── domain/
│   ├── errors/
│   │   ├── auth_errors.go       # Add ErrInvalidRefreshToken
│   │   └── auth_errors_test.go  # Update unit tests
│   ├── session_token.go         # Update with Rotate() method
│   └── session_token_test.go    # Update unit tests
└── application/
    └── interfaces/
        └── cache.go             # Define TokenBlacklist cache interface
```

**Structure Decision**: Following the clean architecture template structure. The interface definition `cache.go` is placed inside `internal/application/interfaces` to separate domain-independent interfaces from infrastructure implementations.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
