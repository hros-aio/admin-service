# Tasks: Biometric Device Login — Integration Test (TSK-BIO-008)

**Input**: Design documents from `/specs/023-biometric-device-login/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (maps to spec.md)
- Include exact file paths in descriptions

## Path Conventions

- Integration tests: `test/integration/`
- No production files are modified in this task

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

*(No setup tasks required — all infrastructure modules, DI wiring, testcontainers, and miniredis are already present in the existing integration test suite.)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core test infrastructure that MUST be complete before test scenarios can be written

**⚠️ CRITICAL**: Scenarios cannot be written until the fixture generator and signing helpers exist

- [x] T001 Create the test file `test/integration/biometric_flow_test.go` with package declaration, all required imports (`crypto/ecdsa`, `crypto/elliptic`, `crypto/sha256`, `crypto/x509`, `encoding/asn1`, `encoding/base64`, `encoding/binary`, `encoding/json`, `math/big`, and existing integration test imports), and the `TestBiometricFlow` test function skeleton
- [x] T002 Implement the `generateTestECDSAKeyPair()` helper in `test/integration/biometric_flow_test.go` — generates a P-256 ECDSA key pair, serialises the public key as base64-std DER-encoded PKIX bytes (for DB seeding), and returns `(*ecdsa.PrivateKey, string)`
- [x] T003 Implement the `buildAuthenticatorData(signCount uint32) []byte` helper in `test/integration/biometric_flow_test.go` — constructs a 37-byte authenticator data buffer: bytes 0–31 = `sha256.Sum256([]byte("localhost"))`, byte 32 = `0x05` (UP+UV flags), bytes 33–36 = big-endian uint32 sign count
- [x] T004 Implement the `buildClientDataJSON(challengeB64url string) []byte` helper in `test/integration/biometric_flow_test.go` — marshals `{"type":"webauthn.get","challenge":"<challengeB64url>","origin":"http://localhost:3000"}` and returns the raw JSON bytes
- [x] T005 Implement the `signWebAuthnAssertion(privKey *ecdsa.PrivateKey, clientDataJSON, authData []byte) []byte` helper in `test/integration/biometric_flow_test.go` — computes `SHA-256(authData || SHA-256(clientDataJSON))`, signs with ECDSA, ASN.1 DER-encodes the signature, and returns raw bytes
- [x] T006 Implement the `base64url(b []byte) string` helper in `test/integration/biometric_flow_test.go` — returns `base64.RawURLEncoding.EncodeToString(b)`

**Checkpoint**: All signing and encoding helpers exist and compile; `TestBiometricFlow` skeleton builds cleanly with `go build ./test/integration/...`

---

## Phase 3: User Story 4 — Full Biometric Flow Integration Test (Priority: P1) 🎯

**Goal**: Prove end-to-end that challenge issuance, cryptographic signing, and verification all work correctly against real PostgreSQL and Redis backends, with valid JWTs issued and sign_count persisted.

**Independent Test**: `go test -p 1 -count=1 -v -run TestBiometricFlow ./test/integration/...`

### Implementation for User Story 4

- [x] T007 [US4] Implement the PostgreSQL testcontainer + miniredis setup block inside `TestBiometricFlow` in `test/integration/biometric_flow_test.go` — spin up `postgres:16-alpine` container via `testcontainers-go/modules/postgres`, run `000001_init.up.sql`, `000002_create_auth_tables.up.sql`, and `000003_add_mfa_to_admin_users.up.sql` migrations using `runMigrationSQLFile` (dollar-quote-aware helper), and start a `miniredis.Run()` instance; mirror the exact setup pattern used in `auth_flow_test.go`
- [x] T008 [US4] Implement test data seeding inside `TestBiometricFlow` in `test/integration/biometric_flow_test.go` — call `generateTestECDSAKeyPair()`, insert a role row and an active admin user row, then update `webauthn_credentials` to `{"id":"test-cred-id","public_key":"<pubKeyBase64>","sign_count":0}` using `db.Exec`
- [x] T009 [US4] Bootstrap the Fx application inside `TestBiometricFlow` in `test/integration/biometric_flow_test.go` — construct `fx.Options` block mirroring `auth_flow_test.go` exactly (all cache providers, repos, application module, kafka mocks, HTTP platform, adapter HTTP module), start the app, wait for the health endpoint at `GET /health` to return 200, defer `app.Stop`
- [x] T010 [US4] Implement the `HappyPath_ChallengeAndVerify` sub-test in `TestBiometricFlow` in `test/integration/biometric_flow_test.go`:
  - POST `{"email":"biometric-test@hros.com"}` to `/v1/auth/biometric/challenge`, assert HTTP 200, decode `challenge` (base64rawurl) and `credential_id`
  - Build `clientDataJSON` from the decoded challenge using `buildClientDataJSON`
  - Build `authenticatorData` with sign count `1` using `buildAuthenticatorData(1)`
  - Sign using `signWebAuthnAssertion`, base64url-encode all fields
  - POST `BiometricVerifyRequest` JSON to `/v1/auth/biometric/verify`, assert HTTP 200, assert `access_token` and `refresh_token` are non-empty strings
  - Query `webauthn_credentials` from DB (scan into `string`) and assert `sign_count` = 1
- [x] T011 [P] [US4] Implement the `InvalidSignature_Returns401` sub-test in `TestBiometricFlow` in `test/integration/biometric_flow_test.go`:
  - POST challenge for the test email, capture challenge bytes
  - Build a valid `authenticatorData` and `clientDataJSON` but corrupt the signature (e.g., flip first byte or submit random bytes as `signature`)
  - POST verify request, assert HTTP 401
- [x] T012 [P] [US4] Implement the `ExpiredChallenge_Returns401` sub-test in `TestBiometricFlow` in `test/integration/biometric_flow_test.go`:
  - Do NOT issue a challenge first (or issue and clear the miniredis key manually)
  - Build a well-formed but unsigned verify request using `buildAuthenticatorData`, `buildClientDataJSON`, and `signWebAuthnAssertion` with a synthetic random challenge
  - POST verify request, assert HTTP 401

**Checkpoint**: All three sub-tests pass with `go test -p 1 -count=1 -v -run TestBiometricFlow ./test/integration/...`

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and full test suite validation

- [x] T013 Run `go fmt ./test/integration/...` and `go vet ./test/integration/...` and resolve any issues in `test/integration/biometric_flow_test.go`
- [x] T014 Run `go test -p 1 -count=1 ./test/integration/...` (all integration tests sequentially) and confirm all tests pass; record output in task notes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: T001 → T002–T006 (T002–T006 can run in parallel once T001 exists)
- **User Story 4 (Phase 3)**: All of Phase 2 must be complete before Phase 3 begins
  - T007 → T008 → T009 → T010 (sequential — each builds on the previous in the same function)
  - T011 and T012 are parallel once T009 (app running) is complete
- **Polish (Phase 4)**: All of Phase 3 must pass before Polish

### Within Phase 2 (parallel opportunities)

```bash
# Once T001 (file + skeleton) exists, run in parallel:
Task T002: generateTestECDSAKeyPair helper
Task T003: buildAuthenticatorData helper
Task T004: buildClientDataJSON helper
Task T005: signWebAuthnAssertion helper
Task T006: base64url helper
```

### Within Phase 3 (parallel opportunities)

```bash
# Once T009 (Fx app running) is complete:
Task T011: InvalidSignature sub-test
Task T012: ExpiredChallenge sub-test
```

---

## Implementation Strategy

### MVP First (Single file, one story)

1. Complete Phase 2 (T001–T006): Create the test file and all crypto helpers.
2. Complete Phase 3 — T007 → T008 → T009 → T010: Bring up infrastructure, seed data, start Fx app, then implement the happy path sub-test.
3. **STOP and VALIDATE**: Run `TestBiometricFlow/HappyPath_ChallengeAndVerify` in isolation.
4. Add T011 and T012 (failure scenarios).
5. Run full test suite (T013 + T014).

### Key Invariants

- All crypto operations mirror exactly what `verifyECDSASignature` in `verify_biometric_usecase.go` expects:
  - Public key: base64-std DER PKIX
  - Signature: ASN.1 DER-encoded ECDSA P-256
  - Signed data: `SHA-256(authenticatorData || SHA-256(clientDataJSON))`
- RP ID hash: `SHA-256("localhost")`
- Allowed origin: `http://localhost:3000`
- Sign count in seed: `0`; in authenticatorData: `1`; expected in DB after verify: `1`
- Use `runMigrationSQLFile` (not `runSQLFile`) for migration 000003 which contains `$$...$$` PL/pgSQL blocks

---

## Notes

- [P] tasks = different files or logically independent sections, no shared state
- [US4] label maps tasks to User Story 4 (Full Biometric Flow Integration Test) from spec.md
- No production code is modified in this task
- Run integration tests with `-p 1` to avoid Docker resource contention between parallel test suites
- Commit after T014 passes using conventional commit: `test(auth): TSK-BIO-008 add biometric flow integration test`
