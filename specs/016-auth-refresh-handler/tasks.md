# Tasks: Auth Refresh Handler

**Input**: Design documents from `/specs/016-auth-refresh-handler/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Unit tests are required for every production file modified.

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

*No blocking foundational tasks needed.*

---

## Phase 3: User Story 1 - Auth Session Token Rotation (Priority: P1) 🎯 MVP

**Goal**: Implement the POST /v1/auth/refresh HTTP endpoint route, handler logic, and unit tests in AuthHandler.

**Independent Test**: Verification via handler unit tests asserting HTTP 200 on success, and HTTP 400/401/403/500 mappings.

### Implementation for User Story 1

- [x] T002 [US1] Update `AuthHandler` struct definition and `NewAuthHandler` constructor in `internal/adapter/http/auth_handler.go` to accept and store `refreshUC *usecase.RefreshSessionUseCase`.
- [x] T003 [US1] Add `POST /v1/auth/refresh` route mapping in `RegisterRoutes` inside `internal/adapter/http/auth_handler.go`.
- [x] T004 [US1] Implement `Refresh(c echo.Context) error` method in `AuthHandler` in `internal/adapter/http/auth_handler.go` to bind request, run struct validation, invoke `RefreshSessionUseCase.Execute`, serialize response, and map errors.
- [x] T005 [US1] Update `internal/adapter/http/auth_handler_test.go` unit tests to cover and assert `/v1/auth/refresh` validation, successful rotation, 401 token invalid/expired mappings, 403 user inactive/locked mappings, and 500 internal usecase error mappings.

**Checkpoint**: User Story 1 is fully functional and testable.

---

## Phase 4: User Story 2 - Pass Remember Me Selection on Login (Priority: P1) 🎯 MVP

**Goal**: Pass the `RememberMe` parameter from the login HTTP request DTO down to the LoginUseCase.

**Independent Test**: Assert mapping parameter in login handler unit tests.

### Implementation for User Story 2

- [x] T006 [US2] Update `Login(c echo.Context) error` method in `AuthHandler` in `internal/adapter/http/auth_handler.go` to map and pass `RememberMe` parameter from `dto.LoginRequest` to `usecase.LoginInput`.
- [x] T007 [US2] Update unit tests for login in `internal/adapter/http/auth_handler_test.go` to assert that the `RememberMe` boolean is correctly propagated to `LoginUseCase`.

**Checkpoint**: User Story 2 is fully functional and testable.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Format checking, linting, and final validation.

- [x] T008 [P] Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T009 Run all tests using `go test -v ./internal/adapter/http/...` and `go test -v ./test/integration/...` to ensure all tests pass cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **User Story 1 (Phase 3)**: Independent, blocks User Story 2 integration test (but handler unit tests can run independently).
- **User Story 2 (Phase 4)**: Independent.
- **Polish (Phase 5)**: Depends on completion of all stories.
