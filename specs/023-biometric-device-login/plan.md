# Implementation Plan: Biometric Device Login (WebAuthn)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/023-biometric-device-login/spec.md`

## Summary

Implement biometric device login (WebAuthn) to support passwordless logins using TouchID, FaceID, or platform authenticators.
This task defines the request and response Data Transfer Objects (DTOs) for the biometric login endpoints and registers them in the OpenAPI contract.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Standard library, `github.com/go-playground/validator/v10`

**Storage**: Redis for Challenge Cache, PostgreSQL (GORM) for credential mapping

**Testing**: go test validator struct checks and JSON marshalling tests

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: N/A (compile-time definitions and validation)

**Constraints**: Clean architecture, zero framework leakage in core layers, HTTP DTOs kept within the adapter layer

**Scale/Scope**: Input validation structures for `/v1/auth/biometric/challenge` and `/v1/auth/biometric/verify`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Clean Architecture: DTOs are defined within `internal/adapter/http/auth/dto/`, isolated from application/domain logic.
- [x] Documentation-First: API paths and schemas documented in `api/openapi.yaml`.
- [x] Unit-Test-Per-File: validation rules and JSON mapping tested in `internal/adapter/http/auth/dto/auth_dto_test.go`.
- [x] Task-Driven: Focus strictly on TSK-BIO-002.
- [x] Observability: N/A.

## Project Structure

### Documentation (this feature)

```text
specs/023-biometric-device-login/
├── plan.md              # This file
├── research.md          # Research (Phase 0)
├── data-model.md        # Data model (Phase 1)
├── quickstart.md        # Quickstart validation guide (Phase 1)
└── tasks.md             # Tasks (Phase 2)
```

### Source Code (repository root)

```text
api/
└── openapi.yaml
internal/
└── adapter/
    └── http/
        └── auth/
            └── dto/
                ├── auth_dto.go
                └── auth_dto_test.go
```

**Structure Decision**: HTTP adapter layout. DTOs are mapped inside the authentication module's `dto` package.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
