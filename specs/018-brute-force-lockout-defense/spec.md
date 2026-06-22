# Feature Specification: Brute-Force Lockout Defense

**Feature Branch**: `018-brute-force-lockout-defense`

**Created**: 2026-06-21

**Updated**: 2026-06-22 (TSK-AUTH-021 — LoginUseCase lockout orchestration)

**Status**: In Progress

**Input**: User description: "TSK-AUTH-018 → TSK-AUTH-020: Define domain primitives (BruteForceCache, ErrAccountLocked, event payloads), Redis cache implementation, and Kafka producer adapter for lockout email notification."

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

### User Story 4 - Kafka Adapter Serializes Lockout Email Event (Priority: P1)

The lockout email notification must flow through a well-defined Kafka adapter that wraps the `EmailSendEvent` domain payload inside a standard JSON event envelope before dispatching it via the Sarama publisher interface.

**Why this priority**: Without the adapter layer, the domain event payload cannot be dispatched over Kafka — the linkage between the domain (what) and infrastructure (how) is missing.

**Independent Test**: Construct an `EmailSendEvent` with a recipient email and unlock timestamp, call the adapter's publish method against a mock `sarama.SyncProducer`, and verify the captured message body deserializes back to a correctly populated `EventEnvelope[EmailSendEvent]`.

**Acceptance Scenarios**:

1. **Given** a valid `EmailSendEvent` (recipient, subject, template, template data including unlock timestamp), **When** the Kafka email producer's `PublishLockoutEmail` is called, **Then** a Sarama `ProducerMessage` is sent to the topic `email.send.v1` with the event serialized inside the standard `EventEnvelope` JSON structure.
2. **Given** the publisher interface is called with the envelope, **When** Sarama returns an error, **Then** the error is wrapped and returned to the caller without panicking.
3. **Given** a zero-value or empty-string recipient email, **When** the adapter is called, **Then** it returns a validation error before attempting to publish.

---

### User Story 3 - Admin Override / Immediate Unlock (Priority: P2)

If a administrator gets locked out, a Super Admin can manually unlock the account before the 30-minute lockout period expires.

**Why this priority**: Necessary operational override for legitimate users who locked themselves out and need immediate access to perform critical operations.

**Independent Test**: Lock an account, verify login is blocked, perform a Super Admin unlock action, and verify that the user can immediately log in again.

**Acceptance Scenarios**:

1. **Given** an account is in locked state, **When** a Super Admin manually unlocks it, **Then** the lockout record and consecutive failed attempt counts are cleared, allowing immediate login.

---

### User Story 5 - LoginUseCase Lockout State Machine (Priority: P1)

The `LoginUseCase` orchestrates all brute-force checks in a strict sequence: it first checks whether the account is already locked, then verifies credentials, then either increments the failure counter (locking on the 5th consecutive failure and notifying by email) or resets it on success.

**Why this priority**: Without this orchestration, the domain primitives, Redis cache, and Kafka adapter remain inert — they only become effective when wired into the login flow.

**Independent Test**: Can be tested in isolation by mocking `BruteForceCache`, the Kafka publisher, and the audit log, then asserting the correct sequence of calls for each branch: already-locked, <5 failures, exactly 5 failures, and successful login.

**Acceptance Scenarios**:

1. **Given** an account is already locked, **When** any login attempt is made (correct or incorrect password), **Then** `ErrAccountLocked` is returned immediately without checking the password.
2. **Given** an account has fewer than 5 consecutive failed attempts, **When** another invalid password attempt is made, **Then** the failure counter is incremented and a credential-error is returned (no lockout applied).
3. **Given** an account has exactly 4 prior consecutive failed attempts, **When** a 5th invalid password attempt is made, **Then** the account is locked, `account.locked` is appended to the audit log, a `email.send` event is published to Kafka (best-effort), and `ErrAccountLocked` is returned.
4. **When** an account is not locked and the correct password is provided, **Then** login succeeds, the failure counter is reset to 0, and a valid session is returned.

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
- **FR-008**: The Kafka adapter MUST wrap the `EmailSendEvent` domain payload inside a standard `EventEnvelope` (containing `id`, `type`, `source`, `version`, `correlation_id`, `occurred_at`, and `data` fields) before publishing.
- **FR-009**: The Kafka adapter MUST publish the lockout email envelope to the topic `email.send.v1` using the project's `EventPublisher` interface backed by `sarama.SyncProducer`. The message key MUST be the recipient email address.
- **FR-010**: `LoginUseCase` MUST call `BruteForceCache.IsLocked()` as the first step before any password verification; if locked, it MUST return `ErrAccountLocked` immediately.
- **FR-011**: `LoginUseCase` MUST call `BruteForceCache.IncrementFailedAttempts()` after each failed password verification.
- **FR-012**: `LoginUseCase` MUST call `BruteForceCache.SetLockout()`, append `account.locked` to the audit log, and publish an `email.send` event via Kafka when the failure count reaches exactly 5 consecutive failures.
- **FR-013**: `LoginUseCase` MUST call `BruteForceCache.ClearFailures()` immediately after a successful password verification to reset the counter.

### Key Entities *(include if feature involves data)*

- **BruteForceCacheState**: Represents the ephemeral status of an email's auth attempts in the cache.
  - Email (string)
  - FailedAttempts (int)
  - LockoutExpiry (timestamp)

- **AuthEvent**: Represents events emitted by the auth domain.
  - **AccountLockedEvent**: Payloads for `account.locked` (audit log entry: user details, timestamp).
  - **EmailSendEvent**: Payload for `email.send` (recipient email, body template/data, subject).

- **EventEnvelope**: Standard Kafka message wrapper applied by the adapter layer.
  - ID (UUID string) — unique event identifier
  - Type (string) — event topic identifier, e.g. `email.send`
  - Source (string) — originating service name
  - Version (int) — schema version
  - CorrelationID (string) — request trace correlation
  - OccurredAt (timestamp) — event creation time
  - Data (generic payload) — the domain event struct

- **EmailKafkaProducer**: Adapter-layer producer responsible for mapping `EmailSendEvent` → `EventEnvelope[EmailSendEvent]` and dispatching via the `EventPublisher` interface.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Verification of lockout status on login attempts must resolve in under 5ms from the cache.
- **SC-002**: Lockout duration is precisely 30 minutes (+/- 5 seconds).
- **SC-003**: Invalidation/lockout check must precede credential verification to prevent expensive password hashing (bcrypt) during active locks, saving CPU resources.
- **SC-004**: All locking and unlocking events must generate corresponding audit logs and Kafka notification events with 100% reliability.
- **SC-006**: `LoginUseCase` unit tests achieve 100% branch coverage across all four branches: already-locked, <5 failures, exactly 5 failures (triggering lockout + audit + Kafka), and successful login (counter reset).
- **SC-005**: The lockout email Kafka message produced by the adapter must deserialize back into a fully-populated `EventEnvelope[EmailSendEvent]` with all required fields present and non-zero, as verified by unit tests against a mock producer.

## Assumptions

- We assume Redis will be the primary caching implementation for the lockout state in production, but the domain/application layers will remain cache-agnostic.
- The system timezone is handled consistently (UTC or matching local time formatting).
- Email notifications will be processed asynchronously by a separate notification service reading from Kafka.
- The `EventEnvelope` generic struct is defined in `internal/adapter/kafka/producer/` or a shared platform package; the adapter owns the envelope construction.
- The `EventPublisher` interface used by the adapter wraps `sarama.SyncProducer` as defined in the tech-stack document. The adapter receives it via dependency injection.
- The Kafka topic for lockout email events is `email.send.v1`, following the project's topic-naming convention `<domain>.<event-name>.v<version>`.
- The message key is the recipient email address to ensure per-user partition ordering.
- Kafka publish errors during lockout notification must be logged but MUST NOT propagate as login errors; the `ErrAccountLocked` result is determined by the Redis state, not the Kafka outcome.
