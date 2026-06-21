# Implementation Plan: Auth Refresh DTOs

**Branch**: `011-auth-refresh-dtos` | **Date**: 2026-06-21 | **Spec**: [specs/011-auth-refresh-dtos/spec.md](spec.md)

**Input**: Feature specification from `/specs/011-auth-refresh-dtos/spec.md`

## Summary

Define request/response DTO structs for authentication token rotation, adding the `remember_me` flag to `LoginRequest` and creating `RefreshRequest` containing the `refresh_token` with strict validation tags. Generate validation unit tests verifying struct binding and validation error paths. Update the OpenAPI 3.0 contract `api/openapi.yaml` to include the `POST /v1/auth/refresh` path and update relevant DTO components.

## Technical Context

**Language/Version**: Go 1.23

**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/go-playground/validator/v10`, `github.com/stretchr/testify`

**Storage**: N/A (Pure DTO and API contract updates. No persistence layer interaction.)

**Testing**: `go test` and `github.com/stretchr/testify` for table-driven unit tests.

**Target Platform**: Linux

**Project Type**: Web Service (HTTP Adapter Layer)

**Performance Goals**: <1ms for serialization, JSON binding, and struct validation processing.

**Constraints**: Strict Clean Architecture rules. Framework bindings (Echo context) and adapter DTOs must remain strictly within the adapter/http layer and must not leak into application or domain layers.

**Scale/Scope**: Auth DTO adapter package and contract updates only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: DTO updates are strictly contained in `internal/adapter/http/auth/dto/`. No framework bindings or validation annotations leak into domain or application layers. (PASS)
2. **Documentation-First**: OpenAPI specification is updated in sync with DTO changes within the same feature slice. (PASS)
3. **Unit-Test-Per-File**: DTO files have matching `_test.go` unit tests validating binding scenarios. (PASS)
4. **Task-Driven**: The plan is focused on and limited to TSK-AUTH-011. (PASS)
5. **Observability**: Handler validations properly return standard JSON error responses with request attributes. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/011-auth-refresh-dtos/
├── plan.md              # This file
├── research.md          # Research decisions
├── data-model.md        # DTO structures and fields
├── quickstart.md        # Run and validation instructions
├── contracts/
│   └── api_contracts.md # API Contract details
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code (repository root)

```text
api/
└── openapi.yaml               # OpenAPI contract

internal/
└── adapter/
    └── http/
        └── auth/
            └── dto/
                ├── auth_dto.go       # DTO definition file
                └── auth_dto_test.go  # DTO unit tests
```

**Structure Decision**: Single project adapter layout. Requests and responses are placed under `internal/adapter/http/auth/dto/` following standard Clean Architecture adapter patterns.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
