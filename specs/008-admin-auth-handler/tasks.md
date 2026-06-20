# Tasks: Admin Auth Handler

**Input**: Design documents from `/specs/008-admin-auth-handler/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Unit tests are MANDATORY per file. Write unit tests using Echo's `httptest` utilities to assert correct 200, 204, 400, 401, and 403 HTTP status code responses.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

*(None required for this slice; project setup is already complete)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 [P] Register `usecase.NewLogoutUseCase` constructor in the application Fx module `internal/application/module.go`
- [x] T002 [P] Create `internal/adapter/http/auth_handler.go` with `AuthHandler` struct, constructor `NewAuthHandler`, and routing registration skeleton `RegisterRoutes`
- [x] T003 [P] Create Fx HTTP adapter module in `internal/adapter/http/module.go` to declare Fx module `"http-adapter"`, providing `NewAuthHandler` and invoking `RegisterRoutes`
- [x] T004 Register the new `"http-adapter"` module in the root app composition inside `internal/app/app.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Admin Login Authentication (Priority: P1) 🎯 MVP

**Goal**: Implement the `POST /v1/auth/login` API route. Bind request payload to DTO, validate, execute `LoginUseCase`, map domain errors to HTTP statuses, and serialize responses.

**Independent Test**: Send a POST request to `/v1/auth/login` with valid/invalid credentials and verify HTTP 200, 400, 401, or 403 responses.

### Implementation for User Story 1

- [x] T005 [US1] Implement `Login` endpoint logic in `internal/adapter/http/auth_handler.go` (bind `dto.LoginRequest`, run struct validation, execute `LoginUseCase`, serialize `dto.LoginResponse`, and map domain errors such as `ErrInvalidCredentials` to HTTP 401, `ErrUserInactive` to HTTP 403, and `ErrUserLocked` to HTTP 403)
- [x] T006 [P] [US1] Implement unit tests in `internal/adapter/http/auth_handler_test.go` covering success (200), invalid credentials (401), account inactive (403), account locked (403), and binding/validation failures (400)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Admin Logout Session Termination (Priority: P1)

**Goal**: Implement the `DELETE /v1/auth/session` API route. Extract the bearer token from the Authorization header, execute `LogoutUseCase` to revoke/delete the session, and return 204 No Content.

**Independent Test**: Send a DELETE request to `/v1/auth/session` with a bearer token and verify the 204 status code response.

### Implementation for User Story 2

- [x] T007 [US2] Implement `Logout` endpoint logic in `internal/adapter/http/auth_handler.go` (extract bearer token from Authorization header, execute `LogoutUseCase.Execute` with token value, map use case output, and return HTTP 204 status)
- [x] T008 [P] [US2] Implement unit tests in `internal/adapter/http/auth_handler_test.go` covering successful session revocation (204) and missing/malformed auth token header (401/400)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T009 Run Go formatting (`go fmt ./...`) and linting (`golangci-lint run`) on all modified packages
- [x] T010 Run quickstart.md validation test scenarios using `go test ./internal/adapter/http/... -race -count=1` to verify end-to-end functionality

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No prerequisites.
- **Foundational (Phase 2)**: Core constructor registration and module composition. BLOCKS all user stories.
- **User Stories (Phase 3+)**: Depend on Foundational phase completion. User stories can be worked on sequentially or in parallel.
- **Polish (Phase 5)**: Depends on both user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on US2.
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on US1.

---

## Parallel Opportunities

- Foundational tasks `T001`, `T002`, and `T003` can be worked on in parallel.
- `T006` (US1 unit tests) can be written in parallel with `T005` (US1 handler implementation).
- `T008` (US2 unit tests) can be written in parallel with `T007` (US2 handler implementation).

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001 to T004)
2. Complete Phase 3: User Story 1 (T005, T006)
3. STOP and VALIDATE: Test User Story 1 independently

### Incremental Delivery

1. Foundation ready (T001 - T004)
2. Add User Story 1 (T005 - T006) → Test independently → Deliver (MVP!)
3. Add User Story 2 (T007 - T008) → Test independently → Deliver
4. Complete Polish & Validation (T009 - T010)
