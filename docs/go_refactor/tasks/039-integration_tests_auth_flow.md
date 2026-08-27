# Task 039: Integration Tests — Auth Flow

## Git Branch

`refactor/039-integration-tests-auth-flow`

## Objective

Write integration tests for the full authentication flow against a real PostgreSQL database.
These tests verify the end-to-end auth lifecycle: session creation, JWT generation, session
validation, and session cleanup — all hitting the real database to catch issues that unit tests
with mocks cannot.

## Dependencies

- Task 037 (integration test infrastructure)
- Task 038 (repository integration tests — test helpers for creating archers)
- Task 017 (auth service implementation)
- Task 009 (auth session repository)

## Acceptance Criteria

- [ ] `backend/tests/integration/auth_flow_test.go` tests the following scenarios:
    - **Session lifecycle**: Create archer → create auth session → find session by token hash →
    verify archer_id matches → delete session → verify it's gone
    - **JWT round-trip with DB**: Create archer → create auth session → build JWT with session
    ID → decode JWT → extract archer ID → verify it matches the created archer
    - **Multiple sessions**: Create archer → create 3 sessions → delete by archer ID →
    verify all 3 are gone
    - **Expired session cleanup**: Create archer → create session with past expiry → call
    DeleteExpired → verify session is removed
    - **Session token hash consistency**: Generate token → hash it → store hash in DB →
    regenerate hash from same raw token → find by hash → verify it matches
- [ ] Each test truncates tables after completion.
- [ ] `go test ./tests/integration/... -v -count=1 -run Auth` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/auth_flow_test.go` |

## Reference

- Python auth tests: [test_auth_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_auth_endpoints.py)
- Python security tests: [test_security_edge_cases.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_security_edge_cases.py)

## Steps

- [ ] **Step 1: Write auth session lifecycle test**

  ```go
  func TestAuthFlow_SessionLifecycle(t *testing.T) {
      ctx := context.Background()
      t.Cleanup(func() { truncateAll(ctx, testPool) })

      archerRepo := repository.NewArcherRepo(testPool)
      authRepo := repository.NewAuthSessionRepo(testPool)

      // Create an archer first
      archerID, err := archerRepo.Create(ctx, model.ArcherCreate{
          Email:         "auth-test@example.com",
          FirstName:     "Auth",
          LastName:      "Tester",
          GoogleSubject: "google-auth-test",
      })
      if err != nil {
          t.Fatalf("create archer: %v", err)
      }

      // Generate session token
      raw, err := auth.GenerateSessionToken(32)
      if err != nil {
          t.Fatalf("generate token: %v", err)
      }
      hash := auth.HashSessionToken(raw)

      // Store session
      err = authRepo.Create(ctx, model.AuthSessionCreate{
          ArcherID:         archerID,
          SessionTokenHash: hash,
          ExpiresAt:        time.Now().Add(time.Hour),
      })
      if err != nil {
          t.Fatalf("create session: %v", err)
      }

      // Find by hash
      session, err := authRepo.FindByTokenHash(ctx, hash)
      if err != nil {
          t.Fatalf("find by hash: %v", err)
      }
      if session.ArcherID != archerID {
          t.Errorf("ArcherID = %v, want %v", session.ArcherID, archerID)
      }

      // Delete and verify gone
      err = authRepo.DeleteByArcherID(ctx, archerID)
      if err != nil {
          t.Fatalf("delete by archer: %v", err)
      }
      session, err = authRepo.FindByTokenHash(ctx, hash)
      if err != nil && session != nil {
          t.Error("session should be gone after delete")
      }
  }
  ```

- [ ] **Step 2: Write JWT round-trip with DB test**

  Test that a JWT built from a real session can be decoded and the claims match the DB records.

- [ ] **Step 3: Write multiple sessions test**

  Test creating multiple sessions for one archer and bulk-deleting them.

- [ ] **Step 4: Write expired session cleanup test**

  Test that `DeleteExpired` only removes sessions past their expiry time.

- [ ] **Step 5: Write token hash consistency test**

  Test that hashing the same raw token bytes always produces the same hash, and that hash
  can be used to look up the session.

- [ ] **Step 6: Run integration tests**

  ```bash
  cd backend
  go test ./tests/integration/... -v -count=1 -run Auth
  ```

- [ ] **Step 7: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  git add -A
  git commit -m "test: add auth flow integration tests with real PostgreSQL"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1 -run Auth` — all tests pass.
- `cd backend && go vet ./...` — clean.
