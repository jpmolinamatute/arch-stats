---
name: backend-type-annotation
trigger: glob
description: Always run Ty to check type annotation in Python files
globs: backend/**/*.py
---

# Python Type Annotation

We use ty for type annotation checking. There are two ways to run type annotation check:

1. Manually:

   ```bash
   cd ./backend
   uv run ty check
   ```

2. Via script (run from project root), this will also run formatting, lint and tests:

   ```bash
   ./scripts/linting.bash --backend
   ```
