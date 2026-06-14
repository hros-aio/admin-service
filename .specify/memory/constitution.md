<!--
Sync Impact Report:
- Version change: N/A → 1.0.0
- List of modified principles:
  - Added I. Clean Architecture & Strict Boundaries
  - Added II. Documentation-First & OpenAPI-Driven
  - Added III. Unit-Test-Per-File (NON-NEGOTIABLE)
  - Added IV. Task-Driven & Atomic Implementation
  - Added V. Observability & Structured Logging
- Added sections: Technical Stack Constraints, Development Workflow
- Templates requiring updates:
  - .specify/templates/plan-template.md (✅ aligned)
  - .specify/templates/spec-template.md (✅ aligned)
  - .specify/templates/tasks-template.md (✅ aligned)
- Follow-up TODOs: None
-->

# HROS Admin Constitution

## Core Principles

### I. Clean Architecture & Strict Boundaries
Every component must adhere to strict dependency directions: adapter/infrastructure -> application -> domain. Business logic is strictly isolated from framework and infrastructure details (Echo, GORM, Redis, Kafka). Domain must not import any infrastructure packages. Application layer must not import Echo or GORM concrete models.

### II. Documentation-First & OpenAPI-Driven
OpenAPI is the primary source of truth for the API contract. Every public REST endpoint must be declared in OpenAPI before implementation. Every request and response must have a schema. Handlers and OpenAPI must be updated together in the same task. Standard error responses are mandatory for all endpoints.

### III. Unit-Test-Per-File (NON-NEGOTIABLE)
Every non-generated .go source file must have a corresponding _test.go file. Unit tests must be fast, deterministic, and free of external dependencies (PostgreSQL, Redis, Kafka, Network). Coverage targets are mandatory: Domain 90%, Application 85%, Adapter 75%, Repository 70%.

### IV. Task-Driven & Atomic Implementation
Features must be implemented via vertical slices, one task at a time as defined in SpecKit tasks.md. Implementation must follow the mandatory workflow: OpenAPI -> Domain -> Application -> Infrastructure -> Adapter. The agent must never implement an entire feature in one pass.

### V. Observability & Structured Logging
Use standard `log/slog` for structured logging with explicit attributes. Never log secrets, tokens, or PII. Include trace_id, request_id, and tenant_id in all logs where available. Every external boundary (HTTP, Kafka, Redis) must log success and failures explicitly.

## Technical Stack Constraints

The following technology stack is non-negotiable for the HROS Admin project:
- **Language**: Go 1.23+
- **Dependency Injection**: Uber Fx
- **HTTP Framework**: Echo Framework
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Messaging**: Kafka (Sarama)
- **Observability**: Structured logging with `log/slog`
- **Architecture**: Clean Architecture with strict layer isolation

## Development Workflow

Every task execution must follow this sequence:
1. Read the current SpecKit artifacts (`spec.md`, `plan.md`, `tasks.md`).
2. Identify and implement exactly one task ID.
3. Update OpenAPI contract if the REST interface is modified.
4. Implement or update unit tests for every changed production file.
5. Verify correctness using `go test ./... -race -count=1` and `golangci-lint`.
6. Justify any architectural complexity in the Implementation Plan.

## Governance

- This Constitution takes absolute precedence over general workflows and tool defaults.
- Amendments require a version bump and an update to this document.
- All code changes must be verified against these principles before finality.
- Use `GEMINI.md` for team-shared conventions and repository-wide workflows.

**Version**: 1.0.0 | **Ratified**: 2026-06-14 | **Last Amended**: 2026-06-14
