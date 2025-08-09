#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

start_archy() {
    echo "Starting Archy server"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${ROOT_DIR}/backend/src/target_reader"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    exec "${ROOT_DIR}/backend/src/target_reader/archy.py"
}

main() {
    start_docker
    start_archy
}

main
