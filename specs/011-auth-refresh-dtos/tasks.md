# Tasks: Auth Refresh DTOs

**Input**: Design documents from `/specs/011-auth-refresh-dtos/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Unit tests for DTO validation mapping are required as per project constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Setup the feature plan and workspace state (already completed during spec/plan phases)

---

## Phase 2: Foundational (Blocking Prerequisites)

*(None required for this slice; base Echo DTO structure is already established)*

---

## Phase 3: User Story 1 - OpenAPI Definition for Token Refresh (Priority: P1) 🎯 MVP

**Goal**: Document the `/v1/auth/refresh` endpoint and its request/response schemas in the OpenAPI specification.

**Independent Test**: Verify that the OpenAPI contract `api/openapi.yaml` parses and validates successfully and documents the new `/v1/auth/refresh` path.

### Implementation for User Story 1

- [X] T002 [US1] Add the `/v1/auth/refresh` path and `RefreshRequest` schema component to `api/openapi.yaml`

**Checkpoint**: OpenAPI definitions are complete and valid.

---

## Phase 4: User Story 2 - Login Request remember_me Mappings (Priority: P1)

**Goal**: Verify that the `LoginRequest` DTO contains the boolean `RememberMe` field.

**Independent Test**: Verify that the `LoginRequest` struct contains the field and matches the OpenAPI definition.

### Implementation for User Story 2

- [X] T003 [P] [US2] Update `LoginRequest` struct in `internal/adapter/http/auth/dto/auth_dto.go` to ensure `RememberMe` field is correctly annotated

**Checkpoint**: `LoginRequest` maps the `remember_me` field.

---

## Phase 5: User Story 3 - Strict Request Validation DTO (Priority: P1)

**Goal**: Define the `RefreshRequest` DTO and create validation unit tests asserting constraints for `LoginRequest` and `RefreshRequest`.

**Independent Test**: Verify validation tag constraints pass/fail correctly under test cases in `internal/adapter/http/auth/dto/auth_dto_test.go`.

### Implementation for User Story 3

- [X] T004 [US3] Define the `RefreshRequest` struct in `internal/adapter/http/auth/dto/auth_dto.go` with required `refresh_token` validation tags
- [X] T005 [P] [US3] Create validation unit tests in `internal/adapter/http/auth/dto/auth_dto_test.go` checking validation constraints on DTO schemas

**Checkpoint**: Validation logic and tests are fully functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Format, lint, and validate changed files.

- [X] T006 Run formatting via `go fmt ./...` and linting via `golangci-lint run` on modified files
- [X] T007 [P] Run contract validation on `api/openapi.yaml` to ensure it passes syntax checks
- [X] T008 Run quickstart.md validation tests using `go test ./internal/adapter/http/auth/dto/...`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. Complete first.
- **User Stories (Phases 3-5)**: Depend on setup completion.
  - US1 (OpenAPI schema changes) is the MVP and should be completed first.
  - US2 and US3 can run in parallel with each other.
- **Polish (Phase 6)**: Depends on all user stories being complete.

---

## Parallel Example: User Stories 2 & 3

```bash
# Developer A implements US2 (remember_me mappings):
Task: "Update LoginRequest struct in internal/adapter/http/auth/dto/auth_dto.go to ensure RememberMe field is correctly annotated"

# Developer B implements US3 (strict DTO validation):
Task: "Define the RefreshRequest struct in internal/adapter/http/auth/dto/auth_dto.go with required refresh_token validation tags"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Implement OpenAPI path and components in `api/openapi.yaml` (T002).
2. Validate spec syntax.

### Incremental Delivery

1. Foundation ready (T001).
2. Implement and test US1 (T002).
3. Implement and test US2 (T003).
4. Implement and test US3 (T004, T005).
5. Format and validate the slice (T006, T007, T008).
