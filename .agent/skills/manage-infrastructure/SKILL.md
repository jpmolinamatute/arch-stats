---
name: manage-infra
description: Starts or stops the Docker Compose infrastructure (Postgres 17).
---

# Skill: Manage Infrastructure

## Visual Studio Code Tasks

- **To Start:** Run the VS Code task: `"Start Docker Compose"`.
    - This builds and starts containers in detached mode using `./docker/docker-compose.yaml`.
- **To Stop:** Run the VS Code task: `"Stop Docker Compose"`.
- **To Reset (Wipe Data):** Run the VS Code task: `"Stop & Remove Volumes"`.

## Docker Compose

From the project root directory:

- **To Start:** Run the Docker Compose command:

   ```bash
   docker compose -f ./docker/docker-compose.yaml up -d
   ```

- **To Stop:** Run the Docker Compose command:

   ```bash
   docker compose -f ./docker/docker-compose.yaml down
   ```

- **To Reset (Wipe Data):** Run the Docker Compose command:

   ```bash
   docker compose -f ./docker/docker-compose.yaml down -v
   ```

## Verification

- Make sure the arch-stats container started successfully before starting the Backend or
   Frontend servers:

   ```bash
   docker compose -f ./docker/docker-compose.yaml ps --all
   ```

   This means:

    - arch-stats-db-1 is healthy
    - arch-stats-migrations-1 exited successfully (exit code 0)
