# Task 029: Add swaggo Annotations + OpenAPI Spec Generation

## Git Branch

`refactor/029-openapi-swaggo-annotations`

## Objective

Add `swaggo/swag` annotations to all Go handler functions and configure `swag init` to generate
an OpenAPI spec (`specs/swagger.json`). The frontend's `npm run generate:types` pipeline
consumes this spec to produce TypeScript types, so the generated spec must match the current
API contract.

## Dependencies

- Task 019–024, 027 (all handlers — annotations go on handler functions)
- Task 025 (chi router — swag needs the router for endpoint discovery)

## Acceptance Criteria

- [ ] Every handler function has `swaggo/swag` annotations including:
    - `@Summary`, `@Description`
    - `@Tags` (matching the Python router tags: Archers, Sessions, Slots, Shots, Faces, Stats, Auth)
    - `@Accept json` / `@Produce json`
    - `@Param` for path params, query params, and request bodies
    - `@Success` and `@Failure` with response types
    - `@Security` for authenticated endpoints
    - `@Router` with path and HTTP method
- [ ] `main.go` has top-level swag annotations:
    - `@title Arch Stats API`
    - `@version v0`
    - `@BasePath /api/v0`
    - `@securityDefinitions.apikey BearerAuth`
- [ ] Running `swag init` generates `backend/specs/swagger.json` and `backend/specs/swagger.yaml`.
- [ ] The generated spec can be consumed by `npm run generate:types` in the frontend to produce
  valid TypeScript types.
- [ ] `go build ./cmd/arch-stats` compiles cleanly.
- [ ] `go vet ./...` reports no issues.

## Files to Modify

| Action | Path |
| ------ | ---- |
| Modify | `backend/cmd/arch-stats/main.go` (top-level swag annotations) |
| Modify | `backend/internal/handler/auth.go` (add swag annotations) |
| Modify | `backend/internal/handler/archer.go` (add swag annotations) |
| Modify | `backend/internal/handler/session.go` (add swag annotations) |
| Modify | `backend/internal/handler/slot.go` (add swag annotations) |
| Modify | `backend/internal/handler/shot.go` (add swag annotations) |
| Modify | `backend/internal/handler/face.go` (add swag annotations) |
| Modify | `backend/internal/handler/live_stats.go` (add swag annotations) |
| Create | `backend/specs/` directory |

## Steps

- [ ] **Step 1: Install swag CLI**

  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

- [ ] **Step 2: Add top-level annotations to `main.go`**

  Add before the `main()` function:

  ```go
  // @title       Arch Stats API
  // @version     v0
  // @description Backend API for archery statistics tracking
  // @BasePath    /api/v0

  // @securityDefinitions.apikey BearerAuth
  // @in   cookie
  // @name access_token
  ```

- [ ] **Step 3: Annotate all handler functions**

  Example for `GetArcher`:

  ```go
  // GetByID godoc
  // @Summary     Get archer by ID
  // @Description Get a single archer by their UUID
  // @Tags        Archers
  // @Accept      json
  // @Produce     json
  // @Param       id path string true "Archer UUID"
  // @Success     200 {object} model.ArcherRead
  // @Failure     404 {object} handler.ErrorResponse
  // @Failure     422 {object} handler.ErrorResponse
  // @Security    BearerAuth
  // @Router      /archer/{id} [get]
  ```

  Repeat for all handler functions across all 7 handler files.

- [ ] **Step 4: Run `swag init`**

  ```bash
  cd backend
  swag init -g cmd/arch-stats/main.go -o specs/
  ```

  Expected: generates `specs/swagger.json`, `specs/swagger.yaml`, `specs/docs.go`.

- [ ] **Step 5: Verify the spec is valid**

  ```bash
  cat backend/specs/swagger.json | python3 -m json.tool > /dev/null
  echo "JSON is valid"
  ```

- [ ] **Step 6: Test frontend type generation**

  ```bash
  cp backend/specs/swagger.json openapi.json
  cd frontend
  npm run generate:types
  ```

  Expected: TypeScript types are generated without errors.

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./cmd/arch-stats
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "feat: add swaggo annotations and OpenAPI spec generation"
  ```

## Verification

- `cd backend && swag init -g cmd/arch-stats/main.go -o specs/` — generates spec without errors.
- `backend/specs/swagger.json` is valid JSON with all endpoints documented.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./cmd/arch-stats` — compiles.
- `cd frontend && npm run generate:types` — produces TypeScript types from the new spec.
