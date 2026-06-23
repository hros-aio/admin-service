# Tasks: MFA Enforcement (Super Admins)

**Input**: Design documents from `/specs/019-mfa-enforcement/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Implement migration (TSK-MFA-001) and domain layer primitives (TSK-MFA-002).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on each other)
- **[Story]**: Maps to user story in spec.md
- Include exact file paths in every task description

---

## Phase 1: Migration (TSK-MFA-001) ✅ Complete

- [x] T001 [US1] Create up migration script at `migrations/000003_add_mfa_to_admin_users.up.sql` to add `totp_secret` (VARCHAR) and `webauthn_credentials` (JSONB) columns to the `admin_users` table, and migrate any existing data from `mfa_secret` to `totp_secret`, before dropping `mfa_secret`.
- [x] T002 [US1] Create down migration script at `migrations/000003_add_mfa_to_admin_users.down.sql` to revert the migration by recreating the `mfa_secret` (VARCHAR) column, restoring its values from `totp_secret`, and dropping `totp_secret` and `webauthn_credentials` columns.

---

## Phase 2: Domain Primitives (TSK-MFA-002) ✅ Complete

- [x] T003 [P] [US2] Update `AdminUser` struct in `internal/domain/admin_user.go` to include `TotpSecret string` and `WebauthnCredentials []byte`.
- [x] T004 [P] [US2] Define the `MFACache` interface in `internal/application/interfaces/mfa_cache.go` to support storing, getting, and deleting partially authenticated contexts.
- [x] T005 [P] [US2] Add specific domain errors `ErrMFAInvalid` and `ErrMFATokenExpired` in `internal/domain/errors/auth_errors.go`.
- [x] T006 [P] [US2] Define the event payload structs `MFASuccessEvent` and `MFAFailedEvent` in `internal/domain/events/auth_events.go`.
- [x] T007 [P] [US2] Implement unit tests to verify the entity updates, error instances, and event serialization in:
  - `internal/domain/admin_user_test.go`
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
  - `internal/application/interfaces/mfa_cache_test.go`
