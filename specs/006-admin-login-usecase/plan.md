# Implementation Plan: Admin Login Use Case

**Branch**: `006-admin-login-usecase` | **Date**: 2026-06-18 | **Spec**: [specs/006-admin-login-usecase/spec.md](specs/006-admin-login-usecase/spec.md)

**Input**: Feature specification from `/specs/006-admin-login-usecase/spec.md`

## Summary
Implement the `LoginUseCase` to handle secure administrator authentication. This includes fetching users by email, bcrypt password verification with constant-time protection, RS256 JWT issuance, session persistence, and audit logging.

## Technical Context

**Language/Version**: Go 1.26.1

**Primary Dependencies**: 
- `golang.org/x/crypto/bcrypt` (Password hashing)
- `github.com/golang-jwt/jwt/v5` (JWT generation)
- `gorm.io/gorm` (Database access)
- `go.uber.org/fx` (Dependency Injection)

**Storage**: PostgreSQL (AdminUser and SessionToken tables)

**Testing**: `github.com/stretchr/testify` (Unit tests with mocks)

**Target Platform**: Linux Server / Containerized

**Project Type**: Web Service / Clean Architecture

**Performance Goals**: < 500ms for login (limited by bcrypt cost factor 12)

**Constraints**: 
- Constant-time processing for invalid emails.
- RS256 signing for JWTs.
- 15-minute access token expiry.

**Scale/Scope**: Auth core logic for HROS Admin Portal.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Boundary Rule**: UseCase will only depend on domain interfaces.
- [x] **Test Rule**: `login_usecase_test.go` will accompany the implementation.
- [x] **Logic Rule**: No business logic in handlers; no GORM/Echo in application layer.
- [x] **Audit Rule**: All login attempts will be logged to an audit interface.

## Project Structure

### Documentation (this feature)

```text
specs/006-admin-login-usecase/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan
├── research.md          # Research findings (bcrypt, JWT, Audit)
├── data-model.md        # Domain entities and repository interfaces
├── quickstart.md        # Validation guide
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code (repository root)

```text
internal/
├── domain/
│   ├── auth/
│   │   ├── admin_user.go       # Existing
│   │   ├── session_token.go    # Existing
│   │   ├── repository.go       # AdminUserRepository + SessionTokenRepository
│   │   └── audit.go            # AuditLogger interface
│   └── errors/
│       └── auth_errors.go      # Existing errors
├── application/
│   ├── usecase/
│   │   ├── login_usecase.go    # TO BE CREATED
│   │   ├── login_usecase_test.go # TO BE CREATED
│   │   ├── logout_usecase.go   # TO BE CREATED (TSK-AUTH-007)
│   │   └── logout_usecase_test.go # TO BE CREATED (TSK-AUTH-007)
│   └── auth/
│       ├── token_provider.go   # JWT interface
│       └── password_helper.go  # Bcrypt interface
└── adapter/
    └── http/
        └── auth/
            └── dto/
                └── auth_dto.go # Existing
```

**Structure Decision**: Standard Clean Architecture layout. UseCases are placed in `internal/application/usecase` as requested by the user.

## Complexity Tracking

*No violations identified.*
