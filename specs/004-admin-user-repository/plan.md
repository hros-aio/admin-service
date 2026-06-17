# Implementation Plan: Admin User Repository (Fetch by Email)

**Branch**: `004-admin-user-repo` | **Date**: 2026-06-17 | **Spec**: [specs/004-admin-user-repository/spec.md](specs/004-admin-user-repository/spec.md)

**Input**: Feature specification from `/specs/004-admin-user-repository/spec.md`

## Summary
Implement `AdminUserRepository` using GORM to fetch users by email. The implementation will follow the **Repository Structure** rule, residing in `internal/infrastructure/repository/auth/` and split into `model.go`, `mapper.go`, and `repository.go`. It will map `gorm.ErrRecordNotFound` to the domain-level `ErrUserNotFound` and use the project's transaction management pattern.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: GORM, sqlmock, Uber Fx

**Storage**: PostgreSQL

**Testing**: `go test` with table-driven tests and `sqlmock`

**Target Platform**: Linux server

**Project Type**: Web Service (Backend)

**Performance Goals**: Indexed relational query (< 50ms p95)

**Constraints**: Strict Clean Architecture boundaries; zero GORM leakage into Domain or Application layers.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Clean Architecture**: Implementation stays in infrastructure layer; maps to domain entities.
- [x] **II. Documentation-First**: Plan, Research, and Data Model documents created before implementation.
- [x] **III. Unit-Test-Per-File**: `repository_test.go` planned with `sqlmock`.
- [x] **IV. Task-Driven**: Plan scoped strictly to repository implementation (TSK-AUTH-004).
- [x] **V. Observability**: Database errors will be logged via structured logging where appropriate (handled by platform/middleware).

## Project Structure

### Documentation (this feature)

```text
specs/004-admin-user-repository/
├── plan.md              # This file
├── research.md          # Research findings and decisions
├── data-model.md        # Entity and GORM model mapping
├── quickstart.md        # Validation scenarios
└── tasks.md             # Implementation tasks (Phase 2)
```

### Source Code (repository root)

```text
internal/
├── domain/
│   └── admin_user.go        # Existing entity and interface
└── infrastructure/
    └── repository/
        └── auth/
            ├── model.go     # GORM model
            ├── mapper.go    # GORM to Domain mapping
            ├── repository.go # Implementation
            └── repository_test.go # Unit tests
```

**Structure Decision**: Single project structure following Clean Architecture and the specialized "Repository Structure" for the Auth module.

## Complexity Tracking

*No violations identified.*
