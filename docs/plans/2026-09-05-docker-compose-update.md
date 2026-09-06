# Task 030: Update Docker Compose Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update `docker/docker-compose.yaml` to upgrade the PostgreSQL service to version 17, eliminate the legacy Flyway `migrations` service, add a clear documentation comment on the Go binary / goose migration strategy, verify compose syntax and PostgreSQL runtime health, mark all tasks in `docs/go_refactor/tasks/030-docker_compose_update.md` and `docs/plans/task.md` as done, and commit changes to `refactor/030-docker-compose-update`.

**Architecture:**
- `docker/docker-compose.yaml`:
  - Upgrade `db` service image from `postgres:15` to `postgres:17`.
  - Remove legacy `migrations` container (previously running `flyway/flyway:latest` mounting `../backend-old/migrations`).
  - Add explanatory comment noting that migrations are now handled by the Go binary via embedded Goose (`./arch-stats migrate` or `APPLY_DB_MIGRATIONS_ON_START=true`).
  - Preserve `db` service configuration (custom `postgresql.conf`, `secondary.conf`, health check, volumes, network) and `emulator` service.
- `docker/.env`:
  - Verified to contain no leftover Flyway environment variables.
- `docs/go_refactor/tasks/030-docker_compose_update.md`:
  - Checked off completely upon successful verification.
- `docs/plans/task.md`:
  - Live table tracker updated per task and marked `DONE`.

**Tech Stack:** Docker Compose v5+, PostgreSQL 17, Goose migration engine (embedded in Go binary).

**Spec:**
- [docs/go_refactor/tasks/030-docker_compose_update.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/030-docker_compose_update.md)
- [docker/docker-compose.yaml](file:///home/juanpa/Projects/arch-stats/docker/docker-compose.yaml)
- [docker/.env](file:///home/juanpa/Projects/arch-stats/docker/.env)

## Global Constraints

- Git branch: `refactor/030-docker-compose-update` branched from latest `main`.
- PostgreSQL image must be updated to `postgres:17`.
- No traces of Flyway service, volume mounts, or configuration may remain in `docker/docker-compose.yaml`.
- All compose commands must validate without warnings or errors (`docker compose -f docker/docker-compose.yaml config`).
- Single-flow execution: exactly one active task tracked in `docs/plans/task.md`.
- Final step must mark all acceptance criteria and steps in `docs/go_refactor/tasks/030-docker_compose_update.md` as completed (`[x]`).

---

## File Structure

```
docker/
├── docker-compose.yaml                  # [MODIFY] Upgrade postgres:17, remove flyway migrations, add migration comment
└── .env                                 # [VERIFY] Verify no leftover Flyway environment variables
docs/
├── plans/
│   ├── task.md                          # [MODIFY] Track Task 030 live checklist progress (table-only)
│   └── 2026-09-05-docker-compose-update.md # [NEW] Implementation plan document
└── go_refactor/
    └── tasks/
        └── 030-docker_compose_update.md # [MODIFY] Mark all acceptance criteria and steps as checked [x]
```

---

## Task Structure

### Task 1: Git Branch Setup & Environment Validation

**Files:**
- Modify: `docs/plans/task.md`
- Inspect: `docker/.env`

**Interfaces:**
- Consumes: `main` branch HEAD
- Produces: `refactor/030-docker-compose-update` branch and clean `docs/plans/task.md` tracker

- [ ] **Step 1: Create and switch to the refactor git branch**

```bash
git checkout -b refactor/030-docker-compose-update
```

- [ ] **Step 2: Initialize `docs/plans/task.md` live tracker**

Write the initial status table for Task 030 to `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Environment Validation | IN_PROGRESS | Create branch `refactor/030-docker-compose-update` and inspect `docker/.env` |
| Task 2: Docker Compose Configuration Update | PENDING | Update postgres to 17, remove Flyway service, add migration comment, validate config |
| Task 3: Container Runtime & Health Verification | PENDING | Spin up postgres:17, verify container health check and SQL version, down container |
| Task 4: Mark Tasks as Done & Final Commit | PENDING | Mark `030-docker_compose_update.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Verify `docker/.env` has no Flyway variables**

Inspect `docker/.env` to ensure no `FLYWAY_*` variables exist:

```bash
grep -i "flyway" docker/.env || echo "No Flyway variables in docker/.env"
```

- [ ] **Step 4: Mark Task 1 as DONE in `docs/plans/task.md`**

Update `Task 1` status to `DONE` and `Task 2` status to `IN_PROGRESS`.

---

### Task 2: Docker Compose Configuration Update

**Files:**
- Modify: `docker/docker-compose.yaml:8-57`

**Interfaces:**
- Consumes: Existing `docker/docker-compose.yaml`
- Produces: Updated `docker/docker-compose.yaml` with PostgreSQL 17 and removed Flyway service

- [ ] **Step 1: Update `docker/docker-compose.yaml`**

In `docker/docker-compose.yaml`:
1. Change `image: postgres:15` to `image: postgres:17`.
2. Remove the entire `migrations:` service block:
```yaml
  migrations:
    image: flyway/flyway:latest
    profiles:
      - dev
    command:
      - migrate
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ../backend-old/migrations:/flyway/sql
    env_file:
      - ./.env
    environment:
      - FLYWAY_URL=jdbc:postgresql://db:5432/${POSTGRES_DB}
      - FLYWAY_USER=${POSTGRES_USER}
      - FLYWAY_PASSWORD=${POSTGRES_PASSWORD}
      - FLYWAY_LOCATIONS=filesystem:/flyway/sql
    networks:
      - arch-stats
```
3. Add explanatory comment regarding migrations:
```yaml
    # Migrations are handled by the Go binary via embedded goose.
    # Run: ./arch-stats migrate
    # Or set APPLY_DB_MIGRATIONS_ON_START=true for auto-migration on startup.
```

The resulting `docker/docker-compose.yaml` should be:

```yaml
name: arch-stats

networks:
  arch-stats:
    driver: bridge

services:
  db:
    image: postgres:17
    profiles:
      - dev
    env_file:
      - ./.env
    ports:
      - "${POSTGRES_PORT}:5432"
    networks:
      - arch-stats
    environment:
      - PGDATA=/var/lib/postgresql/data
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./pg_conf/postgresql.conf:/etc/postgresql/postgresql.conf
      - ./pg_conf/secondary.conf:/etc/postgresql/secondary.conf
    command:
      - postgres
      - -c
      - config_file=/etc/postgresql/postgresql.conf
    healthcheck:
      test:
        - CMD-SHELL
        - pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB -p 5432
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 5s
    # Migrations are handled by the Go binary via embedded goose.
    # Run: ./arch-stats migrate
    # Or set APPLY_DB_MIGRATIONS_ON_START=true for auto-migration on startup.

  emulator:
    build:
      context: ..
      dockerfile: docker/Dockerfile.rpi
    privileged: true
    profiles:
      - emulator
    ports:
      - "2222:22"
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:rw
    tmpfs:
      - /run
      - /run/lock
    cgroup: host
    networks:
      - arch-stats

volumes:
  pgdata:
```

- [ ] **Step 2: Validate compose file syntax**

Run:
```bash
docker compose -f docker/docker-compose.yaml config
```
Expected: Valid YAML output representing services `db` and `emulator`, with exit code 0.

- [ ] **Step 3: Confirm absence of Flyway references**

Run:
```bash
grep -i flyway docker/docker-compose.yaml
```
Expected: No output, exit code 1.

- [ ] **Step 4: Mark Task 2 as DONE in `docs/plans/task.md`**

Update `Task 2` status to `DONE` and `Task 3` status to `IN_PROGRESS`.

---

### Task 3: Container Runtime & Health Verification

**Files:**
- Test target: Docker daemon, `docker/docker-compose.yaml`

**Interfaces:**
- Consumes: Validated `docker/docker-compose.yaml`
- Produces: Verified healthy PostgreSQL 17 container instance

- [ ] **Step 1: Start PostgreSQL 17 in background**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile dev up -d
```
> [!NOTE]
> If existing volume `arch-stats_pgdata` was initialized with PostgreSQL 15, PostgreSQL 17 will report data directory version incompatibility. In that scenario, reset the dev volume via `docker compose -f docker/docker-compose.yaml --profile dev down -v` and re-run `docker compose -f docker/docker-compose.yaml --profile dev up -d`.

- [ ] **Step 2: Wait for healthy status and check `ps`**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile dev ps
```
Expected: `arch-stats-db-1` has status `Up ... (healthy)`.

- [ ] **Step 3: Verify PostgreSQL 17 version inside the container**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile dev exec db psql -U juanpa -d arch-stats -c "SELECT version();"
```
Expected: Output containing `PostgreSQL 17`.

- [ ] **Step 4: Stop containers cleanly**

Run:
```bash
docker compose -f docker/docker-compose.yaml --profile dev down
```
Expected: Containers stopped and removed cleanly.

- [ ] **Step 5: Mark Task 3 as DONE in `docs/plans/task.md`**

Update `Task 3` status to `DONE` and `Task 4` status to `IN_PROGRESS`.

---

### Task 4: Mark Tasks as Done & Final Commit

**Files:**
- Modify: `docs/go_refactor/tasks/030-docker_compose_update.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: Verified configuration and runtime
- Produces: Completed task documents and committed git branch

- [ ] **Step 1: Update `docs/go_refactor/tasks/030-docker_compose_update.md`**

Mark all acceptance criteria and all steps as completed `[x]`:
- `[x] docker/docker-compose.yaml changes:`
  - `[x] PostgreSQL service updated from postgres:15 to postgres:17`
  - `[x] Flyway migrations service removed entirely`
  - `[x] Migration volume mount (../backend/migrations:/flyway/sql) removed`
  - `[x] All Flyway-related environment variables removed`
  - `[x] A comment documents that migrations run via the Go binary on startup`
- `[x] docker/docker-compose.yaml retains:`
  - `[x] PostgreSQL db service with health check`
  - `[x] Emulator service (unchanged)`
  - `[x] Network and volume configuration`
- `[x] docker compose -f docker/docker-compose.yaml config validates without errors.`
- `[x] docker compose -f docker/docker-compose.yaml --profile dev up -d starts PostgreSQL successfully.`
- `[x] Step 1: Remove the Flyway migrations service`
- `[x] Step 2: Update PostgreSQL version`
- `[x] Step 3: Add a comment about migration strategy`
- `[x] Step 4: Clean up .env if needed`
- `[x] Step 5: Validate the compose file`
- `[x] Step 6: Test starting the database`
- `[x] Step 7: Stop and commit`

- [ ] **Step 2: Update `docs/plans/task.md` to reflect all tasks DONE**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Environment Validation | DONE | Create branch `refactor/030-docker-compose-update` and inspect `docker/.env` |
| Task 2: Docker Compose Configuration Update | DONE | Update postgres to 17, remove Flyway service, add migration comment, validate config |
| Task 3: Container Runtime & Health Verification | DONE | Spin up postgres:17, verify container health check and SQL version, down container |
| Task 4: Mark Tasks as Done & Final Commit | DONE | Mark `030-docker_compose_update.md` and `task.md` done, commit changes |
```

- [ ] **Step 3: Commit changes to git**

Run:
```bash
git add docker/docker-compose.yaml docs/go_refactor/tasks/030-docker_compose_update.md docs/plans/task.md docs/plans/2026-09-05-docker-compose-update.md
git commit -m "chore: remove Flyway from docker-compose, update to PostgreSQL 17"
```

- [ ] **Step 4: Verify working tree is clean**

Run:
```bash
git status
```
Expected: `nothing to commit, working tree clean`.

---

## Self-Review Checklist

1. **Spec Coverage:**
   - PostgreSQL updated to 17? Covered in Task 2 & Task 3.
   - Flyway migrations service removed? Covered in Task 2.
   - Migration volume mount removed? Covered in Task 2.
   - All Flyway environment variables removed? Covered in Task 1 & Task 2.
   - Comment documenting migration strategy added? Covered in Task 2.
   - Retains `db` health check, `emulator`, networks, volumes? Covered in Task 2.
   - `docker compose config` validates? Covered in Task 2.
   - `docker compose --profile dev up -d` starts successfully? Covered in Task 3.
   - Marking tasks as done at the end of implementation? Covered in Task 4.
2. **Placeholder Scan:**
   - No "TBD", "TODO", or unwritten code blocks exist in any step.
3. **Type Consistency:**
   - Compose YAML syntax, service names, and commands are consistent across all steps.
