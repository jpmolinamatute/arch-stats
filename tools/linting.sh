#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/tools/lib/manage_docker"

run_python_tests() {
    local pyproject_path="${1}"
    start_docker
    echo "running python tests..."
    pytest --config-file "${pyproject_path}"
    stop_docker
}

run_python_checks() {
    local pyproject_path="${ROOT_DIR}/backend/pyproject.toml"
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    echo "running isort..."
    isort --settings-file "${pyproject_path}" .
    echo "running black..."
    black --config "${pyproject_path}" .
    echo "running mypy..."
    mypy --config-file "${pyproject_path}" .
    echo "running pylint..."
    pylint --rcfile "${pyproject_path}" .
    run_python_tests "${pyproject_path}"
    cd -
}

run_webui_checks() {
    cd "${ROOT_DIR}/webui"
    npm run lint
    npm run format
    cd -

}

main() {
    local needs_webui=false
    local needs_backend=false
    staged_files=$(git diff --cached --name-only)
    for file in $staged_files; do
        if [[ "$file" == webui/* ]]; then
            needs_webui=true
        elif [[ "$file" == backend/* ]]; then
            needs_backend=true
        fi
    done
    if $needs_webui; then
        run_webui_checks
    fi
    if $needs_backend; then
        run_python_checks
    fi
}

main

exit 0
