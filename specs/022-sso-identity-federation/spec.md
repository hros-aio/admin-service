# Feature Specification: SSO Identity Federation

**Feature Branch**: `022-sso-identity-federation`

**Created**: 2026-06-27

**Status**: Approved (TSK-SSO-001)

**Input**: User description: "Define the `SSOStateCache` interface required by the application layer to temporarily hold OAuth/OIDC state and nonce parameters to prevent CSRF. Define specific domain errors `ErrNoAccountLinked` and `ErrInvalidSSOState`. Define the event payload structs for the `login.sso_success` and `login.sso_failed` audit events."

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

### Edge Cases

- **State Expiry / Tampering**: The `SSOStateCache` contract must allow storing state parameters with a short TTL to prevent replay attacks and CSRF.
- **Serialization Safety**: Audit events must define struct tags that permit safe, clean JSON serialization for logging or streaming to Kafka.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define `SSOStateCache` interface with methods to store and verify state parameters (with automatic expiration/TTL support).
- **FR-002**: System MUST define domain error `ErrNoAccountLinked` representing "No admin account linked to this identity".
- **FR-003**: System MUST define domain error `ErrInvalidSSOState` representing "Invalid or expired SSO state parameter".
- **FR-004**: System MUST define event payload struct for `login.sso_success` (SSOSuccessEvent) with appropriate audit metadata.
- **FR-005**: System MUST define event payload struct for `login.sso_failed` (SSOFailedEvent) with audit metadata and failure reason.

### Key Entities *(include if feature involves data)*

- **SSOStateCache**: Interface representing temporary in-memory or Redis-backed storage for OAuth2/OIDC state and nonce strings.
  - State (string): Cryptographically secure random identifier.
  - Value/Metadata: Nonce or transient authentication parameters.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Go code in the domain and application interfaces layers compiles without syntax or import errors.
- **SC-002**: Standard package boundaries are respected: no framework or infrastructure imports (e.g., GORM, Echo, Redis) are present in the domain files.
- **SC-003**: Unit tests achieve 100% code coverage for the newly added domain errors and events.

## Assumptions

- Redis will be the eventual provider for `SSOStateCache` in the infrastructure layer, but the interface definition must remain implementation-agnostic.
- Audit event payloads will be serialized to JSON and published to Kafka, requiring appropriate JSON tags.
