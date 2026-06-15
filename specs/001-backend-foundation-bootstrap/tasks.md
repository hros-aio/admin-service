# Tasks: Backend Foundation Bootstrap

**Input**: Design documents from `/specs/001-backend-foundation-bootstrap/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Unit tests per file are MANDATORY as per HROS Admin Constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1: Developer Bootstrap, US2: Health Check, US3: Quality Guardrails)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go 1.23 module in repo root (`go mod init github.com/<org>/<service-name>`)
- [x] T002 Create repository directory structure per `plan.md` (cmd, internal/app, internal/config, internal/platform/*, internal/shared/*, docs/openapi, migrations, test)
- [x] T003 [P] Create `.gitignore` for Go, environment files, and IDE settings
- [x] T004 [P] Create `.env.example` with mandatory configuration keys from `data-model.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T005 Implement configuration loader and validation in `internal/config/config.go`
- [x] T006 [P] Implement structured logger using `log/slog` in `internal/platform/logger/logger.go`
- [x] T007 Implement Uber Fx application bootstrap in `internal/app/app.go` and `cmd/api/main.go`
- [x] T008 [P] Implement standard JSON response and error formats in `internal/shared/response/` and `internal/shared/errors/`
- [x] T009 [P] Implement Echo HTTP server initialization and middleware (RequestID, Logger, Recovery) in `internal/platform/http/server.go` and `internal/shared/middleware/`
- [x] T010 Implement Postgres connection (GORM) and TxManager in `internal/platform/database/`
- [x] T011 [P] Implement Redis client initialization in `internal/platform/redis/redis.go`
- [x] T012 [P] Implement Kafka Sarama producer/consumer base in `internal/platform/kafka/`
- [x] T013 Implement GORM migration runner in `internal/platform/migration/migration.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Developer Bootstrap (Priority: P1) 🎯 MVP

**Goal**: Reproducible local environment with all infrastructure

**Independent Test**: Run `docker-compose up -d` and verify services are reachable.

- [x] T014 [US1] Create `docker-compose.yaml` for Postgres, Redis, and Kafka in repo root
- [x] T015 [US1] Create initial migration for schema verification in `migrations/000001_init.up.sql`
- [x] T016 [US1] Create `Makefile` with targets for `run`, `test`, `migrate`, `lint`, and `docker-up`
- [x] T017 [US1] Write unit tests for infrastructure initialization (DB, Redis, Kafka) in their respective directories

**Checkpoint**: Developer environment is fully functional

---

## Phase 4: User Story 2 - Health and Connectivity Check (Priority: P1)

**Goal**: Verify system availability and dependency status

**Independent Test**: `curl http://localhost:8080/health` returns 200 OK with dependency status

- [x] T018 [US2] Implement health check handler in `internal/platform/http/health_handler.go`
- [x] T019 [US2] Integrate health check with DB, Redis, and Kafka pings in `internal/platform/http/health_handler.go`
- [x] T020 [US2] Register health route in Echo server in `internal/platform/http/server.go`
- [x] T021 [US2] Write unit tests for health handler in `internal/platform/http/health_handler_test.go`
- [x] T022 [US2] Update `docs/openapi/openapi.yaml` with the health check endpoint (per `contracts/openapi.yaml`)

**Checkpoint**: System health monitoring is operational

---

## Phase 5: User Story 3 - Automated Quality Guardrails (Priority: P2)

**Goal**: CI/CD integration and linting

**Independent Test**: CI pipeline passes on push

- [x] T023 [US3] Configure `golangci-lint` settings in `.golangci.yml`
- [x] T024 [US3] Create GitHub Actions workflow for linting and testing in `.github/workflows/ci.yml`
- [x] T025 [US3] Add `make lint` target to `Makefile` and verify local execution

**Checkpoint**: Code quality and CI guardrails are in place

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T026 [P] Add README.md with project overview and link to `quickstart.md`
- [x] T027 [P] Ensure all files have corresponding `_test.go` files per Constitution
- [x] T028 Run `quickstart.md` validation scenarios end-to-end
- [x] T029 [P] Update `docs/openapi/` with full Swagger UI integration

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup (T001-T004)
- **User Stories (Phase 3+)**: All depend on Foundational (Phase 2) completion
- **Polish (Final Phase)**: Depends on all user stories

### Parallel Opportunities

- T003, T004 (Setup)
- T006, T008, T009, T011, T012 (Foundational - different platform packages)
- T016, T017 (Developer Bootstrap - independent utilities)
- T023, T024 (CI setup)
- All Polish tasks marked [P]

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 & 2 (Foundation)
2. Complete Phase 3 (Developer Bootstrap)
3. Validate with `make run`

### Incremental Delivery

1. Foundation -> Developer Bootstrap -> Health Check -> Quality Guardrails
2. Each story is independently testable and adds observable value.
