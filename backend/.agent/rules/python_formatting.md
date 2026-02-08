---
name: python-formatting
trigger: glob
description: Always run Ruff to Format Python files
globs: backend/**/*.py
---

# Python Formatting

We use ruff to format all Python files. We also use ./backend/pyproject.toml to configure Ruff.
There are two ways to run formatting

1. Manually:

   ```bash
   cd ./backend
   uv run ruff format --config ./pyproject.toml
   ```

2. Via script, this will also run lint, type annotation check and tests:

   ```bash
   ./scripts/linting.bash --backend
   ```
