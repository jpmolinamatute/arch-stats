# Design Document: Restore install_app.bash functionality

## Goal
Restore the functionality of `scripts/install_app.bash` after the user's modifications diverted it from the `main` branch. The script must remain generic and dynamic (deriving the repository and database name from the executing system user) while fixing all parameter passing bugs and variable mismatches.

## Proposed Design (Approach 2)

### Global Script-Wide Variables
We will use script-scoped variables dynamically derived at script loading time:
```bash
APP_NAME="$(whoami)"
OWNER="jpmolinamatute"
USER_AGENT="${APP_NAME}-installer"
ASSET_TARBALL_NAME="${APP_NAME}.tar.xz"
TMP_DIR="$(mktemp -d -t "${APP_NAME}-installer.XXXXXX")"
RELEASE_JSON_FILE="${TMP_DIR}/release.json"
MIGRATION_ZIP_OUT="${TMP_DIR}/${APP_NAME}-migrations.zip"
MIGRATIONS_UNPACK_DIR="${TMP_DIR}/migrations_unpacked"
PG_SOCKET_DIR="/var/run/postgresql"
PG_PORT="5432"

export PATH="${HOME}/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
```

### Argument Overrides and `HOME`
To support both parameterized paths (like `/opt/arch-stats` passed during deployments) and local user execution, we will override `HOME` at the start of `main()` if a parameter is provided:
```bash
main() {
    ...
    HOME="${1:-${HOME}}"
    ...
}
```
This ensures all functions referencing `${HOME}` (like `purge_existing_install` and `install_dependencies`) automatically resolve to the correct directory.

### Functions and Signatures
All function signatures will be simplified back to the main branch version, removing complex argument passing:
*   `cleanup_tmp_workspace()`: Cleans up `$TMP_DIR`.
*   `purge_existing_install()`: Removes existing installation at `${HOME}/backend`.
*   `gh_download(url, out, api_call)`: Standard wrapper around `curl`.
*   `get_repo_meta_data()`: Downloads latest release metadata from GitHub API.
*   `json_get_tarball_url()`: Extracts the tarball download URL using `jq`.
*   `json_get_sha256()`: Extracts the checksum, downloads it, and parses the expected SHA.
*   `verify_sha256(file, expected)`: Compares SHA256 checksums.
*   `extract_app(tar_path)`: Unpacks the tarball to `${HOME}`.
*   `download_migrations_zip()`: Downloads the migrations ZIP archive.
*   `unpack_migrations_zip()`: Unpacks migrations and runs them.
*   `run_migrations(migrations_dir)`: Runs SQL migrations using peer socket authentication.
*   `assert_postgres_socket()`: Verifies that the PG socket exists.
*   `install_dependencies()`: Auto-installs/updates `uv` and runs `uv sync`.

## Verification Plan

### Automated Checks
*   Run `shellcheck scripts/install_app.bash` to verify syntax and bash rules.
*   Run `shfmt -d scripts/install_app.bash` to verify formatting.
