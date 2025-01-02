#!/usr/bin/env bash

set -e

LOCAL_USER="arch-stats"
ROOT_DIR="$(dirname "$(dirname "$(realpath "$0")")")"

print_out() {
    local msg="${1}"
    echo -e "\e[1;32m${msg}\e[0m"
}

print_err() {
    local msg="${1}"
    echo -e "\e[1;31m${msg}\e[0m" >&2
    exit 1
}

compile_rust() {
    cd "${ROOT_DIR}/backend" || exit 1
    print_out "Compiling Rust code..."
    if ! cargo build --target aarch64-unknown-linux-gnu --release --workspace -j "$(nproc)"; then
        print_err "Failed to compile Rust code"
    fi
    cd - || exit 1
}

move_rust_binaries() {
    local rust_binaries=("arrow_reader" "bow_reader" "create_uuid" "main_socket" "server" "target_reader")
    local release_dir="${ROOT_DIR}/backend/target/aarch64-unknown-linux-gnu/release"

    print_out "Moving all Rust binaries to ${LOCAL_USER}1..."
    for binary in "${rust_binaries[@]}"; do
        if [[ -f "${release_dir}/${binary}" ]]; then
            print_out "Moving ${binary}..."
            scp "${release_dir}/${binary}" "${LOCAL_USER}1:/opt/${LOCAL_USER}/backend"
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
    local webui_dir="${ROOT_DIR}/dist/webui/browser"
    print_out "Moving all Rust binaries to ${LOCAL_USER}1..."
    if ! scp "${webui_dir}"/* "${LOCAL_USER}1:/opt/${LOCAL_USER}/webui"; then
        print_err "Failed to move WebUI files"
    fi
}

main() {
    echo "Starting deployment..."
    compile_rust
    move_rust_binaries
    build_webui
    move_webui_files
    echo "Deployment completed successfully."
}

main

exit 0
