# Feature Specification: Admin Auth Handler

**Feature Branch**: `008-admin-auth-handler`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "TSK-AUTH-008: **Layer**: Handler **Description**: Implement Echo HTTP handler `AuthHandler`. Register `POST /auth/login` and `DELETE /auth/session` routes. Bind JSON request to DTOs, execute UseCases, and serialize the domain outputs into standard JSON envelopes. Map domain errors (like `ErrInvalidCredentials`) to HTTP 401. **Input**: `LoginUseCase`, `LogoutUseCase`, Echo Context. **Output**: `internal/adapter/http/auth_handler.go`, `internal/adapter/http/auth_handler_test.go`, `internal/adapter/http/module.go` **Definition of Done**: Routes are wired into Echo via Fx. Handlers contain no business logic. Unit tests using Echo's `httptest` utilities assert correct 200, 204, and 401 HTTP status code responses."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Login Authentication (Priority: P1)

As an HROS Administrator, I want to authenticate with my email and password via the HTTP API, so that I can establish a secure session and obtain access/refresh tokens.

**Why this priority**: Critical MVP requirement. Authenticating users is the entry point to access any protected features in the admin portal.

**Independent Test**: Can be fully tested by sending a POST request to the login endpoint with valid credentials and verifying that tokens are returned with a 200 OK response.

**Acceptance Scenarios**:

1. **Given** valid administrator credentials, **When** a login POST request is received, **Then** the system establishes a session and returns a 200 OK status containing the access and refresh tokens.
2. **Given** invalid administrator credentials, **When** a login POST request is received, **Then** the system returns a 401 Unauthorized status with a standard error response.
3. **Given** a malformed request payload (invalid JSON structure), **When** a login POST request is received, **Then** the system returns a 400 Bad Request status.

---

### User Story 2 - Admin Logout Session Termination (Priority: P1)

As an authenticated HROS Administrator, I want to log out of my active session via the HTTP API, so that my access and refresh tokens are permanently invalidated.

**Why this priority**: Critical security requirement to ensure sessions are properly terminated when requested.

**Independent Test**: Can be fully tested by sending a DELETE request to the session endpoint with a valid active session token, and verifying that the session is terminated with a 204 No Content response.

**Acceptance Scenarios**:

1. **Given** a valid active session token, **When** a session DELETE request is received, **Then** the system deletes/revokes the session token and returns a 204 No Content status.
2. **Given** an invalid or already deleted session token, **When** a session DELETE request is received, **Then** the system handles the revocation idempotently and returns a 204 No Content status.

### Edge Cases

- **Malformed JSON Payload**: Binding failures due to type mismatches, missing required fields, or syntax errors in the JSON body of the login request must be handled and return a 400 Bad Request error.
- **UseCase Execution Failures**: If the underlying `LoginUseCase` or `LogoutUseCase` encounters unexpected issues (e.g. database disconnect), the handler must respond with a 500 Internal Server Error without leaking internal system details.
- **Empty Auth Token on Logout**: If a logout request is received without a session token (or authorization header, depending on extraction mechanism), the handler must respond with a 401 Unauthorized or 400 Bad Request.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST register the route `POST /v1/auth/login` to the `AuthHandler.Login` method.
- **FR-002**: System MUST register the route `DELETE /v1/auth/session` to the `AuthHandler.Logout` method.
- **FR-003**: System MUST bind incoming JSON request payloads to the corresponding request DTOs (e.g., `LoginRequest`).
- **FR-004**: Handlers MUST NOT contain any business logic and MUST delegate all business operations to `LoginUseCase` and `LogoutUseCase`.
- **FR-005**: System MUST serialize successful domain outputs into standard JSON response envelopes.
- **FR-006**: System MUST map domain-specific errors (specifically `ErrInvalidCredentials`) to HTTP 401 Unauthorized, returning a standard JSON error envelope.
- **FR-007**: System MUST wire the `AuthHandler` and its route registrations into Echo using dependency injection (Uber Fx).

### Key Entities *(include if feature involves data)*

- **AuthHandler**: The HTTP adapter handling Echo requests and responses.
- **LoginRequest DTO**: The request schema containing authentication credentials.
- **LoginResponse DTO**: The response schema returning access/refresh tokens.
- **JSON Envelope**: The standard API response wrapper for successful outcomes and error payloads.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of valid login requests result in HTTP 200 OK and return a structured JSON response containing the access token and refresh token.
- **SC-002**: 100% of login requests with incorrect credentials result in HTTP 401 Unauthorized with a structured JSON error payload.
- **SC-003**: 100% of valid session revocation requests return HTTP 204 No Content with no body.
- **SC-004**: Unit tests covering HTTP routes, JSON binding, error mapping, and status codes achieve at least 75% coverage for the handler implementation.

## Assumptions

- **Use Case Availability**: `LoginUseCase` and `LogoutUseCase` interfaces are defined and can be mocked for unit testing.
- **Standard DTOs**: The request/response schemas (DTOs) match those defined in the API contract.
- **Token Extraction**: The session token string can be extracted from the request context or headers/cookies as standard.
