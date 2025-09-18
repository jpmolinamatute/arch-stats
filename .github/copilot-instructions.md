# Arch-Stats - AI Assistant Project Instructions

**Audience:** AI coding agents (Copilot, etc.)  
**Repo:** https://github.com/jpmolinamatute/arch-stats  
**Last updated:** 2025-08-18

## 1. Purpose & Architecture (Big Picture)

Arch-Stats tracks archery performance. Two main runtime surfaces:

1. **Backend (`backend/`)** FastAPI (Python 3.13) + PostgreSQL 17. Sensor simulators (arrow/bow/target readers) write events to DB; server exposes REST `/api/v0/...` + WebSockets for real-time updates. Modules communicate indirectly through the database and Postgres `LISTEN/NOTIFY` (decoupling - do NOT add direct cross-module imports for runtime coupling).
2. **Frontend (`frontend/`)** Vue 3 + Vite SPA consuming REST + (future) WebSocket stream. Build artifacts emitted into `backend/src/server/frontend/` and then served by FastAPI.

Data flow example (shot lifecycle): sensor script -> insert row(s) -> Postgres NOTIFY -> server/websocket (future) -> frontend updates table/visuals. Maintain this unidirectional flow; avoid tight coupling.

## 2. Golden Constraints (Never Violate)

- DB access:
- Models: Pydantic v2 with `model_config = ConfigDict(extra="forbid")` (no legacy `Config` class).
- Strict typing: every Python function annotated; pass mypy (strict) without `# type: ignore` unless justified.
- Formatting: Python (Black + isort + Pylint), JS/TS (ESLint + Prettier). Keep diffs minimal.
- Frontend build output path: `backend/src/server/frontend/` (never change path assumptions in server code).
- Keep FE/BE separation: no importing backend internals into frontend types; generate API types (`npm run generate:types`).

## 3. Repository Layout (Key Directories)

| Path                    | Purpose                                                                       |
| ----------------------- | ----------------------------------------------------------------------------- |
| `backend/src/server/`   | FastAPI app factory, routers, db pool, static serving.                        |
| `backend/src/*_reader/` | Sensor simulators (arrow/bow/target). Each isolated; communicate only via DB. |
| `backend/src/shared/`   | Shared utilities (logging, factories). Keep generic & dependency-light.       |
| `backend/tests/`        | Pytest suites (models + endpoints). Use these as examples for new tests.      |
| `frontend/src/`         | Vue SPA code (components, composables, state).                                |
| `scripts/`              | Automation (linting, start scripts, installation). Bash must be POSIX-safe.   |
| `docker/`               | Postgres compose + config.                                                    |

## 4. Core Workflows (Do These Exactly)

Backend setup:

```bash
cd backend
uv sync --dev --python $(cat ./.python-version)
source ./.venv/bin/activate
docker compose -f docker/docker-compose.yaml up -d
```

Run API locally:

```bash
./scripts/start_uvicorn.bash   # or VS Code task "Start Uvicorn Dev Server"
```

Run sensor bot (simulated shots):

```bash
./scripts/start_archy_bot.bash # adds shot data for manual testing
```

Frontend dev:

```bash
cd frontend
npm install
npm run dev                    # or VS Code task "Start Vite Frontend"
```

Generate API types (run whenever backend OpenAPI changes):

```bash
npm run generate:types
```

Full multi-language lint/test (pre-commit style): staged detection handled by `scripts/linting.bash` - invoke manually to check everything:

```bash
./scripts/linting.bash
```

## 5. Patterns & Conventions

Backend:

- Small async DB helper functions; wrap queries with prepared statements where beneficial; return simple dict/record or Pydantic model.
- Routers: accept/return Pydantic models; enforce validation at boundary; snake_case JSON.
- WebSocket (future/real-time): centralize connection management; handle disconnects gracefully; never block event loop.
- Avoid global mutable state; acquire DB pool via startup dependency injection.

Frontend:

- Composition API `<script setup>` only. Keep state modules in `state/` (Pinia-like simple objects) or composables in `composables/` (`useSession.ts`, `useTarget.ts` as patterns).
- Map view names to components (see `App.vue` `componentsMap`). When adding a new view, extend the discriminated union in `uiManagerStore` and update mapping.
- Always import generated API types instead of redefining (e.g., use `SessionsRead` from `types.generated.ts`).
- Keep reactive sources narrow; compute derivatives with `computed()`. Avoid spreading reactive objects before network serialization.

Scripts:

- Bash scripts must use `set -euo pipefail`; prefer functions + explicit exits (see `scripts/linting.bash` for canonical style).

Testing:

- For new endpoint: add pytest in `backend/tests/endpoints/` mirroring existing naming (`test_<resource>_endpoints.py`). Use existing fixtures (see `conftest.py`).
- For model logic: add to `backend/tests/models/` following current style.

## 6. Adding Features - Checklist

1. Define/adjust Pydantic schema (backend) - keep strict, forbid extra.
2. Implement async DB accessor (single responsibility).
3. Expose via FastAPI router; update OpenAPI-driven types by regenerating frontend types.
4. Add/extend frontend composable for API call; return typed `Promise<...>`.
5. Update state and UI components; ensure minimal coupling (use mapping patterns like in `App.vue`).
6. Write/extend tests (unit + endpoint). Run `./scripts/linting.bash`.
7. Keep diffs focused; no drive-by refactors without justification.

## 7. Things to Avoid

- Introducing new heavy dependencies (both Python & JS) without clear performance/maintainability gain.
- Synchronous I/O in backend request path.
- Committing `frontend/src/types/types.generated.ts` (it is ignored).
- Cross-imports between sensor modules; they must stay DB-decoupled.
- Synchronous DB calls or ad-hoc threads.
- Mixing build outputs into source directories (only emit to `backend/src/server/frontend/`).

## 8. Quick Examples

Minimal async DB pattern (conceptual):

```python
async def fetch_session(pool: Pool, session_id: UUID) -> SessionsRead | None:
    row = await pool.fetchrow("SELECT * FROM sessions WHERE id = $1", session_id)
    return SessionsRead.model_validate(row) if row else None
```

Frontend composable pattern:

```ts
import type { components } from "@/types/types.generated";
type Session = components["schemas"]["SessionsRead"];
export async function getOpenSession(): Promise<Session | null> {
  const res = await fetch("/api/v0/session/open");
  const body = await res.json();
  return body.data ?? null;
}
```

## 9. Reference Documents

- Frontend rules: [./instructions/frontend.instructions.md](./instructions/frontend.instructions.md)
- Backend rules: [./instructions/backend.instructions.md](./instructions/backend.instructions.md)
- Google Python style (linked inside backend rules).

## 10. When Unsure

Prefer smallest change; mirror existing patterns; ask for clarification via PR description instead of speculating. Maintain strict typing & validation guarantees.

---

Provide feedback if any area above is unclear or missing (e.g., test fixtures, WebSocket specifics, or deployment details) and this guide will be updated.
