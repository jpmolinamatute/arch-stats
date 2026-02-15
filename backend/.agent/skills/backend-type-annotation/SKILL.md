---
name: backend-type-annotation
description: How to use ty to check type annotation in Python files
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
