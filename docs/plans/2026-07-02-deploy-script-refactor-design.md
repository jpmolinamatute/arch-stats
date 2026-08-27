# Deploy Script Refactor Design

**Date**: 2026-07-02  
**Feature/Refactor**: Break down `main` in `deploy.bash` into separate functions (`uninstall`, `install`, `update`).

## Overview

The `main` function in `scripts/deploy.bash` is currently long and contains distinct blocks of logic for clean installation, application updates, and uninstallation. Refactoring these blocks into dedicated functions improves readability, maintainability, and clean separation of concerns.

## Design

### Refactored Functions

1. **`uninstall()`**
   - **Arguments**: `host`
   - **Actions**: Uploads the remote uninstaller script and executes it on the remote host.

2. **`install()`**
   - **Arguments**: `host`
   - **Checks**: Verifies `GITHUB_TOKEN`, client ID, JWT secret, and Cloudflared credentials.
   - **Actions**: Renders templates locally, generates environment variables, uploads service configurations, and triggers remote installer script.

3. **`update()`**
   - **Arguments**: `host`
   - **Checks**: Verifies `GITHUB_TOKEN`.
   - **Actions**: Uploads `install_app.bash`, stops service, runs app installer, and starts service.

### Main Control Flow

The `main()` function will:
1. Parse arguments (`host`, `action`).
2. Load environment variables.
3. Verify SSH connection.
4. Auto-resolve the action if empty or set to `install`.
5. Dispatch to `uninstall`, `install`, or `update` with `host` as parameter.
