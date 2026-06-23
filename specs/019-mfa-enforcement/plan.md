# Implementation Plan: MFA Enforcement (Super Admins)

**Branch**: `019-mfa-enforcement` | **Date**: 2026-06-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/019-mfa-enforcement/spec.md`

## Summary

This plan outlines the database migration required to support MFA Enforcement (Super Admins) and WebAuthn credentials storage. We will introduce `totp_secret` (VARCHAR) and `webauthn_credentials` (JSONB) columns to the `admin_users` table. To maintain backward compatibility and preserve existing data, we will copy any existing values in `mfa_secret` to `totp_secret` during the up migration, and revert it during the down migration.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (raw SQL migrations)

**Storage**: PostgreSQL 15+ (specifically using JSONB for WebAuthn credentials storage)

**Testing**: Local PostgreSQL instance migration run and rollback checks.

**Target Platform**: Linux server / local developer machines

**Project Type**: web-service (Go backend)

**Constraints**:
- Migrations must be non-destructive to existing user data.
- The `webauthn_credentials` column must use the `JSONB` format to allow structured queries and future schema flexibility.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Schema migration files are standard PostgreSQL SQL files under the `migrations/` folder. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Database migration task; no API endpoints are added or changed in this specific migration task. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Migration SQL scripts will be run and tested against the local PostgreSQL test instance. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Target task TSK-MFA-001 maps to Phase 1 migration script creation. |
| **V. Observability & Structured Logging** | ✅ PASS | Migration logging will be handled by the database schema executor. |

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
migrations/
├── 000003_add_mfa_to_admin_users.up.sql
└── 000003_add_mfa_to_admin_users.down.sql
```

**Structure Decision**: Standard SQL migrations under `migrations/` directory.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*(No violations)*
