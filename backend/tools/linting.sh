#!/usr/bin/env bash

DIR="${1}"

if [[ ! -d $DIR ]]; then
    echo "ERROR: '${DIR}' is not a valid directory" >&2
    exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
echo "running isort..."
uv run isort --settings-file "${ROOT_DIR}/pyproject.toml" "${DIR}"
echo "running black..."
uv run black --config "${ROOT_DIR}/pyproject.toml" "${DIR}"
echo "running mypy..."
uv run mypy --config-file "${ROOT_DIR}/pyproject.toml" "${DIR}"
echo "running pylint..."
uv run pylint --rcfile "${ROOT_DIR}/pyproject.toml" "${DIR}"
exit 0
