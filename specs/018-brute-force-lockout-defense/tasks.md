# Tasks: Brute-Force Lockout Defense — Kafka Producer Adapter (TSK-AUTH-020)

**Input**: Design documents from `/specs/018-brute-force-lockout-defense/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/kafka.md ✅

**Scope**: This tasks.md covers **only the pending work**: the Kafka producer adapter layer (TSK-AUTH-020).
Phases 1–3 (TSK-AUTH-018 domain primitives, TSK-AUTH-019 Redis cache) are already complete and are recorded below as checkpointed history.

**Tests**: Unit tests are mandatory per constitution Principle III (unit-test-per-file). Every production file must have a corresponding `_test.go`.

**Organization**: Tasks are grouped by implementation phase. Phase 4 is the only actionable phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on each other)
- **[Story]**: Maps to user story in spec.md (US4 = Kafka adapter serializes lockout email event)
- Include exact file paths in every task description

---

## Phase 1: Setup ✅

- [x] T001 Setup feature spec, plan, checklist, and design artifacts in `specs/018-brute-force-lockout-defense/`

---

## Phase 2: Domain Primitives (TSK-AUTH-018) ✅

**Purpose**: Application-layer interface, domain error, and event payload structs — zero external dependencies.

- [x] T002 [P] Define `BruteForceCache` interface in `internal/application/interfaces/brute_force_cache.go`
- [x] T003 [P] Define `ErrAccountLocked` in `internal/domain/errors/auth_errors.go`
- [x] T004 [P] Define `AccountLockedEvent` and `EmailSendEvent` structs in `internal/domain/events/auth_events.go`
- [x] T005 [P] Implement unit tests in:
  - `internal/domain/errors/auth_errors_test.go`
  - `internal/domain/events/auth_events_test.go`
  - `internal/application/interfaces/brute_force_cache_test.go`

**Checkpoint** ✅: Interface compiles; domain errors and events defined; unit tests pass with zero external dependencies.

---

## Phase 3: Redis Cache Implementation (TSK-AUTH-019) ✅

**Purpose**: `BruteForceCache` interface implementation backed by Redis, with fail-open graceful degradation.

- [x] T006 Implement `RedisBruteForceCache` in `internal/infrastructure/cache/brute_force_redis.go` with five methods: `IncrementFailedAttempts`, `GetFailedAttempts`, `SetLockout`, `IsLocked`, `Reset`. Key prefixes: `auth:failed_attempts:{email}` (15-min TTL), `auth:lockout:{email}` (30-min TTL, RFC3339 value). SHA256-masked email in all log output.
- [x] T007 Implement unit tests in `internal/infrastructure/cache/brute_force_redis_test.go` using `miniredis` covering: first-increment TTL assignment, subsequent-increment TTL preservation, cache-miss returns, malformed-value TTL fallback, fail-open on Redis error for all five methods.

**Checkpoint** ✅: Redis implementation complete; all miniredis tests pass; graceful degradation verified.

---

## Phase 4: Kafka Producer Adapter (TSK-AUTH-020) 🔲 Pending

**Goal**: Implement the adapter layer that wraps `EmailSendEvent` domain payloads in the standard `EventEnvelope` and dispatches them to Kafka topic `email.send.v1` via `sarama.SyncProducer`. Wire into the Uber Fx dependency graph.

**Independent Test**: Construct an `EmailSendEvent` with a recipient email and unlock timestamp, call `PublishLockoutEmail` against a `sarama/mocks.MockSyncProducer`, capture the `ProducerMessage`, unmarshal the value into `EventEnvelope[EmailSendEvent]`, and assert all fields are correctly populated.

### Implementation for User Story 4 (US4)

- [x] T008 [P] [US4] Define `EventEnvelope[T any]` generic struct and `EventPublisher` interface in `internal/adapter/kafka/producer/email_events.go`.

  **Exact contract**:
  ```go
  // EventEnvelope is the standard Kafka message wrapper for all events published by this service.
  type EventEnvelope[T any] struct {
      ID            string    `json:"id"`
      Type          string    `json:"type"`
      Source        string    `json:"source"`
      Version       int       `json:"version"`
      CorrelationID string    `json:"correlation_id"`
      OccurredAt    time.Time `json:"occurred_at"`
      Data          T         `json:"data"`
  }

  // EventPublisher defines the adapter contract for publishing domain events to Kafka.
  type EventPublisher interface {
      Publish(ctx context.Context, topic string, key string, event any) error
  }
  ```

  **Also implement** `EmailKafkaProducer` struct and its constructor:
  ```go
  type EmailKafkaProducer struct {
      producer sarama.SyncProducer
      source   string
      logger   *slog.Logger
  }

  func NewEmailKafkaProducer(producer sarama.SyncProducer, logger *slog.Logger) *EmailKafkaProducer
  ```

  **Also implement** `PublishLockoutEmail`:
  - Validate: return `fmt.Errorf("recipient email must not be empty")` if `event.To == ""`
  - Build envelope: `ID` = new UUID (`domain.NewUUID()`), `Type` = `"email.send"`, `Source` = `p.source` (`"admin-service"`), `Version` = `1`, `CorrelationID` = extracted from ctx or `""`, `OccurredAt` = `time.Now().UTC()`, `Data` = `event`
  - Marshal envelope to JSON with `encoding/json`
  - Send via `p.producer.SendMessage(&sarama.ProducerMessage{Topic: "email.send.v1", Key: sarama.StringEncoder(event.To), Value: sarama.ByteEncoder(payload)})`
  - On success: log `slog.InfoContext` with `event="kafka.email_send.published"`, `topic`, `key` (masked)
  - On error: wrap and return `fmt.Errorf("publish lockout email: %w", err)`

- [x] T009 [P] [US4] Implement unit tests in `internal/adapter/kafka/producer/email_events_test.go` using `sarama/mocks.NewMockSyncProducer`.

  **Required test cases**:

  1. `TestEmailKafkaProducer_PublishLockoutEmail_HappyPath`
     - Create `mocks.NewMockSyncProducer(t, nil)` and call `mock.ExpectSendMessageAndSucceed()`
     - Call `PublishLockoutEmail` with a valid `EmailSendEvent{To: "admin@hros.io", Subject: "Account Locked", Template: "account_locked_notification", TemplateData: map[string]interface{}{"email": "admin@hros.io", "unlock_at": "2026-06-21T17:15:00Z"}}`
     - `require.NoError`
     - Capture the message via `mock.YieldMessage()` or by inspecting what was passed; unmarshal `Value` into `EventEnvelope[events.EmailSendEvent]`
     - Assert: `ID` is non-empty, `Type == "email.send"`, `Source == "admin-service"`, `Version == 1`, `OccurredAt` is non-zero, `Data.To == "admin@hros.io"`, `Data.Template == "account_locked_notification"`
     - Assert message `Key` encodes to `"admin@hros.io"`

  2. `TestEmailKafkaProducer_PublishLockoutEmail_SaramaError`
     - Call `mock.ExpectSendMessageAndFail(errors.New("broker unavailable"))`
     - Assert returned error is non-nil and wraps the original error (`errors.Is` or `strings.Contains`)

  3. `TestEmailKafkaProducer_PublishLockoutEmail_EmptyRecipient`
     - Call `PublishLockoutEmail` with `event.To = ""`
     - Assert returned error is non-nil and `SendMessage` was **not** called (mock has no expectation set; mock fails test if unexpectedly called)

- [x] T010 [US4] Create `internal/adapter/kafka/producer/module.go` to wire `EmailKafkaProducer` into the Uber Fx dependency graph.

  **Exact implementation**:
  ```go
  package producer

  import "go.uber.org/fx"

  // Module is the Fx module for the Kafka producer adapter.
  var Module = fx.Module("kafka-producer",
      fx.Provide(NewEmailKafkaProducer),
  )
  ```

  No unit test is required for `module.go` (it is a wire-only bootstrap file with no logic; exempt per coding conventions §3).

- [x] T011 [US4] Register `producer.Module` and `cache.NewRedisBruteForceCache` in `internal/app/app.go`.

  **Changes**:
  - Add import for `"github.com/hros/admin-service/internal/adapter/kafka/producer"` (alias `kafkaProducer`)
  - Add `fx.Provide(cache.NewRedisBruteForceCache)` alongside the existing `cache.NewRedisTokenBlacklist` line
  - Add `kafkaProducer.Module` to the `fx.Options` list after `adapterHttp.Module`

  **Note**: `app.go` already has the `cache` package imported; only the new provider call and producer module need adding. Verify the app still compiles with `go build ./...`.

**Checkpoint**: `go test -race -count=1 ./internal/adapter/kafka/producer/...` passes. `go build ./...` succeeds with all wiring. No real Kafka broker required.

---

## Phase 5: Polish ✅ / 🔲 Post-Phase-4

**Purpose**: Format verification and regression guard after Phase 4 completes.

- [x] T012 Run `go fmt ./internal/adapter/kafka/producer/... ./internal/app/...` and `golangci-lint run` on all modified files. Fix any lint warnings.
- [x] T013 Run the full unit test suite `go test -race -count=1 ./...` to verify zero regressions across all packages including pre-existing tests for TSK-AUTH-018 and TSK-AUTH-019.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phases 1–3** (TSK-AUTH-018, 019): ✅ Complete — no action needed
- **Phase 4** (TSK-AUTH-020): Can start immediately. T008 and T009 are independent files and can be worked on in parallel. T010 has no code dependency on T008/T009 (it only wires the type). T011 depends on T010 existing.
- **Phase 5**: Must follow Phase 4 completion

### Within Phase 4

```
T008 ──┐
       ├──► T010 ──► T011 ──► T012 ──► T013
T009 ──┘
```

- **T008** and **T009** can be worked in parallel (different files, T009 imports T008's types but can be written concurrently by a developer familiar with the types)
- **T010** (`module.go`) can be written in parallel with T008/T009 — it only needs to know the constructor signature
- **T011** (`app.go`) must follow T010
- **T012–T013** (polish) must follow T011

### Parallel Opportunities

```bash
# These can be launched simultaneously:
Task T008: "Define EventEnvelope, EventPublisher, EmailKafkaProducer in internal/adapter/kafka/producer/email_events.go"
Task T009: "Implement unit tests in internal/adapter/kafka/producer/email_events_test.go"
Task T010: "Create module.go in internal/adapter/kafka/producer/module.go"
```

---

## Implementation Strategy

### Minimum Viable Task (TSK-AUTH-020 only)

1. Implement T008 — `email_events.go` (core logic)
2. Implement T009 — `email_events_test.go` (verify correctness)
3. Implement T010 — `module.go` (Fx wiring)
4. Implement T011 — wire into `app.go` (register provider)
5. Run T012–T013 — format, lint, full test suite

**Validation gate**: `go test -race -count=1 ./internal/adapter/kafka/producer/...` must pass before T011.

---

## Notes

- `[P]` tasks operate on different files with no blocking dependency — safe to implement concurrently
- `sarama/mocks.MockSyncProducer` is already imported by `test/integration/auth_flow_test.go` and `session_persistence_flow_test.go` — no new `go.mod` entry required
- Do NOT use `sarama.NewSyncProducer` in tests — always use `mocks.NewMockSyncProducer`
- Email masking is NOT required in the Kafka adapter logs (the `To` field in the log should use a masked or omitted representation — use `slog.String("key_masked", maskEmail(event.To))` if logging the key)
- `domain.NewUUID()` is already defined in `internal/domain/id.go` — use it for envelope ID generation
- `app.go` already imports the `cache` package — only add the new `fx.Provide(cache.NewRedisBruteForceCache)` line; do not restructure the file
