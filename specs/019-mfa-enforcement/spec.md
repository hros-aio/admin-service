# Feature Specification: MFA Enforcement (Super Admins)

**Feature Branch**: `019-mfa-enforcement`

**Created**: 2026-06-23

**Status**: Draft

**Input**: User description: "Feature: MFA Enforcement (Super Admins). Task: TSK-MFA-004. Layer: Cache. Description: Implement the MFACache interface using Redis. Implement StoreToken(ctx, mfaToken, adminID) mapping the short-lived token to the user ID with a strict 5-minute TTL. Implement GetAdminID(ctx, mfaToken) and DeleteToken(ctx, mfaToken)."

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

### User Story 3 - Define API Contracts and HTTP DTOs for MFA Verification (Priority: P1)

As a front-end developer or API consumer, I want the login response to include MFA requirements when MFA is enforced, and I want a dedicated verification endpoint documented in the OpenAPI spec and defined in HTTP DTOs, so that we can implement the MFA user experience correctly.

**Why this priority**: Prerequisite for implementing the HTTP handlers and API endpoints.

**Independent Test**: The DTO structs compile with strict validation tags, and `api/openapi.yaml` passes validation checks.

**Acceptance Scenarios**:

1. **Given** an admin user with MFA enabled logs in, **When** the login endpoint is called, **Then** the response contains `mfa_required: true`, `mfa_token`, and allowed `mfa_methods`.
2. **Given** a client has an `mfa_token`, **When** verifying MFA via `POST /v1/auth/mfa/verify`, **Then** they submit `mfa_token`, `method`, and `code` (if TOTP).
3. **Given** an invalid verification request, **When** parsed by the handler, **Then** structural validation constraints fail with bad request.

---

### User Story 4 - Temporarily Cache MFA Session Context in Redis (Priority: P1)

As a developer, I want the short-lived MFA session token to map to the user's Admin ID in a secure Redis cache with a strict 5-minute TTL, so that we can verify the user context during the MFA validation step.

**Why this priority**: Essential middleware infrastructure to bridge the password login step and the MFA verification endpoint.

**Independent Test**: Token storage, retrieval, and deletion operations succeed using `miniredis` and respect the 5-minute TTL.

**Acceptance Scenarios**:

1. **Given** a successful password login, **When** generating the short-lived MFA token, **Then** `MFACache` stores the mapping of token to Admin ID with a 5-minute expiration time.
2. **Given** an MFA token, **When** verifying MFA, **Then** `MFACache` retrieves the Admin ID.
3. **Given** an MFA token is verified, **When** verification finishes, **Then** `MFACache` deletes the token immediately.
4. **Given** 5 minutes have elapsed since storing the token, **When** trying to retrieve the Admin ID, **Then** `MFACache` returns an error indicating the token has expired.

---

### Edge Cases

- **Transaction isolation**: The migration script must run in a single transaction block so that failures in middle execution revert the table to its previous state.
- **Nullability**: Since existing users do not have MFA secrets set, the new columns must allow `NULL` values.
- **Validation combinations**: The DTO for verification must require `code` if the verification method is `totp`.
- **Redis downtime**: Cache operations must handle Redis client connectivity failures gracefully, returning standard errors.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The database migration MUST add a `totp_secret` column of type `VARCHAR` to the `admin_users` table.
- **FR-002**: The database migration MUST add a `webauthn_credentials` column of type `JSONB` to the `admin_users` table.
- **FR-003**: The migrations MUST be non-destructive, preserving existing records in the `admin_users` table.
- **FR-004**: The migration down script MUST revert the changes by dropping the `totp_secret` and `webauthn_credentials` columns from the `admin_users` table.
- **FR-005**: `AdminUser` domain entity MUST include `TotpSecret` (string) and `WebauthnCredentials` ([]byte) fields.
- **FR-006**: Define the `MFACache` interface with `StoreToken`, `GetAdminID`, and `DeleteToken` methods.
- **FR-007**: Define `ErrMFAInvalid` and `ErrMFATokenExpired` domain errors.
- **FR-008**: Define `MFASuccessEvent` and `MFAFailedEvent` event payload structs.
- **FR-009**: The OpenAPI contract `api/openapi.yaml` MUST define `POST /v1/auth/mfa/verify`.
- **FR-010**: `LoginResponse` schema MUST include optional `mfa_required`, `mfa_token`, and `mfa_methods` properties.
- **FR-011**: The DTO struct `LoginResponse` MUST support optional `MFARequired`, `MFAToken`, and `MFAMethods` fields.
- **FR-012**: Define `MFAVerifyRequest` DTO containing `mfa_token`, `method`, and `code` with validation tags (`required` for token and method, `required_if` for code when method is `totp`).
- **FR-013**: Define the `MFACache` interface with methods: `StoreToken(ctx, mfaToken, adminID)`, `GetAdminID(ctx, mfaToken)`, and `DeleteToken(ctx, mfaToken)`. The `StoreToken` method MUST NOT accept a TTL parameter from callers.
- **FR-014**: Implement `RedisMFACache` mapping the token to the Admin ID key format `auth:mfa_token:{mfaToken}`. It MUST hard-code a 5-minute TTL when storing tokens.
- **FR-015**: The cached keys MUST have a strict 5-minute TTL.

### Key Entities *(include if feature involves data)*

- **AdminUser**: Represents administrators in the system.
  - `TotpSecret` (string): TOTP secret key.
  - `WebauthnCredentials` ([]byte): WebAuthn credential records serialized as JSON.
- **MFACache**: Contract for temporary storage of partially authenticated context.
- **MFASuccessEvent**: Emitted on successful MFA verification.
- **MFAFailedEvent**: Emitted on failed MFA verification.
- **MFAVerifyRequest**: DTO representing the client verification request payload.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Migrations run in less than 5 seconds against a standard PostgreSQL instance.
- **SC-002**: Up migration successfully creates columns `totp_secret` and `webauthn_credentials` in table `admin_users`.
- **SC-003**: Down migration successfully drops columns `totp_secret` and `webauthn_credentials` from table `admin_users`.
- **SC-004**: Zero loss of data in existing `admin_users` records when applying both up and down migrations.
- **SC-005**: 100% test coverage for new errors, event structs, and entity methods.
- **SC-006**: OpenAPI contract `api/openapi.yaml` compiles and passes validation (e.g., using swagger/openapi CLI tool if available).
- **SC-007**: HTTP validation tests verify DTO validator tags correctly block invalid request payloads.
- **SC-008**: Cache store and retrieve operations complete in under 5ms (P95).
- **SC-009**: Expiration test asserts key is missing exactly after 5 minutes (or appropriate mocked duration).

## Assumptions

- PostgreSQL 15+ is used.
- The new columns are optional initially, so they must be nullable.
- Migrations will be executed using the application's migration tool or raw SQL execution tool.
- The go-playground/validator package is used for structural DTO validation in HTTP handlers.
- Redis client is available and properly initialized.
