# Task 031: Configure Air for Hot Reload & Dev Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configure `air` for Go live-reload development in `backend/.air.toml`, ignore temporary build artifacts in `backend/.gitignore`, update `.vscode/tasks.json` to replace legacy Uvicorn tasks with the Go `air` live-reload task, verify `air` builds and restarts correctly, mark all tasks in `docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md` and `docs/plans/task.md` as done, and commit changes to `refactor/031-air-hot-reload-and-dev-environment`.

**Architecture:**
- `backend/.air.toml`:
  - Declarative configuration for `air` (the Go live-reloader).
  - Watches `cmd/` and `internal/` for `.go` and `.toml` changes, building the binary to `./tmp/arch-stats`.
  - Excludes non-source directories (`tmp/`, `tests/`, `vendor/`, `specs/`, `migrations/`, `frontend/`) and `_test.go` files from triggering full server rebuilds.
  - Enables colored log output and sets clean exit options.
- `backend/.gitignore`:
  - Ignores `tmp/` directory created by `air` and `build-errors.log`.
- `.vscode/tasks.json`:
  - Replaces legacy "Start Uvicorn Server" task with "Start Go Server (air)", invoking `air` inside `${workspaceFolder}/backend`.
  - Confirms "Run Go Tests" and "Start Vite Server" tasks are intact and functional.
- `docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md`:
  - Marked completely done (`[x]`) upon successful verification.
- `docs/plans/task.md`:
  - Live table tracker updated per task and marked `DONE`.

**Tech Stack:** Go 1.27+, `github.com/air-verse/air`, VS Code Task Runner, Git.

**Spec:**
- [docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md)
- [.vscode/tasks.json](file:///home/juanpa/Projects/arch-stats/.vscode/tasks.json)
- [.agent/skills/backend-run-server/SKILL.md](file:///home/juanpa/Projects/arch-stats/.agent/skills/backend-run-server/SKILL.md)

## Global Constraints

- Git branch: `refactor/031-air-hot-reload-and-dev-environment` branched from latest `main`.
- `backend/.air.toml` must watch `cmd/` and `internal/` directories for `.go` and `.toml` changes.
- Build command must be `go build -o ./tmp/arch-stats ./cmd/arch-stats`.
- Run command must be `./tmp/arch-stats`.
- `.vscode/tasks.json` must contain zero references to `uvicorn`.
- `backend/.gitignore` must ignore `tmp/` and `build-errors.log`.
- Single-flow execution: exactly one active task tracked in `docs/plans/task.md`.
- Final step must mark all acceptance criteria and steps in `docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md` as completed (`[x]`).

---

## File Structure

```
.vscode/
└── tasks.json                           # [MODIFY] Replace "Start Uvicorn Server" with "Start Go Server (air)"
backend/
├── .air.toml                            # [NEW] Air configuration for watching, building, and running Go server
└── .gitignore                           # [NEW] Ignore tmp/ and build-errors.log
docs/
├── plans/
│   ├── task.md                          # [MODIFY] Track Task 031 live checklist progress (table-only)
│   └── 2026-09-05-air-hot-reload-and-dev-environment.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 031-air_hot_reload_and_dev_environment.md # [MODIFY] Mark all acceptance criteria and steps as checked [x]
```

---

## Task Structure

### Task 1: Git Branch Setup & Environment Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: `main` branch HEAD
- Produces: `refactor/031-air-hot-reload-and-dev-environment` branch, `air` binary in PATH, initialized `docs/plans/task.md` tracker

- [ ] **Step 1: Create and switch to the refactor git branch**

```bash
git checkout -b refactor/031-air-hot-reload-and-dev-environment
```

- [ ] **Step 2: Initialize `docs/plans/task.md` live tracker**

Write the initial status table for Task 031 to `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Environment Initialization | IN_PROGRESS | Create branch `refactor/031-air-hot-reload-and-dev-environment` and install `air` |
| Task 2: Configure Air & Ignore Temporary Build Artifacts | PENDING | Create `backend/.air.toml` and `backend/.gitignore` |
| Task 3: Update VS Code Tasks Configuration | PENDING | Replace Uvicorn task with "Start Go Server (air)" in `.vscode/tasks.json` |
| Task 4: Integration Verification (Air & Tasks) | PENDING | Test `air` live-reload build cycle and verify VS Code tasks syntax |
| Task 5: Mark Tasks as Done & Final Commit | PENDING | Mark `031-air_hot_reload_and_dev_environment.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Ensure `air` tool is installed in Go bin path**

Run:
```bash
go install github.com/air-verse/air@latest
air -v
```
Expected: `air` prints version information (e.g. `air version v1.x.x ...`).

- [ ] **Step 4: Mark Task 1 as DONE in `docs/plans/task.md`**

Update `Task 1` status to `DONE` and `Task 2` status to `IN_PROGRESS`.

---

### Task 2: Configure Air & Ignore Temporary Build Artifacts

**Files:**
- Create: `backend/.air.toml`
- Create: `backend/.gitignore`

**Interfaces:**
- Consumes: Task 031 specification requirements
- Produces: Valid `backend/.air.toml` and `backend/.gitignore`

- [ ] **Step 1: Create `backend/.air.toml`**

Create `backend/.air.toml` with the exact configuration matching the Task 031 specification:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/arch-stats ./cmd/arch-stats"
bin = "./tmp/arch-stats"
full_bin = "./tmp/arch-stats"
include_ext = ["go", "toml"]
exclude_dir = ["tmp", "tests", "vendor", "specs", "migrations", "frontend"]
exclude_regex = ["_test\\.go"]
delay = 1000
log = "build-errors.log"
send_interrupt = true
kill_delay = 500

[log]
time = true

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"

[misc]
clean_on_exit = true
```

- [ ] **Step 2: Create `backend/.gitignore`**

Create `backend/.gitignore` with:

```txt
tmp/
build-errors.log
```

- [ ] **Step 3: Verify git ignores `backend/tmp/` and `backend/build-errors.log`**

Test:
```bash
mkdir -p backend/tmp
touch backend/tmp/dummy.bin backend/build-errors.log
git status --porcelain backend/tmp backend/build-errors.log
```
Expected: No output (files are ignored by Git).
Then clean up dummy test files:
```bash
rm -rf backend/tmp backend/build-errors.log
```

- [ ] **Step 4: Mark Task 2 as DONE in `docs/plans/task.md`**

Update `Task 2` status to `DONE` and `Task 3` status to `IN_PROGRESS`.

---

### Task 3: Update VS Code Tasks Configuration

**Files:**
- Modify: `.vscode/tasks.json`

**Interfaces:**
- Consumes: Existing `.vscode/tasks.json`
- Produces: Updated `.vscode/tasks.json` without Uvicorn references and with `Start Go Server (air)`

- [ ] **Step 1: Replace "Start Uvicorn Server" in `.vscode/tasks.json`**

In `.vscode/tasks.json`, replace:
```json
            {
                "label": "Start Uvicorn Server",
                "type": "shell",
                "command": "${workspaceFolder}/scripts/start_uvicorn.bash",
                "problemMatcher": [],
                "presentation": {
                    "echo": true,
                    "reveal": "always",
                    "close": false,
                    "focus": false,
                    "panel": "dedicated"
                },
                "detail": "Start FastAPI server with Uvicorn in backend-old/server"
            },
```
With:
```json
            {
                "label": "Start Go Server (air)",
                "type": "shell",
                "command": "air",
                "options": {
                    "cwd": "${workspaceFolder}/backend"
                },
                "presentation": {
                    "echo": true,
                    "reveal": "always",
                    "close": false,
                    "focus": false,
                    "panel": "dedicated"
                },
                "problemMatcher": [],
                "detail": "Start Go backend with air for live-reload"
            },
```

- [ ] **Step 2: Verify "Run Go Tests" and "Start Vite Server" tasks are present and properly structured**

Check that `.vscode/tasks.json` retains:
- "Start Docker Compose"
- "Stop Docker Compose"
- "Stop & Remove Volumes"
- "Start Go Server (air)"
- "Start Vite Server"
- "Run Go Tests" (`go test ./... -v -count=1` in `${workspaceFolder}/backend`)

- [ ] **Step 3: Verify JSON syntax and ensure no "uvicorn" strings remain**

Run:
```bash
python3 -m json.tool .vscode/tasks.json > /dev/null
grep -i "uvicorn" .vscode/tasks.json || echo "No uvicorn references found in .vscode/tasks.json"
```
Expected: Valid JSON and "No uvicorn references found in .vscode/tasks.json".

- [ ] **Step 4: Mark Task 3 as DONE in `docs/plans/task.md`**

Update `Task 3` status to `DONE` and `Task 4` status to `IN_PROGRESS`.

---

### Task 4: Integration Verification (Air & Tasks)

**Files:**
- Inspect: `backend/.air.toml`, `backend/.gitignore`, `.vscode/tasks.json`

**Interfaces:**
- Consumes: Configured `air` and Go backend
- Produces: Verified build, execution, and reload capability

- [ ] **Step 1: Test `air` build execution**

Verify `air` can read `backend/.air.toml` and compile `./tmp/arch-stats`:
Run:
```bash
cd backend
air -v
```
Expected: `air` runs and outputs its version.

Test running `air` briefly (e.g. for 3 seconds) using `timeout 3 air || true` or running the build command defined in `air.toml`:
```bash
cd backend
go build -o ./tmp/arch-stats ./cmd/arch-stats
./tmp/arch-stats --help || true
rm -rf ./tmp
```
Expected: Binary builds and outputs help/flags or starts cleanly.

- [ ] **Step 2: Verify `air` dry-run / execution in dev mode**

Run:
```bash
cd backend && timeout 3 air || true
```
Expected: Air starts, builds `cmd/arch-stats` to `tmp/arch-stats`, outputs colored banner/logs, and shuts down cleanly on SIGTERM/timeout with `clean_on_exit` deleting `tmp`.
Ensure any residual `backend/tmp` is cleaned:
```bash
rm -rf backend/tmp backend/build-errors.log
```

- [ ] **Step 3: Verify git status is clean of build artifacts**

Run:
```bash
git status --porcelain
```
Expected: Only modified `.vscode/tasks.json`, new `backend/.air.toml`, new `backend/.gitignore`, and documentation files are listed. No untracked `tmp/` or `build-errors.log`.

- [ ] **Step 4: Mark Task 4 as DONE in `docs/plans/task.md`**

Update `Task 4` status to `DONE` and `Task 5` status to `IN_PROGRESS`.

---

### Task 5: Mark Tasks as Done & Final Commit

**Files:**
- Modify: `docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified configuration and runtime behavior
- Produces: Completed task documents and committed git branch

- [ ] **Step 1: Update `docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md`**

Mark all acceptance criteria and all steps as completed `[x]`:

```markdown
## Acceptance Criteria

- [x] `backend/.air.toml` configures air with:
    - Watch `cmd/`, `internal/` directories for `.go` file changes
    - Build command: `go build -o ./tmp/arch-stats ./cmd/arch-stats`
    - Run command: `./tmp/arch-stats`
    - Exclude `tmp/`, `tests/`, `vendor/`, `specs/`
    - Log coloring enabled
- [x] `.vscode/tasks.json` is updated:
    - "Start Uvicorn Server" task replaced with "Start Go Server (air)"
    - "Start Vite Server" task unchanged
    - Docker Compose tasks unchanged
    - New task: "Run Go Tests" (`cd backend && go test ./... -v`)
- [x] Running `air` in `backend/` starts the Go server and automatically rebuilds on file changes.
- [x] `backend/.gitignore` updated to ignore `tmp/` (air build output).

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/.air.toml` |
| Modify | `.vscode/tasks.json` |
| Modify | `backend/.gitignore` (add `tmp/`) |

## Steps

- [x] **Step 1: Create `.air.toml`**
...
- [x] **Step 2: Update `backend/.gitignore`**
...
- [x] **Step 3: Update `.vscode/tasks.json`**
...
- [x] **Step 4: Test air**
...
- [x] **Step 5: Commit**
```

- [ ] **Step 2: Update `docs/plans/task.md` to reflect all tasks DONE**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Environment Initialization | DONE | Create branch `refactor/031-air-hot-reload-and-dev-environment` and install `air` |
| Task 2: Configure Air & Ignore Temporary Build Artifacts | DONE | Create `backend/.air.toml` and `backend/.gitignore` |
| Task 3: Update VS Code Tasks Configuration | DONE | Replace Uvicorn task with "Start Go Server (air)" in `.vscode/tasks.json` |
| Task 4: Integration Verification (Air & Tasks) | DONE | Test `air` live-reload build cycle and verify VS Code tasks syntax |
| Task 5: Mark Tasks as Done & Final Commit | DONE | Mark `031-air_hot_reload_and_dev_environment.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Commit changes to git**

Run:
```bash
git add backend/.air.toml backend/.gitignore .vscode/tasks.json docs/go_refactor/tasks/031-air_hot_reload_and_dev_environment.md docs/plans/task.md docs/plans/2026-09-05-air-hot-reload-and-dev-environment.md
git commit -m "chore: add air config for hot reload and update VS Code tasks for Go"
```

- [ ] **Step 4: Verify working tree is clean**

Run:
```bash
git status
```
Expected: `nothing to commit, working tree clean`.

---

## Self-Review Checklist

1. **Spec Coverage:**
   - `backend/.air.toml` configures watch dirs, build cmd, run cmd, exclude dirs, log coloring? Covered in Task 2.
   - `.vscode/tasks.json` replaced "Start Uvicorn Server" with "Start Go Server (air)"? Covered in Task 3.
   - "Start Vite Server" task unchanged? Covered in Task 3.
   - Docker Compose tasks unchanged? Covered in Task 3.
   - "Run Go Tests" task verified? Covered in Task 3.
   - `backend/.gitignore` updated to ignore `tmp/` and `build-errors.log`? Covered in Task 2.
   - `air` execution and rebuild behavior tested? Covered in Task 4.
   - Marking tasks as done at the end of implementation? Covered in Task 5.
2. **Placeholder Scan:**
   - No "TBD", "TODO", or unwritten code blocks exist in any step.
3. **Type Consistency:**
   - TOML keys, JSON keys, path names, and command flags are consistent across all steps and match Task 031 specification.
