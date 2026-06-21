# Feature Specification: Auth Refresh DTOs

**Feature Branch**: `011-auth-refresh-dtos`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "Use the existing repository documents as source of truth. TSK-AUTH-011: **Layer**: DTO **Description**: Update the OpenAPI contract `api/openapi.yaml` and HTTP DTOs. Add the `remember_me` (boolean) field to `LoginRequest`. Create `RefreshRequest` (containing `refresh_token`) and update response schemas if necessary. Add strict validation struct tags. **Input**: API Specification (`POST /auth/refresh` and `POST /auth/login`). **Output**: `internal/adapter/http/dto/auth_dto.go`, `api/openapi.yaml` **Definition of Done**: DTOs accurately map the new fields, and the `openapi.yaml` passes validation documenting the 200 and 401 responses. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - OpenAPI Definition for Token Refresh (Priority: P1)

As an integration developer, I want a documented contract for the token refresh endpoint (`POST /v1/auth/refresh`), so that I can understand the expected request/response models.

**Why this priority**: High. Essential for exposing the token refresh feature to client apps in an API-first way.

**Independent Test**: The contract can be verified by running OpenAPI validation on `api/openapi.yaml` to ensure `/v1/auth/refresh` matches specifications.

**Acceptance Scenarios**:

1. **Given** the OpenAPI specification, **When** reviewing the `/v1/auth/refresh` endpoint, **Then** I see it accepts `RefreshRequest` (containing `refresh_token` as a required string) and returns `LoginResponse` (containing `access_token` and `refresh_token`) with a `200 OK` status.
2. **Given** the OpenAPI specification, **When** reviewing the `/v1/auth/refresh` endpoint, **Then** I see the error responses are documented with `400 Bad Request`, `401 Unauthorized`, and `500 Internal Server Error` using the standard `ErrorResponse` schema reference.

---

### User Story 2 - Login Request remember_me Mappings (Priority: P1)

As a client application, I want the `POST /v1/auth/login` endpoint's DTO to include a boolean `remember_me` flag, so that I can control whether the refresh token has an extended expiration.

**Why this priority**: High. Crucial for enabling session duration customization.

**Independent Test**: Asserted via unit tests binding a login request containing the `remember_me` flag.

**Acceptance Scenarios**:

1. **Given** a login payload, **When** providing a boolean value for `remember_me`, **Then** the request is successfully bound and validated.
2. **Given** a login payload, **When** `remember_me` is omitted, **Then** the validation succeeds and defaults to false.

---

### User Story 3 - Strict Request Validation DTO (Priority: P1)

As the HTTP adapter layer, I want request DTOs to have strict validation rules (via Go struct tags), so that invalid requests (like missing required fields) are blocked at the boundary.

**Why this priority**: High. Prevents malformed inputs from reaching the application layer, aligning with the project's coding conventions.

**Independent Test**: Can be tested via unit tests in `internal/adapter/http/auth/dto/auth_dto_test.go` checking struct validation output on valid and invalid payloads.

**Acceptance Scenarios**:

1. **Given** a `RefreshRequest` payload, **When** `refresh_token` is empty or missing, **Then** validation fails.
2. **Given** a `LoginRequest` payload, **When** `email` is invalid or missing, **Then** validation fails.

### Edge Cases

- **Missing Required Fields**: If `refresh_token` is missing or empty, validation fails immediately at the HTTP layer and returns 400 Bad Request.
- **Incorrect Types**: If `remember_me` is provided as a non-boolean (e.g., string `"true"` or number `1`), binding or validation fails.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-011-001**: The system MUST define `RefreshRequest` DTO containing `refresh_token` string.
- **FR-011-002**: The `LoginRequest` DTO MUST contain `remember_me` boolean field.
- **FR-011-003**: The `LoginRequest` and `RefreshRequest` fields MUST use strict validation struct tags (e.g. `validate:"required"` for required fields, `validate:"required,email"` for emails).
- **FR-011-004**: The system MUST update the OpenAPI spec `api/openapi.yaml` to add `/v1/auth/refresh` path.
- **FR-011-005**: The system MUST document standard 200 OK and 401 Unauthorized responses for `/v1/auth/refresh` using components from `api/openapi.yaml`.
- **FR-011-006**: The system MUST ensure the OpenAPI spec `api/openapi.yaml` remains syntactically valid.

### Key Entities *(include if feature involves data)*

- **LoginRequest**: Captures user credentials and the optional `remember_me` flag.
- **RefreshRequest**: Captures the `refresh_token` required to rotate a session.
- **LoginResponse**: Exposes the returned JWT and refresh token.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-011-001**: `api/openapi.yaml` is updated and passes validation.
- **SC-011-002**: Unit tests in `internal/adapter/http/auth/dto/auth_dto_test.go` assert that `LoginRequest` and `RefreshRequest` validation behaves exactly as specified (including validation tag failures for empty fields).
- **SC-011-003**: Compilation succeeds with all HTTP adapter DTOs properly mapped.

## Assumptions

- Standard Echo binder and `go-playground/validator/v10` library are used for bindings and validation.
- Standard error structures will be mapped using the existing `ErrorResponse` schema.
