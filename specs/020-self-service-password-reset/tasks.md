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

