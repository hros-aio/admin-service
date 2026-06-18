# Tasks: Admin Login Use Case

**Input**: Design documents from `/specs/006-admin-login-usecase/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Unit tests per file are MANDATORY as per Constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and domain interface definition

- [x] T001 Define `AuditLogger` interface in `internal/domain/auth/audit.go`
- [x] T002 Define `PasswordHelper` interface in `internal/application/auth/password_helper.go`
- [x] T003 Define `TokenProvider` interface in `internal/application/auth/token_provider.go`
- [x] T004 [P] Update `internal/domain/auth/repository.go` to include `AdminUserRepository` and `SessionTokenRepository` (if not already fully aligned with spec)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T005 Implement bcrypt-based `PasswordHelper` in `internal/infrastructure/auth/bcrypt_helper.go` (include `Compare` and `CompareDummy`)
- [x] T006 [P] Implement RS256-based `TokenProvider` in `internal/infrastructure/auth/jwt_provider.go`
- [x] T007 [P] Implement `AuditLogger` implementation (e.g., logging to slog) in `internal/infrastructure/auth/audit_logger.go`

**Checkpoint**: Foundation ready - user story implementation can now begin.

---

## Phase 3: User Story 1 - Secure Portal Access (Priority: P1) 🎯 MVP

**Goal**: Successfully authenticate an active admin user with valid credentials and issue tokens.

**Independent Test**: Execute `LoginUseCase` with valid credentials and verify it returns a valid JWT and persists a session.

### Implementation for User Story 1

- [x] T008 [US1] Define `LoginInput` and `LoginOutput` structs in `internal/application/usecase/login_types.go`
- [x] T009 [US1] Create `LoginUseCase` structure and constructor in `internal/application/usecase/login_usecase.go`
- [x] T010 [US1] Implement `Execute` method logic for fetching user, comparing password, and generating tokens in `internal/application/usecase/login_usecase.go`
- [x] T011 [US1] Implement session token persistence in `LoginUseCase.Execute` using `SessionTokenRepository` in `internal/application/usecase/login_usecase.go`
- [x] T012 [US1] Add `login.success` audit logging in `LoginUseCase.Execute` in `internal/application/usecase/login_usecase.go`
- [x] T013 [US1] Create unit tests for success scenario in `internal/application/usecase/login_usecase_test.go`

**Checkpoint**: User Story 1 is functional for happy-path logins.

---

## Phase 4: User Story 2 - Failed Login Defense (Priority: P2)

**Goal**: Protect against timing oracle attacks and handle failed authentication securely.

**Independent Test**: Verify constant-time response for non-existent users and correct emission of failure events.

### Implementation for User Story 2

- [x] T014 [US2] Implement constant-time dummy password comparison for missing users in `internal/application/usecase/login_usecase.go`
- [x] T015 [US2] Implement `login.failed` audit logging for credential failures in `internal/application/usecase/login_usecase.go`
- [x] T016 [US2] Add unit tests for invalid email scenario (timing attack defense) in `internal/application/usecase/login_usecase_test.go`
- [x] T017 [US2] Add unit tests for invalid password scenario in `internal/application/usecase/login_usecase_test.go`

**Checkpoint**: User Story 2 ensures security against timing analysis and handles failures.

---

## Phase 5: User Story 3 - Session Integrity (Priority: P3)

**Goal**: Ensure all login attempts are recorded for auditing and compliance.

**Independent Test**: Verify that every call to the use case results in an audit log entry.

### Implementation for User Story 3

- [x] T018 [US3] Ensure comprehensive error handling and logging for all edge cases (account locked, inactive) in `internal/application/usecase/login_usecase.go`
- [x] T019 [US3] Add unit tests for locked and inactive account scenarios in `internal/application/usecase/login_usecase_test.go`

**Checkpoint**: All login scenarios are audited and tested.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final integration and validation.

- [x] T020 [P] Register `LoginUseCase` in Fx module at `internal/application/module.go`
- [x] T021 [P] Run `quickstart.md` validation scenarios using `go test`
- [x] T022 [P] Update documentation in `docs/architecture/SRS.md` if necessary to reflect implementation details

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1 interfaces.
- **User Stories (Phase 3+)**: All depend on Foundational phase completion.
- **Polish (Phase 6)**: Depends on completion of User Stories.

### Parallel Opportunities

- T004 (Repository update) can run in parallel with interface definitions (T001-T003).
- Foundational implementations (T005, T006, T007) can run in parallel.
- US1 (Phase 3) can be worked on as a single thread.
- US2 and US3 add specific security and auditing layers to the core logic.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 & 2.
2. Complete Phase 3 (US1).
3. Validate happy path login works end-to-end.

### Incremental Delivery

1. Foundation ready (Phase 1-2).
2. Login functional (Phase 3).
3. Security hardened (Phase 4).
4. Fully audited and compliant (Phase 5).
