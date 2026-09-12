# Task 006: Backend Integration Tests for Onboarding Flow

## Git Branch

`feature/006-backend-integration-tests`

## Objective

Build comprehensive Go integration tests verifying the full onboarding lifecycle against a live
PostgreSQL test database: Google OAuth initial sign-in, archer personal profile registration,
mandatory bow creation, optional arrow registration, and status transitions from
`needs_registration` to `authenticated`.

## Dependencies

- Task 001 (Database migrations and schema)
- Task 005 (HTTP handlers and Chi router wiring)

## Acceptance Criteria

- [ ] New test file `backend/tests/integration/onboarding_test.go` implements:
    - [ ] Test 1: First-time Google OAuth sign-in creates an `auth` identity record and returns
          `status: needs_registration`.
    - [ ] Test 2: `POST /api/v0/archers` creates personal profile; `GET /api/v0/auth/status` still
          returns `needs_registration` because 0 bows are registered.
    - [ ] Test 3: `POST /api/v0/bows` creates first bow; `GET /api/v0/auth/status` transitions to
          `authenticated`.
    - [ ] Test 4: Subsequent login with same Google credential directly returns `authenticated`
          with the created archer and bow data.
- [ ] New test file `backend/tests/integration/arrow_registration_test.go` implements:
    - [ ] Test 5: Optional arrow registration during onboarding accepts valid batches (>= 3
          arrows), automatically marks them with `status == "in_use"`, and returns HTTP 201.
    - [ ] Test 6: Arrow batches with fewer than 3 arrows are rejected with HTTP 422 validation
          error.
    - [ ] Test 7: Arrow registration can be skipped entirely without blocking `authenticated`
          status.
- [ ] New test file `backend/tests/integration/bow_equipment_test.go` implements:
    - [ ] Test 8: Multiple bows can be registered by the same archer.
    - [ ] Test 9: `GET /api/v0/bows` returns all registered bows ordered by creation.
- [ ] All integration tests pass cleanly with race detection enabled:
      `go test -race ./tests/integration/... -v`.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/onboarding_test.go` |
| Create | `backend/tests/integration/arrow_registration_test.go` |
| Create | `backend/tests/integration/bow_equipment_test.go` |

## Reference

- [PRD.md](../PRD.md)
- [helpers_test.go](../../../backend/tests/integration/helpers_test.go)
- [story_time.md](../../../backend/migrations/story_time.md)

## Steps

- [ ] **Step 1: Implement `onboarding_test.go`**

  Create test suite executing the multi-step flow using `httptest.Server` or direct router calls:

  ```go
  func TestOnboardingFlow_CompleteLifecycle(t *testing.T) {
      ts := setupIntegrationServer(t)
      defer ts.Close()

      // Step 1: Initial OAuth login
      loginResp := ts.loginWithGoogle(t, "test-google-sub-123")
      assert.Equal(t, model.AuthStatusNeedsRegistration, loginResp.Status)

      // Step 2: Submit personal profile
      archerResp := ts.createArcher(t, loginResp.ArcherID, model.ArcherCreate{
          Email:       "archer@example.com",
          FirstName:   "Robin",
          LastName:    "Hood",
          DateOfBirth: "1995-05-15",
          Gender:      model.GenderMale,
      })
      assert.Equal(t, http.StatusCreated, archerResp.StatusCode)

      // Status check should still be needs_registration (missing bow)
      statusResp := ts.getAuthStatus(t, loginResp.AccessToken)
      assert.Equal(t, model.AuthStatusNeedsRegistration, statusResp.Status)

      // Step 3: Register first bow
      bowResp := ts.createBow(t, loginResp.AccessToken, model.BowCreate{
          Name:       "Sherwood Recurve",
          Bowstyle:   model.BowstyleRecurve,
          DrawWeight: 36.0,
      })
      assert.Equal(t, http.StatusCreated, bowResp.StatusCode)

      // Status check must now transition to authenticated
      statusResp = ts.getAuthStatus(t, loginResp.AccessToken)
      assert.Equal(t, model.AuthStatusAuthenticated, statusResp.Status)
  }
  ```

- [ ] **Step 2: Implement `arrow_registration_test.go`**

  Write tests verifying arrow set validation (minimum 3 arrows per set):

  ```go
  func TestArrowRegistration_BatchMinimum(t *testing.T) {
      ts := setupIntegrationServer(t)
      defer ts.Close()
      token := ts.authenticateUser(t)

      // Batch with 2 arrows must fail with 422
      resp := ts.createArrowBatch(t, token, model.ArrowBatchCreate{
          ArrowSet: 1,
          Count:    2,
      })
      assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

      // Batch with 6 arrows must succeed with 201
      resp = ts.createArrowBatch(t, token, model.ArrowBatchCreate{
          ArrowSet: 1,
          Count:    6,
      })
      assert.Equal(t, http.StatusCreated, resp.StatusCode)
  }
  ```

- [ ] **Step 3: Implement `bow_equipment_test.go`**

  Write tests for multiple bow registration and listing.

- [ ] **Step 4: Run integration test suite**

  ```bash
  cd backend
  go test -race ./tests/integration/... -v
  ```

- [ ] **Step 5: Commit changes**

  ```bash
  git add backend/tests/integration
  git commit -m "test(integration): add full onboarding lifecycle and equipment test suite"
  ```

## Verification

- `cd backend && go test -race ./tests/integration/... -v` runs all tests with 0 failures.
- Zero data race warnings or database deadlocks reported.
