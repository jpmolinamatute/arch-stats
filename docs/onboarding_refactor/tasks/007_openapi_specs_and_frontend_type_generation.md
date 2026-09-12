# Task 007: OpenAPI Swaggo Annotations and Frontend Type Generation

## Git Branch

`feature/007-openapi-specs-typegen`

## Objective

Add and update Swaggo annotations across all new and modified HTTP handlers, regenerate OpenAPI 2.0
/ 3.0 specification documents, and run the automated type generator to produce synchronized
TypeScript definitions for the frontend.

## Dependencies

- Task 005 (HTTP handlers and Chi router wiring)

## Acceptance Criteria

- [ ] Swaggo annotations added/updated in:
    - [ ] `backend/internal/handler/bow.go` (`POST /api/v0/bows`, `GET /api/v0/bows`)
    - [ ] `backend/internal/handler/arrow.go` (`POST /api/v0/arrows`, `GET /api/v0/arrows`)
    - [ ] `backend/internal/handler/archer.go`
          (`POST /api/v0/archers`, `GET /api/v0/archers/{archer_id}`)
    - [ ] `backend/internal/handler/auth.go` (`GET /api/v0/auth/status`, `POST /api/v0/auth/login`)
- [ ] Swaggo CLI (`swag init`) regenerates:
    - [ ] `backend/docs/swagger.json`
    - [ ] `backend/docs/swagger.yaml`
    - [ ] `backend/docs/docs.go`
- [ ] Running `./scripts/generate_fe_types.bash` updates `frontend/src/types/types.generated.ts`
      with:
    - [ ] `components['schemas']['BowRead']` and `BowCreate`
    - [ ] `components['schemas']['ArrowRead']` (including `status: ArrowStatus`) and `ArrowCreate`
    - [ ] `components['schemas']['ArrowStatus']` (`'in_use' | 'damaged' | 'lost'`)
    - [ ] Updated `components['schemas']['ArcherRead']` (without legacy fields)
    - [ ] Updated `components['schemas']['ArcherCreate']` (strictly personal fields)
- [ ] `cd frontend && npx vue-tsc -b` runs type checking and compiles cleanly.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/internal/handler/bow.go` |
| Modify | `backend/internal/handler/arrow.go` |
| Modify | `backend/internal/handler/archer.go` |
| Modify | `backend/internal/handler/auth.go` |
| Modify | `backend/docs/swagger.json` |
| Modify | `backend/docs/swagger.yaml` |
| Modify | `backend/docs/docs.go` |
| Generate | `frontend/src/types/types.generated.ts` (gitignored; do NOT commit) |

> [!NOTE]
> `frontend/src/types/types.generated.ts` is gitignored and must **NOT** be committed. It must
> be generated locally using `./scripts/generate_fe_types.bash` (or `npm run generate:types`)
> to ensure frontend types remain synchronized with the backend OpenAPI schema.

## Reference

- [generate_fe_types.bash](../../../scripts/generate_fe_types.bash)
- [types.generated.ts][types-gen]
- [PRD.md](../../../docs/onboarding_refactor/PRD.md)

[types-gen]: ../../../frontend/src/types/types.generated.ts

## Steps

- [ ] **Step 1: Add Swaggo annotations to `bow.go`**

  Annotate `Create` and `List` methods:

  ```go
  // Create godoc
  // @Summary     Register Bow
  // @Description Register a new bow for the authenticated archer
  // @Tags        Equipment
  // @Accept      json
  // @Produce     json
  // @Param       request body model.BowCreate true "Bow details"
  // @Success     201 {object} model.BowRead
  // @Failure     401 {object} model.ErrorResponse
  // @Failure     422 {object} model.ErrorResponse
  // @Router      /bows [post]
  ```

- [ ] **Step 2: Add Swaggo annotations to `arrow.go`**

  Annotate `Create` and `List` methods for arrows:

  ```go
  // Create godoc
  // @Summary     Register Arrows
  // @Description Register one or more arrows for the authenticated archer
  // @Tags        Equipment
  // @Accept      json
  // @Produce     json
  // @Param       request body model.ArrowBatchCreate true "Arrow set details"
  // @Success     201 {array} model.ArrowRead
  // @Failure     401 {object} model.ErrorResponse
  // @Failure     422 {object} model.ErrorResponse
  // @Router      /arrows [post]
  ```

- [ ] **Step 3: Update annotations on `archer.go` and `auth.go`**

  Update request and response schemas to reflect the personal-only profile and new auth status.

- [ ] **Step 4: Regenerate OpenAPI specification files**

  ```bash
  cd backend
  swag init -g cmd/arch-stats/main.go -o docs --parseDependency --parseInternal
  ```

- [ ] **Step 5: Run frontend type generation script**

  ```bash
  ./scripts/generate_fe_types.bash
  ```

  Verify `frontend/src/types/types.generated.ts` contains the new schema interfaces.

- [ ] **Step 6: Run frontend type verification**

  ```bash
  cd frontend
  npx vue-tsc -b
  ```

- [ ] **Step 7: Commit changes**

  Commit OpenAPI specs (frontend generated types are gitignored and not committed):

  ```bash
  git add backend/docs
  git commit -m "chore(api): update OpenAPI specs for onboarding and equipment endpoints"
  ```

## Verification

- `swag init` completes without errors or parse warnings.
- `frontend/src/types/types.generated.ts` diff shows new `BowRead`, `BowCreate`, `ArrowRead`, and
  `ArrowCreate` types.
- `cd frontend && npx vue-tsc -b` passes with zero type errors.
