# Tasks: SSO Identity Federation

**Input**: Design documents from `/specs/022-sso-identity-federation/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Define the `SSOStateCache` interface, specific domain errors, and event payload structs. Create migration scripts, define request DTOs, and update OpenAPI contracts.

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

---

## Phase 2: Database Migration (TSK-SSO-002)

- [x] T005 [P] [US2] Create SQL migration script `migrations/000005_add_sso_to_admin_users.up.sql` to add unique `sso_identity_id` and `sso_provider` to `admin_users`.
- [x] T006 [P] [US2] Create SQL migration script `migrations/000005_add_sso_to_admin_users.down.sql` to drop these columns.
- [x] T007 [P] [US2] Implement integration test in `test/integration/sso_migration_test.go` verifying that UP migration adds columns with correct constraint and DOWN migration drops them.

---

## Phase 3: DTO & API Contract (TSK-SSO-003)

- [x] T008 [P] [US3] Define `SSOCallbackRequest` DTO in `internal/adapter/http/auth/dto/auth_dto.go` with validation tags.
- [x] T009 [P] [US3] Add unit tests for `SSOCallbackRequest` validation in `internal/adapter/http/auth/dto/auth_dto_test.go`.
- [x] T010 [P] [US3] Document `GET /auth/sso/initiate` and `GET /auth/sso/callback` endpoints in `api/openapi.yaml`.
