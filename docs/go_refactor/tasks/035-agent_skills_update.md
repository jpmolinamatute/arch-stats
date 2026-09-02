# Task 035: Update `.agent/` Skills for Go Tooling

## Git Branch

`refactor/035-agent-skills-update`

## Objective

Update backend-related `.agent/skills/` to reference Go tooling instead of Python, consolidate
Python-era formatting and type-checking skills into a unified `backend-linting` skill, and
introduce a `backend-go-coding` skill enforcing Effective Go best practices. The agent skills
guide AI coding assistants on how to work with the codebase — they must reflect the new Go
toolchain.

## Dependencies

- Task 032 (CI workflow uses golangci-lint)

## Acceptance Criteria

- [x] A new local skill `backend-go-coding` is created:
    - Enforces [Effective Go](https://go.dev/doc/effective_go) best practices (formatting, package &
      variable naming, control flow, functions/multi-return, error handling patterns, struct
      embedding, interfaces, concurrency).
    - Mandates calling the `backend-linting` skill after writing or modifying Go code.
    - Mandates calling the `backend-tests` skill after linting.
- [x] `backend-linting` consolidates formatting, linting, and type checking:
    - References `golangci-lint run ./...` and `golangci-lint run --fix` (with gofumpt, govet,
      staticcheck).
- [x] Redundant Python-era skills are removed:
    - `backend-formatting` is deleted (absorbed into `backend-linting`).
    - `backend-type-annotation` is deleted (absorbed into `backend-linting` / staticcheck / compiler).
    - `backend-virtual-environment` is deleted (Go uses modules, no virtual environments).
- [x] The remaining backend skills are updated for Go tooling:
    - `backend-package-management` → reference `go get` / `go mod tidy` instead of uv.
    - `backend-run-server` → reference `air` instead of Uvicorn.
    - `backend-scripts` → reference `go run` instead of `uv run`.
    - `backend-tests` → reference `go test` instead of pytest.
- [x] The `frontend-*` skills remain unchanged (frontend is still Vue 3 / TypeScript).
- [x] `.agent/rules/instructions.md` is updated to reflect the Go backend:
    - Update "Big Picture" section.
    - Update "Non-Negotiables" (pgx instead of asyncpg, Go structs instead of Pydantic, etc.).
    - Update "Repo Map" (new directory structure).
    - Update "Core Workflows" (Go commands instead of Python).
    - Update "Backend Patterns".
- [x] Each created/updated skill file has accurate commands and instructions that work when executed.

## Files to Modify

| Action | Path |
| ------ | ---- |
| Create | `.agent/skills/backend-go-coding/SKILL.md` |
| Modify | `.agent/skills/backend-linting/SKILL.md` |
| Modify | `.agent/skills/backend-package-management/SKILL.md` |
| Modify | `.agent/skills/backend-run-server/SKILL.md` |
| Modify | `.agent/skills/backend-scripts/SKILL.md` |
| Modify | `.agent/skills/backend-tests/SKILL.md` |
| Delete | `.agent/skills/backend-formatting/SKILL.md` |
| Delete | `.agent/skills/backend-type-annotation/SKILL.md` |
| Delete | `.agent/skills/backend-virtual-environment/SKILL.md` |
| Modify | `.agent/rules/instructions.md` |

## Steps

- [x] **Step 1: Create `backend-go-coding` skill**

  Create `.agent/skills/backend-go-coding/SKILL.md` to guide agents writing Go code in `backend/`:
    - Emphasize idioms and standards from [Effective Go](https://go.dev/doc/effective_go):
        - Formatting (`gofmt` / `gofumpt` conventions)
        - Names (package names, Getters, interface names, MixedCaps)
        - Control structures (`if`, `for`, `switch`, type switch, short variable declarations)
        - Functions (multiple return values, named result parameters, `defer`)
        - Data (allocation with `make` vs `new`, constructors and composite literals)
        - Methods and interfaces (pointer vs value receivers, interface satisfaction, type assertions)
        - Error handling (errors as values, wrapping, sentinel errors, avoiding unnecessary `panic`)
        - Concurrency (goroutines, channels, sync primitives)
    - Mandate execution flow:
        1. Follow Effective Go idioms when writing or modifying Go code.
        2. Call the `backend-linting` skill to verify formatting and linting.
        3. Call the `backend-tests` skill to ensure tests pass.

- [x] **Step 2: Update `backend-linting`**

  Consolidate linting, formatting, and type-checking into `.agent/skills/backend-linting/SKILL.md`:
    - `golangci-lint run ./...` for complete linting and static checks (govet, staticcheck, errcheck,
      revive, etc.)
    - `golangci-lint run --fix ./...` or `gofumpt -l -w .` for formatting and autofixes
    - Run from `backend/` directory

- [x] **Step 3: Remove Redundant Python Skills**

  Delete obsolete skill directories:
    - `.agent/skills/backend-formatting/`
    - `.agent/skills/backend-type-annotation/`
    - `.agent/skills/backend-virtual-environment/`

- [x] **Step 4: Update `backend-package-management`**

  Replace uv with Go module commands:
    - `go get <package>` to add dependencies
    - `go mod tidy` to clean up `go.mod` and `go.sum`
    - `go mod download` to download dependencies

- [x] **Step 5: Update `backend-run-server`**

  Replace Uvicorn with:
    - `air` for dev mode (hot reload)
    - `go run ./cmd/arch-stats` for one-shot run
    - `go build ./cmd/arch-stats && ./arch-stats` for production

- [x] **Step 6: Update `backend-scripts`**

  Replace `uv run` with `go run`.

- [x] **Step 7: Update `backend-tests`**

  Replace pytest with:
    - `go test ./... -v` for all tests
    - `go test ./internal/service/... -v` for a specific package
    - `go test -run TestName ./...` for a specific test
    - `go test -race ./...` for race detection

- [x] **Step 8: Update `.agent/rules/instructions.md`**

  Update all sections to reflect Go backend:
    - Tech stack: Go 1.24+, chi, pgx, squirrel
    - DB access: pgx only (no ORM)
    - Formatting & Linting: golangci-lint, gofumpt
    - Repo map: `backend/cmd/`, `backend/internal/` structure
    - Core workflows: `air`, `go test`, `docker compose`
    - Backend patterns: repository pattern, constructor injection, error wrapping, reference
      `backend-go-coding`

- [x] **Step 9: Commit**

  ```bash
  git add -A
  git commit -m "chore: update agent skills and rules for Go backend tooling"
  ```

## Verification

- `grep -rn "ruff\|pytest\|uvicorn\|uv run\|asyncpg\|pydantic" .agent/skills/backend-*` — no
  Python references in backend skills.
- `grep -rn "ruff\|pytest\|uvicorn\|asyncpg" .agent/rules/instructions.md` — no Python
  references in instructions.
- Skills `backend-formatting`, `backend-type-annotation`, and `backend-virtual-environment` no
  longer exist.
- New skill `backend-go-coding` exists and links to `backend-linting` and `backend-tests`.
- Each skill's commands are valid Go commands that work when executed.
