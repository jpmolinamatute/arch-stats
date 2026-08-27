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

- [ ] The `backend-old/` directory is deleted from the repository.
- [ ] All references to `backend-old/` are removed from:
    - `.github/workflows/` — no workflow references backend-old
    - `scripts/` — no script references backend-old
    - `docker/` — no docker config references backend-old
    - `.agent/` — no skill references backend-old
    - `docs/` — task files can still reference it historically but no active config
    - `.gitmodules` — if the migrations submodule pointed into backend-old
- [ ] The `.github/actions/uv-setup/` custom action is deleted (no longer needed).
- [ ] The repository compiles, tests pass, and CI pipelines work without `backend-old/`.
- [ ] `git log --oneline -1` shows a clean commit message.

## Files to Delete

| Action | Path |
| ------ | ---- |
| Delete | `backend-old/` (entire directory) |
| Delete | `.github/actions/uv-setup/` (if it exists and is Python-only) |

## Steps

- [ ] **Step 1: Verify the Go backend is fully functional**

  ```bash
  cd backend
  go build ./cmd/arch-stats
  go test ./... -v -count=1
  go vet ./...
  ```

  All must pass before proceeding.

- [ ] **Step 2: Search for any remaining references**

  ```bash
  grep -rn "backend-old" . --include="*.yaml" --include="*.yml" --include="*.bash" \
    --include="*.json" --include="*.toml" --include="*.md" \
    | grep -v ".git/" | grep -v "docs/tasks/"
  ```

  Fix any active references found (docs/tasks/ references are historical and acceptable).

- [ ] **Step 3: Delete `backend-old/`**

  ```bash
  git rm -r backend-old/
  ```

- [ ] **Step 4: Delete `uv-setup` action if it exists**

  ```bash
  git rm -r .github/actions/uv-setup/ 2>/dev/null || true
  ```

- [ ] **Step 5: Verify the build still works**

  ```bash
  cd backend
  go build ./cmd/arch-stats
  go test ./... -v -count=1
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "chore: delete backend-old/ Python reference code — Go port complete"
  ```

## Verification

- `ls backend-old/` — "No such file or directory".
- `cd backend && go build ./cmd/arch-stats` — compiles.
- `cd backend && go test ./... -v -count=1` — all tests pass.
- `grep -rn "backend-old" . --include="*.yaml" --include="*.bash" --include="*.json" | grep -v ".git/" | grep -v "docs/tasks/"` — no active references.
