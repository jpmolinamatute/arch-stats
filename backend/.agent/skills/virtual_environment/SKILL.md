---
trigger: glob
name: python-virtual-environment
description: Activate the python virtual environment in the backend directory.
globs: backend/**/*.py
---

# Activate Python Virtual Environment

Always activate the python virtual environment before running any python command.

```bash
cd ./backend
source .venv/bin/activate
```

if there is no virtual environment, create one:

```bash
cd ./backend
uv sync
```
