# Feature Specification: Admin Account Activation (Accept Invite)

**Feature Branch**: `021-admin-account-activation`

**Created**: 2026-06-27

**Status**: Approved

**Input**: User description: "Define the `InviteToken` domain entity and the `InviteTokenRepository` interface required to fetch and consume tokens. Define specific domain errors `ErrInviteExpired`, `ErrInviteUsed`, and `ErrPasswordWeak`. Define the event payload structs for the `admin.activated` and `invite.accepted` audit events, as well as the in-app notification Kafka event (`notification.send`) targeting the inviter."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Account Activation via Invite (Priority: P1)

Newly invited administrators can click their invitation link, specify a password, and activate their account.

**Why this priority**: Crucial for onboarding new administrators securely. Without this, no invited administrator can access the portal.

**Independent Test**: Can be tested by invoking the activation endpoint with a valid token, verifying database state changes (admin active, token used) and event emission.

**Acceptance Scenarios**:

1. **Given** a valid and unused invite token, **When** a user submits a strong password (min 10 chars, 1 uppercase, 1 number, 1 special), **Then** the account status becomes active, the invite token is marked as used, and the password hash is updated.
2. **Given** an expired invite token (older than 48 hours), **When** activation is attempted, **Then** return `ErrInviteExpired` error.
3. **Given** an already used invite token, **When** activation is attempted, **Then** return `ErrInviteUsed` error.
4. **Given** a weak password (does not meet constraints), **When** activation is attempted, **Then** return `ErrPasswordWeak` error.

---

### Edge Cases

- **Token Replay**: Attempting to reuse an invite token multiple times must fail on the second attempt.
- **Expiration Check**: Tokens expired by even 1 second must be rejected.
- **Weak Password Inputs**: Password validation must strictly enforce complexity constraints.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support fetching an invite token by its unique token string.
- **FR-002**: System MUST check if the invite token has expired (creation + 48 hours).
- **FR-003**: System MUST check if the invite token has already been used.
- **FR-004**: System MUST transition the admin user status to active upon successful activation.
- **FR-005**: System MUST record `admin.activated` and `invite.accepted` audit events.
- **FR-006**: System MUST dispatch a `notification.send` Kafka event to notify the inviter that their invite was accepted.

### Key Entities *(include if feature involves data)*

- **InviteToken**: Domain entity representing a secure invitation token linked to an admin user.
  - `ID`: Unique identifier (UUID).
  - `AdminID`: ID of the admin user being invited (UUID).
  - `Token`: Unique cryptographically secure token string.
  - `ExpiresAt`: Time when the token expires (typically created_at + 48 hours).
  - `UsedAt`: Time when the token was consumed (null if unused).
  - `CreatedBy`: ID of the Super Admin who created the invite (UUID).
  - `CreatedAt`: Time when the token was created.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Account activation is processed in under 500ms.
- **SC-002**: 100% of expired or reused tokens are rejected.
- **SC-003**: 100% of password validation failures block account activation.

## Assumptions

- **Constraints**: Password validation rules require min 10 characters, 1 uppercase letter, 1 number, and 1 special character.
- **Mail System**: Email delivery and actual token generation is handled by the Admin Management service.
