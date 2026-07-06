#!/usr/bin/env bash

set -Eeuo pipefail

# We need GITHUB_TOKEN to download the latest version of the bundle app from GitHub
: "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"

export PATH="${HOME}/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

# shellcheck disable=SC2329 # Invoked indirectly via 'trap cleanup_tmp_workspace EXIT' in main
cleanup_tmp_workspace() {
    local tmp_dir="${1}"
    if [[ -d "$tmp_dir" ]]; then
        rm -rf "$tmp_dir"
        log_info "Removed temp workspace: ${tmp_dir}"
    fi
}

# Remove existing application directory to ensure clean install
purge_existing_install() {
    local backend_dir="${HOME}/backend"
    if [[ -d "$backend_dir" ]]; then
        log_info "Removing existing install at: $backend_dir"
        rm -rf -- "$backend_dir"
    fi
}

gh_download() {
    local app_name="${1}"
    local url="${2}"
    local out="${3}"
    local api_call="${4:-false}"
    local curl_ec
    local user_agent="${app_name}-installer"
    local curl_opts=(
        -fsSL
        --max-time 60
        --connect-timeout 10
        --retry 3
        --retry-delay 2
        --retry-all-errors
        --user-agent "${user_agent}"
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
    local app_name="${1}"
    local base_url="${2}"
    local release_json_file="${3}"
    local api_url="${base_url}/releases/latest"

    log_info "Resolving latest release metadata from GitHub API for ${app_name}"
    gh_download "${app_name}" "$api_url" "${release_json_file}" true
    log_info "Saved release JSON to ${release_json_file}"
}

# Echoes the browser_download_url for the tarball asset
json_get_tarball_url() {
    local asset_tarball_name="${1}"
    local release_json_file="${2}"
    local url
    url="$(jq -r --arg name "${asset_tarball_name}" '.assets[] | select(.name==$name) | .browser_download_url' "${release_json_file}")"
    if [[ -z "$url" || "$url" == "null" ]]; then
        log_error "Could not find asset '${asset_tarball_name}' in latest release."
        exit 1
    fi
    echo "$url"
}

json_get_sha256() {
    local app_name="${1}"
    local asset_tarball_name="${2}"
    local tmp_dir="${3}"
    local release_json_file="${4}"
    local checksum_asset_url checksum_file sha
    checksum_asset_url="$(jq -r --arg name "${asset_tarball_name}.sha256" '.assets[] | select(.name==$name) | .browser_download_url' "${release_json_file}")"

    if [[ -n "$checksum_asset_url" && "$checksum_asset_url" != "null" ]]; then
        checksum_file="${tmp_dir}/${asset_tarball_name}.sha256"

        gh_download "${app_name}" "$checksum_asset_url" "$checksum_file" false
        sha="$(grep -Eoi '^[0-9a-f]{64}' "$checksum_file" || true)"
    else
        log_error "Checksum asset ${asset_tarball_name}.sha256 not found in release assets. Failing."
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
    local asset_tarball_name="${2}"

    log_info "Extracting ${asset_tarball_name} into ${HOME}"
    if ! tar -xJf "$tar_path" -C "${HOME}"; then
        log_error "Failed to extract application tarball"
        exit 1
    fi
}

download_migrations_zip() {
    local app_name="${1}"
    local base_url="${2}"
    local migration_zip_out="${3}"
    local zip_url
    zip_url="${base_url}-migrations/zipball/main"

    log_info "Downloading migrations zip from: $zip_url"
    if ! gh_download "${app_name}" "$zip_url" "${migration_zip_out}" false; then
        log_error "Failed to download migrations zip"
        exit 12
    fi
    log_info "Migrations zip saved to: ${migration_zip_out}"
}

# Run migrations pointing at the local PostgreSQL DB via Unix socket
run_migrations() {
    local app_name="${1}"
    local migrations_dir="${2}"
    log_info "Running migrations '${migrations_dir}'"
    while IFS= read -r -d '' f; do
        # - ON_ERROR_STOP stops on SQL errors
        if ! psql -h /var/run/postgresql -p 5432 -v ON_ERROR_STOP=1 -U "${app_name}" -d "${app_name}" -f "${f}"; then
            log_error "Migration failed"
            exit 13
        fi
    done < <(find "$migrations_dir" -maxdepth 1 -type f -name '*.sql' -print0 | sort -z)
    log_info "Migrations completed successfully"
}

unpack_migrations_zip() {
    local migrations_dir
    local app_name="${1}"
    local migrations_unpack_dir="${2}"
    local migration_zip_out="${3}"
    if [[ ! -f "${migration_zip_out}" ]]; then
        log_error "Migrations zip not found at ${migration_zip_out}. Did download_migrations_zip run?"
        exit 1
    fi
    mkdir -p "${migrations_unpack_dir}"
    log_info "Unpacking migrations zip into ${migrations_unpack_dir}"
    # GitHub zipball root contains a single top-level directory; we extract all then normalize path.
    if ! unzip -q "${migration_zip_out}" -d "${migrations_unpack_dir}"; then
        log_error "Failed to unzip migrations archive"
        exit 1
    fi
    # Capture the extracted top-level directory (there should be exactly one).

    migrations_dir="$(find "${migrations_unpack_dir}" -mindepth 1 -maxdepth 1 -type d | head -n1)"
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
    run_migrations "${app_name}" "${migrations_dir}"
}

assert_postgres_socket() {
    local socket_path="/var/run/postgresql/.s.PGSQL.5432"
    if [[ ! -d /var/run/postgresql/ ]]; then
        log_error "PostgreSQL socket directory not found at /var/run/postgresql/. Is PostgreSQL running?"
        exit 10
    fi
    if [[ ! -S "${socket_path}" ]]; then
        log_error "PostgreSQL socket not found at ${socket_path}. Is PostgreSQL running?"
        exit 10
    fi
    log_info "Detected PostgreSQL socket: ${socket_path}"
}

install_dependencies() {
    local backend_dir="${HOME}/backend"

    if ! command -v uv >/dev/null 2>&1; then
        log_info "uv not found, installing uv..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
    else
        log_info "Running 'uv self update'"
        if ! uv self update; then
            log_error "uv self update failed"
            exit 5
        fi
    fi

    cd "${backend_dir}" || {
        log_error "backend directory not found: ${backend_dir}"
        exit 4
    }

    log_info "Syncing production dependencies (no dev, frozen)"
    if ! uv sync --no-dev --frozen --python "$(cat .python-version)"; then
        log_error "uv sync failed"
        exit 6
    fi

    log_info "Dependency installation completed successfully"
}

main() {
    local tar_url checksum_hash tar_path base_url tmp_dir app_name
    app_name="$(whoami)"
    local asset_tarball_name="${app_name}.tar.xz"
    tmp_dir="$(mktemp -d -t "${app_name}-installer.XXXXXX")"
    local migration_zip_out="${tmp_dir}/${app_name}-migrations.zip"
    local migrations_unpack_dir="${tmp_dir}/migrations_unpacked"
    local release_json_file="${tmp_dir}/release.json"

    base_url="https://api.github.com/repos/jpmolinamatute/${app_name}"
    trap 'cleanup_tmp_workspace "${tmp_dir}"' EXIT

    if [[ -d "${HOME}" ]]; then
        if [[ ! -r "${HOME}" || ! -w "${HOME}" ]]; then
            log_error "directory ${HOME} exists but is not readable and/or writable"
            exit 2
        fi
        cd "${HOME}"
    else
        log_error "HOME is missing or is not a real directory"
        exit 2
    fi
    purge_existing_install
    assert_postgres_socket
    get_repo_meta_data "${app_name}" "${base_url}" "${release_json_file}"
    tar_url="$(json_get_tarball_url "${asset_tarball_name}" "${release_json_file}")"
    checksum_hash="$(json_get_sha256 "${app_name}" "${asset_tarball_name}" "${tmp_dir}" "${release_json_file}")"
    tar_path="${tmp_dir}/${asset_tarball_name}"
    gh_download "${app_name}" "$tar_url" "$tar_path" false
    verify_sha256 "$tar_path" "$checksum_hash"
    extract_app "$tar_path" "${asset_tarball_name}"
    download_migrations_zip "${app_name}" "${base_url}" "${migration_zip_out}"
    unpack_migrations_zip "${app_name}" "${migrations_unpack_dir}" "${migration_zip_out}"
    install_dependencies

    exit 0
}

main "$@"
