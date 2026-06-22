# Implementation Plan: Brute-Force Lockout Defense

**Branch**: `018-brute-force-lockout-defense` | **Date**: 2026-06-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/018-brute-force-lockout-defense/spec.md`

## Summary

This plan implements a three-phase brute-force lockout defense for the HROS Admin Service.

**Phase 1 (TSK-AUTH-018 — ✅ Done)**: Domain layer primitives — `BruteForceCache` interface, `ErrAccountLocked` error, and `AccountLockedEvent` / `EmailSendEvent` payload structs with unit tests.

**Phase 2 (TSK-AUTH-019 — ✅ Done)**: Redis infrastructure — `RedisBruteForceCache` implementing the `BruteForceCache` interface using `auth:failed_attempts:{email}` (15-min TTL) and `auth:lockout:{email}` (30-min TTL) keys, with fail-open graceful degradation and PII-safe logging. Tested via miniredis.

**Phase 3 (TSK-AUTH-020 — 🔲 Pending)**: Kafka adapter layer — `EventEnvelope[T]` generic struct, `EventPublisher` interface, and `EmailKafkaProducer` that wraps `EmailSendEvent` into the standard envelope and dispatches to topic `email.send.v1` via `sarama.SyncProducer`. Wired into the Uber Fx dependency graph. Tested with `sarama/mocks`.

The login use case will be extended (post-Phase 3) to consume `BruteForceCache` for lockout checking, incrementing, and resetting — this wiring is tracked in the application module and `LoginUseCase`.

---

## Technical Context

**Language/Version**: Go 1.26.1 (module `github.com/hros/admin-service`)

**Primary Dependencies**:
- `github.com/IBM/sarama` — Kafka sync producer and mock producer for tests
- `github.com/redis/go-redis/v9` — Redis client used in `RedisBruteForceCache`
- `github.com/alicebob/miniredis/v2` — In-memory Redis for unit tests (no external infra)
- `github.com/stretchr/testify` — `assert` / `require` for all test suites
- `go.uber.org/fx` — Dependency injection and module wiring
- Standard library only for domain and application layers

**Storage**:
- Redis — ephemeral brute-force state store (`auth:failed_attempts:{email}`, `auth:lockout:{email}`)
- No PostgreSQL changes required for this feature

**Testing**:
- Unit tests: `go test ./... -race -count=1` with miniredis for Redis, `sarama/mocks` for Kafka
- Coverage targets (constitution): Application ≥85%, Infrastructure ≥70%, Adapter ≥75%
- Zero external runtime dependencies in unit tests

**Target Platform**: Linux server / local developer machines

**Project Type**: web-service (Go backend, Clean Architecture)

**Performance Goals**:
- Lockout check via `BruteForceCache.IsLocked`: p95 < 5ms (Redis single-key GET)
- Kafka event publish: async, non-blocking to the caller; failure is logged but does not fail login

**Constraints**:
- Domain layer (`internal/domain`) must import **zero** infrastructure packages
- Application layer (`internal/application`) must import **zero** Echo or GORM packages
- `BruteForceCache` must fail-open: Redis unavailability must never block a legitimate login
- Kafka publishing errors must be logged but must not propagate as login errors
- Email addresses must never appear in log output — use SHA256-truncated mask

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|---------|
| **I. Clean Architecture & Strict Boundaries** | ✅ PASS | Domain events and errors have no external imports. `BruteForceCache` interface lives in `application/interfaces`. Redis implementation lives in `infrastructure/cache`. Kafka adapter lives in `adapter/kafka/producer`. |
| **II. Documentation-First & OpenAPI-Driven** | ✅ PASS | No new REST endpoints added; this feature is internal (login flow enrichment) and async Kafka event publishing. OpenAPI contract unchanged for this feature. |
| **III. Unit-Test-Per-File (NON-NEGOTIABLE)** | ✅ PASS | Every production file produced in all three phases has a corresponding `_test.go` file. |
| **IV. Task-Driven & Atomic Implementation** | ✅ PASS | Three distinct task IDs (TSK-AUTH-018, 019, 020) map to three implementation phases. No cross-task scope bleed. |
| **V. Observability & Structured Logging** | ✅ PASS | `RedisBruteForceCache` uses `slog.ErrorContext`/`WarnContext`/`InfoContext` with structured attributes. Emails masked via SHA256. `EmailKafkaProducer` will log publish success and failures. |

---

## Research Decisions

### Decision 1: Fail-Open Strategy for Redis

**Decision**: All `BruteForceCache` methods degrade gracefully on Redis errors — they log the error and return a safe default (0 attempts, not-locked, no error returned).

**Rationale**: Redis unavailability must never block a legitimate administrator login. The lockout defense is a best-effort security layer; liveness takes precedence over security hardening when the cache is down.

**Alternatives rejected**: Fail-closed (block all logins on Redis error) — unacceptable operational risk.

---

### Decision 2: PII Protection in Redis Keys and Logs

**Decision**: Email addresses are used verbatim as the Redis key suffix (`auth:failed_attempts:{email}`, `auth:lockout:{email}`) but are SHA256-truncated to 12 hex chars in all `slog` output via `maskEmail()`.

**Rationale**: The key pattern must be deterministic for correctness (lookup by email at login time). Logs must not leak PII per constitution Principle V.

**Alternatives rejected**: Hashing keys in Redis — makes debugging and manual inspection impossible without the source email; adds complexity with no operational benefit since Redis access is internal.

---

### Decision 3: EventEnvelope Ownership

**Decision**: The `EventEnvelope[T any]` generic struct and `EventPublisher` interface are defined in `internal/adapter/kafka/producer/` (the adapter owns the envelope), not in a shared platform package.

**Rationale**: The tech-stack document defines the envelope shape but does not mandate a shared package. Placing it in the adapter keeps the domain and application layers free of Kafka coupling. Other future producers can reference the same struct or define their own.

**Alternatives rejected**: Shared `internal/platform/kafka/envelope.go` — couples unrelated packages to Kafka-specific data shapes prematurely.

---

### Decision 4: Kafka Publish Errors Must Not Fail Login

**Decision**: `LoginUseCase.Execute` will call `EmailKafkaProducer.PublishLockoutEmail` in a non-blocking best-effort manner. If publishing fails, the error is logged but `ErrAccountLocked` is still returned to the caller.

**Rationale**: SC-004 requires 100% reliability for audit log and Kafka event generation, but a Kafka broker outage must not prevent the lockout from taking effect. The lockout itself (Redis state) is the safety mechanism; the email is a notification.

**Alternatives rejected**: Blocking publish with error propagation — a Kafka outage would cause `ErrInternalServer` instead of `ErrAccountLocked`, exposing infrastructure state to the caller.

---

### Decision 5: Message Key = Recipient Email

**Decision**: The Sarama `ProducerMessage.Key` is set to `sarama.StringEncoder(event.To)`.

**Rationale**: Using the recipient email as the partition key guarantees that lockout emails for the same user are delivered in-order to the same partition. This matches the tech-stack convention: `<tenant_id>:<aggregate_id>` — email is the effective aggregate ID for this event.

---

## Project Structure

### Documentation (this feature)

```text
specs/018-brute-force-lockout-defense/
├── plan.md                        # This file
├── spec.md                        # Feature specification
├── tasks.md                       # Task list (phases 2–5)
└── checklists/
    └── requirements.md            # Specification quality checklist
```

### Source Code — Completed (TSK-AUTH-018, TSK-AUTH-019)

```text
internal/
├── domain/
│   ├── errors/
│   │   ├── auth_errors.go                  ✅  ErrAccountLocked defined
│   │   └── auth_errors_test.go             ✅  Unit tests
│   └── events/
│       ├── auth_events.go                  ✅  AccountLockedEvent, EmailSendEvent
│       └── auth_events_test.go             ✅  Serialization unit tests
├── application/
│   └── interfaces/
│       ├── brute_force_cache.go            ✅  BruteForceCache interface
│       └── brute_force_cache_test.go       ✅  Interface compile / contract tests
└── infrastructure/
    └── cache/
        ├── brute_force_redis.go            ✅  RedisBruteForceCache implementation
        └── brute_force_redis_test.go       ✅  Miniredis unit tests (5 test functions)
```

### Source Code — Pending (TSK-AUTH-020)

```text
internal/
└── adapter/
    └── kafka/
        └── producer/
            ├── email_events.go             🔲  EventEnvelope[T], EventPublisher,
            │                                   EmailKafkaProducer.PublishLockoutEmail
            ├── email_events_test.go        🔲  sarama/mocks unit tests
            └── module.go                   🔲  Fx provider for EmailKafkaProducer
```

### Wiring Changes Required (Post TSK-AUTH-020)

```text
internal/app/app.go                         🔲  Add cache.NewRedisBruteForceCache
                                                and EmailKafkaProducer to fx.Options
internal/application/module.go             🔲  (future) LoginUseCase extended to
                                                accept BruteForceCache dependency
```

**Structure Decision**: The adapter layer (`internal/adapter/kafka/producer`) is the correct home for the envelope, publisher interface, and email producer, per Clean Architecture: adapters translate between domain events and infrastructure transports. No changes to domain or application packages are required for Phase 3.

---

## Key Data Models

### EventEnvelope (adapter/kafka/producer)

```go
type EventEnvelope[T any] struct {
    ID            string    `json:"id"`             // UUID v4
    Type          string    `json:"type"`            // e.g. "email.send"
    Source        string    `json:"source"`          // e.g. "admin-service"
    Version       int       `json:"version"`         // 1
    CorrelationID string    `json:"correlation_id"`  // from context or generated
    OccurredAt    time.Time `json:"occurred_at"`     // UTC
    Data          T         `json:"data"`
}
```

### EmailKafkaProducer (adapter/kafka/producer)

```go
type EventPublisher interface {
    Publish(ctx context.Context, topic string, key string, event any) error
}

type EmailKafkaProducer struct {
    producer sarama.SyncProducer
    source   string
    logger   *slog.Logger
}

func (p *EmailKafkaProducer) PublishLockoutEmail(
    ctx context.Context,
    event events.EmailSendEvent,
) error
```

### Kafka Topic & Key Convention

| Field | Value |
|-------|-------|
| Topic | `email.send.v1` |
| Key | `event.To` (recipient email address) |
| Partition strategy | Key-based (per-user ordering) |
| Serialization | JSON (standard library `encoding/json`) |

### Redis Key Layout

| Key Pattern | TTL | Description |
|-------------|-----|-------------|
| `auth:failed_attempts:{email}` | 15 min sliding | Consecutive failed login counter |
| `auth:lockout:{email}` | 30 min fixed | Lockout expiry timestamp (RFC3339) |

---

## API Contract

No new REST endpoints are added by this feature. The existing `POST /v1/auth/login` will internally call `BruteForceCache.IsLocked` before password verification and return `HTTP 423` or `HTTP 403` for a locked account.

The Kafka `email.send.v1` event is consumed by a downstream notification service and is not part of this service's REST API contract.

---

## Quickstart Validation

```bash
# Run all unit tests for this feature (no Docker required)
go test -race -count=1 \
  ./internal/domain/errors/... \
  ./internal/domain/events/... \
  ./internal/application/interfaces/... \
  ./internal/infrastructure/cache/... \
  ./internal/adapter/kafka/producer/...

# Run all tests in the repository
go test ./... -race -count=1

# Format and lint
go fmt ./...
golangci-lint run
```

**Expected outcome**: All tests pass. No races detected. No lint errors in modified files.

---

## Complexity Tracking

*No constitution violations found.*

All complexity is justified:

| Concern | Justification |
|---------|---------------|
| Generic `EventEnvelope[T any]` | Required by tech-stack doc; avoids `interface{}` payload casting at the consumer side |
| SHA256 email masking in logs | Required by constitution Principle V (no PII in logs) |
| Fail-open Redis degradation | Required by constitution Principle I (liveness over security hardening when infra is down) |
| Separate `module.go` per adapter layer | Required by constitution Principle I and Uber Fx convention (each bounded context wires its own providers) |
