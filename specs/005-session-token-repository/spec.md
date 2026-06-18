# Feature Specification: Session Token Repository (GORM)

**Feature Branch**: `005-session-token-repository`

**Created**: 2026-06-18

**Status**: Draft

**Input**: User description: "Use the existing repository documents as source of truth. TSK-AUTH-005: Layer: Repository Description: Implement SessionTokenRepository using GORM to insert new session tokens on login and delete tokens by value on explicit logout. Input: SessionTokenRepository interface. Output: internal/infrastructure/database/session_token_repository.go, internal/infrastructure/database/session_token_repository_test.go Definition of Done: Create and Delete operations map correctly to GORM statements. Unit tests pass. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## Clarifications

### Session 2026-06-18
- Q: Repository Interface Alignment? → A: Add `DeleteByToken(ctx context.Context, token string) error` to the domain interface to best match the "delete by value" requirement, as session tokens are short-lived security artifacts without audit retention requirements.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Store Session Token on Login (Priority: P1)

As the system, I need to persist a new session token (refresh token) whenever an admin user successfully logs in so that the session can be tracked and refreshed later without re-authenticating.

**Why this priority**: Essential for the "Remember Me" and refresh token functionality. Without this, users would have to log in frequently.

**Independent Test**: Can be verified by creating a `SessionToken` domain entity, calling the repository's save method, and asserting that the record exists in the `session_tokens` table with the correct `admin_id` and token value.

**Acceptance Scenarios**:

1. **Given** a valid `SessionToken` entity for an authenticated user, **When** I call the `Save` method, **Then** the token should be stored in the database.
2. **Given** a database connection error, **When** I attempt to save a token, **Then** the system should return an appropriate infrastructure-related error wrapped for the domain.

---

### User Story 2 - Remove Session Token on Explicit Logout (Priority: P2)

As the system, I need to remove a specific session token from the database when an admin user logs out so that the token is immediately invalidated and cannot be used again.

**Why this priority**: Critical security requirement to ensure that logged-out sessions cannot be hijacked or reused via the refresh token.

**Independent Test**: Can be verified by calling the `DeleteByToken` method with a specific token value and asserting that the corresponding record is no longer present in the database.

**Acceptance Scenarios**:

1. **Given** an active session token "abc-123" exists in the database, **When** I call `DeleteByToken` for token "abc-123", **Then** the record should be removed from the `session_tokens` table.
2. **Given** a token value that does not exist in the database, **When** I attempt to delete it via `DeleteByToken`, **Then** the system should complete without error (idempotent behavior).

---

### Edge Cases

- **Token Expiry**: How does the repository handle tokens that have expired but are still in the database? (Assumption: The repository only manages persistence; expiry logic is handled by the application/domain layer).
- **Duplicate Tokens**: What happens if the system attempts to save a token value that already exists? (Assumption: Should return a conflict error or handle via UNIQUE constraint mapping).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a mechanism to persist `SessionToken` entities using GORM.
- **FR-002**: System MUST implement `DeleteByToken(ctx, token)` to allow deleting session tokens by their unique token value string.
- **FR-003**: System MUST update the `SessionTokenRepository` domain interface to include the `DeleteByToken(ctx context.Context, token string) error` method signature.
- **FR-004**: System MUST use `context.Context` for all GORM operations to respect request timeouts and cancellation.
- **FR-005**: System MUST map GORM-specific errors (e.g., connection failures, constraint violations) to domain-level errors or generic infrastructure errors.
- **FR-006**: System MUST NOT leak `gorm.DB` or other infrastructure-specific types outside the repository implementation.

### Key Entities *(include if feature involves data)*

- **Session Token**: Represents a persistent session for an admin. Key attributes include `AdminID`, `RefreshToken`, `ExpiresAt`, and `IPAddress`.
- **Admin User**: The owner of the session token.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of valid `Save` operations result in a correctly persisted record in the `session_tokens` table.
- **SC-002**: 100% of `DeleteByToken` requests for a specific token value result in the removal of that token from the database.
- **SC-003**: Deletion operations are idempotent; calling delete on a non-existent token does not produce an error.
- **SC-004**: All repository operations complete within the performance targets of an indexed relational database query.

## Assumptions

- **Schema Existence**: The `session_tokens` table exists with a UNIQUE index on the `refresh_token` column.
- **Domain Model**: The `SessionToken` struct in `internal/domain/session_token.go` is the authoritative definition of the entity.
- **Dependency Injection**: The GORM database connection will be injected via Uber Fx.
- **Clean Architecture**: The repository will be implemented in the infrastructure layer but will only interact with domain entities and models.
