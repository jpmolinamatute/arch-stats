# Arch-Stats Backend

> This directory contains all Python runtime modules that power Arch-Stats: the API surface and the sensor simulation processes. This README gives a unified, high level onboarding view. For project vision see [../README.md](../README.md); for coding rules see [.github/instructions/backend.instructions.md](../.github/instructions/backend.instructions.md).

## Table of Contents

- [Arch-Stats Backend](#arch-stats-backend)
  - [Table of Contents](#table-of-contents)
  - [TL;DR Quick Start](#tldr-quick-start)
  - [Architecture](#architecture)
  - [What You'll Need](#what-youll-need)
  - [Setup: Step-by-Step](#setup-step-by-step)
    - [1. Install Dependencies](#1-install-dependencies)
    - [2. Activate Venv](#2-activate-venv)
    - [3. Start PostgreSQL](#3-start-postgresql)
    - [4. Seed / Simulate (optional)](#4-seed--simulate-optional)
  - [Running Processes](#running-processes)
  - [Project Structure](#project-structure)
  - [Managing dependencies](#managing-dependencies)
    - [Add new dependency](#add-new-dependency)
    - [Update all dependencies](#update-all-dependencies)
    - [Update a dependency](#update-a-dependency)
    - [Remove a dependency](#remove-a-dependency)
  - [Quality: Lint, Type, Test](#quality-lint-type-test)
  - [Testing Details](#testing-details)
  - [Troubleshooting](#troubleshooting)
  - [Contribution \& Change Checklist](#contribution--change-checklist)

## TL;DR Quick Start

```bash
cd backend
uv sync --dev --python $(cat ./.python-version)
source .venv/bin/activate
docker compose -f docker/docker-compose.yaml up -d
./scripts/start_uvicorn.bash          # or VS Code task
./scripts/start_archy_bot.bash        # optional simulated shots
pytest -q
```

## Architecture

All Python runtime modules communicate *only* through PostgreSQL (tables + `LISTEN/NOTIFY`). There are **no direct cross imports** between runtime modules; this preserves loose coupling and makes individual processes restartable and testable in isolation.

Canonical data flow:

```text
producer module (e.g. sensor simulator) -> parameterized INSERT/UPDATE -> PostgreSQL NOTIFY -> (optional real-time consumer) -> frontend
```

Design constraints (enforced in review):

- Database access: `asyncpg` only (no ORM, no psycopg2, no synchronous DB calls).
- Pydantic v2 models with `extra="forbid"` at I/O boundaries.
- Strict typing + mypy clean (avoid unnecessary `Any`).
- SQL is explicit & parameterized (no string interpolation of values).
- No global mutable state shared across modules; configuration via environment.

## What You'll Need

- **uv** (environment + dependency manager)
- **Python 3.13** (see [.python-version](./.python-version))
- **Docker / Docker Compose**
- **Linux** (native or WSL)
- Optional: VS Code tasks (see [.vscode/](../.vscode/))

## Setup: Step-by-Step

### 1. Install Dependencies

```bash
cd backend
uv sync --dev --python $(cat ./.python-version)
```

### 2. Activate Venv

```bash
source .venv/bin/activate
python --version  # expect 3.13.x
```

### 3. Start PostgreSQL

```bash
docker compose -f docker/docker-compose.yaml up -d
# Stop: docker compose -f docker/docker-compose.yaml down
```

### 4. Seed / Simulate (optional)

Run sensor bots after DB is up (see [scripts/](../scripts/)).

## Running Processes

Runtime programs are started via helper scripts in `scripts/` (see file headers for usage) or via editor tasks. All processes share the same virtual environment and database. Start PostgreSQL before launching any process.

## Project Structure

```text
backend/
├── pyproject.toml
├── scripts/                 # automation (linting, start, etc.)
├── docker/                  # docker-compose + configs
├── src/
│   ├── server/              # FastAPI app + routers + db pool
│   ├── arrow_reader/        # arrow simulator
│   ├── bow_reader/          # bow simulator
│   ├── target_reader/       # target simulator
│   └── shared/              # shared utilities (lightweight)
└── tests/                   # pytest suites (models + endpoints)
```

## Managing dependencies

The backend uses [UV](https://docs.astral.sh/uv/), [pyproject.toml](./pyproject.toml) and [uv.lock](./uv.lock) to manage dependencies

### Add new dependency

```bash
uv add <package>
uv sync
```

### Update all dependencies

Update every dependency in the project (within existing version constraints) to the newest compatible versions and refresh the virtual environment:

```bash
# Re-resolve all versions to latest allowed & update lock file
uv lock --upgrade

# Apply the new lock to the virtual environment
uv sync
```

### Update a dependency

Upgrade a single package (and its transitive dependencies) without touching the rest:

```bash
uv add --upgrade <package>
# or explicitly target a version / specifier
uv add <package>@<version>
```

After upgrading:

```bash
./scripts/linting.bash   # ensure formatting, lint, type, tests still pass
```

If an upgrade introduces breaking changes, revert selectively by editing the version spec in `pyproject.toml` (or pinning with `==`) and re-running `uv lock && uv sync`.

### Remove a dependency

Remove a package and update the environment:

```bash
uv remove <package>
uv sync
```

Clean up or refactor affected modules, then run the full quality suite:

```bash
./scripts/linting.bash
```

If the package provided type stubs or runtime side-effects (e.g. registering Pydantic validators), make sure an alternative exists before removal. For multi-module usage (server + readers), confirm none of the other three modules still require it.

## Quality: Lint, Type, Test

| Tool | Command |
| ---- | ------- |
| Black | `black --config pyproject.toml ./src` |
| isort | `isort --settings-file pyproject.toml ./src` |
| mypy | `mypy --config-file pyproject.toml ./src` |
| Pylint | `pylint --rcfile pyproject.toml ./src` |
| Pytest | `pytest -v` |
| All (script) | `./scripts/linting.bash` |

CI mirrors these (see workflows in [.github/workflows/backend_linting.yaml](../.github/workflows/backend_linting.yaml)).

## Testing Details

- Tests run locally against the active PostgreSQL container.
- HTTP-facing behavior uses in-process clients (no external network calls).
- Shared fixtures live in `tests/conftest.py`.
- Prefer deterministic factories and parametrization over ad-hoc loops.

## Troubleshooting

| Symptom | Fix |
| ------- | --- |
| Connection errors | Confirm container: `docker ps`; check env vars; ensure pool init in startup |
| Failing mypy on Pydantic models | Ensure `model_config = ConfigDict(extra="forbid")` and correct field types |
| Hanging test suite | Look for un-awaited coroutines or unclosed DB connections |
| Real-time / async issues | Verify event loop configuration and graceful shutdown settings |

Quick DB sanity:

```bash
docker exec -it arch-stats-postgres psql -U postgres -c '\dt'
```

## Contribution & Change Checklist

1. Open issue / confirm scope.
2. Implement minimal change (no drive-by refactors).
3. Maintain module isolation (no cross sensor imports).
4. Add/extend tests (success + failure cases).
5. Run lint/type/test script clean.
6. Update docs / README if behavior or setup changes.
7. Regenerate frontend API types if OpenAPI changed.
8. Submit PR referencing issue; include concise rationale.
