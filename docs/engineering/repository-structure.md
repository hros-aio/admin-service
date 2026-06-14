# Repository Structure

## 1. Target Structure

Use this structure for a single Golang service using Echo, OpenAPI, GORM, Redis, Slog, Uber Fx, Sarama, Clean Architecture, RESTful APIs, and unit tests.

```text
.
├── api
│   ├── openapi.yaml
│   ├── components
│   │   ├── schemas.yaml
│   │   ├── parameters.yaml
│   │   └── responses.yaml
│   └── paths
│       └── users.yaml
├── cmd
│   └── api
│       └── main.go
├── config
│   ├── config.example.yaml
│   └── local.yaml
├── deployments
│   ├── docker
│   │   └── Dockerfile
│   └── compose
│       └── docker-compose.yaml
├── docs
│   ├── tech-stack.md
│   ├── coding-conventions.md
│   ├── repository-structure.md
│   ├── testing-strategy.md
│   └── implementation-rules.md
├── internal
│   ├── app
│   │   ├── module.go
│   │   └── lifecycle.go
│   ├── bootstrap
│   │   ├── fx.go
│   │   └── logger.go
│   ├── domain
│   │   └── user
│   │       ├── entity.go
│   │       ├── entity_test.go
│   │       ├── errors.go
│   │       ├── repository.go
│   │       └── value_object.go
│   ├── application
│   │   └── user
│   │       ├── service.go
│   │       ├── service_test.go
│   │       ├── input.go
│   │       ├── output.go
│   │       └── mapper.go
│   ├── adapter
│   │   ├── http
│   │   │   ├── router.go
│   │   │   ├── middleware
│   │   │   │   ├── error_middleware.go
│   │   │   │   ├── request_id_middleware.go
│   │   │   │   └── recover_middleware.go
│   │   │   └── user
│   │   │       ├── handler.go
│   │   │       ├── handler_test.go
│   │   │       ├── request.go
│   │   │       ├── response.go
│   │   │       └── mapper.go
│   │   └── kafka
│   │       └── user
│   │           ├── consumer.go
│   │           ├── consumer_test.go
│   │           └── handler.go
│   ├── infrastructure
│   │   ├── postgres
│   │   │   ├── client.go
│   │   │   ├── transaction.go
│   │   │   └── transaction_test.go
│   │   ├── repository
│   │   │   └── user
│   │   │       ├── model.go
│   │   │       ├── mapper.go
│   │   │       ├── repository.go
│   │   │       └── repository_test.go
│   │   ├── redis
│   │   │   ├── client.go
│   │   │   ├── cache.go
│   │   │   └── cache_test.go
│   │   └── kafka
│   │       ├── producer.go
│   │       ├── producer_test.go
│   │       ├── consumer_group.go
│   │       └── envelope.go
│   └── platform
│       ├── config
│       │   ├── config.go
│       │   └── config_test.go
│       ├── validator
│       │   ├── validator.go
│       │   └── validator_test.go
│       └── clock
│           ├── clock.go
│           └── clock_test.go
├── migrations
│   ├── 000001_create_users.up.sql
│   └── 000001_create_users.down.sql
├── test
│   ├── integration
│   │   ├── postgres_test.go
│   │   ├── redis_test.go
│   │   └── kafka_test.go
│   └── testutil
│       ├── logger.go
│       ├── fixture.go
│       └── mock.go
├── .env.example
├── .gitignore
├── Makefile
├── go.mod
└── go.sum
```

---

## 2. Directory Responsibilities

### `api/`

Contains OpenAPI contract.

Rules:

- Must be updated when REST API changes.
- Must define request/response schemas.
- Must define standard error responses.
- Must be reviewed with handler changes.

---

### `cmd/api/`

Application entrypoint.

Responsibilities:

- Start Fx application
- Load bootstrap module
- No business logic
- No route definitions

Example:

```go
package main

import "github.com/<org>/<service-name>/internal/app"

func main() {
    app.New().Run()
}
```

---

### `internal/domain/`

Pure business domain.

Allowed:

- Entities
- Value objects
- Domain errors
- Domain repository interfaces
- Domain services if needed

Forbidden:

- Echo
- GORM
- Redis
- Kafka
- Fx
- HTTP DTOs
- Database models
- Config loading

Example:

```go
package user

type Repository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id ID) (*User, error)
    FindByEmail(ctx context.Context, email Email) (*User, error)
}
```

---

### `internal/application/`

Application use cases.

Responsibilities:

- Orchestrate business flow
- Call domain repositories
- Control transactions
- Publish events through interfaces
- Perform authorization checks if needed

Forbidden:

- Echo context
- GORM model
- Redis concrete client
- Sarama concrete client

Example:

```go
type UserService struct {
    repo domain.UserRepository
    tx   TxManager
    log  *slog.Logger
}
```

---

### `internal/adapter/http/`

HTTP adapter using Echo.

Responsibilities:

- Register routes
- Bind request DTOs
- Validate request DTOs
- Map HTTP requests to application inputs
- Map application outputs to HTTP responses
- Convert application/domain errors into HTTP responses through middleware

Forbidden:

- Business rules
- SQL queries
- Kafka publishing directly from handlers

---

### `internal/adapter/kafka/`

Kafka message handlers.

Responsibilities:

- Decode event envelope
- Validate event type/version
- Call application use cases
- Return processing result

Forbidden:

- Business logic inside consumer handlers
- Direct DB writes bypassing use cases

---

### `internal/infrastructure/`

Concrete implementation of external technologies.

Contains:

- GORM repositories
- PostgreSQL connection
- Redis client and cache implementation
- Sarama producer/consumer implementation
- Transaction manager

Rules:

- Infrastructure depends on domain/application interfaces.
- Infrastructure must not define business use cases.

---

### `internal/platform/`

Cross-cutting technical components.

Allowed:

- Config loader
- Validator wrapper
- Clock abstraction
- ID generator
- Observability helpers

Rules:

- Keep platform packages small.
- Do not put business logic here.
- Avoid becoming a `utils` dumping ground.

---

### `migrations/`

Database migration scripts.

Rules:

- Every schema change must have up/down migration.
- Migration files must be deterministic.
- Do not rely on GORM AutoMigrate in production.

---

### `test/`

Shared test helpers and integration tests.

Rules:

- Unit tests stay next to source files.
- Integration tests stay under `test/integration`.
- Shared fixtures stay under `test/testutil`.

---

## 3. Dependency Direction

Allowed imports:

```text
adapter/http        -> application, domain
adapter/kafka       -> application, domain
application         -> domain
infrastructure      -> application interfaces, domain
bootstrap/app       -> all layers for wiring
cmd/api             -> internal/app only
```

Forbidden imports:

```text
domain              -> application, adapter, infrastructure, platform runtime
application         -> adapter/http, adapter/kafka, infrastructure/postgres
adapter/http        -> gorm, sarama, go-redis
```

---

## 4. Example User Module

### Domain

```text
internal/domain/user/
├── entity.go
├── entity_test.go
├── errors.go
├── repository.go
└── value_object.go
```

### Application

```text
internal/application/user/
├── service.go
├── service_test.go
├── input.go
├── output.go
└── mapper.go
```

### HTTP Adapter

```text
internal/adapter/http/user/
├── handler.go
├── handler_test.go
├── request.go
├── response.go
└── mapper.go
```

### Repository Implementation

```text
internal/infrastructure/repository/user/
├── model.go
├── mapper.go
├── repository.go
└── repository_test.go
```

---

## 5. Fx Module Layout

Root app module:

```go
package app

import (
    "go.uber.org/fx"
)

func New() *fx.App {
    return fx.New(
        BootstrapModule,
        PlatformModule,
        InfrastructureModule,
        ApplicationModule,
        AdapterModule,
    )
}
```

Feature module:

```go
var UserModule = fx.Module("user",
    fx.Provide(
        userapp.NewService,
        userrepo.NewRepository,
        userhttp.NewHandler,
    ),
    fx.Invoke(userhttp.RegisterRoutes),
)
```

Rules:

- Use `fx.Module` to group dependencies by concern.
- Use `fx.Provide` for constructors.
- Use `fx.Invoke` only for startup registration.
- Do not put business logic in Fx modules.

---

## 6. Route Registration

Routes must be registered in adapter/http packages.

```go
func RegisterRoutes(e *echo.Echo, h *UserHandler) {
    g := e.Group("/v1/users")

    g.POST("", h.Create)
    g.GET("/:id", h.GetByID)
    g.PATCH("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

Rules:

- Do not register routes in `main.go`.
- Do not instantiate handlers manually in route files.
- All handlers must be injected through Fx.

---

## 7. Migration Naming

Use sequential migration files:

```text
000001_create_users.up.sql
000001_create_users.down.sql
000002_create_user_idempotency_keys.up.sql
000002_create_user_idempotency_keys.down.sql
```

Migration rule:

```text
No schema-changing PR/task is complete without migration files.
```

---

## 8. Test File Placement

Unit tests stay beside the file being tested:

```text
service.go
service_test.go
repository.go
repository_test.go
handler.go
handler_test.go
```

Integration tests stay under:

```text
test/integration/
```

Generated files do not require direct unit tests.

---

## 9. Generated Code

Generated code must be placed in predictable locations:

```text
internal/generated/openapi/
```

Rules:

- Do not manually edit generated files.
- Generated files must include a header.
- Regeneration command must be available in Makefile.

Example:

```makefile
openapi-generate:
	oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml
```
