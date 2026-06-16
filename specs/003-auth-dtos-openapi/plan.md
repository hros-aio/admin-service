# Implementation Plan: Authentication DTOs and OpenAPI

**Branch**: `002-authentication-service` | **Date**: 2026-06-16 | **Spec**: [specs/003-auth-dtos-openapi/spec.md](spec.md)

**Input**: Feature specification from `/specs/003-auth-dtos-openapi/spec.md`

## Summary
Define HTTP DTOs for authentication (LoginRequest, LoginResponse) and update the OpenAPI contract to include `POST /v1/auth/login` and `DELETE /v1/auth/session` endpoints. This task ensures that the API contract is established before business logic implementation, adhering to the Documentation-First principle.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Echo Framework, github.com/go-playground/validator/v10

**Storage**: N/A for this task

**Testing**: `go test ./... -race -count=1`

**Target Platform**: Linux

**Project Type**: web-service

**Performance Goals**: N/A (contract only)

**Constraints**: Clean Architecture (adapter layer isolation), OpenAPI 3.0.3

**Scale/Scope**: 2 endpoints, 2 DTOs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Gate I**: Clean Architecture adherence. DTOs are correctly placed in `internal/adapter/http/dto`.
- [x] **Gate II**: Documentation-First & OpenAPI-Driven. OpenAPI is updated before/during DTO definition.
- [x] **Gate III**: Unit-Test-Per-File. Tests will be added if DTOs have mapping logic.
- [x] **Gate IV**: Task-Driven & Atomic Implementation. Task is strictly scoped to DTOs and OpenAPI.

## Project Structure

### Documentation (this feature)

```text
specs/003-auth-dtos-openapi/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # OpenAPI (api/openapi.yaml)
└── tasks.md             # To be created by /speckit.tasks
```

### Source Code (repository root)

```text
api/
└── openapi.yaml         # API Contract

internal/
└── adapter/
    └── http/
        └── auth/
            └── dto/
                └── auth_dto.go # DTO definitions
```

**Structure Decision**: Single project layout. DTOs will be placed in `internal/adapter/http/auth/dto` to maintain domain-based organization within the adapter layer.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

N/A
