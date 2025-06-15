#!/usr/bin/env bash

set -eu
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
NEEDS_WEBUI=false
NEEDS_BACKEND_SERVER=false
NEEDS_BACKEND_TARGET_READER=false

run_python_checks() {
    local dir="${1}"
    if [[ ! -d $dir ]]; then
        echo "ERROR: '${dir}' is not a valid directory" >&2
        exit 2
    fi
    source "${ROOT_DIR}/backend/.venv/bin/activate"
    cd "${dir}"
    echo "running isort..."
    isort --settings-file "${ROOT_DIR}/backend/pyproject.toml" "${dir}"
    echo "running black..."
    black --config "${ROOT_DIR}/backend/pyproject.toml" "${dir}"
    echo "running mypy..."
    mypy --config-file "${ROOT_DIR}/backend/pyproject.toml" "${dir}"
    echo "running pylint..."
    pylint --rcfile "${ROOT_DIR}/backend/pyproject.toml" "${dir}"
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
        elif [[ "$file" == backend/server/* ]]; then
            NEEDS_BACKEND_SERVER=true
        elif [[ "$file" == backend/target_reader/* ]]; then
            NEEDS_BACKEND_TARGET_READER=true
        fi
    done
    if $NEEDS_WEBUI; then
        run_webui_checks
    fi
    if $NEEDS_BACKEND_SERVER; then
        run_python_checks "${ROOT_DIR}/backend/server"
    fi
    if $NEEDS_BACKEND_TARGET_READER; then
        run_python_checks "${ROOT_DIR}/backend/target_reader"
    fi
}

main

exit 0
