#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd -P)"

log_info() { echo "INFO: $*"; }
log_error() { echo "ERROR: $*" >&2; }

create_app_user() {
    local app_user="${1}"
    if ! id -u "${app_user}" >/dev/null 2>&1; then
        log_info "Creating system user '${app_user}'..."
        useradd -r -m -d "/opt/${app_user}" -s "/usr/sbin/nologin" "${app_user}"
    fi
}

generate_env_file() {
    local app_user="${1}"
    local postgres_password="${2}"
    local user_dir
    local jwt_secret
    user_dir="$(getent passwd "${app_user}" | cut -d: -f6)"

    jwt_secret="$(openssl rand -hex 32)"

    cat <<EOF >"${user_dir}/.env"
POSTGRES_USER="${app_user}"
POSTGRES_PASSWORD="${postgres_password}"
POSTGRES_DB="${app_user}"
POSTGRES_HOST="localhost"
POSTGRES_PORT="5433"
POSTGRES_SOCKET_DIR="/var/run/postgresql"
ARCH_STATS_SERVER_PORT="8001"
ARCH_STATS_DEV_MODE="false"
POSTGRES_POOL_MIN_SIZE="1"
POSTGRES_POOL_MAX_SIZE="10"
ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID=""
VITE_GOOGLE_CLIENT_ID=""
ARCH_STATS_JWT_SECRET="${jwt_secret}"
ARCH_STATS_JWT_ALGORITHM="HS256"
ARCH_STATS_JWT_TTL_MINUTES="60"
EOF

    chown "${app_user}:${app_user}" "${user_dir}/.env"
    chmod 600 "${user_dir}/.env"
}

install_os_packages() {
    log_info "Configuring Cloudflare repository..."
    local codename="bookworm"
    if [[ -f /etc/os-release ]]; then
        codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-bookworm}")"
    fi
    curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
    echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared ${codename} main" | tee /etc/apt/sources.list.d/cloudflared.list >/dev/null

    log_info "Installing OS packages (cloudflared, postgresql, postgresql-contrib, openssl)..."
    apt-get update -y
    apt-get install -y \
        cloudflared \
        postgresql \
        postgresql-contrib \
        openssl

    systemctl enable --now postgresql
}

setup_postgres() {
    local app_user="${1}"
    local postgres_password="${2}"
    log_info "Setting up PostgreSQL user and database..."

    # Run PostgreSQL commands from /tmp to avoid "could not change directory to '/root': Permission denied"
    (
        cd /tmp
        if ! sudo -u postgres psql -t -c '\du' | cut -d \| -f 1 | grep -q "${app_user}"; then
            log_info "Creating PostgreSQL user '${app_user}'..."
            sudo -u postgres psql -c "CREATE USER \"${app_user}\" WITH PASSWORD '${postgres_password}';"
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
    log_info "Setting up cloudflared..."
    src_cred_file="$(find /tmp -maxdepth 1 -name "*.json" | head -n 1)"
    if [[ ! -f "${src_cred_file}" || ! -f "${src_cert_file}" ]]; then
        log_error "Cloudflared credentials file or cert file not found in /tmp. Aborting."
        exit 23
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
    local postgres_password
    if [[ $EUID -ne 0 ]]; then
        log_error "Please run as root."
        exit 1
    fi

    install_os_packages
    postgres_password="$(openssl rand -hex 16)"
    create_app_user "${app_user}"
    generate_env_file "${app_user}" "${postgres_password}"
    setup_postgres "${app_user}" "${postgres_password}"
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
