# Backend instructions

**Audience:** Python developers working in [backend/](../../backend/)

---
applyTo: "**/*.py"
---

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

### Validation & Single Source of Truth

- Pydantic schemas define external contract & input validation.
- Database constraints (NOT NULL, UNIQUE, FK, CHECK) must mirror critical invariants; validation is **duplicated only for invariants that protect data integrity** (e.g. FK, uniqueness) to prevent drift.
- Do not add business-only rules (like max distance heuristics) into DB constraints—keep those at the schema/service layer.

## Project expectations

- **No synchronous DB clients**, and do not introduce ORMs.
- **WebSocket handlers** must be robust: handle disconnects, timeouts, and backpressure sensibly.
- Follow REST naming patterns already used (snake_case JSON keys).
- Log useful context (request id, principal, resource id) without spamming.

### Logging Format

Emit structured JSON (preferred) or key=value single line. Minimum fields:

`ts level module msg request_id=<uuid> resource=<entity> latency_ms=<int>`

If using JSON: `{ "ts": "2025-08-20T12:34:56.123Z", "level": "INFO", "module": "sessions.router", "msg": "session opened", "request_id": "...", "resource": {"type": "session", "id": "..."}, "latency_ms": 12 }`

Never log PII; truncate arrays/large payloads.

### Connection Pool & Concurrency

Environment variables (documented here, default fallback in code):

| Var | Purpose | Suggested Dev Default |
|-----|---------|-----------------------|
| `PG_POOL_MIN` | minimum connections | 1 |
| `PG_POOL_MAX` | maximum connections | 10 |
| `API_MAX_CONCURRENT_REQUESTS` | (future) semaphore limit | 100 |

Guideline: keep `POOL_MAX` <= (CPU cores * 2) for typical IO-bound FastAPI usage. Reassess with load tests.

### Prepared Statements Guidance

Use prepared statements when a query is executed frequently (rule of thumb: > 50 times per process lifetime OR inside high-throughput endpoints). Name statements deterministically: `<area>_<table>_<action>` (e.g., `sessions_select_open`). Avoid preparing one-off migration queries.

### WebSocket (Future Contract Stub)

When implementing the real-time channel, follow this envelope shape (JSON text frames):

```json
{
	"type": "shot.created",      // event name (snake_case + domain)
	"ts": "2025-08-20T12:34:56.789Z",
	"request_id": "<uuid or null>",
	"data": { /* event-specific payload */ }
}
```

Rules:
1. Never block send loop; queue with backpressure (drop oldest after N=1000 if client slow, log WARN).
2. Heartbeat: server -> client `{"type":"heartbeat","ts":"..."}` every 30s.
3. No client->server mutation messages until an explicit spec section is added.

### Error Model

Success responses: HTTP 2xx with `HTTPResponse` wrapper (`code`, `data`, `errors=[]`).

Client mistakes (validation, not found, conflict): return appropriate 4xx AND still use wrapper with `data=null`, `errors=[..]` so frontend parsing is uniform. Do *not* return 200 + populated `errors` for true failures.

Server unexpected failures: 500 with wrapper, log stack (not returned). FastAPI exception handlers centralize formatting.

### Migrations Process

Current approach: raw SQL migration files under `backend/src/server/models/migrations/` (create directory if absent) named with ordered prefix: `YYYYMMDDHHMM__short_description.sql`.

Checklist to add a migration:
1. Create SQL file (idempotent where feasible; guard with `IF NOT EXISTS`).
2. Include `-- migrate:up` section (and optional `-- migrate:down`).
3. Provide accompanying minimal test asserting new columns / tables exist.
4. Run locally against fresh DB (`docker compose down -v && up`) to verify.
5. Document any destructive change in PR description.

Future tooling placeholder: if adopting a migration runner, integrate here; until then, a lightweight script will apply pending files ordered lexicographically.

### OpenAPI -> Frontend Types Workflow

Any change to routers or schemas that alters OpenAPI (new field, endpoint, description affecting generated types) requires:
1. Regenerate: `npm run generate:types` (frontend directory) while backend running.
2. Verify FE build; DO NOT commit `types.generated.ts` (ignored).
3. Mention in PR checklist: "API types regenerated locally".

### Sensor Simulators (arrow/bow/target)

Target cadence assumptions (guidance for realistic data):
- Bow reader: emits `arrow_engage` and `arrow_release` with 2–8s average gap during active session.
- Target reader: landing detection within configurable window (default 3s) after release; if absent, mark miss.
- Arrow reader: ID association event occurs prior to `arrow_engage` (<=1s).

Tests may mock these timings; do not bake fixed sleeps into production logic—use timestamps from DB rows.

## Testing discipline

- Unit-test business logic without hitting the DB where feasible.
- Integration tests use Dockerized Postgres (see `docker/docker-compose.yaml`).
- For HTTP, use `httpx.AsyncClient` against the FastAPI app factory.
- Seed/factories for test data accepted; ensure determinism.

## Server feature Integration Checklist

When adding or updating a server features, ensure changes are consistent across routers, schemas, models, and tests.

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

- See [backend/README.md](../../backend/README.md) for setup, structure, and VS Code tasks.
 - For style/consistency: see top-level architecture guide `.github/copilot-instructions.md`.
