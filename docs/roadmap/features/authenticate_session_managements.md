Here is the implementation breakdown for the **Authentication & Session Management** epic, structured into independent.

### Feature Name: Core Credential Authentication (Login & Logout)
**Purpose:** Establish the fundamental identity verification flow using email/password, issue RS256 JWT access tokens, and allow explicit session invalidation.
**API Impact:**
*   Implement `POST /auth/login` (Base implementation without MFA or lockout).
*   Implement `DELETE /auth/session`.
**Database Impact:** 
*   Read-only access to `admin_users` to verify email, status, and bcrypt password hash.
*   Insert/Delete records in `session_tokens` to track active device sessions.
**Events:** 
*   Append to `audit_logs` service: `login.success`, `login.failed`, `logout.success`.
**Dependencies:** 
*   Global Audit Middleware (must be ready to record the login events).
**Acceptance Criteria:**
*   Successful login returns a 15-minute RS256 access token and HTTP 200.
*   Invalid credentials or unregistered emails return HTTP 401 with generic "Invalid email or password" and take the same processing time (preventing timing oracles).
*   Calling the logout endpoint successfully invalidates the specific session token.

### Feature Name: Session Refresh & Persistence
**Purpose:** Support continuous admin workflows via secure JWT refresh rotation and optional 30-day persistent sessions.
**API Impact:**
*   Implement `POST /auth/refresh`.
*   Update `POST /auth/login` to parse the `remember_me` boolean.
**Database Impact:**
*   Update/Rotate refresh token values inside the `session_tokens` table.
**Events:**
*   Append to `audit_logs`: `session.refreshed`.
**Dependencies:**
*   Core Credential Authentication feature.
**Cache Requirements:**
*   Implement a Redis blacklist for immediately revoked access tokens before their 15-minute expiry (e.g., post-logout).
**Acceptance Criteria:**
*   Submitting a valid refresh token issues a new access/refresh token pair.
*   If `remember_me` is true, the refresh token expires in 30 days; otherwise, it expires at the end of the browser session.

### Feature Name: Brute-Force Lockout Defense
**Purpose:** Prevent password guessing attacks by temporarily locking admin accounts after consecutive failed attempts.
**API Impact:**
*   Update `POST /auth/login` to check lockout state before evaluating credentials and record failure increments.
**Database Impact:**
*   No strict schema changes, but relies on `admin_users` status check.
**Events:**
*   Append to `audit_logs`: `account.locked`.
*   Publish `email.send` Kafka event to notify the admin of the lockout and unlock timestamp.
**Dependencies:**
*   Core Credential Authentication feature.
**Cache Requirements:**
*   Redis required to track failed attempts per email. Key format: `auth:failed_attempts:{email}` with a 15-minute TTL. 
*   Redis required to track lockout state. Key format: `auth:lockout:{email}` with a 30-minute TTL.
**Acceptance Criteria:**
*   5 consecutive failed login attempts within 15 minutes lock the account for 30 minutes.
*   Attempts during the lockout return HTTP 401 with an `ACCOUNT_LOCKED` code.
*   A successful login before 5 attempts clears the failure counter in Redis.

### Feature Name: MFA Enforcement (Super Admins)
**Purpose:** Require a secondary authentication factor (TOTP/WebAuthn) for Super Admin logins to satisfy zero-trust security requirements.
**API Impact:**
*   Update `POST /auth/login` to return a short-lived `mfa_token` (HTTP 200) instead of a JWT pair if the user role is Super Admin.
*   Implement `POST /auth/mfa/verify`.
**Database Impact:**
*   Schema changes to `admin_users` to store `totp_secret` or WebAuthn credentials.
**Events:**
*   Append to `audit_logs`: `mfa.failed`, `mfa.success`.
**Dependencies:**
*   Core Credential Authentication feature.
**Cache Requirements:**
*   Redis cache to store the short-lived `mfa_token` mapping to the partially authenticated user context (TTL: 5 minutes).
**Acceptance Criteria:**
*   Super Admins do not receive JWTs until the MFA challenge is successfully verified.
*   Invalid MFA codes return HTTP 401 `MFA_INVALID`.

### Feature Name: Self-Service Password Reset
**Purpose:** Allow admins to securely recover their accounts without relying on internal IT support.
**API Impact:**
*   Implement `POST /auth/password-reset/request`.
*   Implement `POST /auth/password-reset/confirm`.
**Database Impact:**
*   Update `admin_users.password_hash`.
*   Delete all existing rows for the user in `session_tokens` upon successful reset to force re-authentication everywhere.
**Events:**
*   Append to `audit_logs`: `password.reset_requested`, `password.reset_completed`.
*   Publish `email.send` Kafka event containing the secure single-use reset link.
**Dependencies:**
*   Kafka Outbox/Producer implementation for email events.
**Cache Requirements:**
*   Redis cache to store the single-use reset token. Key format: `auth:reset_token:{token}` mapping to the admin ID with a 60-minute TTL.
**Acceptance Criteria:**
*   Reset requests return HTTP 200 immediately to prevent email enumeration.
*   The new password must meet complexity constraints (min 10 chars, 1 upper, 1 number, 1 special) or return HTTP 422 `PASSWORD_WEAK`.
*   Reusing an expired or consumed token returns HTTP 400 `TOKEN_EXPIRED` or `TOKEN_USED`.

### Feature Name: Admin Account Activation (Accept Invite)
**Purpose:** Enable newly invited administrators to set their initial password and activate their accounts.
**API Impact:**
*   Implement `POST /auth/accept-invite`.
**Database Impact:**
*   Verify and consume the token from the `invite_tokens` table.
*   Update `admin_users.password_hash` and set `status` to `active`.
**Events:**
*   Append to `audit_logs`: `admin.activated`, `invite.accepted`.
*   Publish internal in-app notification event to the inviter.
**Dependencies:**
*   Admin Management epic (specifically the `POST /admins/invite` feature that generates the token).
**Acceptance Criteria:**
*   Providing a valid invite token and matching password constraints successfully activates the account and returns HTTP 200.
*   Using an invite link older than 48 hours returns HTTP 400 `INVITE_EXPIRED`.

### Feature Name: SSO Identity Federation
**Purpose:** Map enterprise identity assertions (SAML 2.0 / OIDC) to existing admin accounts.
**API Impact:**
*   Implement SSO callback and redirect endpoints (e.g., `GET /auth/sso/callback`).
**Database Impact:**
*   Lookup `admin_users` by email or SSO Identity ID.
*   Insert into `session_tokens` upon success.
**Events:**
*   Append to `audit_logs`: `login.sso_success`, `login.sso_failed`.
**Dependencies:**
*   IdP Configuration parameters (Client ID/Secret, SAML ACS URLs).
**Cache Requirements:**
*   Redis cache to store OAuth/OIDC state/nonce parameters to prevent CSRF during the federation flow.
**Acceptance Criteria:**
*   Valid assertions from the IdP matching an existing admin email seamlessly issue a JWT pair.
*   If the IdP assertion has no matching HROS admin account, return an error: "No admin account linked to this identity".

### Feature Name: Biometric Device Login (WebAuthn)
**Purpose:** Support fast, secure login using trusted device authenticators like FaceID or TouchID.
**API Impact:**
*   Implement WebAuthn challenge/registration endpoints (e.g., `POST /auth/biometric/challenge`, `POST /auth/biometric/verify`).
**Database Impact:**
*   Schema changes to `admin_users` to support `webauthn_credentials` (storing public key, credential ID, and sign count).
**Events:**
*   Append to `audit_logs`: `login.biometric_success`.
**Dependencies:**
*   None.
**Cache Requirements:**
*   Redis cache required to hold the cryptographic WebAuthn challenge payload between the request and verification steps.
**Acceptance Criteria:**
*   Successful WebAuthn verification issues a JWT token pair, bypassing the need for a password.