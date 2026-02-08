---
name: python-tests
trigger: glob
description: Always run pytest in Python files
globs: backend/**/*.py
---

# Python Tests

We use Pytest for unit testing. We also use ./backend/pyproject.toml to configure Pytest.
There are two ways to run tests check:

1. Manually:

   ```bash
   docker compose -f ./docker/docker-compose.yaml up -d  # this will start the PostgreSQL Database
   cd ./backend
   uv run pytest -vv
   cd -
   docker compose -f ./docker/docker-compose.yaml down  # this will stop the PostgreSQL Database
   ```
  
2. Via script, this will also run formatting, lint and type annotation check:

   ```bash
   ./scripts/linting.bash --backend
   ```
