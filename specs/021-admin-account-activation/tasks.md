# Tasks: Admin Account Activation (Accept Invite)

**Input**: Design documents from `/specs/021-admin-account-activation/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Define the `InviteToken` domain entity, specific domain errors, and event payload structs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel
- **[Story]**: Maps to user story in spec.md
- Include exact file paths in every task description

---

## Phase 1: Domain Primitives (TSK-ACT-001)

- [x] T001 [P] [US1] Define `InviteToken` domain entity and `InviteTokenRepository` interface in `internal/domain/invite_token.go`.
- [x] T002 [P] [US1] Define `ErrInviteExpired`, `ErrInviteUsed` (and ensure `ErrPasswordWeak` is defined) in `internal/domain/errors/auth_errors.go`.
- [x] T003 [P] [US1] Define payload structs for `admin.activated` (`AdminActivatedEvent`), `invite.accepted` (`InviteAcceptedEvent`) audit events, and `notification.send` (`NotificationSendEvent`) Kafka event in `internal/domain/events/auth_events.go`.
- [x] T004 [P] [US1] Implement unit tests to verify the interfaces, errors, and events in:
  - `internal/domain/invite_token_test.go`
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
