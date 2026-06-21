# Tasks: Refresh Session Use Case

**Input**: Design documents from `/specs/014-refresh-session-usecase/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

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

**Purpose**: Infrastructure dependencies that must be completed first.

- [x] T002 [US1] Update the `AuditLogger` interface in `internal/domain/auth/audit.go` to include `LogSessionRefreshed(ctx context.Context, userID string)`.
- [x] T003 [US1] Update `SlogAuditLogger` implementation in `internal/infrastructure/auth/audit_logger.go` to implement `LogSessionRefreshed` and update its unit tests in `internal/infrastructure/auth/audit_logger_test.go`.

---

## Phase 3: User Story 1 - Rotate Active Token Session (Priority: P1) 🎯 MVP

**Goal**: Implement `RefreshSessionUseCase` business logic and unit test coverage.

**Independent Test**: Unit tests pass verifying successful token rotation and error propagation (validation, expiration, revocation, database faults).

### Implementation for User Story 1

- [x] T004 [US1] Create the use case file `internal/application/usecase/refresh_session_usecase.go` implementing the token refresh workflow.
- [x] T005 [US1] Create unit tests in `internal/application/usecase/refresh_session_usecase_test.go` verifying the use case against mocked interfaces.
- [x] T006 [US1] Wire the `NewRefreshSessionUseCase` provider in `internal/application/module.go` via Fx.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting, linting, and final validation.

- [x] T007 Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T008 Run tests using `go test -v ./internal/application/usecase/...` to verify the usecase implementation passes cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Complete first.
- **Foundational (Phase 2)**: Core AuditLogger methods updated.
- **User Story 1 (Phase 3)**: Usecase logic implemented and tested.
- **Polish (Phase 4)**: Polish and final checks.
