# Implementation Plan: Biometric Device Login (WebAuthn) - UseCase Layer (TSK-BIO-006)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Implement `VerifyBiometricUseCase` which verifies WebAuthn cryptographic signatures against the stored public key, handles replay prevention using cached challenges, updates the credential sign count monotonically, logs audit trails, and issues a valid session (JWT + database-persisted session token).

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `crypto/ecdsa`, `crypto/sha256`, `crypto/x509`, `encoding/asn1`, `encoding/base64`, `encoding/json`, `encoding/pem`

**Verification Flow (W3C WebAuthn compliant)**:
1. Accept email, credentialID, clientDataJSON, authenticatorData, and signature.
2. Retrieve and consume the cached challenge from Redis via `WebAuthnChallengeCache.VerifyAndConsumeChallenge()`.
3. Unmarshal `clientDataJSON` and compare the base64url-decoded challenge field with the cached challenge bytes.
4. Retrieve the `AdminUser` from Postgres, parse their PEM public key from `webauthn_credentials`, and verify the credential ID matches.
5. Compute the hash of `clientDataJSON` using SHA-256.
6. Verify the ECDSA signature over the concatenation of `authenticatorData` and the client data hash using the user's public key.
7. Upon successful validation, update the sign count in the repository via `UpdateWebAuthnSignCount()`.
8. Log a `login.biometric_success` event.
9. Generate JWT access and refresh tokens, persist a `SessionToken` record to the database, and return the session details.

**Testing**: Mocked tests in `verify_biometric_usecase_test.go` with 100% coverage using mock repositories/caches/loggers.

## Constitution Check

- [x] Clean Architecture: All repository, cache, token generation, and audit logging actions are decoupled via interfaces.
- [x] Security-First: Challenge verification uses single-use consumption and cryptographically secure ECDSA P-256 checks.
- [x] Unit-Test-Per-File: Unit tests located in `verify_biometric_usecase_test.go`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | N/A        | N/A                                 |
