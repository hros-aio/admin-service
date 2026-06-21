# Feature Specification: Refresh Session Use Case

**Feature Branch**: `014-refresh-session-usecase`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-014: Layer: UseCase Description: Implement RefreshSessionUseCase. Workflow: Accept refresh token string → fetch via repository → check expiry → generate new JWT access token → generate new refresh token string → update session in DB → emit session.refreshed to the audit log interface. Input: RefreshRequest data mapped to domain input, SessionTokenRepository, Audit event publisher. Output: internal/application/usecase/refresh_session_usecase.go, internal/application/usecase/refresh_session_usecase_test.go Definition of Done: The business logic securely rotates the token pair, updates the DB, emits the audit event, and returns the new tokens. 100% unit test coverage using mocked repositories. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rotate Active Token Session (Priority: P1) 🎯 MVP

As an authenticated user or client application, I want to trade my valid refresh token for a new access token and rotated refresh token, so that I can extend my session securely without needing to re-login.

**Why this priority**: High. Essential for enabling secure, frictionless authentication session persistence.

**Independent Test**: Verification via unit tests mocking the `SessionTokenRepository`, `TokenProvider`, and `AuditLogger` to assert that correct input maps to successful token updates, database persistence, and audit logging.

**Acceptance Scenarios**:

1. **Given** a valid, unexpired refresh token, **When** `Execute` is called on the refresh session usecase, **Then** GORM updates the token in the database, a `session.refreshed` event is published, and a new access/refresh token pair is returned.
2. **Given** an expired refresh token, **When** `Execute` is called, **Then** the usecase returns `ErrTokenExpired`.
3. **Given** a revoked refresh token, **When** `Execute` is called, **Then** the usecase returns `ErrInvalidRefreshToken`.
4. **Given** a non-existent refresh token string, **When** `Execute` is called, **Then** the usecase returns `ErrInvalidRefreshToken`.

### Edge Cases

- **Database Save Failure**: If updating the session in the database fails, the usecase must return a wrapped error and must not return any new tokens to prevent token discrepancy.
- **Audit Logging Failure**: Audit log interface is non-blocking or handles logging errors cleanly without failing the main token refresh flow.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-014-001**: The system MUST implement `RefreshSessionUseCase` in the application layer.
- **FR-AUTH-014-002**: The usecase MUST accept `RefreshInput` containing the `RefreshToken` string.
- **FR-AUTH-014-003**: The usecase MUST fetch the session token using `SessionTokenRepository.FindByToken`.
- **FR-AUTH-014-004**: The usecase MUST validate the fetched session token (check if expired via `IsExpired()` or revoked via `IsRevoked()`).
- **FR-AUTH-014-005**: The usecase MUST generate a new access token and a new refresh token using `TokenProvider`.
- **FR-AUTH-014-006**: The usecase MUST rotate the session token model state using `Rotate()` helper, and persist it to the database via `SessionTokenRepository.UpdateToken`.
- **FR-AUTH-014-007**: The usecase MUST log the token refresh event via the `AuditLogger` interface by calling `LogSessionRefreshed(ctx context.Context, userID string)`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-014-001**: 100% test coverage achieved for `refresh_session_usecase.go` under `refresh_session_usecase_test.go`.
- **SC-AUTH-014-002**: The usecase handles all validation paths (success, expired, revoked, not found) correctly and returns proper domain error instances.

## Assumptions

- We will extend the `AuditLogger` interface with the `LogSessionRefreshed` method to fulfill the audit logging requirement.
- The `TokenProvider` dependency exists and exposes `GenerateAccessToken` and `GenerateRefreshToken` methods.
