# Implementation Plan: Biometric Device Login (WebAuthn) - Handler Layer (TSK-BIO-007)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Implement Echo HTTP handlers for biometric login challenges (`POST /v1/auth/biometric/challenge`) and verification (`POST /v1/auth/biometric/verify`). Bind client request bodies to DTO structs, invoke the respective UseCases (`GenerateBiometricChallengeUseCase` and `VerifyBiometricUseCase`), map any business domain errors appropriately (such as 401 Unauthorized for invalid signatures or unregistered accounts), serialize successful JWT login responses matching the OpenAPI spec, and register routes/handlers in Echo using the Uber Fx container.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/labstack/echo/v4`, `github.com/hros/admin-service/internal/application/usecase`

**Route Mappings**:
- `POST /v1/auth/biometric/challenge` -> `authBiometricHandler.Challenge`
- `POST /v1/auth/biometric/verify` -> `authBiometricHandler.Verify`

**Request/Response DTOs**:
1. **BiometricChallengeRequest**:
   ```go
   type BiometricChallengeRequest struct {
       Email string `json:"email" validate:"required,email"`
   }
   ```
2. **BiometricChallengeResponse**:
   ```go
   type BiometricChallengeResponse struct {
       Challenge    string `json:"challenge"`
       CredentialID string `json:"credential_id"`
   }
   ```
3. **BiometricVerifyRequest**:
   ```go
   type BiometricVerifyRequest struct {
       Email             string `json:"email" validate:"required,email"`
       CredentialID      string `json:"credential_id" validate:"required"`
       AuthenticatorData string `json:"authenticator_data" validate:"required"`
       ClientDataJSON    string `json:"client_data_json" validate:"required"`
       Signature         string `json:"signature" validate:"required"`
       RememberMe        bool   `json:"remember_me"`
   }
   ```

**Response Serialization**:
- Challenge response: Returns `HTTP 200` with `BiometricChallengeResponse` containing base64url challenge and credential ID.
- Verification response: Returns `HTTP 200` with `LoginResponse` containing accessToken, refreshToken, and user details.
- Invalid requests (binding or validation failures): Returns `HTTP 400 Bad Request`.
- Verification errors:
  - `ErrBiometricNotRegistered` -> `HTTP 401 Unauthorized`
  - `ErrInvalidBiometricSignature` -> `HTTP 401 Unauthorized`
  - `ErrChallengeNotFoundOrExpired` -> `HTTP 401 Unauthorized`
  - `ErrUserInactive` -> `HTTP 401 Unauthorized`
  - `ErrUserLocked` -> `HTTP 401 Unauthorized`
- Other errors: `HTTP 500 Internal Server Error`.

**Fx DI Registration**:
- Inject handlers via `fx.Provide(NewAuthBiometricHandler)` in `internal/adapter/http/module.go`.
- Register endpoints in `RegisterRoutes` inside `internal/adapter/http/auth_handler.go` (or a dedicated route registry).

## Constitution Check

- [x] Clean Architecture: HTTP handlers reside strictly in the adapter layer and translate frameworks to UseCase inputs. No business logic leaks into handlers.
- [x] Documentation-First: Endpoints are declared in `api/openapi.yaml`. Update OpenAPI schema if required.
- [x] Unit-Test-Per-File: Handlers tested exhaustively in `auth_biometric_handler_test.go` checking HTTP statuses and error mapping.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
