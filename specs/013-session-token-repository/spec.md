# Feature Specification: Session Token Repository Updates

**Feature Branch**: `013-session-token-repository`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-013: Layer: Repository Description: Update SessionTokenRepository with a FindByToken(ctx context.Context, token string) method and an UpdateToken(ctx context.Context, session *domain.SessionToken) method to facilitate refresh rotation. Input: SessionTokenRepository interface. Output: internal/infrastructure/database/session_token_repository.go, internal/infrastructure/database/session_token_repository_test.go Definition of Done: Repository methods successfully query and update the database utilizing GORM correctly. Unit tests pass using sqlmock. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Session Token DB Operations for Rotation (Priority: P1) 🎯 MVP

As the token rotation usecase, I want to fetch an existing session token by its refresh token value and update its state (such as token string, expiration, or revocation) in the database, so that refresh token rotation is persisted securely.

**Why this priority**: High. Essential for enabling secure token rotation flows.

**Independent Test**: Verification via unit tests using `sqlmock` asserting that `FindByToken` and `UpdateToken` translate correctly into GORM SELECT and UPDATE database operations.

**Acceptance Scenarios**:

1. **Given** a valid refresh token string, **When** `FindByToken` is called, **Then** GORM performs a SELECT query and returns the mapped `SessionToken` domain entity.
2. **Given** an updated `SessionToken` entity (with a rotated refresh token or updated expiry), **When** `UpdateToken` is called, **Then** GORM performs an UPDATE query to save the changes to the database.

### Edge Cases

- **Token Not Found**: If no token matches the query in `FindByToken`, a GORM record-not-found error should be returned cleanly.
- **Nil Session Input**: If `UpdateToken` is called with a nil pointer, it must return an error gracefully rather than panic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-013-001**: The domain layer `SessionTokenRepository` interface in `internal/domain/session_token.go` MUST declare `UpdateToken(ctx context.Context, session *SessionToken) error`.
- **FR-AUTH-013-002**: The infrastructure layer `GormSessionTokenRepository` MUST implement `UpdateToken` using GORM.
- **FR-AUTH-013-003**: The implementation MUST utilize `sqlmock` in unit tests to mock and verify SQL statements.
- **FR-AUTH-013-004**: The existing `FindByToken` method MUST continue to be fully implemented and tested.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-013-001**: Unit tests in `internal/infrastructure/repository/auth/session_token_repository_test.go` achieve at least 70% coverage.
- **SC-AUTH-013-002**: GORM queries match standard database update/select statements under `sqlmock`.

## Assumptions

- The active implementation of `SessionTokenRepository` resides under `internal/infrastructure/repository/auth` and is wired to Echo/Fx; we will update this path to respect the approved repository structure.
