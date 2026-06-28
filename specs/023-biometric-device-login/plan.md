# Implementation Plan: Biometric Device Login (WebAuthn)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/023-biometric-device-login/spec.md`

## Summary

Implement biometric device login (WebAuthn) to support passwordless logins using TouchID, FaceID, or platform authenticators.
This task defines the domain interfaces, errors, and success events for Biometric Authentication.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Standard library only (no external dependencies for domain interfaces)

**Storage**: Redis for Challenge Cache, PostgreSQL (GORM) for credential mapping

**Testing**: go test with mock interface implementations

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: <50ms cache check latency

**Constraints**: Clean architecture, zero infrastructure imports in domain/interfaces

**Scale/Scope**: Transient challenges cached per user session (TTL: 60s)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Clean Architecture: No external dependencies in internal/domain or application/interfaces.
- [x] Documentation-First: Interfaces fully documented.
- [x] Unit-Test-Per-File: Every package file has a corresponding test.
- [x] Task-Driven: Focus strictly on TSK-BIO-001.
- [x] Observability: Biometric login event maps to standard slog fields.

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
internal/
├── application/
│   └── interfaces/
│       └── webauthn_cache.go
├── domain/
│   ├── errors/
│   │   └── auth_errors.go
│   └── events/
│       └── auth_events.go
```

**Structure Decision**: Standard Go Clean Architecture layout. Domain files go to internal/domain, interfaces go to internal/application/interfaces.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
