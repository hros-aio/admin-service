# Feature Specification: Authentication Service

**Feature Branch**: `002-authentication-service`

**Created**: 2026-06-16

**Status**: Draft

**Input**: User description: "Implement authentication for the HROS Admin Portal. This includes migrations for admin_users and session_tokens tables, password hashing with bcrypt, JWT-based session management, and RBAC foundation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Login (Priority: P1)

As an HROS Admin, I want to securely log in to the portal using my email and password so that I can access administrative features.

**Why this priority**: High. Blocking requirement for all other features (Tenants, Plans, etc.).

**Independent Test**: Can be tested by sending a POST request to `/v1/auth/login` with valid and invalid credentials.

**Acceptance Scenarios**:

1. **Given** valid admin credentials, **When** I log in, **Then** I receive a 200 OK status with an access token and refresh token.
2. **Given** invalid credentials, **When** I log in, **Then** I receive a 401 Unauthorized status with a standard error response.
3. **Given** a deactivated account, **When** I log in, **Then** I receive a 403 Forbidden status.

---

### User Story 2 - Token Refresh (Priority: P1)

As a logged-in admin, I want my session to remain active without re-entering credentials so that my workflow is not interrupted.

**Why this priority**: High. Essential for user experience and security (short-lived access tokens).

**Independent Test**: Can be tested by sending a POST request to `/v1/auth/refresh` with a valid refresh token.

**Acceptance Scenarios**:

1. **Given** a valid refresh token, **When** I request a refresh, **Then** I receive a new access token.
2. **Given** a revoked refresh token, **When** I request a refresh, **Then** I receive a 401 Unauthorized status.

---

### User Story 3 - Logout (Priority: P2)

As an admin, I want to be able to log out of the portal so that my session is terminated securely.

**Why this priority**: Medium. Important for security on shared devices.

**Independent Test**: Can be tested by sending a POST request to `/v1/auth/logout`.

**Acceptance Scenarios**:

1. **Given** an active session, **When** I log out, **Then** my refresh token is revoked in the database and I can no longer use it.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-001**: System MUST store admin users with hashed passwords (bcrypt).
- **FR-AUTH-002**: System MUST support Role-Based Access Control (RBAC) via `roles` and `role_permissions` tables.
- **FR-AUTH-003**: System MUST provide JWT-based authentication.
- **FR-AUTH-004**: System MUST persist refresh tokens in the `session_tokens` table for session management and revocation.
- **FR-AUTH-005**: System MUST validate that the admin user domain matches authorized HROS domains (per PRD).
- **FR-AUTH-006**: System MUST lock accounts after 5 failed attempts (per PRD).

### Key Entities *(include if feature involves data)*

- **AdminUser**: Represents an HROS administrator.
- **Role**: Represents a set of permissions.
- **Permission**: Defines access to specific modules (Tenant, Plan, etc.).
- **SessionToken**: Represents an active or historical refresh token.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-001**: Password hashing uses bcrypt with a minimum cost of 10.
- **SC-AUTH-002**: Access tokens expire in 15 minutes; Refresh tokens expire in 24 hours (or 30 days if persistent).
- **SC-AUTH-003**: Login latency is under 500ms (excluding network).
- **SC-AUTH-004**: Migrations execute without errors in forward and backward directions.

## Assumptions

- **Domain Restriction**: Initially, we will use a configurable list of allowed email domains.
- **MFA**: MFA is defined in PRD but will be implemented in a subsequent phase (out of scope for this epic).
- **SSO**: SSO is out of scope for this initial implementation.
