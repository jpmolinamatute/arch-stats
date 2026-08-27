# Task 035: Update `.agent/` Skills for Go Tooling

## Git Branch

`refactor/035-agent-skills-update`

## Objective

Update all backend-related `.agent/skills/` to reference Go tooling instead of Python. The
agent skills guide AI coding assistants on how to work with the codebase — they must reflect
the new Go toolchain.

## Dependencies

- Task 025 (Go backend is fully functional)
- Task 031 (air hot reload configured)
- Task 032 (CI workflow uses golangci-lint)

## Acceptance Criteria

- [ ] The following skills are updated to reference Go tooling:
    - `backend-formatting` → reference `gofmt` / `golangci-lint` instead of Ruff
    - `backend-linting` → reference `golangci-lint` instead of Ruff
    - `backend-package-management` → reference `go get` / `go mod tidy` instead of uv
    - `backend-run-server` → reference `air` instead of Uvicorn
    - `backend-scripts` → reference `go run` instead of `uv run`
    - `backend-tests` → reference `go test` instead of pytest
    - `backend-type-annotation` → reference `go vet` / static analysis instead of Ty
    - `backend-virtual-environment` → either remove (Go doesn't use venvs) or update to
    explain Go module system
- [ ] The `frontend-*` skills remain unchanged (frontend is still Vue 3 / TypeScript).
- [ ] `.agent/rules/instructions.md` is updated to reflect the Go backend:
    - Update "Big Picture" section
    - Update "Non-Negotiables" (pgx instead of asyncpg, Go structs instead of Pydantic, etc.)
    - Update "Repo Map" (new directory structure)
    - Update "Core Workflows" (Go commands instead of Python)
    - Update "Backend Patterns"
- [ ] Each updated skill file has accurate commands that work when executed.

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `.agent/skills/backend-formatting/SKILL.md` |
| Modify | `.agent/skills/backend-linting/SKILL.md` |
| Modify | `.agent/skills/backend-package-management/SKILL.md` |
| Modify | `.agent/skills/backend-run-server/SKILL.md` |
| Modify | `.agent/skills/backend-scripts/SKILL.md` |
| Modify | `.agent/skills/backend-tests/SKILL.md` |
| Modify | `.agent/skills/backend-type-annotation/SKILL.md` |
| Modify or Delete | `.agent/skills/backend-virtual-environment/SKILL.md` |
| Modify | `.agent/rules/instructions.md` |

## Steps

- [ ] **Step 1: Update `backend-formatting`**

  Replace Ruff formatting with Go formatting:
    - `gofmt -w .` for formatting
    - `golangci-lint run --fix` for auto-fixable lint issues

- [ ] **Step 2: Update `backend-linting`**

  Replace Ruff linting with:
    - `golangci-lint run ./...` for linting
    - `go vet ./...` for vet checks

- [ ] **Step 3: Update `backend-package-management`**

  Replace uv with Go module commands:
    - `go get <package>` to add dependencies
    - `go mod tidy` to clean up
    - `go mod download` to download dependencies

- [ ] **Step 4: Update `backend-run-server`**

  Replace Uvicorn with:
    - `air` for dev mode (hot reload)
    - `go run ./cmd/arch-stats` for one-shot run
    - `go build ./cmd/arch-stats && ./arch-stats` for production

- [ ] **Step 5: Update `backend-scripts`**

  Replace `uv run` with `go run`.

- [ ] **Step 6: Update `backend-tests`**

  Replace pytest with:
    - `go test ./... -v` for all tests
    - `go test ./internal/service/... -v` for specific package
    - `go test -run TestName ./...` for specific test
    - `go test -race ./...` for race detection

- [ ] **Step 7: Update `backend-type-annotation`**

  Replace Ty with:
    - `go vet ./...` for static analysis
    - `golangci-lint run` includes type checking via staticcheck

- [ ] **Step 8: Handle `backend-virtual-environment`**

  Either delete (Go doesn't use venvs) or rewrite to explain Go modules:
    - `go mod init` creates a module
    - `go mod tidy` manages dependencies
    - No activation step needed

- [ ] **Step 9: Update `.agent/rules/instructions.md`**

  Update all sections to reflect Go backend:
    - Tech stack: Go 1.27.0, chi, pgx, squirrel
    - DB access: pgx only (no ORM)
    - Formatting: golangci-lint, gofmt
    - Repo map: new `backend/cmd/`, `backend/internal/` structure
    - Core workflows: `air`, `go test`, `docker compose`
    - Backend patterns: repository pattern, constructor injection, error wrapping

- [ ] **Step 10: Commit**

  ```bash
  git add -A
  git commit -m "chore: update agent skills and rules for Go backend tooling"
  ```

## Verification

- `grep -rn "ruff\|pytest\|uvicorn\|uv run\|asyncpg\|pydantic" .agent/skills/backend-*` — no
  Python references in backend skills.
- `grep -rn "ruff\|pytest\|uvicorn\|asyncpg" .agent/rules/instructions.md` — no Python
  references in instructions.
- Each skill's commands are valid Go commands that work when executed.
