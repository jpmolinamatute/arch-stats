# Task 033: Rewrite Build Artifact CI Workflow

## Git Branch

`refactor/033-ci-build-artifact-workflow`

## Objective

Rewrite the `.github/workflows/build_artifact.yaml` CI workflow for Go. The pipeline changes
fundamentally: build the frontend, cross-compile the Go binary with embedded frontend assets
for `linux/arm64` (Raspberry Pi 5), and publish the single binary as a GitHub Release.

## Dependencies

- Task 028 (frontend embedding via `//go:embed`)
- Task 029 (OpenAPI spec generation via swag)
- Task 025 (Go binary compiles with all handlers)

## Acceptance Criteria

- [ ] `.github/workflows/build_artifact.yaml` is rewritten with this pipeline:
  1. Checkout code
  2. Set up Go 1.27.0
  3. Set up Node.js (for frontend build)
  4. Install frontend dependencies (`npm ci`)
  5. Generate OpenAPI spec (`swag init`)
  6. Generate frontend types (`npm run generate:types`)
  7. Build frontend (`npm run build`)
  8. Copy frontend build output to `backend/frontend/` (for `//go:embed`)
  9. Cross-compile Go binary: `GOOS=linux GOARCH=arm64 go build -o arch-stats ./cmd/arch-stats`
  10. Generate checksum: `sha256sum arch-stats > arch-stats.sha256`
  11. Create GitHub Release with the binary and checksum
- [ ] The release artifact is a single binary (not a tarball).
- [ ] No Python, uv, Flyway, or venv references remain.
- [ ] The `uv-setup` custom action is no longer referenced.
- [ ] The `npm-setup` custom action is still referenced for the frontend.

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `.github/workflows/build_artifact.yaml` |

## Reference

- Current workflow: [build_artifact.yaml](file:///home/juanpa/Projects/arch-stats/.github/workflows/build_artifact.yaml)
- Plan §11 and §12: cross-compile, single binary, GitHub Release

## Steps

- [ ] **Step 1: Rewrite the workflow**

  Replace the entire contents of `.github/workflows/build_artifact.yaml`:

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

- [ ] **Step 2: Validate YAML syntax**

  ```bash
  python3 -c "import yaml; yaml.safe_load(open('.github/workflows/build_artifact.yaml'))"
  ```

- [ ] **Step 3: Verify no Python references remain**

  ```bash
  grep -i "uv\|python\|uvicorn\|flyway\|tar\|venv" .github/workflows/build_artifact.yaml
  ```

  Expected: no results.

- [ ] **Step 4: Commit**

  ```bash
  git add -A
  git commit -m "ci: rewrite build artifact workflow for Go single binary release"
  ```

## Verification

- `cat .github/workflows/build_artifact.yaml` — references Go cross-compilation, not Python.
- YAML is valid.
- No references to uv, Python, Flyway, tar, or venv.
- Release artifact is a single binary, not a tarball.
