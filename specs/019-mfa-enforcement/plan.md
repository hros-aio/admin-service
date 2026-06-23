# Implementation Plan: MFA Enforcement (Super Admins)

**Branch**: `019-mfa-enforcement` | **Date**: 2026-06-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/019-mfa-enforcement/spec.md`

## Summary

This plan outlines the implementation of MFA Enforcement (Super Admins).

**Phase 1 (TSK-MFA-001 — ✅ Done)**: Database migrations to add `totp_secret` and `webauthn_credentials` to the `admin_users` table with full up/down idempotency.

**Phase 2 (TSK-MFA-002 — ✅ Done)**: Domain layer primitives — update `AdminUser` entity, `MFACache` interface, errors, and events.

**Phase 3 (TSK-MFA-003 — 🔲 Pending)**: DTO and OpenAPI Contract updates. We will update `api/openapi.yaml` to define `/v1/auth/mfa/verify` and update `LoginResponse` fields, and update `internal/adapter/http/auth/dto/auth_dto.go` to add validation tags and define `MFAVerifyRequest`.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/go-playground/validator/v10` for DTO validations.

**Storage**: None in this phase.

**Testing**: Unit tests for DTO validation tags (verifying invalid formats are blocked).

**Target Platform**: Linux server / local developer machines

**Project Type**: web-service (Go backend)

**Constraints**:
- The OpenAPI YAML must be fully valid according to the OpenAPI 3.0.3 specification.
- Response payloads must map correctly between Go models and JSON representations.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | DTO files reside purely in the HTTP adapter boundary (`internal/adapter/http/auth/dto`), completely decoupled from domain and application layers. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | The API endpoint and request/response models are added to `api/openapi.yaml` in this phase before handler logic is built. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Every created/updated DTO file has a corresponding `_test.go` file validating validation tags. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Target task TSK-MFA-003 maps to Phase 3 DTO & OpenAPI updates. |
| **V. Observability & Structured Logging** | ✅ PASS | Standard ErrorResponse schemas are used for bad requests and validation errors. |

## Project Structure

### Documentation (this feature)

```text
specs/019-mfa-enforcement/
├── plan.md              # This file
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Task definitions
```

### Source Code (repository root)

```text
api/
└── openapi.yaml           # Updated to define verification endpoints and response models
internal/
└── adapter/
    └── http/
        └── auth/
            └── dto/
                ├── auth_dto.go     # Updated LoginResponse and added MFAVerifyRequest
                └── auth_dto_test.go # Updated tests to cover mfa verification validations
```

**Structure Decision**: Standard Go files matching clean architecture structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*(No violations)*
