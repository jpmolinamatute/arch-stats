# Delete `backend-old/` Directory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Permanently delete the `backend-old/` directory and remove all stale references across scripts, configuration files, and documentation, completing the Go port repository transition and marking Task 036 as completed.

**Architecture:** Remove `backend-old/` via `git rm -r` and delete any untracked local artifacts (`rm -rf backend-old/`). Clean up active configuration files (`.gitignore`, `.vscode/settings.json`, `.vscode/launch.json`, `scripts/generate_fe_types.bash`, `scripts/README.md`, `README.md`, and `frontend/README.md`) so that no active tooling, scripts, or workflows reference `backend-old` or legacy Uvicorn/Python artifacts. Run full Go and scripts test suites to verify integrity, then update task tracking documents and create a clean git commit.

**Tech Stack:** Git, Bash, Go 1.27+, VS Code configuration, npm / Vite.

**Spec:** [docs/go_refactor/tasks/036-delete_backend_old.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/036-delete_backend_old.md)

## Global Constraints

- Target branch: `refactor/036-delete-backend-old`
- Directory to delete: `backend-old/` (entire directory, tracked and untracked)
- Zero active references to `backend-old/` in `.github/workflows/`, `scripts/`, `docker/`, `.agent/`, `.vscode/`, root configurations, or `frontend/README.md` (historical references in `docs/` are permitted)
- All Go tests (`cd backend && go test ./... -v -count=1`), `go vet ./...`, and `go build ./cmd/arch-stats` must pass cleanly
- At the end of implementation, mark all acceptance criteria and steps in [docs/go_refactor/tasks/036-delete_backend_old.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/036-delete_backend_old.md) as completed (`[x]`) and update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking all tasks as `DONE`
- Clean git commit message matching convention: `chore: delete backend-old/ Python reference code — Go port complete`

---

## File Structure

```
.
├── .gitignore                           # [MODIFY] Remove backend-old log ignore rules
├── .vscode/
│   ├── launch.json                      # [MODIFY] Replace Python uvicorn configs with Go debug config
│   └── settings.json                    # [MODIFY] Remove Python/ruff/pytest/ty backend-old settings
├── README.md                            # [MODIFY] Update Backend (Legacy) link to Backend
├── backend-old/                         # [DELETE] Entire legacy Python backend directory
├── docs/
│   ├── go_refactor/
│   │   └── tasks/
│   │       └── 036-delete_backend_old.md # [MODIFY] Mark all acceptance criteria & steps as done
│   └── plans/
│       ├── 2026-09-06-delete-backend-old.md # [NEW] This implementation plan document
│       └── task.md                      # [MODIFY] Live checklist table tracking Task 036
├── frontend/
│   └── README.md                        # [MODIFY] Replace legacy Uvicorn/Python references with Go backend
└── scripts/
    ├── README.md                        # [MODIFY] Update backend merge requirements to Go
    └── generate_fe_types.bash           # [MODIFY] Remove legacy Python openapi_via_script fallback
```

---

### Task 1: Git Branch Setup and Pre-Deletion Baseline Verification

**Interfaces:**

- Consumes: `main` git branch with existing uncommitted changes in
  `docs/go_refactor/tasks/036-delete_backend_old.md`
- Produces:
    - Feature branch `refactor/036-delete-backend-old`
    - Active task tracking table in `docs/plans/task.md`
    - Verified passing baseline for `backend` Go build, vet, and tests

- [x] **Step 1: Check out feature branch `refactor/036-delete-backend-old`**

Run:

```bash
git switch -c refactor/036-delete-backend-old
```

- [x] **Step 2: Initialize live task tracker in `docs/plans/task.md`**

Replace `docs/plans/task.md` with:

```markdown
| Task | Status | Description |
| ---- | ------ | ----------- |
| Task 1: Git Branch Setup & Baseline Verification | IN_PROGRESS | Branch `refactor/036-delete-backend-old`, initialize task.md, baseline test suite |
| Task 2: Clean Active Configuration & Script References | PENDING | Remove backend-old references in .gitignore, .vscode, scripts, and READMEs |
| Task 3: Delete `backend-old/` Directory | PENDING | Remove `backend-old/` from git and disk, verify zero active references remain |
| Task 4: Full Repository Build & Test Suite Verification | PENDING | Run Go build, vet, test, script linting, and verify clean workspace |
| Task 5: Mark Task 036 Completed & Commit | PENDING | Update task 036 checklist [x], task.md to DONE, commit chore |
```

- [x] **Step 3: Run Go baseline verification**

Run:

```bash
cd backend && go build ./cmd/arch-stats && go vet ./... && go test ./... -v -count=1
```

Expected: PASS for all packages.

- [x] **Step 4: Update `docs/plans/task.md` for Task 1 completion**

Mark Task 1 as `DONE` in `docs/plans/task.md`.

---

### Task 2: Clean Active Configuration & Script References

**Files:**

- Modify: `scripts/generate_fe_types.bash`
- Modify: `scripts/README.md`
- Modify: `.gitignore`
- Modify: `.vscode/settings.json`
- Modify: `.vscode/launch.json`
- Modify: `README.md`
- Modify: `frontend/README.md`

**Interfaces:**

- Consumes: Existing files containing stale `backend-old` / legacy Python configurations
- Produces: Clean configuration and script files with valid syntax and no dead references

- [x] **Step 1: Clean `scripts/generate_fe_types.bash`**

Remove the `openapi_via_script` function and its fallback invocation in `main()`.
In `scripts/generate_fe_types.bash`:
Remove lines 22-30:

```bash
openapi_via_script() {
    local openapi_source="$1"
    (
        cd "${ROOT_DIR}/backend-old"
        export PYTHONPATH="${ROOT_DIR}/backend-old:${ROOT_DIR}/backend-old/src"
        echo "Info: Generating OpenAPI spec from script"
        uv run ./tools/generate_openapi.py "${openapi_source}"
    )
}
```

And replace lines 62-70 with:

```bash
        if [[ -f "${ROOT_DIR}/backend/specs/swagger.json" ]]; then
            openapi_source="${ROOT_DIR}/openapi.json"
            convert_swagger_to_openapi "${ROOT_DIR}/backend/specs/swagger.json" "${openapi_source}"
        else
            openapi_source="${ROOT_DIR}/openapi.json"
        fi
```

Verify bash syntax:

```bash
bash -n scripts/generate_fe_types.bash
```

- [x] **Step 2: Clean `.gitignore`**

Remove lines 15-16 from `.gitignore`:

```gitignore
backend-old/src/core/logs/*
!backend-old/src/core/logs/.gitkeep
```

- [x] **Step 3: Update `scripts/README.md`**

In `scripts/README.md`:
Replace line 197:

```markdown
| **Backend** | `backend-old/**` | Black, Isort, MyPy, Pylint, Tests |
```

with:

```markdown
| **Backend** | `backend/**` | golangci-lint, gofumpt, go test |
```

Replace line 228:

```markdown
- **Backend (Legacy)**: [backend-old/README.md](../backend-old/README.md)
```

with:

```markdown
- **Backend**: [backend/](../backend)
```

- [x] **Step 4: Update root `README.md`**

In `README.md`:
Replace line 7:

```markdown
- [Backend (Legacy)](./backend-old/README.md)
```

with:

```markdown
- [Backend](./backend)
```

- [x] **Step 5: Clean `.vscode/settings.json`**

Remove all Python/pytest/ruff/ty settings targeting `backend-old`. The resulting
`.vscode/settings.json` content:

```json
{
    "eslint.enable": true,
    "eslint.validate": [
        "javascript",
        "typescript",
        "vue"
    ],
    "eslint.workingDirectories": [
        "./frontend"
    ],
    "eslint.format.enable": true,
    "[shellscript]": {
        "editor.defaultFormatter": "foxundermoon.shell-format",
        "editor.formatOnSave": true
    },
    "shellcheck.enable": true,
    "shellcheck.executablePath": "/usr/bin/shellcheck",
    "shellcheck.run": "onSave",
    "shellcheck.customArgs": [
        "--shell=bash",
        "--color=never",
        "--external-sources"
    ],
    "shellformat.path": "/usr/bin/shfmt",
    "shellformat.flag": "-i 4 -s -ci -w",
    "sqlfluff.dialect": "postgres",
    "sqlfluff.executablePath": "/usr/bin/sqlfluff",
    "sqlfluff.shell": true,
    "typescript.tsdk": "./frontend/node_modules/typescript/lib",
    "typescript.enablePromptUseWorkspaceTsdk": true,
    "typescript.format.placeOpenBraceOnNewLineForFunctions": true,
    "vitest.rootConfig": "./frontend/vitest.config.ts",
    "cSpell.words": [
        "notabool"
    ],
    "typescript.tsserver.enableTracing": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--config=${workspaceFolder}/backend/.golangci.yml"
    ],
    "go.lintOnSave": "workspace",
    "gopls": {
        "formatting.gofumpt": true
    },
    "[go]": {
        "editor.formatOnSave": true,
        "editor.defaultFormatter": "golang.go",
        "editor.codeActionsOnSave": {
            "source.organizeImports": "explicit"
        }
    },
    "go.testFlags": [
        "-v",
        "-count=1"
    ],
    "go.testTimeout": "30s",
    "go.testEnvFile": "${workspaceFolder}/backend/.env"
}
```

Verify JSON validity:

```bash
jq . .vscode/settings.json >/dev/null
```

- [x] **Step 6: Update `.vscode/launch.json`**

Replace legacy Python launch configs with a Go debug launch config:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Go Backend",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/backend/cmd/arch-stats",
            "cwd": "${workspaceFolder}/backend",
            "envFile": "${workspaceFolder}/backend/.env"
        }
    ]
}
```

Verify JSON validity:

```bash
jq . .vscode/launch.json >/dev/null
```

- [x] **Step 7: Update `frontend/README.md`**

In `frontend/README.md`:

1. Line 25: Replace `It communicates with a Python backend (FastAPI/Uvicorn) served at
   http://localhost:8000.` with `It communicates with a Go backend (Chi router) served at
   http://localhost:8000.`
2. Line 34: Replace `npm run build emits to ../backend/src/frontend/.` with `npm run build emits to
   ../backend/frontend/.`
3. Lines 73-78: Replace:

    ```bash
    ./scripts/start_uvicorn.bash
    ```

    with:

    ```bash
    cd backend && air
    ```

4. Line 85: Replace `- Start backend API: Run task Start Uvicorn Server.` with `- Start backend API:
   Run task Start Go Server (air).`
5. Lines 221-238: Replace:

    ```markdown
    In production, the frontend is built into static files and served directly by the backend (Uvicorn).
    
    1. **Build the Frontend**:
    
        ```bash
        npm run build
        ```
    
        This compiles the app to `../backend/src/frontend/`.
    
    2. **Serve with Uvicorn**:
        When you start the backend:
    
        ```bash
        ./scripts/start_uvicorn.bash
        ```
    
        Uvicorn serves the API at `/api/v0` AND the static frontend files at the root `/`.
    ```

    with:

    ```markdown
    In production, the frontend is built into static files and served directly by the Go backend.

    1. **Build the Frontend**:

        ```bash
        npm run build
        ```

        This compiles the app to `../backend/frontend/`.

    2. **Serve with Go Backend**:
        When you run the backend:

        ```bash
        cd backend && go run ./cmd/arch-stats
        ```

        The Go server serves the API at `/api/v0` AND the static frontend files at the root `/`.
    ```

6. Line 247: Replace `../backend/src/frontend/` with `../backend/frontend/`
7. Line 259: Replace `[backend/README.md](../backend/README.md)` with `[backend/](../backend)`

- [x] **Step 8: Update `docs/plans/task.md` for Task 2 completion**

Mark Task 2 as `DONE` in `docs/plans/task.md`.

---

### Task 3: Delete `backend-old/` Directory

**Files:**

- Delete: `backend-old/` (entire directory)

**Interfaces:**

- Consumes: Tracked git directory `backend-old/` and any untracked artifacts
- Produces: Completely removed `backend-old/` tree; zero references in active repository files

- [x] **Step 1: Delete `backend-old/` from git tracking and disk**

Run:

```bash
git rm -r backend-old/
rm -rf backend-old/
```

- [x] **Step 2: Verify `backend-old/` does not exist**

Run:

```bash
ls backend-old/ 2>&1 || true
```

Expected output: `ls: cannot access 'backend-old/': No such file or directory`

- [x] **Step 3: Verify no active references remain in the repository**

Run:

```bash
git grep -i "backend-old" -- ':!docs'
```

Expected output: Empty (exit status 1, zero matches outside `docs/`).

- [x] **Step 4: Update `docs/plans/task.md` for Task 3 completion**

Mark Task 3 as `DONE` in `docs/plans/task.md`.

---

### Task 4: Full Repository Build & Test Suite Verification

**Files:**

- None (verification task)

**Interfaces:**

- Consumes: Cleaned workspace without `backend-old/`
- Produces: Confirmed test results and build artifacts

- [x] **Step 1: Compile the Go backend binary**

Run:

```bash
cd backend && go build -o ./arch-stats ./cmd/arch-stats
```

Expected: Zero errors, binary created.

- [x] **Step 2: Run `go vet`**

Run:

```bash
cd backend && go vet ./...
```

Expected: Zero issues reported.

- [x] **Step 3: Run all Go unit and integration tests**

Run:

```bash
cd backend && go test ./... -v -count=1
```

Expected: All tests PASS.

- [x] **Step 4: Run bash script linting and formatting check**

Run:

```bash
./scripts/linting.bash --scripts
```

Expected: Shellcheck and shfmt pass with zero errors.

- [x] **Step 5: Verify git status**

Run:

```bash
git status
```

Confirm all deleted files and modified files are properly tracked.

- [x] **Step 6: Update `docs/plans/task.md` for Task 4 completion**

Mark Task 4 as `DONE` in `docs/plans/task.md`.

---

### Task 5: Mark Task 036 Completed & Commit

**Files:**

- Modify: `docs/go_refactor/tasks/036-delete_backend_old.md`
- Modify: `docs/plans/task.md`

**Interfaces:**

- Consumes: Verified clean build and passing tests
- Produces:
  - All acceptance criteria and steps checked (`[x]`) in `docs/go_refactor/tasks/036-delete_backend_old.md`
  - All tasks marked `DONE` in `docs/plans/task.md`
  - Clean git commit on `refactor/036-delete-backend-old`

- [x] **Step 1: Update `docs/go_refactor/tasks/036-delete_backend_old.md`**

Mark all acceptance criteria and steps as checked (`[x]`):

```markdown
# Task 036: Delete `backend-old/` Directory

## Git Branch

`refactor/036-delete-backend-old`

## Objective

Delete the `backend-old/` directory now that the Go refactoring is complete, all tests pass,
and the Go backend has been validated. This is the final cleanup step of the repository strategy.

## Dependencies

- All tasks 001–035 must be merged
- All tasks 037–041 (integration tests) should ideally be merged before this
- The Go backend is fully functional and deployed

## Acceptance Criteria

- [x] The `backend-old/` directory is deleted from the repository.
- [x] All references to `backend-old/` are removed from:
    - `.github/workflows/` — no workflow references backend-old
    - `scripts/` — no script references backend-old
    - `docker/` — no docker config references backend-old
    - `.agent/` — no skill references backend-old
    - `docs/` — task files can still reference it historically but no active config
    - `.gitmodules` — if the migrations submodule pointed into backend-old
- [x] The repository compiles, tests pass, and CI pipelines work without `backend-old/`.
- [x] `git log --oneline -1` shows a clean commit message.

## Files to Delete

| Action | Path |
| ------ | ---- |
| Delete | `backend-old/` (entire directory) |

## Steps

- [x] **Step 1: Verify the Go backend is fully functional**

  ```bash
  cd backend
  go build ./cmd/arch-stats
  go test ./... -v -count=1
  go vet ./...
  ```

  All must pass before proceeding.

- [x] **Step 2: Search for any remaining references**

  ```bash
  grep -rn "backend-old" . --include="*.yaml" --include="*.yml" --include="*.bash" \
    --include="*.json" --include="*.toml" --include="*.md" \
    | grep -v ".git/" | grep -v "docs/tasks/"
  ```

  Fix any active references found (docs/tasks/ references are historical and acceptable).

- [x] **Step 3: Delete `backend-old/`**

  ```bash
  git rm -r backend-old/
  ```

- [x] **Step 4: Verify the build still works**

  ```bash
  cd backend
  go build ./cmd/arch-stats
  go test ./... -v -count=1
  ```

- [x] **Step 5: Commit**

  ```bash
  git add -A
  git commit -m "chore: delete backend-old/ Python reference code — Go port complete"
  ```

## Verification

- `ls backend-old/` — "No such file or directory".
- `cd backend && go build ./cmd/arch-stats` — compiles.
- `cd backend && go test ./... -v -count=1` — all tests pass.
- `grep -rn "backend-old" . --include="*.yaml" --include="*.bash" --include="*.json" | grep -v ".git/" | grep -v "docs/tasks/"` — no active references.
```

- [x] **Step 2: Update `docs/plans/task.md`**

Ensure all tasks are marked as `DONE` in `docs/plans/task.md`:

| Task | Status | Description |
| ---- | ------ | ----------- |
| Task 1: Git Branch Setup & Baseline Verification | DONE | Branch `refactor/036-delete-backend-old`, initialize task.md, baseline test suite |
| Task 2: Clean Active Configuration & Script References | DONE | Remove backend-old references in .gitignore, .vscode, scripts, and READMEs |
| Task 3: Delete `backend-old/` Directory | DONE | Remove `backend-old/` from git and disk, verify zero active references remain |
| Task 4: Full Repository Build & Test Suite Verification | DONE | Run Go build, vet, test, script linting, and verify clean workspace |
| Task 5: Mark Task 036 Completed & Commit | DONE | Update task 036 checklist [x], task.md to DONE, commit chore |


- [x] **Step 3: Stage and commit all changes**

Run:

```bash
git add -A
git commit -m "chore: delete backend-old/ Python reference code — Go port complete"
```

- [x] **Step 4: Verify the commit**

Run:

```bash
git log --oneline -1
```

Expected: Clean commit log showing `chore: delete backend-old/ Python reference code — Go port complete`.
