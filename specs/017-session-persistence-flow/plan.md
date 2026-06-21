# Implementation Plan: Session Persistence Flow Test

**Branch**: `017-session-persistence-flow` | **Date**: 2026-06-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/017-session-persistence-flow/spec.md`

## Summary

This plan outlines the testing infrastructure and adapter changes required to verify the complete remember-me login, token refresh, and access token blacklist-on-logout integration flow (TSK-AUTH-017).
We will spin up PostgreSQL and Redis test containers using `testcontainers-go`.
Additionally, we will update the `Logout` handler in `auth_handler.go` to support receiving both the access token and the refresh token simultaneously, enabling full verification of the blacklist functionality during integration testing.

## Technical Context

**Language/Version**: Go 1.23+ (using project's Go 1.26.1)

**Primary Dependencies**: Echo Framework, GORM, go-redis/v9, testcontainers-go (PostgreSQL & Generic Redis)

**Storage**: PostgreSQL, Redis

**Testing**: Standard library testing with `testcontainers-go`, `testify/assert`, and `testify/require`

**Target Platform**: Linux server / local developer machines running Docker

**Project Type**: web-service (integration testing)

**Performance Goals**: Test container startup and execution under 25s

**Constraints**:
- Must not leak containers; all containers must be properly terminated on test completion.
- Handlers must remain backwards-compatible for logout tests that only send the refresh token in the `Authorization` header.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: PASS. The integration test verifies the fully assembled application layer wired via Fx, maintaining correct dependency directions.
- **Documentation-First**: PASS. The specification is completed in `spec.md`.
- **Unit-Test-Per-File**: PASS. Integration tests belong to the `test/integration/` package and verify multi-component behaviors end-to-end.
- **Task-Driven & Atomic Implementation**: PASS. The tasks will be written to `tasks.md` sequentially.
- **Observability & Structured Logging**: PASS. Logging behavior is verified through standard slog outputs discarded/monitored during test runs.

## Project Structure

### Documentation (this feature)

```text
specs/017-session-persistence-flow/
├── plan.md              # This file
├── research.md          # Container selection and handler updates
├── data-model.md        # Database schema/cache key assertions
├── quickstart.md        # Scenario verification guide
├── contracts/
│   └── api.md           # Extended Logout header details
└── tasks.md             # Tasks list (generated next)
```

### Source Code (repository root)

```text
internal/
└── adapter/
    └── http/
        └── auth_handler.go                       # Update Logout to support X-Refresh-Token header

test/
└── integration/
    └── session_persistence_flow_test.go         # Implementation of integration tests
```

**Structure Decision**: Added new integration test in the existing `test/integration` folder and updated `internal/adapter/http/auth_handler.go` to populate the `AccessToken` parameter on logout.

## Complexity Tracking

*No violations found.*
