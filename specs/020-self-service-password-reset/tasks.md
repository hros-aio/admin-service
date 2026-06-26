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

---

## Phase 2: DTO & OpenAPI Contract (TSK-PR-002)

- [x] T005 [P] [US2] Update `internal/adapter/http/auth/dto/auth_dto.go` to add `PasswordResetRequest` and `PasswordResetConfirmRequest` structs with validation tags.
- [x] T006 [P] [US2] Update `internal/adapter/http/auth/dto/auth_dto_test.go` to test validations and JSON mapping of `PasswordResetRequest` and `PasswordResetConfirmRequest`.
- [x] T007 [P] [US2] Update `api/openapi.yaml` to document `/v1/auth/password-reset/request` and `/v1/auth/password-reset/confirm` endpoints, detailing error responses for 200, 400 (`TOKEN_EXPIRED`, `TOKEN_USED`), and 422 (`PASSWORD_WEAK`).

---

## Phase 3: Redis Cache (TSK-PR-003)

- [x] T008 [P] [US3] Implement `RedisPasswordResetCache` in `internal/infrastructure/cache/password_reset_redis.go` using Redis client connection and key prefix `auth:reset_token:{token}` with a strict 60-minute TTL.
- [x] T009 [P] [US3] Implement unit tests in `internal/infrastructure/cache/password_reset_redis_test.go` using `miniredis` to verify cache operations and strict expiration.

---

## Phase 4: Kafka Event Producer (TSK-PR-004)

- [x] T010 [P] [US4] Update `internal/adapter/kafka/producer/email_events.go` to add `PublishPasswordResetEmail` method on `EmailKafkaProducer` dispatching `events.EmailSendEvent` to topic `email.send.v1`.
- [x] T011 [P] [US4] Update `internal/adapter/kafka/producer/email_events_test.go` to test `PublishPasswordResetEmail` serialization, formatting, and dispatching.

---

## Phase 5: Repository Updates (TSK-PR-005)

- [x] T012 [P] [US5] Update `AdminUserRepository` interface in `internal/domain/admin_user.go` to add `UpdatePassword(ctx, id, newHash)`.
- [x] T013 [P] [US5] Update `SessionTokenRepository` interface in `internal/domain/session_token.go` to add `DeleteAllByAdminID(ctx, adminID)`.
- [x] T014 [P] [US5] Implement GORM repository method `UpdatePassword` in `internal/infrastructure/repository/auth/repository.go`.
- [x] T015 [P] [US5] Implement GORM repository method `DeleteAllByAdminID` in `internal/infrastructure/repository/auth/session_token_repository.go`.
- [x] T016 [P] [US5] Implement unit tests for `UpdatePassword` in `internal/infrastructure/repository/auth/repository_test.go` using `sqlmock`.
- [x] T017 [P] [US5] Implement unit tests for `DeleteAllByAdminID` in `internal/infrastructure/repository/auth/session_token_repository_test.go` using `sqlmock`.
- [x] T018 [P] [US5] Update test mocks in `internal/adapter/http/auth_handler_test.go` and `internal/application/usecase/login_usecase_test.go`.

---

## Phase 6: Usecase Implementation (TSK-PR-006)

- [x] T019 [P] [US3] Define `PasswordResetNotifier` interface in `internal/application/interfaces/password_reset_notifier.go` and compliance tests in `internal/application/interfaces/password_reset_notifier_test.go`.
- [x] T020 [P] [US3] Implement `RequestPasswordResetUseCase` in `internal/application/usecase/request_password_reset_usecase.go`.
- [x] T021 [P] [US3] Implement unit tests for `RequestPasswordResetUseCase` in `internal/application/usecase/request_password_reset_usecase_test.go`.
- [x] T022 [P] [US3] Update `AuditLogger` interface in `internal/domain/auth/audit.go` to support `LogPasswordResetRequested`.
- [x] T023 [P] [US3] Implement `LogPasswordResetRequested` in `internal/infrastructure/auth/audit_logger.go` and add unit tests in `internal/infrastructure/auth/audit_logger_test.go`.

---

## Phase 7: Confirm Password Reset Usecase (TSK-PR-007)

- [x] T024 [P] [US4] Implement `ConfirmPasswordResetUseCase` in `internal/application/usecase/confirm_password_reset_usecase.go`.
- [x] T025 [P] [US4] Implement unit tests for `ConfirmPasswordResetUseCase` in `internal/application/usecase/confirm_password_reset_usecase_test.go`.
- [x] T026 [P] [US4] Update `AuditLogger` interface in `internal/domain/auth/audit.go` and `SlogAuditLogger` in `internal/infrastructure/auth/audit_logger.go` to support `LogPasswordResetCompleted`.
- [x] T027 [P] [US4] Update unit tests in `internal/infrastructure/auth/audit_logger_test.go` for `LogPasswordResetCompleted`.

---

## Phase 8: HTTP Handlers (TSK-PR-008)

- [x] T028 [P] [US5] Implement HTTP handler `RequestPasswordReset` and `ConfirmPasswordReset` in `internal/adapter/http/auth_handler.go`.
- [x] T029 [P] [US5] Implement unit tests in `internal/adapter/http/auth_handler_test.go` to assert correct HTTP mappings and validation errors.
- [x] T030 [P] [US5] Wire new usecases and map handlers to Echo routes in `internal/adapter/http/auth_handler.go` (and wire via Fx modules).

---

## Phase 9: Integration Tests (TSK-PR-009)

- [ ] T031 [P] [US6] Create integration test `TestPasswordResetFlow` in `test/integration/password_reset_flow_test.go` utilizing actual PostgreSQL and Redis docker containers.
- [ ] T032 [P] [US6] Verify the successful password reset flow, database updates, active session wipe, and token expiration rejection behavior.
