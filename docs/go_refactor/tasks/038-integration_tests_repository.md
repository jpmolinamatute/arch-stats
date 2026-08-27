# Task 038: Integration Tests — Repository Layer

## Git Branch

`refactor/038-integration-tests-repository`

## Objective

Write integration tests for all repository implementations against a real PostgreSQL database
using the testcontainers infrastructure from Task 037. These tests verify that the squirrel-built
queries actually execute correctly against PostgreSQL, including edge cases like unique constraints,
FK violations, and NULL handling.

## Dependencies

- Task 037 (integration test infrastructure)
- Tasks 008–013 (all repository implementations)

## Acceptance Criteria

- [ ] `backend/tests/integration/repo_archer_test.go` tests:
    - Create an archer → verify it can be found by ID
    - Create an archer → verify it can be found by email
    - Create an archer → verify it can be found by Google subject
    - Create duplicate email → verify conflict/unique constraint error
    - Update an archer → verify changes are persisted
    - FindByID with non-existent ID → verify nil/no rows
    - FindAll with filters → verify correct filtering
- [ ] `backend/tests/integration/repo_auth_session_test.go` tests:
    - Create an auth session → verify it can be found by token hash
    - DeleteByArcherID → verify all sessions for that archer are removed
    - DeleteExpired → verify only expired sessions are removed
- [ ] `backend/tests/integration/repo_session_test.go` tests:
    - Create a shooting session → verify FindByID returns it
    - FindOpen → verify it returns the session with status "open"
    - Close a session → verify status changes and ended_at is set
    - Create multiple sessions → verify FindAll returns correct list
- [ ] `backend/tests/integration/repo_slot_test.go` tests:
    - Create a slot → verify FindBySessionID returns it
    - Create multiple slots → verify ordering by slot_number
    - CountBySessionID → verify correct count
    - Delete a slot → verify it's removed
- [ ] `backend/tests/integration/repo_shot_test.go` tests:
    - Create a shot → verify FindBySlotID returns it
    - Create multiple shots → verify ordering by arrow_number
    - Update a shot score → verify change persisted
    - Delete a shot → verify it's removed
- [ ] `backend/tests/integration/repo_face_target_test.go` tests:
    - FindAll faces → verify faces are returned
    - FindByType → verify filtering works
    - Create a target → verify FindBySlotID returns it
- [ ] Each test truncates tables after completion (using `t.Cleanup`).
- [ ] `go test ./tests/integration/... -v -count=1` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/repo_archer_test.go` |
| Create | `backend/tests/integration/repo_auth_session_test.go` |
| Create | `backend/tests/integration/repo_session_test.go` |
| Create | `backend/tests/integration/repo_slot_test.go` |
| Create | `backend/tests/integration/repo_shot_test.go` |
| Create | `backend/tests/integration/repo_face_target_test.go` |

## Reference

- Python model tests were in `backend-old/tests/models/` (currently empty — `core/__init__.py` only)
- Python endpoint tests provide behavioral expectations:
  [test_archer_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_archer_endpoints.py),
  [test_session_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_session_endpoints.py)

## Steps

- [ ] **Step 1: Write `repo_archer_test.go`**

  Each test function follows this pattern:

  ```go
  func TestArcherRepo_CreateAndFindByID(t *testing.T) {
      ctx := context.Background()
      t.Cleanup(func() { truncateAll(ctx, testPool) })

      repo := repository.NewArcherRepo(testPool)

      // Create
      id, err := repo.Create(ctx, model.ArcherCreate{
          Email:          "test@example.com",
          FirstName:      "Test",
          LastName:       "Archer",
          GoogleSubject:  "google-sub-123",
          // ... other required fields
      })
      if err != nil {
          t.Fatalf("Create() error: %v", err)
      }

      // Find
      archer, err := repo.FindByID(ctx, id)
      if err != nil {
          t.Fatalf("FindByID() error: %v", err)
      }
      if archer.Email != "test@example.com" {
          t.Errorf("Email = %q, want %q", archer.Email, "test@example.com")
      }
  }
  ```

- [ ] **Step 2: Write `repo_auth_session_test.go`**

  Test create → find by token hash → delete by archer ID → verify gone.

- [ ] **Step 3: Write `repo_session_test.go`**

  Test create → find open → close → verify status/ended_at.

- [ ] **Step 4: Write `repo_slot_test.go`**

  Test create → find by session ID → count → delete. Requires a session to exist first (FK).

- [ ] **Step 5: Write `repo_shot_test.go`**

  Test create → find by slot ID → update score → delete. Requires a slot to exist first (FK).

- [ ] **Step 6: Write `repo_face_target_test.go`**

  Test find all faces → create target → find by slot ID.

- [ ] **Step 7: Run all integration tests**

  ```bash
  cd backend
  go test ./tests/integration/... -v -count=1
  ```

  Expected: all tests pass.

- [ ] **Step 8: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 9: Commit**

  ```bash
  git add -A
  git commit -m "test: add repository integration tests with real PostgreSQL"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1` — all tests pass.
- `cd backend && go vet ./...` — clean.
- Each test properly cleans up (truncates tables) via `t.Cleanup`.
