# Feature Specification: SSO Identity Federation

**Feature Branch**: `022-sso-identity-federation`

**Created**: 2026-06-27

**Status**: Approved (TSK-SSO-007)

**Input**: User description: "Define the `SSOStateCache` interface required by the application layer to temporarily hold OAuth/OIDC state and nonce parameters to prevent CSRF. Define specific domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState`. Define the event payload structs for the `login.sso_success` and `login.sso_failed` audit events. Create up/down SQL migration scripts to add SSO mapping fields to the `admin_users` table. Add an `sso_identity_id` (VARCHAR, UNIQUE) and `sso_provider` (VARCHAR) column to reliably map IdP assertions to admin accounts beyond just email matching. Define the HTTP request and response DTOs for the SSO endpoints (e.g., `SSOCallbackRequest` containing `code` and `state`). Update the OpenAPI contract `api/openapi.yaml` to document the `GET /auth/sso/initiate` and `GET /auth/sso/callback` endpoints. Implement the `SSOStateCache` interface using Redis. Implement `StoreState(ctx, state, nonce)` mapping the state string with a short TTL (e.g., 10 minutes). Implement `VerifyAndConsumeState(ctx, state)` to fetch the nonce and delete the key atomically, preventing CSRF replay attacks. Update `AdminUserRepository` with a `FindByEmailOrSSO(ctx, email, ssoID)` method. This enables the use case to look up an internal admin account using either the exact `sso_identity_id` returned by the IdP or matching their work email. Implement `InitiateSSOUseCase`. Workflow: Generate a secure random `state` and `nonce`. Store them in Redis via `SSOStateCache.StoreState()`. Construct and return the fully formatted authorization redirect URL for the configured Identity Provider (SAML/OIDC). Implement `CallbackSSOUseCase`. Workflow: Accept `state` and IdP `code`/assertion. Verify state via `SSOStateCache.VerifyAndConsumeState()`; return `ErrInvalidSSOState` if invalid. Exchange code for IdP profile data. Call `AdminUserRepository.FindByEmailOrSSO()`. If no match, emit `login.sso_failed` to audit log and return `ErrNoAccountLinked`. If successful, issue JWT access/refresh tokens, generate a `SessionToken`, save it to the DB, and emit `login.sso_success` to the audit log."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - SSO State and Event Definition (Priority: P1)

Define core domain contracts, errors, and events for federated identity login (SAML 2.0 / OIDC).

**Why this priority**: Core domain model definition must precede use cases and handlers to ensure strict domain isolation in clean architecture.

**Independent Test**: Verification via unit tests that the interface compiles, errors are properly defined, and event payloads serialise as expected.

**Acceptance Scenarios**:

1. **Given** an application layer use case, **When** validating or saving OAuth/OIDC state and nonce, **Then** the use case depends only on the `SSOStateCache` interface.
2. **Given** a failed SSO login, **When** no admin account is linked to the returned identity, **Then** the domain error `ErrNoAccountLinked` is returned.
3. **Given** a failed SSO callback validation, **When** the state or nonce parameter is invalid/expired, **Then** the domain error `ErrInvalidSSOState` is returned.
4. **Given** a successful SSO login, **When** auditing, **Then** the `login.sso_success` event payload contains the operator email and timestamp.
5. **Given** a failed SSO login, **When** auditing, **Then** the `login.sso_failed` event payload contains the provider name, failure reason, and error details.

---

### User Story 2 - Database Migration for SSO Fields (Priority: P1)

Create SQL migrations to add mapping fields to `admin_users` to facilitate SSO identity matching.

**Why this priority**: Database structure must support mapping federated identities before any login logic or mapping query can be implemented.

**Independent Test**: Run migrations up and down against a test PostgreSQL instance and check database schema metadata.

**Acceptance Scenarios**:

1. **Given** the database, **When** migration 000005 UP is run, **Then** columns `sso_identity_id` and `sso_provider` exist on the `admin_users` table, and `sso_identity_id` has a unique constraint.
2. **Given** the migrated database, **When** migration 000005 DOWN is run, **Then** the columns `sso_identity_id` and `sso_provider` are removed from the `admin_users` table.

---

### User Story 3 - SSO REST API contracts and DTO definitions (Priority: P1)

Document endpoints and declare request/response DTO structures for initiating SSO redirects and handling IdP callback responses.

**Why this priority**: The HTTP contract and interface DTOs define the API boundary between frontend clients, backend routing, and application use cases.

**Independent Test**: Unit tests on DTO bindings and structure validation tags.

**Acceptance Scenarios**:

1. **Given** a request to initiate SSO, **When** calling `GET /auth/sso/initiate`, **Then** the server responds with `302 Found` redirect to the IdP.
2. **Given** a redirection callback from the IdP, **When** parsing query parameters, **Then** the parameters map to `SSOCallbackRequest` containing `code` and `state`.
3. **Given** an IdP callback where the identity is not associated with any active admin user, **When** validated, **Then** return `401 Unauthorized` with error code `NO_ACCOUNT_LINKED`.

---

### User Story 4 - Redis State Caching for SSO (Priority: P1)

Implement the `SSOStateCache` using Redis to store transient state parameters during the authorization flow.

**Why this priority**: Storing state parameters with a short TTL prevents CSRF and replay attacks during the authentication flow.

**Independent Test**: Unit tests against a mocked Redis client verifying state storing, retrieving, and atomic deletion.

**Acceptance Scenarios**:

1. **Given** an active Redis client, **When** storing state and nonce with `StoreState`, **Then** the value is successfully cached in Redis with a 10-minute TTL.
2. **Given** a cached state, **When** verifying with `GetState`, **Then** the nonce is returned.
3. **Given** a cached state, **When** state is deleted with `DeleteState` or consumed, **Then** it is removed from Redis so that subsequent gets return `ErrInvalidSSOState`.

---

### User Story 5 - Database Lookup for Federated Identity (Priority: P1)

Implement a secure lookup capability in the admin user repository to fetch an admin account using either the exact SSO identity ID (scoped by provider) or their work email.

**Why this priority**: Correctly looking up user records is essential for verifying authorization assertions during the SSO callback.

**Independent Test**: Unit tests against a mocked database verifying query logic and mapping.

**Acceptance Scenarios**:

1. **Given** a federated identity ID and email, **When** the account is looked up, **Then** retrieve the matching admin user if either the email or SSO identity (scoped by provider) matches.
2. **Given** no matching record exists, **When** looked up, **Then** return a not found error.
3. **Given** the email and SSO identity point to different users, **When** looked up, **Then** return a conflict error.

---

### User Story 6 - SSO Initiation (Priority: P1)

Implement the SSO initiation use case to securely generate state variables, register them in the state cache, and build the redirect authorization URL.

**Why this priority**: Business logic must govern the secure generation of CSRF parameters and authorization redirect parameters to protect the system boundary.

**Independent Test**: Unit tests using mocked state cache and provider configurations verifying correct redirect URL generation.

**Acceptance Scenarios**:

1. **Given** a request to initiate SSO for a supported provider, **When** execution runs, **Then** generate a random state/nonce, cache them, and construct the correct redirect URL containing these parameters.
2. **Given** an unsupported provider name, **When** execution runs, **Then** return an error.

---

### User Story 7 - SSO Callback handling (Priority: P1)

Implement the SSO callback handling use case to consume the state, verify Identity Provider assertions, lookup the matching user, and issue authorization tokens.

**Why this priority**: Callback use case orchestrates the transition from federated assertion to local application session creation, ensuring security check enforcement.

**Independent Test**: Unit tests using mocked domain entities, repositories, and caches verifying state validation, lookup, token generation, session saving, and audit logs.

**Acceptance Scenarios**:

1. **Given** a redirect callback, **When** processed with a valid state, **Then** verify and consume the state, exchange the code for user profile, lookup local user, generate session and JWT tokens, and emit success audit event.
2. **Given** a redirect callback with invalid or expired state, **When** processed, **Then** fail with `ErrInvalidSSOState`.
3. **Given** a valid callback but the user has no linked admin user account, **When** processed, **Then** emit failure audit event and return `ErrNoAccountLinked`.

---

### User Story 8 - SSO HTTP Handlers (Priority: P1)

Implement Echo HTTP handlers for initiating and handling redirects from the Identity Provider.

**Why this priority**: Handlers are the external entry point for clients, exposing the SSO endpoints over the REST API.

**Independent Test**: Unit tests asserting redirect status codes, response headers, HTTP cookies, and HTTP error response mappings.

**Acceptance Scenarios**:

1. **Given** a valid initiate request to `GET /auth/sso/initiate`, **When** parsed, **Then** invoke the use case and redirect to the Identity Provider with HTTP 302.
2. **Given** a callback request to `GET /auth/sso/callback`, **When** the code exchange and user mapping succeed and the `Accept` header prefers HTML, **Then** set the HTTP-only refresh cookie and redirect to the dashboard.
3. **Given** a callback request to `GET /auth/sso/callback`, **When** the code exchange and user mapping succeed and the `Accept` header does not prefer HTML, **Then** set the HTTP-only refresh cookie and return a JSON response containing the access token.
4. **Given** a callback request to `GET /auth/sso/callback`, **When** no account is linked to the identity, **Then** respond with HTTP 401 and the message "No admin account linked to this identity".

---

### Edge Cases

- **State Expiry / Tampering**: The `SSOStateCache` contract must allow storing state parameters with a short TTL to prevent replay attacks and CSRF.
- **Serialization Safety**: Audit events must define struct tags that permit safe, clean JSON serialization for logging or streaming to Kafka.
- **Migration Idempotency**: Migration scripts must handle columns that already exist/do not exist gracefully to avoid failing if run repeatedly.
- **DTO Validation**: The callback request MUST validate that both `code` and `state` parameters are present and non-empty.
- **Atomic Deletion / Replay Protection**: To prevent state reuse, state parameters must be deleted from Redis immediately upon validation.
- **SSO Identity and Email Discrepancies**: If a user is registered with a different SSO ID than email, the lookup MUST still resolve the user correctly, returning a conflict error if they point to different accounts.
- **Provider Misconfiguration**: If client configuration parameters are missing or invalid, the initiation request must gracefully return a clear, structured configuration error instead of generating malformed redirect URLs.
- **Code Exchange Failure**: If the IdP code exchange fails or returned profile assertions are malformed, callback handling must abort cleanly and log the failure.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define `SSOStateCache` interface with methods to store and verify state parameters (with automatic expiration/TTL support).
- **FR-002**: System MUST define domain error `ErrNoAccountLinked` representing "No admin account linked to this identity".
- **FR-003**: System MUST define domain error `ErrInvalidSSOState` representing "Invalid or expired SSO state parameter".
- **FR-004**: System MUST define event payload struct for `login.sso_success` (SSOSuccessEvent) with appropriate audit metadata.
- **FR-005**: System MUST define event payload struct for `login.sso_failed` (SSOFailedEvent) with audit metadata and failure reason.
- **FR-006**: Up migration MUST add `sso_identity_id` (VARCHAR, UNIQUE) and `sso_provider` (VARCHAR) to `admin_users` table.
- **FR-007**: Down migration MUST drop `sso_identity_id` and `sso_provider` from `admin_users` table.
- **FR-008**: System MUST document `GET /auth/sso/initiate` and `GET /auth/sso/callback` in `api/openapi.yaml`.
- **FR-009**: System MUST define `SSOCallbackRequest` DTO containing `code` and `state` query parameters with validation tags.
- **FR-010**: System MUST implement `SSOStateCache` interface using Redis (`internal/infrastructure/cache/sso_state_redis.go`).
- **FR-011**: `StoreState` MUST map the state string to the nonce with a 10-minute TTL.
- **FR-012**: Retrieval MUST verify and support deletion of the state parameter.
- **FR-013**: System MUST update the admin user repository to support lookup by either email or SSO provider and SSO identity ID.
- **FR-014**: The repository lookup MUST participate correctly in the active transaction context.
- **FR-015**: System MUST implement `InitiateSSOUseCase` to generate state/nonce, cache them, and construct the redirect URL.
- **FR-016**: State and nonce parameters MUST be generated using a cryptographically secure random source.
- **FR-017**: The constructed redirect URL MUST include client ID, redirect URI, response type, scope, state, and nonce parameters.
- **FR-018**: System MUST implement `CallbackSSOUseCase` to verify state/code, fetch provider profile, perform a provider-scoped admin-user lookup referencing the provider identifier and identity ID, create session, and issue JWT tokens.
- **FR-019**: State verification MUST atomically consume the state parameter.
- **FR-020**: Successful SSO authentication MUST publish the `login.sso_success` audit event to Kafka.
- **FR-021**: Unlinked identity assertions MUST publish the `login.sso_failed` audit event and return `ErrNoAccountLinked`.
- **FR-022**: System MUST implement Echo HTTP handlers for `GET /auth/sso/initiate` and `GET /auth/sso/callback`.
- **FR-023**: The initiate handler MUST invoke `InitiateSSOUseCase` and execute an HTTP 302 redirect.
- **FR-024**: The callback handler MUST invoke `CallbackSSOUseCase`, set the refresh session token in an HTTP-only cookie, and either redirect to the frontend dashboard (for browser requests requesting HTML) or return a JSON response containing the access token.
- **FR-025**: The callback handler MUST map `ErrNoAccountLinked` to HTTP 401 with the message "No admin account linked to this identity".

### Key Entities *(include if feature involves data)*

- **SSOStateCache**: Interface representing temporary in-memory or Redis-backed storage for OAuth2/OIDC state and nonce strings.
  - State (string): Cryptographically secure random identifier.
  - Value/Metadata: Nonce or transient authentication parameters.
- **AdminUser**: Database model representing admin account. Added fields:
  - `sso_identity_id` (string): Maps federated ID.
  - `sso_provider` (string): Records IdP name.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Go code in the domain and application interfaces layers compiles without syntax or import errors.
- **SC-002**: Standard package boundaries are respected: no framework or infrastructure imports (e.g., GORM, Echo, Redis) are present in the domain files.
- **SC-003**: Unit tests achieve 100% code coverage for the newly added domain errors and events.
- **SC-004**: Migration SQL executes successfully forward and backward without errors.
- **SC-005**: OpenAPI specification `api/openapi.yaml` compiles and validates successfully with Swagger/OpenAPI tools.
- **SC-006**: Unit tests verify validation tags on the `SSOCallbackRequest` DTO structure.
- **SC-007**: Unit tests verify Redis state caching, TTL, and deletion workflow with a mocked Redis client.
- **SC-008**: Unit tests verify that lookup by email or SSO resolves the correct record or returns conflict/not-found results appropriately.
- **SC-009**: Unit tests verify correct generation of state/nonce, caching, and redirect URL construction for configured providers.
- **SC-010**: Unit tests verify `CallbackSSOUseCase` workflow under successful, unlinked, conflict, and invalid-state execution flows.
- **SC-011**: Echo handlers for SSO endpoints are successfully registered in the Fx dependency injection module.
- **SC-012**: Unit tests in `internal/adapter/http/auth_sso_handler_test.go` achieve at least 80% coverage and assert correct HTTP status codes, redirect locations, cookies, and error response mappings.

## Assumptions

- Redis will be the eventual provider for `SSOStateCache` in the infrastructure layer, but the interface definition must remain implementation-agnostic.
- Audit event payloads will be serialized to JSON and published to Kafka, requiring appropriate JSON tags.
