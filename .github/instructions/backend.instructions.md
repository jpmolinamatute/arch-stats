# Copilot Instructions \* Backend (FastAPI + Python 3.13)

**Audience:** Python developers working in [backend/](../../backend/)

---

## applyTo: "\*_/_.py"

## Tech stack

- **Language:** Python 3.13 (strict typing)
- **Web framework:** FastAPI
- **Data models/validation:** Pydantic **v2.x** (`ConfigDict`, `Field`, `extra="forbid"`)
- **DB:** PostgreSQL 15+ via **`asyncpg` only** (no ORMs)
- **Async:** idiomatic `async/await` with connection pools/transactions
- **Real-time:** WebSockets
- **Testing:** pytest (+ pytest-asyncio, httpx.AsyncClient)
- **Quality:** Black (format), isort (imports), mypy (strict), Pylint (lint)

## Coding standards

- [Follow Google's Python Guide Line](./google_standards.md)
- Every function and module must be **fully typed**; `mypy` should pass with no ignores.
- prefer Pydantic models for I/O boundaries over dataclasses.
- Pydantic v2 config via `model_config = ConfigDict(...)`; no inner `Config` class.
- Validate all inputs at router boundaries; return precise HTTP status codes.
- Keep DB access **isolated** in small async functions using `asyncpg` with prepared statements when helpful.
- Avoid global state; pass pools/clients via dependency injection or startup wiring.

## Project expectations

- **No synchronous DB clients**, and do not introduce ORMs.
- **WebSocket handlers** must be robust: handle disconnects, timeouts, and backpressure sensibly.
- Follow REST naming patterns already used (snake_case JSON keys).
- Log useful context (request id, principal, resource id) without spamming.

## Testing discipline

- Unit-test business logic without hitting the DB where feasible.
- Integration tests use Dockerized Postgres (see `docker/docker-compose.yaml`).
- For HTTP, use `httpx.AsyncClient` against the FastAPI app factory.
- Seed/factories for test data accepted; ensure determinism.

## Feature Integration Checklist

When adding or updating backend features, ensure changes are consistent across routers, schemas, models, and tests.

### Routers (`backend/src/server/routers/`)

- Keep **endpoint code minimal**: no business logic in routes, only orchestration.
- Use dependency-injected services and async DB helpers.
- Follow REST naming and status code patterns used elsewhere.
- Document inputs/outputs with Pydantic models.

### Schemas (`backend/src/server/schema/`)

- All request/response bodies must use **Pydantic v2 models**.
- `extra="forbid"` and strict typing for all fields.
- Use `Field(...)` for defaults, constraints, and descriptions.
- Version schemas when changing existing contracts.

### Models (`backend/src/server/models/`)

- Add or modify DB table fields with clear migrations.
- Use only `asyncpg` queries; parameterized, never interpolated SQL.
- Update helper functions or query builders when schema changes.
- Keep models dumb: only describe fields and queries, not business rules.

### Tests (`backend/tests/{endpoints,models}/`)

- For each new/updated route: add tests under `tests/endpoints/`.
- For each new/updated model: add tests under `tests/models/`.
- Cover both success and failure paths (e.g. invalid payloads, DB errors).
- Use factories or seed data to keep tests deterministic.
- Use `httpx.AsyncClient` for HTTP tests, transaction rollbacks for DB tests.

**Rule of thumb:** any new endpoint must have:

1. Minimal router code
2. Schema(s) for validation
3. Matching model/queries
4. Unit and integration tests

Failing to check one of these boxes = incomplete feature.

## References

- See [backend/README.md](../../backend/README.md) for setup, structure, and VS Code tasks.
