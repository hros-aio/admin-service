# Tasks: Session Token Repository Updates

**Input**: Design documents from `/specs/013-session-token-repository/`

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

*(None required for this database repository slice)*

---

## Phase 3: User Story 1 - Session Token DB Operations (Priority: P1) 🎯 MVP

**Goal**: Update the `SessionTokenRepository` domain interface and its GORM infrastructure implementation with the `UpdateToken` method, and write unit tests for both `FindByToken` and `UpdateToken` using `sqlmock`.

**Independent Test**: Unit tests pass with `sqlmock` confirming GORM correctly translates `UpdateToken` into SQL UPDATE statements.

### Implementation for User Story 1

- [x] T002 [US1] Update the `SessionTokenRepository` interface in `internal/domain/session_token.go` to include `UpdateToken(ctx context.Context, session *SessionToken) error`.
- [x] T003 [US1] Implement `UpdateToken(ctx context.Context, session *domain.SessionToken) error` in `internal/infrastructure/repository/auth/session_token_repository.go` using GORM's `Save` method.
- [x] T004 [US1] Add unit tests for `FindByToken` and `UpdateToken` in `internal/infrastructure/repository/auth/session_token_repository_test.go` using `sqlmock` to verify successful operations and error handling.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting, linting, and final validation.

- [x] T005 Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T006 Run tests using `go test -v ./internal/infrastructure/repository/auth/...` to verify the repository implementation passes cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Complete first.
- **User Story 1 (Phase 3)**: Depends on setup completion.
- **Polish (Phase 4)**: Depends on User Story 1 completion.
