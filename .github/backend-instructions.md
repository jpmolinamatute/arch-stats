# Copilot Instructions * Backend (FastAPI + Python 3.13)

**Audience:** Python developers working in [backend/](../backend/)

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

- [Follow Google's Python Guide Line](./google_standars.md)
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

## Commands (examples)

- Format: `black --config ./pyproject.toml ./src && isort --settings-file ./pyproject.toml ./src`
- Type-check: `mypy --config-file ./pyproject.toml ./src`
- Lint: `pylint --rcfile ./pyproject.toml ./src`
- Tests: `pytest -v`

## References

- See [backend/README.md](../backend/README.md) for setup, structure, and VS Code tasks.
