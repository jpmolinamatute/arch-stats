#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

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
    echo "Running isort..."
    isort --settings-file "${pyproject_path}" .
    echo "Running black..."
    black --config "${pyproject_path}" .
    echo "Running mypy..."
    mypy --config-file "${pyproject_path}" .
    echo "Running pylint..."
    pylint --rcfile "${pyproject_path}" .
    run_python_tests "${pyproject_path}"
    cd -
}

run_frontend_checks() {
    cd "${ROOT_DIR}/frontend"
    echo "Running JS/TS linter"
    npm run lint
    echo "Running JS/TS formatter"
    npm run format
    cd -

}

run_bash_checks() {
    cd "${ROOT_DIR}/scripts"
    echo "Running bash linter"
    shellcheck --shell=bash --color=always -x ./*\.bash
    echo "Running bash formatter"
    shfmt --language-dialect bash --write -i 4 ./*\.bash
    cd -
}

main() {
    local needs_frontend=false
    local needs_backend=false
    local needs_scripts=false
    staged_files=$(git diff --cached --name-only)
    for file in $staged_files; do
        if [[ "$file" == frontend/* ]]; then
            needs_frontend=true
        elif [[ "$file" == backend/* ]]; then
            needs_backend=true
        elif [[ "$file" == scripts/* ]]; then
            needs_scripts=true
        fi
    done
    if $needs_frontend; then
        run_frontend_checks
    fi

    if $needs_backend; then
        run_python_checks
    fi

    if $needs_scripts; then
        run_bash_checks
    fi
    exit 0
}

main
