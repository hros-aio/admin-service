# Implementation Plan: Admin Account Activation (Accept Invite)

**Branch**: `021-admin-account-activation` | **Date**: 2026-06-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/021-admin-account-activation/spec.md`

## Summary

This plan outlines the implementation of Admin Account Activation (Accept Invite).

**Phase 1 (TSK-ACT-001 — ✅ Done)**: Define the `InviteToken` domain entity and `InviteTokenRepository` interface in `internal/domain/invite_token.go`. Define domain errors `ErrInviteExpired`, `ErrInviteUsed` in `internal/domain/errors/auth_errors.go`. Define event payload structs for the `admin.activated`, `invite.accepted` audit events, and `notification.send` Kafka event in `internal/domain/events/auth_events.go`.

**Phase 2 (TSK-ACT-002 — ✅ Done)**: Create SQL migration scripts `migrations/000004_create_invite_tokens.up.sql` and `migrations/000004_create_invite_tokens.down.sql` to establish the `invite_tokens` table.

**Phase 3 (TSK-ACT-003 — ✅ Done)**: Define the `AcceptInviteRequest` HTTP DTO in `internal/adapter/http/auth/dto/auth_dto.go`. Document the `POST /v1/auth/accept-invite` endpoint in `api/openapi.yaml`.

**Phase 4 (TSK-ACT-004 — ✅ Done)**: Implement `InviteTokenRepository` and update `AdminUserRepository` with `ActivateAccount` in `internal/infrastructure/repository/auth/`.

**Phase 5 (TSK-ACT-005 — ✅ Done)**: Implement Kafka producer for `notification.send` event in `internal/adapter/kafka/producer/notification_events.go` to dispatch in-app notifications to the original inviter.

**Phase 6 (TSK-ACT-006 — ✅ Done)**: Implement `AcceptInviteUseCase` in `internal/application/usecase/accept_invite_usecase.go` to orchestrate token validation, password hashing, account activation, token consumption, and event emission.

**Phase 7 (TSK-ACT-007 — ✅ Done)**: Wire `AcceptInviteUseCase` into `AuthHandler` in `internal/adapter/http/auth_handler.go`. Replace the stub `AcceptInvite` handler body with: bind → validate → invoke use case → map `ErrInviteExpired`/`ErrInviteUsed` → 400, `ErrPasswordWeak` → 422, catch-all → 500; success → 200 OK. Add unit tests in `internal/adapter/http/auth_handler_test.go` covering all HTTP mapping branches using `httptest`.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: None (Standard Go modules).

**Storage**: PostgreSQL (for invite tokens and admin users).

## Constitution Check

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Repository implementation implements domain interfaces using GORM. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Updating planning and task files prior to coding. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Adding test cases to repository unit tests. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Focusing solely on task TSK-ACT-004. |
| **V. Observability & Structured Logging** | ✅ PASS | Errors from GORM execution mapped to appropriate domain errors. |

## Project Structure

### Documentation

```text
specs/021-admin-account-activation/
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Task definitions
```

### Source Code

```text
api/
└── openapi.yaml                 # OpenAPI contract documentation
internal/
├── adapter/
│   └── http/
│       ├── auth_handler.go      # AcceptInvite handler — wire AcceptInviteUseCase, map errors
│       ├── auth_handler_test.go # Unit tests for AcceptInvite handler (httptest)
│       └── auth/
│           └── dto/
│               ├── auth_dto.go  # Add AcceptInviteRequest DTO
│               └── auth_dto_test.go # Test AcceptInviteRequest validation
├── domain/
│   ├── invite_token.go          # InviteToken entity & repository interface
│   ├── invite_token_test.go     # Tests for InviteToken entity methods
│   ├── errors/
│   │   ├── auth_errors.go       # Add ErrInviteExpired, ErrInviteUsed
│   │   └── auth_errors_test.go  # Test coverage for errors
│   └── events/
│       ├── auth_events.go       # Add AdminActivatedEvent, InviteAcceptedEvent, NotificationSendEvent
│       └── auth_events_test.go  # Test coverage for events
├── infrastructure/
│   └── repository/
│       └── auth/
│           ├── invite_token_repository.go # GORM InviteTokenRepository implementation
│           ├── invite_token_repository_test.go # Unit tests for GORM InviteTokenRepository
│           ├── repository.go            # Update GormAdminUserRepository with ActivateAccount
│           └── repository_test.go       # Update unit tests for GormAdminUserRepository
migrations/
├── 000004_create_invite_tokens.up.sql   # SQL script to create invite_tokens table
└── 000004_create_invite_tokens.down.sql # SQL script to drop invite_tokens table
```
