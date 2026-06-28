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
2. **Given** a cached authentication challenge, **When** the admin submits a valid signature matching their registered biometric credential, **Then** the system authenticates the user and logs a biometric success event.

---

### Edge Cases

- **ErrBiometricNotRegistered**: The admin attempts to log in using biometrics but has no registered biometric credentials.
- **ErrInvalidBiometricSignature**: The signature submitted during verification does not match the cached challenge or stored public key.
- **Challenge Expired/Not Found**: The challenge cache has cleared (due to TTL) or the request has an invalid challenge, rejecting the verification step.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST generate a cryptographically secure challenge for WebAuthn authentication.
- **FR-002**: The system MUST cache the challenge using `WebAuthnChallengeCache` with a short TTL (e.g., 60 seconds).
- **FR-003**: The system MUST support storing multiple biometric credentials per admin user.
- **FR-004**: The system MUST verify biometric signatures during the login phase using the cached challenge and registered public keys.
- **FR-005**: The system MUST log a `login.biometric_success` audit event containing the admin ID, email, and credential ID upon successful verification.
- **FR-006**: The system MUST return `ErrBiometricNotRegistered` if a biometric login is requested for an account with no biometric credentials.
- **FR-007**: The system MUST return `ErrInvalidBiometricSignature` if the cryptographic verification fails.

### Key Entities *(include if feature involves data)*

- **WebAuthnChallengeCache**: Transient repository holding challenge payload associated with an admin during the multi-step handshake.
- **BiometricCredential**: Domain representation of a registered authenticator containing Credential ID, Public Key, and Sign Count.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Admins can log in using biometric verification in under 3 seconds.
- **SC-002**: Cryptographic challenges are automatically cleared from cache after 60 seconds.
- **SC-003**: Unauthorized login attempts using modified challenges or signatures are rejected 100% of the time.

## Assumptions

- Admins have devices supporting modern WebAuthn/FIDO2 authenticators.
- A secure HTTPS channel is established (WebAuthn requires HTTPS).
- The client browser supports the WebAuthn API.
