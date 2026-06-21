# Implementation Plan: Auth Refresh Handler

**Branch**: `016-auth-refresh-handler` | **Date**: 2026-06-21 | **Spec**: [spec.md](file:///home/ren0503/new-hros/admin-service/specs/016-auth-refresh-handler/spec.md)

**Input**: Feature specification from `/specs/016-auth-refresh-handler/spec.md`

## Summary

This plan outlines the adapter HTTP layer changes required for TSK-AUTH-016.
We will expose the `POST /v1/auth/refresh` endpoint in `AuthHandler`, mapping it to a new `Refresh` handler method.
The handler binds request parameters to `dto.RefreshRequest`, validates them using validator, executes `RefreshSessionUseCase.Execute`, and returns `dto.LoginResponse` on success.
We will map domain validation and token expiration errors to standard 401/403 HTTP status codes.
Additionally, we will update `AuthHandler.Login` to read the `remember_me` parameter from `dto.LoginRequest` and propagate it to `LoginUseCase.Execute`.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Echo Framework, validator/v10, Uber Fx

**Storage**: Redis (implied for blacklist), PostgreSQL (implied for sessions)

**Testing**: Echo test packages (`httptest`), stretchr/testify mocks

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: Sub-10ms handler processing time

**Constraints**: Handlers must not contain business logic; all logic belongs to the use cases.

**Scale/Scope**: Limit changes strictly to the HTTP adapter layer (`auth_handler.go` and `auth_handler_test.go`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: PASS. HTTP Handlers belong to the adapter layer; they invoke UseCases through clean interfaces and map results to HTTP responses.
- **Documentation-First**: PASS. The specification is completed in `spec.md`.
- **Unit-Test-Per-File**: PASS. `auth_handler.go` has a matching `auth_handler_test.go` file. All test scenarios will be covered there.
- **Task-Driven & Atomic Implementation**: PASS. The plan will generate sequential, granular steps.
- **Observability & Structured Logging**: PASS. Uses standard error mapping envelopes and Echo default tracing.

## Project Structure

### Documentation (this feature)

```text
specs/016-auth-refresh-handler/
├── plan.md              # This file
├── research.md          # Handler error mappings and route decisions
├── data-model.md        # Mappings to existing DTO structs
├── quickstart.md        # Scenario verification guide
├── contracts/
│   └── api.md           # API endpoints contracts
└── tasks.md             # Tasks list (generated next)
```

### Source Code (repository root)

```text
internal/
└── adapter/
    └── http/
        ├── auth_handler.go          # Expose refresh endpoint, update login mapping
        └── auth_handler_test.go     # Add endpoint and input/output unit tests
```

**Structure Decision**: Exposing routes directly in the existing `internal/adapter/http/auth_handler.go` adapter layout.

## Complexity Tracking

*No complexity tracking violations.*
