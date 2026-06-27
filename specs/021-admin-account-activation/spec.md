# Feature Specification: Admin Account Activation (Accept Invite)

**Feature Branch**: `021-admin-account-activation`

**Created**: 2026-06-27

**Status**: Approved — Updated 2026-06-27 (TSK-ACT-008: Integration Test Layer)

**Input**: User description: "Define the `InviteToken` domain entity and the `InviteTokenRepository` interface required to fetch and consume tokens. Define specific domain errors `ErrInviteExpired`, `ErrInviteUsed`, and `ErrPasswordWeak`. Define the event payload structs for the `admin.activated` and `invite.accepted` audit events, as well as the in-app notification Kafka event (`notification.send`) targeting the inviter."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Account Activation via Invite (Priority: P1)

Newly invited administrators can click their invitation link, specify a password, and activate their account.

**Why this priority**: Crucial for onboarding new administrators securely. Without this, no invited administrator can access the portal.

**Independent Test**: Can be tested by invoking the activation endpoint with a valid token, verifying database state changes (admin active, token used) and event emission.

**Acceptance Scenarios**:

1. **Given** a valid and unused invite token, **When** a user submits a strong password (min 10 chars, 1 uppercase, 1 number, 1 special), **Then** the account status becomes active, the invite token is marked as used, and the password hash is updated.
2. **Given** an expired invite token (older than 48 hours), **When** activation is attempted, **Then** return `ErrInviteExpired` error.
3. **Given** an already used invite token, **When** activation is attempted, **Then** return `ErrInviteUsed` error.
4. **Given** a weak password (does not meet constraints), **When** activation is attempted, **Then** return `ErrPasswordWeak` error.

---

### User Story 2 - HTTP Handler Routing (Accept Invite Endpoint) (Priority: P1)

The system exposes `POST /v1/auth/accept-invite` over HTTP so that API clients can activate accounts by submitting an invite token and password. The handler is responsible only for request binding, validation, response serialization, and error-code mapping — zero business logic lives in this layer.

**Why this priority**: Without the HTTP entry point, the use case cannot be reached by any client. The handler is the final integration step for this feature.

**Independent Test**: Can be fully tested with `httptest` against a stubbed `AcceptInviteUseCase`, verifying HTTP status codes and response bodies without a real database.

**Acceptance Scenarios**:

1. **Given** a valid `AcceptInviteRequest` body, **When** `AcceptInviteUseCase.Execute` returns `nil`, **Then** the handler responds `200 OK` with `{"message": "Account activated successfully."}`.
2. **Given** a malformed JSON body, **When** binding fails, **Then** the handler responds `400 Bad Request` with an error envelope.
3. **Given** a structurally valid body that fails struct validation (e.g., missing required fields), **When** validation fails, **Then** the handler responds `400 Bad Request`.
4. **Given** a valid body where `AcceptInviteUseCase.Execute` returns `ErrInviteExpired`, **When** the use case is called, **Then** the handler responds `400 Bad Request` with `code: "INVITE_EXPIRED"`.
5. **Given** a valid body where `AcceptInviteUseCase.Execute` returns `ErrInviteUsed`, **When** the use case is called, **Then** the handler responds `400 Bad Request` with `code: "INVITE_USED"`.
6. **Given** a valid body where `AcceptInviteUseCase.Execute` returns `ErrPasswordWeak`, **When** the use case is called, **Then** the handler responds `422 Unprocessable Entity` with `code: "PASSWORD_WEAK"`.
7. **Given** a valid request on a handler wired with a nil AcceptInviteUseCase, **When** the request is processed, **Then** the handler returns `500 Internal Server Error` without panicking.

---

### User Story 3 - Full Activation Flow Integration Test (Priority: P1)

The system has an integration test that proves every layer of the acceptance flow works together: an HTTP request flows through the Echo handler, use case, repository, and Kafka producer against a real PostgreSQL database and a mock Kafka producer.

**Why this priority**: Unit tests mock each layer in isolation. Only an integration test can prove that the handler, use case, DB schema, and notification pipeline are wired correctly together as a whole.

**Independent Test**: Runs with `go test ./test/integration/... -run TestAcceptInviteFlow` using testcontainers for PostgreSQL.

**Acceptance Scenarios**:

1. **Given** a seeded pending admin user and a valid (non-expired, unconsumed) invite token in PostgreSQL, **When** `POST /v1/auth/accept-invite` is called with a strong password, **Then** the response is `200 OK`; `admin_users.status` is `active`; `admin_users.password_hash` is updated; `invite_tokens.consumed_at` is non-null; and exactly one Kafka notification message is published.
2. **Given** a seeded invite token with `expires_at` in the past, **When** `POST /v1/auth/accept-invite` is called, **Then** the response is `400 INVITE_EXPIRED`; `admin_users.status` remains `pending`; `invite_tokens.consumed_at` remains null.

---

### Edge Cases

- **Token Replay**: Attempting to reuse an invite token multiple times must fail on the second attempt.
- **Expiration Check**: Tokens expired by even 1 second must be rejected.
- **Weak Password Inputs**: Password validation must strictly enforce complexity constraints.
- **Handler Serialization**: Error responses from the handler must match the shared error envelope shape (`code`, `message`, `details`, `trace_id`) as defined in the OpenAPI spec.
- **No Business Logic in Handler**: The handler must not perform token expiry checks, password hashing, or any domain computation; all logic is delegated to the use case.
- **Integration — DB State After Success**: After a successful activation, the database row for the admin user must reflect `status = active` and a new `password_hash`; the `invite_tokens` row must have a non-null `consumed_at`.
- **Integration — Expired Token DB Seed**: Seeding a token with `expires_at` in the past must cause the activation endpoint to return `400 INVITE_EXPIRED` without modifying any database row.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support fetching an invite token by its unique token string.
- **FR-002**: System MUST check if the invite token has expired (creation + 48 hours).
- **FR-003**: System MUST check if the invite token has already been used.
- **FR-004**: System MUST transition the admin user status to active upon successful activation.
- **FR-005**: System MUST record `admin.activated` and `invite.accepted` audit events.
- **FR-006**: System MUST dispatch a `notification.send` Kafka event to notify the inviter that their invite was accepted.
- **FR-007**: The `POST /v1/auth/accept-invite` HTTP endpoint MUST bind the request body to `AcceptInviteRequest`, invoke `AcceptInviteUseCase`, and map domain errors to HTTP status codes: `ErrInviteExpired` → 400, `ErrInviteUsed` → 400, `ErrPasswordWeak` → 422; all other errors → 500. Successful activation MUST return 200 OK. The handler MUST be wired into the Echo router via Uber Fx and MUST contain zero business logic.
- **FR-008**: An integration test in `test/integration/accept_invite_flow_test.go` MUST prove the full activation flow end-to-end using a real PostgreSQL instance (testcontainers). The test MUST: seed a pending admin user and a valid invite token; POST `{ token, password }` to `POST /v1/auth/accept-invite`; verify `admin_users.status = active`, a changed `password_hash`, a non-null `consumed_at` on the token row, and a Kafka notification message published. A second test case MUST seed an expired token and verify the endpoint returns `400 INVITE_EXPIRED` without mutating any DB row.

### Key Entities *(include if feature involves data)*

- **InviteToken**: Domain entity representing a secure invitation token linked to an admin user.
  - `ID`: Unique identifier (UUID).
  - `AdminID`: ID of the admin user being invited (UUID).
  - `Token`: Unique cryptographically secure token string.
  - `ExpiresAt`: Time when the token expires (typically created_at + 48 hours).
  - `UsedAt`: Time when the token was consumed (null if unused).
  - `CreatedBy`: ID of the Super Admin who created the invite (UUID).
  - `CreatedAt`: Time when the token was created.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Account activation is processed in under 500ms.
- **SC-002**: 100% of expired or reused tokens are rejected.
- **SC-003**: 100% of password validation failures block account activation.
- **SC-004**: The HTTP handler returns the correct status code and error code for every mapped domain error 100% of the time.
- **SC-005**: The integration test suite passes consistently (0 flakes across 3 consecutive runs), proving that the HTTP layer, use case, database mutations, and Kafka notification work correctly together against a real PostgreSQL instance.

## Assumptions

- **Constraints**: Password validation rules require min 10 characters, 1 uppercase letter, 1 number, and 1 special character.
- **Mail System**: Email delivery and actual token generation is handled by the Admin Management service.
