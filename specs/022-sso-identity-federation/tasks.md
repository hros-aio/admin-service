# Tasks: SSO Identity Federation

**Input**: Design documents from `/specs/022-sso-identity-federation/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Define the `SSOStateCache` interface, specific domain errors, and event payload structs. Create migration scripts, define request DTOs, update OpenAPI contracts, implement Redis cache, implement DB lookup query, implement SSO initiation logic, and implement SSO callback logic.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can parallelize
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

---

## Phase 4: Redis Cache Layer (TSK-SSO-004)

- [x] T011 [P] [US4] Implement `RedisSSOStateCache` in `internal/infrastructure/cache/sso_state_redis.go` conforming to the `interfaces.SSOStateCache` interface.
- [x] T012 [P] [US4] Implement unit tests using `go-redis` mock (like `redismock`) in `internal/infrastructure/cache/sso_state_redis_test.go`.
- [x] T013 [P] [US4] Register `RedisSSOStateCache` in the Uber Fx dependency injection module if required.

---

## Phase 5: Repository Layer (TSK-SSO-005)

- [x] T014 [P] [US5] Add `FindByEmailOrSSO(ctx, email, ssoID)` method to `AdminUserRepository` interface in `internal/domain/admin_user.go`.
- [x] T015 [P] [US5] Implement `FindByEmailOrSSO` in `GormAdminUserRepository` inside `internal/infrastructure/repository/auth/repository.go`.
- [x] T016 [P] [US5] Implement unit tests for `FindByEmailOrSSO` in `internal/infrastructure/repository/auth/repository_test.go` using `sqlmock`.

---

## Phase 6: SSO Initiation Use Case (TSK-SSO-006)

- [x] T017 [P] [US6] Implement `InitiateSSOUseCase` in `internal/application/usecase/initiate_sso_usecase.go`.
- [x] T018 [P] [US6] Implement unit tests for `InitiateSSOUseCase` in `internal/application/usecase/initiate_sso_usecase_test.go`.

---

## Phase 7: SSO Callback Use Case (TSK-SSO-007)

- [x] T019 [P] [US7] Implement `CallbackSSOUseCase` in `internal/application/usecase/callback_sso_usecase.go`.
- [x] T020 [P] [US7] Implement unit tests for `CallbackSSOUseCase` in `internal/application/usecase/callback_sso_usecase_test.go`.

---

## Phase 8: SSO HTTP Handlers (TSK-SSO-008)

- [x] T021 [P] [US8] Implement Echo HTTP handlers for `GET /auth/sso/initiate` and `GET /auth/sso/callback` in `internal/adapter/http/auth_sso_handler.go`.
- [x] T022 [P] [US8] Implement unit tests for the SSO HTTP handlers in `internal/adapter/http/auth_sso_handler_test.go`.
- [x] T023 [US8] Register the SSO HTTP handlers and routes in `internal/adapter/http/module.go`.

