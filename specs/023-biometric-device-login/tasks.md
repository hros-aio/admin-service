# Tasks: Biometric Device Login (WebAuthn) - Repository Layer (TSK-BIO-004)

**Input**: Design documents from `/specs/023-biometric-device-login/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `internal/`, `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

*(No setup tasks required for this repository update; using existing infrastructure)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*(No foundational schema migrations or routing setup required; `webauthn_credentials` JSONB column already exists)*

---

## Phase 3: User Story 2 - Biometric Login (Priority: P1)

**Goal**: Implement `UpdateWebAuthnSignCount` repository method to persist WebAuthn sign counts and mitigate authenticator cloning attacks.

**Independent Test**: Verify using `sqlmock` that the GORM query correctly performs a PostgreSQL `jsonb_set` atomic update on the `webauthn_credentials` column.

### Implementation for User Story 2

- [x] T001 [US2] Define `UpdateWebAuthnSignCount(ctx context.Context, adminID string, newCount uint32) error` in `AdminUserRepository` interface in `internal/domain/admin_user.go`
- [x] T002 [US2] Implement `UpdateWebAuthnSignCount(ctx context.Context, adminID string, newCount uint32) error` method in `GormAdminUserRepository` inside `internal/infrastructure/repository/auth/repository.go`
- [x] T003 [P] [US2] Create unit tests in `internal/infrastructure/repository/auth/repository_test.go` using `sqlmock` to verify the GORM query structure, parameters, rows affected matching, and database error handling for `UpdateWebAuthnSignCount`

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and overall testing check

- [x] T004 Run `go fmt` and `go test` for all affected packages, including `internal/infrastructure/repository/auth/...`, `internal/adapter/http/auth_handler_test.go`, `internal/application/usecase/login_usecase_test.go`, and `internal/application/usecase/accept_invite_usecase_test.go` to confirm mock compilation and test correctness

---

## Dependencies & Execution Order

### Phase Dependencies

- **User Story 2 (Phase 3)**: Can start immediately since database and code frameworks are already set up.
- **Polish (Phase 4)**: Depends on User Story 2 implementation and tests being complete.

### Parallel Opportunities

- Unit tests (T003) and implementation (T002) can be drafted concurrently after interface definition (T001) is complete.

---

## Implementation Strategy

### MVP First (User Story 2 Only)

1. Define interface method.
2. Write GORM implementation.
3. Write `sqlmock` unit tests.
4. Verify all tests pass.
