# Tasks: Admin User Repository (Fetch by Email)

**Input**: Design documents from `/specs/004-admin-user-repository/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Unit tests per file are MANDATORY as per Constitution Principle III.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create directory structure for the auth repository in `internal/infrastructure/repository/auth/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure setup

- [X] T002 [P] Define `adminUserModel` with GORM tags in `internal/infrastructure/repository/auth/model.go`
- [X] T003 [P] Implement `toDomain` and `fromDomain` mapping functions in `internal/infrastructure/repository/auth/mapper.go`

---

## Phase 3: User Story 1 - Retrieve Admin User by Email (Priority: P1) 🎯 MVP

**Goal**: Implement the core functionality to fetch an admin user by email using GORM.

**Independent Test**: Verify via `sqlmock` that a valid email returns the expected `AdminUser` domain entity.

### Implementation for User Story 1

- [X] T004 [P] [US1] Implement `NewGormAdminUserRepository` constructor in `internal/infrastructure/repository/auth/repository.go`
- [X] T005 [US1] Implement `FindByEmail` method using GORM in `internal/infrastructure/repository/auth/repository.go`
- [X] T006 [US1] Create unit test for `FindByEmail` success case using `sqlmock` in `internal/infrastructure/repository/auth/repository_test.go`

---

## Phase 4: User Story 2 - Handle Missing Admin User (Priority: P2)

**Goal**: Ensure the repository handles "not found" and database errors correctly without leaking GORM details.

**Independent Test**: Verify via `sqlmock` that a missing email returns `domainErrors.ErrUserNotFound`.

### Implementation for User Story 2

- [X] T007 [US2] Update `FindByEmail` to map `gorm.ErrRecordNotFound` to `domainErrors.ErrUserNotFound` in `internal/infrastructure/repository/auth/repository.go`
- [X] T008 [US2] Implement unit tests for "not found" and database error cases in `internal/infrastructure/repository/auth/repository_test.go`

---

## Phase 5: Polish & Integration

**Purpose**: Finalize implementation and wire into the application.

- [X] T009 [P] Register `GormAdminUserRepository` in the root Fx module in `internal/app/app.go`
- [X] T010 [P] Run `go test ./internal/infrastructure/repository/auth/...` to verify all cases
- [X] T011 [P] Run `golangci-lint run ./internal/infrastructure/repository/auth/...` for quality check

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1.
- **User Story 1 (Phase 3)**: Depends on Phase 2.
- **User Story 2 (Phase 4)**: Depends on Phase 3.
- **Polish (Phase 5)**: Depends on Phase 4.

### Parallel Opportunities

- T002 and T003 can be implemented in parallel.
- T009, T010, and T011 can be performed in parallel after implementation is complete.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

The MVP focus is the successful retrieval of a user by email. This proves the end-to-end flow from the database to the domain entity.

1. Complete Setup and Foundational models/mappers.
2. Implement the success path for `FindByEmail`.
3. Validate with unit tests.

### Incremental Delivery

1. Foundation ready (Model + Mapper).
2. US1: Success path implemented and tested.
3. US2: Error handling and mapping refined.
4. Final Integration: Fx wiring and linting.
