# Testing Strategy

## 1. Testing Goals

Testing must prove that each file and each layer behaves correctly in isolation and that infrastructure boundaries work correctly in integration.

Testing priorities:

1. Domain correctness
2. Use case correctness
3. HTTP contract behavior
4. Repository persistence behavior
5. Kafka idempotency and offset safety
6. Redis cache behavior
7. Fx wiring correctness

Rule:

```text
Every non-generated .go source file must have a corresponding _test.go file or a documented exemption.
```

---

## 2. Test Pyramid

Use this pyramid:

```text
Many unit tests
Some integration tests
Few end-to-end tests
```

Unit tests should be fast and deterministic.
Integration tests may use containers or dedicated local services.
End-to-end tests should cover only critical flows.

---

## 3. Unit Test Rules

Unit tests must not require:

- Real PostgreSQL
- Real Redis
- Real Kafka
- Network access
- File system writes unless testing file behavior
- Sleep-based timing

Use mocks, fakes, or in-memory implementations.

Example command:

```bash
go test ./... -race -count=1
```

---

## 4. Unit Test Per File Policy

For each source file:

```text
<name>.go -> <name>_test.go
```

Examples:

```text
user_service.go -> user_service_test.go
user_handler.go -> user_handler_test.go
user_repository.go -> user_repository_test.go
cache.go -> cache_test.go
producer.go -> producer_test.go
```

Allowed exemptions:

- `main.go`
- generated files
- type-only files covered by package tests
- Fx module files with no logic
- static constants-only files

Every exemption must be documented in the PR/task summary.

---

## 5. Domain Tests

Domain tests should validate pure business behavior.

Test examples:

- Entity creation succeeds with valid input
- Entity creation fails with invalid input
- Value object validation
- State transitions
- Invariant protection

Example:

```go
func TestNewUser(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        fullName string
        wantErr error
    }{
        {
            name:     "valid user",
            email:    "john@example.com",
            fullName: "John Doe",
        },
        {
            name:    "invalid email",
            email:   "invalid",
            wantErr: domain.ErrInvalidEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := domain.NewUser(tt.email, tt.fullName)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.email, user.Email.String())
        })
    }
}
```

---

## 6. Application Use Case Tests

Use case tests should mock repositories, transaction manager, cache, and event publisher.

Must test:

- Happy path
- Validation failure
- Domain error propagation
- Repository error propagation
- Transaction rollback
- Event publishing behavior
- Idempotency behavior if applicable

Example:

```go
func TestUserService_Create(t *testing.T) {
    repo := new(MockUserRepository)
    tx := NewFakeTxManager()
    publisher := new(MockEventPublisher)
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))

    svc := NewUserService(UserServiceParams{
        Users:     repo,
        Tx:        tx,
        Publisher: publisher,
        Log:       logger,
    })

    repo.On("FindByEmail", mock.Anything, domain.Email("john@example.com")).
        Return(nil, domain.ErrUserNotFound)
    repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.User")).
        Return(nil)
    publisher.On("Publish", mock.Anything, "user.created.v1", mock.Anything, mock.Anything).
        Return(nil)

    out, err := svc.Create(context.Background(), CreateUserInput{
        Email: "john@example.com",
        Name:  "John Doe",
    })

    require.NoError(t, err)
    require.NotNil(t, out)
    repo.AssertExpectations(t)
    publisher.AssertExpectations(t)
}
```

---

## 7. HTTP Handler Tests

Handler tests should use Echo test utilities and mocked use cases.

Must test:

- Request binding failure
- Request validation failure
- Successful response
- Error response mapping
- Status code
- JSON response body

Example:

```go
func TestUserHandler_Create(t *testing.T) {
    e := echo.New()
    body := strings.NewReader(`{"email":"john@example.com","name":"John Doe"}`)
    req := httptest.NewRequest(http.MethodPost, "/v1/users", body)
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    ctx := e.NewContext(req, rec)

    svc := new(MockUserUseCase)
    svc.On("Create", mock.Anything, CreateUserInput{
        Email: "john@example.com",
        Name:  "John Doe",
    }).Return(&UserOutput{
        ID:    "01HZX3QK0YS6M2Z1WD9S7VN9VJ",
        Email: "john@example.com",
        Name:  "John Doe",
    }, nil)

    h := NewUserHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))

    err := h.Create(ctx)

    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, rec.Code)
    assert.JSONEq(t, `{"id":"01HZX3QK0YS6M2Z1WD9S7VN9VJ","email":"john@example.com","name":"John Doe"}`, rec.Body.String())
}
```

---

## 8. Repository Tests

Repository tests can be either:

1. Unit tests with `sqlmock`
2. Integration tests with real PostgreSQL through testcontainers

Preferred strategy:

- Use unit tests for query construction and error mapping.
- Use integration tests for migrations, constraints, and real persistence behavior.

Repository tests must cover:

- Insert success
- Find success
- Not found mapping
- Unique constraint mapping
- Transaction behavior
- Context cancellation where practical

---

## 9. Redis Tests

Redis cache tests must cover:

- Set/Get success
- Cache miss behavior
- Delete behavior
- TTL behavior
- Serialization/deserialization error
- Redis unavailable behavior

Unit tests may use mocks.
Integration tests may use testcontainers.

Cache miss rule:

```text
Cache miss is not a business error.
```

---

## 10. Kafka Tests

Kafka producer tests must cover:

- Envelope encoding
- Topic selection
- Message key selection
- Header propagation
- Sarama error mapping

Kafka consumer tests must cover:

- Valid message processing
- Invalid JSON handling
- Unsupported event version
- Idempotency behavior
- Handler error behavior
- Offset commit after success only

Consumer behavior rule:

```text
A Kafka message is considered consumed only after application handler succeeds.
```

---

## 11. Fx Wiring Tests

Fx wiring should be tested with `fx.ValidateApp` where possible.

Example:

```go
func TestAppModule(t *testing.T) {
    err := fx.ValidateApp(
        BootstrapModule,
        PlatformModule,
        InfrastructureModule,
        ApplicationModule,
        AdapterModule,
    )

    require.NoError(t, err)
}
```

Use this to catch missing providers early.

---

## 12. OpenAPI Contract Tests

Contract tests must verify that implemented routes match OpenAPI.

Minimum checks:

- OpenAPI file is valid YAML.
- Required paths exist.
- Required schemas exist.
- Error response schema exists.
- Generated server/types are up to date if generation is used.

Suggested command:

```bash
make openapi-lint
make openapi-generate-check
```

---

## 13. Test Data Strategy

Use builders for complex test data.

Example:

```go
type UserBuilder struct {
    email string
    name  string
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        email: "john@example.com",
        name:  "John Doe",
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.email = email
    return b
}

func (b *UserBuilder) Build(t *testing.T) *domain.User {
    t.Helper()
    user, err := domain.NewUser(b.email, b.name)
    require.NoError(t, err)
    return user
}
```

Rules:

- Avoid huge shared fixtures.
- Each test should own its setup.
- Test data must be clear and intentional.

---

## 14. Coverage Expectations

Minimum expectations:

| Layer | Minimum Coverage |
|---|---:|
| Domain | 90% |
| Application | 85% |
| HTTP adapter | 75% |
| Repository | 70% |
| Infrastructure wrappers | 70% |

Coverage is not a replacement for meaningful assertions.

---

## 15. CI Test Commands

Recommended Makefile targets:

```makefile
test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

test-integration:
	go test ./test/integration/... -count=1

lint:
	golangci-lint run ./...
```

CI must run:

```bash
make lint
make test-race
make test-cover
make openapi-lint
```

---

## 16. Anti-Patterns

Forbidden in tests:

- Tests depending on execution order
- Real external services in unit tests
- `time.Sleep` for synchronization
- Ignoring errors
- Assertion-free tests
- Testing private implementation details too tightly
- Over-mocking simple value objects
- Snapshot tests for unstable payloads

---

## 17. Definition of Done for Tests

A task is not complete unless:

- New source files have matching unit tests.
- Changed behavior has updated tests.
- OpenAPI changes have contract checks.
- Repository changes have persistence tests.
- Kafka handlers have success and failure tests.
- Redis logic has miss/error tests.
- `go test ./... -race -count=1` passes.
