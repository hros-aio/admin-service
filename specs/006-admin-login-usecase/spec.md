# Feature Specification: Admin Login Use Case

**Feature Branch**: `006-admin-login-usecase`

**Created**: 2026-06-18

**Status**: Draft

**Input**: User description: "Implement LoginUseCase. Workflow: Fetch user by email → verify bcrypt password (cost factor 12) → issue RS256 JWT access token (15-min expiry) → generate session token → save to DB → emit login.success or login.failed to audit log interface. Note: Ensure constant-time processing for invalid emails to prevent timing oracle attacks. Input: LoginRequest mapped data, AdminUserRepository, SessionTokenRepository, JWT Signing Key."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Secure Portal Access (Priority: P1)

As an Admin User, I want to securely log in using my corporate credentials so that I can access the super admin management features.

**Why this priority**: Core entry point for all administrative functions. Without authentication, the system is inaccessible.

**Independent Test**: Can be tested by providing valid email/password combination and verifying that a valid JWT is returned and a session record is created in the database.

**Acceptance Scenarios**:

1. **Given** a registered admin user with email "admin@hros.io" and password "SecurePass123!", **When** they submit these credentials to the login service, **Then** a JWT access token is issued with a 15-minute expiration and a new session token record is saved in the database.
2. **Given** a successful login, **When** the system processes the request, **Then** a "login.success" event is emitted to the audit log with the user's ID.

---

### User Story 2 - Failed Login Defense (Priority: P2)

As a security-conscious administrator, I want the system to reject invalid login attempts without leaking information about whether an email exists in the system.

**Why this priority**: Prevents user enumeration and brute-force attacks via timing analysis.

**Independent Test**: Can be tested by measuring response times for both "existing email" and "non-existent email" with incorrect passwords; the variance should be statistically insignificant.

**Acceptance Scenarios**:

1. **Given** an email that does not exist in the database, **When** a login attempt is made, **Then** the system performs a dummy password comparison to ensure constant-time response and returns a generic "Invalid credentials" error.
2. **Given** a failed login attempt, **When** the error occurs, **Then** a "login.failed" event is emitted to the audit log with the attempted email.

---

### User Story 3 - Session Integrity (Priority: P3)

As a system auditor, I want every login to be tracked so that we have a record of who accessed the system and when.

**Why this priority**: Compliance and security forensics.

**Independent Test**: Verify that for every login request (success or fail), there is a corresponding entry in the audit interface.

**Acceptance Scenarios**:

1. **Given** any login attempt, **When** processed by the use case, **Then** the audit log interface is invoked before returning the response to the caller.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST fetch the admin user by email using the `AdminUserRepository`.
- **FR-002**: The system MUST verify the provided password against the stored bcrypt hash using a cost factor of 12.
- **FR-003**: The system MUST implement a "dummy" password comparison when a user is not found to prevent timing oracle attacks.
- **FR-004**: On successful verification, the system MUST generate an RS256-signed JWT with a 15-minute expiration.
- **FR-005**: On successful verification, the system MUST generate a session token (refresh token) and persist it using the `SessionTokenRepository`.
- **FR-006**: The system MUST emit a "login.success" event to the audit log interface on successful login, including the user's ID and timestamp.
- **FR-007**: The system MUST emit a "login.failed" event to the audit log interface on failed login, including the attempted email and timestamp.
- **FR-008**: The system MUST return a domain-specific "Unauthenticated" error on any credential failure, which maps to a 401 Unauthorized status.

### Key Entities *(include if feature involves data)*

- **AdminUser**: Represents the administrative user account, containing email and bcrypt-hashed password.
- **SessionToken**: Represents an active authenticated session, linking a user to a persistent token in the database.
- **LoginResponse**: Data transfer object containing the access token (JWT) and relevant user metadata.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of valid login requests result in a signed RS256 JWT verifiable by its public key.
- **SC-002**: Variance in response time between "valid email/wrong password" and "invalid email/any password" is less than 5% (to prevent timing attacks).
- **SC-003**: 100% of login attempts (success and failure) are recorded in the audit log interface.
- **SC-004**: Access tokens (JWT) expire exactly 15 minutes after issuance.

## Assumptions

- **JWT Signing Key**: The RS256 private key is securely provided via the platform's configuration module.
- **Audit Interface**: A standard audit log interface is available in the application or domain layer for emission of security events.
- **Password Utility**: A utility for bcrypt comparison with cost factor support is available or will be implemented as part of this feature slice.
- **Persistence**: Database schema for users and session tokens is already migrated (as per TSK-AUTH-001).
