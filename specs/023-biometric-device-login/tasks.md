# Tasks: Biometric Device Login (WebAuthn) - UseCase Layer (TSK-BIO-005)

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

*(No setup tasks required; using existing infrastructure)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*(No foundational tasks required)*

---

## Phase 3: User Story 2 - Biometric Login (Priority: P1)

**Goal**: Implement `GenerateBiometricChallengeUseCase` to generate and cache WebAuthn login challenges.

**Independent Test**: Verify using mocked repositories and caches that the usecase correctly fetches the user, parses the JSONB credential, generates a cryptographically secure challenge, caches it, and returns the expected string.

### Implementation for User Story 2

- [x] T001 [US2] Implement `GenerateBiometricChallengeUseCase` and input/output structs in `internal/application/usecase/generate_biometric_challenge_usecase.go`
- [x] T002 [P] [US2] Create unit tests in `internal/application/usecase/generate_biometric_challenge_usecase_test.go` with 100% coverage using Go mock repositories
- [x] T003 [US2] Register `GenerateBiometricChallengeUseCase` inside Fx module in `internal/application/module.go`

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and overall testing check

- [x] T004 Run `go fmt` and `go test` for `internal/application/...` to verify compilation and test correctness

---

## Dependencies & Execution Order

### Phase Dependencies

- **User Story 2 (Phase 3)**: Can start immediately.
- **Polish (Phase 4)**: Depends on User Story 2 being complete.

### Parallel Opportunities

- Unit tests (T002) and implementation (T001) can be worked on concurrently by stubbing the interface first.

---

## Implementation Strategy

### MVP First (User Story 2 Only)

1. Implement UseCase and structs.
2. Register UseCase in Fx.
3. Write mock-based unit tests.
4. Run formatting and verify all tests pass.
