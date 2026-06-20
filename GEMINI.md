# Gemini Project Instructions

You are implementing this repository using GitHub SpecKit artifacts only.

## Source of Truth

Always read these files before planning or coding:

- @docs/product/PRD.md
- @docs/engineering/tech-stack.md
- @docs/engineering/coding-conventions.md
- @docs/engineering/repository-structure.md
- @docs/engineering/testing-strategy.md
- @docs/engineering/implementation-rules.md
- @docs/engineering/foundation.md

## Architecture Rules

This project uses:

- Golang
- Echo Framework
- Swagger/OpenAPI
- Postgres with GORM
- Redis
- Slog
- Uber Fx for DI
- Kafka with Sarama
- Clean Architecture
- RESTful APIs
- Unit test per file

## Hard Rules

Do not implement a whole feature at once.

Only implement the current SpecKit task selected by the user.

Before editing code:

1. Read the current `specs/<feature>/spec.md`.
2. Read the current `specs/<feature>/plan.md`.
3. Read the current `specs/<feature>/tasks.md`.
4. Identify the exact task ID.
5. Explain the files you will modify.
6. Implement only that task.
7. Add or update unit tests for every changed production file.
8. Run formatting and tests.

## Forbidden

- Do not bypass clean architecture.
- Do not put business logic in Echo handlers.
- Do not access GORM directly from usecase layer.
- Do not access Redis directly from usecase layer.
- Do not publish Kafka events directly from handlers.
- Do not create files outside the approved repository structure.
- Do not modify unrelated files.
- Do not skip tests.