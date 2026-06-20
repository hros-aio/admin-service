# Tasks: Auth Integration Test

**Input**: Design documents from `/specs/009-auth-integration-test/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: This is a test suite implementation task. The final target is the successful execution of integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Add `testcontainers-go` and its PostgreSQL module to the project dependencies in `go.mod` by running `go get`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*(None required for this slice; project setup is already complete)*

---

## Phase 3: User Story 1 - E2E Admin Authentication Flow (Priority: P1) 🎯 MVP

**Goal**: Implement the integration test `test/integration/auth_flow_test.go` to spin up a real PostgreSQL container, apply schema migrations, seed a test administrator, start the Fx Echo application, and execute API requests to verify successful login and logout.

**Independent Test**: The test is runnable via `go test ./test/integration/... -v` and successfully verifies the full authentication cycle.

### Implementation for User Story 1

- [x] T002 [US1] Create the integration test file `test/integration/auth_flow_test.go` containing the main `TestAuthFlow` runner and the setup of the `testcontainers-go` PostgreSQL container with clean defer teardown
- [x] T003 [US1] Implement a raw SQL migration runner helper inside `test/integration/auth_flow_test.go` to parse and execute statements in `migrations/000001_init.up.sql` and `migrations/000002_create_auth_tables.up.sql` against the GORM connection
- [x] T004 [US1] Implement a seeding helper inside `test/integration/auth_flow_test.go` to insert the system Role and a test AdminUser (with email `test-admin@hros.com` and password `password123` hashed via bcrypt)
- [x] T005 [US1] Implement the Fx application bootstrap in `test/integration/auth_flow_test.go` which overrides the configuration with the test container's database URL and starts the Echo HTTP server on a random port
- [x] T006 [US1] Write test assertions in `test/integration/auth_flow_test.go` that execute a `POST /v1/auth/login` request, verify the 200 status code containing access and refresh tokens, and verify that sending a `DELETE /v1/auth/session` with the refresh token returns a 204 No Content status
- [x] T007 [US1] Write additional test assertions in `test/integration/auth_flow_test.go` verifying that sending incorrect credentials to `POST /v1/auth/login` results in a 401 Unauthorized response

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T008 Run Go formatting (`go fmt ./...`) and linting (`golangci-lint run`) on all modified packages
- [x] T009 Run quickstart.md validation test scenarios using `go test ./test/integration/... -v -count=1` to verify end-to-end functionality

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. Must complete before writing code.
- **User Stories (Phase 3+)**: Depend on Setup phase completion.
- **Polish (Phase 4)**: Depends on all user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 1 setup.

---

## Parallel Opportunities

- Setup tasks and code implementation are largely sequential in this test slice since they build up the same single test file (`test/integration/auth_flow_test.go`).

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 3: User Story 1 (T002 to T007)
3. STOP and VALIDATE: Test User Story 1 independently

### Incremental Delivery

1. Setup ready (T001)
2. Implement container setup and GORM connection (T002)
3. Implement SQL migration execution (T003)
4. Implement data seeding (T004)
5. Bootstrap Fx app server (T005)
6. Add happy path login/logout request assertions (T006)
7. Add error path login assertions (T007)
8. Complete formatting and execution runs (T008, T009)
