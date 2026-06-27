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

---

## Phase 2: Migration Layer (TSK-ACT-002)

- [x] T005 [P] [US1] Create migration up script `migrations/000004_create_invite_tokens.up.sql` to establish `invite_tokens` table.
- [x] T006 [P] [US1] Create migration down script `migrations/000004_create_invite_tokens.down.sql` to drop `invite_tokens` table.
- [x] T007 [P] [US1] Run database migrations forward and backward to verify they execute successfully.

---

## Phase 3: DTO & OpenAPI Contract (TSK-ACT-003)

- [x] T008 [P] [US1] Define `AcceptInviteRequest` DTO in `internal/adapter/http/auth/dto/auth_dto.go` with strict validation tags.
- [x] T009 [P] [US1] Add unit tests for `AcceptInviteRequest` validation in `internal/adapter/http/auth/dto/auth_dto_test.go`.
- [x] T010 [P] [US1] Document the `POST /v1/auth/accept-invite` endpoint in `api/openapi.yaml`.


