# Tech Stack

## 1. Project Target

This project is a Golang backend service using:

- **Language:** Go
- **HTTP framework:** Echo
- **API contract:** OpenAPI / Swagger
- **Database:** PostgreSQL
- **ORM:** GORM
- **Cache / distributed primitives:** Redis
- **Logging:** `log/slog`
- **Dependency Injection:** Uber Fx
- **Message broker:** Kafka with Sarama
- **Architecture:** Clean Architecture
- **API style:** RESTful APIs
- **Testing:** Unit test per source file, integration tests for infrastructure boundaries

The service must be designed as a production-grade backend service, not as a demo application. The codebase must be easy to split into multiple services later.

---

## 2. Architecture Style

Use **Clean Architecture** with strict dependency direction:

```text
HTTP / Kafka / CLI / Cron
        ↓
Application Use Cases
        ↓
Domain Models + Domain Interfaces
        ↑
Infrastructure Implementations
```

Dependency rule:

```text
adapter -> application -> domain
infrastructure -> domain/application interfaces
platform -> shared runtime wiring only
```

Forbidden dependency direction:

```text
domain must not import Echo, GORM, Redis, Sarama, Fx, slog, or config packages
application must not import Echo or GORM concrete models
handler must not contain business logic
repository must not contain business workflow
```

---

## 3. Go Runtime

Use Go modules.

Recommended baseline:

```go
module github.com/<org>/<service-name>

go 1.23
```

Rules:

- Use `context.Context` as the first argument for I/O and business operations.
- Do not use package-level mutable global variables.
- Do not hide errors with `_` unless explicitly justified.
- Do not use reflection-heavy abstractions unless necessary.
- Prefer small interfaces owned by the consumer package.

---

## 4. HTTP Framework: Echo

Echo is used only in the HTTP adapter layer.

Responsibilities:

- Routing
- Request binding
- Request validation trigger
- Response serialization
- Middleware composition
- Error mapping

Echo must not appear in domain or application layer.

Example handler shape:

```go
type UserHandler struct {
    svc UserUseCase
    log *slog.Logger
}

func NewUserHandler(svc UserUseCase, log *slog.Logger) *UserHandler {
    return &UserHandler{svc: svc, log: log}
}

func (h *UserHandler) Create(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return NewHTTPError(http.StatusBadRequest, "invalid_request", err)
    }

    input := CreateUserInput{
        Email: req.Email,
        Name:  req.Name,
    }

    result, err := h.svc.Create(c.Request().Context(), input)
    if err != nil {
        return err
    }

    return c.JSON(http.StatusCreated, ToUserResponse(result))
}
```

---

## 5. API Contract: OpenAPI / Swagger

Use **OpenAPI-first** as the primary contract.

Recommended location:

```text
api/openapi.yaml
api/components/*.yaml
api/paths/*.yaml
```

Rules:

- Every public REST endpoint must be declared in OpenAPI.
- Every request body must have a schema.
- Every response must have a schema.
- Every error response must use the standard error format.
- OpenAPI must be updated in the same task as handler changes.
- Swagger UI may be exposed in non-production or protected environments only.

Standard error response:

```yaml
ErrorResponse:
  type: object
  required:
    - code
    - message
  properties:
    code:
      type: string
      example: validation_error
    message:
      type: string
      example: Request validation failed
    details:
      type: object
      additionalProperties: true
    trace_id:
      type: string
      example: 01HZX3QK0YS6M2Z1WD9S7VN9VJ
```

Allowed implementation options:

1. **OpenAPI-first:** maintain `api/openapi.yaml`, optionally generate DTOs/stubs.
2. **Swagger annotations:** allowed only if the generated document is committed and reviewed.

Do not maintain two conflicting API contracts.

---

## 6. Database: PostgreSQL

PostgreSQL is the primary transactional store.

Use cases:

- Core business entities
- Transactional consistency
- Relational constraints
- Audit references
- Idempotency keys
- Outbox table if needed

Rules:

- Use UUID or ULID for public IDs.
- Use database constraints for uniqueness and foreign keys where applicable.
- Every table must include `created_at` and `updated_at`.
- Soft delete is allowed only when the business explicitly needs recovery/history.
- Do not expose database IDs directly if public IDs are required.
- Use migrations. Do not rely on GORM AutoMigrate in production.

Common table fields:

```sql
id uuid primary key,
created_at timestamptz not null default now(),
updated_at timestamptz not null default now(),
deleted_at timestamptz null
```

---

## 7. ORM: GORM

GORM is used only inside repository implementations.

Allowed package:

```text
internal/infrastructure/postgres
internal/infrastructure/repository
```

Rules:

- Do not return GORM models from repositories.
- Repositories must map between database models and domain models.
- All queries must receive `context.Context`.
- Use explicit `Select`, `Where`, `Order`, and `Limit` for non-trivial queries.
- Avoid hidden preloads. Preload only when the use case requires it.
- Use transactions through an application-facing transaction manager.

Example repository method:

```go
func (r *UserGormRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var model UserModel
    err := r.db.WithContext(ctx).
        Where("email = ?", email).
        First(&model).Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, domain.ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("find user by email: %w", err)
    }

    user := model.ToDomain()
    return &user, nil
}
```

---

## 8. Redis

Redis is used for cache, distributed locks, rate limiting, idempotency, and short-lived runtime state.

Rules:

- Redis must not be a source of truth for business records.
- Every cache key must have an owner, namespace, and TTL.
- Cache invalidation must be explicit.
- Use JSON only when schema evolution is manageable.
- Avoid storing large blobs.

Key format:

```text
<service>:<module>:<entity>:<id>
```

Example:

```text
user-svc:user:profile:01HZX3QK0YS6M2Z1WD9S7VN9VJ
```

Recommended cache interface:

```go
type Cache interface {
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
}
```

---

## 9. Logging: slog

Use Go standard `log/slog`.

Rules:

- Use structured logs only.
- Never log secrets, tokens, passwords, private keys, or full PII payloads.
- Include `trace_id`, `request_id`, `tenant_id`, and `user_id` where available.
- Use `Info` for business milestones.
- Use `Warn` for recoverable abnormal states.
- Use `Error` for failed operations requiring attention.
- Do not use `fmt.Println` in application code.

Example:

```go
log.ErrorContext(ctx, "failed to create user",
    slog.String("email", input.Email),
    slog.String("error", err.Error()),
)
```

---

## 10. Dependency Injection: Uber Fx

Uber Fx is used for application composition only.

Allowed location:

```text
internal/app
internal/bootstrap
```

Rules:

- Fx modules wire dependencies.
- Business logic must not depend on Fx.
- Constructors must be explicit.
- Avoid service locator patterns.
- Avoid hidden global singleton state.

Example module:

```go
var UserModule = fx.Module("user",
    fx.Provide(
        application.NewUserService,
        repository.NewUserGormRepository,
        http.NewUserHandler,
    ),
    fx.Invoke(http.RegisterUserRoutes),
)
```

---

## 11. Kafka: Sarama

Kafka is used for asynchronous integration and event-driven workflows.

Rules:

- Use Sarama only in infrastructure/message packages.
- Application layer publishes domain/application events through interfaces.
- Consumers must be idempotent.
- Commit offsets only after successful processing.
- Include correlation ID and event ID in every message.
- Message handlers must support retry and dead-letter strategy.

Event envelope:

```go
type EventEnvelope[T any] struct {
    ID            string    `json:"id"`
    Type          string    `json:"type"`
    Source        string    `json:"source"`
    Version       int       `json:"version"`
    CorrelationID string    `json:"correlation_id"`
    OccurredAt    time.Time `json:"occurred_at"`
    Data          T         `json:"data"`
}
```

Producer interface:

```go
type EventPublisher interface {
    Publish(ctx context.Context, topic string, key string, event any) error
}
```

---

## 12. Configuration

Configuration must be loaded from environment variables, config files, or secret manager.

Rules:

- No hardcoded credentials.
- No direct `os.Getenv` scattered across the codebase.
- Load config once at bootstrap.
- Validate config at startup.

Example config structure:

```go
type Config struct {
    App      AppConfig
    HTTP     HTTPConfig
    Postgres PostgresConfig
    Redis    RedisConfig
    Kafka    KafkaConfig
}
```

---

## 13. Recommended Dependencies

Core dependencies:

```text
github.com/labstack/echo/v4
gorm.io/gorm
gorm.io/driver/postgres
github.com/redis/go-redis/v9
go.uber.org/fx
github.com/IBM/sarama
```

Testing dependencies:

```text
github.com/stretchr/testify
go.uber.org/goleak
github.com/DATA-DOG/go-sqlmock
github.com/testcontainers/testcontainers-go
```

OpenAPI/Swagger options:

```text
github.com/swaggo/echo-swagger
github.com/swaggo/swag
github.com/oapi-codegen/oapi-codegen/v2
```

Use only one API contract workflow per service.
