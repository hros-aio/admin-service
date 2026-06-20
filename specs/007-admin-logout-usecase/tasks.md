# Tasks: Admin Logout Use Case

**Input**: Design documents from `/specs/007-admin-logout-usecase/`

**Prerequisites**: plan.md (required), spec.md (required)

**Tests**: Unit tests are MANDATORY per file. Write unit tests to achieve 100% coverage on the new usecase code.

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

- [X] T001 Update `AuditLogger` interface in `internal/domain/auth/audit.go` to include `LogLogoutSuccess(ctx context.Context, token string)`
- [X] T002 Implement `LogLogoutSuccess` in `internal/infrastructure/auth/audit_logger.go`
- [X] T003 Create and implement unit tests in `internal/infrastructure/auth/audit_logger_test.go` to verify `LogLogoutSuccess` format and behavior

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Secure Session Termination (Priority: P1) 🎯 MVP

**Goal**: Implement the primary logout usecase flow: delete the session token by value via the repository and emit a successful logout audit event.

**Independent Test**: Verify `LogoutUseCase` executes successfully and emits a success audit log.

### Implementation for User Story 1

- [X] T004 [US1] Create the struct definitions and implement `LogoutUseCase.Execute` in `internal/application/usecase/logout_usecase.go` for valid tokens
- [X] T005 [US1] Create and write unit tests in `internal/application/usecase/logout_usecase_test.go` to verify the execution flow for a valid token, checking repository deletion and audit log emission

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Idempotent Logout (Priority: P2)

**Goal**: Support idempotent deletion when the session token does not exist in persistence, and handle database/repository errors gracefully.

**Independent Test**: Verify `LogoutUseCase` handles non-existent/invalid tokens gracefully and propagates repository errors.

### Implementation for User Story 2

- [X] T006 [US2] Update `LogoutUseCase.Execute` in `internal/application/usecase/logout_usecase.go` to ensure idempotent handling (returns success when deleting a non-existent token) and correct error propagation on repository failures
- [X] T007 [US2] Add unit tests in `internal/application/usecase/logout_usecase_test.go` covering non-existent tokens and repository deletion error scenarios

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, formatting, and linting

- [X] T008 [P] Verify code formatting and run `golangci-lint` to ensure no linting errors
- [X] T009 Run quickstart.md validation test scenarios using `go test ./internal/application/usecase/... -race -count=1`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No prerequisites - can start immediately. BLOCKS all user stories.
- **User Stories (Phase 3+)**: All depend on Foundational phase completion.
- **Polish (Phase 5)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories.
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May build upon US1 structure.

---

## Parallel Opportunities

- T003 (unit tests for SlogAuditLogger) can be implemented in parallel with T002.
- Polish tasks (Phase 5) can run in parallel where possible.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001, T002, T003)
2. Complete Phase 3: User Story 1 (T004, T005)
3. STOP and VALIDATE: Test User Story 1 independently

### Incremental Delivery

1. Foundation ready
2. Add User Story 1 → Test independently → Deliver (MVP!)
3. Add User Story 2 → Test independently → Deliver
4. Complete Polish and run validation scenarios
