# Tasks: Self-Service Password Reset

**Input**: Design documents from `/specs/020-self-service-password-reset/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Define the `PasswordResetCache` interface, specific domain errors, and event payload structs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel
- **[Story]**: Maps to user story in spec.md
- Include exact file paths in every task description

---

## Phase 1: Domain Primitives (TSK-PR-001)

- [x] T001 [P] [US1] Define `PasswordResetCache` interface in `internal/application/interfaces/password_reset_cache.go`.
- [x] T002 [P] [US1] Define `ErrTokenExpired`, `ErrTokenUsed`, and `ErrPasswordWeak` in `internal/domain/errors/auth_errors.go`.
- [x] T003 [P] [US1] Define `PasswordResetRequestedEvent` and `PasswordResetCompletedEvent` in `internal/domain/events/auth_events.go`.
- [x] T004 [P] [US1] Implement unit tests to verify the interfaces, errors, and events in:
  - `internal/application/interfaces/password_reset_cache_test.go`
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
