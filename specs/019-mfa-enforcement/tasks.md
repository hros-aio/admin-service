# Tasks: MFA Enforcement (Super Admins)

**Input**: Design documents from `/specs/019-mfa-enforcement/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Implement migration (TSK-MFA-001), domain primitives (TSK-MFA-002), DTO/OpenAPI contract (TSK-MFA-003), and Redis cache (TSK-MFA-004).

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

---

## Phase 3: DTO & OpenAPI Contract (TSK-MFA-003) ✅ Complete

- [x] T008 [P] [US3] Update `api/openapi.yaml` to define `/v1/auth/mfa/verify` endpoint, document its responses (200, 401 with `MFA_INVALID` and `MFA_TOKEN_EXPIRED`), update `LoginResponse` fields, and add `MFAVerifyRequest` schema.
- [x] T009 [P] [US3] Update `LoginResponse` and implement `MFAVerifyRequest` with validation tags in `internal/adapter/http/auth/dto/auth_dto.go`.
- [x] T010 [P] [US3] Update `internal/adapter/http/auth/dto/auth_dto_test.go` to test validations and JSON mapping of `MFAVerifyRequest` and `LoginResponse`.

---

## Phase 4: Redis Cache (TSK-MFA-004) ✅ Complete

- [x] T011 [P] [US4] Update `MFACache` interface definition in `internal/application/interfaces/mfa_cache.go` and its unit tests to use `StoreToken`, `GetAdminID`, and `DeleteToken` methods mapping to the Admin ID.
- [x] T012 [P] [US4] Implement `RedisMFACache` in `internal/infrastructure/cache/mfa_redis.go` using Redis client connection and key prefix `auth:mfa_token:{mfaToken}` with 5-minute TTL.
- [x] T013 [P] [US4] Implement unit tests in `internal/infrastructure/cache/mfa_redis_test.go` using `miniredis` to verify cache operations and strict expiration.

---

## Phase 5: LoginUseCase MFA Challenge Interception (TSK-MFA-005) 🔲 Pending

Story goal: Update `LoginUseCase` to check user roles and issue an intermediate MFA challenge token for Super Admin logins instead of JWT pairs and persistent sessions.

Independent test criteria: Unit tests cover user role resolution, checking for `"Super Admin"`, secure generation of token, storage failures, and correct branching between standard and Super Admin login paths.

- [x] T014 [US5] Add `GetRoleNameByID(ctx context.Context, roleID string) (string, error)` method to `AdminUserRepository` interface in `internal/domain/admin_user.go`.
- [x] T015 [P] [US5] Implement `GetRoleNameByID` method in `GormAdminUserRepository` in `internal/infrastructure/repository/auth/repository.go`.
- [x] T016 [US5] Update `LoginUseCase` in `internal/application/usecase/login_usecase.go` to check if user's role is `"Super Admin"`. If true, generate a cryptographically secure random `mfa_token` (e.g., 32-byte hex string), store in `MFACache`, log intermediate success (redacting token), and return a `LoginOutput` containing the token with `MfaRequired: true`, bypassing JWT generation and session creation.
- [x] T017 [P] [US5] Add unit tests in `internal/application/usecase/login_usecase_test.go` to achieve 100% statement and branch coverage for the Super Admin role check and MFA token redirection.

