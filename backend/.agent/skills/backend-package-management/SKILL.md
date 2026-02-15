---
name: backend-package-management
description: How to use uv to install/uninstall/update python packages
---

# Python Package Management

We use uv to install, uninstall and update python packages. We also use
./backend/pyproject.toml to configure uv.

## Install dependencies

```bash
cd ./backend
uv add <package_name>
# or
uv add --dev <package_name>
```

## Uninstall dependencies

```bash
cd ./backend
uv remove <package_name>
# or
uv remove --dev <package_name>
```

## Update dependencies

```bash
cd ./backend
uv sync --upgrade
```
