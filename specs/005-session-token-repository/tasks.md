# Tasks: Session Token Repository (GORM)

**Input**: Design documents from `/specs/005-session-token-repository/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Unit tests are MANDATORY per Constitution Principle III. Integration tests for infrastructure boundaries are required per tech-stack.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Update domain interfaces and shared mappers

- [X] T00- [ ] T001 Update `SessionTokenRepository` interface in `internal/domain/session_token.go` to include `DeleteByToken(ctx context.Context, token string) error`
- [X] T00- [ ] T002 Update `mapper.go` in `internal/infrastructure/repository/auth/mapper.go` to include `SessionToken` to/from domain mappers
- [X] T00- [ ] T003 Define `SessionTokenModel` in `internal/infrastructure/repository/auth/model.go` with GORM tags per `data-model.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core implementation file and Fx registration

- [X] T00- [ ] T004 Create `session_token_repository.go` in `internal/infrastructure/repository/auth/` with struct definition and constructor
- [X] T00- [ ] T005 [P] Register `NewGormSessionTokenRepository` in the auth infrastructure Fx module (if applicable, otherwise ensure it's providable)

**Checkpoint**: Infrastructure scaffold ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Store Session Token on Login (Priority: P1) 🎯 MVP

**Goal**: Persist a new session token to the database using GORM

**Independent Test**: `go test -v internal/infrastructure/repository/auth/session_token_repository_test.go -run TestSave`

### Implementation for User Story 1

- [X] T00- [ ] T006 [US1] Implement `Save(ctx context.Context, token *domain.SessionToken)` in `internal/infrastructure/repository/auth/session_token_repository.go`
- [X] T00- [ ] T007 [US1] Utilize `platformDB.GetTx(ctx, r.db)` in `Save` method for transaction support
- [X] T00- [ ] T008 [P] [US1] Create `session_token_repository_test.go` and implement `TestSave_Success` using `sqlmock`
- [X] T00- [ ] T009 [P] [US1] Implement `TestSave_Error` in `internal/infrastructure/repository/auth/session_token_repository_test.go` to verify error mapping

**Checkpoint**: User Story 1 (Persistence) is fully functional and tested independently

---

## Phase 4: User Story 2 - Remove Session Token on Explicit Logout (Priority: P2)

**Goal**: Remove a specific session token by value from the database

**Independent Test**: `go test -v internal/infrastructure/repository/auth/session_token_repository_test.go -run TestDeleteByToken`

### Implementation for User Story 2

- [X] T01- [ ] T010 [US2] Implement `DeleteByToken(ctx context.Context, token string)` in `internal/infrastructure/repository/auth/session_token_repository.go`
- [X] T01- [ ] T011 [US2] Ensure `DeleteByToken` is idempotent (no error if token not found)
- [X] T01- [ ] T012 [P] [US2] Implement `TestDeleteByToken_Success` in `internal/infrastructure/repository/auth/session_token_repository_test.go` using `sqlmock`
- [X] T01- [ ] T013 [P] [US2] Implement `TestDeleteByToken_NotFound` in `internal/infrastructure/repository/auth/session_token_repository_test.go` to verify idempotency

**Checkpoint**: User Story 2 (Deletion) is fully functional and tested independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Integration testing and final validation

- [ ] T014 [P] Create integration test in `test/integration/session_token_repository_test.go` using `testcontainers` (if environment supports it)
- [X] T01- [ ] T015 Verify 100% unit test coverage for `internal/infrastructure/repository/auth/session_token_repository.go`
- [X] T01- [ ] T016 [P] Run `golangci-lint run internal/infrastructure/repository/auth/...`
- [X] T01- [ ] T017 [P] Execute final validation scenarios from `quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Base interface and models must exist
- **Foundational (Phase 2)**: Depends on Setup - provides the implementation file
- **User Stories (Phase 3-4)**: Depend on Foundational phase. Can be worked on in parallel once `session_token_repository.go` exists.
- **Polish (Phase 5)**: Depends on both user stories being complete.

### User Story Dependencies

- **US1**: Independent after Phase 2
- **US2**: Independent after Phase 2

### Parallel Opportunities

- T002 and T003 can be done in parallel
- T008 and T009 (tests for US1) can be done in parallel with implementation T006-T007
- T012 and T013 (tests for US2) can be done in parallel with implementation T010-T011
- Integration tests (T014) can be prepared in parallel with unit tests

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Update domain interface (T001)
2. Setup infrastructure models/mappers (T002, T003)
3. Implement `Save` method (T006, T007)
4. Verify with unit tests (T008, T009)

### Incremental Delivery

1. Foundation + US1 → "Login storage ready"
2. Add US2 → "Logout invalidation ready"
3. Integration testing → "Production ready"

---

## Notes

- All GORM operations must use `ctx` for cancellation
- Repository must NOT return `gorm.DB` or models
- ID generation should be handled by the domain or mapping layer (if UUIDs are generated in DB, ensure they are retrieved)
- `session_tokens` table is assumed to have unique index on `refresh_token` per `data-model.md`
