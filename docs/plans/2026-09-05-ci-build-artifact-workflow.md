# Task 033: Rewrite Build Artifact CI Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `.github/workflows/build_artifact.yaml` to replace the legacy Python/Flyway pipeline with a Go cross-compilation pipeline that builds the Vue 3 frontend, embeds the assets into a `linux/arm64` binary, generates SHA-256 checksums, publishes a single binary release on GitHub Releases, and marks all task items as done.

**Architecture:**
- Workflow file `.github/workflows/build_artifact.yaml`:
  - Strips out Postgres service, Flyway migrations, uv setup, Python environment, tarball creation, and legacy scripts.
  - Implements sequential steps: checkout repository, set up Go 1.27.0 (`actions/setup-go@v5`), setup Node.js and dependencies (`./.github/actions/npm-setup`), install swag CLI, generate OpenAPI specs (`swag init`), copy spec to `openapi.json`, generate frontend types (`npm run generate:types`), build frontend SPA (`npm run build`), copy `frontend/dist` to `backend/frontend` for `//go:embed`, cross-compile static Go binary for `linux/arm64` with stripped symbols (`-s -w`), compute SHA-256 checksum, and publish both binary and checksum to GitHub Releases via `softprops/action-gh-release@v2`.
- Documentation & Task Tracking:
  - Update `docs/plans/task.md` live table-only tracker across all tasks.
  - Check off all Acceptance Criteria and Steps in `docs/go_refactor/tasks/033-ci_build_artifact_workflow.md` as `[x]`.

**Tech Stack:** GitHub Actions, Go 1.27 (`GOOS=linux GOARCH=arm64 CGO_ENABLED=0`), Swaggo (`swag`), Node.js 24, Vite / Vue 3, `softprops/action-gh-release@v2`.

**Spec:**
- [docs/go_refactor/tasks/033-ci_build_artifact_workflow.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/033-ci_build_artifact_workflow.md)
- [.github/workflows/build_artifact.yaml](file:///home/juanpa/Projects/arch-stats/.github/workflows/build_artifact.yaml)

## Global Constraints

- Git branch: `refactor/033-ci-build-artifact-workflow` branched from `main`.
- The release artifact must be a single executable binary (`arch-stats`), not a tarball or archive.
- Target platform: `linux/arm64` (Raspberry Pi 5) with `CGO_ENABLED=0`.
- No Python, uv, Flyway, venv, or tar references may remain in `.github/workflows/build_artifact.yaml`.
- The `uv-setup` custom action must be removed; the `npm-setup` custom action must be retained.
- YAML syntax must be strictly valid (`python3 -c "import yaml; yaml.safe_load(open('.github/workflows/build_artifact.yaml'))"`).
- Single-flow execution model: exactly one active task tracked in `docs/plans/task.md`.
- Final step must mark all acceptance criteria and steps in `docs/go_refactor/tasks/033-ci_build_artifact_workflow.md` and `docs/plans/task.md` as completed (`[x]` / `DONE`).

---

## File Structure

```
.github/
└── workflows/
    └── build_artifact.yaml                  # [MODIFY] Rewrite workflow for Go single-binary arm64 release
docs/
├── plans/
│   ├── task.md                              # [MODIFY] Live checklist tracker (table-only)
│   └── 2026-09-05-ci-build-artifact-workflow.md # [NEW] Implementation plan
└── go_refactor/
    └── tasks/
        └── 033-ci_build_artifact_workflow.md # [MODIFY] Mark all acceptance criteria and steps as checked [x]
```

---

## Task Structure

### Task 1: Git Branch Setup & Live Tracker Initialization

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Clean working tree on `main` branch
- Produces: `refactor/033-ci-build-artifact-workflow` branch and initialized `docs/plans/task.md` tracker

- [ ] **Step 1: Create and switch to the refactor git branch**

```bash
git checkout -b refactor/033-ci-build-artifact-workflow
```

- [ ] **Step 2: Initialize `docs/plans/task.md` live tracker**

Write the initial status table for Task 033 to `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | IN_PROGRESS | Create branch `refactor/033-ci-build-artifact-workflow` and initialize `task.md` |
| Task 2: Workflow Pipeline Rewrite | PENDING | Rewrite `.github/workflows/build_artifact.yaml` for Go arm64 single-binary release |
| Task 3: Local Dry-Run & Pipeline Step Verification | PENDING | Validate swag init, type generation, frontend build, and arm64 cross-compilation |
| Task 4: Mark Tasks as Done, Documentation Updates & Final Commit | PENDING | Mark `033-ci_build_artifact_workflow.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Verify clean working directory**

Run:
```bash
git status
```
Expected: On branch `refactor/033-ci-build-artifact-workflow`, only `docs/plans/task.md` (and the plan file) staged or untracked.

- [ ] **Step 4: Mark Task 1 as DONE in `docs/plans/task.md`**

Update `Task 1` status to `DONE` and `Task 2` status to `IN_PROGRESS` in `docs/plans/task.md`.

---

### Task 2: Workflow Pipeline Rewrite (`.github/workflows/build_artifact.yaml`)

**Files:**
- Modify: `.github/workflows/build_artifact.yaml`

**Interfaces:**
- Consumes: Task 033 workflow specification
- Produces: Syntactically valid, Python-free GitHub Actions workflow file

- [ ] **Step 1: Rewrite `.github/workflows/build_artifact.yaml`**

Replace the entire content of `.github/workflows/build_artifact.yaml` with:

```yaml
name: Build Artifact

on:
  push:
    branches:
      - main

permissions:
  contents: write

defaults:
  run:
    shell: bash

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.27.0"

      - name: Setup Node & Install Dependencies
        uses: ./.github/actions/npm-setup

      - name: Install swag CLI
        run: go install github.com/swaggo/swag/cmd/swag@latest

      - name: Generate OpenAPI spec
        working-directory: ./backend
        run: swag init -g cmd/arch-stats/main.go -o specs/

      - name: Copy spec for frontend type generation
        run: cp backend/specs/swagger.json openapi.json

      - name: Generate frontend types
        working-directory: ./frontend
        run: npm run generate:types

      - name: Build frontend
        working-directory: ./frontend
        run: npm run build

      - name: Copy frontend build to backend for embedding
        run: |
          rm -rf backend/frontend
          cp -r frontend/dist backend/frontend

      - name: Cross-compile Go binary (linux/arm64)
        working-directory: ./backend
        env:
          GOOS: linux
          GOARCH: arm64
          CGO_ENABLED: "0"
        run: go build -ldflags="-s -w" -o "${GITHUB_WORKSPACE}/arch-stats" ./cmd/arch-stats

      - name: Generate checksum
        run: sha256sum "${GITHUB_WORKSPACE}/arch-stats" > "${GITHUB_WORKSPACE}/arch-stats.sha256"

      - name: Create GitHub release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: v${{ github.run_number }}
          name: Build v${{ github.run_number }}
          target_commitish: ${{ github.sha }}
          make_latest: true
          files: |
            ${{ github.workspace }}/arch-stats
            ${{ github.workspace }}/arch-stats.sha256
          fail_on_unmatched_files: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Validate YAML syntax using Python YAML parser**

Run:
```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/build_artifact.yaml'))"
```
Expected: Exits with code 0 and no syntax errors.

- [ ] **Step 3: Verify no Python, uv, Flyway, or tar references remain**

Run:
```bash
grep -iE "uv|python|uvicorn|flyway|tar|venv" .github/workflows/build_artifact.yaml || echo "CLEAN: No legacy references found"
```
Expected: `CLEAN: No legacy references found`.

- [ ] **Step 4: Verify custom action references**

Verify that `uv-setup` is gone and `npm-setup` is referenced:
```bash
grep "uv-setup" .github/workflows/build_artifact.yaml || echo "CLEAN: No uv-setup found"
grep "npm-setup" .github/workflows/build_artifact.yaml
```
Expected:
`CLEAN: No uv-setup found`
`uses: ./.github/actions/npm-setup`

- [ ] **Step 5: Mark Task 2 as DONE in `docs/plans/task.md`**

Update `Task 2` status to `DONE` and `Task 3` status to `IN_PROGRESS` in `docs/plans/task.md`.

---

### Task 3: Local Dry-Run & Pipeline Step Verification

**Files:**
- Inspect: `backend/cmd/arch-stats/main.go`
- Inspect: `backend/specs/`
- Inspect: `frontend/`

**Interfaces:**
- Consumes: Local Go and Node environment
- Produces: Verified execution of every command included in the CI workflow

- [ ] **Step 1: Test Swagger doc generation**

Run:
```bash
cd backend && swag init -g cmd/arch-stats/main.go -o specs/
```
Expected: `create swagger.json at specs/swagger.json` with exit code 0.

- [ ] **Step 2: Test frontend type generation and build**

Run:
```bash
cp backend/specs/swagger.json openapi.json
cd frontend && npm run generate:types && npm run build
```
Expected: Successful type generation and Vite build output in `frontend/dist`.

- [ ] **Step 3: Test Go arm64 cross-compilation with embedded frontend**

Run:
```bash
cd /home/juanpa/Projects/arch-stats
mkdir -p /tmp/arch-stats-dist
cp -r frontend/dist/* /tmp/arch-stats-dist/
cd backend
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/arch-stats-verify ./cmd/arch-stats
file /tmp/arch-stats-verify
sha256sum /tmp/arch-stats-verify > /tmp/arch-stats-verify.sha256
cat /tmp/arch-stats-verify.sha256
```
Expected:
Executable identified as `ELF 64-bit LSB executable, ARM aarch64`, statically linked, stripped, and checksum produced.

- [ ] **Step 4: Clean up temporary verification artifacts and keep workspace clean**

Run:
```bash
rm -rf /tmp/arch-stats-verify* /tmp/arch-stats-dist openapi.json
git checkout backend/specs/
```
Expected: Workspace restored to clean state with only workflow changes.

- [ ] **Step 5: Mark Task 3 as DONE in `docs/plans/task.md`**

Update `Task 3` status to `DONE` and `Task 4` status to `IN_PROGRESS` in `docs/plans/task.md`.

---

### Task 4: Mark Tasks as Done, Documentation Updates & Final Commit

**Files:**
- Modify: `docs/go_refactor/tasks/033-ci_build_artifact_workflow.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified `.github/workflows/build_artifact.yaml` and dry-run verification
- Produces: Updated checklist in task spec, finalized live tracker, and committed git branch

- [ ] **Step 1: Update `docs/go_refactor/tasks/033-ci_build_artifact_workflow.md`**

Mark all Acceptance Criteria and Steps as completed (`[x]`):
- `[x] .github/workflows/build_artifact.yaml is rewritten with this pipeline:`
  - `[x] 1. Checkout code`
  - `[x] 2. Set up Go 1.27.0`
  - `[x] 3. Set up Node.js (for frontend build)`
  - `[x] 4. Install frontend dependencies (npm ci)`
  - `[x] 5. Generate OpenAPI spec (swag init)`
  - `[x] 6. Generate frontend types (npm run generate:types)`
  - `[x] 7. Build frontend (npm run build)`
  - `[x] 8. Copy frontend build output to backend/frontend/ (for //go:embed)`
  - `[x] 9. Cross-compile Go binary: GOOS=linux GOARCH=arm64 go build -o arch-stats ./cmd/arch-stats`
  - `[x] 10. Generate checksum: sha256sum arch-stats > arch-stats.sha256`
  - `[x] 11. Create GitHub Release with the binary and checksum`
- `[x] The release artifact is a single binary (not a tarball).`
- `[x] No Python, uv, Flyway, or venv references remain.`
- `[x] The uv-setup custom action is no longer referenced.`
- `[x] The npm-setup custom action is still referenced for the frontend.`
- `[x] Step 1: Rewrite the workflow`
- `[x] Step 2: Validate YAML syntax`
- `[x] Step 3: Verify no Python references remain`
- `[x] Step 4: Commit`

- [ ] **Step 2: Update `docs/plans/task.md` to reflect all tasks DONE**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Live Tracker Initialization | DONE | Create branch `refactor/033-ci-build-artifact-workflow` and initialize `task.md` |
| Task 2: Workflow Pipeline Rewrite | DONE | Rewrite `.github/workflows/build_artifact.yaml` for Go arm64 single-binary release |
| Task 3: Local Dry-Run & Pipeline Step Verification | DONE | Validate swag init, type generation, frontend build, and arm64 cross-compilation |
| Task 4: Mark Tasks as Done, Documentation Updates & Final Commit | DONE | Mark `033-ci_build_artifact_workflow.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Commit changes to git**

Run:
```bash
git add .github/workflows/build_artifact.yaml docs/go_refactor/tasks/033-ci_build_artifact_workflow.md docs/plans/task.md docs/plans/2026-09-05-ci-build-artifact-workflow.md
git commit -m "ci: rewrite build artifact workflow for Go single binary release"
```

- [ ] **Step 4: Verify working tree clean**

Run:
```bash
git status
```
Expected: `nothing to commit, working tree clean`.

---

## Self-Review Checklist

1. **Spec Coverage:**
   - `.github/workflows/build_artifact.yaml` rewritten with 11-step pipeline? Covered in Task 2.
   - Release artifact is a single binary (not a tarball)? Covered in Task 2.
   - No Python, uv, Flyway, or venv references? Covered in Task 2.
   - `uv-setup` removed and `npm-setup` retained? Covered in Task 2.
   - YAML syntax verified? Covered in Task 2.
   - Cross-compilation for `linux/arm64` tested? Covered in Task 3.
   - Marking tasks as done at the end of implementation? Covered in Task 4 (both `033-ci_build_artifact_workflow.md` and `task.md`).
2. **Placeholder Scan:**
   - No "TBD", "TODO", or unwritten code blocks exist in any step.
3. **Type and Syntax Consistency:**
   - Exact YAML structure, paths, and flags match the task specification.
