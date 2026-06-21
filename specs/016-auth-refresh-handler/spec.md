# Feature Specification: Auth Refresh Handler

**Feature Branch**: `016-auth-refresh-handler`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "TSK-AUTH-016: Layer: Handler Description: Implement the Echo HTTP handler for POST /auth/refresh. Bind the request body to RefreshRequest, trigger RefreshSessionUseCase, map ErrInvalidRefreshToken to HTTP 401, and serialize the successful response. Also update the POST /auth/login handler to pass remember_me into the use case. Input: RefreshSessionUseCase, updated LoginUseCase, Echo Context. Output: internal/adapter/http/auth_handler.go, internal/adapter/http/auth_handler_test.go Definition of Done: The /auth/refresh endpoint is successfully wired via Fx and returns HTTP 200 on success or HTTP 401 on failure. Unit tests assert correct JSON mapping and status codes. Do not create a new epic if this belongs to an existing feature. Do not expand scope beyond the provided task."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Auth Session Token Rotation (Priority: P1) 🎯 MVP

As a client application holding a near-expired or valid session refresh token, I want to call the refresh endpoint to obtain a new access/refresh token pair, so that my user's active session is securely extended without prompt.

**Why this priority**: High. Core adapter handler mapping that exposes token rotation functionality to frontend applications.

**Independent Test**: Verified via Echo handler unit tests checking the endpoint route `/v1/auth/refresh`, binding valid and invalid refresh token payloads, calling the use case, asserting the use case calls match, and verifying the HTTP response body and status code.

**Acceptance Scenarios**:

1. **Given** a valid refresh token payload `{"refresh_token": "valid-token"}`, **When** the HTTP request `POST /v1/auth/refresh` is received, **Then** the handler triggers `RefreshSessionUseCase` and returns HTTP 200 with the new `access_token` and `refresh_token` in the response body.
2. **Given** an invalid or expired refresh token payload, **When** the HTTP request `POST /v1/auth/refresh` is received, **Then** the handler maps the use case errors (`ErrInvalidRefreshToken` or `ErrTokenExpired`) to HTTP 401 Unauthorized.
3. **Given** an empty refresh token string in the request payload, **When** the HTTP request is received, **Then** structural validation fails and the handler returns HTTP 400 Bad Request.

---

### User Story 2 - Pass Remember Me Selection on Login (Priority: P1) 🎯 MVP

As an administrator logging in, I want my choice of "remember me" to be passed from the HTTP request down to the business logic, so that my session expiration calculates correctly based on my preference.

**Why this priority**: High. Ensures end-to-end alignment of the `RememberMe` login preference.

**Independent Test**: Verified via Echo handler unit tests asserting that the `RememberMe` boolean field from `LoginRequest` is successfully mapped and passed to `LoginUseCase`.

**Acceptance Scenarios**:

1. **Given** a login payload with `"remember_me": true`, **When** the login endpoint executes, **Then** the handler passes `RememberMe = true` in the `LoginInput` parameters of the `LoginUseCase` execution.

---

### Edge Cases

- **Malformed JSON Body**: If the request body contains invalid JSON, the handler must return HTTP 400 Bad Request.
- **UseCase Internal Failures**: If the `RefreshSessionUseCase` fails with unexpected database errors, the handler must degrade safely, returning HTTP 500 Internal Server Error without leaking internal context.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-AUTH-016-001**: The system MUST register the Echo HTTP route `POST /v1/auth/refresh` pointing to `AuthHandler.Refresh`.
- **FR-AUTH-016-002**: The `AuthHandler.Refresh` method MUST bind the request body to `dto.RefreshRequest`.
- **FR-AUTH-016-003**: The handler MUST validate the bound struct using the validator framework, returning HTTP 400 on failure.
- **FR-AUTH-016-004**: The handler MUST invoke `RefreshSessionUseCase.Execute` with the bound `RefreshToken`.
- **FR-AUTH-016-005**: If the use case returns `ErrInvalidRefreshToken` or `ErrTokenExpired`, the handler MUST return HTTP 401 Unauthorized using the standard `sharedErrors.ErrorResponse` envelope.
- **FR-AUTH-016-006**: The system MUST return HTTP 403 Forbidden for specific known domain errors `ErrUserInactive` and `ErrUserLocked`. All other unexpected or unknown errors (system failures) MUST return HTTP 500 Internal Server Error.
- **FR-AUTH-016-007**: Upon successful usecase completion, the handler MUST respond with HTTP 200 containing the `access_token` and `refresh_token` mapped to `dto.LoginResponse`.
- **FR-AUTH-016-008**: The `AuthHandler.Login` method MUST map the `RememberMe` boolean from `LoginRequest` into the `LoginInput` struct passed to `LoginUseCase.Execute`.

### Key Entities

- **RefreshRequest**: The DTO payload capturing the refresh token.
- **LoginResponse**: The response payload returning token details.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-AUTH-016-001**: 100% test coverage of the updated `AuthHandler` endpoints under `auth_handler_test.go`.
- **SC-AUTH-016-002**: Verification that `POST /v1/auth/refresh` responds with 200 for successful rotation and 401 for invalid refresh tokens.
- **SC-AUTH-016-003**: Handlers contain only request binding, validation, usecase execution, and error mapping logic (no business decisions).

## Assumptions

- **Injecting RefreshSessionUseCase**: `RefreshSessionUseCase` is already registered in the application dependencies and can be successfully injected into `AuthHandler` via its constructor.
