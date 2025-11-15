#!/usr/bin/env bash

# This script is meant to be run on a remote machine to install arch-stats by local_installer.bash
# over SSH.

set -Eeuo pipefail

# Config
APP="arch-stats"
SYSTEM_SERVICE="${APP}.service"
ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"

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

Exit status codes:
    1   Generic fatal error / not root / extraction failure
    2   System user missing / cannot resolve home directory
    7   Dependency installer script missing or not executable
    14  Dependency installation (runuser) failure
    20  Service stop failure (only if unit exists and remains active)
    21  Service start failure
    22  Service not active after start

EOF
}
# Run dependency installation script as the app user
install_app_as_user() {
    local user_dir
    local script_path

    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    script_path="${ROOT_DIR}/install_app.bash"

    if [[ ! -d "${user_dir}" ]]; then
        log_error "System user '$APP' does not exist. Cannot determine home directory."
        exit 2
    fi

    if [[ ! -x "$script_path" ]]; then
        log_error "Dependency installer not found or not executable: $script_path"
        exit 7
    fi

    log_info "Running dependency installer as ${APP}: $script_path ${user_dir}"

    if ! runuser -u "${APP}" "${script_path}" "${user_dir}"; then
        log_error "Dependency installation failed for ${APP} (script exit)"
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

main() {
    # Help flag
    if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
        print_help
        exit 0
    fi

    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    stop_system_service
    install_app_as_user
    start_system_service
    assert_system_service_running
    log_info "Remote installation completed."
    exit 0
}

main "$@"
