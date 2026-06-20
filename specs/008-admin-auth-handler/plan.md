# Implementation Plan: Admin Auth Handler

**Branch**: `008-admin-auth-handler` | **Date**: 2026-06-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-admin-auth-handler/spec.md`

## Summary

Implement the HTTP adapter layer for admin authentication (`AuthHandler`) to handle `POST /v1/auth/login` and `DELETE /v1/auth/session` endpoints. The handler binds incoming requests to DTOs, validates inputs using `validator`, executes the relevant UseCases (`LoginUseCase` and `LogoutUseCase`), and returns formatted JSON responses or error envelopes.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Echo Framework v4, Uber Fx, `go-playground/validator/v10`

**Storage**: PostgreSQL (persisted session tokens and user accounts, accessed via repository interfaces inside use cases)

**Testing**: Echo `httptest` utilities, `testify/assert`, `testify/mock`

**Target Platform**: Linux server

**Project Type**: Clean Architecture Go Web Service

**Performance Goals**: Sub-50ms handler processing overhead.

**Constraints**: Handlers contain no business logic. CENTRALIZED/handler-level error mapping. Unit test per file (MANDATORY). Coverage targets: Adapter layer 75%.

**Scale/Scope**: Adapter layer routes, JSON binding, DTO definition, and Fx DI wiring.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: Yes. The HTTP adapter layer depends only on the Application layer UseCases and domain errors. No GORM or DB details are exposed.
- **Documentation-First**: Yes. API contract is defined in `api/openapi.yaml`, and feature specification is at `spec.md`.
- **Unit-Test-Per-File**: Yes. `auth_handler_test.go` will be created alongside `auth_handler.go` to cover all scenarios.
- **Observability**: Yes. Standard HTTP request logging via middleware, and structured error responses include trace/request ID.

## Project Structure

### Documentation (this feature)

```text
specs/008-admin-auth-handler/
├── plan.md              # This file
├── research.md          # Research/Unknowns
├── data-model.md        # Request/Response schemas
└── quickstart.md        # Manual verification commands
```

### Source Code

```text
internal/
├── app/
│   └── app.go                            # Register the HTTP adapter module
├── application/
│   └── module.go                         # Ensure NewLogoutUseCase is registered
├── adapter/
│   └── http/
│       ├── auth_handler.go               # HTTP routes and handler logic
│       ├── auth_handler_test.go          # Unit tests for the AuthHandler
│       └── module.go                     # Uber Fx module for HTTP adapter
```

**Structure Decision**: Standard Go adapter layout conforming to Clean Architecture.

## Complexity Tracking

No violations. The implementation uses standard Echo and Uber Fx patterns.
