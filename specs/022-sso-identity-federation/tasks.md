# Tasks: SSO Identity Federation

**Input**: Design documents from `/specs/022-sso-identity-federation/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Define the `SSOStateCache` interface, specific domain errors, and event payload structs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel
- **[Story]**: Maps to user story in spec.md
- Include exact file paths in every task description

---

## Phase 1: Domain Primitives (TSK-SSO-001)

- [x] T001 [P] [US1] Define `SSOStateCache` interface in `internal/application/interfaces/sso_state_cache.go`.
- [x] T002 [P] [US1] Define domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState` in `internal/domain/errors/auth_errors.go`.
- [x] T003 [P] [US1] Define event payload structs for `login.sso_success` (`SSOSuccessEvent`) and `login.sso_failed` (`SSOFailedEvent`) audit events in `internal/domain/events/auth_events.go`.
- [x] T004 [P] [US1] Implement unit tests to verify the interfaces, errors, and events in:
  - `internal/application/interfaces/sso_state_cache_test.go`
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
