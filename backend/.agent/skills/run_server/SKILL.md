---
name: run-server
description: Starts the FastAPI backend server using the dedicated VS Code task and bash script.
---

# Skill: Run Backend Server

This skill automates the execution of the FastAPI development server while ensuring all
environmental non-negotiables are met.

## Prerequisites

Before starting the server, you must verify the following:

1. **Docker Infrastructure:** Ensure the PostgreSQL 17 containers are running.
   - Command: `docker compose -f docker/docker-compose.yaml up -d`.
2. **Environment:** You must be using the Python 3.14 environment managed by `uv`.

## Execution Steps

To launch the server, use the integrated VS Code task runner:

1. **Trigger Task:** Execute the VS Code task labeled `"Start Uvicorn Dev Server"`.
2. **Manual Fallback:** If the task runner is unavailable, execute the canonical script
   directly from the project root:

   ```bash
   ./scripts/start_uvicorn.bash
   ```

## Verification

- Monitor the dedicated terminal panel for the Uvicorn startup sequence.
- Confirm the server is serving FastAPI on the expected local port.
- Data Flow Check: Ensure the server is successfully connecting to the asyncpg pool.

## Troubleshooting

- If the server fails to start, check if the database port is blocked or if Docker
  containers are unhealthy.
- Ensure no other instances of Uvicorn are running (pgrep -af uvicorn).
