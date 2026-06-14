# Base Source Init Package

## Objective
To bootstrap the foundational backend codebase for the HROS Super Admin Portal. This initialization provides a robust, production-ready scaffold using Golang, Uber Fx, and Clean Architecture principles. It establishes all required infrastructure adapters, API routing, logging, error handling, and testing patterns without implementing any business or domain logic.

## Architecture Constraints
*   **Clean Architecture**: Strict dependency flow must be maintained: `infrastructure` and `adapter` depend on `application`, which depends on `domain`.
*   **Dependency Injection**: Use Uber Fx for all wiring (`fx.Module`, `fx.Provide`, `fx.Invoke`). Fx must not leak into business logic.
*   **HTTP Layer**: Echo is restricted to `internal/adapter/http/`. Handlers only bind, validate, call use cases, and serialize responses.
*   **Persistence Layer**: GORM is restricted to `internal/infrastructure/database/`. Repositories map DB models to domain entities and must accept `context.Context`.
*   **Error Handling**: Centralized mapping of domain errors to HTTP statuses in the HTTP adapter layer.
*   **Testing**: Unit tests are required per source file. External infrastructure requires integration tests.

## Repository Structure
The scaffolding will adhere to the following directory structure:
*   `api/` - OpenAPI definitions.
*   `cmd/api/` - Application entrypoint.
*   `internal/domain/` - Pure domain interfaces and error types.
*   `internal/application/` - Use case interfaces.
*   `internal/adapter/http/` - Echo server, routing, and middleware.
*   `internal/adapter/kafka/` - Sarama consumer handlers.
*   `internal/infrastructure/` - Concrete implementations for DB, Redis, and Kafka.
*   `internal/platform/` - Cross-cutting concerns (config, logging).
*   `migrations/` - SQL migration scripts.
*   `test/` - Integration testing fixtures.

## Init Tasks

### Task ID
TSK-INIT-001

### Task Name
Project Scaffolding & Fx Bootstrapping

### Goal
Establish the Go module, directory tree, and the root Uber Fx application lifecycle.

### Scope
Create the fundamental directory structure and the `main.go` entrypoint using Uber Fx for graceful startup and shutdown. 

### Files Expected To Create/Modify
*   `go.mod`
*   `go.sum`
*   `cmd/api/main.go`
*   `cmd/api/main_test.go`
*   `Makefile`

### Dependencies
*   None.

### Acceptance Criteria
*   The Go module is initialized (`go mod init`).
*   The `cmd/api/main.go` file contains a valid `fx.New` setup and invokes an empty lifecycle hook.
*   The application starts, runs, and shuts down cleanly on SIGINT/SIGTERM.
*   `Makefile` includes targets for `build`, `test`, `lint`, and `run`.

### Unit Test Requirements
*   `main_test.go` uses `fx.ValidateApp` to ensure the dependency graph is valid without starting the server.

### Integration Test Requirements
*   None.

### Out of Scope
*   Server implementation, configuration loading, or any business logic.

---

### Task ID
TSK-INIT-002

### Task Name
Platform Configuration Loader

### Goal
Implement a configuration loader that reads from environment variables safely and makes the configuration available to the Fx dependency graph.

### Scope
Create a strongly typed configuration struct and a loader function in the `platform` layer.

### Files Expected To Create/Modify
*   `internal/platform/config/config.go`
*   `internal/platform/config/config_test.go`
*   `internal/platform/config/module.go` (Fx module definition)

### Dependencies
*   TSK-INIT-001

### Acceptance Criteria
*   Configuration is read from environment variables.
*   Includes base fields: `AppEnv`, `Port`, `DatabaseURL`, `RedisURL`, `KafkaBrokers`.
*   Application fails to start (panics or returns Fx error) if required variables are missing.
*   Provided to Fx via `fx.Provide`.

### Unit Test Requirements
*   Test successful loading with valid environment variables.
*   Test expected failure when required variables are missing.

### Integration Test Requirements
*   None.

### Out of Scope
*   External secret managers (e.g., AWS Secrets Manager/Vault).

---

### Task ID
TSK-INIT-003

### Task Name
Structured Logging (Slog) Component

### Goal
Standardize JSON logging across the platform using `log/slog`.

### Scope
Implement an Fx-provided slog logger that automatically formats output as JSON and includes request-scoped attributes.

### Files Expected To Create/Modify
*   `internal/platform/logger/logger.go`
*   `internal/platform/logger/logger_test.go`
*   `internal/platform/logger/module.go`

### Dependencies
*   TSK-INIT-002 (Config to determine log level)

### Acceptance Criteria
*   Logger is configured for JSON output.
*   Logger includes helper functions to extract `trace_id` and `request_id` from `context.Context`.
*   Never logs passwords or PII payloads.

### Unit Test Requirements
*   Verify JSON formatting.
*   Verify extraction of context variables into log attributes.

### Integration Test Requirements
*   None.

### Out of Scope
*   Log forwarding agents (e.g., Fluentd, Datadog).

---

### Task ID
TSK-INIT-004

### Task Name
Domain Error Types & Centralized Mapping

### Goal
Define standard domain errors and implement the centralized Echo middleware to map these errors to HTTP status codes.

### Scope
Create reusable domain error structs (NotFound, Validation, Conflict, Unauthorized, Forbidden) and an Echo HTTP error handler to format them.

### Files Expected To Create/Modify
*   `internal/domain/errors/errors.go`
*   `internal/domain/errors/errors_test.go`
*   `internal/adapter/http/middleware/error_handler.go`
*   `internal/adapter/http/middleware/error_handler_test.go`

### Dependencies
*   TSK-INIT-003 (Logging)

### Acceptance Criteria
*   Domain errors implement the standard `error` interface and allow wrapping.
*   The Echo error handler intercepts these errors and outputs the standard `{"error": {"code", "message", "details"}}` JSON envelope.
*   HTTP status mapping adheres strictly to the conventions (e.g., NotFound -> 404, Conflict -> 409).

### Unit Test Requirements
*   Test mapping of each domain error type to its corresponding HTTP status code and JSON envelope.
*   Test fallback behavior for unknown internal errors (500).

### Integration Test Requirements
*   None.

### Out of Scope
*   Business-specific error codes (e.g., `TENANT_NOT_FOUND`).

---

### Task ID
TSK-INIT-005

### Task Name
Echo Server & OpenAPI Foundation

### Goal
Initialize the Echo HTTP server framework, establish base middleware, and create the OpenAPI definition file.

### Scope
Provide the Echo instance via Fx, configure standard middleware (Recover, Request ID, CORS), and set up the `api/openapi.yaml` contract.

### Files Expected To Create/Modify
*   `api/openapi.yaml`
*   `internal/adapter/http/server.go`
*   `internal/adapter/http/server_test.go`
*   `internal/adapter/http/module.go`

### Dependencies
*   TSK-INIT-002, TSK-INIT-003, TSK-INIT-004

### Acceptance Criteria
*   `api/openapi.yaml` contains standard error response schemas and global headers.
*   Echo server starts on the port defined in Config and is wired to the Fx lifecycle `OnStart` and `OnStop` hooks.
*   Base middlewares (including the error handler from TSK-INIT-004) are registered.

### Unit Test Requirements
*   Test server initialization without panic.
*   Validate `openapi.yaml` syntax.

### Integration Test Requirements
*   None.

### Out of Scope
*   Registering any business route handlers.

---

### Task ID
TSK-INIT-006

### Task Name
PostgreSQL & GORM Infrastructure

### Goal
Establish relational database connectivity, connection pooling, and the transaction management foundation.

### Scope
Initialize GORM with the PostgreSQL driver, define the application-facing Transaction Manager interface, and provide it via Fx.

### Files Expected To Create/Modify
*   `internal/application/interfaces/transaction.go`
*   `internal/infrastructure/database/postgres.go`
*   `internal/infrastructure/database/transaction_manager.go`
*   `internal/infrastructure/database/postgres_test.go`
*   `internal/infrastructure/database/module.go`

### Dependencies
*   TSK-INIT-002, TSK-INIT-003

### Acceptance Criteria
*   GORM connects to Postgres using credentials from Config.
*   GORM `AutoMigrate` is explicitly disabled.
*   `TransactionManager` interface exposes `RunInTransaction(ctx, func(ctx) error) error`.

### Unit Test Requirements
*   Test provider initialization and error handling for bad connection strings.

### Integration Test Requirements
*   `postgres_test.go` must use `testcontainers` to spin up a real PostgreSQL instance and verify ping/connection success.

### Out of Scope
*   Any database models, repository implementations, or schema migrations.

---

### Task ID
TSK-INIT-007

### Task Name
Redis Cache Infrastructure

### Goal
Initialize the distributed cache and locking store using `go-redis`.

### Scope
Create the Redis connection provider and a base Cache interface in the application layer implemented by the infrastructure layer.

### Files Expected To Create/Modify
*   `internal/application/interfaces/cache.go`
*   `internal/infrastructure/cache/redis.go`
*   `internal/infrastructure/cache/redis_test.go`
*   `internal/infrastructure/cache/module.go`

### Dependencies
*   TSK-INIT-002, TSK-INIT-003

### Acceptance Criteria
*   Redis connects using Config variables.
*   Cache implementation supports standard `Set(ctx, key, val, TTL)` and `Get(ctx, key)` methods.
*   Cache failures gracefully degrade (log an error but do not crash the application).

### Unit Test Requirements
*   Test cache interface methods using a mock Redis client.
*   Test graceful degradation when Redis returns an error.

### Integration Test Requirements
*   Use `testcontainers` to verify actual Set/Get/TTL behavior against a real Redis instance.

### Out of Scope
*   Specific cache keys (e.g., Dashboard KPIs or rate limiters).

---

### Task ID
TSK-INIT-008

### Task Name
Kafka Producer & Consumer Infrastructure

### Goal
Provide asynchronous event-driven communication utilities using the Sarama library.

### Scope
Create the Kafka connection provider, define the standard Event Envelope struct, the Publisher interface, and the Consumer handler skeleton.

### Files Expected To Create/Modify
*   `internal/application/interfaces/publisher.go`
*   `internal/domain/events/envelope.go`
*   `internal/infrastructure/message/sarama_producer.go`
*   `internal/infrastructure/message/sarama_producer_test.go`
*   `internal/infrastructure/message/module.go`

### Dependencies
*   TSK-INIT-002, TSK-INIT-003

### Acceptance Criteria
*   `envelope.go` defines the required `event_id`, `correlation_id`, `type`, and `version` fields.
*   Sarama SyncProducer is configured and provided to Fx.
*   Publisher interface implements `Publish(ctx, topic, key, envelope)` and properly serializes the envelope.

### Unit Test Requirements
*   Use Sarama's built-in `mocks` package to test successful message production and proper JSON encoding of the envelope.
*   Test Sarama error mapping.

### Integration Test Requirements
*   None.

### Out of Scope
*   Concrete event definitions (e.g., `tenant.created`) or actual consumer business logic implementation.