# Tasks: Auth Token Rotation

**Input**: Design documents from `/specs/010-auth-token-rotation/`

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

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*(None required for this slice; base infrastructure is already complete)*

---

## Phase 3: User Story 1 - Session Token Rotation (Priority: P1) 🎯 MVP

**Goal**: Update the `SessionToken` domain entity to support explicit expiration tracking and add a `Rotate()` helper method to generate secure token strings.

**Independent Test**: Verification via unit tests in `internal/domain/session_token_test.go` asserting that `Rotate()` generates a secure random hex token string and updates the model's fields correctly.

### Implementation for User Story 1

- [x] T002 [US1] Update the `SessionToken` domain entity in `internal/domain/session_token.go` to add the `Rotate(newExpiry time.Time) (string, error)` helper method using `crypto/rand`
- [x] T003 [P] [US1] Update unit tests in `internal/domain/session_token_test.go` to assert correct rotation logic, new token format, updated expiration, and random number generator failure handling

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Token Blacklisting (Priority: P2)

**Goal**: Define a `TokenBlacklist` cache interface in the application layer for immediate JWT or refresh token revocation.

**Independent Test**: The interface file compiles successfully and is ready for usecase injections and mock implementations.

### Implementation for User Story 2

- [x] T004 [US2] Create the application interface file `internal/application/interfaces/cache.go` defining the `TokenBlacklist` interface with `Add` and `Exists` methods

**Checkpoint**: At this point, User Stories 1 and 2 are complete.

---

## Phase 5: User Story 3 - Specific Domain Error Identification (Priority: P2)

**Goal**: Define the `ErrInvalidRefreshToken` domain error and ensure standard token errors are comparable.

**Independent Test**: Verification via unit tests in `internal/domain/errors/auth_errors_test.go` asserting the new error definitions are present and correct.

### Implementation for User Story 3

- [x] T005 [US3] Update the domain errors file `internal/domain/errors/auth_errors.go` to define the `ErrInvalidRefreshToken` variable using `errors.New`
- [x] T006 [P] [US3] Update unit tests in `internal/domain/errors/auth_errors_test.go` to assert the definition and error message of `ErrInvalidRefreshToken`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Formatting, linting, and final validation of the domain layer code changes.

- [x] T007 Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files
- [x] T008 Run quickstart.md validation tests using `go test ./internal/domain/...` to verify domain logic passes cleanly

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. Complete first.
- **User Stories (Phases 3-5)**: Depends on setup completion.
  - US1 (P1) is the MVP and should be completed first.
  - US2 (P2) and US3 (P2) can run in parallel with each other after US1 is completed.
- **Polish (Phase 6)**: Depends on all user stories being complete.

### Within Each User Story

- The entity/error definitions must be updated before updating their corresponding test files.

### Parallel Opportunities

- Once Phase 3 (US1) is complete, Phase 4 (US2) and Phase 5 (US3) can be worked on concurrently since they affect different packages and files (`internal/application/interfaces/cache.go` vs `internal/domain/errors/auth_errors.go`).

---

## Parallel Example: User Stories 2 & 3

```bash
# Developer A implements US2 (Token Blacklisting):
Task: "Create the application interface file internal/application/interfaces/cache.go defining the TokenBlacklist interface..."

# Developer B implements US3 (Specific Domain Errors):
Task: "Update the domain errors file internal/domain/errors/auth_errors.go to define the ErrInvalidRefreshToken variable..."
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Implement rotation method in domain entity (T002).
2. Add unit tests for rotation logic (T003).
3. **STOP and VALIDATE**: Run `go test ./internal/domain/...` to assert success.

### Incremental Delivery

1. Foundation ready (T001).
2. Implement and test US1 (T002, T003).
3. Implement US2 (T004).
4. Implement and test US3 (T005, T006).
5. Format and validate the slice (T007, T008).
