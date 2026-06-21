# Feature Specification: Session Persistence Flow Test

**Feature Branch**: `017-session-persistence-flow`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-017: Layer: Tests Description: Implement an integration test for the complete Session Refresh & Persistence flow. Use testcontainers for PostgreSQL and Redis to execute a full flow: Login with remember_me=true, verify the 30-day DB expiry, perform a session refresh, and verify the old access token is correctly placed on the Redis blacklist post-rotation/logout. Input: Fully assembled Auth module, testcontainers PostgreSQL and Redis instances. Output: test/integration/session_persistence_flow_test.go Definition of Done: Integration tests pass, successfully interacting with real database schemas and Redis caches, proving the refresh and revocation cycle works end-to-end. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verify Long-term Session Persistence (Priority: P1)

As an administrator using the client application, I want my session to remain active and valid for a long duration (30 days) when I choose "Remember Me", so that I do not have to log in repeatedly.

**Why this priority**: High. This verifies the end-to-end integration of the "remember me" feature, ensuring database sessions persist correctly.

**Independent Test**: Verified by invoking a login request with `remember_me=true`, inspecting the session expiration date in the test database, and confirming that the expiration timestamp matches the 30-day configuration.

**Acceptance Scenarios**:

1. **Given** a login request with `remember_me = true`, **When** the integration flow executes, **Then** the returned session token is persisted in the PostgreSQL database with an expiration date set to 30 days from the current time.

---

### User Story 2 - Verify Secure Session Rotation (Priority: P1)

As a security-conscious administrator, I want my active session to rotate its refresh token when refreshed, and I want the old access token to be blacklisted in Redis to prevent reuse, ensuring that my session is kept secure.

**Why this priority**: High. Prevents session replay attacks and guarantees token rotation is fully secure.

**Independent Test**: Verified by calling the refresh session endpoint with a valid refresh token, asserting a new token pair is returned, verifying the new refresh session is updated in PostgreSQL, and asserting that the old access token is added to the Redis blacklist cache.

**Acceptance Scenarios**:

1. **Given** an active session, **When** a refresh request is submitted with the valid refresh token, **Then** a new access token and refresh token are returned, the session record in PostgreSQL is updated with the new token and expiration, and the old access token is written to the Redis blacklist cache.
2. **Given** a recently rotated old access token, **When** a request is made using that token to access a protected resource, **Then** the request is rejected as unauthorized.

---

### Edge Cases

- **Expired Token Reuse**: If the integration test attempts to refresh using a token whose expiration has passed, the refresh endpoint must reject it.
- **Double Refresh Attempt**: If the client attempts to refresh using a refresh token that has already been rotated (old refresh token), the system must reject the request.
- **Revoked Session Post-Logout**: After calling logout, the session must be completely removed/invalidated in the PostgreSQL database, and the access token must be blacklisted in Redis.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-017-001**: The integration test suite MUST bootstrap real PostgreSQL and Redis instances using `testcontainers`.
- **FR-AUTH-017-002**: The integration test suite MUST execute database migrations against the test PostgreSQL database container before running the test cases.
- **FR-AUTH-017-003**: The integration test flow MUST execute `POST /v1/auth/login` with `remember_me=true` and verify the created session in PostgreSQL has an `ExpiresAt` value set to approximately 30 days from the current time.
- **FR-AUTH-017-004**: The integration test flow MUST execute `POST /v1/auth/refresh` using the active refresh token, verifying that a new access/refresh token pair is generated.
- **FR-AUTH-017-005**: The integration test flow MUST verify that after rotation, the old access token is added to the Redis blacklist.
- **FR-AUTH-017-006**: The integration test flow MUST execute `DELETE /v1/auth/session` (Logout) using the authorization header containing the access token, verifying that the session is deleted in PostgreSQL and the token is blacklisted in Redis.

### Key Entities

- **Session Flow Integration Test**: The test suite that orchestrates test container lifecycles, database migrations, endpoint calls, and direct state assertions.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-017-001**: The test suite executes and passes successfully within Go integration test suites.
- **SC-AUTH-017-002**: 100% of the steps in the login, refresh, blacklist, and logout flows are asserted correctly against real containerized services.
- **SC-AUTH-017-003**: Database and cache states are verified directly in PostgreSQL and Redis after each endpoint action to ensure state consistency.

## Assumptions

- Testcontainers requires a running Docker daemon on the host.
- The PostgreSQL schema migrations are available inside the repository at `migrations/` and can be loaded dynamically.
- System time in containers and the test suite are synchronized or handled within acceptable delta limits.
