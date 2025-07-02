#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/check_docker
. "${ROOT_DIR}/tools/lib/check_docker"

start_uvicorn() {
    echo "Starting Uvicorn server"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${ROOT_DIR}/backend/server"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    exec uvicorn --loop uvloop --lifespan on --reload --ws websockets --http h11 --use-colors --log-level debug --timeout-graceful-shutdown 10 --factory server.app:run

}

main() {
    check_docker
    start_uvicorn
}

main
