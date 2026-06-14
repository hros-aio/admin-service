# HROS Admin Service

Backend service for the Human Resources Operating System (HROS) Admin module.

## Architecture

This project follows **Clean Architecture** principles and uses **Uber Fx** for dependency injection.

### Directory Structure

- `cmd/api`: Application entry point.
- `internal/app`: Fx module definitions and application bootstrap.
- `internal/config`: Configuration management.
- `internal/platform`: Infrastructure adapters (HTTP, Database, Redis, Kafka).
- `internal/shared`: Shared utilities, response formats, and middleware.
- `docs/openapi`: API specifications.
- `migrations`: Database migrations.

## Getting Started

Refer to the [Quickstart Guide](specs/001-backend-foundation-bootstrap/quickstart.md) for local development setup.

## Development

### Prerequisites
- Go 1.23+
- Docker and Docker Compose
- Makefile
- golangci-lint

### Common Commands
- `make run`: Start the application locally.
- `make test`: Run unit tests.
- `make lint`: Run the linter.
- `make docker-up`: Start infrastructure dependencies.
- `make migrate`: Run database migrations.
