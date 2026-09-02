# Agent Skills Update for Go Tooling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update all backend `.agent/skills/` and `.agent/rules/` to reference the Go toolchain and Effective Go idioms, consolidating formatting and static analysis into `backend-linting`, removing obsolete Python skills, and adding a `backend-go-coding` skill.

**Architecture:** Create `.agent/skills/backend-go-coding/SKILL.md` enforcing Effective Go guidelines and chaining into linting and testing. Consolidate formatting, type-checking, and static analysis into `.agent/skills/backend-linting/SKILL.md` backed by `golangci-lint` and `gofumpt`. Delete obsolete Python skills (`backend-formatting`, `backend-type-annotation`, `backend-virtual-environment`). Modernize `backend-package-management`, `backend-run-server`, `backend-scripts`, and `backend-tests` with Go commands. Update `.agent/rules/instructions.md` and `.agent/rules/new-backend-code.md` to reflect the Go backend architecture and workflows.

**Tech Stack:** Go 1.27+, `golangci-lint`, `gofumpt`, `go test`, `air`, `pgx/v5`, Chi router, Antigravity agent skills/rules format.

**Spec:** [docs/go_refactor/tasks/035-agent_skills_update.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/035-agent_skills_update.md)

## Global Constraints

- Git branch: `refactor/035-agent-skills-update`
- Frontend skills (`frontend-*`) must remain completely untouched (Vue 3 / TypeScript).
- No Python toolchain references (`ruff`, `pytest`, `uvicorn`, `uv run`, `asyncpg`, `pydantic`, `venv`) in backend skills or rules.
- Consolidated linter skill must reference `golangci-lint run ./...` and `golangci-lint run --fix ./...` (or `gofumpt -l -w .`), runnable from `backend/` or via `./scripts/linting.bash --go`.
- `backend-go-coding` must enforce Effective Go standards and mandate a 3-step workflow: code -> lint -> test.
- Deleted skill directories must be completely removed: `backend-formatting`, `backend-type-annotation`, `backend-virtual-environment`.
- All commands in skill markdown files must be syntactically valid and executable in this repository.

---

## File Structure

```
.agent/
├── rules/
│   ├── instructions.md                      # [MODIFY] Rewrite backend sections for Go architecture
│   ├── new-backend-code.md                  # [MODIFY] Update trigger glob to Go and skills to Go toolchain
│   └── new-frontend-code.md                 # (unchanged)
└── skills/
    ├── backend-formatting/                  # [DELETE] Absorbed into backend-linting
    │   └── SKILL.md
    ├── backend-go-coding/                   # [NEW] Effective Go idioms + lint/test chain
    │   └── SKILL.md
    ├── backend-linting/                     # [MODIFY] Consolidated Go linter, formatter, type checker
    │   └── SKILL.md
    ├── backend-package-management/          # [MODIFY] Go modules (go get, go mod tidy)
    │   └── SKILL.md
    ├── backend-run-server/                  # [MODIFY] air, go run, binary execution
    │   └── SKILL.md
    ├── backend-scripts/                     # [MODIFY] go run scripts
    │   └── SKILL.md
    ├── backend-tests/                       # [MODIFY] go test flags, race detector, linting.bash --go
    │   └── SKILL.md
    ├── backend-type-annotation/             # [DELETE] Absorbed into Go compiler / staticcheck / linting
    │   └── SKILL.md
    └── backend-virtual-environment/         # [DELETE] Obsolete (Go uses go.mod modules)
        └── SKILL.md
```

---

### Task 1: Create `backend-go-coding` Skill

**Files:**
- Create: `.agent/skills/backend-go-coding/SKILL.md`

**Interfaces:**
- Consumes: [Effective Go](https://go.dev/doc/effective_go) patterns, `backend-linting` skill, `backend-tests` skill
- Produces: Definitive guideline for agents writing or modifying Go code in `backend/`

- [ ] **Step 1: Create `.agent/skills/backend-go-coding/SKILL.md`**

Write `.agent/skills/backend-go-coding/SKILL.md` with YAML frontmatter, Effective Go idioms, and mandatory lint/test workflow:

```markdown
---
name: backend-go-coding
description: Standards, idioms, and best practices for writing Go code in the backend following Effective Go
---

# Go Coding Standards & Best Practices

All Go code written in `backend/` must adhere to idiomatic Go as defined by [Effective Go](https://go.dev/doc/effective_go) and the project architecture standards.

## Mandatory Execution Workflow

Whenever creating or modifying Go code, follow this sequence:

1. **Write idiomatic Go** adhering to the conventions below.
2. **Run linting and formatting** using the `backend-linting` skill:
   ```bash
   cd backend && golangci-lint run ./...
   ```
3. **Run tests** using the `backend-tests` skill:
   ```bash
   cd backend && go test ./... -v
   ```

Do not consider Go code complete until both linting and testing pass with zero errors.

## Effective Go Conventions

### 1. Formatting & Style
- Never argue about formatting: all code must be formatted using `gofumpt` (enforced by `golangci-lint`).
- Use tabs for indentation, spaces for alignment.
- Keep line lengths reasonable, but do not break lines artificially.

### 2. Naming
- **Package names**: Short, clear, lowercase, single-word names (e.g., `config`, `service`, `model`, `repository`, `handler`). Do not use underscores or mixedCaps.
- **Getters**: Omit the `Get` prefix. For field `owner`, name the method `Owner()`, not `GetOwner()`. Use `SetOwner(...)` for setters.
- **Interface names**: One-method interfaces end in `-er` (e.g., `Reader`, `Writer`, `Closer`). Keep interfaces small and define them where they are consumed.
- **MixedCaps**: Use camelCase or PascalCase (`MixedCaps`), never snake_case for Go identifiers. Acronyms stay uppercase (`URL`, `HTTP`, `ID`, `JWT`, `DSN`).

### 3. Control Flow
- Avoid unnecessary indentation: handle errors and guard clauses early with fast return.
  ```go
  // Good:
  if err != nil {
      return nil, err
  }
  // proceed with happy path
  ```
- Use short variable declarations in `if` statements when scoping variables:
  ```go
  if err := cfg.Validate(); err != nil {
      return err
  }
  ```
- Use type switches to inspect concrete types from interface values.

### 4. Functions & Returns
- Return multiple values instead of out-parameters or synthetic wrapper structs.
- The standard error return is always the last return value (`(Result, error)`).
- Use `defer` immediately after resource allocation to guarantee cleanup (e.g., rows closing, unlock):
  ```go
  rows, err := pool.Query(ctx, query, args...)
  if err != nil {
      return nil, err
  }
  defer rows.Close()
  ```

### 5. Data & Allocation
- Use composite literals (`&Config{DevMode: true}`) instead of `new()`.
- Use `make` only for slices, maps, and channels where capacity or length is needed.
- Provide constructor functions (`New...`) when struct initialization requires validation, defaults, or dependency injection.

### 6. Methods & Interfaces
- Choose pointer receivers when the method modifies the struct or the struct is large; use value receivers for small immutable value objects.
- Keep receivers consistent across a type's method set.
- Satisfy interfaces implicitly—do not declare explicit interface implementations.
- Accept interfaces, return concrete structs (e.g., accept `repository.ArcherRepository` interface in service constructors, return `*Service`).

### 7. Error Handling
- Errors are values: inspect and handle them explicitly.
- Use `fmt.Errorf("...: %w", err)` to wrap lower-level errors with context.
- Use sentinel errors defined in `internal/apperror` (e.g., `apperror.ErrNotFound`, `apperror.ErrUnauthorized`).
- Check wrapped errors using `errors.Is(err, apperror.ErrNotFound)` or `errors.As`.
- **Never panic** in library, repository, service, or handler code. Reserve `panic` strictly for truly unrecoverable program initialization failures.

### 8. Concurrency
- "Do not communicate by sharing memory; instead, share memory by communicating."
- Always propagate `context.Context` as the first argument in blocking, I/O, or database calls.
- Always ensure goroutines have a well-defined termination condition to prevent goroutine leaks.
- Use `sync.WaitGroup` or `errgroup.Group` for coordinated concurrency.
```

- [ ] **Step 2: Verify skill formatting and link syntax**

Run:
```bash
test -f .agent/skills/backend-go-coding/SKILL.md && head -n 10 .agent/skills/backend-go-coding/SKILL.md
```
Expected: File exists and displays valid frontmatter.

- [ ] **Step 3: Commit**

```bash
git add .agent/skills/backend-go-coding/SKILL.md
git commit -m "feat: create backend-go-coding skill for Effective Go standards"
```

---

### Task 2: Update and Consolidate `backend-linting` Skill

**Files:**
- Modify: `.agent/skills/backend-linting/SKILL.md`

**Interfaces:**
- Consumes: `golangci-lint`, `gofumpt`, `scripts/linting.bash --go`
- Produces: Consolidated skill covering Go formatting, linting, and static analysis

- [ ] **Step 1: Replace `.agent/skills/backend-linting/SKILL.md`**

Replace `.agent/skills/backend-linting/SKILL.md` with:

```markdown
---
name: backend-linting
description: How to run golangci-lint and gofumpt to lint, format, and check Go backend code
---

# Go Backend Linting & Formatting

We use `golangci-lint` (which includes `govet`, `staticcheck`, `errcheck`, `gofumpt`, `revive`, and `ineffassign`) to lint, format, and statically analyze all Go code under `backend/`. Configuration is defined in `backend/.golangci.yml`.

## How to Run

### 1. Fast Linting (From `backend/`)

Run from the `backend/` directory:

```bash
cd backend
golangci-lint run ./...
```

### 2. Auto-fix Formatting and Lint Issues

To automatically fix format errors and auto-fixable lint issues:

```bash
cd backend
golangci-lint run --fix ./...
# or format directly with gofumpt:
gofumpt -l -w .
```

### 3. Full Project Go Verification Script (From project root)

Runs formatter (`gofumpt`), linter (`golangci-lint`), and unit tests (`go test`):

```bash
./scripts/linting.bash --go
```

## IDE Integration

Antigravity IDE / VS Code automatically formats on save using `gofumpt` and displays inline lint diagnostics via `golangci-lint`.
```

- [ ] **Step 2: Test the commands in the skill**

Run:
```bash
cd backend && golangci-lint run ./...
```
Expected: 0 issues reported.

Run:
```bash
./scripts/linting.bash --go
```
Expected: PASS for formatter, linter, and tests.

- [ ] **Step 3: Commit**

```bash
git add .agent/skills/backend-linting/SKILL.md
git commit -m "feat: consolidate formatting and linting into backend-linting skill"
```

---

### Task 3: Remove Redundant Python Skills

**Files:**
- Delete: `.agent/skills/backend-formatting/SKILL.md`
- Delete: `.agent/skills/backend-formatting/`
- Delete: `.agent/skills/backend-type-annotation/SKILL.md`
- Delete: `.agent/skills/backend-type-annotation/`
- Delete: `.agent/skills/backend-virtual-environment/SKILL.md`
- Delete: `.agent/skills/backend-virtual-environment/`

**Interfaces:**
- Consumes: Consolidated `backend-linting`
- Produces: Removal of obsolete Python skill directories

- [ ] **Step 1: Delete obsolete directories using git rm**

Run:
```bash
git rm -r .agent/skills/backend-formatting
git rm -r .agent/skills/backend-type-annotation
git rm -r .agent/skills/backend-virtual-environment
```

- [ ] **Step 2: Verify deletion**

Run:
```bash
ls -d .agent/skills/backend-formatting .agent/skills/backend-type-annotation .agent/skills/backend-virtual-environment 2>/dev/null || true
```
Expected: No such file or directory for all three.

- [ ] **Step 3: Commit**

```bash
git commit -m "chore: remove obsolete Python skills (formatting, type-annotation, venv)"
```

---

### Task 4: Update Remaining Backend Skills (`package-management`, `run-server`, `scripts`, `tests`)

**Files:**
- Modify: `.agent/skills/backend-package-management/SKILL.md`
- Modify: `.agent/skills/backend-run-server/SKILL.md`
- Modify: `.agent/skills/backend-scripts/SKILL.md`
- Modify: `.agent/skills/backend-tests/SKILL.md`

**Interfaces:**
- Consumes: Go module commands, `air` dev server, `go run`, `go test`
- Produces: Accurate, runnable Go guidance across all backend development skills

- [ ] **Step 1: Update `.agent/skills/backend-package-management/SKILL.md`**

Replace contents with:

```markdown
---
name: backend-package-management
description: How to manage Go module dependencies using go get, go mod tidy, and go mod download
---

# Go Package & Dependency Management

We use Go modules (`go.mod` and `go.sum` in `backend/`) to manage backend dependencies.

All commands should be executed from the `backend/` directory:

```bash
cd backend
```

## Add or Update Dependencies

```bash
# Add or upgrade a dependency
go get github.com/example/pkg@latest

# Add a specific version
go get github.com/example/pkg@v1.2.3
```

## Clean and Prune Dependencies

Always run `go mod tidy` after adding or removing package imports to remove unused modules and update checksums in `go.sum`:

```bash
go mod tidy
```

## Download Cached Dependencies

```bash
go mod download
```

## Verify Dependencies

```bash
go mod verify
```
```

- [ ] **Step 2: Update `.agent/skills/backend-run-server/SKILL.md`**

Replace contents with:

```markdown
---
name: backend-run-server
description: How to start the Go backend server in development mode (with air live-reload) or production mode
---

# Run Backend Server

This skill describes how to run the Go backend server in local development or production mode.

## Prerequisites

1. **Docker Infrastructure:** PostgreSQL 17 container must be running:
   ```bash
   docker compose -f docker/docker-compose.yaml up -d
   ```
2. **Environment Variables:** Verify `backend/.env` exists or required environment variables are set (see `internal/config/config.go`).

## Execution Options

### Option 1: Live-Reload Dev Mode with `air` (Recommended for development)

From the `backend/` directory:

```bash
cd backend
air
```

Or trigger the VS Code task: `"Start Go Server (air)"`.

`air` monitors `.go` and `.toml` files in `backend/` and automatically recompiles and restarts the server on change.

### Option 2: One-Shot Run (Direct execution)

To run without live reload:

```bash
cd backend
go run ./cmd/arch-stats
```

### Option 3: Compile and Run Binary (Production-like)

```bash
cd backend
go build -o arch-stats ./cmd/arch-stats
./arch-stats
```

## Verification

- Check stdout for the startup banner: `arch-stats starting (dev_mode=true)`
- Verify database connection pool initialization succeeds.
- Confirm server is listening on the configured port (default `:8080`).
```

- [ ] **Step 3: Update `.agent/skills/backend-scripts/SKILL.md`**

Replace contents with:

```markdown
---
name: backend-scripts
description: How to run Go backend scripts and standalone utilities using go run
---

# Running Go Backend Scripts

To run standalone Go utilities or main programs in the `backend/` module:

```bash
cd backend
go run ./cmd/<utility_name> [args...]
```

For single-file scripts or migration tools:

```bash
cd backend
go run ./path/to/script.go
```
```

- [ ] **Step 4: Update `.agent/skills/backend-tests/SKILL.md`**

Replace contents with:

```markdown
---
name: backend-tests
description: How to run Go backend unit and integration tests using go test
---

# Go Backend Tests

We use Go's standard testing toolchain (`go test`) for unit and integration testing in `backend/`.

## Running Tests

All commands are executed from the `backend/` directory:

```bash
cd backend
```

### 1. Run All Tests

```bash
go test ./... -v
```

### 2. Run with Race Detection (Recommended before commit)

```bash
go test -race ./... -v
```

### 3. Run Tests for a Specific Package

```bash
go test ./internal/config/... -v
go test ./internal/apperror/... -v
go test ./internal/repository/... -v
```

### 4. Run a Specific Test Function

```bash
go test ./internal/config/... -run TestNewLogger_DevMode -v
```

### 5. Bypass Test Cache

```bash
go test ./... -count=1 -v
```

### 6. Full Suite via Root Script

From the project root:

```bash
./scripts/linting.bash --go
```
```

- [ ] **Step 5: Verify all updated skills are valid**

Run:
```bash
grep -rn "ruff\|pytest\|uvicorn\|uv run\|backend-old" .agent/skills/backend-*
```
Expected: No output.

- [ ] **Step 6: Commit**

```bash
git add .agent/skills/backend-package-management/SKILL.md \
        .agent/skills/backend-run-server/SKILL.md \
        .agent/skills/backend-scripts/SKILL.md \
        .agent/skills/backend-tests/SKILL.md
git commit -m "feat: update backend package-management, run-server, scripts, and tests skills for Go"
```

---

### Task 5: Update `.agent/rules/instructions.md` and `.agent/rules/new-backend-code.md`

**Files:**
- Modify: `.agent/rules/instructions.md`
- Modify: `.agent/rules/new-backend-code.md`

**Interfaces:**
- Consumes: Current Go backend architecture (`backend/cmd/`, `backend/internal/`, `backend/migrations/`)
- Produces: Up-to-date agent rules reflecting Go backend patterns, workflows, and skills

- [ ] **Step 1: Update `.agent/rules/instructions.md`**

Replace `.agent/rules/instructions.md` with:

```markdown
---
trigger: model_decision
description: Project Overview
---

# Arch-Stats – AI Coding Agent Guide

Audience: AI coding agents
Repo: <https://github.com/jpmolinamatute/arch-stats>
Last updated: 2026-09-02

## Big Picture

- Two surfaces: Backend (Go 1.27+, Chi router, pgx/v5, PostgreSQL 17) and Frontend (Vue 3 + Vite SPA). FE build
  outputs go into `backend/internal/handler/static/dist/` (or embedded via `//go:embed`) and are served by Chi.
- Data flow (shot lifecycle): sensor/bot inserts rows → Postgres NOTIFY → WebSocket hub (`internal/websocket`)
  → frontend renders live updates. Keep flow unidirectional and components loosely coupled.

## Non‑Negotiables

- DB access: `pgx/v5` only with `pgxpool.Pool` (no ORM, no sync DB). Queries use parameterized SQL ($1, $2) or Squirrel query builder.
- Go structs with JSON tags for requests/responses; validate input explicitly.
- Strict error handling: wrap errors with `%w` or `apperror.Wrap`; use sentinel errors in `internal/apperror`. Never swallow errors or use `# type: ignore`-style escapes.
- Formatting & Linting: Go (`golangci-lint`, `gofumpt`), JS/TS (`ESLint`, `Prettier`), Bash (`shellcheck`, `shfmt`). Keep diffs minimal.
- FE/BE separation: FE consumes generated TypeScript types; do not import backend internals into FE.
- Follow Effective Go conventions; reference the `backend-go-coding` skill.

## Repo Map (targets to read first)

- Backend app entrypoint: `backend/cmd/arch-stats/main.go`.
- Configuration & logging: `backend/internal/config/`.
- Domain models: `backend/internal/model/`.
- Database access: `backend/internal/repository/`.
- Business logic: `backend/internal/service/`.
- HTTP handlers & routing: `backend/internal/handler/`.
- HTTP middleware: `backend/internal/middleware/`.
- WebSocket real-time: `backend/internal/websocket/`.
- Application errors: `backend/internal/apperror/`.
- DB migrations: `backend/migrations/*.sql` executed via Goose (`goose postgres "<DSN>" up`).
- Frontend app: `frontend/src/App.vue`, `frontend/src/composables/`, `frontend/src/api/`, and generated
  types in `frontend/src/types/`.
- Scripts: `scripts/linting.bash`, `scripts/generate_fe_types.bash`, etc.

## Core Workflows

- Start services (Docker + API + FE):
  - Start Docker Compose:
    ```bash
    docker compose -f docker/docker-compose.yaml up -d
    ```
  - Start Go Server (dev live-reload):
    ```bash
    cd backend && air
    ```
  - Start Frontend Vite Server:
    ```bash
    cd frontend && npm install && npm run dev
    ```

- Run Go tests:
  ```bash
  cd backend && go test -race ./... -v
  ```

- Run Go linting & formatting:
  ```bash
  cd backend && golangci-lint run ./...
  # or full suite:
  ./scripts/linting.bash --go
  ```

- Generate FE API types when backend models change:
  ```bash
  cd frontend && npm run generate:types
  ```

## Backend Patterns

- Handlers: Accept `http.ResponseWriter` and `*http.Request`; parse JSON; call service layer; serialize JSON responses using standard helpers.
- Service Layer: Coordinates business logic, interacts with repositories, returns domain models and `apperror` sentinel errors.
- Repositories: Accept `context.Context` as first parameter; use `*pgxpool.Pool` or `pgx.Tx`; return domain models from `internal/model/`.
- Error handling:
  ```go
  if err != nil {
      return nil, fmt.Errorf("failed to fetch session: %w", err)
  }
  ```
- Dependency Injection: Constructors (`New...`) take explicit interfaces or dependencies; no global mutable state.

## Frontend Patterns

- Composition API with `<script setup>` only; state modules in `state/`, composables in `composables/`.
- Views mapped via `componentsMap` in `App.vue`.
- Always consume generated OpenAPI types from `@/types/types.generated`.
- Keep reactive sources narrow; derive with `computed()`.

## Scripts & Tasks

- Canonical bash style in `scripts/` with `set -euo pipefail`.
- Use `./scripts/linting.bash --go` to verify Go changes before committing.

## References

- Go Coding Standards: `.agent/skills/backend-go-coding/SKILL.md`
- Linting: `.agent/skills/backend-linting/SKILL.md`
- Tests: `.agent/skills/backend-tests/SKILL.md`
- Frontend rules: `.agent/rules/google_standards_typescript.md`
```

- [ ] **Step 2: Update `.agent/rules/new-backend-code.md`**

Update `.agent/rules/new-backend-code.md` to target `backend/**/*.go` and reference the new Go skills:

```markdown
---
trigger: glob
globs: backend/**/*.go
---

# New Go Backend Code

## Coding Standards

When creating or modifying Go backend code, follow these rules:

* Adhere to [Effective Go](https://go.dev/doc/effective_go) standards. Use the `backend-go-coding` skill.
* Code must pass linting and formatting without errors. Use the `backend-linting` skill (`cd backend && golangci-lint run ./...`).
* Always write unit tests for new code (covering both success and error paths). All tests must pass. Use the `backend-tests` skill (`cd backend && go test -race ./... -v`).
* Use `internal/apperror` sentinel errors and wrap errors using `%w` or `apperror.Wrap`.
* Do not introduce external dependencies without checking `backend-package-management`.

## Workflow Chain

Whenever writing Go code:
1. Implement code according to `backend-go-coding`.
2. Run `backend-linting` to format and check.
3. Run `backend-tests` to ensure tests pass.

## Feedback

* Explain what was done and why it was done in all agent responses.
```

- [ ] **Step 3: Verify rules syntax and references**

Run:
```bash
grep -rn "ruff\|pytest\|uvicorn\|asyncpg\|pydantic" .agent/rules/instructions.md .agent/rules/new-backend-code.md
```
Expected: No output.

- [ ] **Step 4: Commit**

```bash
git add .agent/rules/instructions.md .agent/rules/new-backend-code.md
git commit -m "docs: update agent instructions and new-backend-code rules for Go"
```

---

### Task 6: Verification, Acceptance Checklist & Task Status Update

**Files:**
- Modify: `docs/go_refactor/tasks/035-agent_skills_update.md`

**Interfaces:**
- Consumes: All updated and created files
- Produces: Full validation pass and updated task checklist

- [ ] **Step 1: Check for obsolete Python terms in skills**

Run:
```bash
grep -rn "ruff\|pytest\|uvicorn\|uv run\|asyncpg\|pydantic" .agent/skills/backend-*
```
Expected: No matches found.

- [ ] **Step 2: Check for obsolete Python terms in rules**

Run:
```bash
grep -rn "ruff\|pytest\|uvicorn\|asyncpg" .agent/rules/
```
Expected: No matches found.

- [ ] **Step 3: Confirm deleted skills are removed**

Run:
```bash
test ! -d .agent/skills/backend-formatting && \
test ! -d .agent/skills/backend-type-annotation && \
test ! -d .agent/skills/backend-virtual-environment && \
echo "All obsolete skills successfully deleted"
```
Expected: "All obsolete skills successfully deleted".

- [ ] **Step 4: Confirm frontend skills remain untouched**

Run:
```bash
git status .agent/skills/frontend-*
```
Expected: Nothing modified.

- [ ] **Step 5: Run full Go checks**

Run:
```bash
./scripts/linting.bash --go
```
Expected: Success exit code 0.

- [ ] **Step 6: Update `docs/go_refactor/tasks/035-agent_skills_update.md`**

Check off all acceptance criteria and steps in `docs/go_refactor/tasks/035-agent_skills_update.md`:
- [x] A new local skill `backend-go-coding` is created
- [x] `backend-linting` consolidates formatting, linting, and type checking
- [x] Redundant Python-era skills are removed
- [x] The remaining backend skills are updated for Go tooling
- [x] The `frontend-*` skills remain unchanged
- [x] `.agent/rules/instructions.md` is updated to reflect the Go backend
- [x] Each created/updated skill file has accurate commands and instructions that work when executed

- [ ] **Step 7: Final commit**

```bash
git add docs/go_refactor/tasks/035-agent_skills_update.md
git commit -m "docs: complete task 035 agent skills update checklist"
```
