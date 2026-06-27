# Feature Specification: SSO Identity Federation

**Feature Branch**: `022-sso-identity-federation`

**Created**: 2026-06-27

**Status**: Approved (TSK-SSO-005)

**Input**: User description: "Define the `SSOStateCache` interface required by the application layer to temporarily hold OAuth/OIDC state and nonce parameters to prevent CSRF. Define specific domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState`. Define the event payload structs for the `login.sso_success` and `login.sso_failed` audit events. Create up/down SQL migration scripts to add SSO mapping fields to the `admin_users` table. Add an `sso_identity_id` (VARCHAR, UNIQUE) and `sso_provider` (VARCHAR) column to reliably map IdP assertions to admin accounts beyond just email matching. Define the HTTP request and response DTOs for the SSO endpoints (e.g., `SSOCallbackRequest` containing `code` and `state`). Update the OpenAPI contract `api/openapi.yaml` to document the `GET /auth/sso/initiate` and `GET /auth/sso/callback` endpoints. Implement the `SSOStateCache` interface using Redis. Implement `StoreState(ctx, state, nonce)` mapping the state string with a short TTL (e.g., 10 minutes). Implement `VerifyAndConsumeState(ctx, state)` to fetch the nonce and delete the key atomically, preventing CSRF replay attacks. Update `AdminUserRepository` with a `FindByEmailOrSSO(ctx, email, ssoID)` method. This enables the use case to look up an internal admin account using either the exact `sso_identity_id` returned by the IdP or matching their work email."

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

Implement database lookup in the `AdminUserRepository` to fetch an admin account using either the exact SSO identity ID or their work email.

**Why this priority**: Correctly looking up user records is essential for verifying authorization assertions during the SSO callback.

**Independent Test**: Unit tests against a mocked GORM connection verifying the query execution structure and GORM mapping.

**Acceptance Scenarios**:

1. **Given** a federated identity ID and email, **When** queried using `FindByEmailOrSSO`, **Then** retrieve the matching `AdminUser` if either the `email` or `sso_identity_id` matches.
2. **Given** no matching record in the database, **When** queried using `FindByEmailOrSSO`, **Then** return `ErrUserNotFound` error.

---

### Edge Cases

- **State Expiry / Tampering**: The `SSOStateCache` contract must allow storing state parameters with a short TTL to prevent replay attacks and CSRF.
- **Serialization Safety**: Audit events must define struct tags that permit safe, clean JSON serialization for logging or streaming to Kafka.
- **Migration Idempotency**: Migration scripts must handle columns that already exist/do not exist gracefully to avoid failing if run repeatedly.
- **DTO Validation**: The callback request MUST validate that both `code` and `state` parameters are present and non-empty.
- **Atomic Deletion / Replay Protection**: To prevent state reuse, state parameters must be deleted from Redis immediately upon validation.
- **SSO Identity and Email Discrepancies**: If a user is registered with a different SSO ID than email, the repository MUST still resolve the user row correctly using the `OR` condition.

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
- **FR-013**: System MUST update `AdminUserRepository` interface with `FindByEmailOrSSO(ctx, email, ssoID)` method.
- **FR-014**: Gorm implementation MUST query using `email = ? OR sso_identity_id = ?` under transaction context if active.

### Key Entities *(include if feature involves data)*

- **SSOStateCache**: Interface representing temporary in-memory or Redis-backed storage for OAuth2/OIDC state and nonce strings.
  - State (string): Cryptographically secure random identifier.
  - Value/Metadata: Nonce or transient authentication parameters.
- **AdminUser**: Database model representing admin account. Added fields:
  - `sso_identity_id` (string, unique): Maps federated ID.
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
- **SC-008**: Unit tests verify `FindByEmailOrSSO` GORM query statement and mapping with `sqlmock`.

## Assumptions

- Redis will be the eventual provider for `SSOStateCache` in the infrastructure layer, but the interface definition must remain implementation-agnostic.
- Audit event payloads will be serialized to JSON and published to Kafka, requiring appropriate JSON tags.
