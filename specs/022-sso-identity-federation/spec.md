# Feature Specification: SSO Identity Federation

**Feature Branch**: `022-sso-identity-federation`

**Created**: 2026-06-27

**Status**: Approved (TSK-SSO-002)

**Input**: User description: "Define the `SSOStateCache` interface required by the application layer to temporarily hold OAuth/OIDC state and nonce parameters to prevent CSRF. Define specific domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState`. Define the event payload structs for the `login.sso_success` and `login.sso_failed` audit events. Create up/down SQL migration scripts to add SSO mapping fields to the `admin_users` table. Add an `sso_identity_id` (VARCHAR, UNIQUE) and `sso_provider` (VARCHAR) column to reliably map IdP assertions to admin accounts beyond just email matching."

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

### Edge Cases

- **State Expiry / Tampering**: The `SSOStateCache` contract must allow storing state parameters with a short TTL to prevent replay attacks and CSRF.
- **Serialization Safety**: Audit events must define struct tags that permit safe, clean JSON serialization for logging or streaming to Kafka.
- **Migration Idempotency**: Migration scripts must handle columns that already exist/do not exist gracefully to avoid failing if run repeatedly.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define `SSOStateCache` interface with methods to store and verify state parameters (with automatic expiration/TTL support).
- **FR-002**: System MUST define domain error `ErrNoAccountLinked` representing "No admin account linked to this identity".
- **FR-003**: System MUST define domain error `ErrInvalidSSOState` representing "Invalid or expired SSO state parameter".
- **FR-004**: System MUST define event payload struct for `login.sso_success` (SSOSuccessEvent) with appropriate audit metadata.
- **FR-005**: System MUST define event payload struct for `login.sso_failed` (SSOFailedEvent) with audit metadata and failure reason.
- **FR-006**: Up migration MUST add `sso_identity_id` (VARCHAR, UNIQUE) and `sso_provider` (VARCHAR) to `admin_users` table.
- **FR-007**: Down migration MUST drop `sso_identity_id` and `sso_provider` from `admin_users` table.

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

## Assumptions

- Redis will be the eventual provider for `SSOStateCache` in the infrastructure layer, but the interface definition must remain implementation-agnostic.
- Audit event payloads will be serialized to JSON and published to Kafka, requiring appropriate JSON tags.
