# Implementation Plan: Redis Token Blacklist

**Branch**: `master` | **Date**: 2026-06-21 | **Spec**: [specs/012-redis-token-blacklist/spec.md](spec.md)

**Input**: Feature specification from `/specs/012-redis-token-blacklist/spec.md`

## Summary

Implement the `TokenBlacklist` interface in the Redis infrastructure layer. This implementation will store revoked JWT access/refresh token identifiers (e.g. JTIs) in Redis with an accurate TTL matching the remaining token lifetime (capped at 15 minutes max). Graceful degradation handles Redis connection errors safely to prevent blocking application flows.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: `github.com/redis/go-redis/v9`, `go.uber.org/fx`, `log/slog`, `github.com/alicebob/miniredis/v2` (for testing)

**Storage**: Redis

**Testing**: `go test`, unit tests with `miniredis` to mock the Redis client.

**Target Platform**: Linux

**Project Type**: Backend Web Service (Infrastructure Layer)

**Performance Goals**: <2ms cache access latency.

**Constraints**: Clean Architecture. Do not expose `go-redis` details to the application or domain layers.

**Scale/Scope**: Storing short-lived JTIs. Graceful failure path logs errors and allows flow degradation instead of causing app panic.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Clean Architecture**: The Redis blacklist implementation resides strictly in the infrastructure layer (`internal/infrastructure/cache`) and only communicates with the application layer via the `TokenBlacklist` interface. (PASS)
2. **Documentation-First**: Spec is finalized. No REST API routes are added/changed, so no OpenAPI updates are needed. (PASS)
3. **Unit-Test-Per-File**: `token_blacklist_redis.go` will be created with its corresponding `token_blacklist_redis_test.go` unit test file. (PASS)
4. **Task-Driven**: Plan is strictly scoped to TSK-AUTH-012. (PASS)
5. **Observability**: Structured logging (`log/slog`) is used for cache operations and connection failures. (PASS)

## Project Structure

### Documentation (this feature)

```text
specs/012-redis-token-blacklist/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── quickstart.md        # Phase 1 output
```

### Source Code (repository root)

```text
internal/
├── application/
│   └── interfaces/
│       └── cache.go                        # Existing TokenBlacklist interface
└── infrastructure/
    └── cache/
        ├── token_blacklist_redis.go        # Redis implementation
        └── token_blacklist_redis_test.go   # Unit tests with miniredis
```

**Structure Decision**: Place the `TokenBlacklist` implementation under `internal/infrastructure/cache` to ensure clean separation of infrastructure from application interfaces.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
