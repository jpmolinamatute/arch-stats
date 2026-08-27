# Task 030: Update Docker Compose — Remove Flyway, Add goose

## Git Branch

`refactor/030-docker-compose-update`

## Objective

Update the Docker Compose configuration to remove the Flyway migration container and replace it
with a goose-based approach. The Go binary can run migrations on startup, so a separate migration
container may not be needed. Update the PostgreSQL service to version 17.

## Dependencies

- Task 005 (goose migration tooling integrated into Go binary)
- Task 025 (Go binary runs migrations on startup when configured)

## Acceptance Criteria

- [ ] `docker/docker-compose.yaml` changes:
    - PostgreSQL service updated from `postgres:15` to `postgres:17`
    - Flyway `migrations` service removed entirely
    - Migration volume mount (`../backend/migrations:/flyway/sql`) removed
    - All Flyway-related environment variables removed
    - A comment documents that migrations run via the Go binary on startup
- [ ] `docker/docker-compose.yaml` retains:
    - PostgreSQL `db` service with health check
    - Emulator service (unchanged)
    - Network and volume configuration
- [ ] `docker compose -f docker/docker-compose.yaml config` validates without errors.
- [ ] `docker compose -f docker/docker-compose.yaml --profile dev up -d` starts PostgreSQL
    successfully.

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `docker/docker-compose.yaml` |
| Modify | `docker/.env` (if Flyway-specific vars exist) |

## Reference

- Current config: [docker-compose.yaml](file:///home/juanpa/Projects/arch-stats/docker/docker-compose.yaml)
- Flyway service to remove: lines 37-56

## Steps

- [ ] **Step 1: Remove the Flyway migrations service**

  Delete the entire `migrations:` service block from `docker/docker-compose.yaml`.

- [ ] **Step 2: Update PostgreSQL version**

  Change `image: postgres:15` to `image: postgres:17`.

- [ ] **Step 3: Add a comment about migration strategy**

  ```yaml
  # Migrations are handled by the Go binary via embedded goose.
  # Run: ./arch-stats migrate
  # Or set APPLY_DB_MIGRATIONS_ON_START=true for auto-migration on startup.
  ```

- [ ] **Step 4: Clean up `.env` if needed**

  Remove any Flyway-specific environment variables from `docker/.env`.

- [ ] **Step 5: Validate the compose file**

  ```bash
  docker compose -f docker/docker-compose.yaml config
  ```

  Expected: valid YAML output, no errors.

- [ ] **Step 6: Test starting the database**

  ```bash
  docker compose -f docker/docker-compose.yaml --profile dev up -d
  docker compose -f docker/docker-compose.yaml --profile dev ps
  ```

  Expected: `db` service is healthy.

- [ ] **Step 7: Stop and commit**

  ```bash
  docker compose -f docker/docker-compose.yaml --profile dev down
  git add -A
  git commit -m "chore: remove Flyway from docker-compose, update to PostgreSQL 17"
  ```

## Verification

- `docker compose -f docker/docker-compose.yaml config` — validates cleanly.
- `grep -i flyway docker/docker-compose.yaml` — returns no results.
- `docker compose -f docker/docker-compose.yaml --profile dev up -d` — PostgreSQL starts.
