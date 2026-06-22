# Tasks: Brute-Force Lockout Defense — LoginUseCase Orchestration (TSK-AUTH-021)

**Input**: Design documents from `/specs/018-brute-force-lockout-defense/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/kafka.md ✅, TSK-AUTH-020 ✅

**Scope**: This tasks.md now covers **all pending work**: TSK-AUTH-021 — the LoginUseCase lockout state-machine (Phase 5).
Phases 1–4 (TSK-AUTH-018 domain primitives, TSK-AUTH-019 Redis cache, TSK-AUTH-020 Kafka adapter, TSK-AUTH-020 wiring) are already complete and recorded below as checkpointed history.

**Tests**: Unit tests are mandatory per constitution Principle III (unit-test-per-file). Every production file must have a corresponding `_test.go`.

**Organization**: Tasks are grouped by implementation phase. Phase 5 is the only actionable phase.

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

## Phase 5: LoginUseCase Lockout Orchestration (TSK-AUTH-021) 🔲 Pending

**Goal**: Update `LoginUseCase.Execute` to implement the three-step brute-force state machine:
1. Pre-check: `BruteForceCache.IsLocked()` — short-circuit with `ErrAccountLocked` if true.
2. On bad password: `BruteForceCache.IncrementFailedAttempts()`; when count reaches 5, call `SetLockout()`, append `account.locked` audit event, and publish `email.send` via `EmailKafkaProducer` (best-effort, errors logged not propagated).
3. On good password: `BruteForceCache.ClearFailures()` to reset the counter before returning a session.

**Independent Test**: Mock `BruteForceCache`, `EmailKafkaProducer.PublishLockoutEmail`, and the audit-log writer; assert the exact sequence of calls for all four branches.

### Implementation for User Story 5 (US5)

- [x] T014 [US5] Update `LoginUseCase` in `internal/application/usecase/login_usecase.go`.

  **Constructor signature change** (inject new dependencies):
  ```go
  func NewLoginUseCase(
      userRepo      domain.AdminUserRepository,
      sessionRepo   domain.SessionTokenRepository,
      password      auth.PasswordHelper,
      tokens        auth.TokenProvider,
      audit         authDomain.AuditLogger,
      bruteForce    interfaces.BruteForceCache,
      lockoutNotify interfaces.LockoutNotifier,
      logger        *slog.Logger,
  ) *LoginUseCase
  ```

  **Execute orchestration (strict ordering)**:
  1. `if locked, _ := bruteForce.IsLocked(ctx, req.Email); locked { return ErrAccountLocked }`
  2. Verify password via bcrypt; if invalid:
     - `count, _ := bruteForce.IncrementFailedAttempts(ctx, req.Email)`
     - `if count >= 5 { bruteForce.SetLockout(ctx, req.Email); auditLog.Append(ctx, "account.locked", ...); emailPub.PublishLockoutEmail(ctx, events.EmailSendEvent{...}) }`
     - Return `ErrInvalidCredentials`
  3. Password valid:
     - `bruteForce.ClearFailures(ctx, req.Email)`
     - Issue session tokens and return result

  **Notes**:
  - `ClearFailures` wraps the existing `Reset` method on the `BruteForceCache` interface (or is a semantic alias; use whichever name the interface exposes).
  - Kafka publish errors must be logged with `slog.ErrorContext` and NOT returned — `ErrAccountLocked` is the single outcome of the 5th-failure branch.
  - All `BruteForceCache` method errors must be fail-open (log and continue).
  - Audit-log call may use an existing `AuditLogger` interface already in the application layer, or a minimal new one — do NOT add a GORM call to the use case.

- [x] T015 [US5] Implement unit tests in `internal/application/usecase/login_usecase_test.go`.

  **Required test cases (100% branch coverage)**:

  1. `TestLoginUseCase_AlreadyLocked`
     - Mock `IsLocked` returns `(true, nil)`
     - Assert use case returns `ErrAccountLocked`
     - Assert password check, `IncrementFailedAttempts`, `SetLockout`, `ClearFailures`, and email publisher are **not** called

  2. `TestLoginUseCase_InvalidPassword_LessThan5Failures`
     - Mock `IsLocked` returns `(false, nil)`; bcrypt check fails; `IncrementFailedAttempts` returns `(3, nil)`
     - Assert use case returns `ErrInvalidCredentials`
     - Assert `SetLockout`, audit log, and email publisher are **not** called

  3. `TestLoginUseCase_InvalidPassword_FifthFailure_TriggersLockout`
     - Mock `IsLocked` returns `(false, nil)`; bcrypt check fails; `IncrementFailedAttempts` returns `(5, nil)`
     - Assert `SetLockout` is called once
     - Assert audit log `Append` is called with event `"account.locked"`
     - Assert `PublishLockoutEmail` is called with correct `To` field
     - Assert use case returns `ErrAccountLocked`

  4. `TestLoginUseCase_ValidPassword_ClearsFailures`
     - Mock `IsLocked` returns `(false, nil)`; bcrypt check passes; `ClearFailures` returns `nil`
     - Assert `IncrementFailedAttempts`, `SetLockout`, and email publisher are **not** called
     - Assert use case returns a valid session result

**Checkpoint**: `go test -race -count=1 ./internal/application/usecase/...` passes with 100% statement coverage on `login_usecase.go`. `go build ./...` succeeds.

---

## Phase 6: Polish 🔲 Post-Phase-5

**Purpose**: Format verification and regression guard after Phase 5 completes.

- [ ] T016 Run `go fmt ./internal/application/usecase/... ./internal/app/...` and `golangci-lint run` on all modified files. Fix any lint warnings.
- [ ] T017 Run the full unit test suite `go test -race -count=1 ./...` to verify zero regressions across all packages.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phases 1–4** (TSK-AUTH-018, 019, 020): ✅ Complete — no action needed
- **Phase 5** (TSK-AUTH-021): Can start immediately. T014 and T015 are independent files and can be worked in parallel.
- **Phase 6**: Must follow Phase 5 completion

### Within Phase 5

```bash
T014 (login_usecase.go)
       ├──► T016 ──► T017
T015 (login_usecase_test.go)
```

- **T014** and **T015** can be worked in parallel (different files; T015 author needs the type signature from T014 but can write test stubs concurrently)
- **T016–T017** (polish) must follow T015

### Parallel Opportunities

```bash
# These can be launched simultaneously:
Task T014: "Update LoginUseCase in internal/application/usecase/login_usecase.go"
Task T015: "Implement unit tests in internal/application/usecase/login_usecase_test.go"
```

---

## Implementation Strategy

### Minimum Viable Task (TSK-AUTH-021 only)

1. Implement T014 — `login_usecase.go` (lockout state machine)
2. Implement T015 — `login_usecase_test.go` (100% branch coverage)
3. Run T016–T017 — format, lint, full test suite

**Validation gate**: `go test -race -count=1 ./internal/application/usecase/...` must pass with 100% coverage before T016.

---

## Notes

- `[P]` tasks operate on different files with no blocking dependency — safe to implement concurrently
- `BruteForceCache.IncrementFailedAttempts` returns `(count int, err error)` — use the count (not a second `GetFailedAttempts` call) to determine whether lockout threshold has been reached
- `ClearFailures` is the semantic name used in the task description; use the actual method name exposed on the `BruteForceCache` interface (`Reset` or equivalent as defined in `internal/application/interfaces/brute_force_cache.go`)
- Do NOT call `sarama.NewSyncProducer` directly from the use case — inject `*EmailKafkaProducer` or its interface via constructor
- Do NOT add GORM or Redis calls inside `LoginUseCase` — all cache and Kafka interactions go through the injected interfaces
- `app.go` wiring changes (registering the new `LoginUseCase` constructor parameters) are included in T014 scope if they require changes to `internal/application/module.go`
