# Implementation Rules

## 1. Non-Negotiable Rules

These rules must be followed by human developers and AI coding agents.

1. Do not implement business logic in HTTP handlers.
2. Do not import Echo outside HTTP adapter packages.
3. Do not import GORM outside infrastructure repository packages.
4. Do not import Sarama outside infrastructure Kafka packages.
5. Do not import go-redis outside infrastructure Redis packages.
6. Do not import Fx inside domain or application business logic.
7. Do not return database models from use cases.
8. Do not expose domain entities directly as HTTP responses.
9. Do not create endpoints without updating OpenAPI.
10. Do not create `.go` source files without corresponding `_test.go` files unless explicitly exempted.

---

## 2. Implementation Workflow

Every feature must follow this order:

```text
1. Read requirements
2. Update OpenAPI contract
3. Define domain model/rules
4. Define application input/output
5. Implement use case
6. Implement repository interface if needed
7. Implement infrastructure repository/cache/broker
8. Implement HTTP handler
9. Register route through Fx
10. Add unit tests per file
11. Add integration tests if infrastructure changed
12. Run lint/test/openapi checks
```

Do not start from the database or HTTP handler unless the task is explicitly infrastructure-only.

---

## 3. Feature Slice Rule

Implement one vertical slice at a time.

A valid vertical slice may include:

```text
OpenAPI path
HTTP request/response DTO
HTTP handler
Application use case
Domain entity/rule
Repository interface
Repository implementation
Migration
Unit tests
```

Do not implement multiple unrelated features in one task.

---

## 4. Clean Architecture Boundaries

### Domain Layer

Allowed:

```text
standard library
pure domain packages
```

Forbidden:

```text
Echo
GORM
Redis
Sarama
Fx
Viper/config loader
HTTP DTO
DB model
```

### Application Layer

Allowed:

```text
domain interfaces
domain entities
application input/output
slog through injected logger
```

Forbidden:

```text
Echo context
GORM DB
Redis client
Sarama producer
HTTP request/response structs
```

### Adapter Layer

Allowed:

```text
Echo
HTTP DTOs
Kafka decoded messages
application use cases
```

Forbidden:

```text
SQL queries
business workflow
transaction orchestration
```

### Infrastructure Layer

Allowed:

```text
GORM
Redis client
Sarama
external SDKs
```

Forbidden:

```text
HTTP routing
business use case orchestration
```

---

## 5. API Implementation Rules

Every REST endpoint must have:

- OpenAPI path
- Request schema if body exists
- Response schema
- Error response schema
- Handler method
- Handler unit test
- Use case method
- Use case unit test

HTTP status rules:

| Action | Success Status |
|---|---:|
| Create sync resource | 201 |
| Create async command | 202 |
| Read resource | 200 |
| List resources | 200 |
| Partial update | 200 |
| Delete resource | 204 |

Error format must be consistent:

```json
{
  "code": "validation_error",
  "message": "Request validation failed",
  "details": {},
  "trace_id": "01HZX3QK0YS6M2Z1WD9S7VN9VJ"
}
```

---

## 6. Handler Rules

Handler is only allowed to:

1. Bind request
2. Validate request
3. Map request to application input
4. Call use case
5. Map output to response
6. Return response

Example:

```go
func (h *UserHandler) Create(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return NewHTTPError(http.StatusBadRequest, "invalid_request", err)
    }

    if err := h.validator.Validate(req); err != nil {
        return NewHTTPError(http.StatusBadRequest, "validation_error", err)
    }

    out, err := h.users.Create(c.Request().Context(), CreateUserInput{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return err
    }

    return c.JSON(http.StatusCreated, ToUserResponse(out))
}
```

Forbidden inside handler:

```go
h.db.Create(...)
h.redis.Set(...)
h.kafka.SendMessage(...)
complex if/else business rules
```

---

## 7. Use Case Rules

Use case must orchestrate business flow.

Allowed:

- Validate business rules
- Call repositories
- Control transactions
- Call cache interface if needed
- Publish events through interface
- Return application output

Example:

```go
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*UserOutput, error) {
    existing, err := s.users.FindByEmail(ctx, domain.Email(input.Email))
    if err == nil && existing != nil {
        return nil, domain.ErrUserAlreadyExists
    }
    if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
        return nil, err
    }

    var output *UserOutput
    err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
        user, err := domain.NewUser(input.Email, input.Name)
        if err != nil {
            return err
        }

        if err := s.users.Save(ctx, user); err != nil {
            return err
        }

        if err := s.events.Publish(ctx, "user.created.v1", user.ID.String(), UserCreatedEvent{
            ID:    user.ID.String(),
            Email: user.Email.String(),
        }); err != nil {
            return err
        }

        output = ToUserOutput(user)
        return nil
    })
    if err != nil {
        return nil, err
    }

    return output, nil
}
```

---

## 8. Repository Rules

Repository interface belongs to domain or application layer.
Repository implementation belongs to infrastructure layer.

Interface:

```go
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id ID) (*User, error)
    FindByEmail(ctx context.Context, email Email) (*User, error)
}
```

Implementation:

```go
type UserGormRepository struct {
    db *gorm.DB
}
```

Rules:

- Always use `db.WithContext(ctx)`.
- Map GORM errors to domain/application errors.
- Do not leak `gorm.ErrRecordNotFound` outside repository.
- Do not return DB model outside repository.
- Keep SQL/GORM query logic in repository only.

---

## 9. Transaction Rules

Transaction boundary must be application-controlled.

Infrastructure provides:

```go
type GormTxManager struct {
    db *gorm.DB
}

func (m *GormTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
    return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txCtx := context.WithValue(ctx, txContextKey{}, tx)
        return fn(txCtx)
    })
}
```

Repositories must detect transaction from context if present.

Rule:

```text
Do not call db.Transaction directly inside handlers or repositories for business workflows.
```

---

## 10. Kafka Rules

Kafka publishing must happen through application-facing interface.

Required envelope:

```go
type Envelope[T any] struct {
    ID            string    `json:"id"`
    Type          string    `json:"type"`
    Source        string    `json:"source"`
    Version       int       `json:"version"`
    CorrelationID string    `json:"correlation_id"`
    OccurredAt    time.Time `json:"occurred_at"`
    Data          T         `json:"data"`
}
```

Producer rules:

- Use message key for partition stability.
- Include event ID.
- Include correlation ID.
- Return publish errors to use case.
- Consider outbox pattern for critical events.

Consumer rules:

- Decode envelope.
- Validate event type and version.
- Check idempotency key.
- Call application use case.
- Mark idempotency success.
- Commit offset after success only.
- Send failed poison messages to DLQ after retry policy.

---

## 11. Redis Rules

Redis is allowed for:

- Cache
- Idempotency key
- Rate limit
- Short-lived lock
- Runtime session-like state

Redis is not allowed for:

- Primary business persistence
- Long-term audit log
- Replacing database transactions

Cache rule:

```text
Database is source of truth. Redis is optimization or coordination.
```

---

## 12. Logging Rules

Every external boundary must log meaningful events:

- HTTP request start/end through middleware
- Use case important business result
- Repository unexpected failure
- Kafka consume failure
- Redis unavailable when relevant

Required attributes where available:

```text
trace_id
request_id
correlation_id
tenant_id
user_id
resource_id
operation
```

Forbidden:

- Logging passwords
- Logging tokens
- Logging full request body with PII
- Using `fmt.Println`
- Swallowing errors after logging

---

## 13. OpenAPI Rules

Every API task must update:

```text
api/openapi.yaml
api/paths/*.yaml
api/components/*.yaml
```

OpenAPI must define:

- Operation ID
- Tags
- Request schema
- Response schema
- Error responses
- Security requirement if protected

Example path:

```yaml
/v1/users:
  post:
    operationId: createUser
    tags:
      - Users
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreateUserRequest'
    responses:
      '201':
        description: User created
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserResponse'
      '400':
        $ref: '#/components/responses/BadRequest'
      '409':
        $ref: '#/components/responses/Conflict'
      '500':
        $ref: '#/components/responses/InternalServerError'
```

---

## 14. Unit Test Rules for AI Agents

When implementing or modifying a file, AI agents must also implement or update the corresponding test file.

Required test cases:

- Happy path
- Invalid input
- Dependency returns error
- Boundary condition
- Context cancellation where relevant
- Idempotency behavior where relevant

For every new function, add at least one direct test or prove it is covered through an existing table test.

---

## 15. Review Checklist

Before marking a task complete, verify:

```text
[ ] OpenAPI updated
[ ] Handler implemented
[ ] Use case implemented
[ ] Domain rules implemented
[ ] Repository interface added/updated
[ ] Infrastructure implementation added/updated
[ ] Migration added if DB changed
[ ] Unit test per source file added/updated
[ ] Integration test added if infrastructure behavior changed
[ ] Fx module updated
[ ] Logging added at boundary/failure points
[ ] Error mapping added/updated
[ ] gofmt/go vet/golangci-lint pass
[ ] go test ./... -race -count=1 pass
```

---

## 16. Forbidden Shortcuts

Do not do these:

```text
- Put everything in main.go
- Put SQL inside HTTP handler
- Return gorm model as JSON
- Use panic for normal error flow
- Ignore errors from Kafka/Redis/Postgres
- Add endpoint without OpenAPI
- Add source file without test file
- Use global DB/Redis/Kafka clients
- Use Fx as service locator
- Skip context.Context
- Create large generic utils package
- Mix multiple unrelated features in one task
```

---

## 17. Done Means Done

A feature is done only when:

1. It compiles.
2. It follows clean architecture boundaries.
3. It has OpenAPI contract.
4. It has unit tests per source file.
5. It has integration tests for changed infrastructure behavior.
6. It has migrations for schema changes.
7. It has structured logs.
8. It passes lint and race tests.
9. It can be reviewed by another AI or human without hidden assumptions.

## 18. Git Rules

### Conventional Commit

All commits must follow:

<type>(<scope>): <summary>

Allowed types:

- feat
- fix
- refactor
- test
- docs
- chore
- ci

Examples:

feat(auth): add jwt validation
fix(cache): prevent redis memory leak

### Task Tracking

Each commit must reference exactly one task.

Example:

feat(employee): EMP-001 create employee repository

### Commit Size

A commit should contain changes for only one task.

Do not combine multiple tasks into a single commit.