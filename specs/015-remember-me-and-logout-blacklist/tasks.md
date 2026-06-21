# Tasks: Remember Me and Logout Blacklist

**Input**: Design documents from `/specs/015-remember-me-and-logout-blacklist/`

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

**Purpose**: Core infrastructure that must be complete before any user story can be implemented

*No blocking foundational tasks needed.*

---

## Phase 3: User Story 1 - Remember Me Session Expiration (Priority: P1) 🎯 MVP

**Goal**: Implement dynamic session expiration logic based on the `RememberMe` parameter during login.

**Independent Test**: Assert dynamic expiration calculation in the `LoginUseCase` unit tests.

### Implementation for User Story 1

- [x] T002 [US1] Update `LoginInput` struct in `internal/application/usecase/login_types.go` to add `RememberMe` boolean field.
- [x] T003 [US1] Update `LoginUseCase.Execute` in `internal/application/usecase/login_usecase.go` to set `ExpiresAt` to 30 days if `RememberMe` is true, or 24 hours (short-lived browser session) if false, and set `IsPersistent` to `RememberMe`.
- [x] T004 [US1] Update unit tests in `internal/application/usecase/login_usecase_test.go` to cover and verify the dynamic session expiration logic.

**Checkpoint**: User Story 1 is fully functional and testable.

---

## Phase 4: User Story 2 - Immediate Access Token Blacklisting on Logout (Priority: P1) 🎯 MVP

**Goal**: Update `LogoutUseCase` to extract the JWT access token's JTI and add it to the `TokenBlacklist` cache.

**Independent Test**: Assert parsing, claim extraction, and blacklist caching in the `LogoutUseCase` unit tests.

### Implementation for User Story 2

- [x] T005 [P] [US2] Update the JWT Token Provider implementation in `internal/infrastructure/auth/jwt_provider.go` to include a unique `jti` (JWT ID) claim when generating access tokens, and update `internal/infrastructure/auth/jwt_provider_test.go` to assert this claim exists.
- [x] T006 [US2] Update `LogoutInput` struct in `internal/application/usecase/logout_usecase.go` to add `AccessToken` string field.
- [x] T007 [US2] Update the `LogoutUseCase` struct and NewLogoutUseCase constructor in `internal/application/usecase/logout_usecase.go` to accept and store the `interfaces.TokenBlacklist` dependency.
- [x] T008 [US2] Implement unverified JWT access token parsing, JTI and expiration extraction, and `TokenBlacklist.Add` integration in `LogoutUseCase.Execute` in `internal/application/usecase/logout_usecase.go`.
- [x] T009 [US2] Update unit tests in `internal/application/usecase/logout_usecase_test.go` to verify the JTI extraction, remaining TTL calculation, and blacklisting behavior using a mock of `TokenBlacklist`.

**Checkpoint**: User Story 2 is fully functional and testable.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Format checking, linting, and final validation.

- [x] T010 [P] Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T011 Run all unit tests using `go test -v ./internal/application/usecase/...` to ensure all tests pass cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **User Story 1 (Phase 3)**: Independent, can be developed first.
- **User Story 2 (Phase 4)**: Independent, can be developed in parallel or after User Story 1.
- **Polish (Phase 5)**: Depends on completion of all stories.

### Parallel Opportunities

- T002 and T005 can start in parallel as they touch different files (`login_types.go` and `jwt_provider.go`).
- Phase 3 (US1) and Phase 4 (US2) can be worked on in parallel by different developers.
