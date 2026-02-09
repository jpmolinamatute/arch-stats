---
name: package-management
trigger: glob
description: When installing/uninstalling/update python packages
globs: backend/**/*.py
---

# Python Package Management

We use uv to install, uninstall and update python packages. We also use ./backend/pyproject.toml
to configure uv.

## Installing dependencies

```bash
cd ./backend
uv add <package_name>
# or
uv add --dev <package_name>
```

## Removing dependencies

```bash
cd ./backend
uv remove <package_name>
# or
uv remove --dev <package_name>
```

## Updating dependencies

```bash
cd ./backend
uv sync --upgrade
```
