# Task 011: End-to-End Verification and Linting Compliance

## Git Branch

`feature/011-e2e-verification-and-linting`

## Objective

Execute exhaustive full-stack verification across database migrations, backend Go services and
handlers, and frontend Vue 3 components and composables. Confirm that all obsolete subsystems
(`slot`, `target`, `websocket`, `open_participants`, and legacy monolithic `/auth/register`) have
been completely purged from the codebase. Ensure zero linting, formatting, or type checking
errors across the entire repository.

## Dependencies

- Tasks 001 through 010.

## Acceptance Criteria

- [ ] Database verification:
    - [ ] `./backend/migrations/scripts/run_migration_tests.bash` succeeds with 0 failures
    - [ ] Migrations roll down and up cleanly without SQL errors
- [ ] Backend code quality and tests:
    - [ ] `gofumpt -l -w backend/` produces no diffs
    - [ ] `cd backend && golangci-lint run ./...` passes with 0 issues
    - [ ] `cd backend && go test -race ./...` passes with 0 failures
- [ ] Frontend code quality and tests:
    - [ ] `cd frontend && npm run type-check` passes with 0 TypeScript errors
    - [ ] `cd frontend && npm run lint` passes with 0 ESLint errors
    - [ ] `cd frontend && npm run test` passes with 0 Vitest test failures
- [ ] Obsolete subsystem purge verification:
    - [ ] No occurrences of `slot_id` in active backend routes or frontend stores
    - [ ] No occurrences of `websocket` hub, client, or routes in active application files
    - [ ] No remaining references to dropped functions (`get_next_lane`,
          `get_available_targets`, etc.)
    - [ ] No occurrences of `/auth/register` or `registerNewArcher` in active application files
- [ ] Documentation and Markdownlint compliance:
    - [ ] All task files in `docs/simplify_refactor/tasks/` pass markdownlint against
          `.markdownlint.json` with 0 errors

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `docs/simplify_refactor/tasks/011_end_to_end_verification_and_linting.md` |

## Reference

- [PRD.md](../PRD.md)
- [linting.bash](../../../scripts/linting.bash)
- [.markdownlint.json](../../../.markdownlint.json)

## Steps

- [ ] **Step 1: Run database migration validation**

  ```bash
  ./backend/migrations/scripts/run_migration_tests.bash
  ```

  Verify all tables, views, check constraints, and triggers pass.

- [ ] **Step 2: Run Go formatting and linting**

  ```bash
  cd backend
  gofumpt -l -w .
  golangci-lint run ./...
  ```

  Fix any formatting or linter warnings immediately.

- [ ] **Step 3: Run backend unit and integration test suites**

  ```bash
  cd backend
  go test -race ./... -v
  ```

  Verify 100% pass rate.

- [ ] **Step 4: Run frontend linting, type-checking, and tests**

  ```bash
  cd frontend
  npm run lint
  npm run type-check
  npm run test
  ```

  Confirm all tests pass and types are consistent.

- [ ] **Step 5: Verify purge of obsolete symbols**

  Search for obsolete symbols across the repository:

  ```bash
  git grep -i "useSlot" frontend/src/ || true
  git grep -i "SlotLetter" backend/internal/ || true
  git grep -i "open_participants" backend/ || true
  git grep -i "auth/register" backend/ frontend/ || true
  git grep -i "registerNewArcher" frontend/ || true
  ```

  Confirm that no active application code references these legacy entities.

- [ ] **Step 6: Run Markdownlint over all task documents**

  ```bash
  /home/juanpa/.npm/bin/markdownlint -c .markdownlint.json docs/simplify_refactor/tasks/*.md
  ```

  Verify 0 errors.

- [ ] **Step 7: Commit final verification status**

  ```bash
  git commit --allow-empty -m "chore: complete end-to-end verification of simplified refactor"
  ```

## Verification

- All automated tests, linters, and type checkers exit with code 0.
- Markdownlint validates all tasks with 0 warnings or errors.
