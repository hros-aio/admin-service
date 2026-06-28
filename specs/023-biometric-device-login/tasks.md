# Tasks: Biometric Device Login (WebAuthn)

**Input**: Design documents from `/specs/023-biometric-device-login/`

**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `internal/`, `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 [P] Ensure specs files are created and checked in

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*(None required for domain definitions)*

## Phase 3: User Story 1 - Biometric Domain Definition (Priority: P1) 🎯 MVP

**Goal**: Define the WebAuthn challenge cache interface, domain errors, and biometric success event.

**Independent Test**: Verification that the domain and application layers compile successfully and all unit tests pass with zero external dependencies.

### Implementation for User Story 1

- [x] T002 [P] [US1] Define specific domain errors in `internal/domain/errors/auth_errors.go`
- [x] T003 [P] [US1] Define the event payload struct in `internal/domain/events/auth_events.go`
- [x] T004 [P] [US1] Define the WebAuthnChallengeCache interface in `internal/application/interfaces/webauthn_cache.go`
- [x] T005 [P] [US1] Create unit tests for domain errors in `internal/domain/errors/auth_errors_test.go`
- [x] T006 [P] [US1] Create unit tests for event payload serialization in `internal/domain/events/auth_events_test.go`
- [x] T007 [P] [US1] Create a mock/stub check to verify compilation and correctness of `internal/application/interfaces/webauthn_cache_test.go`

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T008 [P] Run go formatting and all tests to confirm zero regressions

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **User Story 1 (Phase 3)**: Depends on Phase 1 Setup.

### Parallel Opportunities

- Tasks T002, T003, and T004 can be started in parallel as they define independent types/interfaces.
- Tests T005, T006, and T007 can be written in parallel.

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Implement Domain Errors, Event structs, and Challenge Cache interface.
2. Add corresponding unit tests.
3. Validate and verify.
