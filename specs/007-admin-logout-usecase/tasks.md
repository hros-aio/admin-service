# Tasks: Admin Logout Use Case

**Input**: Design documents from `/specs/007-admin-logout-usecase/`

**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: Unit tests per file are MANDATORY as per Constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Update domain interfaces and infrastructure implementations to support logout auditing.

- [ ] T001 [US3] Add `LogLogoutSuccess(ctx context.Context, token string)` to `AuditLogger` interface in `internal/domain/auth/audit.go`
- [ ] T002 [US3] Implement `LogLogoutSuccess` in `internal/infrastructure/auth/audit_logger.go`
- [ ] T003 [P] Add mock implementation of `LogLogoutSuccess` to test files where `AuditLogger` mock is used (e.g., `internal/application/usecase/login_usecase_test.go`) to ensure they compile

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the types required for the logout use case.

- [ ] T004 [P] [US1] Create `LogoutInput` struct in `internal/application/usecase/logout_types.go`

**Checkpoint**: Foundation ready - use case implementation can now begin.

---

## Phase 3: User Story 1 - Session Termination (Priority: P1) 🎯 MVP

**Goal**: Implement the core session deletion functionality.

**Independent Test**: Execute `LogoutUseCase` with a valid token and verify it calls `DeleteByToken` on repository.

### Implementation for User Story 1

- [ ] T005 [US1] Create `LogoutUseCase` structure and constructor in `internal/application/usecase/logout_usecase.go`
- [ ] T006 [US1] Implement `Execute` method in `internal/application/usecase/logout_usecase.go` that calls `DeleteByToken` on `SessionTokenRepository`
- [ ] T007 [US1] Create unit tests for successful logout scenario in `internal/application/usecase/logout_usecase_test.go`

**Checkpoint**: User Story 1 is functional for happy-path logout.

---

## Phase 4: User Story 2 - Failed Logout (Priority: P2)

**Goal**: Handle missing session tokens correctly.

**Independent Test**: Execute `LogoutUseCase` with an invalid token and assert it returns `ErrTokenNotFound`.

### Implementation for User Story 2

- [ ] T008 [US2] Implement error checking and propagation of `ErrTokenNotFound` or repository errors in `LogoutUseCase.Execute`
- [ ] T009 [US2] Add unit tests for session token not found and repository error cases in `internal/application/usecase/logout_usecase_test.go`

**Checkpoint**: Failed attempts are safely rejected and signaled.

---

## Phase 5: User Story 3 - Audit Trail (Priority: P3)

**Goal**: Emit audit logs for successful logouts.

**Independent Test**: Verify audit logger invocation on successful logout.

### Implementation for User Story 3

- [ ] T010 [US3] Add audit event trigger `LogLogoutSuccess` to `LogoutUseCase.Execute` on success
- [ ] T011 [US3] Add unit test assertion to verify `LogLogoutSuccess` was called in `internal/application/usecase/logout_usecase_test.go`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Fx module registration and cleanup.

- [ ] T012 [P] Register `LogoutUseCase` in Fx module at `internal/application/module.go`
- [ ] T013 [P] Run `go test ./...` and ensure all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Must be completed first to define the interface.
- **Foundational (Phase 2)**: Defines the input type.
- **User Stories (Phase 3-5)**: Implement the core functionality step-by-step.
- **Polish (Phase 6)**: Fx registration.
