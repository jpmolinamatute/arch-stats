# Arch-Stats \* Copilot Instructions (Project-Level)

**Audience:** VS Code Copilot  
**Repo:** https://github.com/jpmolinamatute/arch-stats  
**Last updated:** 2025-08-18

## What this project is

Arch-Stats collects and analyzes archery performance data. The frontend is a Vue 3 SPA built with Vite. The backend is a FastAPI service using PostgreSQL 15+ via `asyncpg`. Real-time features use WebSockets.

## How to help

- Prefer **concise, correct** code over cleverness.
- Follow the language-specific rules in the linked instruction files:
  - [Frontend rules](./frontend-instructions.md)
  - [Backend rules](./backend-instructions.md)
  - [Procedures (CI/CD) rules](./procedures-instructions.md)

## Monorepo layout (high level)

- [frontend/](../frontend/): Vue 3 SPA.
- [backend/](../backend/): FastAPI app and sensor modules. Strict typing and linting enforced.
- [docker/](../docker/): Compose for PostgreSQL and local integration.
- [scripts/](../scripts/): A Miscellaneous scripts for various tasks.

## Golden constraints (do not violate)

- **Backend DB access:** use **`asyncpg` only** (no ORMs, no psycopg2).
- **Python models:** **Pydantic v2** with `model_config = ConfigDict(...)` and `extra="forbid"`.
- **Typing:** mypy-strict. All functions must have explicit types.
- **Formatting/Linting:** Python via Black + isort + Pylint; TS via ESLint + Prettier.
- **Frontend build:** output goes to `backend/src/server/frontend/`.
- **Linux-first:** assume Bash + POSIX paths; Raspberry Pi 5 runtime.

## Things to avoid

- Generating code that adds new server-side libraries not already used.
- Synchronous DB calls or ad-hoc threads.
- Leaking implementation across layers (keep FE/BE boundaries clean).

## Useful references

- Project README (user-facing): [README.md](../README.md)
- Backend README (dev-facing): [backend/README.md](../backend/README.md)
- Frontend README (dev-facing): [frontend/README.md](../frontend/README.md)
- Procedures README (dev-facing): [scripts/README.md](../scripts/README.md)
