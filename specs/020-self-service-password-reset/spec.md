# Feature Specification: Self-Service Password Reset

**Feature Branch**: `020-self-service-password-reset`

**Created**: 2026-06-25

**Status**: Draft

**Input**: User description: "Feature: Self-Service Password Reset. Task: TSK-PR-001. Layer: Domain. Description: Define the PasswordResetCache interface required to store the single-use reset token. Define the specific domain errors ErrTokenExpired, ErrTokenUsed, and ErrPasswordWeak. Define the event payload structs for the password.reset_requested and password.reset_completed audit events, as well as the email.send Kafka event for the reset link."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Domain Primitives for Password Reset (Priority: P1)

As a developer, I want to define the cache interface `PasswordResetCache` to store the single-use reset tokens, domain errors `ErrTokenExpired`, `ErrTokenUsed`, and `ErrPasswordWeak`, and event payloads for `password.reset_requested`, `password.reset_completed`, and `email.send` Kafka events, so that the domain layer has the necessary primitives for the password reset feature.

**Why this priority**: Essential domain setup required before implementing the password reset application and repository logic.

**Independent Test**: The structs, interfaces, and errors compile without external infrastructure dependencies, and event payload structures serialize correctly to JSON.

**Acceptance Scenarios**:

1. **Given** a password reset flow, **When** caching the reset token context, **Then** `PasswordResetCache` can store, retrieve, and delete the context using a unique token.
2. **Given** validation or expiration checks fail, **When** returning errors, **Then** `ErrTokenExpired`, `ErrTokenUsed`, and `ErrPasswordWeak` are returned.
3. **Given** password reset events are triggered, **When** serializing event payloads, **Then** they can be serialized to JSON and contain correct metadata fields.

---

### User Story 2 - Define API Contracts and HTTP DTOs for Password Reset (Priority: P1)

As a front-end developer or API consumer, I want the password reset endpoints (`POST /v1/auth/password-reset/request` and `POST /v1/auth/password-reset/confirm`) documented in the OpenAPI spec and defined in HTTP DTOs, so that we can implement the password reset UI and clients correctly.

**Why this priority**: Prerequisite for implementing the password reset HTTP handlers.

**Independent Test**: The DTO structs compile with strict validation tags, and `api/openapi.yaml` passes validation checks.

**Acceptance Scenarios**:

1. **Given** an admin user wants to request a password reset, **When** they submit their email, **Then** they send a `PasswordResetRequest` and receive a success response.
2. **Given** a user has a password reset token, **When** they submit their new password, **Then** they send a `PasswordResetConfirmRequest` containing token, password, and password_confirmation.
3. **Given** structural validation constraints are violated (e.g. missing fields, mismatched confirmation, invalid email), **When** parsed by the handler, **Then** validation fails with a bad request response.
4. **Given** a password reset attempt fails due to business rules, **When** returning errors, **Then** `api/openapi.yaml` documents `400` errors for `TOKEN_EXPIRED` and `TOKEN_USED`, and `422` error for `PASSWORD_WEAK`.

### User Story 3 - Request Password Reset UseCase (Priority: P1)

As a user who forgot my password, I want to request a password reset by providing my email address, so that I can receive a secure reset link via email.

**Why this priority**: Essential application logic to enable users to trigger a password reset flow.

**Independent Test**: The usecase handles both existing and unregistered emails, returning success immediately without disclosing whether the email exists. For existing emails, it generates a secure token, saves it to the cache, emits the audit log, and publishes the email send Kafka event.

**Acceptance Scenarios**:

1. **Given** a password reset request with a registered email, **When** processed by `RequestPasswordResetUseCase`, **Then** a secure token is generated and cached, a `password.reset_requested` audit log is emitted, an `email.send` Kafka event is published, and success is returned.
2. **Given** a password reset request with an unregistered email, **When** processed by `RequestPasswordResetUseCase`, **Then** the usecase returns success immediately without storing a token or publishing any events.
3. **Given** a database error occurs during user lookup, **When** processed, **Then** the usecase propagates the database error.

### User Story 4 - Confirm Password Reset UseCase (Priority: P1)

As a user who has requested a password reset, I want to confirm my password reset by providing the token and my new password, so that my password is securely updated and I can log in.

**Why this priority**: Core workflow enabling actual password updates and account recovery.

**Independent Test**: The usecase validates password complexity rules, atomically consumes the reset token from the cache, hashes the password, persists it, invalidates all active sessions for the user, and logs the audit event.

**Acceptance Scenarios**:

1. **Given** a valid token and a password meeting complexity constraints, **When** confirmed, **Then** the new password is hashed (bcrypt cost 12) and saved, all existing sessions for the user are deleted, the cache token is atomically marked as used, a `password.reset_completed` audit log is emitted, and success is returned.
2. **Given** a weak password (not meeting complexity: min 10 chars, 1 upper, 1 number, 1 special), **When** confirmed, **Then** the usecase returns `ErrPasswordWeak`.
3. **Given** a token that is missing or expired in the cache, **When** confirmed, **Then** the usecase returns `ErrTokenExpired`.
4. **Given** a token that has already been consumed/used, **When** confirmed, **Then** the usecase returns `ErrTokenUsed`.

### User Story 5 - Password Reset HTTP Handlers (Priority: P1)

As a client application, I want to submit password reset requests and confirmations via HTTP endpoints, so that the backend can invoke the corresponding business logic and return appropriate HTTP status codes and payloads.

**Why this priority**: Core HTTP interface for self-service password reset.

**Independent Test**: The Echo handlers bind incoming payloads, validate request structures, execute the use cases, and return mapped HTTP status codes matching the API contracts.

**Acceptance Scenarios**:

1. **Given** a valid password reset request (email), **When** posted to `/v1/auth/password-reset/request`, **Then** the handler triggers `RequestPasswordResetUseCase` and returns HTTP 200.
2. **Given** a valid password reset confirmation (token and new password), **When** posted to `/v1/auth/password-reset/confirm`, **Then** the handler triggers `ConfirmPasswordResetUseCase` and returns HTTP 200.
3. **Given** a confirmation request with an expired or missing token, **When** posted, **Then** the handler returns HTTP 400 with a structured error `TOKEN_EXPIRED`.
4. **Given** a confirmation request with an already consumed token, **When** posted, **Then** the handler returns HTTP 400 with a structured error `TOKEN_USED`.
5. **Given** a confirmation request with a weak password, **When** posted, **Then** the handler returns HTTP 422 with a structured error `PASSWORD_WEAK`.
6. **Given** request binding or validation fails on either endpoint, **When** posted, **Then** the handler returns HTTP 400 with a validation error description.


## Edge Cases

- **Token Replay**: A token must be single-use. Once verified, it should be deleted/marked as used. `ErrTokenUsed` is returned if the token has already been consumed.
- **Expiration**: A token must expire after 60 minutes. `ErrTokenExpired` is returned if the token is verified after expiration.
- **Weak Password**: `ErrPasswordWeak` is returned if the password does not meet complexity requirements (min 10 characters, 1 uppercase, 1 number, 1 special character).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Define the `PasswordResetCache` interface with methods: `StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error`, `ConsumeToken(ctx context.Context, token string) (string, error)`, and `DeleteToken(ctx context.Context, token string) error`.
- **FR-002**: Define the specific domain errors:
  - `ErrTokenExpired` = `errors.New("reset token has expired")`
  - `ErrTokenUsed` = `errors.New("reset token has already been used")`
  - `ErrPasswordWeak` = `errors.New("password does not meet complexity requirements")`
- **FR-003**: Define the event payload structs in `internal/domain/events/auth_events.go`:
  - `PasswordResetRequestedEvent` containing details of the password reset request (`Email`, `Token`, `IPAddress`, `UserAgent`, `OccurredAt`).
  - `PasswordResetCompletedEvent` containing details of the password reset completion (`Email`, `IPAddress`, `UserAgent`, `OccurredAt`).
  - `EmailSendEvent` is already defined in `auth_events.go`. If needed, ensure its payload struct is documented or reused.
- **FR-004**: Define `PasswordResetRequest` HTTP DTO with `Email` validate tag `"required,email"`.
- **FR-005**: Define `PasswordResetConfirmRequest` HTTP DTO with `Token` (validate `"required"`), `Password` (validate `"required"`), and `PasswordConfirmation` (validate `"required,eqfield=Password"`).
- **FR-006**: Update `api/openapi.yaml` to document `POST /v1/auth/password-reset/request` returning 200 and standard error responses.
- **FR-007**: Update `api/openapi.yaml` to document `POST /v1/auth/password-reset/confirm` returning 200, 400 (`TOKEN_EXPIRED`, `TOKEN_USED`), and 422 (`PASSWORD_WEAK`) error responses.
- **FR-008**: Implement `RedisPasswordResetCache` in package `cache` implementing the `PasswordResetCache` interface.
- **FR-009**: The key format stored in Redis MUST be `auth:reset_token:{token}` and it MUST enforce a strict 60-minute TTL (or the duration passed by the caller).
- **FR-010**: All logs from the Redis cache implementation must redact the token portion of the key (e.g., logging `auth:reset_token:[REDACTED]`) to prevent leaking raw tokens.
- **FR-011**: The `EmailKafkaProducer` MUST implement a method `PublishPasswordResetEmail` to serialize and publish `EmailSendEvent` to topic `email.send.v1`.
- **FR-012**: Implement `RequestPasswordResetUseCase` accepting `RequestPasswordResetInput` containing the email.
- **FR-013**: Use `AdminUserRepository` to look up the user by email.
- **FR-014**: If the user is not found, return success (nil error) immediately to prevent email enumeration.
- **FR-015**: If the user is found, generate a cryptographically secure 32-byte (256-bit) single-use token and encode it as hex.
- **FR-016**: Store the token in `PasswordResetCache` associated with the admin user's ID for a strict 60-minute TTL.
- **FR-017**: Emit the `password.reset_requested` audit log using the Audit Log interface.
- **FR-018**: Publish the `email.send` Kafka event via `PasswordResetNotifier` using the `events.EmailSendEvent` payload with template `password_reset_request`. The `events.EmailSendEvent.TemplateData` map MUST explicitly include `email` and `token` keys.
- **FR-019**: Implement `ConfirmPasswordResetUseCase` accepting `ConfirmPasswordResetInput` containing the token, new password, IPAddress, and UserAgent.
- **FR-020**: Validate that the password meets complexity constraints (minimum 10 characters, at least 1 uppercase letter, 1 number, and 1 special character) or return `ErrPasswordWeak`.
- **FR-021**: Retrieve the admin user's ID associated with the token by atomically consuming it from `PasswordResetCache` (marking it as used). If the token is missing or expired, return `ErrTokenExpired`. If the token is already consumed, return `ErrTokenUsed`.
- **FR-022**: Hash the new password with bcrypt at cost 12 and update it via `AdminUserRepository.UpdatePassword`.
- **FR-023**: Invalidate/delete the token from the cache on publication failure via `PasswordResetCache.DeleteToken`.
- **FR-024**: Delete all active sessions for the user in `session_tokens` via `SessionTokenRepository.DeleteAllByAdminID` to force re-authentication.
- **FR-025**: Emit the `password.reset_completed` audit log using `AuditLogger` with the `events.PasswordResetCompletedEvent` payload.
- **FR-026**: Implement `RequestPasswordReset` HTTP handler mapped to route `POST /v1/auth/password-reset/request`.
- **FR-027**: Implement `ConfirmPasswordReset` HTTP handler mapped to route `POST /v1/auth/password-reset/confirm`.
- **FR-028**: Bind incoming requests to DTOs `PasswordResetRequest` and `PasswordResetConfirmRequest` and validate them at the HTTP boundary.
- **FR-029**: Map use-case errors to HTTP responses:
  - Map `ErrTokenExpired` to HTTP 400 with error code `TOKEN_EXPIRED`.
  - Map `ErrTokenUsed` to HTTP 400 with error code `TOKEN_USED`.
  - Map `ErrPasswordWeak` to HTTP 422 with error code `PASSWORD_WEAK`.
- **FR-030**: Register the routes and handlers with the Echo framework and wire via Uber Fx.


### Key Entities

- **PasswordResetCache**: Interface for managing temporary single-use reset tokens.
- **PasswordResetRequestedEvent**: Emitted when a password reset is requested.
- **PasswordResetCompletedEvent**: Emitted when a password reset is successfully completed.
- **EmailSendEvent**: Existing struct representing a request to send an email.
- **PasswordResetRequest**: DTO representing the request payload.
- **PasswordResetConfirmRequest**: DTO representing the confirm payload.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% test coverage for new error definitions, event structs, and cache interface behaviors.
- **SC-002**: Domain interfaces and error types compile without importing any external infrastructure or framework dependencies (e.g. Echo, GORM, Redis, Sarama).
- **SC-003**: Event structs serialize correctly to and from JSON.
- **SC-004**: 100% test coverage for validation rules of `PasswordResetRequest` and `PasswordResetConfirmRequest`.
- **SC-005**: OpenAPI contract `api/openapi.yaml` validates successfully.
- **SC-006**: Redis cache store, retrieve, and delete operations complete successfully.
- **SC-007**: 100% test coverage for `RedisPasswordResetCache` using `miniredis`.
- **SC-008**: 100% test coverage for `PublishPasswordResetEmail` using Sarama mocks.
- **SC-009**: 100% unit test coverage for `RequestPasswordResetUseCase` asserting that registered emails generate tokens and dispatch events, unregistered emails return success without side-effects, and database errors are propagated.
- **SC-010**: 100% unit test coverage for `ConfirmPasswordResetUseCase` asserting password strength rules, token checks, session cleanup, database persistence, and audit logging.
- **SC-011**: 100% unit test coverage for `RequestPasswordReset` and `ConfirmPasswordReset` HTTP handlers, asserting correct HTTP status mappings, validation errors, and JSON response payloads.
- **SC-012**: Request password reset and confirm password reset endpoints are successfully registered with the Echo router and wired via Fx bootstrap.

## Assumptions

- The cache TTL for the token will be 60 minutes.
- Password complexity requirements are verified at the domain/usecase layer before invoking the password reset.

