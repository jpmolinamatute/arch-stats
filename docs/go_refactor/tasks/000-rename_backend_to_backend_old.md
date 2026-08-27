# Task 000: Rename `backend/` to `backend-old/`

## Git Branch

`refactor/000-rename-backend-to-backend-old`

## Objective

Rename the existing `backend/` directory to `backend-old/` so it remains available as a
reference during the Go port. This is a prerequisite for creating the new Go `backend/`
directory in the next task.

## Dependencies

**None**: this is the first task.

## Acceptance Criteria

- [x] The directory `backend-old/` exists at the project root and contains all original Python
  backend code (src/, tests/, migrations/, pyproject.toml, etc.).
- [x] The directory `backend/` no longer exists.
- [x] All references to `backend/` in project configuration files are updated to `backend-old/`:
    - `.github/workflows/backend_linting.yaml`
    - `.github/workflows/build_artifact.yaml`
    - `docker/docker-compose.yaml` (migrations volume mount)
    - `.gitmodules` (if applicable)
    - `.vscode/tasks.json` (if it references backend paths)
    - `scripts/start_uvicorn.bash`
    - `scripts/linting.bash`
    - `scripts/generate_fe_types.bash`
    - `cspell.json` (if it references backend paths)
- [x] The Python virtual environment inside `backend-old/.venv/` still works (no broken
  symlinks from the rename).
- [x] CI workflows that reference the backend directory are updated so they still pass
  (even if they run against `backend-old/`).

## Files to Modify

| Action | Path |
| ------ | ---- |
| Rename | `backend/` → `backend-old/` |
| Modify | `.github/workflows/backend_linting.yaml` |
| Modify | `.github/workflows/build_artifact.yaml` |
| Modify | `docker/docker-compose.yaml` |
| Modify | `scripts/start_uvicorn.bash` |
| Modify | `scripts/linting.bash` |
| Modify | `scripts/generate_fe_types.bash` |
| Modify | Any other files that reference `backend/` paths |

## Steps

- [x] **Step 1: Identify all references to `backend/`**

  ```bash
  grep -rn "backend/" --include="*.yaml" --include="*.yml" --include="*.bash" \
    --include="*.json" --include="*.toml" --include="*.md" . \
    | grep -v "backend-old" | grep -v ".git/" | grep -v "node_modules"
  ```

- [x] **Step 2: Rename the directory**

  ```bash
  git mv backend backend-old
  ```

- [x] **Step 3: Update all configuration files**

  Replace every `backend/` reference with `backend-old/` in the files identified in Step 1.
  Pay special attention to:
    - Workflow `working-directory` fields
    - Volume mount paths in docker-compose
    - Script paths in bash scripts

- [x] **Step 4: Verify the Python environment still works**

  ```bash
  cd backend-old
  source .venv/bin/activate
  python -c "import fastapi; print('OK')"
  ```

- [x] **Step 5: Run existing linting to verify nothing broke**

  ```bash
  cd backend-old
  source .venv/bin/activate
  uv run ruff check src/
  ```

- [x] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "refactor: rename backend/ to backend-old/ for Go port reference"
  ```

## Verification

- `ls -la backend-old/` shows the full Python backend.
- `ls backend/` returns "No such file or directory".
- `grep -rn "backend/" .github/ scripts/ docker/ | grep -v backend-old | grep -v ".git/"` returns
  no results (all references updated).
