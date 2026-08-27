# Task 034: Update Deployment Scripts + Systemd Service

## Git Branch

`refactor/034-deployment-scripts-and-systemd`

## Objective

Update the deployment scripts and systemd service file for the new single-binary Go deployment
model. The deployment simplifies drastically: download the binary, verify checksum, place it,
and run migrations — no venv, no `uv sync`, no tarball extraction.

## Dependencies

- Task 033 (CI produces the single Go binary as a release)
- Task 005 (Go binary can run migrations via `arch-stats migrate`)

## Acceptance Criteria

- [ ] `scripts/install_app.bash` is simplified:
    - Downloads the Go binary from GitHub Releases (not a tarball)
    - Verifies checksum
    - Places binary at `/opt/arch-stats/arch-stats`
    - Sets executable permissions
    - Runs migrations: `/opt/arch-stats/arch-stats migrate`
    - No venv creation, no `uv sync`, no pip
- [ ] `scripts/remote_installer.bash` is updated:
    - Same stop/start flow but simpler internals
    - References the binary, not the venv
- [ ] `scripts/deploy.bash` is updated for the new artifact type.
- [ ] Systemd service file (referenced in scripts or `Pi/arch-stats.service`) is updated:
    - `ExecStart=/opt/arch-stats/arch-stats` (single binary)
    - No `WorkingDirectory` pointing to venv
    - Environment variables loaded from `/opt/arch-stats/.env`
- [ ] `scripts/start_uvicorn.bash` is removed or replaced with a Go equivalent.
- [ ] All scripts pass `shellcheck`:

    ```bash
    shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash
    ```

## Files to Modify/Delete

| Action | Path |
| ------ | ---- |
| Modify | `scripts/install_app.bash` |
| Modify | `scripts/remote_installer.bash` |
| Modify | `scripts/deploy.bash` |
| Delete | `scripts/start_uvicorn.bash` |
| Modify | `scripts/templates/` (if systemd template exists) |

## Reference

- Current installer: [install_app.bash](file:///home/juanpa/Projects/arch-stats/scripts/install_app.bash)
- Current remote installer: [remote_installer.bash](file:///home/juanpa/Projects/arch-stats/scripts/remote_installer.bash)
- Current deploy: [deploy.bash](file:///home/juanpa/Projects/arch-stats/scripts/deploy.bash)
- Plan §12: single binary deployment model

## Steps

- [ ] **Step 1: Update `install_app.bash`**

  Simplify to:
  1. Check if running as root
  2. Download binary from GitHub Releases
  3. Verify SHA-256 checksum
  4. Stop service if running
  5. Place binary at `/opt/arch-stats/arch-stats`
  6. `chmod +x /opt/arch-stats/arch-stats`
  7. Run migrations: `/opt/arch-stats/arch-stats migrate`
  8. Start/restart service

  Remove all Python/venv/uv references.

- [ ] **Step 2: Update `remote_installer.bash`**

  Update to reference the binary instead of the tarball. Same SSH + stop/start flow.

- [ ] **Step 3: Update `deploy.bash`**

  Update to download the binary (not tarball) and deploy.

- [ ] **Step 4: Delete `scripts/start_uvicorn.bash`**

  ```bash
  git rm scripts/start_uvicorn.bash
  ```

- [ ] **Step 5: Update systemd service template**

  ```ini
  [Unit]
  Description=Arch Stats Server
  After=network.target postgresql.service
  Requires=postgresql.service

  [Service]
  Type=simple
  User=arch-stats
  Group=arch-stats
  ExecStart=/opt/arch-stats/arch-stats
  EnvironmentFile=/opt/arch-stats/.env
  Restart=on-failure
  RestartSec=5

  [Install]
  WantedBy=multi-user.target
  ```

- [ ] **Step 6: Run shellcheck**

  ```bash
  shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash
  ```

- [ ] **Step 7: Verify no Python/venv references remain**

  ```bash
  grep -rn "venv\|uvicorn\|uv sync\|pip\|python" scripts/install_app.bash scripts/deploy.bash
  ```

  Expected: no results.

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "chore: update deployment scripts for Go single binary model"
  ```

## Verification

- `shellcheck scripts/install_app.bash scripts/remote_installer.bash scripts/deploy.bash` — clean.
- `grep -rn "venv\|uvicorn\|uv sync" scripts/` — no Python references in deployment scripts.
- `scripts/start_uvicorn.bash` no longer exists.
- Systemd service references `/opt/arch-stats/arch-stats`.
