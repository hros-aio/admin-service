# Implementation Plan: MFA Enforcement (Super Admins)

**Branch**: `019-mfa-enforcement` | **Date**: 2026-06-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/019-mfa-enforcement/spec.md`

## Summary

This plan outlines the implementation of MFA Enforcement (Super Admins).

**Phase 1 (TSK-MFA-001 — ✅ Done)**: Database migrations to add `totp_secret` and `webauthn_credentials` to the `admin_users` table with full up/down idempotency.

**Phase 2 (TSK-MFA-002 — ✅ Done)**: Domain layer primitives — update `AdminUser` entity, `MFACache` interface, errors, and events.

**Phase 3 (TSK-MFA-003 — ✅ Done)**: DTO and OpenAPI Contract updates.

**Phase 4 (TSK-MFA-004 — ✅ Done)**: Redis Cache implementation. We will implement `RedisMFACache` mapping the token to the user's Admin ID with a strict 5-minute TTL.

**Phase 5 (TSK-MFA-005 — ✅ Done)**: Intercept Super Admin login in `LoginUseCase`.
- Update `AdminUserRepository` to include a method `GetRoleNameByID(ctx context.Context, roleID string) (string, error)` for role validation.
- Implement `GetRoleNameByID` in the GORM repository.
- Update `LoginUseCase.Execute` to fetch the user's role name and check if it is `"Super Admin"`.
- Generate a secure random 32-byte hex string using `crypto/rand` as the `mfa_token`.
- Cache the `mfa_token` mapping to `user.ID` using `MFACache.StoreToken()`.
- Return a `LoginOutput` indicating MFA is required, containing the `mfa_token` and permitted methods, without generating access/refresh tokens or session records.
- Add comprehensive unit tests in `login_usecase_test.go` to cover all branches.

**Phase 6 (TSK-MFA-006 — ✅ Done)**: Implement `VerifyMFAUseCase`.
- Create `VerifyMFAUseCase` accepting `VerifyMFAInput` (mfa_token, method, code) and returning `VerifyMFAOutput`.
- Fetch the admin user ID from `MFACache` using `mfa_token`. If token is missing/expired, return `ErrMFATokenExpired`.
- Fetch `AdminUser` from repository. If user not found, return appropriate error.
- Verify the code using standard TOTP library `github.com/pquerna/otp`.
- If verification fails, publish `mfa.failed` audit event and return `ErrMFAInvalid`.
- If verification succeeds, publish `mfa.success` audit event, issue the standard JWT access/refresh token pair, store a persistent session record in `SessionTokenRepository`, and delete the `mfa_token` from `MFACache`.
- Add unit tests in `verify_mfa_usecase_test.go` achieving 100% statement and branch coverage.

**Phase 7 (TSK-MFA-007 — ✅ Done)**: HTTP Handler for MFA Verification.
- Inject `VerifyMFAUseCase` into `AuthHandler`.
- Register path `POST /v1/auth/mfa/verify` mapped to `AuthHandler.VerifyMFA` in route registration.
- Update `AuthHandler.Login` to properly map `MFARequired`, `MFAToken`, and `MFAMethods` fields from output to client response envelope (already partially structured, but ensure it is strictly tested).
- Implement `VerifyMFA` method on `AuthHandler` to parse, validate, and execute verification request, returning standard JWT access/refresh token pair or mapping domain errors `ErrMFAInvalid` / `ErrMFATokenExpired` to `401 Unauthorized` responses matching contract codes.
- Implement comprehensive handler-level unit and integration tests inside `auth_handler_test.go` checking all scenarios (success, validation error, invalid token/code, usecase/internal errors).


## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/redis/go-redis/v9` for Redis caching, `github.com/pquerna/otp` for TOTP validation. `github.com/alicebob/miniredis/v2` for unit testing.

**Storage**: Redis for the temporary MFA cache (`auth:mfa_token:{mfaToken}`).

**Testing**: Unit tests using miniredis checking store, retrieve, delete, and expiration/TTL.

**Target Platform**: Linux server / local developer machines

**Project Type**: web-service (Go backend)

**Constraints**:
- Redis keys must follow the prefix format `auth:mfa_token:{mfaToken}`.
- Cache TTL must be set to exactly 5 minutes on storage.
- ErrMFATokenExpired must be returned if the token is not found.
- All logs from the Redis cache implementation must redact the token portion of the key (e.g., logging `auth:mfa_token:[REDACTED]`) to prevent leaking raw MFA tokens.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Redis Cache implementation lives inside the infrastructure boundary (`internal/infrastructure/cache/mfa_redis.go`) and implements the application interface contract `interfaces.MFACache`. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | Relies on existing API specifications for MFA. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | The Redis implementation will have its corresponding `_test.go` file with full test cases. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Target task TSK-MFA-004 maps to Phase 4 Redis cache implementation. |
| **V. Observability & Structured Logging** | ✅ PASS | Redis caching failures and validations will use structured logs with safe logging properties. |

## Project Structure

### Documentation (this feature)

```text
specs/019-mfa-enforcement/
├── plan.md              # This file
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Task definitions
```

### Source Code (repository root)

```text
internal/
├── adapter/
│   └── http/
│       ├── auth_handler.go           # Wired with VerifyMFAUseCase and custom error mappings
│       └── auth_handler_test.go      # Integration and unit tests for auth handlers
├── application/
│   └── usecase/
│       ├── verify_mfa_usecase.go     # VerifyMFAUseCase implementation
│       └── verify_mfa_usecase_test.go # Unit tests for verify_mfa_usecase.go
└── infrastructure/
    └── cache/
        ├── mfa_redis.go     # Redis implementation of MFACache
        └── mfa_redis_test.go # Unit tests using miniredis
```

**Structure Decision**: Standard Go files matching clean architecture structure.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*(No violations)*
