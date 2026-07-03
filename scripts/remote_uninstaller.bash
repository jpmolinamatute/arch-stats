#!/usr/bin/env bash

set -euo pipefail

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

main() {
    local app_user="${1}"
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    log_info "Stopping and disabling services..."
    for svc in "${app_user}.service" cloudflared.service; do
        if systemctl is-active --quiet "$svc" 2>/dev/null; then
            log_info "Stopping $svc..."
            systemctl stop "$svc" || true
        fi
        if systemctl is-enabled --quiet "$svc" 2>/dev/null; then
            log_info "Disabling $svc..."
            systemctl disable "$svc" || true
        fi
    done

    log_info "Removing systemd unit files..."
    rm -f "/etc/systemd/system/${app_user}.service" \
        /etc/systemd/system/cloudflared.service
    systemctl daemon-reload

    log_info "Dropping PostgreSQL database and user..."
    (
        cd /tmp
        if command -v dropdb >/dev/null 2>&1; then
            sudo -u postgres dropdb --if-exists "${app_user}" || true
        fi
        if command -v dropuser >/dev/null 2>&1; then
            sudo -u postgres dropuser --if-exists "${app_user}" || true
        fi
    )

    log_info "Purging cloudflared package..."
    if command -v cloudflared >/dev/null 2>&1; then
        apt-get purge -y cloudflared || true
    fi
    rm -rf /etc/cloudflared /usr/share/keyrings/cloudflare-main.gpg /etc/apt/sources.list.d/cloudflared.list

    log_info "Deleting user ${app_user} and home directory..."
    if id -u "${app_user}" >/dev/null 2>&1; then
        userdel -r "${app_user}" || true
    fi

    log_info "Uninstallation completed successfully."
    exit 0
}

main "$@"
