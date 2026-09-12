# Task 005: HTTP Handlers and Chi Router Wiring

## Git Branch

`feature/005-http-handlers-and-routing`

## Objective

Implement HTTP handlers and register endpoints in the Chi router for bow registration, arrow
registration, personal archer profile creation, and authentication status verification.

## Dependencies

- Task 002 (Domain model structs)
- Task 004 (Service layer and onboarding business logic)

## Acceptance Criteria

- [ ] New handler `backend/internal/handler/bow.go` implements:
    - [ ] `POST /api/v0/bows`: registers a bow for the authenticated archer (HTTP 201).
    - [ ] `GET /api/v0/bows`: returns all bows owned by the authenticated archer (HTTP 200).
- [ ] New handler `backend/internal/handler/arrow.go` implements:
    - [ ] `POST /api/v0/arrows`: registers arrows for the authenticated archer, supporting both
          single arrow and batch arrow set creation (HTTP 201; `status` defaults to `in_use` if
          omitted).
    - [ ] `GET /api/v0/arrows`: returns all arrows owned by the authenticated archer including
          `status` (HTTP 200).
- [ ] Modified `backend/internal/handler/archer.go`:
    - [ ] `POST /api/v0/archers`: creates initial archer profile during onboarding step 3.
    - [ ] `GET /api/v0/archers/{archer_id}`: retrieves profile by ID.
    - [ ] Validation tags and error responses match updated `model.ArcherCreate`.
- [ ] Modified `backend/internal/handler/auth.go`:
    - [ ] `GET /api/v0/auth/status`: returns `authenticated` or `needs_registration`.
    - [ ] `POST /api/v0/auth/login` & `/api/v0/auth/google`: returns authenticated payload or
          registration requirements with HTTP 200.
- [ ] Chi router in `backend/cmd/arch-stats/main.go` registers all new routes under `/api/v0`
      with appropriate auth middleware.
- [ ] Unit tests for all handlers verifying status codes, request body parsing, validation errors,
      and response formatting.
- [ ] `cd backend && go test ./internal/handler/... -v` passes.
- [ ] `cd backend && golangci-lint run ./internal/handler/...` reports zero issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/bow.go` |
| Create | `backend/internal/handler/bow_test.go` |
| Create | `backend/internal/handler/arrow.go` |
| Create | `backend/internal/handler/arrow_test.go` |
| Modify | `backend/internal/handler/archer.go` |
| Modify | `backend/internal/handler/archer_test.go` |
| Modify | `backend/internal/handler/auth.go` |
| Modify | `backend/internal/handler/auth_test.go` |
| Modify | `backend/cmd/arch-stats/main.go` |

## Reference

- [PRD.md](../PRD.md)
- [main.go](../../../backend/cmd/arch-stats/main.go)
- [auth.go](../../../backend/internal/handler/auth.go)
- [archer.go](../../../backend/internal/handler/archer.go)

## Steps

- [ ] **Step 1: Write failing handler tests for Bow and Arrow endpoints**

  Create `bow_test.go` and `arrow_test.go` verifying:
    - `POST /api/v0/bows` returns 201 and created `BowRead` JSON on valid input.
    - `POST /api/v0/bows` returns 422 on invalid draw weight or missing bowstyle.
    - `POST /api/v0/arrows` returns 201 on valid batch registration.
    - `POST /api/v0/arrows` returns 422 when batch count is less than 3.
    - `GET /api/v0/bows` and `GET /api/v0/arrows` return 200 with JSON arrays.

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Implement `bow.go` handler**

  Implement `BowHandler`:

  ```go
  type BowHandler struct {
      bows BowService
  }

  func (h *BowHandler) Create(w http.ResponseWriter, r *http.Request) {
      archerID, err := middleware.GetArcherID(r.Context())
      if err != nil {
          WriteAppError(w, apperror.ErrUnauthorized)
          return
      }
      var req model.BowCreate
      if err := ReadJSON(r, &req); err != nil {
          WriteAppError(w, err)
          return
      }
      req.ArcherID = archerID
      bow, err := h.bows.CreateBow(r.Context(), req)
      if err != nil {
          WriteAppError(w, err)
          return
      }
      _ = WriteJSON(w, http.StatusCreated, bow)
  }
  ```

  Implement `List` endpoint.

- [ ] **Step 4: Implement `arrow.go` handler**

  Implement `ArrowHandler` with `Create` and `List` methods, extracting `archerID` from context.

- [ ] **Step 5: Update `archer.go` and `auth.go` handlers**

  Update `CreateArcher` payload reading to match new personal-only schema. Ensure `GET /auth/status`
  inspects profile completeness and returns current status.

- [ ] **Step 6: Wire dependencies and routes in `cmd/arch-stats/main.go`**

  Instantiate `BowRepo`, `ArrowRepo`, `AuthIdentityRepo`, `BowService`, `ArrowService`,
  `BowHandler`, and `ArrowHandler`. Mount routes:

  ```go
  r.Route("/api/v0", func(r chi.Router) {
      r.Group(func(r chi.Router) {
          r.Use(authMiddleware)
          r.Post("/archers", archerHandler.Create)
          r.Get("/archers/{archer_id}", archerHandler.GetByID)
          r.Post("/bows", bowHandler.Create)
          r.Get("/bows", bowHandler.List)
          r.Post("/arrows", arrowHandler.Create)
          r.Get("/arrows", arrowHandler.List)
      })
  })
  ```

- [ ] **Step 7: Run handler tests and linter**

  ```bash
  cd backend
  go test ./internal/handler/... -v
  golangci-lint run ./internal/handler/...
  ```

- [ ] **Step 8: Commit changes**

  ```bash
  git add backend/internal/handler backend/cmd/arch-stats
  git commit -m "feat(handler): add bow and arrow endpoints and wire routes in main"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` exits with code 0.
- `cd backend && golangci-lint run ./...` reports no issues.
- `cd backend && go build ./...` compiles cleanly without errors.
