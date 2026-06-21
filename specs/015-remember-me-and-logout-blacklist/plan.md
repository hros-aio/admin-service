# Implementation Plan: Remember Me and Logout Blacklist

**Branch**: `015-remember-me-and-logout-blacklist` | **Date**: 2026-06-21 | **Spec**: [spec.md](file:///home/ren0503/new-hros/admin-service/specs/015-remember-me-and-logout-blacklist/spec.md)

**Input**: Feature specification from `/specs/015-remember-me-and-logout-blacklist/spec.md`

## Summary

This feature adds session-length control and explicit access-token revocation to the authentication flows.
We will update `LoginUseCase` to process a `RememberMe` parameter, setting a 30-day session expiry if true, or a 24-hour browser-session expiry if false.
Additionally, we will update `LogoutUseCase` to parse the caller's JWT access token, extract its `jti` claim, calculate its remaining time-to-live (`ttl`), and register it in the `TokenBlacklist` cache (Redis) to prevent immediate token reuse.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: Echo Framework, golang-jwt/jwt/v5, go-redis/v9, Uber Fx

**Storage**: PostgreSQL (GORM) for persistent session tokens, Redis for the temporary token blacklist

**Testing**: Standard library testing with `testify/assert` and `testify/mock`

**Target Platform**: Linux server

**Project Type**: web-service

**Performance Goals**: Low latency cache checks (<5ms)

**Constraints**: Strict layer boundaries (UseCase must not import GORM or Redis direct clients; it must use defined repository and blacklist abstractions)

**Scale/Scope**: Update `LoginUseCase` and `LogoutUseCase` only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Clean Architecture & Strict Boundaries**: PASS. Usecases communicate with GORM/Redis exclusively through interfaces (`SessionTokenRepository`, `TokenBlacklist`).
- **Documentation-First**: PASS. Feature specification is written and finalized in `spec.md`.
- **Unit-Test-Per-File**: PASS. Both `login_usecase.go` and `logout_usecase.go` already have unit test files. These will be updated and must achieve 100% coverage.
- **Task-Driven & Atomic Implementation**: PASS. The implementation plan will generate specific, granular tasks to be executed sequentially.
- **Observability & Structured Logging**: PASS. Logging uses standard `log/slog` and will output non-sensitive execution details.

## Project Structure

### Documentation (this feature)

```text
specs/015-remember-me-and-logout-blacklist/
├── plan.md              # This file
├── research.md          # Decisions on token exp and JWT unverified parsing
├── data-model.md        # SessionToken expires_at and blacklist mappings
├── quickstart.md        # Scenario verification guide
├── contracts/
│   └── usecases.md      # Signature and DTO changes for the use cases
└── tasks.md             # Implementation tasks
```

### Source Code (repository root)

```text
internal/
├── application/
│   └── usecase/
│       ├── login_types.go                # Update LoginInput with RememberMe
│       ├── login_usecase.go              # Calculate dynamic expiry
│       ├── login_usecase_test.go         # Assert remember me behaviors
│       ├── logout_usecase.go             # Implement JTI blacklisting
│       └── logout_usecase_test.go        # Assert JTI extraction and blacklisting
```

**Structure Decision**: Single project, adapting the existing clean architecture directories in `internal/application/usecase/`.

## Complexity Tracking

*No violations found.*
