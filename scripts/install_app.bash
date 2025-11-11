#!/usr/bin/env bash

set -Eeuo pipefail

: "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"
: "${POSTGRES_USER:?Environment variable POSTGRES_USER is not set}"
: "${POSTGRES_DB:?Environment variable POSTGRES_DB is not set}"

REPO="arch-stats"
OWNER="jpmolinamatute"
USER_AGENT="${REPO}-installer"
ASSET_TARBALL_NAME="${REPO}.tar.xz"
TMP_DIR="$(mktemp -d -t "${REPO}-installer.XXXXXX")"
RELEASE_JSON_FILE="${TMP_DIR}/release.json"
MIGRATION_ZIP_OUT="${TMP_DIR}/${REPO}-migrations.zip"
MIGRATIONS_UNPACK_DIR="${TMP_DIR}/migrations_unpacked"
PG_SOCKET_DIR="${PG_SOCKET_DIR:-/var/run/postgresql}"
PG_PORT="${PG_PORT:-5432}"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

print_help() {
    local script
    script="${0##*/}"
    cat <<EOF
Usage: $script APP_HOME_DIR
       $script --help

APP_HOME_DIR:
    (required) Absolute path to the application system user's home directory
    where the release archive will be extracted. Example: /home/arch-stats

Description:
    Automates deployment of the latest release:
      - Fetches release metadata and tarball from GitHub and verifies SHA-256
      - Extracts into the system user's home directory
      - Downloads ephemeral SQL migrations and applies them
      - Installs backend dependencies (production only)
      - (Service management is not performed by this script version)

Environment Variables:
    GITHUB_TOKEN    (required) GitHub token with repo:read access
    POSTGRES_USER   (required) Database user owning the target schema
    POSTGRES_DB     (required) Target database name
    PG_SOCKET_DIR   (optional) PostgreSQL socket dir (default: /var/run/postgresql)
    PG_PORT         (optional) PostgreSQL port (default: 5432)

Exit Status Codes:
    1   Generic fatal error / extraction failure
    2   Missing or invalid APP_HOME_DIR
    10  PostgreSQL socket missing
    11  No SQL migrations found
    12  Network metadata/download failure
    13  Migration failure
    15  Download helper (curl) failure

Example:
    $script /home/arch-stats

EOF
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

# Remove existing application directory to ensure clean install
purge_existing_install() {
    local app_home_dir="${1}/backend"
    if [[ -d "$app_home_dir" ]]; then
        log_info "Removing existing install at: $app_home_dir"
        rm -rf -- "$app_home_dir"
    fi
}

gh_download() {
    local url="$1"
    local out="$2"
    local api_call="${3:-false}"
    local curl_ec
    local curl_opts=(
        -fsSL
        --max-time 60
        --connect-timeout 10
        --retry 3
        --retry-delay 2
        --retry-all-errors
        --user-agent "${USER_AGENT:-arch-stats-installer}"
        -H "Authorization: Bearer ${GITHUB_TOKEN:-}"
        --output "$out"
    )
    local gh_api_headers=(
        -H "Accept: application/vnd.github+json"
        -H "X-GitHub-Api-Version: 2022-11-28"
    )
    if [[ "$api_call" == "true" ]]; then
        curl_opts+=("${gh_api_headers[@]}")
    fi
    if ! curl "${curl_opts[@]}" "$url"; then
        curl_ec=$?
        log_error "Download failed (curl exit=${curl_ec}) url=${url}"
        exit 15
    fi
}

get_repo_meta_data() {
    local api_url="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"

    log_info "Resolving latest release metadata from GitHub API"
    if ! gh_download "$api_url" "${RELEASE_JSON_FILE}" true; then
        log_error "Failed to fetch latest release metadata"
        exit 12
    fi
    log_info "Saved release JSON to ${RELEASE_JSON_FILE}"
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
        checksum_file="${TMP_DIR}/${ASSET_TARBALL_NAME}.sha256"
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
        log_error "sha256 mismatch. Expected='${expected}' Actual='${actual}'"
        exit 1
    fi
    log_info "sha256 OK: $actual"
}

extract_app() {
    local tar_path="$1"
    local backend_dir="${2}"
    log_info "Extracting ${ASSET_TARBALL_NAME} into ${backend_dir}"
    if ! tar -xJf "$tar_path" -C "${backend_dir}"; then
        log_error "Failed to extract application tarball"
        exit 1
    fi
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

# Run migrations pointing at the local PostgreSQL via Unix socket
run_migrations() {
    local migrations_dir="${1}"
    log_info "Running migrations '${migrations_dir}'"
    while IFS= read -r -d '' f; do
        if ! psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f "${f}"; then
            log_error "Migration failed"
            exit 13
        fi
    done < <(find "$migrations_dir" -maxdepth 1 -type f -name '*.sql' -print0 | sort -z)

    log_info "Migrations completed successfully"
}

# Unpack the downloaded migrations zipball under TMP_DIR.
unpack_migrations_zip() {
    local migrations_dir
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

    migrations_dir="$(find "${MIGRATIONS_UNPACK_DIR}" -mindepth 1 -maxdepth 1 -type d | head -n1)"
    if [[ -z "${migrations_dir}" ]]; then
        log_error "Could not determine extracted root directory"
        exit 1
    fi
    log_info "Migrations extracted root: ${migrations_dir}"
    # List discovered SQL files for visibility.
    local sql_count
    sql_count="$(find "${migrations_dir}" -maxdepth 1 -type f -name '*.sql' | wc -l | tr -d ' ')"
    log_info "Found ${sql_count} SQL migration file(s) in archive"
    if [[ "${sql_count}" -eq 0 ]]; then
        log_error "No SQL migrations found (count=0)"
        exit 11
    fi
    run_migrations "${migrations_dir}"
}

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

install_dependencies() {
    local app_home_dir="${1}/backend"

    cd "$app_home_dir" || {
        log_error "backend directory not found: $app_home_dir"
        exit 4
    }
    log_info "Running 'uv self update'"
    if ! uv self update; then
        log_error "uv self update failed"
        exit 5
    fi

    log_info "Syncing production dependencies (no dev, frozen)"
    if ! uv sync --no-dev --frozen --python "$(cat .python-version)"; then
        log_error "uv sync failed"
        exit 6
    fi

    log_info "Dependency installation completed successfully"
}

main() {
    # Help flag
    if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
        print_help
        exit 0
    fi
    trap cleanup_tmp_workspace EXIT
    local home_dir="${1}"
    if [[ ! -d "$home_dir" ]]; then
        log_error "required argument: home_dir is missing or is not a real directory"
        log_error "Usage: $0 /path/to/arch-stats"
        exit 2
    fi
    purge_existing_install "${home_dir}"
    assert_postgres_socket
    get_repo_meta_data
    tar_url="$(json_get_tarball_url)"
    checksum_hash="$(json_get_sha256)"
    tar_path="${TMP_DIR}/${ASSET_TARBALL_NAME}"
    gh_download "$tar_url" "$tar_path"
    verify_sha256 "$tar_path" "$checksum_hash"
    extract_app "$tar_path" "${home_dir}"
    download_migrations_zip
    unpack_migrations_zip
    install_dependencies "${home_dir}"

    exit 0
}

main "$@"
