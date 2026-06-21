# Feature Specification: Brute-Force Lockout Defense

**Feature Branch**: `018-brute-force-lockout-defense`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-018: Define the BruteForceCache interface required by the application layer. Define the specific domain error ErrAccountLocked. Define the event payload structs for account.locked (audit) and email.send (Kafka notification). Input: Feature specifications (SRS-AUTH-002, Brute-Force Lockout Defense Roadmap)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Account Lockout on Failed Attempts (Priority: P1)

When an administrator makes 5 consecutive failed login attempts with incorrect credentials within 15 minutes, their account is temporarily locked for 30 minutes to prevent brute-force password guessing.

**Why this priority**: Core functionality to protect against credential stuffing and brute-force password attacks.

**Independent Test**: Can be tested by attempting to log in with invalid credentials 5 times in a row, then verifying that subsequent login attempts fail with a lockout error code, even if correct credentials are provided during the lockout window.

**Acceptance Scenarios**:

1. **Given** a user has no active failed attempts, **When** they make 5 consecutive invalid login attempts within 15 minutes, **Then** their account is locked for 30 minutes, and any subsequent attempts during this window return a lockout error status.
2. **Given** a user has 4 failed attempts, **When** they perform a successful login, **Then** their failed attempt counter is reset to 0.
3. **Given** a user has 3 failed attempts, **When** 15 minutes pass with no login attempts, **Then** their failed attempt counter is cleared or expires, and they start fresh.

---

### User Story 2 - Lockout Notification via Email (Priority: P1)

When an admin account gets locked, the system automatically emails the account holder notifying them of the lockout event and providing the unlock timestamp.

**Why this priority**: Required for security transparency, so the user knows if someone else is trying to guess their password, and knows when they can try to log in again.

**Independent Test**: Trigger a lockout on an account and verify that an email sending event payload is generated containing the recipient's email address and the unlock timestamp.

**Acceptance Scenarios**:

1. **Given** a user is on their 5th consecutive failed login attempt, **When** the attempt fails and triggers a lockout, **Then** the system publishes a notification event to send a lockout email to the user.

---

### User Story 3 - Admin Override / Immediate Unlock (Priority: P2)

If a administrator gets locked out, a Super Admin can manually unlock the account before the 30-minute lockout period expires.

**Why this priority**: Necessary operational override for legitimate users who locked themselves out and need immediate access to perform critical operations.

**Independent Test**: Lock an account, verify login is blocked, perform a Super Admin unlock action, and verify that the user can immediately log in again.

**Acceptance Scenarios**:

1. **Given** an account is in locked state, **When** a Super Admin manually unlocks it, **Then** the lockout record and consecutive failed attempt counts are cleared, allowing immediate login.

---

### Edge Cases

- **Attempts during active lockout**: When an account is locked, any login attempt (even with correct password) must fail and must NOT extend the lockout duration beyond the original 30-minute expiration.
- **Concurrent failed attempts**: If a user attempts to log in concurrently from multiple clients, the failed attempts must be tracked accurately and atomic counters must be incremented.
- **Non-existent accounts**: If an invalid email is targeted, the system must simulate or handle attempts in a way that prevents timing oracles or email enumeration, while avoiding caching failure states for non-existent users if that creates a DoS vulnerability.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST track consecutive failed login attempts by email address within a sliding window of 15 minutes.
- **FR-002**: System MUST temporarily lock an account for 30 minutes when consecutive failed attempts reach 5.
- **FR-003**: System MUST reject all login attempts for a locked email during the lockout duration with a specific lockout error status (`ACCOUNT_LOCKED`).
- **FR-004**: System MUST reset the consecutive failed login attempts counter to 0 immediately upon a successful login.
- **FR-005**: System MUST trigger an audit log event (`account.locked`) when an account is locked.
- **FR-006**: System MUST trigger a notification event (`email.send`) containing the user's email, lockout event details, and the unlock timestamp.
- **FR-007**: System MUST provide an interface to query, increment, and reset the lockout/failure state in a cache.

### Key Entities *(include if feature involves data)*

- **BruteForceCacheState**: Represents the ephemeral status of an email's auth attempts in the cache.
  - Email (string)
  - FailedAttempts (int)
  - LockoutExpiry (timestamp)

- **AuthEvent**: Represents events emitted by the auth domain.
  - **AccountLockedEvent**: Payloads for `account.locked` (audit log entry: user details, timestamp).
  - **EmailSendEvent**: Payload for `email.send` (recipient email, body template/data, subject).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Verification of lockout status on login attempts must resolve in under 5ms from the cache.
- **SC-002**: Lockout duration is precisely 30 minutes (+/- 5 seconds).
- **SC-003**: Invalidation/lockout check must precede credential verification to prevent expensive password hashing (bcrypt) during active locks, saving CPU resources.
- **SC-004**: All locking and unlocking events must generate corresponding audit logs and Kafka notification events with 100% reliability.

## Assumptions

- We assume Redis will be the primary caching implementation for the lockout state in production, but the domain/application layers will remain cache-agnostic.
- The system timezone is handled consistently (UTC or matching local time formatting).
- Email notifications will be processed asynchronously by a separate notification service reading from Kafka.
