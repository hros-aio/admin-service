# Feature Specification: Admin Logout Use Case

**Feature Branch**: `007-admin-logout-usecase`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "Use the existing repository documents as source of truth. TSK-AUTH-007: Layer: UseCase Description: Implement LogoutUseCase. Workflow: Accept current session token → delete token from session_tokens via repository → emit logout.success to audit log interface. Input: Session token string (extracted from Context). Output: internal/application/usecase/logout_usecase.go, internal/application/usecase/logout_usecase_test.go Definition of Done: Tokens are successfully removed from persistence and audit events are triggered. Unit tests pass with 100% coverage. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Secure Session Termination (Priority: P1)

As a logged-in Administrator, I want to securely log out of the HROS Admin Portal so that my active session is terminated, the refresh token is revoked/deleted from the system, and no unauthorized access can occur using that token.

**Why this priority**: High. Essential security practice to revoke credentials and prevent session hijacking on shared devices or after usage is complete.

**Independent Test**: Can be tested by executing the `LogoutUseCase` with a valid session token, verifying that the token is deleted from the `session_tokens` persistence via the repository, and verifying that the `logout.success` event is emitted.

**Acceptance Scenarios**:

1. **Given** a valid session token "valid-token-123", **When** the logout use case is executed, **Then** the token is permanently deleted from the database using the repository interface.
2. **Given** a successful logout execution, **When** the workflow completes, **Then** a "logout.success" event is emitted to the audit log interface.

---

### User Story 2 - Idempotent Logout (Priority: P2)

As a system security mechanism, I want logout requests for already deleted or invalid tokens to behave gracefully and not result in unhandled application crashes or generic server errors.

**Why this priority**: Prevents internal server errors and noise in error logs for double-logout attempts or expired sessions.

**Independent Test**: Can be tested by invoking the `LogoutUseCase` with a non-existent or already deleted token and verifying that the system behaves correctly (e.g. returns a Success or handles the non-existence gracefully without system errors, or raises a designated unauthenticated error depending on implementation requirements).

**Acceptance Scenarios**:

1. **Given** a session token that does not exist in the database, **When** the logout use case is executed, **Then** the use case deletes the token (which is a no-op or returns successfully) and completes without system error.

### Edge Cases

- **Empty Token**: If an empty session token is provided to the use case, it should return an error or handle it as an invalid request.
- **Repository Failure**: If the repository encountered a database failure during token deletion, the usecase must return a clear repository-level error and not emit a success audit event.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST accept a session token string as input.
- **FR-002**: The system MUST delete the token from the persistence repository using `SessionTokenRepository.DeleteByToken`.
- **FR-003**: The system MUST emit a "logout.success" event to the audit log interface upon successful deletion of the token.
- **FR-004**: If the repository returns an error, the system MUST propagate the error and MUST NOT emit a success audit event.

### Key Entities *(include if feature involves data)*

- **SessionToken**: Represents the token string identifier to be invalidated.
- **AuditLogger**: The interface used to emit security audit logs, which needs to support logging logout success.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of logout requests with valid tokens result in the token being permanently deleted from the `session_tokens` table.
- **SC-002**: 100% of successful logout operations trigger a "logout.success" audit event to the audit logger.
- **SC-003**: 100% unit test coverage for the `LogoutUseCase` implementation.

## Assumptions

- **Audit log interface update**: The audit log interface (`AuditLogger`) will be extended to support a method for logging logout success (e.g. `LogLogoutSuccess(ctx context.Context, token string)` or similar), or a generic audit event method if one exists.
- **Session Token Context**: The session token string is extracted from the request context before it is passed to the use case.
