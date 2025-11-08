#!/usr/bin/env bash

set -Eeuo pipefail

: "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"

# Exit codes (installer):
# 1  : Generic fatal error / not root / extraction failure
# 7  : Dependency installer script missing or not executable
# 10 : PostgreSQL socket missing
# 11 : No SQL migrations found
# 12 : Network metadata/download failure
# 13 : Flyway migration failure
# 14 : Dependency installation (runuser) failure
# 20 : Service stop failure (only if unit exists and remains active)
# 21 : Service start failure
# 22 : Service not active after start

# Config
REPO="arch-stats"
SYSTEM_USER="archy"
SYSTEM_HOME=""
SYSTEM_SERVICE="${REPO}.service"
OWNER="jpmolinamatute"
USER_AGENT="${REPO}-installer"
ASSET_TARBALL_NAME="${REPO}.tar.xz"
TMP_DIR="$(mktemp -d -t "${REPO}-installer.XXXXXX")"
ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"
RELEASE_JSON_FILE="${TMP_DIR}/release.json"
MIGRATION_ZIP_OUT="${TMP_DIR}/${REPO}-migrations.zip"
MIGRATIONS_UNPACK_DIR="${TMP_DIR}/migrations_unpacked"
PG_SOCKET_DIR="${PG_SOCKET_DIR:-/var/run/postgresql}"
PG_PORT="${PG_PORT:-5432}"
MIGRATIONS_DIR=""

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

# Print usage and exit code reference
print_help() {
    local script
    script="${0##*/}"
    cat <<EOF
Usage: $script [--help]

Automates deployment of the latest release:
    - Fetches release metadata and tarball from GitHub and verifies SHA-256
    - Extracts into the system user's home directory
    - Downloads ephemeral SQL migrations and runs Flyway
    - Installs backend dependencies as the app user
    - Stops/starts the systemd service around the deployment

Requirements:
    - Must be run as root
    - Environment variables:
            GITHUB_TOKEN            (required) GitHub token with repo:read access
            PG_SOCKET_DIR           (optional) Default: /var/run/postgresql
            PG_PORT                 (optional) Default: 5432

Exit status codes:
    1   Generic fatal error / not root / extraction failure
    7   Dependency installer script missing or not executable
    10  PostgreSQL socket missing
    11  No SQL migrations found
    12  Network metadata/download failure
    13  Flyway migration failure
    14  Dependency installation (runuser) failure
    20  Service stop failure (only if unit exists and remains active)
    21  Service start failure
    22  Service not active after start

EOF
}

# Remove existing application directory to ensure clean install
purge_existing_install() {
    local app_dir
    app_dir="${SYSTEM_HOME}/${REPO}"
    if [[ -d "$app_dir" ]]; then
        log_info "Removing existing install at: $app_dir"
        rm -rf -- "$app_dir"
    else
        log_info "No existing install to remove at: $app_dir"
    fi
}

# Extract tarball into ${SYSTEM_HOME} (tar contains ${REPO} root) and fix ownership
extract_app() {
    local tar_path="$1"
    purge_existing_install
    log_info "Extracting ${ASSET_TARBALL_NAME} into ${SYSTEM_HOME}"
    if ! tar -xJf "$tar_path" -C "${SYSTEM_HOME}"; then
        log_error "Failed to extract application tarball"
        exit 1
    fi
    log_info "Setting ownership for ${SYSTEM_HOME}/${REPO} to ${SYSTEM_USER}:${SYSTEM_USER}"
    chown -R "${SYSTEM_USER}:${SYSTEM_USER}" "${SYSTEM_HOME}/${REPO}"
}

# Run dependency installation script as the app user
install_backend_dependencies_as_user() {
    local app_dir script_path
    app_dir="${SYSTEM_HOME}/${REPO}"
    script_path="${ROOT_DIR}/install_dependencies.bash"

    if [[ ! -x "$script_path" ]]; then
        log_error "Dependency installer not found or not executable: $script_path"
        exit 7
    fi

    log_info "Running dependency installer as ${SYSTEM_USER}: $script_path $app_dir"
    if ! runuser -u "${SYSTEM_USER}" -- "$script_path" "$app_dir"; then
        log_error "Dependency installation failed for ${SYSTEM_USER} (script exit)"
        exit 14
    fi
}

# Stop the systemd service for this application
stop_system_service() {
    if ! systemctl list-unit-files --type=service | grep -q "^${SYSTEM_SERVICE}"; then
        log_info "Service unit not installed; skipping stop: ${SYSTEM_SERVICE}"
        return 0
    fi
    log_info "Stopping systemd service: ${SYSTEM_SERVICE}"
    if ! systemctl stop "${SYSTEM_SERVICE}"; then
        log_error "Failed to stop service: ${SYSTEM_SERVICE}"
        exit 20
    fi
    if systemctl is-active --quiet "${SYSTEM_SERVICE}"; then
        log_error "Service still active after stop: ${SYSTEM_SERVICE}"
        exit 20
    fi
    log_info "Service stopped: ${SYSTEM_SERVICE}"
}

# Start the systemd service for this application
start_system_service() {
    log_info "Starting systemd service: ${SYSTEM_SERVICE}"
    if ! systemctl start "${SYSTEM_SERVICE}"; then
        log_error "Failed to start service: ${SYSTEM_SERVICE}"
        exit 21
    fi
    log_info "Start command issued for: ${SYSTEM_SERVICE}"
}

# Assert the systemd service is running (active)
assert_system_service_running() {
    log_info "Checking service is active: ${SYSTEM_SERVICE}"
    if ! systemctl is-active --quiet "${SYSTEM_SERVICE}"; then
        # Surface a concise status line to aid debugging
        local state
        state="$(systemctl is-active "${SYSTEM_SERVICE}" || true)"
        log_error "Service not active (state=${state}): ${SYSTEM_SERVICE}"
        exit 22
    fi
    log_info "Service is active: ${SYSTEM_SERVICE}"
}

# Download helper that sends Authorization for private assets and follows redirects
gh_download() {
    local url="$1" out="$2"
    curl -fL --max-time 60 --connect-timeout 10 \
        --retry 3 --retry-delay 2 --retry-all-errors \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        --user-agent "${USER_AGENT}" --output "$out" \
        "$url"
}

# shellcheck disable=SC2329 # Invoked indirectly via 'trap cleanup_tmp_workspace EXIT' in main
cleanup_tmp_workspace() {
    # shellcheck disable=SC2317
    if [[ -d "$TMP_DIR" ]]; then
        # shellcheck disable=SC2317
        rm -rf "$TMP_DIR"
        log_info "Removed temp workspace: ${TMP_DIR}"
    fi
}

get_repo_meta_data() {
    local api_url="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
    log_info "Resolving latest release metadata from GitHub API"
    if ! curl -fsSL --max-time 60 --connect-timeout 10 \
        --retry 3 --retry-delay 2 --retry-all-errors \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        --user-agent "${USER_AGENT}" \
        --output "${RELEASE_JSON_FILE}" \
        "$api_url"; then
        log_error "Failed to fetch latest release metadata"
        exit 12
    fi
    log_info "Saved release JSON to ${RELEASE_JSON_FILE}"
}

# Download the private migrations repository as a ZIP into TMP_DIR.
download_migrations_zip() {
    local repo="arch-stats-migrations"
    local zip_url

    zip_url="https://api.github.com/repos/${OWNER}/${repo}/zipball/main"

    log_info "Downloading migrations zip from: $zip_url"
    if ! gh_download "$zip_url" "${MIGRATION_ZIP_OUT}"; then
        log_error "Failed to download migrations zip"
        exit 12
    fi
    log_info "Migrations zip saved to: ${MIGRATION_ZIP_OUT}"
}

# Unpack the downloaded migrations zipball under TMP_DIR.
unpack_migrations_zip() {
    if [[ ! -f "${MIGRATION_ZIP_OUT}" ]]; then
        log_error "Migrations zip not found at ${MIGRATION_ZIP_OUT}. Did download_migrations_zip run?"
        exit 1
    fi
    mkdir -p "${MIGRATIONS_UNPACK_DIR}"
    log_info "Unpacking migrations zip into ${MIGRATIONS_UNPACK_DIR}"
    # GitHub zipball root contains a single top-level directory; we extract all then normalize path.
    if ! unzip -q "${MIGRATION_ZIP_OUT}" -d "${MIGRATIONS_UNPACK_DIR}"; then
        log_error "Failed to unzip migrations archive"
        exit 1
    fi
    # Capture the extracted top-level directory (there should be exactly one).

    MIGRATIONS_DIR="$(find "${MIGRATIONS_UNPACK_DIR}" -mindepth 1 -maxdepth 1 -type d | head -n1)"
    if [[ -z "${MIGRATIONS_DIR}" ]]; then
        log_error "Could not determine extracted root directory"
        exit 1
    fi
    log_info "Migrations extracted root: ${MIGRATIONS_DIR}"
    # List discovered SQL files for visibility.
    local sql_count
    sql_count="$(find "${MIGRATIONS_DIR}" -maxdepth 1 -type f -name '*.sql' | wc -l | tr -d ' ')"
    log_info "Found ${sql_count} SQL migration file(s) in archive"
    if [[ "${sql_count}" -eq 0 ]]; then
        log_error "No SQL migrations found (count=0)"
        exit 11
    fi

}

# Ensure the default PostgreSQL Unix socket exists before attempting migrations
assert_postgres_socket() {
    local socket_path
    socket_path="${PG_SOCKET_DIR}/.s.PGSQL.${PG_PORT}"
    if [[ ! -d "${PG_SOCKET_DIR}" ]]; then
        log_error "PostgreSQL socket directory not found at ${PG_SOCKET_DIR}. Is PostgreSQL running?"
        exit 10
    fi

    if [[ ! -S "${socket_path}" ]]; then
        log_error "PostgreSQL socket not found at ${socket_path}. Is PostgreSQL running?"
        exit 10
    fi
    log_info "Detected PostgreSQL socket: ${socket_path}"
}

# Run Flyway migrations pointing at the local PostgreSQL via Unix socket
run_flyway_migrations() {
    export FLYWAY_LOCATIONS="filesystem:${MIGRATIONS_DIR}"
    log_info "Running Flyway migrations using FLYWAY_LOCATIONS=${FLYWAY_LOCATIONS}"
    if ! flyway -baselineOnMigrate=true migrate; then
        log_error "Flyway migration failure"
        exit 13
    fi
    log_info "Flyway migrations completed successfully"
}

# Echoes the browser_download_url for the tarball asset
json_get_tarball_url() {
    local url
    url="$(jq -r --arg name "$ASSET_TARBALL_NAME" '.assets[] | select(.name==$name) | .browser_download_url' "${RELEASE_JSON_FILE}")"
    if [[ -z "$url" || "$url" == "null" ]]; then
        log_error "Could not find asset '$ASSET_TARBALL_NAME' in latest release."
        exit 1
    fi
    echo "$url"
}

json_get_sha256() {
    local checksum_asset_url checksum_file sha
    checksum_asset_url="$(jq -r --arg name "${ASSET_TARBALL_NAME}.sha256" '.assets[] | select(.name==$name) | .browser_download_url' "${RELEASE_JSON_FILE}")"

    if [[ -n "$checksum_asset_url" && "$checksum_asset_url" != "null" ]]; then
        checksum_file="${TMP_DIR}/$(basename "${ASSET_TARBALL_NAME}").sha256"
        log_info "Downloading checksum asset"
        gh_download "$checksum_asset_url" "$checksum_file"
        sha="$(grep -Eoi '^[0-9a-f]{64}' "$checksum_file" || true)"
    else
        log_error "Checksum asset ${ASSET_TARBALL_NAME}.sha256 not found in release assets. Failing."
        exit 1
    fi

    if [[ -z "$sha" ]]; then
        log_error "Checksum file did not contain a valid 64 hex sha256. Failing."
        exit 1
    fi
    echo "$sha"
}

verify_sha256() {
    local actual
    local file="$1" expected="$2"

    if [[ -z "$expected" ]]; then
        log_error "sha256 verification (no expected hash provided)"
        exit 1
    fi

    actual="$(sha256sum "$file" | awk '{print $1}')"
    if [[ "$actual" != "$expected" ]]; then
        log_error "sha256 mismatch. Expected=$expected Actual=$actual"
        exit 1
    fi
    log_info "sha256 OK: $actual"
}

main() {
    local tar_url checksum_hash tar_path

    # Help flag
    if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
        print_help
        exit 0
    fi
    trap cleanup_tmp_workspace EXIT
    if [[ $EUID -ne 0 ]]; then
        echo "Please run as root." >&2
        exit 1
    fi
    SYSTEM_HOME="$(getent passwd "$SYSTEM_USER" | cut -d: -f6)"
    stop_system_service
    # Verify local PostgreSQL is reachable via Unix socket
    assert_postgres_socket

    get_repo_meta_data

    tar_url="$(json_get_tarball_url)"
    checksum_hash="$(json_get_sha256)"
    tar_path="${TMP_DIR}/${ASSET_TARBALL_NAME}"
    gh_download "$tar_url" "$tar_path"
    verify_sha256 "$tar_path" "$checksum_hash"
    extract_app "$tar_path"
    download_migrations_zip
    unpack_migrations_zip
    run_flyway_migrations
    install_backend_dependencies_as_user
    start_system_service
    assert_system_service_running
    echo "Remote installation completed."
    exit 0
}

main "$@"
