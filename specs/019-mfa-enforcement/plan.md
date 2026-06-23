# Implementation Plan: MFA Enforcement (Super Admins)

**Branch**: `019-mfa-enforcement` | **Date**: 2026-06-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/019-mfa-enforcement/spec.md`

## Summary

This plan outlines the implementation of MFA Enforcement (Super Admins).

**Phase 1 (TSK-MFA-001 — ✅ Done)**: Database migrations to add `totp_secret` and `webauthn_credentials` to the `admin_users` table with full up/down idempotency.

**Phase 2 (TSK-MFA-002 — 🔲 Pending)**: Domain layer primitives. We will update the `AdminUser` domain entity, define the `MFACache` interface, add the specific domain errors `ErrMFAInvalid` and `ErrMFATokenExpired`, and define event payloads for `mfa.success` and `mfa.failed`.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (Go standard library for domain primitives)

**Storage**: PostgreSQL 15+ (migrations already created), Redis for the temporary MFA cache

**Testing**: Unit tests for domain entity updates, errors serialization, and event structures.

**Target Platform**: Linux server / local developer machines

**Project Type**: web-service (Go backend)

**Constraints**:
- Domain layer (`internal/domain`) must have zero external infrastructure/framework dependencies.
- Cache interfaces must reside under `internal/application/interfaces` to maintain layer boundaries.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Domain entity, errors, and events have zero dependencies. `MFACache` interface is placed under application layer boundaries (`internal/application/interfaces`). |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | API and handlers will be addressed in subsequent tasks; this task is purely domain/primitives. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Every created/updated file will have a corresponding `_test.go` file with unit tests. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Target task TSK-MFA-002 maps to Phase 2 domain primitives creation. |
| **V. Observability & Structured Logging** | ✅ PASS | Domain events will contain fields appropriate for audit logging and downstream analysis. |

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
internal/
├── domain/
│   ├── admin_user.go     # Updated to include TotpSecret and WebauthnCredentials
│   ├── errors/
│   │   └── auth_errors.go # Updated to include ErrMFAInvalid and ErrMFATokenExpired
│   └── events/
│       └── auth_events.go # Updated to include mfa.success and mfa.failed payload structs
└── application/
    └── interfaces/
        └── mfa_cache.go   # Created MFACache interface
```

**Structure Decision**: Standard Go files matching clean architecture structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*(No violations)*
