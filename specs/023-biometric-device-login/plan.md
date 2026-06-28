# Implementation Plan: Biometric Device Login (WebAuthn) - UseCase Layer (TSK-BIO-005)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Implement `GenerateBiometricChallengeUseCase` which accepts an email, validates the user has a registered biometric credential, generates a 32-byte secure random challenge, stores it in the `WebAuthnChallengeCache` for 5 minutes, and returns the base64url-encoded challenge along with the credential ID to the client.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `crypto/rand`, `encoding/base64`, `encoding/json`

**Storage**: Redis (WebAuthn Challenge Cache) and PostgreSQL (Admin User Repository)

**Testing**: Standard library testing with Go mocks (`github.com/stretchr/testify/mock`)

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: Sub-millisecond execution for memory-bound challenge generation

**Constraints**: Clean Architecture. Usecase logic goes into `internal/application/usecase/generate_biometric_challenge_usecase.go`, and tests in `internal/application/usecase/generate_biometric_challenge_usecase_test.go`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Clean Architecture: UseCase layer depends only on domain boundaries. Infrastructure components are mocked.
- [x] Documentation-First: Requirements and outcomes documented in `specs/023-biometric-device-login/spec.md`.
- [x] Unit-Test-Per-File: Unit tests for the usecase are located in `generate_biometric_challenge_usecase_test.go`.
- [x] Task-Driven: Focus strictly on TSK-BIO-005.
- [x] Observability: Structured logs for success/failure in challenge generation.

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
└── application/
    └── usecase/
        ├── generate_biometric_challenge_usecase.go        # UseCase implementation
        └── generate_biometric_challenge_usecase_test.go   # UseCase Unit Tests (mocked)
```

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
