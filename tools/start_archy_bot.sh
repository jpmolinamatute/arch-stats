#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/check_docker
. "${ROOT_DIR}/tools/lib/check_docker"

start_archy() {
    echo "Starting Archy server"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${ROOT_DIR}/backend/src/target_reader"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    exec "${ROOT_DIR}/backend/src/target_reader/archy.py"
}

main() {
    check_docker
    start_archy
}

main
