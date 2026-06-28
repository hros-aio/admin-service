# Implementation Plan: Biometric Device Login (WebAuthn) - Cache Layer

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Implement the `WebAuthnChallengeCache` interface using Redis.
Define `StoreChallenge(ctx, email, challenge, ttl)` mapping the email to the secure challenge string with a strict short TTL (e.g., 5 minutes).
Define `VerifyAndConsumeChallenge(ctx, email)` to fetch and delete the challenge atomically.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/redis/go-redis/v9`

**Storage**: Redis for Challenge Cache

**Testing**: `github.com/alicebob/miniredis/v2` for local mocked Redis client unit tests

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: Sub-millisecond Redis response time for cache hits

**Constraints**: Clean architecture, interface in `application/interfaces`, implementation in `infrastructure/cache`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Clean Architecture: Cache interface in `internal/application/interfaces/webauthn_cache.go`, implementation in `internal/infrastructure/cache/webauthn_redis.go`.
- [x] Documentation-First: TTL and lifecycle documented in `specs/023-biometric-device-login/spec.md`.
- [x] Unit-Test-Per-File: `webauthn_redis_test.go` corresponding to `webauthn_redis.go`.
- [x] Task-Driven: Focus strictly on TSK-BIO-003.
- [x] Observability: Structured logging for store/fetch/consume events.

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
│       ├── webauthn_cache.go          # Cache Interface
│       └── webauthn_cache_test.go     # Compilation fake/mock check
└── infrastructure/
    └── cache/
        ├── webauthn_redis.go          # Redis Implementation
        └── webauthn_redis_test.go     # Cache Unit Tests
```

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
