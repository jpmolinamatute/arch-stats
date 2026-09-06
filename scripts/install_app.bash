#!/usr/bin/env bash

set -Eeu

: "${GITHUB_TOKEN:?Environment variable GITHUB_TOKEN is not set}"

export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

# shellcheck disable=SC2329
cleanup_tmp_workspace() {
    local tmp_dir="${1}"
    if [[ -d "${tmp_dir}" ]]; then
        rm -rf "${tmp_dir}"
        log_info "Removed temp workspace: ${tmp_dir}"
    fi
}

gh_download() {
    local url="${1}"
    local out="${2}"
    local api_call="${3:-false}"
    local curl_ec
    local curl_opts=(
        -fsSL
        --max-time 120
        --connect-timeout 15
        --retry 3
        --retry-delay 2
        --retry-all-errors
        --user-agent "arch-stats-installer"
        -H "Authorization: Bearer ${GITHUB_TOKEN}"
        --output "${out}"
    )
    if [[ "${api_call}" == "true" ]]; then
        curl_opts+=(
            -H "Accept: application/vnd.github+json"
            -H "X-GitHub-Api-Version: 2022-11-28"
        )
    fi
    if ! curl "${curl_opts[@]}" "${url}"; then
        curl_ec=$?
        log_error "Download failed (curl exit=${curl_ec}) url=${url}"
        exit 15
    fi
}

get_release_metadata() {
    local base_url="${1}"
    local release_json_file="${2}"
    local api_url="${base_url}/releases/latest"

    log_info "Resolving latest release metadata from GitHub API"
    gh_download "${api_url}" "${release_json_file}" true
}

json_get_asset_url() {
    local asset_name="${1}"
    local release_json_file="${2}"
    local url
    url="$(jq -r --arg name "${asset_name}" '.assets[] | select(.name==$name) | .browser_download_url' "${release_json_file}")"
    if [[ -z "${url}" || "${url}" == "null" ]]; then
        log_error "Could not find asset '${asset_name}' in latest release."
        exit 1
    fi
    echo "${url}"
}

get_expected_sha256() {
    local checksum_url="${1}"
    local checksum_file="${2}"
    local sha

    gh_download "${checksum_url}" "${checksum_file}" false
    sha="$(grep -Eoi '^[0-9a-f]{64}' "${checksum_file}" | head -n 1 || true)"
    if [[ -z "${sha}" ]]; then
        log_error "Checksum file did not contain a valid 64-hex SHA-256 hash."
        exit 1
    fi
    echo "${sha}"
}

verify_sha256() {
    local file="${1}"
    local expected="${2}"
    local actual

    actual="$(sha256sum "${file}" | awk '{print $1}')"
    if [[ "${actual}" != "${expected}" ]]; then
        log_error "SHA-256 mismatch. Expected='${expected}' Actual='${actual}'"
        exit 1
    fi
    log_info "SHA-256 OK: ${actual}"
}

assert_postgres_socket() {
    local socket_path="/var/run/postgresql/.s.PGSQL.5432"
    if [[ ! -d /var/run/postgresql/ || ! -S "${socket_path}" ]]; then
        log_error "PostgreSQL socket not found at ${socket_path}. Is PostgreSQL running?"
        exit 10
    fi
    log_info "Detected PostgreSQL socket: ${socket_path}"
}

install_migrations() {
    local base_url="${1}"
    local install_dir="${2}"
    local tmp_dir="${3}"
    local app_user="${4}"
    local target_migrations="${install_dir}/migrations"
    local staged_migrations="/tmp/deploy_assets/migrations"

    mkdir -p "${target_migrations}"

    if [[ -d "${staged_migrations}" ]] && compgen -G "${staged_migrations}/*.sql" >/dev/null; then
        log_info "Installing migrations from staged assets..."
        cp "${staged_migrations}"/*.sql "${target_migrations}/"
    else
        local zip_url="${base_url}-migrations/zipball/main"
        local zip_file="${tmp_dir}/migrations.zip"
        local unpack_dir="${tmp_dir}/migrations_unpacked"

        log_info "Downloading migrations archive from ${zip_url}..."
        gh_download "${zip_url}" "${zip_file}" false

        mkdir -p "${unpack_dir}"
        unzip -q "${zip_file}" -d "${unpack_dir}"

        local src_dir
        src_dir="$(find "${unpack_dir}" -mindepth 1 -maxdepth 1 -type d | head -n 1)"
        if [[ -z "${src_dir}" ]]; then
            log_error "Failed to locate extracted migrations directory."
            exit 11
        fi
        cp "${src_dir}"/*.sql "${target_migrations}/"
    fi

    chown -R "${app_user}:${app_user}" "${target_migrations}"
    chmod 755 "${target_migrations}"
    chmod 644 "${target_migrations}"/*.sql
    log_info "Installed SQL migrations to ${target_migrations}"
}

main() {
    local app_user="${1:-arch-stats}"
    local app_name="arch-stats"
    local repo="jpmolinamatute/arch-stats"
    local base_url="https://api.github.com/repos/${repo}"
    local install_dir="/opt/${app_user}"
    local target_bin="${install_dir}/${app_name}"

    if [[ ${EUID} -ne 0 ]]; then
        log_error "This script must be run as root."
        exit 1
    fi

    local tmp_dir
    tmp_dir="$(mktemp -d -t "arch-stats-installer.XXXXXX")"
    trap 'cleanup_tmp_workspace "${tmp_dir}"' EXIT

    assert_postgres_socket

    local release_json="${tmp_dir}/release.json"
    get_release_metadata "${base_url}" "${release_json}"

    local bin_url checksum_url expected_sha
    bin_url="$(json_get_asset_url "${app_name}" "${release_json}")"
    checksum_url="$(json_get_asset_url "${app_name}.sha256" "${release_json}")"

    local dl_bin="${tmp_dir}/${app_name}"
    local dl_checksum="${tmp_dir}/${app_name}.sha256"

    log_info "Downloading ${app_name} binary..."
    gh_download "${bin_url}" "${dl_bin}" false

    log_info "Downloading ${app_name}.sha256 checksum..."
    expected_sha="$(get_expected_sha256 "${checksum_url}" "${dl_checksum}")"

    log_info "Verifying SHA-256 checksum..."
    verify_sha256 "${dl_bin}" "${expected_sha}"

    if systemctl is-active --quiet "${app_user}.service" 2>/dev/null; then
        log_info "Stopping active ${app_user}.service before updating binary..."
        systemctl stop "${app_user}.service" || true
    fi

    mkdir -p "${install_dir}"
    log_info "Installing binary to ${target_bin}..."
    install -m 755 -o "${app_user}" -g "${app_user}" "${dl_bin}" "${target_bin}"

    install_migrations "${base_url}" "${install_dir}" "${tmp_dir}" "${app_user}"

    log_info "Running database migrations as user ${app_user}..."
    if ! runuser -u "${app_user}" -- bash -c "cd '${install_dir}' && '${target_bin}' migrate"; then
        log_error "Database migrations failed."
        exit 13
    fi
    log_info "Database migrations completed successfully."

    if [[ -f "/etc/systemd/system/${app_user}.service" ]]; then
        log_info "Starting ${app_user}.service..."
        systemctl restart "${app_user}.service"
    fi

    log_info "Installation completed successfully."
    exit 0
}

main "$@"
