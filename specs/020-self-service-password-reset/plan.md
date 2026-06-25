# Implementation Plan: Self-Service Password Reset

**Branch**: `020-self-service-password-reset` | **Date**: 2026-06-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/020-self-service-password-reset/spec.md`

## Summary

This plan outlines the implementation of Self-Service Password Reset.

**Phase 1 (TSK-PR-001 — ✅ Done)**: Define the `PasswordResetCache` interface, domain error variables (`ErrTokenExpired`, `ErrTokenUsed`, `ErrPasswordWeak`), and domain event payloads (`PasswordResetRequestedEvent`, `PasswordResetCompletedEvent`, `EmailSendEvent`).

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (Standard Go modules).

**Storage**: Memory cache interfaces in domain.

**Testing**: Unit tests verifying serialization and compilation of structures.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Only defining interfaces, errors, and event payload structures in domain and application interface boundaries. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Creating specification files before writing code. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | All domain code updates will be covered by matching `_test.go` unit tests. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | TSK-PR-001 maps to Domain Primitive implementation. |
| **V. Observability & Structured Logging** | ✅ PASS | Log formats and events are defined in domain event models. |

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
├── application/
│   └── interfaces/
│       └── password_reset_cache.go      # Interface for password reset cache
├── domain/
│   ├── errors/
│   │   └── auth_errors.go               # Add ErrTokenExpired, ErrTokenUsed, ErrPasswordWeak
│   └── events/
│       └── auth_events.go               # Add PasswordResetRequestedEvent, PasswordResetCompletedEvent
```
