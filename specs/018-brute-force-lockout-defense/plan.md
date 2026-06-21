# Implementation Plan: Brute-Force Lockout Defense Domain Definitions

**Branch**: `018-brute-force-lockout-defense` | **Date**: 2026-06-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/018-brute-force-lockout-defense/spec.md`

## Summary

This plan outlines the implementation of the Domain layer definitions for the Brute-Force Lockout Defense feature. We will define the `BruteForceCache` interface in the application layer, define the specific domain error `ErrAccountLocked` in the domain layer, and define the event payload structs for `account.locked` (audit event) and `email.send` (notification event) in the domain layer.

No infrastructure or external dependencies will be imported by these definitions, satisfying clean architecture boundaries.

## Technical Context

**Language/Version**: Go 1.23+ (using project's Go 1.26.1)

**Primary Dependencies**: None (Standard Library only)

**Storage**: N/A (Interfaces/structs only)

**Testing**: Standard library testing with `testify/assert` and `testify/require`

**Target Platform**: Linux server / local developer machines

**Project Type**: Domain layer code definitions

**Performance Goals**: N/A

**Constraints**:
- Strict Clean Architecture: No imports from infrastructure, adapter, or framework packages in `internal/domain` or `internal/application/interfaces`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: PASS. We are defining abstractions (interfaces) and pure domain structures (errors and events) which have no outward dependencies.
- **Documentation-First**: PASS. The specification is completed in `spec.md`.
- **Unit-Test-Per-File**: PASS. We will add a corresponding unit test file for every Go production file modified or created.
- **Task-Driven & Atomic Implementation**: PASS. The tasks will be written to `tasks.md` sequentially.
- **Observability & Structured Logging**: PASS. Event payloads support serialization suitable for audit logging.

## Project Structure

### Documentation (this feature)

```text
specs/018-brute-force-lockout-defense/
├── plan.md              # This file
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Requirements verification checklist
└── tasks.md             # Tasks list (generated next)
```

### Source Code (repository root)

```text
internal/
├── application/
│   └── interfaces/
│       ├── brute_force_cache.go                  # BruteForceCache cache interface
│       └── brute_force_cache_test.go             # Unit tests for brute_force_cache interfaces (dummy or compile tests)
└── domain/
    ├── errors/
    │   ├── auth_errors.go                        # Define ErrAccountLocked
    │   └── auth_errors_test.go                   # Update unit tests
    └── events/
        ├── auth_events.go                        # Define account.locked and email.send event payloads
        └── auth_events_test.go                   # Unit tests for events serialization
```

**Structure Decision**: Place the domain and application layers in their respective clean architecture packages without any concrete adapter/infrastructure references.

## Complexity Tracking

*No violations found.*
