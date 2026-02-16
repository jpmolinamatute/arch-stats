---
name: backend-virtual-environment
description: Activate the python virtual environment in the backend directory.
---

# Activate Python Virtual Environment

Most task can be done via uv, but in certain cases you may want to run python commands
directly, in that case you need to activate the python virtual environment first.

```bash
cd ./backend
source .venv/bin/activate
```

if there is no virtual environment, create one:

```bash
cd ./backend
uv sync
```
