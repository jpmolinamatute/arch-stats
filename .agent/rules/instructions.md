---
trigger: model_decision
description: Project Overview
---

# Arch-Stats – AI Coding Agent Guide

Audience: AI coding agents
Repo: <https://github.com/jpmolinamatute/arch-stats>  
Last updated: 2026-02-13

## Big Picture

- Two surfaces: Backend (FastAPI, Python 3.14, PostgreSQL 17) and Frontend (Vue 3 + Vite SPA). FE build
  outputs go into `backend/src/frontend/` and are served by FastAPI.
- Data flow (shot lifecycle): sensor/bot inserts rows → Postgres NOTIFY → server/WebSocket (planned)
  → frontend renders updates. Keep flow unidirectional and components loosely coupled.

## Non‑Negotiables

- DB access: `asyncpg` only (no ORM, no sync DB).
- Pydantic v2 models with `model_config = ConfigDict(extra="forbid")`.
- Strict typing across Python; avoid `# type: ignore` unless justified.
- Formatting: Python (Ruff, Ty), JS/TS (ESLint, Prettier). Keep diffs minimal.
- FE/BE separation: FE imports generated OpenAPI types; do not import backend internals in FE.
- FE build output path must remain `backend/src/frontend/`.

## Repo Map (targets to read first)

- Backend app: `backend/src/app.py`, routers under `backend/src/routers/`, models in `backend/src/models/`,
  schema in `backend/src/schema/`, core utilities in `backend/src/core/`.
- DB migrations: `backend/migrations/*.sql` orchestrate tables like archers, sessions, arrows, shots.
- Frontend app: `frontend/src/App.vue`, `frontend/src/composables/`, `frontend/src/api/`, and generated
  types in `frontend/src/types/`.
- Tasks: VS Code tasks for Docker, Uvicorn, Vite; see workspace tasks and `scripts/`.

## Core Workflows

- Start services (Docker + API + FE):
    - VS Code tasks: Start Docker Compose, Start Uvicorn Server, Start Vite Server.
    - Or run manually:

    ```bash
    docker compose -f docker/docker-compose.yaml up -d
    ./scripts/start_uvicorn.bash
    (cd frontend && npm install && npm run dev)
    ```

- Backend env setup (per backend/README):

  ```bash
  cd backend
  uv sync --dev --python $(cat ./.python-version)
  source ./.venv/bin/activate
  ```

- Generate FE API types when backend OpenAPI changes:

  ```bash
  cd frontend
  npm run generate:types
  ```

- Lint + tests (multi-language):

  ```bash
  ./scripts/linting.bash
  ```

## Backend Patterns

- DB helpers: small, async functions using `asyncpg`; consider prepared statements; return Pydantic
  models or dicts.
- Routers: accept/return Pydantic; snake_case JSON; validation at edges.
- Avoid global mutable state; obtain DB pool via startup DI in `core/db_pool.py`.
- Future real-time: centralize WebSocket connection management; never block the event loop.
- Imports at top; inline imports only for `if TYPE_CHECKING:`.

Example (conceptual):

```python
async def fetch_session(pool: Pool, session_id: UUID) -> SessionsRead | None:
    row = await pool.fetchrow("SELECT * FROM sessions WHERE id = $1", session_id)
    return SessionsRead.model_validate(row) if row else None
```

## Frontend Patterns

- Composition API with `<script setup>` only; state modules in `state/`, composables in `composables/`
  (see `useSession.ts`, `useTarget.ts`).
- Views mapped via `componentsMap` in `App.vue`; extend `uiManagerStore` discriminated union when
  adding views.
- Always consume generated OpenAPI types (e.g., `components["schemas"]["SessionsRead"]`).
- Keep reactive sources narrow; derive with `computed()`; avoid spreading reactive objects before network
  serialization.

Example composable:

```ts
import type { components } from "@/types/types.generated";
type Session = components["schemas"]["SessionsRead"];
export async function getOpenSession(): Promise<Session | null> {
  const res = await fetch("/api/v0/session/open");
  const body = await res.json();
  return body?.data ?? null;
}
```

## Scripts & Tasks

- Canonical bash style in `scripts/` with `set -euo pipefail`; see `scripts/linting.bash`.

## Testing

- Endpoint tests live in `backend/tests/endpoints/` (pattern: `test_<resource>_endpoints.py`),
  fixtures in `backend/tests/conftest.py`.
- Model logic tests in `backend/tests/models/`.

## References

- Backend rules: `.agent/rules/google_standards_python.md`
- Frontend rules: `.agent/rules/google_standards_typescript.md`
- Top-level docs: `backend/README.md`, `frontend/README.md`, `scripts/README.md`

When unsure: prefer smallest changes, mirror existing patterns, and keep strict typing/validation.
Open a PR with clarification notes instead of speculative refactors.
