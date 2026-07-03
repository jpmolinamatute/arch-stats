#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

setup_app_user() {
    local app_user="${1}"
    local user_dir

    if ! id -u "${app_user}" >/dev/null 2>&1; then
        log_info "Creating system user '${app_user}'..."
        useradd -r -m -d "/opt/${app_user}" -s "/usr/sbin/nologin" "${app_user}"
    fi
    user_dir="$(getent passwd "${app_user}" | cut -d: -f6)"

    if [[ -f "/tmp/.env" ]]; then
        mv "/tmp/.env" "${user_dir}/.env"
        chown "${app_user}:${app_user}" "${user_dir}/.env"
        chmod 600 "${user_dir}/.env"
    fi
}

setup_postgres() {
    local app_user="${1}"
    log_info "Setting up PostgreSQL..."
    if ! systemctl list-units --type=service | grep -q "postgresql"; then
        log_info "Installing PostgreSQL..."
        apt-get update -y && apt-get install -y postgresql postgresql-contrib
        systemctl enable --now postgresql
    fi

    # Run PostgreSQL commands from /tmp to avoid "could not change directory to '/root': Permission denied"
    (
        cd /tmp
        if ! sudo -u postgres psql -t -c '\du' | cut -d \| -f 1 | grep -q "${app_user}"; then
            log_info "Creating PostgreSQL user '${app_user}'..."
            sudo -u postgres createuser "${app_user}"
        fi

        if ! sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw "${app_user}"; then
            log_info "Creating PostgreSQL database '${app_user}'..."
            sudo -u postgres createdb -O "${app_user}" "${app_user}"
        fi
    )
    log_info "PostgreSQL setup complete."
}

setup_cloudflared() {
    local app_user="${1}"
    local src_cred_file
    local dest_cred_file
    local src_cert_file="/tmp/cert.pem"
    local dest_cert_file
    local user_dir
    local cf_dir
    local codename="bookworm"
    log_info "Setting up cloudflared..."
    src_cred_file="$(find /tmp -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ ! -f "${src_cred_file}" || ! -f "${src_cert_file}" ]]; then
        log_error "Cloudflared credentials file or cert file not found in /tmp. Aborting."
        exit 23
    fi

    if ! command -v cloudflared >/dev/null 2>&1; then
        log_info "Installing cloudflared..."
        curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null

        if [[ -f /etc/os-release ]]; then
            codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-bookworm}")"
        fi
        echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared ${codename} main" | tee /etc/apt/sources.list.d/cloudflared.list >/dev/null
        apt-get update -y
        apt-get install -y cloudflared
    fi

    user_dir="$(getent passwd "${app_user}" | cut -d: -f6)"
    cf_dir="${user_dir}/.cloudflared"
    dest_cred_file="${cf_dir}/$(basename "${src_cred_file}")"
    dest_cert_file="${cf_dir}/$(basename "${src_cert_file}")"
    mkdir -p "${cf_dir}"
    mv "${src_cred_file}" "${dest_cred_file}"
    mv "${src_cert_file}" "${dest_cert_file}"
    chmod 700 "${cf_dir}"
    chmod 600 "${dest_cred_file}" "${dest_cert_file}"
    chown -R "${app_user}:${app_user}" "${cf_dir}"

    mkdir -p "/etc/cloudflared"
    mv /tmp/cloudflared_config.yaml /etc/cloudflared/cloudflared_config.yaml
    mv /tmp/cloudflared.service /etc/systemd/system/
    chmod 644 /etc/systemd/system/cloudflared.service /etc/cloudflared/cloudflared_config.yaml

    systemctl daemon-reload
    systemctl enable --now cloudflared.service
    log_info "Cloudflared setup complete."
}

register_app_service() {
    local app_user="${1}"
    mv "/tmp/${app_user}.service" "/etc/systemd/system/"
    chmod 644 "/etc/systemd/system/${app_user}.service"
    systemctl daemon-reload
    systemctl enable --now "${app_user}.service"
}

install_app_as_user() {
    local app_user="${1}"
    local user_dir
    local script_path="${ROOT_DIR}/install_app.bash"
    user_dir="$(getent passwd "$app_user" | cut -d: -f6)"
    chmod +x "$script_path"
    log_info "Running application installer as ${app_user}..."
    if ! runuser -u "${app_user}" -- bash "${script_path}" "${user_dir}"; then
        log_error "Application installation failed."
        exit 14
    fi
}

main() {
    local app_user="${1}"
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    setup_app_user "${app_user}"
    setup_postgres "${app_user}"
    setup_cloudflared "${app_user}"
    register_app_service "${app_user}"
    install_app_as_user "${app_user}"

    log_info "Starting ${app_user} service..."
    systemctl start "${app_user}.service"

    if ! systemctl is-active --quiet "${app_user}.service"; then
        log_error "Service ${app_user}.service failed to start."
        exit 22
    fi

    log_info "Cleaning up temporary installation files..."
    rm -f /tmp/remote_installer.bash /tmp/install_app.bash \
        /tmp/cloudflared_config.yaml "/tmp/${app_user}.service" \
        /tmp/cloudflared.service /tmp/cloudflared-update.service /tmp/cloudflared-update.timer \
        /tmp/*.json

    log_info "Remote installation completed successfully."
    exit 0
}

main "$@"
