# Tasks: Biometric Device Login (WebAuthn) - UseCase Layer (TSK-BIO-006)

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

---

## Phase 3: User Story 2 - Biometric Login (Priority: P1)

**Goal**: Implement `VerifyBiometricUseCase` to verify signatures, persist sign count updates, and issue session tokens.

**Independent Test**: Verify using mocked repositories and caches that the usecase correctly verifies FIDO2 ECDSA signatures, increments sign count atomically, logs success audit events, and issues session tokens.

### Implementation for User Story 2

- [x] T004 [US2] Implement `VerifyBiometricUseCase` and `VerifyBiometricInput` in `internal/application/usecase/verify_biometric_usecase.go`
- [x] T005 [P] [US2] Create unit tests in `internal/application/usecase/verify_biometric_usecase_test.go` with 100% coverage using Go mocks
- [x] T006 [US2] Register `VerifyBiometricUseCase` inside Fx module in `internal/application/module.go`

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Formatting and overall testing check

- [x] T007 Run `go fmt` and `go test` for all affected packages (`internal/domain/...`, `internal/infrastructure/...`, `internal/application/...`, `internal/adapter/...`) and verify all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: Must be implemented first to allow compilation of packages with the new audit logging capabilities.
- **User Story 2 (Phase 3)**: Depends on Phase 2.
- **Polish (Phase 4)**: Depends on User Story 2 implementation and tests being complete.

---

## Implementation Strategy

### MVP First (User Story 2 Only)

1. Extend `AuditLogger` interface and update mock definitions so the project compiles.
2. Implement cryptographic verification logic inside `VerifyBiometricUseCase`.
3. Register the new UseCase in Fx.
4. Write comprehensive unit tests including cryptographic success and validation failure cases.
5. Verify that all project tests pass.
