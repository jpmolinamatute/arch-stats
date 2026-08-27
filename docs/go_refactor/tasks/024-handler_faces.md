# Task 024: Build HTTP Handler — Faces Endpoints

## Git Branch

`refactor/024-handler-faces`

## Objective

Implement the faces HTTP handler in `internal/handler/`, porting the Python `faces_router.py`.
This provides read-only endpoints for listing target face definitions and getting a specific
face by type. These endpoints do NOT require authentication (public data).

## Dependencies

- Task 016 (face service)
- Task 019 (handler helpers)

## Acceptance Criteria

- [ ] `backend/internal/handler/face.go` implements `FaceHandler` with methods:
    - `ListFaces(w, r)` — GET `/api/v0/faces` — returns list of face summaries (face_type + face_name)
    - `GetFace(w, r)` — GET `/api/v0/faces/{face_type}` — returns full face definition
- [ ] These endpoints do NOT go through auth middleware (public).
- [ ] GetFace with unknown face_type returns 404.
- [ ] Unit tests using `httptest` with mock service verify:
    - ListFaces returns 200 + JSON array of face summaries
    - GetFace with valid type returns 200 + full face JSON
    - GetFace with unknown type returns 404
- [ ] `go test ./internal/handler/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/handler/face.go` |
| Create | `backend/internal/handler/face_test.go` |

## Reference

- Python router: [faces_router.py](file:///home/juanpa/Projects/arch-stats/backend/src/routers/v0/faces_router.py)
- 2 endpoints, no auth required

## Steps

- [ ] **Step 1: Write failing tests**

  Create `backend/internal/handler/face_test.go`:
    - Define mock `faceService` interface
    - Test list and get by type

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 3: Implement `face.go`**

- [ ] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend && go test ./internal/handler/... -v
  ```

- [ ] **Step 5: Run go vet and build**

  ```bash
  cd backend && go vet ./... && go build ./...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add faces HTTP handler with list and get-by-type endpoints"
  ```

## Verification

- `cd backend && go test ./internal/handler/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
