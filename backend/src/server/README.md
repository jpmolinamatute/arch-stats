# Arch-Stats Backend Server

## Table of Contents

- [Arch-Stats Backend Server](#arch-stats-backend-server)
  - [Table of Contents](#table-of-contents)
  - [TL;DR Quick Start](#tldr-quick-start)
  - [Audience \& Scope](#audience--scope)
  - [Overview](#overview)
  - [Non-Negotiable Constraints](#non-negotiable-constraints)
  - [Quick Start](#quick-start)
  - [Environment \& Configuration](#environment--configuration)
  - [Request Lifecycle](#request-lifecycle)
  - [Architecture Principles](#architecture-principles)
  - [File Structure](#file-structure)
    - [Entry Point](#entry-point)
    - [Database Layer](#database-layer)
    - [Data Validation (Schemas)](#data-validation-schemas)
    - [Routers (API Endpoints)](#routers-api-endpoints)
    - [WebSocket Endpoint](#websocket-endpoint)
  - [Error Handling \& Responses](#error-handling--responses)
  - [Mini End-to-End Example](#mini-end-to-end-example)
  - [Testing Guidance](#testing-guidance)
  - [Performance \& Safety Notes](#performance--safety-notes)
  - [Extending: Checklist](#extending-checklist)
  - [Dev Command Cheat Sheet](#dev-command-cheat-sheet)
  - [Further References](#further-references)

## TL;DR Quick Start

```bash
./scripts/start_uvicorn.bash          # or VS Code task
./scripts/start_archy_bot.bash        # optional simulated shots
```

## Audience & Scope

This document is for Python developers working specifically inside `backend/src/server/`. For broader repository vision or sensor simulators, see the top-level [README](../../../README.md) and backend root [README](../../README.md).

## Overview

The server is an async **FastAPI** application exposing REST (and a WebSocket) backed by **PostgreSQL** via **`asyncpg`** (no ORM). All request/response I/O is validated with **Pydantic v2** strict models.

## Non-Negotiable Constraints

- DB access: `asyncpg` only (no psycopg2, no ORM).
- Strict Pydantic v2 models (`model_config = ConfigDict(extra="forbid")`).
- Fully typed Python; `mypy` (strict) must pass.
- Parameterized SQL only (no f-string interpolation of user input).
- No runtime cross-import coupling with sensor modules—communication only through the database and (future) LISTEN/NOTIFY.
- Keep router functions thin: orchestration only; no business logic.

## Quick Start

```bash
# From repo root
cd backend
uv sync --dev --python $(cat .python-version)
source .venv/bin/activate
docker compose -f docker/docker-compose.yaml up -d
./scripts/start_uvicorn.bash            # start API
pytest -q                                # run tests
```

Visit: <http://localhost:8000/api/swagger>

## Environment & Configuration

Runtime settings are loaded in [shared/settings.py](../shared/settings.py) (Pydantic settings model). Typical environment variables:

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`
- `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `ARCH_STATS_DEV_MODE` (optional)
Add new settings fields only if required at runtime; prefer deriving defaults in code.

## Request Lifecycle

```mermaid
flowchart LR
    CLIENT --> REQ[HTTP Request]
    REQ --> VAL_IN[Request Pydantic Model]
    VAL_IN --> DB[(PostgreSQL)]
    DB --> VAL_OUT[Response Pydantic Model]
    VAL_OUT --> WRAP[Response Wrapper]
    WRAP --> CLIENT
```

## Architecture Principles

- Explicit async I/O; no hidden abstractions.
- Clear boundary: Routers call a focused DB helper (one purpose).
- Fail fast: reject invalid input before any DB work.
- Deterministic tests (no timing races / sleeps).
- Small, composable functions (prefer clarity over cleverness).

## File Structure

```text
server/
├── app.py            # App factory (startup, shutdown, router include)
├── (settings moved to shared/)  
├── (db pool lives in shared/)  
├── models/           # CRUD/query helpers per resource + DBBase
├── routers/          # FastAPI APIRouter groupings + websocket
├── schema/           # Pydantic request/response/filter models
├── frontend/         # Built frontend assets (served as static)
└── README.md
```

### Entry Point

- [app.py](./app.py): Creates FastAPI instance, registers routers, mounts static frontend, wires startup/shutdown events (e.g., pool creation/close).
- Startup acquires an asyncpg pool (see [db_pool.py](./db_pool.py)); dependency functions fetch that pool per request (no globals leaked).

### Database Layer

Located in [models/](./models/):

- `DBBase` ([models/base_db.py](./models/base_db.py)) supplies shared helpers (fetch, fetchrow, execute) and translates low-level failures into domain exceptions.
- Resource modules (e.g., `sessions_db.py`, `shots_db.py`) define parameterized SQL + thin, typed wrapper functions.
- Always return either primitives/record dicts or validated Pydantic models, never raw driver rows past the boundary.

### Data Validation (Schemas)

Located in [schema/](./schema/):

- Naming pattern: `XyzCreate`, `XyzRead`, `XyzUpdate`, `XyzFilters`.
- All models: `extra="forbid"`.
- Use explicit optional fields (e.g., `value: int | None = None`) for PATCH semantics.
- Filter models drive query parameter parsing in routers.

### Routers (API Endpoints)

Located in [routers/](./routers/):

- One router file per resource (sessions, shots, arrows, targets).
- Wrap return values with helper functions in [routers/utils.py](./routers/utils.py) to produce a consistent `{"code": int, "data": ..., "errors": []}` envelope which the frontend type generator consumes.

### WebSocket Endpoint

- [websocket.py](./routers/websocket.py): Handles (future) real-time shot updates via Postgres NOTIFY or buffered in-memory fan-out. Keep it non-blocking and resilient to disconnects.

## Error Handling & Responses

Mapped exceptions (see raising in DB layer):

| Exception | HTTP | Meaning |
|-----------|------|---------|
| `DBNotFound` | 404 | Resource does not exist |
| `DBException` | 422 | Invalid input / integrity / constraint |
| Unhandled | 500 | Internal error |

Responses use envelope helpers in [`routers/utils.py`](./routers/utils.py) for consistency.

## Mini End-to-End Example

Below is a minimal (illustrative) pattern for adding a new resource "widgets" (not in repo, example only):

```python
# schema/widgets.py
from pydantic import BaseModel, ConfigDict, Field
from typing import Optional
class WidgetCreate(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: str = Field(..., min_length=1)
    size: int

class WidgetRead(WidgetCreate):
    id: str

class WidgetUpdate(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: Optional[str] = None
    size: Optional[int] = None
```

```python
# models/widgets_db.py
from .base_db import DBBase
from typing import Sequence
class WidgetsDB(DBBase):
    async def insert(self, name: str, size: int) -> str:
        row = await self.pool.fetchrow(
            "INSERT INTO widgets (name,size) VALUES ($1,$2) RETURNING id",
            name, size,
        )
        return str(row["id"])
```

```python
# routers/widgets_router.py
from fastapi import APIRouter, Depends
from ..schema.widgets import WidgetCreate, WidgetRead
from ..models.widgets_db import WidgetsDB
from .utils import ok_response
from ..db_pool import get_pool

router = APIRouter(prefix="/api/v0/widget", tags=["widgets"])

def get_db(pool=Depends(get_pool)) -> WidgetsDB:
    return WidgetsDB(pool)

@router.post("", response_model=dict)
async def create_widget(payload: WidgetCreate, db: WidgetsDB = Depends(get_db)):
    wid = await db.insert(payload.name, payload.size)
    return ok_response(data=wid)
```

## Testing Guidance

- Endpoint tests: [../../tests/endpoints/](../../tests/endpoints/) using `httpx.AsyncClient`.
- Model / DB logic tests: [../../tests/models/](../../tests/models/).
- Use existing fixtures (`conftest.py`) for pool / app startup.
- Cover both success and failure (404, 422 paths).
- Avoid sleeps, prefer deterministic fixtures and explicit inserts.

## Performance & Safety Notes

- Use prepared/parameterized statements (asyncpg handles this when reusing identical SQL).
- Avoid N+1 loops; batch where possible (e.g., single `IN` queries).
- Keep transactions small; only open when atomic multi-step operations required.
- Never hold a connection across `await` boundaries doing unrelated work, fetch row(s), then release.

## Extending: Checklist

1. Define / adjust Pydantic schema(s) (`Create`, `Read`, `Update`, `Filters`).
2. Add DB helper methods (parameterized SQL only).
3. Add / extend router endpoints; wrap responses (`ok_response`, `error_response`).
4. Write tests (model + endpoint).
5. Run `./scripts/linting.bash`.
6. Start server, verify new OpenAPI; regenerate frontend types (`npm run generate:types`).
7. Keep diffs minimal; no drive-by refactors.

## Dev Command Cheat Sheet

| Task | Command |
|------|---------|
| Start DB (Docker) | `docker compose -f docker/docker-compose.yaml up -d` |
| Run API | `./scripts/start_uvicorn.bash` |
| Run simulated shots | `./scripts/start_archy_bot.bash` |
| All lint/type/test | `./scripts/linting.bash` |
| Tests only | `pytest -q` |
| Regenerate frontend API types | `(cd ../frontend && npm run generate:types)` |

## Further References

- Project-wide backend rules: [.github/instructions/backend.instructions.md](../../../.github/instructions/backend.instructions.md)
- Root backend overview: [../../README.md](../../README.md)
- Response envelope pattern: [routers/utils.py](./routers/utils.py)
- DB base class: [models/base_db.py](./models/base_db.py)
