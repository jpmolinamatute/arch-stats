# Arch Stats Backend

## Description

The **Arch Stats backend** is a high-performance REST API that powers the Arch Stats platform.
It manages archery session data, user authentication, and statistical analysis. Built with modern
Python async technologies, it ensures low-latency responses and scalable data handling.

## Development Guidelines

**Audience:** Developers working in the `backend/` directory.

## Architecture

The backend is built on a robust, asynchronous stack:

- **Framework**: [FastAPI](https://fastapi.tiangolo.com/) (High performance, easy-to-use)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Database Access**: [asyncpg](https://github.com/MagicStack/asyncpg) (No ORM, raw SQL for performance)
- **Validation**: [Pydantic v2](https://docs.pydantic.dev/)
- **Server**: [Uvicorn](https://www.uvicorn.org/)
- **Package Manager**: [uv](https://github.com/astral-sh/uv)

## Golden Constraints (must-follow)

These rules apply across the backend. Violations will be flagged in CI and during review:

- **DB access**: Use `asyncpg` only. No ORM. No synchronous DB calls.
- **Pydantic v2**: Use `model_config = ConfigDict(extra="forbid")`. Do not use inner `Config` classes.
- **Strict typing**: Annotate every function. `mypy --strict` must pass
    without `# type: ignore` unless justified.
- **Formatting**: Python via Black, isort, Pylint. Keep diffs minimal and focused.
- **Frontend build output path**: The frontend build must emit to `backend/src/frontend/`.
- **FE/BE separation**: Routers return Pydantic models; do not couple frontend types directly.

## Prerequisites

Ensure your environment meets these requirements:

- **Python**: v3.14
- **Docker**: Required for the database and migrations.
- **uv**: Required for package management.

## Setup

Follow these steps to get the backend running:

1. **Install Dependencies**:

    This command creates the virtual environment at `backend/.venv` and installs all dependencies.

    ```bash
    cd ./backend
    uv sync
    cd ..
    ```

2. **Start the Server**:

    You can start the server using the automated script or manually.

    **Option A: Automated Script (Recommended)**

    This script activates the virtual environment, starts Docker (database), and runs Uvicorn.

    ```bash
    ./scripts/start_uvicorn.bash
    ```

    **Option B: Manual Setup**

    1. Activate the virtual environment:

        ```bash
        source ./backend/.venv/bin/activate
        ```

    2. Start the database:

        ```bash
        docker compose -f ./docker/docker-compose.yaml up --build --detach
        ```

    3. Start Uvicorn:

        ```bash
        cd backend/src
        exec uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors \
        --log-level debug --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 app:run
        ```

    The API will be available at `http://localhost:8000`.

    > [!NOTE]
    > In both approaches, remember to stop the database when you are done:
    >
    > ```bash
    > docker compose -f ./docker/docker-compose.yaml down
    > ```

## Quickstart (VS Code tasks)

If you are using this repository in VS Code, the workspace defines helpful tasks:

- Start database: Run task `Start Docker Compose`.
- Stop database: Run task `Stop Docker Compose` (or `Stop & Remove Volumes`).
- Start backend API: Run task `Start Uvicorn Server`.
- Start frontend (in separate terminal): Run task `Start Vite Server`.

These match the manual commands shown above.

## Scripts

Common commands for development:

| Command | Description |
| :--- | :--- |
| `./scripts/start_uvicorn.bash` | Start the FastAPI development server with hot reload. |
| `./scripts/linting.bash --backend` | Run all linters (black, isort, ty, pylint). |
| `pytest` | Run the test suite (requires active venv). |
| `docker compose ... up` | Start PostgreSQL and run migrations. |

Additional useful scripts:

- `./tools/generate_openapi.py`: Regenerate `openapi.json` from the running backend.
- `./scripts/create_pr.bash`: Open a pre-filled PR on GitHub.

## Git Hooks & Safety Net

To prevent committing broken code, "activate" the git pre-commit hook by creating a symlink to the
linting script:

```bash
ln -sfr ./scripts/linting.bash ./.git/hooks/pre-commit
```

This ensures that all linters and tests pass before a commit is allowed.

### Pull Requests

After pushing your changes, use the helper script to create a Pull Request:

```bash
./scripts/create_pr.bash
```

> [!IMPORTANT]
> Workflows are triggered selectively based on the files changed (e.g., frontend changes do not
> trigger backend tests). All **triggered** workflows must pass for a PR to be mergeable.

## Database & Migrations

- Migrations live in `backend/migrations/` and are applied by the `docker-compose` setup in `docker/docker-compose.yaml`.
- To bootstrap the database locally:

  ```bash
  docker compose -f ./docker/docker-compose.yaml up --build --detach
  ```

- To reset the environment (including volumes):

  ```bash
  docker compose -f ./docker/docker-compose.yaml down -v
  ```

## Running Tests

Run tests from the `backend/` environment:

```bash
cd ./backend
source ./.venv/bin/activate
pytest -q
```

Use the provided fixtures in `backend/tests/conftest.py`.
Endpoint tests live under `backend/tests/endpoints/` and model tests under
`backend/tests/models/`.

## Structure

The codebase is organized for separation of concerns:

- **`backend/src/routers/`**: Route definitions and HTTP status handling. **No business logic here.**
- **`backend/src/models/`**: Business logic and database interactions.
- **`backend/src/core/`**: Core business logic and utilities.
- **`backend/src/schema/`**: Pydantic models for I/O validation.
- **`backend/src/app.py`**: Application entry point.
- **`backend/tests/`**: Unit tests using pytest.

### Backend contribution pattern

When adding a new feature or endpoint:

1. Define/adjust strict Pydantic schema in `backend/src/schema/` using `ConfigDict(extra="forbid")`.
2. Implement a small async DB accessor in `backend/src/models/` using `asyncpg`.
3. Expose via a router in `backend/src/routers/` that accepts and returns Pydantic models.
4. Regenerate OpenAPI and frontend types (see "Frontend integration" below).
5. Add tests in `backend/tests/` (models and endpoints). Run `./scripts/linting.bash --backend`.

### Style Guidelines

> [!IMPORTANT]
> Follow the [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html).

### Coding Standards

1. **Typing**: Every function and class must be fully typed. `mypy` must pass without ignores.
2. **Pydantic v2**: Use `model_config = ConfigDict(...)`. Do not use inner `Config` classes.
3. **Database**: Use `asyncpg` with prepared statements. Keep DB access in `models/`.
4. **No ORM**: We use raw SQL for control and performance.
5. **Dependency Injection**: Pass DB pools and config via FastAPI dependencies.

### Linting & formatting

Run the full linting suite for the backend:

```bash
./scripts/linting.bash --backend
```

This runs **isort**, **black**, **mypy**, and **pylint**. Fix issues before opening a PR.

### Environment Setup

- **Virtual Env**: Always work inside `backend/.venv`.
- **Package Management**: Use **`uv`** for everything.
    - `uv add <package>`
    - `uv remove <package>`
    - **NEVER** use `pip install`.

### Frontend integration

In production, the backend (Uvicorn) serves both the API and the frontend static files.

1. Build the frontend in `frontend/`:

   ```bash
   cd ./frontend
   npm install
   npm run build
   ```

   This outputs files to `backend/src/frontend/`.

2. The backend serves these static files at `/` alongside the API.

If you modify backend routes or schemas, regenerate API types in the frontend:

```bash
cd ./frontend
npm run generate:types
```

### Example Endpoint Pattern

```python
@router.get("/{id}", response_model=ArcherRead, status_code=status.HTTP_200_OK)
async def get_archer(
    id: UUID,
    current_user: UUID = Depends(require_auth),
    model: ArcherModel = Depends(),
) -> ArcherRead:
    if current_user != id:
        raise HTTPException(status_code=403, detail="Forbidden")
    return await model.get_one(id)
```

### Code Quality

All code must pass the strict linting suite:

```bash
./scripts/linting.bash --backend
```

This runs **isort**, **black**, **mypy**, and **pylint**.

### Production Serving

In production, the backend (Uvicorn) serves both the API and the frontend static files.

1. **Frontend Build**: The frontend must be built first (`npm run build` in `frontend/`), which
   outputs files to `backend/src/frontend/`.
2. **Static Files**: The application mounts these static files at the root path `/`.

This allows the entire application to be deployed as a single service.

## Troubleshooting

- If `uv sync` fails, verify your Python version matches `backend/.python-version` and
  that `uv` is installed.
- If the API cannot connect to Postgres, ensure Docker is running and the
  compose stack is up. Use `docker compose -f ./docker/docker-compose.yaml logs`
  to inspect services.
- If types in the frontend are out of date, run `npm run generate:types`
  after the backend starts so OpenAPI is available.
- For port conflicts, check if another service is using `8000` or the Postgres port defined in `docker/docker-compose.yaml`.

## References

- **Frontend**: [frontend/README.md](../frontend/README.md)
- **Scripts**: [scripts/README.md](../scripts/README.md)
