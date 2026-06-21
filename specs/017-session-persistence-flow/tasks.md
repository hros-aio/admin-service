# Tasks: Session Persistence Flow Test

**Input**: Design documents from `/specs/017-session-persistence-flow/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: This task list includes integration tests in `test/integration/` and updates unit tests in `internal/adapter/http/`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Setup the feature plan and workspace state (already completed during spec/plan phases)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure updates that MUST be complete before user story testing can begin

**⚠️ CRITICAL**: No user story integration testing can begin until the handler is updated to pass the access token on logout

- [x] T002 [P] Update `Logout` handler method in `internal/adapter/http/auth_handler.go` to support receiving the Access Token (via `Authorization` header) and the Refresh Token (via `X-Refresh-Token` header) simultaneously for blacklist registration and session deletion.
- [x] T003 [P] Update unit tests in `internal/adapter/http/auth_handler_test.go` to assert the updated `Logout` header parsing logic (checking that if `X-Refresh-Token` header is present, `AccessToken` is populated with the Authorization Bearer token and `RefreshToken` is populated with `X-Refresh-Token`).

**Checkpoint**: Foundation ready - logout handler supports multi-token inputs, and integration flow testing can begin.

---

## Phase 3: User Story 1 - Verify Long-term Session Persistence (Priority: P1) 🎯 MVP

**Goal**: Verify login with `remember_me=true` creates a persistent 30-day session in containerized PostgreSQL database.

**Independent Test**: Execute login integration test using testcontainers PostgreSQL.

### Implementation for User Story 1

- [x] T004 [US1] Create integration test file `test/integration/session_persistence_flow_test.go` and implement the test setup booting the PostgreSQL container via `testcontainers-go` and executing migration scripts.
- [x] T005 [US1] Implement `TestSessionPersistenceFlow` login integration flow asserting that logging in with `remember_me=true` creates a session in PostgreSQL with an expiration date set to 30 days in the future and `is_persistent=true`.

**Checkpoint**: User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Verify Secure Session Rotation & Blacklisting (Priority: P1)

**Goal**: Verify session token rotation during refresh and access token JTI blacklisting on logout using containerized PostgreSQL and Redis.

**Independent Test**: Execute integration test checking `/v1/auth/refresh` rotates tokens, and `/v1/auth/session` blacklists the old access token in containerized Redis.

### Implementation for User Story 2

- [x] T006 [US2] Update `test/integration/session_persistence_flow_test.go` to boot a containerized Redis instance using generic `testcontainers-go` container options.
- [x] T007 [US2] Implement token rotation integration test verifying `POST /v1/auth/refresh` rotates the session token and updates the database record.
- [x] T008 [US2] Implement logout integration test verifying `DELETE /v1/auth/session` deletes the session token in PostgreSQL and writes the access token JTI to the Redis blacklist cache, checking keys directly via the Redis container client.
- [x] T009 [US2] Implement integration test case verifying old access token rejection after token rotation flow in `test/integration/session_persistence_flow_test.go`.
- [x] T010 [US2] Implement integration test cases verifying refresh endpoint edge cases (invalid, expired, and empty tokens) in `test/integration/session_persistence_flow_test.go`.

**Checkpoint**: User Stories 1 and 2 are fully functional and verified end-to-end.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Format checking, linting, and final validation.

- [x] T011 [P] Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T012 Run all integration tests using `go test -v ./test/integration/...` to ensure all tests pass cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - blocks integration flow tests since handler needs to support the token blacklist mapping.
- **User Stories (Phase 3+)**: Depend on Foundational completion.
- **Polish (Final Phase)**: Depends on all stories being completed.

### Parallel Opportunities

- T002 and T003 can be worked on in parallel.
- Once Phase 2 completes, Phase 3 (PostgreSQL flow) and Phase 4 (Redis container integration and rotation/blacklisting tests) are implemented sequentially in the same test file to verify the progressive login-refresh-logout cycle.
