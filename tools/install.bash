#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
TAR_FILE="arch-stats.tar.xz"

export PATH="$HOME/.local/bin:${PATH}"

checking_download_tar_file() {
    if tar -Jtvf "${SCRIPT_DIR}/${TAR_FILE}" > /dev/null; then
        echo "The tar file is valid."
    else
        echo "The tar file is corrupted or invalid." >&2
        exit 1
    fi
}

download_tar_file(){
    local url="https://github.com/jpmolinamatute/arch-stats/releases/download/latest/${TAR_FILE}"
    echo "Downloading ${TAR_FILE} from ${url}..."
    wget --output-document="${SCRIPT_DIR}/${TAR_FILE}" "${url}"
    checking_download_tar_file
    echo "Download completed successfully."
}

untar_files() {
    echo "Extracting ${TAR_FILE}..."
    tar -xJf "${SCRIPT_DIR}/${TAR_FILE}" -C "${HOME}"
    echo "Extraction completed successfully."
    rm -f "${SCRIPT_DIR}/${TAR_FILE}"
}

install_uv(){
    if ! command -v uv >/dev/null 2>&1; then
        echo "Installing uv..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
        if ! uv --version; then
            echo "ERROR: uv does not appear to be installed correctly" >&2
            exit 1
        fi
    fi

}


install_python_dependencies() {
    local backend_dir="${HOME}/backend"
    install_uv
    echo "Installing Python dependencies..."
    cd "${backend_dir}" || exit 1
    uv sync --no-dev --python "$(cat .python-version)"
    echo "Python dependencies installed successfully."
}

main() {
    echo "Starting Arch-Stats installation process..."
    download_tar_file
    untar_files
    install_python_dependencies
    echo "Arch-Stats installation completed successfully."
    exit 0
}

main "$@"
