#!/usr/bin/env bash

# Purpose: Install Arch-Stats into $HOME, create .venv via uv, and write backend/.env
# Usage: run as the *target* user (e.g., via remote_installer.bash):
#   install.bash <db_user> <db_name> <db_socket_dir> <server_port>

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
TAR_FILE="arch-stats.tar.xz"

# Ensure user-local bin (uv installer) is reachable
export PATH="${HOME}/.local/bin:${PATH}"

checking_download_tar_file() {
    if tar -Jtvf "${SCRIPT_DIR}/${TAR_FILE}" >/dev/null; then
        echo "The tar file is valid."
    else
        echo "The tar file is corrupted or invalid." >&2
        exit 1
    fi
}

download_tar_file() {
    local url="https://github.com/jpmolinamatute/arch-stats/releases/download/latest/${TAR_FILE}"
    echo "Downloading ${TAR_FILE} from ${url}..."
    wget --output-document="${SCRIPT_DIR}/${TAR_FILE}" "${url}"
    checking_download_tar_file
    echo "Download completed successfully."
}

untar_files() {
    echo "Extracting ${TAR_FILE} into ${HOME}..."
    tar -xJf "${SCRIPT_DIR}/${TAR_FILE}" -C "${HOME}"
    rm -f "${SCRIPT_DIR}/${TAR_FILE}"
    echo "Extraction completed successfully."
}

install_uv() {
    if ! command -v uv >/dev/null 2>&1; then
        echo "Installing uv..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
        if ! uv --version >/dev/null 2>&1; then
            echo "ERROR: uv does not appear to be installed correctly" >&2
            exit 1
        fi
    fi
}

install_python_dependencies() {
    local backend_dir="${HOME}/backend"
    echo "Installing Python dependencies with uv..."
    install_uv
    cd "${backend_dir}"
    uv sync --no-dev --python "$(cat .python-version)"
    echo "Python dependencies installed successfully."
}

create_env_file() {
    local db_user="${1:?db_user required}"
    local db_name="${2:?db_name required}"
    local db_socket_dir="${3:-/var/run/postgresql}"
    local server_port="${4:-8000}"

    local env_file="${HOME}/backend/.env"
    local venv_bin="${HOME}/backend/.venv/bin"

    if [[ -f "${env_file}" ]]; then
        rm -f "${env_file}"
    fi

    echo "Creating ${env_file}..."
    cat >"${env_file}" <<EOF
# Runtime environment for Arch-Stats backend (read by systemd EnvironmentFile)
PATH=${venv_bin}:${HOME}/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
POSTGRES_USER=${db_user}
POSTGRES_DB=${db_name}
POSTGRES_SOCKET_DIR=${db_socket_dir}
ARCH_STATS_SERVER_PORT=${server_port}
ARCH_STATS_WS_CHANNEL=${USER}
EOF

    chmod 600 "${env_file}"
    echo "Environment file created at: ${env_file}"
}

main() {
    echo "Starting Arch-Stats installation process..."
    if [[ $# -lt 2 ]]; then
        echo "Usage: $0 <db_user> <db_name> [db_socket_dir=/var/run/postgresql] [server_port=8000]" >&2
        exit 2
    fi

    download_tar_file
    untar_files
    install_python_dependencies
    create_env_file "$@"
    echo "Arch-Stats installation completed successfully."
    exit 0
}

main "$@"
