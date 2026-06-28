# Feature Specification: Biometric Device Login (WebAuthn)

**Feature Branch**: `023-biometric-device-login`

**Created**: 2026-06-28

**Status**: Draft

**Input**: User description: "Biometric Device Login (WebAuthn)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Biometric Registration (Priority: P1)

Admins can register their biometric authenticator (TouchID, FaceID, or platform authenticators) so they can use it for subsequent logins.

**Why this priority**: Core step to enable biometric login. Users must register before they can authenticate.

**Independent Test**: Can be tested by initiating the registration flow, passing the challenge to the client WebAuthn API, and submitting the result to complete registration.

**Acceptance Scenarios**:

1. **Given** a logged-in admin, **When** they request a registration challenge, **Then** the system generates a cryptographically secure challenge, caches it, and returns the public key options.
2. **Given** a cached registration challenge, **When** the admin submits the signature from their authenticator, **Then** the system verifies the signature and saves the credential metadata.

---

### User Story 2 - Biometric Login (Priority: P1)

Admins can log in to the administrative portal using their registered biometric device instead of their password.

**Why this priority**: Key user journey that delivers the primary value of the feature (passwordless biometric login).

**Independent Test**: Can be tested by requesting a login challenge for an email, scanning biometrics, and verifying the signature to obtain an active session.

**Acceptance Scenarios**:

1. **Given** an admin email, **When** they request an authentication challenge, **Then** the system generates a challenge, caches it, and returns it to the client.
2. **Given** a cached authentication challenge, **When** the admin submits a valid signature matching their registered biometric credential, **Then** the system authenticates the user, logs a biometric success event, and updates the stored biometric credential sign count.

---

### User Story 3 - Biometric API Payload & Validation (Priority: P2)

Clients and automated integrators must be able to use standardized, fully-validated request and response payloads when interacting with the biometric login endpoints.

**Why this priority**: Ensures API contract robustness, validation of client parameters, and clear error responses for integration correctness.

**Independent Test**: Test with valid and invalid requests to challenge and verification endpoints to confirm strict validator constraints and OpenAPI spec conformance.

**Acceptance Scenarios**:

1. **Given** an invalid email in a challenge request, **When** submitted to the challenge endpoint, **Then** the system rejects it with a 400 Bad Request response.
2. **Given** missing cryptographic fields (e.g. credential ID or signature) in a verification request, **When** submitted to the verify endpoint, **Then** the system rejects it with a 400 Bad Request response.

---

### Edge Cases

- **Unregistered Biometric Device**: The admin attempts to log in using biometrics but has no registered biometric credentials.
- **Invalid Cryptographic Signature**: The signature submitted during verification does not match the cached challenge or stored public key.
- **Challenge Expired or Missing**: The cryptographic challenge has expired or is not found in transient storage, rejecting the verification step.
- **Cloned Authenticator Detection**: The sign count returned by the authenticator is less than or equal to the stored sign count, indicating potential device cloning.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST generate a cryptographically secure challenge for WebAuthn authentication.
- **FR-002**: The system MUST cache the challenge in a transient store with a short TTL (e.g., 5 minutes).
- **FR-003**: The system MUST support storing multiple biometric credentials per admin user.
- **FR-004**: The system MUST verify biometric signatures during the login phase using the cached challenge and registered public keys.
- **FR-005**: The system MUST log an audit event for successful biometric login containing the admin ID, email, and credential ID.
- **FR-006**: The system MUST fail the login attempt with a clear error indicating the device is not registered if a biometric login is requested for an account with no biometric credentials.
- **FR-007**: The system MUST fail the login attempt with a clear signature verification error if the cryptographic signature check fails.
- **FR-008**: The system MUST define request schemas validating client-supplied parameters for the biometric challenge endpoint (requiring a valid email).
- **FR-009**: The system MUST define request schemas validating client-supplied parameters for the biometric verification endpoint (requiring email, credential ID, authenticator data, client data JSON, and cryptographic signature).
- **FR-010**: The OpenAPI contract MUST document the biometric challenge and verification endpoints, including successful outcomes and failure codes (such as 400 Bad Request and 401 Unauthorized).
- **FR-011**: The system MUST update the sign count of the verified biometric credential in persistent storage (e.g., inside the user's `webauthn_credentials` JSONB field) post-verification to mitigate authenticator cloning attacks.

### Key Entities *(include if feature involves data)*

- **Transient Challenge Storage**: A temporary storage mechanism holding challenge payloads associated with an admin, subject to a short-lived expiration TTL during the multi-step authentication handshake.
- **Biometric Credential**: Information representing a registered authenticator containing Credential ID, Public Key, and Sign Count.
- **Biometric Challenge Payload**: Structured parameters containing the target email.
- **Biometric Verification Payload**: Structured parameters containing the target email, credential ID, signature, and client/authenticator context.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Admins can log in using biometric verification in under 3 seconds.
- **SC-002**: Cryptographic challenges are automatically cleared from cache after 5 minutes.
- **SC-003**: Unauthorized login attempts using modified challenges or signatures are rejected 100% of the time.
- **SC-004**: All malformed or incomplete client request payloads are rejected by API validation gates.
- **SC-005**: 100% of successful WebAuthn authentications result in the persistent update of the credential's sign count.

## Assumptions

- Admins have devices supporting modern WebAuthn/FIDO2 authenticators.
- A secure HTTPS channel is established (WebAuthn requires HTTPS).
- The client browser supports the WebAuthn API.
