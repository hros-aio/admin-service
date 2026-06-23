# Feature Specification: MFA Enforcement (Super Admins)

**Feature Branch**: `019-mfa-enforcement`

**Created**: 2026-06-23

**Status**: Draft

**Input**: User description: "Feature: MFA Enforcement (Super Admins). Task: TSK-MFA-001. Layer: Migration. Description: Create up/down SQL migration scripts to add MFA credential storage to the admin_users table. Add a totp_secret (VARCHAR) column and a webauthn_credentials (JSONB) column to support both RFC 6238 and FIDO2 standards. Input: Database Domain Model specifications. Output: migrations/000003_add_mfa_to_admin_users.up.sql, 000003_add_mfa_to_admin_users.down.sql. Definition of Done: Migrations execute successfully forward and backward against a local PostgreSQL instance without dropping existing user data."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add MFA credential storage columns to admin_users table (Priority: P1)

As a database administrator or developer, I want to add `totp_secret` (VARCHAR) and `webauthn_credentials` (JSONB) columns to the `admin_users` table so we can support both TOTP and WebAuthn MFA standards for Super Admin users.

**Why this priority**: Essential schema setup required before implementing the MFA logic.

**Independent Test**: Migrations execute forward and backward successfully without losing existing user records.

**Acceptance Scenarios**:

1. **Given** the database has existing users in the `admin_users` table, **When** the up migration `000003_add_mfa_to_admin_users.up.sql` is run, **Then** `totp_secret` and `webauthn_credentials` columns are added to `admin_users`, and existing user data is intact.
2. **When** the down migration `000003_add_mfa_to_admin_users.down.sql` is run, **Then** those columns are removed, and existing user data remains intact.

---

### Edge Cases

- **Transaction isolation**: The migration script must run in a single transaction block so that failures in middle execution revert the table to its previous state.
- **Nullability**: Since existing users do not have MFA secrets set, the new columns must allow `NULL` values.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The database migration MUST add a `totp_secret` column of type `VARCHAR` to the `admin_users` table.
- **FR-002**: The database migration MUST add a `webauthn_credentials` column of type `JSONB` to the `admin_users` table.
- **FR-003**: The migrations MUST be non-destructive, preserving existing records in the `admin_users` table.
- **FR-004**: The migration down script MUST revert the changes by dropping the `totp_secret` and `webauthn_credentials` columns from the `admin_users` table.

### Key Entities *(include if feature involves data)*

- **admin_users**: Represents administrators in the system.
  - `totp_secret` (VARCHAR): Encrypted TOTP secret key (null if not set).
  - `webauthn_credentials` (JSONB): Array/structure representing FIDO2/WebAuthn credential records (null or empty array if not set).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Migrations run in less than 5 seconds against a standard PostgreSQL instance.
- **SC-002**: Up migration successfully creates columns `totp_secret` and `webauthn_credentials` in table `admin_users`.
- **SC-003**: Down migration successfully drops columns `totp_secret` and `webauthn_credentials` from table `admin_users`.
- **SC-004**: Zero loss of data in existing `admin_users` records when applying both up and down migrations.

## Assumptions

- PostgreSQL 15+ is used.
- The new columns are optional initially, so they must be nullable.
- Migrations will be executed using the application's migration tool or raw SQL execution tool.
