# Task 007: OpenAPI Specs and Frontend Type Generation

## Git Branch

`feature/007-openapi-and-fe-types-simplify`

## Objective

Update the declarative Swaggo annotations across backend HTTP handlers to document the simplified
session, shot, and live stats API surfaces. Regenerate the Swagger 2.0 specification, convert and
enrich it to OpenAPI 3.0 using project automation scripts, and generate updated TypeScript type
definitions for the Vue 3 frontend in `frontend/src/types/types.generated.ts`.

## Dependencies

- Task 005 (HTTP handlers, route refactoring, and websocket teardown).

## Acceptance Criteria

- [ ] Swaggo annotations updated across backend handler files:
    - [ ] `backend/internal/handler/session.go`:
        - `POST /api/v0/session`: documents `model.SessionCreate` and `model.SessionId`
        - `GET /api/v0/session/open`: documents return of active `model.SessionRead` or null
        - `GET /api/v0/session/{id}`: documents `model.SessionRead`
        - `GET /api/v0/session`: documents `[]model.SessionRead`
        - `PATCH /api/v0/session/{id}/close`: documents `model.SessionClose`
        - `DELETE /api/v0/session/{id}`: documents 204 No Content
        - `GET /api/v0/session/{id}/live-stats`: documents `model.SessionLiveStatsRead`
        - Obsolete session endpoints removed from annotations
    - [ ] `backend/internal/handler/shot.go`:
        - `POST /api/v0/shot`: documents single or array `model.ShotCreate` with `session_id`
        - `GET /api/v0/shot/by-session/{session_id}`: documents `[]model.ShotRead`
        - `GET /api/v0/shot/count-by-session/{session_id}`: documents integer count
        - `DELETE /api/v0/shot/{shot_id}`: documents 204 No Content
        - Obsolete slot-based shot annotations removed
    - [ ] Obsolete Swaggo doc annotations for `slot`, `stats/ws`, and `/auth/register`
          completely purged
- [ ] Automated spec and type generation succeeds:
    - [ ] `backend/specs/swagger.json` and `swagger.yaml` regenerated via `swag init`
    - [ ] `openapi.json` generated and enriched via `scripts/enrich_openapi.py`
    - [ ] `frontend/src/types/types.generated.ts` generated via `scripts/generate_fe_types.bash`
- [ ] Generated TypeScript interfaces verified:
    - [ ] `SessionCreate` contains `bow_id`, `distance`, `face_type`, `shots_per_end`, etc.
    - [ ] `SessionRead` contains all canonical session properties including `status: SessionStatus`
    - [ ] `SessionClose` contains mandatory `status: SessionStatus` (`'bad' | 'neutral' | 'good'`),
          `was_goal_achieved`, `did_well`, `need_work`
    - [ ] `SessionStatus` enum generated (`'not_rated' | 'bad' | 'neutral' | 'good'`)
    - [ ] `ShotCreate` and `ShotRead` reference `session_id` (not `slot_id`)
    - [ ] `SessionLiveStatsRead` contains `stats` and `scores`
    - [ ] No obsolete `Slot`, `WebSocketMessage`, or `AuthRegistrationRequest` schemas in
          generated types
- [ ] `cd frontend && npm run type-check` compiles without syntax errors in generated types.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/handler/session.go` |
| Modify | `backend/internal/handler/shot.go` |
| Modify | `backend/specs/docs.go` |
| Modify | `backend/specs/swagger.json` |
| Modify | `backend/specs/swagger.yaml` |
| Modify | `openapi.json` |
| Modify | `frontend/src/types/types.generated.ts` |

## Reference

- [PRD.md](../PRD.md)
- [generate_fe_types.bash](../../../scripts/generate_fe_types.bash)
- [enrich_openapi.py](../../../scripts/enrich_openapi.py)

## Steps

- [ ] **Step 1: Update Swaggo annotations in `handler/session.go`**

  Annotate all session lifecycle routes and the new `live-stats` endpoint. Add tags, summary,
  parameters, responses, and security declarations.

- [ ] **Step 2: Update Swaggo annotations in `handler/shot.go`**

  Annotate `/api/v0/shot` (POST single/batch), `/by-session/{session_id}` (GET),
  `/count-by-session/{session_id}` (GET), and `/{shot_id}` (DELETE).

- [ ] **Step 3: Run backend Swaggo regeneration**

  ```bash
  cd backend
  swag init -g cmd/arch-stats/main.go -o specs/ --parseInternal --useStructName --requiredByDefault
  ```

- [ ] **Step 4: Run frontend type generation script**

  ```bash
  ./scripts/generate_fe_types.bash
  ```

- [ ] **Step 5: Verify generated TypeScript types**

  Inspect `frontend/src/types/types.generated.ts`:
    - Confirm `SessionCreate` includes `bow_id: string`, `distance: number`, etc.
    - Confirm `ShotCreate` includes `session_id: string` and optional `arrow_id?: string | null`.
    - Confirm obsolete slot interfaces are removed.
    - Confirm `/auth/register` and `AuthRegistrationRequest` are completely absent.

- [ ] **Step 6: Commit changes**

  ```bash
  git add backend/specs/ openapi.json frontend/src/types/types.generated.ts
  git commit -m "chore(api): regenerate OpenAPI specs and frontend TypeScript types"
  ```

## Verification

- `./scripts/generate_fe_types.bash` completes with exit code 0.
- `frontend/src/types/types.generated.ts` exports updated schemas matching the PRD specification.
