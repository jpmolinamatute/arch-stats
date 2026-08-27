# Task 031: Configure `air` for Hot Reload + VS Code Tasks

## Git Branch

`refactor/031-air-hot-reload-and-dev-environment`

## Objective

Set up `air` for Go live-reload during development and update VS Code tasks to reference Go
tooling instead of Uvicorn. This completes the local development environment story for the
Go backend.

## Dependencies

- Task 025 (Go binary is a runnable HTTP server)
- Task 030 (Docker Compose updated for Go workflow)

## Acceptance Criteria

- [ ] `backend/.air.toml` configures air with:
    - Watch `cmd/`, `internal/` directories for `.go` file changes
    - Build command: `go build -o ./tmp/arch-stats ./cmd/arch-stats`
    - Run command: `./tmp/arch-stats`
    - Exclude `tmp/`, `tests/`, `vendor/`, `specs/`
    - Log coloring enabled
- [ ] `.vscode/tasks.json` is updated:
    - "Start Uvicorn Server" task replaced with "Start Go Server (air)"
    - "Start Vite Server" task unchanged
    - Docker Compose tasks unchanged
    - New task: "Run Go Tests" (`cd backend && go test ./... -v`)
- [ ] Running `air` in `backend/` starts the Go server and automatically rebuilds on file changes.
- [ ] `backend/.gitignore` updated to ignore `tmp/` (air build output).

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/.air.toml` |
| Modify | `.vscode/tasks.json` |
| Modify | `backend/.gitignore` (add `tmp/`) |

## Steps

- [ ] **Step 1: Create `.air.toml`**

  Create `backend/.air.toml`:

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

- [ ] **Step 2: Update `backend/.gitignore`**

  Add:

  ```txt
  tmp/
  build-errors.log
  ```

- [ ] **Step 3: Update `.vscode/tasks.json`**

  Replace the "Start Uvicorn Server" task with:

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
  }
  ```

  Add a "Run Go Tests" task:

  ```json
  {
      "label": "Run Go Tests",
      "type": "shell",
      "command": "go",
      "args": ["test", "./...", "-v", "-count=1"],
      "options": {
          "cwd": "${workspaceFolder}/backend"
      },
      "presentation": {
          "echo": true,
          "reveal": "always",
          "close": false,
          "panel": "shared"
      },
      "problemMatcher": ["$go"],
      "detail": "Run all Go backend tests"
  }
  ```

- [ ] **Step 4: Test air**

  ```bash
  cd backend
  go install github.com/air-verse/air@latest
  air
  ```

  Expected: air watches files, builds, and starts the server. Modifying a `.go` file triggers
  a rebuild.

- [ ] **Step 5: Commit**

  ```bash
  git add -A
  git commit -m "chore: add air config for hot reload and update VS Code tasks for Go"
  ```

## Verification

- `cd backend && air` — starts the Go server with live reload.
- `.vscode/tasks.json` contains "Start Go Server (air)" task (no Uvicorn reference).
- `cat backend/.air.toml` — valid TOML configuration.
- `grep -r "uvicorn" .vscode/tasks.json` — returns no results.
