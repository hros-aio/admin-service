# Feature Specification: Authentication DTOs and OpenAPI

**Feature Branch**: `003-auth-dtos-openapi`

**Created**: 2026-06-16

**Status**: Draft

**Input**: User description: "Use the existing repository documents as source of truth. TSK-AUTH-003: **Layer**: DTO **Description**: Define HTTP request and response DTOs for the authentication endpoints. Includes `LoginRequest` (email, password, remember_me) and `LoginResponse` (access_token, refresh_token). Add struct tags for strict OpenAPI request validation. Update the `api/openapi.yaml` contract. **Input**: API Specification for `POST /auth/login` and `DELETE /auth/session`. **Output**: `internal/adapter/http/dto/auth_dto.go`, `api/openapi.yaml` **Definition of Done**: DTOs properly map required fields; `openapi.yaml` passes validation and correctly documents 200, 204, and 401 error responses. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Login (Priority: P1)

As an HROS Admin, I want to provide my credentials securely so that I can access the management portal.

**Why this priority**: Critical. Entry point for all administrative actions.

**Independent Test**: Can be tested by sending a POST request to `/v1/auth/login` with valid credentials and receiving tokens.

**Acceptance Scenarios**:

1. **Given** a valid email and password, **When** I log in, **Then** I receive a 200 OK status with an access token and a refresh token.
2. **Given** invalid credentials, **When** I log in, **Then** I receive a 401 Unauthorized status with a standard error response.
3. **Given** a malformed email, **When** I log in, **Then** I receive a 400 Bad Request status due to validation failure.

---

### User Story 2 - Terminate Session (Priority: P1)

As an HROS Admin, I want to log out of my current session so that my access is revoked and my account remains secure on shared devices.

**Why this priority**: Critical for security compliance.

**Independent Test**: Can be tested by sending a DELETE request to `/v1/auth/session` with a valid token and verifying the session is terminated.

**Acceptance Scenarios**:

1. **Given** an active session, **When** I log out, **Then** I receive a 204 No Content status and my refresh token is invalidated.

---

### Edge Cases

- **Expired Tokens**: How does the system handle a logout request with an already expired token? (Should still return success or appropriate error).
- **Missing Required Fields**: Login request missing password or email should be caught by strict validation.
- **Malformed JWT**: Logout request with invalid token format should return 401.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define a `LoginRequest` containing `email` (string, required, email format), `password` (string, required), and `remember_me` (boolean).
- **FR-002**: System MUST define a `LoginResponse` containing `access_token` (string) and `refresh_token` (string).
- **FR-003**: System MUST validate `LoginRequest` fields using strict rules before processing.
- **FR-004**: System MUST provide OpenAPI definitions for `POST /v1/auth/login` and `DELETE /v1/auth/session`.
- **FR-005**: System MUST document standard error responses (400, 401, 500) for authentication endpoints in OpenAPI.

### Key Entities *(include if feature involves data)*

- **LoginRequest**: Data Transfer Object for capturing admin credentials and session preference.
- **LoginResponse**: Data Transfer Object for returning session tokens to the client.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `api/openapi.yaml` passes validation against the OpenAPI 3.0/3.1 specification.
- **SC-002**: DTO structs in `internal/adapter/http/dto/auth_dto.go` include all required fields and validation tags.
- **SC-003**: API documentation correctly reflects the 200 OK, 204 No Content, and 401 Unauthorized scenarios.
- **SC-004**: Validation tags in DTOs match the constraints defined in the OpenAPI schema.

## Assumptions

- **Standard Error Format**: Authentication errors will follow the `ErrorResponse` schema defined in the project foundation.
- **Remember Me Logic**: The `remember_me` flag will be used by the application layer to determine the refresh token's expiration (handled in a later task).
- **Endpoint Versioning**: Endpoints will use the `/v1` prefix as per project conventions.
