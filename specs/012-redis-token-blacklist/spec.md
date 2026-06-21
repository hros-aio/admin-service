# Feature Specification: Redis Token Blacklist

**Feature Branch**: `012-redis-token-blacklist`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-012: Layer: Cache Description: Implement the TokenBlacklist interface in the Redis infrastructure layer. This implementation must store revoked JWT access tokens (e.g., after logout or rotation) and set the Redis TTL to match the exact remaining lifetime of the JWT (up to 15 minutes max). Input: TokenBlacklist application interface, Redis client connection. Output: internal/infrastructure/cache/token_blacklist_redis.go, internal/infrastructure/cache/token_blacklist_redis_test.go Definition of Done: The Redis cache safely stores JTI (JWT IDs) with accurate TTLs. Unit tests pass with a mocked Redis client. Graceful degradation handles Redis connection errors safely. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Revoke Access/Refresh Tokens (Priority: P1) 🎯 MVP

As an application security component, I want to store revoked token identifiers (such as JWT IDs) in Redis with an accurate TTL matching the remaining token lifetime, so that revoked tokens are instantly and reliably rejected during their valid window.

**Why this priority**: High. Essential for securing active sessions, handling logouts, and executing refresh token rotation.

**Independent Test**: Verification via unit tests with a mocked Redis client verifying that calling `Add` writes the token ID to Redis with the correct TTL (capped at 15 minutes), and calling `Exists` returns true if blacklisted and false if not.

**Acceptance Scenarios**:

1. **Given** a valid token identifier and a remaining lifetime of 10 minutes, **When** the token is added to the blacklist, **Then** it is stored in the cache with a TTL of exactly 10 minutes.
2. **Given** a token identifier and a remaining lifetime of 20 minutes, **When** the token is added to the blacklist, **Then** it is stored in the cache with a TTL capped at 15 minutes (900 seconds).
3. **Given** a token identifier that was blacklisted, **When** checked, **Then** the cache indicates it is blacklisted.
4. **Given** a token identifier that was not blacklisted, **When** checked, **Then** the cache indicates it is not blacklisted.
5. **Given** a connection issue with the cache store, **When** a token is added or checked, **Then** the operation gracefully degrades (e.g., logging the error) and returns a handled error instead of crashing or blocking the application.

### Edge Cases

- **TTL Capping**: If the token remaining lifetime is longer than 15 minutes, the TTL must be capped at 15 minutes to conserve cache storage.
- **Connection Error Handling**: If Redis goes down, operations must fail-safe (graceful degradation) without causing server panic or hanging requests.
- **Key Collisions**: Key naming must be isolated to prevent conflicts with other cache usages.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-012-001**: The system MUST implement the application layer `TokenBlacklist` interface in the infrastructure layer using Redis.
- **FR-AUTH-012-002**: The implementation MUST store blacklisted token identifiers in Redis.
- **FR-AUTH-012-003**: The TTL set on the Redis key MUST match the token's remaining lifetime up to a maximum limit of 15 minutes (900 seconds).
- **FR-AUTH-012-004**: Keys in Redis MUST be prefixed (e.g. `blacklist:`) to avoid namespace collisions.
- **FR-AUTH-012-005**: If the Redis client returns an error (e.g., connection timed out or down), the implementation MUST handle the error gracefully, log the error using structured logs (`slog`), and return a clear, non-panicking error.

### Key Entities

- **TokenBlacklist**: Cache interface defining the `Add` and `Exists` methods.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-012-001**: Unit tests in `internal/infrastructure/cache/token_blacklist_redis_test.go` cover all primary code paths and edge cases (success, TTL capping, connection failures) and pass successfully.
- **SC-AUTH-012-002**: Verification of exact TTL mapping under test (10 mins -> 10 mins; 20 mins -> 15 mins cap).
- **SC-AUTH-012-003**: Graceful error handling is explicitly tested and verified.

## Assumptions

- The Redis client from `"github.com/redis/go-redis/v9"` is already set up and provided via dependency injection.
- Revoked JWTs are identified by their unique JTI (JWT ID) or raw token string.
