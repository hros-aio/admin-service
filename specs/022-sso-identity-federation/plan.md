# Implementation Plan: SSO Identity Federation

**Branch**: `022-sso-identity-federation` | **Date**: 2026-06-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/022-sso-identity-federation/spec.md`

## Summary

This plan outlines the implementation of the Domain and Application Interface definitions, as well as the database schema updates for the SSO Identity Federation.

**Phase 1 (TSK-SSO-001)**: Define `SSOStateCache` interface in `internal/application/interfaces/sso_state_cache.go`. Define domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState` in `internal/domain/errors/auth_errors.go`. Define event payload structs for the `login.sso_success` and `login.sso_failed` audit events in `internal/domain/events/auth_events.go`.

**Phase 2 (TSK-SSO-002)**: Create SQL migration scripts `migrations/000005_add_sso_to_admin_users.up.sql` and `migrations/000005_add_sso_to_admin_users.down.sql` to add SSO mapping fields (`sso_identity_id`, `sso_provider`) to `admin_users` table.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (pure Go standard library for domain layer).

**Storage**: PostgreSQL (migration adds columns to `admin_users` table).

## Constitution Check

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Database schema changes support domain requirements without introducing domain leaks. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Written plan and task definitions updated prior to implementation. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Database migration is covered by an integration migration test. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Focusing only on task TSK-SSO-002. |
| **V. Observability & Structured Logging** | ✅ PASS | Columns allow tracking the identity provider of authenticated users. |

## Project Structure

### Documentation

```text
specs/022-sso-identity-federation/
├── checklists/
│   └── requirements.md  # Quality checklist
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Task definitions
```

### Source Code

```text
internal/
├── application/
│   └── interfaces/
│       ├── sso_state_cache.go      # SSOStateCache interface
│       └── sso_state_cache_test.go # Unit tests/verifications for SSOStateCache interface
├── domain/
│   ├── errors/
│   │   ├── auth_errors.go          # ErrNoAccountLinked and ErrInvalidSSOState
│   │   └── auth_errors_test.go     # Unit tests for domain errors
│   └── events/
│       ├── auth_events.go          # Event payload structs for SSO success/failure
│       └── auth_events_test.go     # Unit tests for event payloads
migrations/
├── 000005_add_sso_to_admin_users.up.sql   # SQL migration up script
└── 000005_add_sso_to_admin_users.down.sql # SQL migration down script
test/
└── integration/
    └── sso_migration_test.go       # Integration test for the database migration
```

**Structure Decision**: Clean Architecture database migration scripts.
