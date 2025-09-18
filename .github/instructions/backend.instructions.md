---
applyTo: "**/*.py"
---

# Backend instructions

**Audience:** Python developers working in `backend/` directory.

## Tech stack

- **Language:** Python 3.14 (strict typing)
- **Web framework:** FastAPI
- **Data models/validation:** Pydantic **v2.x** (`ConfigDict`, `Field`, `extra="forbid"`)
- **DB:** PostgreSQL 17+ via **`asyncpg` only** (no ORMs)
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

### Logger

import and use the core logger from `backend/src/core/logger.py`:

```python
from core import LoggerFactory

logger = LoggerFactory().get_logger(__name__)
logger.info("message %s", {"key": value})
```

### Connection Pool & Concurrency

Environment variables (documented here, default fallback in code):

| Var                           | Purpose                  | Suggested Dev Default |
| ----------------------------- | ------------------------ | --------------------- |
| `POSTGRES_POOL_MIN_SIZE`      | minimum connections      | 1                     |
| `POSTGRES_POOL_MAX_SIZE`      | maximum connections      | 10                    |
| `API_MAX_CONCURRENT_REQUESTS` | (future) semaphore limit | 100                   |

Guideline: keep `POOL_MAX` <= (CPU cores \* 2) for typical IO-bound FastAPI usage. Reassess with load tests.

### WebSocket (Future Contract Stub)

When implementing the real-time channel, follow this envelope shape (JSON text frames):

```json
{
  "type": "shot.created", // event name (snake_case + domain)
  "ts": "2025-08-20T12:34:56.789Z",
  "request_id": "<uuid or null>",
  "data": {
    /* event-specific payload */
  }
}
```

Rules:

1. Never block send loop; queue with backpressure (drop oldest after N=1000 if client slow, log WARN).
2. Heartbeat: server -> client `{"type":"heartbeat","ts":"..."}` every 30s.
3. No client->server mutation messages until an explicit spec section is added.

### OpenAPI -> Frontend Types Workflow

Any change to routers or schemas that alters OpenAPI (new field, endpoint, description affecting generated types) requires:

1. Regenerate: `npm run generate:types` (frontend directory) while backend running.
2. Verify FE build; DO NOT commit `types.generated.ts` (ignored).
3. Mention in PR checklist: "API types regenerated locally".

## Testing discipline

- Unit-test business logic without hitting the DB where feasible.
- Integration tests use Dockerized Postgres (see `docker/docker-compose.yaml`).
- For HTTP, use `httpx.AsyncClient` against the FastAPI app factory.
- Seed/factories for test data accepted; ensure determinism.

## Server feature Integration Checklist

When adding or updating a server features, ensure changes are consistent across routers, schemas, models, and tests.

### Routers (`backend/src/routers/`)

- Keep **endpoint code minimal**: no business logic in routes, only orchestration.
- Use dependency-injected services and async DB helpers.
- Validate inputs/outputs with Pydantic models.
- Follow REST naming conventions used elsewhere.
- Routers ALWAYS must return HTTPStatus and a appropriate status code patterns.

```python
from fastapi import APIRouter, Depends, Response, status, HTTPException
from uuid import UUID
from models import ArcherModel
from schemas import ArcherRead


router = APIRouter()


@router.get("/{archer_id}", response_model=ArcherRead, status_code=status.HTTP_200_OK)
async def get_archer(
    archer_id: UUID,
    archer_model: ArcherModel = Depends(),
) -> ArcherRead:
    try:
        return await archer_model.get_one(archer_id)
    except DBNotFound as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Archer not found") from e
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail="Archer not found") from e
```

### Schemas (`backend/src/schema/`)

- All request/response bodies must use **Pydantic v2 models**.
- `extra="forbid"` and strict typing for all fields.
- Use `Field(...)` for defaults, constraints, and descriptions.
- Version schemas when changing existing contracts.

### Models (`backend/src/models/`)

- Add or modify DB table fields with clear migrations.
- Keep models dumb: only describe fields and queries, not business rules.

### Tests (`backend/tests/{endpoints,models}/`)

- For each new/updated route: add tests under `tests/endpoints/`.
- For each new/updated model: add tests under `tests/models/`.
- Cover both success and failure paths (e.g. invalid payloads, DB errors).
- Use factories or seed data to keep tests deterministic.
- Use `httpx.AsyncClient` for HTTP tests, transaction rollbacks for DB tests.

### Example Endpoint Test Pattern (Reference)

```python
async def test_create_session(client: AsyncClient) -> None:
	payload = {"is_opened": True, "start_time": datetime.utcnow().isoformat()+"Z", "location": "range A", "is_indoor": False, "distance": 18}
	resp = await client.post("/api/v0/session", json=payload)
	assert resp.status_code == 200
	body = resp.json()
	assert body["code"] == 200
	assert body["errors"] == []
```

Use similar minimal assertions + any domain specific follow-ups.

**Rule of thumb:** any new endpoint must have:

1. Minimal router code
2. Schema(s) for validation
3. Matching model/queries
4. Unit and integration tests

Failing to check one of these boxes = incomplete feature.

## References

- For style/consistency: see top-level architecture guide `.github/copilot-instructions.md`.
