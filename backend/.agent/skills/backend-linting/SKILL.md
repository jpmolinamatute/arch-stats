---
name: backend-linting
description: How to run Ruff to lint python files
---

# Python Linting

We use ruff to lint all Python files. We also use ./backend/pyproject.toml to configure Ruff.
There are two ways to run linting

1. Manually:

   ```bash
   cd ./backend
   uv run ruff check --fix --config ./pyproject.toml
   ```

2. Via script (run from project root), this will also run formatting, type annotation check and tests:

   ```bash
   ./scripts/linting.bash --backend
   ```
