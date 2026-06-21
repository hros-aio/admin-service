# Tasks: Redis Token Blacklist

**Input**: Design documents from `/specs/012-redis-token-blacklist/`

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

*(None required for this cache infrastructure slice)*

---

## Phase 3: User Story 1 - Implement TokenBlacklist in Redis (Priority: P1) 🎯 MVP

**Goal**: Implement the `TokenBlacklist` interface in the Redis infrastructure layer, storing revoked token identifiers (such as JWT IDs) with accurate TTLs capped at 15 minutes.

**Independent Test**: Unit tests pass with `miniredis` asserting `Add` and `Exists` behaviors, including capping TTL at 15 minutes, key prefixing, and graceful connection failure handling.

### Implementation for User Story 1

- [x] T002 [US1] Create the infrastructure cache directory and implement `TokenBlacklist` interface in `internal/infrastructure/cache/token_blacklist_redis.go` using the go-redis client connection.
- [x] T003 [P] [US1] Create unit tests in `internal/infrastructure/cache/token_blacklist_redis_test.go` using a mocked Redis client (`miniredis`) to verify token storage, correct TTL setting, capping to 15m, existence checks, and graceful error handling on Redis failure.
- [x] T004 [US1] Wire the `NewRedisTokenBlacklist` provider in `internal/app/app.go` using Fx, so it is injected wherever the `TokenBlacklist` application interface is requested.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting, linting, and final validation.

- [x] T005 Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files.
- [x] T006 Run tests using `go test -v ./internal/infrastructure/cache/...` to verify the cache implementation passes cleanly.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Complete first.
- **User Story 1 (Phase 3)**: Depends on setup completion.
- **Polish (Phase 4)**: Depends on User Story 1 completion.
