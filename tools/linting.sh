#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
NEEDS_WEBUI=false
NEEDS_BACKEND=false

run_python_tests() {
    local compose_files="${ROOT_DIR}/docker/docker-compose.yaml"
    is_running=$(docker compose -f "${compose_files}" ps --all --quiet | wc -l)
    if [[ ${is_running} -eq 0 ]]; then
        echo "Starting docker compose"
        docker compose -f "${compose_files}" up --detach --build
        sleep 2
    fi

    echo "running python tests..."
    pytest --config-file "${ROOT_DIR}/backend/pyproject.toml"
    echo "Stopping docker compose"
    docker compose -f "${compose_files}" down -v
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
    staged_files=$(git diff --cached --name-only)
    for file in $staged_files; do
        if [[ "$file" == webui/* ]]; then
            NEEDS_WEBUI=true
        elif [[ "$file" == backend/* ]]; then
            NEEDS_BACKEND=true
        fi
    done
    if $NEEDS_WEBUI; then
        run_webui_checks
    fi
    if $NEEDS_BACKEND; then
        run_python_checks
    fi
}

main

exit 0
