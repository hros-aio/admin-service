# Tasks: Biometric Device Login (WebAuthn) - Handler Layer (TSK-BIO-007)

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

- [x] T001 [US2] Update `AuditLogger` interface in `internal/domain/auth/audit.go` to include `LogBiometricSuccess(ctx context.Context, event events.BiometricSuccessEvent)`
- [x] T002 [US2] Implement `LogBiometricSuccess` in `SlogAuditLogger` in `internal/infrastructure/auth/audit_logger.go`
- [x] T003 [US2] Update the `mockAuditLogger` and `mockAcceptInviteAuditLogger` structs in `internal/application/usecase/login_usecase_test.go`, `internal/application/usecase/accept_invite_usecase_test.go`, and `internal/adapter/http/auth_handler_test.go` to mock the new audit method and satisfy interface constraints
- [x] T004 [US2] Implement `VerifyBiometricUseCase` and `VerifyBiometricInput` in `internal/application/usecase/verify_biometric_usecase.go`
- [x] T005 [P] [US2] Create unit tests in `internal/application/usecase/verify_biometric_usecase_test.go` with 100% coverage using Go mocks
- [x] T006 [US2] Register `VerifyBiometricUseCase` inside Fx module in `internal/application/module.go`
- [x] T007 Run `go fmt` and `go test` for all affected packages (`internal/domain/...`, `internal/infrastructure/...`, `internal/application/...`, `internal/adapter/...`) and verify all tests pass

---

## Phase 3: User Story 3 - Biometric Handlers & API Routing (Priority: P1)

**Goal**: Implement Echo HTTP handlers for biometric challenge generation and verification, map domain errors to HTTP statuses, serialize JWT response, and register routing in Echo via Fx.

**Independent Test**: Verify using Echo HTTP test recorder that `POST /v1/auth/biometric/challenge` and `POST /v1/auth/biometric/verify` return 200 OK with correct payloads, invalid inputs return 400 Bad Request, and business verification failures map to 401 Unauthorized.

### Implementation for User Story 3

- [x] T008 [US3] Update DTO definitions in `internal/adapter/http/auth/dto/auth_dto.go` and update schemas in `api/openapi.yaml` to include `credential_id` in challenge response and `remember_me` in verify request
- [x] T009 [US3] Implement `AuthBiometricHandler` and error mapping in `internal/adapter/http/auth_biometric_handler.go`
- [x] T010 [P] [US3] Create handler unit tests in `internal/adapter/http/auth_biometric_handler_test.go` with 100% coverage
- [x] T011 [US3] Register `AuthBiometricHandler` inside Echo Fx module in `internal/adapter/http/module.go` and configure endpoint routing in Echo

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and overall testing check

- [x] T012 Run `go fmt` and `go test -count=1 ./...` for all affected packages and verify all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)** and **UseCase (Phase 3)**: Must be complete to supply UseCases to the handler layer.
- **Biometric Handlers (Phase 3)**: Implements endpoint routing and binding.
- **Polish (Phase 4)**: Final verification step.

---

## Implementation Strategy

### MVP First (User Story 3 Only)

1. Update the DTOs and OpenAPI specs to support necessary fields (credential ID, remember me).
2. Create `AuthBiometricHandler` with Challenge and Verify handlers translating Echo context to UseCase calls.
3. Hook handler routing up in Fx.
4. Exhaustively test with `auth_biometric_handler_test.go`.
5. Run formatting and full test validation.
