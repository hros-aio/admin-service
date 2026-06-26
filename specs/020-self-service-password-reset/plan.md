# Implementation Plan: Self-Service Password Reset

**Branch**: `020-self-service-password-reset` | **Date**: 2026-06-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/020-self-service-password-reset/spec.md`

## Summary

This plan outlines the implementation of Self-Service Password Reset.

**Phase 1 (TSK-PR-001 — ✅ Done)**: Define the `PasswordResetCache` interface, domain error variables (`ErrTokenExpired`, `ErrTokenUsed`, `ErrPasswordWeak`), and domain event payloads (`PasswordResetRequestedEvent`, `PasswordResetCompletedEvent`, `EmailSendEvent`).

**Phase 2 (TSK-PR-002 — ✅ Done)**: Define `PasswordResetRequest` and `PasswordResetConfirmRequest` DTOs with validation tags and update the OpenAPI contract.

**Phase 3 (TSK-PR-003 — ✅ Done)**: Implement the Redis cache for password reset tokens (`RedisPasswordResetCache`) with a strict 60-minute TTL.

**Phase 4 (TSK-PR-004 — ✅ Done)**: Implement the Kafka producer event payload mapping for the `email.send` event containing the secure single-use reset link.

**Phase 5 (TSK-PR-005 — ✅ Done)**: Update repositories (`AdminUserRepository`, `SessionTokenRepository`) to support password updates and session revocation.

**Phase 6 (TSK-PR-006 — ✅ Done)**: Implement the `RequestPasswordResetUseCase` application service.

**Phase 7 (TSK-PR-007 — ✅ Done)**: Implement the `ConfirmPasswordResetUseCase` application service.

**Phase 8 (TSK-PR-008 — ✅ Done)**: Implement the password reset HTTP handlers and wire them in Echo/Fx.

**Phase 9 (TSK-PR-009 — ✅ Done)**: Implement the full flow integration tests.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (Standard Go modules).

**Storage**: Redis (for token caching), PostgreSQL (for user lookup).

**Testing**: Unit tests verifying use case behavior with mock implementations of cache, repository, audit log, and notifier.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Usecase interacts with repository, cache, audit logger, and notifier via clean domain and application interfaces. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Updating specifications and plan prior to implementing the code. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | All updates and new files are covered by matching `_test.go` unit tests. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Implementing TSK-PR-006 incrementally. |
| **V. Observability & Structured Logging** | ✅ PASS | Log formats and events are emitted using slog structured logging conventions. |

## Project Structure

### Documentation (this feature)

```text
specs/020-self-service-password-reset/
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Task definitions
```

### Source Code (repository root)

```text
internal/
├── adapter/
│   └── http/
│       ├── auth_handler.go              # Map POST /v1/auth/password-reset/request and confirm
│       └── auth_handler_test.go         # Test HTTP handlers, binding, and error mapping
├── application/
│   ├── interfaces/
│   │   ├── password_reset_cache.go      # Interface for password reset cache
│   │   └── password_reset_notifier.go   # Interface for password reset Kafka publisher
│   └── usecase/
│       ├── request_password_reset_usecase.go  # Request password reset usecase
│       ├── request_password_reset_usecase_test.go # Tests for request password reset usecase
│       ├── confirm_password_reset_usecase.go  # Confirm password reset usecase
│       └── confirm_password_reset_usecase_test.go # Tests for confirm password reset usecase
├── domain/
│   ├── errors/
│   │   └── auth_errors.go               # Add ErrTokenExpired, ErrTokenUsed, ErrPasswordWeak
│   └── events/
│       └── auth_events.go               # Add PasswordResetRequestedEvent, PasswordResetCompletedEvent
test/
└── integration/
    └── password_reset_flow_test.go      # Full flow integration test using testcontainers
```
