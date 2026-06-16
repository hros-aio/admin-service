# Tasks: Authentication Service

**Input**: Design documents from `/specs/002-authentication-service/`

**Prerequisites**: Foundation Bootstrap complete.

**Tests**: Unit tests per file are MANDATORY.

---

## Phase 1: Database Foundation

- [X] **TSK-AUTH-001**: Create up/down SQL migration scripts for the `roles`, `role_permissions`, `admin_users`, and `session_tokens` tables. Ensure `email` has a UNIQUE index and `password_hash` supports bcrypt strings. `session_tokens` must have a UNIQUE index on `refresh_token` and foreign key to `admin_users.id`.
    - **Layer**: Migration
    - **Output**: `migrations/000002_create_auth_tables.up.sql` and `000002_create_auth_tables.down.sql`
    - **Acceptance**: Migrations execute successfully forward and backward.

- [ ] **TSK-AUTH-002**: Seed initial system roles (`Super Admin`, `Manager`, `Auditor`, etc.) and their default permissions.
    - **Layer**: Migration
    - **Output**: `migrations/000003_seed_roles.up.sql`
    - **Acceptance**: Roles and permissions are present in the database after migration.

---

## Phase 2: Domain Layer

- [ ] **TSK-AUTH-003**: Define Auth domain entities (`AdminUser`, `Role`, `SessionToken`) and repository interfaces in `internal/domain/auth/`.
- [ ] **TSK-AUTH-004**: Implement password hashing and verification utility in `internal/domain/auth/password.go`.

---

## Phase 3: Infrastructure Layer

- [ ] **TSK-AUTH-005**: Implement GORM models and repository implementation for Auth entities in `internal/infrastructure/repository/auth/`.
- [ ] **TSK-AUTH-006**: Write integration tests for Auth repositories.

---

## Phase 4: Application Layer

- [ ] **TSK-AUTH-007**: Implement `AuthService` with `Login` use case.
- [ ] **TSK-AUTH-008**: Implement `AuthService` with `Refresh` and `Logout` use cases.
- [ ] **TSK-AUTH-009**: Implement JWT generator and validator.

---

## Phase 5: Adapter Layer

- [ ] **TSK-AUTH-010**: Implement `AuthHandler` for `/v1/auth/login`, `/refresh`, and `/logout`.
- [ ] **TSK-AUTH-011**: Update OpenAPI contract with Auth endpoints.
- [ ] **TSK-AUTH-012**: Implement Auth middleware for Echo to protect routes.
- [ ] **TSK-AUTH-013**: Wire Auth module into Fx lifecycle.
