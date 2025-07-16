# Arch-Stats: ChatGPT Instruction File

## Project Overview

Arch-Stats is a monorepo designed for collecting, storing, and analyzing archery performance data. It leverages a FastAPI server (Python 3.12, Pydantic v2.x), a TypeScript frontend (using Vite), and several custom Python sensor modules for real-time shot/event capture. Data is stored in PostgreSQL Docker container and development is optimized for low-latency, hot-reload feedback, and VS Code tasks.

---

## Directory Structure

- `backend/`
  - `.venv/` — Python virtual environment (Python 3.12 and managed by uv)
  - `server/` — FastAPI application with asyncpg, Pydantic v2.x, pytest, Black, isort, mypy, etc.
  - `target_reader/` — Sensor handling modules in Python
- `frontend/`
  - Node/npm project (npm 11.4.1), Vite, TypeScript, Prettier, ESLint
- `docker/`
  - Docker Compose configuration and related scripts
- `.vscode/`
  - VS Code tasks and keybinding
- `.env`
  - All critical environment variables

---

## Development Workflow & Automation

- **Server**:
  - Use Uvicorn (with hot reload) via VS Code Tasks or CLI for serving FastAPI from `backend/server`.
  - Code formatting: Black (4 spaces), import sorting with isort, strict typing with mypy, linting with Pylint.
  - Database interactions use asyncpg; real-time events managed with PostgreSQL LISTEN/NOTIFY.
  - All config/testing tools expect a single `.venv` under `backend/`.
  - Testing: `pytest` configured, type checking enforced in CI and on save.
- **Frontend**:
  - Use Vite (`npm run dev`) for hot-reload development, configured with Prettier (4 spaces), ESLint, and VS Code integration.
  - TypeScript workspace version is enforced.
  - API calls proxied to backend during development via Vite proxy settings.
  - Build output is served statically from the backend in production.
- **Target Reader**:
  - Same as Server but code location is under `backend/target_reader`.
- **Bow Reader**:
  - Same as Server but code location is under `backend/bow_reader`.
- **Arrow Reader**:
  - Same as Server but code location is under `backend/arrow_reader`.
- **Docker**:
  - Docker Compose is used for orchestrating DB and supporting services.
  - VS Code Tasks exist for `up`, `down`, and full volume teardown.
- **VS Code**:
  - Editor set to format on save.
  - ESLint, Prettier, Python, Docker extensions recommended and used for diagnostics and code actions.
  - VS Code Tasks automate starting/stopping backend, frontend, and Docker Compose in correct dependency order.

---

## Coding Standards

- **Python**:

  - Black for formatting, isort for imports, mypy for typing, Pylint for linting.
  - Pydantic schemas use **v2.x** syntax and features (including new type system).
  - All code must use 4-space indentation.

- **TypeScript/JavaScript**:
  - Prettier and ESLint enforce style (4 spaces, semi, singleQuote, etc.).
  - Vite as the build system, with workspace TypeScript enforced.

---

## Environment

- **Python**: 3.12 (all back-end development and venv)
- **npm**: 11.4.1 (all frontend development)
- **Pydantic**: 2.x
- **OS**: Linux

---

## Troubleshooting & Preferences

- Prefer **concise, actionable explanations and code snippets**.
- Show debugging and problem-solving steps explicitly and in order.
- Assume all commands and paths are for Linux environments.
- When suggesting code, **reflect current project structure and config files**.
- For VS Code, always suggest settings that fit a monorepo
- Emphasize automation, hot reload, and maximizing feedback in development.

---

## Testing & CI

- **Backend**: pytest (in `backend/server/tests`), mypy, Black, isort, and Pylint as part of the test and commit workflow.
- **Frontend**: (Fill in your actual test runner, e.g., Vitest, Jest, or leave blank if not used yet.)

---

## Special Notes

- API and database models follow Pydantic v2.x standards.
- TypeScript is always configured for project-wide consistency (via workspace version in `frontend`).
- Environment variables are stored in `./.env` at the project root. There are two symlinks in `./frontend/.env` and `./backend/.env` both point to `./.env`.
