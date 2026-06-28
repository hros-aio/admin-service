# Tasks: Biometric Device Login (WebAuthn) - Cache Layer

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

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Domain definition and basic error structure

- [x] T001 [P] Define `ErrChallengeNotFoundOrExpired` domain error in `internal/domain/errors/auth_errors.go` and add `VerifyAndConsumeChallenge` method to `internal/application/interfaces/webauthn_cache.go`

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core interface stubs that must be complete before implementation begins

- [x] T002 Update `fakeWebAuthnChallengeCache` stub in `internal/application/interfaces/webauthn_cache_test.go` to implement `VerifyAndConsumeChallenge` and verify compilation

## Phase 3: User Story 1 - WebAuthn Redis Cache Implementation (Priority: P1) 🎯 MVP

**Goal**: Implement the Redis-backed challenge cache with a 5-minute TTL and atomic Lua script verification.

**Independent Test**: Verify that all Cache unit tests pass with `miniredis` stubbing.

### Implementation for User Story 1

- [x] T003 [P] [US1] Create `internal/infrastructure/cache/webauthn_redis.go` implementing `WebAuthnChallengeCache` interface using Go-Redis and Lua script
- [x] T004 [P] [US1] Create unit tests in `internal/infrastructure/cache/webauthn_redis_test.go` using `miniredis` to verify storage, TTL, deletion, and atomic verification

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and overall testing check

- [x] T005 Run `go fmt` and `go test` for both `internal/infrastructure/cache/...` and `internal/application/interfaces/...` (exercising interface tests modified in T002) to confirm compilation and test correctness

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1)
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2)
- **Polish (Phase 4)**: Depends on User Story 1 (Phase 3)

### Parallel Opportunities

- Tasks T003 and T004 can be implemented concurrently by preparing interface stubs first.

---

## Parallel Example: User Story 1

```bash
# Verify compilation and unit tests:
go test ./internal/application/interfaces/...
go test ./internal/infrastructure/cache/...
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Define error and updated interface.
2. Update interface compilation stubs.
3. Write Redis implementation and corresponding unit tests.
4. Verify all tests pass.
