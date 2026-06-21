# Tasks: Brute-Force Lockout Defense

**Input**: Design documents from `/specs/018-brute-force-lockout-defense/`

**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: Includes unit tests in `internal/domain/errors/`, `internal/domain/events/`, and `internal/application/interfaces/`.

**Organization**: Tasks are grouped by logical implementation phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

- [x] T001 Setup the feature spec, plan, and checklist (already completed)

---

## Phase 2: Foundational Domain & Interface Definitions (TSK-AUTH-018)

**Purpose**: Establish domain layer primitives and interfaces for the lockout mechanism.

- [x] T002 [P] [US1] Define the `BruteForceCache` interface in `internal/application/interfaces/brute_force_cache.go` containing methods to check/increment failed login attempts, set/check lockout status, and reset lockout state.
- [x] T003 [P] [US1] Define the domain error `ErrAccountLocked` in `internal/domain/errors/auth_errors.go`.
- [x] T004 [P] [US1] Define event payload structs for `account.locked` (audit) and `email.send` (notification) in `internal/domain/events/auth_events.go`.
- [x] T005 [P] [US1] Implement unit tests verifying formatting, serialization, error checks, and interface completeness in:
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
  - `internal/application/interfaces/brute_force_cache_test.go` (if needed, or verify compile correctness)

**Checkpoint**: Interface compiles, domain errors/events defined, and unit tests pass with zero external infrastructure dependencies.

---

## Phase 2.5: Infrastructure Cache Implementation (TSK-AUTH-019)

**Purpose**: Implement the `BruteForceCache` interface using Redis.

- [x] T008 [P] [US1] Implement the `BruteForceCache` interface using Redis in `internal/infrastructure/cache/brute_force_redis.go` using `auth:failed_attempts:{email}` (15-min TTL) and `auth:lockout:{email}` (30-min TTL).
- [x] T009 [P] [US1] Implement unit tests verifying Redis caching logic and graceful degradation in `internal/infrastructure/cache/brute_force_redis_test.go` using an in-memory Redis server (miniredis).

**Checkpoint**: Redis cache safely tracks attempts and lockout states with exact TTLs, and tests pass with graceful degradation.

---

## Phase 3: Polish

- [x] T006 Run formatting (`go fmt ./...`) and linting (`golangci-lint run`) on modified files.
- [x] T007 Run all unit tests to verify zero regressions.
