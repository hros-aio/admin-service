# Tasks: Authentication DTOs and OpenAPI

**Input**: Design documents from `/specs/003-auth-dtos-openapi/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Unit tests for DTO validation mapping are required as per project constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create `api/` directory at repository root
- [X] T002 Migrate `docs/openapi/openapi.yaml` to `api/openapi.yaml` to align with project constitution
- [X] T003 Create directory `internal/adapter/http/auth/dto/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure prerequisites

- [X] T004 [P] Update `api/openapi.yaml` with standard `ErrorResponse` component if missing (from foundation docs)

---

## Phase 3: User Story 1 - Admin Login (Priority: P1) 🎯 MVP

**Goal**: Define the DTOs and OpenAPI contract for the login endpoint.

**Independent Test**: Verify `api/openapi.yaml` contains `POST /v1/auth/login` and `LoginRequest`/`LoginResponse` schemas.

### Implementation for User Story 1

- [X] T005 [P] [US1] Define `LoginRequest` and `LoginResponse` structs in `internal/adapter/http/auth/dto/auth_dto.go`
- [X] T006 [P] [US1] Add `validate` tags to `LoginRequest` fields in `internal/adapter/http/auth/dto/auth_dto.go`
- [X] T007 [US1] Create unit tests for DTO validation tags in `internal/adapter/http/auth/dto/auth_dto_test.go`
- [X] T008 [US1] Add `POST /v1/auth/login` path definition to `api/openapi.yaml`
- [X] T009 [US1] Add `LoginRequest` and `LoginResponse` schemas to `components/schemas` in `api/openapi.yaml`

**Checkpoint**: User Story 1 contract and DTOs are complete and verified.

---

## Phase 4: User Story 2 - Terminate Session (Priority: P1)

**Goal**: Define the OpenAPI contract for the session termination endpoint.

**Independent Test**: Verify `api/openapi.yaml` contains `DELETE /v1/auth/session` with 204 No Content response.

### Implementation for User Story 2

- [X] T010 [US2] Add `DELETE /v1/auth/session` path definition to `api/openapi.yaml`
- [X] T011 [US2] Document `401 Unauthorized` error response for `DELETE /v1/auth/session` using `$ref` to `ErrorResponse`

**Checkpoint**: User Story 2 contract is complete and verified.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T012 [P] Update `docs/openapi/index.html` (if exists) to point to `api/openapi.yaml`
- [X] T013 Run `golangci-lint` to ensure DTO code follows project standards
- [X] T014 [P] Final validation of `api/openapi.yaml` against OpenAPI spec
- [X] T015 Run quickstart.md validation scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup (T002).
- **User Stories (Phase 3 & 4)**: Depend on Foundational (T004). US1 and US2 are independent.
- **Polish (Phase 5)**: Depends on all user stories completion.

### Parallel Opportunities

- T005 and T006 can be done together as they affect the same file.
- T008 and T010 can be done in parallel once the foundation is set.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Setup and Foundational phases.
2. Implement User Story 1 (Login DTOs and OpenAPI).
3. Validate US1 against documentation requirements.

### Incremental Delivery

1. Foundation ready (Phase 1 & 2).
2. Login contract ready (Phase 3).
3. Session termination contract ready (Phase 4).
4. Full contract polish (Phase 5).
