---
name: manage-infra
description: Starts or stops the Docker Compose infrastructure (Postgres 17).
---

# Skill: Manage Infrastructure

## Actions

- **To Start:** Run the VS Code task: `"Start Docker Compose"`.
    - This builds and starts containers in detached mode using `./docker/docker-compose.yaml`.
- **To Stop:** Run the VS Code task: `"Stop Docker Compose"`.
- **To Reset (Wipe Data):** Run the VS Code task: `"Stop & Remove Volumes"`.

## Verification

- Run `docker ps` to ensure the database container is healthy before starting the Backend
  or Frontend servers.
