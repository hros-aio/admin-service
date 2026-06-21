# Feature Specification: Auth Token Rotation

**Feature Branch**: `010-auth-token-rotation`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-010: **Layer**: Domain **Description**: Update the `SessionToken` domain entity to support explicit expiration tracking (`ExpiresAt`) and add a `Rotate()` helper method to generate new token strings. Define a `TokenBlacklist` cache interface to handle immediate JWT revocation. Define specific domain errors like `ErrInvalidRefreshToken` and `ErrTokenExpired`. **Input**: Feature specifications. **Output**: `internal/domain/session_token.go`, `internal/application/interfaces/cache.go`, `internal/domain/errors/auth_errors.go` **Definition of Done**: Domain models support expiry logic and rotation mechanisms without any external dependencies. The cache interface is explicitly defined for application use. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Session Token Rotation (Priority: P1)

As a security-conscious system administrator, I want the system to support rotation of session refresh tokens upon use, so that replay attacks using stolen refresh tokens are mitigated.

**Why this priority**: High. Essential for securing persistent user sessions and implementing standard refresh token rotation security practices.

**Independent Test**: Can be tested via unit tests on the `SessionToken` domain entity by creating a token, calling `Rotate()`, and asserting that a new cryptographically secure token string is generated and the entity's inner state (like `RefreshToken` and expiration) is updated correctly.

**Acceptance Scenarios**:

1. **Given** a valid `SessionToken` entity, **When** `Rotate` is called with a new expiration duration, **Then** a new secure random token string is returned, and the entity is updated with the new token string and the new expiration time.
2. **Given** a `SessionToken` entity, **When** checking its expiration status, **Then** the system correctly identifies if the token has expired based on the `ExpiresAt` field.

---

### User Story 2 - Token Blacklisting (Priority: P2)

As the application security middleware, I want access to a standardized cache interface to blacklist specific tokens immediately, so that revoked sessions (e.g., from forced logout or security breach detection) are instantly blocked.

**Why this priority**: Medium. Crucial for handling immediate revocation of active JWTs or refresh tokens before their natural expiry.

**Independent Test**: Can be tested by defining a unit test or mock checking the `TokenBlacklist` interface invocations and verifying that `Add` and `Exists` operations map correctly without carrying any database implementation details.

**Acceptance Scenarios**:

1. **Given** a revoked refresh token, **When** added to the blacklist with a specific TTL, **Then** a subsequent check confirms it is blacklisted until the TTL expires.

---

### User Story 3 - Specific Domain Error Identification (Priority: P2)

As a developer using the domain model, I want specific and distinct domain errors for token validation failures (like expired vs. invalid tokens), so that the application layer can respond with precise client-facing error codes.

**Why this priority**: Medium. Essential for distinguishing between a token that is simply expired (which triggers a fresh login) and one that is structurally invalid or does not exist.

**Independent Test**: Unit tests assert that the errors package defines `ErrInvalidRefreshToken` and standardizes token validation error constants.

**Acceptance Scenarios**:

1. **Given** an expired session token, **When** validated by the domain layer, **Then** the specific `ErrTokenExpired` error is returned.
2. **Given** a malformed or invalid refresh token string, **When** validated by the domain layer, **Then** the specific `ErrInvalidRefreshToken` error is returned.

### Edge Cases

- **Random Generator Failures**: If the system's cryptographically secure random number generator (`crypto/rand`) fails during token rotation, the method must return a clear error rather than succeeding with weak/predictable fallback values.
- **Immediate Re-rotation**: What happens if a token is rotated multiple times in a short window? (Assumption: The rotation is an atomic update operation of the session entity state, and persistence/concurrency handling is done at the repository layer).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-010-001**: The `SessionToken` domain entity MUST support explicit expiration tracking via the `ExpiresAt` field.
- **FR-AUTH-010-002**: The `SessionToken` domain entity MUST provide a `Rotate(newExpiry time.Time) (string, error)` method that generates a new cryptographically secure random token string, updates the entity's `RefreshToken` and `ExpiresAt` values, and returns the newly generated token.
- **FR-AUTH-010-003**: The system MUST define a `TokenBlacklist` interface under the application layer interfaces package.
- **FR-AUTH-010-004**: The `TokenBlacklist` interface MUST include `Add(ctx context.Context, token string, ttl time.Duration) error` and `Exists(ctx context.Context, token string) (bool, error)`.
- **FR-AUTH-010-005**: The system MUST define the `ErrInvalidRefreshToken` domain error in the `internal/domain/errors/auth_errors.go` file.
- **FR-AUTH-010-006**: The system MUST ensure `ErrTokenExpired` is defined and used for expired token scenarios.
- **FR-AUTH-010-007**: Domain entity validation and rotation methods MUST NOT import any external dependencies, frameworks, or database libraries.

### Key Entities *(include if feature involves data)*

- **SessionToken**: The domain entity representing a persistent session refresh token. Key attributes: `AdminID`, `RefreshToken`, `ExpiresAt`, `IPAddress`, `UserAgent`, `CreatedAt`, `RevokedAt`, and `RevokeReason`.
- **TokenBlacklist**: A cache interface used to query and persist short-lived revocations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-010-001**: 100% of unit tests for token expiration verification and secure rotation helper pass successfully.
- **SC-AUTH-010-002**: The domain error variables are exported and easily comparable via `errors.Is`.
- **SC-AUTH-010-003**: Blacklist cache interface is defined cleanly in `internal/application/interfaces/cache.go` without any reference to Redis or GORM implementations.

## Assumptions

- **Secure Source**: The standard library `crypto/rand` is fully functional and sufficient for secure random token generation (which does not depend on any third-party frameworks).
- **Domain Scope**: The persistence of blacklisted tokens (using Redis cache wrapper) is out of scope for this task and will be implemented in infrastructure layers as a subsequent task.
