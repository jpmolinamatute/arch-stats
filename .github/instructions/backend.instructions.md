---
applyTo: "**/*.py"
---

# Backend Development Instructions for GitHub Copilot

These instructions define how GitHub Copilot must generate and modify backend code in this repository.

## Project Overview

- The backend is a **FastAPI** application located under `backend/src/`.
- The entry point is `backend/src/app.py`.
- The Python version used by the project is defined in `backend/.python-version`.
- Library versions are defined in `backend/pyproject.toml`.
- PostgreSQL and database migrations are handled through Docker Compose.
- The backend uses a **virtual environment** located at `backend/.venv`.
- Python packages are managed using **`uv`**.
- All code must comply with the configured **code quality tools**:
  - `black` for formatting
  - `isort` for import ordering
  - `mypy` for static type checking
  - `pylint` for linting and style enforcement
  - These tools can be run manually or via the script:
    `./scripts/linting.bash --lint-backend`

## Environment Setup

Copilot must respect and operate within the project’s virtual environment.

### Virtual Environment Activation

Before running any command or Python process, Copilot must activate the backend virtual environment:

```bash
source ./backend/.venv/bin/activate
```

### Package Management

All Python package operations (install, upgrade, remove, etc.) must be performed using **`uv`**, not pip or poetry.
For example:

```bash
uv add fastapi
uv remove sqlalchemy
```

Copilot must never suggest raw `pip install` commands.

### Docker

Start the backend environment using Docker Compose:

```bash
docker compose -f ./docker/docker-compose.yaml up --build
```

Or automatically via the VS Code task: **`Start Docker Compose`**.

Stop the environment manually:

```bash
docker compose -f ./docker/docker-compose.yaml down
# or
docker compose -f ./docker/docker-compose.yaml down -v
```

Or via the VS Code tasks: **`Stop Docker Compose`** or **`Stop & Remove Volumes`**.

### FastAPI Dev Server

Run the FastAPI server after Docker is running and the virtual environment is activated:

```bash
source ./backend/.venv/bin/activate
cd ./backend/src
exec uvicorn --loop asyncio --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory --limit-concurrency 10 app:run
```

Or automatically via the VS Code task: **`Start Uvicorn Dev Server`**.

## Code Structure Rules

Copilot must follow these structural and architectural conventions.

### Directories

- **`backend/src/routers/`**:
  Define only route declarations and HTTP status handling.
  Keep routers small and free of business logic.
- **`backend/src/models/`** and **`backend/src/core/`**:
  Implement all business logic here.
  Use clean separation of concerns.
- **`backend/src/schema/`**:
  Contain all **Pydantic** models (v2) used for I/O validation.
  - Models that validate data **coming from the outside world to the DB**.
  - Models that validate data **coming from the DB to the outside world**.

## Coding Standards

Copilot must adhere to the following rules and conventions when generating or modifying backend Python code.

1. **Follow [Google’s Python Style Guide](./google_standards.md)**.
2. Every module, class, and function must be **fully typed**.
   `mypy` should pass without `# type: ignore` comments.
3. Always place **import statements at the top** of the file.
4. Use **Pydantic models** for all I/O boundaries (never dataclasses).
5. Use **Pydantic v2 syntax**:

   ```python
   model_config = ConfigDict(...)
   ```

   Do not use an inner `Config` class.

6. Validate all inputs at router boundaries using Pydantic schemas.
7. Return **precise HTTP status codes** using `HTTPException` as needed.
8. Keep **database access isolated** in small async functions under `models/` or `core/`.
9. Use **`asyncpg`** for database access. Use prepared statements when beneficial.
10. Avoid global state. Pass database pools, clients, and other dependencies through **dependency injection** or startup wiring.
11. Ensure Copilot **uses the Python version** defined in `backend/.python-version`.
12. Use library versions and dependencies defined in `backend/pyproject.toml`.
13. Always activate the virtual environment before running or testing code.
14. Manage packages exclusively through **`uv`** commands.

### Code Quality Requirements

All backend code must pass the following tools without errors or warnings:

- isort — import sorting (uses pyproject.toml for settings)
- black — code formatting
- mypy — static type checking
- pylint — linting and style enforcement

Copilot must write code compatible with all these tools.

#### Running Tools Individually

First, activate the virtual environment and move into the backend directory:

```bash
source ./backend/.venv/bin/activate
cd backend
```

Then run:

```bash
isort --settings-file ./pyproject.toml ./src ./tests
black --config ./pyproject.toml ./src ./tests
mypy --config-file ./pyproject.toml ./src ./tests
pylint --rcfile ./pyproject.toml ./src ./tests
```

#### Running All at Once

To execute all code quality checks in one step, run the following from the project root directory:

```bash
./scripts/linting.bash --lint-backend
```

Copilot must generate code that passes all these checks.

### Example Endpoint Pattern

Copilot must model all FastAPI endpoints according to this style and structure:

```python
from fastapi import APIRouter, Depends, Response, status, HTTPException, Request
from uuid import UUID
from models import ArcherModel
from schemas import ArcherRead

router = APIRouter()

async def require_auth(request: Request) -> UUID:
    """Validate JWT from auth cookie and return authenticated archer ID.

    Raises:
        HTTPException: If the cookie is missing or invalid (401).
    """
    token = request.cookies.get("arch_stats_auth")
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        )
    try:
        sub = decode_token(token, "sub")
        if not isinstance(sub, str):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User is not authorized to use this endpoint",
            )
        return UUID(sub)
    except Exception as exc:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User is not authorized to use this endpoint",
        ) from exc


@router.get("/{archer_id}", response_model=ArcherRead, status_code=status.HTTP_200_OK)
async def get_archer(
    archer_id: UUID,
    current_archer_id: UUID = Depends(require_auth),
    archer_model: ArcherModel = Depends(),
) -> ArcherRead:
    """Retrieve one archer by ID. Enforces caller authorization."""
    if current_archer_id != archer_id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")

    try:
        return await archer_model.get_one(archer_id)
    except DBNotFound as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Archer not found",
        ) from e
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail="Archer not found",
        ) from e
```

**Copilot must respect these design principles:**

- The `response_model` matches the function’s return type.
- No hardcoded HTTP responses.
- Return consistent `HTTPException` structures.
- Validate user authorization at the router boundary.
- Business logic resides in models or core modules, not routers.

## Summary

Copilot must:

- Activate the virtual environment from `backend/.venv`.
- Use `uv` for all Python dependency management.
- Infer Python version from `backend/.python-version`.
- Infer dependency versions from `backend/pyproject.toml`.
- Keep routers minimal, typed, and validated.
- Centralize business logic in `models` or `core`.
- Use Pydantic v2 and asyncpg properly.
- Follow Google’s style guide.
- Maintain full typing and static validation compatibility.
- Run linting and formatting tools without errors.
