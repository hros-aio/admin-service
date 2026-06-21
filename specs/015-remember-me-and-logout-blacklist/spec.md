# Feature Specification: Remember Me and Logout Blacklist

**Feature Branch**: `015-remember-me-and-logout-blacklist`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-015: Layer: UseCase Description: Update the existing LoginUseCase to accept and process the remember_me parameter. If true, set the refresh token ExpiresAt to 30 days; if false, set it to a browser-session standard (or short-lived). Update LogoutUseCase to extract the current JWT's JTI and add it to the TokenBlacklist cache. Input: LoginUseCase, LogoutUseCase, TokenBlacklist interface. Output: internal/application/usecase/login_usecase.go, internal/application/usecase/logout_usecase.go, their respective test files. Definition of Done: Login correctly calculates session expiration, and Logout correctly invalidates the specific access token in Redis. Unit tests successfully assert this updated behavior. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remember Me Session Expiration (Priority: P1) 🎯 MVP

As an administrator logging in, I want to choose whether my login session is persistent ("remember me") or short-lived, so that my account access matches my security requirements (persistent on private devices, short-lived on public ones).

**Why this priority**: High. Essential for satisfying the user requirements of session persistence options and avoiding premature session termination or security leakage.

**Independent Test**: Can be verified via unit tests mocking the dependencies of `LoginUseCase` and checking the computed `ExpiresAt` and `IsPersistent` fields on the generated `SessionToken` entity for both true and false inputs.

**Acceptance Scenarios**:

1. **Given** a login request with `RememberMe` set to `true`, **When** the login executes successfully, **Then** the persisted session has `ExpiresAt` set to 30 days from the current time and `IsPersistent` set to `true`.
2. **Given** a login request with `RememberMe` set to `false`, **When** the login executes successfully, **Then** the persisted session has `ExpiresAt` set to a short-lived duration (24 hours) and `IsPersistent` set to `false`.

---

### User Story 2 - Immediate Access Token Blacklisting on Logout (Priority: P1) 🎯 MVP

As a security-conscious system administrator logging out, I want my current JWT access token to be immediately blacklisted, so that it cannot be reused by any client for subsequent API requests, even before its natural expiration.

**Why this priority**: High. Essential for preventing session hijacking and unauthorized API access using captured access tokens after a user has logged out.

**Independent Test**: Can be verified via unit tests mocking the `SessionTokenRepository` and `TokenBlacklist` interfaces to assert that a logout request with an access token extracts the JTI and adds it to the blacklist with the correct TTL, while also deleting the refresh token.

**Acceptance Scenarios**:

1. **Given** a logout request with a valid access token and refresh token, **When** the logout executes successfully, **Then** the refresh token is deleted from persistence, and the access token's JTI is added to the token blacklist cache with a TTL equal to the token's remaining time-to-live.
2. **Given** a logout request where the access token is missing or has no JTI, **When** the logout executes, **Then** the use case deletes the refresh token and proceeds without adding any JTI to the blacklist cache.
3. **Given** a logout request where the access token is already expired, **When** the logout executes, **Then** the use case deletes the refresh token and does not add the expired JTI to the blacklist.

---

### Edge Cases

- **Access Token Parse Error**: If the provided access token is malformed, the use case should log the issue, but it must still proceed with the deletion of the refresh token so the user is logged out from their session database-side.
- **Cache Connection Failure**: If adding the JTI to the `TokenBlacklist` cache fails due to network/Redis errors, the use case must log the error and proceed to ensure the refresh token deletion still completes (minimizing disruption).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-015-001**: The `LoginUseCase` MUST accept a `RememberMe` boolean field within `LoginInput`.
- **FR-AUTH-015-002**: If `RememberMe` is `true`, the `LoginUseCase` MUST calculate the session's expiration (`ExpiresAt`) as exactly 30 days from now, and set `IsPersistent` to `true`.
- **FR-AUTH-015-003**: If `RememberMe` is `false` (or omitted), the `LoginUseCase` MUST calculate the session's expiration (`ExpiresAt`) as exactly 24 hours (1 day) from now, and set `IsPersistent` to `false`.
- **FR-AUTH-015-004**: The `LogoutUseCase` MUST accept both `RefreshToken` and `AccessToken` strings in its `LogoutInput`.
- **FR-AUTH-015-005**: The `LogoutUseCase` MUST parse the provided `AccessToken` (without signature validation if using unverified parsing, as signature check is handled by middleware) to extract the `jti` (JWT ID) and `exp` (expiration timestamp) claims.
- **FR-AUTH-015-006**: If a `jti` is successfully extracted and the token's expiration `exp` is in the future, the `LogoutUseCase` MUST calculate the remaining time-to-live (`ttl = exp - now`).
- **FR-AUTH-015-007**: The `LogoutUseCase` MUST add the extracted `jti` to the `TokenBlacklist` cache with the computed `ttl` using the `TokenBlacklist.Add` interface.
- **FR-AUTH-015-008**: The `LogoutUseCase` MUST delete the session token using `SessionTokenRepository.DeleteByToken` regardless of whether blacklisting succeeds or is skipped.

### Key Entities

- **SessionToken**: The domain model representing the persistent or short-lived refresh token session.
- **TokenBlacklist**: The application interface used to query and store revoked JWT access tokens.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-015-001**: 100% of unit test coverage for `login_usecase.go` and `logout_usecase.go` verified by `go test`.
- **SC-AUTH-015-002**: Login sessions are generated with a 30-day expiry when `RememberMe` is `true` and a 24-hour expiry when `RememberMe` is `false`.
- **SC-AUTH-015-003**: Access tokens successfully generate with unique `jti` (JWT ID) values, which are blacklisted for their exact remaining lifetime upon logout.

## Assumptions

- **Access Token Claims**: Access tokens will include the standard JWT claims `jti` (UUID) and `exp` (unix timestamp) to facilitate blacklist expiration mapping.
- **Token Signature Verification**: Token signature verification is out of scope for the logout use case itself, as requests reaching the secure logout handler are already authenticated and verified by network/middleware boundaries. The use case will parse the token claims unverified to avoid needing the RSA public/private key.
