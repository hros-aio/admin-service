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

**Phase 8 (TSK-ACT-008 — ✅ Done)**: Wire missing components (`AcceptInviteUseCase` and `NotificationKafkaProducer`) into Fx modules (`internal/application/module.go`, `internal/adapter/kafka/producer/module.go`, and `internal/app/app.go`). Implement an end-to-end integration test in `test/integration/accept_invite_flow_test.go` using `testcontainers` for PostgreSQL and a mock Sarama producer to verify the full activation flow and error handling.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Testcontainers for Go (PostgreSQL), testify, miniredis, Sarama mock sync producer.

**Storage**: PostgreSQL (for invite tokens and admin users).

## Constitution Check

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Integration test proves the end-to-end flow without breaking architecture boundaries. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Updating planning and task files prior to coding. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Handled in previous phases. Integration test checks multi-component interactions. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Focusing solely on task TSK-ACT-008. |
| **V. Observability & Structured Logging** | ✅ PASS | End-to-end flow verified including events fired to Kafka and database mutations. |

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
├── app/
│   └── app.go                   # Bind NotificationKafkaProducer as NotificationPublisher
├── adapter/
│   ├── http/
│   │   ├── auth_handler.go      # AcceptInvite handler
│   │   └── auth_handler_test.go # Unit tests for AcceptInvite handler
│   ├── kafka/
│   │   └── producer/
│   │       ├── module.go        # Register NewNotificationKafkaProducer
│   │       └── notification_events.go # NotificationKafkaProducer implementation
│   └── http/
│       └── auth/
│           └── dto/
│               ├── auth_dto.go  # AcceptInviteRequest DTO
│               └── auth_dto_test.go # Test AcceptInviteRequest validation
├── application/
│   ├── module.go                # Register NewAcceptInviteUseCase
│   └── usecase/
│       ├── accept_invite_usecase.go # AcceptInviteUseCase implementation
│       └── accept_invite_usecase_test.go # Unit tests for AcceptInviteUseCase
├── domain/
│   ├── invite_token.go          # InviteToken entity & repository interface
│   ├── errors/
│   │   └── auth_errors.go       # ErrInviteExpired, ErrInviteUsed
│   └── events/
│       └── auth_events.go       # Event payload structs
├── infrastructure/
│   └── repository/
│       └── auth/
│           ├── invite_token_repository.go # GORM InviteTokenRepository
│           └── repository.go            # GormAdminUserRepository update
migrations/
├── 000004_create_invite_tokens.up.sql   # UP migration
└── 000004_create_invite_tokens.down.sql # DOWN migration
test/
└── integration/
    └── accept_invite_flow_test.go       # E2E activation flow integration test (TSK-ACT-008)
```
