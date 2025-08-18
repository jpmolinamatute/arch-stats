# Copilot Instructions * Procedures (CI/CD, Releases, Deployments)

**Audience:** DevOps engineers and maintainers

## CI pipeline (GitHub Actions)

- Run on push/PR to `main` and feature branches.
- Jobs:
  1) **Backend quality**: `uv` env setup; run Black, isort, mypy, Pylint, pytest.
  2) **Frontend quality**: install Node; run ESLint and Prettier checks; ensure build success.
  3) **Shell quality**: run `shellcheck` and `shfmt` on `scripts/*.bash`.
  4) **Artifact build**: `npm run build` emits FE assets into `backend/src/server/frontend/`.
- Cache Python and Node dependencies for speed.
- Fail fast; annotate errors where possible.

## Releases

- Use semver tags: `vMAJOR.MINOR.PATCH`.
- Release notes summarize:
  - API changes (breaking/behavioral)
  - DB migrations (if any)
  - Frontend UX changes
  - Operational notes (config flags, env vars)

## Deployments

- Target: Linux (Raspberry Pi 5) environment.
- **Docker Compose** for PostgreSQL; backend runs as a systemd service or container depending on environment.
- `.env` is symlinked into backend and frontend processes so configuration stays consistent.
- Frontend is served by the backend (FastAPI) in production from `backend/src/server/frontend/`.

## Guardrails for automation

- Never publish images or artifacts with unpinned base versions.
- Ensure DB migrations (if introduced) are idempotent; include rollback notes.
- Keep secrets in GitHub Actions Secrets or a proper secret manager; never in repo.

## Useful scripts & locations

- `docker/docker-compose.yaml`: local DB
- `tools/linting.sh`: convenience runner for Python linters
- `.vscode/tasks.json`: standardized local tasks for dev

## References

- Top-level `README.md` (project intro)
- `backend/README.md` (backend dev guide)
- `frontend/README.md` (frontend dev guide)
