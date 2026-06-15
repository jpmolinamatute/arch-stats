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
    23  Cloudflared service activation failure

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
# Install and configure cloudflared tunnel
install_cloudflared() {
    # Check if there is a credentials JSON file in /tmp/
    local cred_file
    cred_file="$(find /tmp -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ ! -f "${cred_file}" ]]; then
        log_info "No Cloudflared credentials file found in /tmp. Skipping Cloudflared setup."
        return 0
    fi

    local tunnel_id
    tunnel_id="$(basename "${cred_file}" .json)"
    log_info "Setting up Cloudflared tunnel: ${tunnel_id}"

    # Install cloudflared binary if not present
    if ! command -v cloudflared >/dev/null 2>&1; then
        log_info "Installing cloudflared via apt..."

        # Add Cloudflare GPG key
        if ! curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null; then
            log_error "Failed to download Cloudflare GPG key"
            exit 1
        fi

        # Add Cloudflare apt repository
        local release_codename="bookworm"
        if [[ -f /etc/os-release ]]; then
            # shellcheck disable=SC1091
            release_codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-bookworm}")"
        fi
        echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared ${release_codename} main" | tee /etc/apt/sources.list.d/cloudflared.list >/dev/null

        # Install package
        if ! apt-get update -y && apt-get install -y cloudflared; then
            log_error "Failed to install cloudflared package."
            exit 1
        fi
    fi

    # Determine app home dir
    local user_dir
    user_dir="$(getent passwd "$APP" | cut -d: -f6)"
    if [[ ! -d "${user_dir}" ]]; then
        log_error "System user '$APP' does not exist. Cannot determine home directory."
        exit 2
    fi

    # Copy credential file to app home dir
    local cloudflared_dir="${user_dir}/.cloudflared"
    mkdir -p "${cloudflared_dir}"
    chmod 700 "${cloudflared_dir}"

    local dest_cred_file="${cloudflared_dir}/${tunnel_id}.json"
    cp "${cred_file}" "${dest_cred_file}"
    chmod 600 "${dest_cred_file}"
    chown -R "${APP}:${APP}" "${cloudflared_dir}"

    # Copy config file to /etc/cloudflared/config.yml
    log_info "Installing cloudflared configuration file..."
    mkdir -p "/etc/cloudflared"
    chmod 755 "/etc/cloudflared"

    if [[ ! -f "/tmp/config.yml" ]]; then
        log_error "Cloudflared config file not found at /tmp/config.yml"
        exit 24
    fi
    cp "/tmp/config.yml" "/etc/cloudflared/config.yml"
    chmod 644 "/etc/cloudflared/config.yml"

    # Copy systemd files
    log_info "Copying cloudflared systemd files..."
    cp "/tmp/cloudflared.service" "/tmp/cloudflared-update.service" "/tmp/cloudflared-update.timer" "/etc/systemd/system/"
    chmod 644 /etc/systemd/system/cloudflared.service /etc/systemd/system/cloudflared-update.service /etc/systemd/system/cloudflared-update.timer

    # Enable and start services
    log_info "Activating cloudflared systemd services..."
    if ! systemctl daemon-reload; then
        log_error "Failed to reload systemd daemon"
        exit 24
    fi

    if ! systemctl enable cloudflared.service; then
        log_error "Failed to enable cloudflared.service"
        exit 24
    fi
    if ! systemctl restart cloudflared.service; then
        log_error "Failed to start/restart cloudflared.service"
        exit 24
    fi

    if ! systemctl enable cloudflared-update.timer; then
        log_error "Failed to enable cloudflared-update.timer"
        exit 24
    fi
    if ! systemctl restart cloudflared-update.timer; then
        log_error "Failed to start/restart cloudflared-update.timer"
        exit 24
    fi

    # Validate active state
    if ! systemctl is-active --quiet cloudflared.service; then
        log_error "cloudflared.service is not active after activation"
        exit 24
    fi
    if ! systemctl is-active --quiet cloudflared-update.timer; then
        log_error "cloudflared-update.timer is not active after activation"
        exit 24
    fi

    log_info "cloudflared services activated successfully"

    # Clean up temp files from /tmp
    log_info "Cleaning up temporary cloudflared files..."
    rm -f "/tmp/cloudflared.service" "/tmp/cloudflared-update.service" "/tmp/cloudflared-update.timer" "/tmp/config.yml" "${cred_file}"
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
    install_cloudflared
    log_info "Remote installation completed."
    exit 0
}

main "$@"
