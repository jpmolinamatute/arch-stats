---
name: uv
description: Manage Python virtual environments, manage dependencies, linting/formatting files,and run scripts.
---

# UV Skill

Any Python command must be run from ./backend directory and use uv.

## Activating the virtual environment

```bash
cd ./backend
source .venv/bin/activate
```

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

## Linting

we use ruff as our linter and it's configured in ./backend/pyproject.toml

```bash
cd ./backend
uv run ruff check --fix --config ./pyproject.toml
```

## Formatting

we use ruff as our formatter and it's configured in ./backend/pyproject.toml

```bash
cd ./backend
uv run ruff format --config ./pyproject.toml
```

## Type Checking

we use ty

```bash
cd ./backend
uv run ty check
```

## Running scripts

```bash
cd ./backend
uv run <script_name>
```
