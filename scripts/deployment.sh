#!/usr/bin/env bash

set -e

ROOT_DIR="$(dirname "$(dirname "$(realpath "$0")")")"

# shellcheck source=./lib/helpers
. "${ROOT_DIR}/scripts/lib/helpers"

compile_rust() {
    cd "${ROOT_DIR}/backend" || exit 1
    print_out "Compiling Rust code..."
    if ! cargo build --target aarch64-unknown-linux-gnu --release --workspace -j "$(nproc)"; then
        print_err "Failed to compile Rust code"
    fi
    cd - || exit 1
}

move_rust_binaries() {
    local app_user="${1}"
    local rust_binaries=("arrow_reader" "bow_reader" "create_uuid" "main_socket" "server" "target_reader")
    local release_dir="${ROOT_DIR}/backend/target/aarch64-unknown-linux-gnu/release"

    print_out "Moving all Rust binaries to ${app_user}1..."
    for binary in "${rust_binaries[@]}"; do
        if [[ -f "${release_dir}/${binary}" ]]; then
            print_out "Moving ${binary}..."
            scp "${release_dir}/${binary}" "${app_user}1:/opt/${app_user}/backend"
        else
            print_err "Binary ${binary} not found"
        fi
    done
}

build_webui() {
    local webui_dir="${ROOT_DIR}/webui"
    print_out "Building WebUI..."
    cd "${webui_dir}" || exit 1
    if ! ng build --delete-output-path -c production; then
        print_err "Failed to build WebUI"
    fi
}

move_webui_files() {
    local app_user="${1}"
    local webui_dir="${ROOT_DIR}/dist/webui/browser"
    print_out "Moving all Rust binaries to ${app_user}1..."
    if ! scp "${webui_dir}"/* "${app_user}1:/opt/${app_user}/webui"; then
        print_err "Failed to move WebUI files"
    fi
}

main() {
    local app_user="arch-stats"
    echo "Starting deployment..."
    compile_rust
    move_rust_binaries "${app_user}"
    build_webui
    move_webui_files "${app_user}"
    echo "Deployment completed successfully."
}

main

exit 0
