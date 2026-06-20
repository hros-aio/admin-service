# Implementation Plan: Auth Integration Test

**Branch**: `009-auth-integration-test` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/009-auth-integration-test/spec.md`

## Summary

Write an end-to-end integration test (`test/integration/auth_flow_test.go`) that starts a PostgreSQL database container via `testcontainers-go`, runs raw SQL migrations from the `migrations/` directory against the container, seeds a test admin user (with bcrypt-hashed password and associated roles), boots the Echo server using concrete dependencies, and executes HTTP calls (`POST /v1/auth/login` and `DELETE /v1/auth/session`) to verify the full authentication flow.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/testcontainers/testcontainers-go`, `github.com/testcontainers/testcontainers-go/modules/postgres`, `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/labstack/echo/v4`

**Storage**: PostgreSQL (real containerized instance)

**Testing**: Standard testing package (`testing`), `github.com/stretchr/testify` for assertions.

**Target Platform**: Local/CI Linux env with running Docker daemon.

**Project Type**: Go Clean Architecture Web Service Integration Test.

**Performance Goals**: Integration test completes within 30 seconds.

**Constraints**: Requires access to a local Docker socket (`/var/run/docker.sock`). Requires GORM PostgreSQL driver to connect to the test container.

**Scale/Scope**: Integration testing layer under `test/integration/`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: Yes. The integration test verifies wiring and concrete connections across all layers (adapter -> application -> domain -> database) without mock bypasses.
- **Documentation-First**: Yes. Specification at `spec.md` and this plan are finalized before writing the tests.
- **Unit-Test-Per-File**: N/A. This is an integration test suite under `test/integration/`, which is already a testing target.
- **Observability**: Yes. Server startup and teardown, as well as HTTP request/response exchanges, will be logged.

## Project Structure

### Documentation (this feature)

```text
specs/009-auth-integration-test/
├── plan.md              # This file
├── research.md          # Research/Unknowns
├── data-model.md        # Seeded schemas
└── quickstart.md        # Run command details
```

### Source Code

```text
test/
└── integration/
    └── auth_flow_test.go                 # New integration test file
```

**Structure Decision**: Standard Go integration test layout under `/test/integration/`.

## Complexity Tracking

No violations. The setup uses standard `testcontainers-go` patterns for testing repository/handler integrations.
