#!/usr/bin/env bash

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
echo "running isort..."
uv run isort --settings-file "${ROOT_DIR}/pyproject.toml" "${ROOT_DIR}/server"
echo "running black..."
uv run black --config "${ROOT_DIR}/pyproject.toml" "${ROOT_DIR}/server"
echo "running mypy..."
uv run mypy --config-file "${ROOT_DIR}/pyproject.toml" "${ROOT_DIR}/server"
echo "running pylint..."
uv run pylint --rcfile "${ROOT_DIR}/pyproject.toml" "${ROOT_DIR}/server"
exit 0
