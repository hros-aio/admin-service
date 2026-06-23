# Feature Specification: MFA Enforcement (Super Admins)

**Feature Branch**: `019-mfa-enforcement`

**Created**: 2026-06-23

**Status**: Draft

**Input**: User description: "Feature: MFA Enforcement (Super Admins). Task: TSK-MFA-002. Layer: Domain. Description: Update the AdminUser domain entity to include TotpSecret and WebauthnCredentials. Define the MFACache interface required by the application layer to temporarily hold the authenticated context. Define the specific domain errors ErrMFAInvalid and ErrMFATokenExpired. Define the event payload structs for mfa.success and mfa.failed."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add MFA credential storage columns to admin_users table (Priority: P1)

As a database administrator or developer, I want to add `totp_secret` (VARCHAR) and `webauthn_credentials` (JSONB) columns to the `admin_users` table so we can support both TOTP and WebAuthn MFA standards for Super Admin users.

**Why this priority**: Essential schema setup required before implementing the MFA logic.

**Independent Test**: Migrations execute forward and backward successfully without losing existing user records.

**Acceptance Scenarios**:

1. **Given** the database has existing users in the `admin_users` table, **When** the up migration `000003_add_mfa_to_admin_users.up.sql` is run, **Then** `totp_secret` and `webauthn_credentials` columns are added to `admin_users`, and existing user data is intact.
2. **When** the down migration `000003_add_mfa_to_admin_users.down.sql` is run, **Then** those columns are removed, and existing user data remains intact.

---

### User Story 2 - Define Domain Primitives for MFA Verification (Priority: P1)

As a developer, I want domain structures for MFA credentials inside the `AdminUser` entity, a cache interface `MFACache` to hold user context during MFA checks, domain errors `ErrMFAInvalid` / `ErrMFATokenExpired`, and event payloads for `mfa.success` / `mfa.failed`.

**Why this priority**: Core domain definition required before any application layer verification logic can be constructed.

**Independent Test**: The structs, interfaces, and errors compile without external dependencies, and event payload structures serialize correctly to JSON.

**Acceptance Scenarios**:

1. **Given** an `AdminUser` struct, **When** initialized with MFA credentials, **Then** `TotpSecret` and `WebauthnCredentials` are retrievable as Go primitives.
2. **Given** a short-lived MFA verification flow, **When** caching the authentication context, **Then** `MFACache` can store, retrieve, and delete the context using a unique token.
3. **Given** validation or expiration checks fail, **When** returning errors, **Then** `ErrMFAInvalid` and `ErrMFATokenExpired` are returned.

---

### Edge Cases

- **Transaction isolation**: The migration script must run in a single transaction block so that failures in middle execution revert the table to its previous state.
- **Nullability**: Since existing users do not have MFA secrets set, the new columns must allow `NULL` values.
- **Cache serialization**: The cached admin user context must serialize to JSON and deserialize without data loss.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The database migration MUST add a `totp_secret` column of type `VARCHAR` to the `admin_users` table.
- **FR-002**: The database migration MUST add a `webauthn_credentials` column of type `JSONB` to the `admin_users` table.
- **FR-003**: The migrations MUST be non-destructive, preserving existing records in the `admin_users` table.
- **FR-004**: The migration down script MUST revert the changes by dropping the `totp_secret` and `webauthn_credentials` columns from the `admin_users` table.
- **FR-005**: `AdminUser` domain entity MUST include `TotpSecret` (string) and `WebauthnCredentials` ([]byte) fields.
- **FR-006**: Define the `MFACache` interface with `Store`, `Get`, and `Delete` methods.
- **FR-007**: Define `ErrMFAInvalid` and `ErrMFATokenExpired` domain errors.
- **FR-008**: Define `MFASuccessEvent` and `MFAFailedEvent` event payload structs.

### Key Entities *(include if feature involves data)*

- **AdminUser**: Represents administrators in the system.
  - `TotpSecret` (string): TOTP secret key.
  - `WebauthnCredentials` ([]byte): WebAuthn credential records serialized as JSON.
- **MFACache**: Contract for temporary storage of partially authenticated context.
- **MFASuccessEvent**: Emitted on successful MFA verification.
- **MFAFailedEvent**: Emitted on failed MFA verification.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Migrations run in less than 5 seconds against a standard PostgreSQL instance.
- **SC-002**: Up migration successfully creates columns `totp_secret` and `webauthn_credentials` in table `admin_users`.
- **SC-003**: Down migration successfully drops columns `totp_secret` and `webauthn_credentials` from table `admin_users`.
- **SC-004**: Zero loss of data in existing `admin_users` records when applying both up and down migrations.
- **SC-005**: 100% test coverage for new errors, event structs, and entity methods.

## Assumptions

- PostgreSQL 15+ is used.
- The new columns are optional initially, so they must be nullable.
- Migrations will be executed using the application's migration tool or raw SQL execution tool.
