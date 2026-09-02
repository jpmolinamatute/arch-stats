---
trigger: model_decision
description: Project Overview
---

# Arch-Stats – AI Coding Agent Guide

Audience: AI coding agents
Repo: <https://github.com/jpmolinamatute/arch-stats>
Last updated: 2026-09-02

## Big Picture

- Two surfaces: Backend (Go 1.27+, Chi router, pgx/v5, PostgreSQL 17) and Frontend (Vue 3 + Vite SPA). FE build
  outputs go into `backend/internal/handler/static/dist/` (or embedded via `//go:embed`) and are served by Chi.
- Data flow (shot lifecycle): sensor/bot inserts rows → Postgres NOTIFY → WebSocket hub (`internal/websocket`)
  → frontend renders live updates. Keep flow unidirectional and components loosely coupled.

## Non‑Negotiables

- DB access: `pgx/v5` only with `pgxpool.Pool` (no ORM, no sync DB). Queries use parameterized SQL ($1, $2) or Squirrel query builder.
- Go structs with JSON tags for requests/responses; validate input explicitly.
- Strict error handling: wrap errors with `%w` or `apperror.Wrap`; use sentinel errors in `internal/apperror`. Never swallow errors.
- Formatting & Linting: Go (`golangci-lint`, `gofumpt`), JS/TS (`ESLint`, `Prettier`), Bash (`shellcheck`, `shfmt`). Keep diffs minimal.
- FE/BE separation: FE consumes generated TypeScript types; do not import backend internals into FE.
- Follow Effective Go conventions; reference the `backend-go-coding` skill.

## Repo Map (targets to read first)

- Backend app entrypoint: `backend/cmd/arch-stats/main.go`.
- Configuration & logging: `backend/internal/config/`.
- Domain models: `backend/internal/model/`.
- Database access: `backend/internal/repository/`.
- Business logic: `backend/internal/service/`.
- HTTP handlers & routing: `backend/internal/handler/`.
- HTTP middleware: `backend/internal/middleware/`.
- WebSocket real-time: `backend/internal/websocket/`.
- Application errors: `backend/internal/apperror/`.
- DB migrations: `backend/migrations/*.sql` executed via Goose (`goose postgres "<DSN>" up`).
- Frontend app: `frontend/src/App.vue`, `frontend/src/composables/`, `frontend/src/api/`, and generated
  types in `frontend/src/types/`.
- Scripts: `scripts/linting.bash`, `scripts/generate_fe_types.bash`, etc.

## Core Workflows

- Start services (Docker + API + FE):
  - Start Docker Compose:
    ```bash
    docker compose -f docker/docker-compose.yaml up -d
    ```
  - Start Go Server (dev live-reload):
    ```bash
    cd backend && air
    ```
  - Start Frontend Vite Server:
    ```bash
    cd frontend && npm install && npm run dev
    ```

- Run Go tests:
  ```bash
  cd backend && go test -race ./... -v
  ```

- Run Go linting & formatting:
  ```bash
  cd backend && golangci-lint run ./...
  # or full suite:
  ./scripts/linting.bash --go
  ```

- Generate FE API types when backend models change:
  ```bash
  cd frontend && npm run generate:types
  ```

## Backend Patterns

- Handlers: Accept `http.ResponseWriter` and `*http.Request`; parse JSON; call service layer; serialize JSON responses using standard helpers.
- Service Layer: Coordinates business logic, interacts with repositories, returns domain models and `apperror` sentinel errors.
- Repositories: Accept `context.Context` as first parameter; use `*pgxpool.Pool` or `pgx.Tx`; return domain models from `internal/model/`.
- Error handling:
  ```go
  if err != nil {
      return nil, fmt.Errorf("failed to fetch session: %w", err)
  }
  ```
- Dependency Injection: Constructors (`New...`) take explicit interfaces or dependencies; no global mutable state.

## Frontend Patterns

- Composition API with `<script setup>` only; state modules in `state/`, composables in `composables/`.
- Views mapped via `componentsMap` in `App.vue`.
- Always consume generated OpenAPI types from `@/types/types.generated`.
- Keep reactive sources narrow; derive with `computed()`.

## Scripts & Tasks

- Canonical bash style in `scripts/` with `set -euo pipefail`.
- Use `./scripts/linting.bash --go` to verify Go changes before committing.

## References

- Go Coding Standards: `.agent/skills/backend-go-coding/SKILL.md`
- Linting: `.agent/skills/backend-linting/SKILL.md`
- Tests: `.agent/skills/backend-tests/SKILL.md`
- Frontend rules: `.agent/rules/google_standards_typescript.md`
