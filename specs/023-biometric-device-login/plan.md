# Implementation Plan: Biometric Device Login — Integration Test (TSK-BIO-008)

**Branch**: `023-biometric-device-login` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

## Summary

Implement a Go integration test (`test/integration/biometric_flow_test.go`) that exercises
the full biometric login flow — challenge → sign → verify — against real PostgreSQL and
Redis backends provisioned by `testcontainers`. The test seeds an admin user with a
pre-generated ECDSA P-256 public key, executes the live HTTP endpoints wired via Uber Fx,
cryptographically signs the challenge with the matching private key, and asserts that valid
JWTs are issued and that the `sign_count` in the database is incremented.

No new production files are created. This task adds only one test file.

---

## Technical Context

**Language/Version**: Go 1.23+

**Test infrastructure** (already in project):
- `testcontainers-go` + `testcontainers-go/modules/postgres` — real PostgreSQL 16-alpine container
- `miniredis/v2` — in-process Redis for test isolation (already used in all other flow tests)
- `uber/fx` — full application DI bootstrapped identically to `auth_flow_test.go`
- `stretchr/testify` — assertions

**Cryptographic strategy** (stdlib only — no new dependencies):
- Generate an ECDSA P-256 private key with `crypto/elliptic` and `crypto/ecdsa`
- Serialize the public key as DER-encoded PKIX bytes with `crypto/x509`
- Persist public key in `webauthn_credentials` JSONB as:
  ```json
  {"id":"test-cred-id","public_key":"<base64std-DER>","sign_count":0}
  ```
- Build the `clientDataJSON` payload manually (same struct the usecase parses):
  ```json
  {"type":"webauthn.get","challenge":"<base64rawurl-of-challenge>","origin":"http://localhost:3000"}
  ```
- Build `authenticatorData` (≥37 bytes):
  - Bytes 0–31: `SHA-256("localhost")` — the RP ID hash
  - Byte 32: `0x05` — UP (bit 0) + UV (bit 2) flags set
  - Bytes 33–36: sign count = `0x00000001` (big-endian uint32, incremented from 0)
- Compute the ECDSA P-256 signature over `SHA-256(authenticatorData || SHA-256(clientDataJSON))`
  using ASN.1 DER encoding (same as `verifyECDSASignature`)
- Base64url-encode all binary fields before sending to the HTTP endpoint

**Allowed origins (hardcoded in usecase)**:
- `http://localhost:3000` ← used in tests

**RP ID hash** (hardcoded in usecase):
- `SHA-256("localhost")` ← used in tests

**Sign count monotonicity**:
- Seed `sign_count = 0` in DB
- Send `sign_count = 1` in `authenticatorData`
- Assert DB column incremented to 1 after verify

**Response shapes**:
- Challenge: `{"challenge": "<base64rawurl>", "credential_id": "<string>"}`
- Verify success: `{"access_token": "...", "refresh_token": "..."}`
- Verify failure (bad sig, expired challenge): HTTP 401

**Test structure** (matches existing `TestAuthFlow` pattern):
1. Spin up PostgreSQL testcontainer
2. Run SQL migrations (`000001_init.up.sql`, `000002_create_auth_tables.up.sql`)
3. Seed role + admin user + `webauthn_credentials` JSONB
4. Spin up miniredis
5. Bootstrap Fx app (same `opts` pattern as `auth_flow_test.go`)
6. Wait for health endpoint ready
7. Execute sub-tests in sequence using `t.Run`

**Scenarios** (sub-tests):
| # | Name | Action | Expected |
|---|------|---------|----------|
| 1 | `HappyPath_ChallengeAndVerify` | Full valid flow | HTTP 200, JWT tokens, sign_count incremented |
| 2 | `InvalidSignature_Returns401` | Corrupt signature bytes | HTTP 401 |
| 3 | `ExpiredChallenge_Returns401` | Verify without issuing challenge first | HTTP 401 |

---

## Constitution Check

- [x] **Clean Architecture**: Test lives in `test/integration/` (outside `internal/`). No production logic is modified.
- [x] **Documentation-First**: No new API endpoints. No OpenAPI changes required.
- [x] **Unit-Test-Per-File**: No new production files — no new `_test.go` unit files required.
- [x] **Task-Driven**: Implements exactly TSK-BIO-008 — one integration test file.
- [x] **Observability**: Integration test logger discards output (`io.Discard`) — consistent with all other flow tests.

---

## Phase 0: Research

**Decision: Cryptographic approach for test signing**
- **Chosen**: stdlib `crypto/ecdsa` + `crypto/asn1` + `crypto/x509` — no new imports.
- **Rationale**: The usecase `verifyECDSASignature` function uses `x509.ParsePKIXPublicKey` and `asn1.Unmarshal`. The test must produce output that passes through this exact code path.
- **Alternative rejected**: `github.com/go-webauthn/webauthn` — would add a large indirect dependency and is unnecessary since the usecase performs its own low-level crypto.

**Decision: Redis backend**
- **Chosen**: `miniredis` (already in use in all 7 flow tests).
- **Rationale**: `miniredis` supports the `EVAL` Lua script used by `VerifyAndConsumeChallenge`. Confirmed by existing `webauthn_redis_test.go`.

**Decision: Fx bootstrap**
- **Chosen**: Mirror `auth_flow_test.go` `fx.Options` block verbatim, adding no new Fx modules.
- **Rationale**: All modules are already registered. The biometric UseCases are included in `application.Module`.

---

## Phase 1: Design

### Data Model

No schema changes. The existing `admin_users.webauthn_credentials` JSONB column (type `[]byte` in GORM model) holds the credential. The test seeds it directly via raw SQL:

```sql
UPDATE admin_users
   SET webauthn_credentials = '{"id":"test-cred-id","public_key":"<base64>","sign_count":0}'::jsonb
 WHERE email = 'biometric-test@hros.com';
```

### Test Helper Functions (within `biometric_flow_test.go`)

| Function | Purpose |
|----------|---------|
| `generateTestECDSAKeyPair()` | Generates ECDSA P-256 key pair, returns `(*ecdsa.PrivateKey, string)` where the string is base64-std DER public key for DB seeding |
| `buildAuthenticatorData(signCount uint32)` | Returns 37-byte authenticator data with SHA-256("localhost") RP ID hash, flags=0x05, and big-endian sign count |
| `buildClientDataJSON(challengeB64 string)` | Returns marshalled `clientData` JSON with `type=webauthn.get`, extracted challenge, `origin=http://localhost:3000` |
| `signWebAuthnAssertion(privKey, clientDataJSON, authData)` | Computes `SHA-256(authData || SHA-256(clientDataJSON))`, ECDSA-signs it, ASN.1-encodes, returns `[]byte` |
| `base64url(b []byte) string` | Returns `base64.RawURLEncoding.EncodeToString(b)` |

### HTTP Flow in Test

```
POST /v1/auth/biometric/challenge
  body: {"email":"biometric-test@hros.com"}
  → 200 {"challenge":"<base64rawurl>","credential_id":"test-cred-id"}

Build clientDataJSON using challenge from response
Build authenticatorData (sign_count=1)
Compute ECDSA signature

POST /v1/auth/biometric/verify
  body: {
    "email":"biometric-test@hros.com",
    "credential_id":"test-cred-id",
    "authenticator_data":"<base64rawurl>",
    "client_data_json":"<base64rawurl>",
    "signature":"<base64rawurl>",
    "remember_me":false
  }
  → 200 {"access_token":"<jwt>","refresh_token":"<jwt>"}
```

### Assertions After Verify

1. HTTP status 200
2. `access_token` is non-empty
3. `refresh_token` is non-empty
4. DB query: `SELECT webauthn_credentials FROM admin_users WHERE email = ?` → `sign_count` = 1

### DB sign_count Verification

After the verify HTTP call:
```go
var rawCreds []byte
db.Raw("SELECT webauthn_credentials FROM admin_users WHERE email = ?", email).Scan(&rawCreds)
var cred struct {
    SignCount uint32 `json:"sign_count"`
}
json.Unmarshal(rawCreds, &cred)
assert.Equal(t, uint32(1), cred.SignCount)
```

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Manual ECDSA signing in test | Required to produce a cryptographically valid assertion matching the exact usecase verification path | No external WebAuthn library; test must replicate the protocol in-band |
| Health-poll loop | Testcontainer + Fx startup race condition mitigation | Already used in all flow tests |
