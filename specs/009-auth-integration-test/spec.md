# Feature Specification: Auth Integration Test

**Feature Branch**: `009-auth-integration-test`

**Created**: 2026-06-20

**Status**: Draft

**Input**: User description: "TSK-AUTH-009: **Layer**: Tests **Description**: Write an integration test for the Core Authentication slice. Use `testcontainers` to spin up a real PostgreSQL database, apply migrations, seed a test admin user, and execute HTTP requests against the Echo server to verify the full Login and Logout flow. **Input**: Fully assembled Auth module, `testcontainers` PostgreSQL image. **Output**: `test/integration/auth_flow_test.go` **Definition of Done:** Integration test successfully spins up infrastructure, verifies a successful login yielding a valid JWT, and verifies a successful 204 logout response."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - E2E Admin Authentication Flow (Priority: P1) 🎯 MVP

As a system verification utility, I want to run a complete integration test for the authentication flow, so that I can ensure the database migrations, GORM repositories, UseCases, password hashing, token providers, and Echo HTTP handlers all work together correctly against a real database instance.

**Why this priority**: High. Essential for verifying that the entire system compiles, wires, and runs successfully against concrete external systems (PostgreSQL database) without mocking core dependencies.

**Independent Test**: The integration test itself is runnable via `go test ./test/integration/...` and performs the complete sequence of database initialization, schema migration, test user seeding, HTTP server start, login authentication, and logout session revocation.

**Acceptance Scenarios**:

1. **Given** a PostgreSQL container is initialized, **When** database migrations are executed, **Then** all database tables are created successfully.
2. **Given** a seeded active admin user, **When** an HTTP POST request is sent to `/v1/auth/login` with correct credentials, **Then** the server returns 200 OK and valid session tokens.
3. **Given** a successful login response, **When** an HTTP DELETE request is sent to `/v1/auth/session` with the refresh token in the Authorization header, **Then** the server revokes the session and returns 204 No Content.
4. **Given** incorrect credentials, **When** an HTTP POST request is sent to `/v1/auth/login`, **Then** the server returns 401 Unauthorized.

### Edge Cases

- **PostgreSQL Startup Failures**: If the PostgreSQL container fails to start or times out, the integration tests must abort cleanly and fail with a clear descriptive message.
- **Migration Errors**: If the SQL migration scripts contain syntax or integrity errors when run against a clean database, the test suite must halt immediately.
- **Invalid Credentials**: Ensuring that invalid authentication attempts fail correctly with a 401 Unauthorized status and return the standard JSON error envelope.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST boot a PostgreSQL database container using `testcontainers-go` before starting the integration tests.
- **FR-002**: System MUST apply all up migration scripts (`migrations/*.sql`) to the containerized PostgreSQL database.
- **FR-003**: System MUST seed a test administrator record (email: `test-admin@hros.com`, active status, valid password hash) into the `admin_users` table.
- **FR-004**: System MUST bootstrap the Echo HTTP application using concrete dependencies (GORM, repositories, audit log, token provider).
- **FR-005**: System MUST execute a real HTTP client call `POST /v1/auth/login` to authenticate the user and assert a 200 OK response containing access/refresh tokens.
- **FR-006**: System MUST execute a real HTTP client call `DELETE /v1/auth/session` with the returned refresh token and assert a 204 No Content response.
- **FR-007**: System MUST perform clean tear-down of the PostgreSQL container and Echo server after test execution completes.

### Key Entities *(include if feature involves data)*

- **TestDatabase**: Real PostgreSQL instance managed via Docker container.
- **EchoServer**: Real Echo HTTP routing server instance.
- **TestHttpClient**: HTTP client executing REST requests.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of integration test executions pass cleanly when run via `go test ./test/integration/...`.
- **SC-002**: All migration scripts execute successfully against the containerized PostgreSQL instance.
- **SC-003**: No mocks are used for database repositories, password helpers, token generation, or Echo handler controllers.
- **SC-004**: Cleanup processes run successfully, leaving no hanging container instances or memory leaks.

## Assumptions

- **Docker Environment**: The environment executing the integration tests has Docker installed and running (required for `testcontainers-go`).
- **Dependencies**: The `testcontainers-go` library will be imported and added to the project's development dependencies.
- **Migrations Path**: The SQL migration files are located in the `migrations/` directory relative to the project root and are readable by the test suite.
