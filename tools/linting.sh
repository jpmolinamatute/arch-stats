#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
# shellcheck source=./lib/check_docker
. "${ROOT_DIR}/tools/lib/check_docker"

run_python_tests() {
    check_docker
    echo "running python tests..."
    pytest --config-file "${ROOT_DIR}/backend/pyproject.toml"
    stop_docker
}

run_python_checks() {
    cd "${ROOT_DIR}/backend"
    export PYTHONPATH="${ROOT_DIR}/backend/src"
    # shellcheck source=../backend/.venv/bin/activate
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    echo "running isort..."
    isort --settings-file "${ROOT_DIR}/backend/pyproject.toml" .
    echo "running black..."
    black --config "${ROOT_DIR}/backend/pyproject.toml" .
    echo "running mypy..."
    mypy --config-file "${ROOT_DIR}/backend/pyproject.toml" .
    echo "running pylint..."
    pylint --rcfile "${ROOT_DIR}/backend/pyproject.toml" .
    run_python_tests
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
