# Task 040: Integration Tests — HTTP Handlers / Endpoints

## Git Branch

`refactor/040-integration-tests-http-handlers`

## Objective

Port the existing Python endpoint tests from `backend-old/tests/endpoints/` to Go integration
tests. These tests exercise the full HTTP stack (handler → service → repository → PostgreSQL)
using `httptest.Server` with the real chi router and a testcontainers PostgreSQL instance. This
is the primary behavioral verification that the Go port matches the Python API contract.

## Dependencies

- Task 037 (integration test infrastructure)
- Task 025 (chi router wiring — full app assembled)
- Task 019–024 (all HTTP handlers)

## Acceptance Criteria

- [ ] `backend/tests/integration/endpoint_archer_test.go` ports tests from
  `test_archer_endpoints.py`:
    - GET `/api/v0/archers` — list archers (authenticated)
    - GET `/api/v0/archers/:id` — get single archer
    - Unauthenticated requests → 401
- [ ] `backend/tests/integration/endpoint_auth_test.go` ports tests from
  `test_auth_endpoints.py`:
    - POST `/api/v0/auth/login` — with valid/invalid credentials
    - POST `/api/v0/auth/register` — with valid/missing fields
    - POST `/api/v0/auth/logout` — clears session
    - GET `/api/v0/auth/me` — returns authenticated archer
- [ ] `backend/tests/integration/endpoint_session_test.go` ports tests from
  `test_session_endpoints.py`:
    - POST `/api/v0/sessions` — create session
    - GET `/api/v0/sessions` — list sessions
    - GET `/api/v0/sessions/open` — get open session
    - PUT `/api/v0/sessions/:id/close` — close session
    - Cannot create session when one is already open → 409
- [ ] `backend/tests/integration/endpoint_slot_test.go` ports tests from
  `test_slot_endpoints.py`:
    - POST `/api/v0/slots` — create slot in open session
    - GET `/api/v0/slots?session_id=X` — list slots by session
    - PUT `/api/v0/slots/:id` — update slot
    - DELETE `/api/v0/slots/:id` — delete slot
    - Create slot in closed session → 422
- [ ] `backend/tests/integration/endpoint_shot_test.go` ports tests from
  `test_shot_endpoints.py`:
    - POST `/api/v0/shots` — create shot in slot
    - GET `/api/v0/shots?slot_id=X` — list shots by slot
    - PUT `/api/v0/shots/:id` — update shot
    - DELETE `/api/v0/shots/:id` — delete shot
- [ ] `backend/tests/integration/endpoint_faces_test.go` ports tests from
  `test_faces_endpoints.py`:
    - GET `/api/v0/faces` — list available faces
- [ ] `backend/tests/integration/endpoint_security_test.go` ports tests from
  `test_security_edge_cases.py`:
    - Cross-user access attempts → 403
    - Expired JWT → 401
    - Malformed JWT → 401
    - Missing auth cookie → 401
- [ ] All tests verify JSON response bodies match the expected API contract (field names,
  types, structure).
- [ ] Each test truncates tables after completion.
- [ ] `go test ./tests/integration/... -v -count=1` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/tests/integration/endpoint_archer_test.go` |
| Create | `backend/tests/integration/endpoint_auth_test.go` |
| Create | `backend/tests/integration/endpoint_session_test.go` |
| Create | `backend/tests/integration/endpoint_slot_test.go` |
| Create | `backend/tests/integration/endpoint_shot_test.go` |
| Create | `backend/tests/integration/endpoint_faces_test.go` |
| Create | `backend/tests/integration/endpoint_security_test.go` |
| Modify | `backend/tests/integration/helpers_test.go` (add HTTP test helpers) |

## Reference

Port tests from these Python files:

- [test_archer_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_archer_endpoints.py) (3.3KB)
- [test_auth_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_auth_endpoints.py) (6.7KB)
- [test_session_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_session_endpoints.py) (25.7KB)
- [test_slot_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_slot_endpoints.py) (33.8KB)
- [test_shot_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_shot_endpoints.py) (21.4KB)
- [test_faces_endpoints.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_faces_endpoints.py) (1.0KB)
- [test_security_edge_cases.py](file:///home/juanpa/Projects/arch-stats/backend/tests/endpoints/test_security_edge_cases.py) (4.8KB)

## Steps

- [ ] **Step 1: Add HTTP test helpers to `helpers_test.go`**

  Add utilities for creating an `httptest.Server` with the full chi router:

  ```go
  func newTestServer() *httptest.Server {
      // Build the full app with all handlers wired to testPool
      router := buildTestRouter(testPool)
      return httptest.NewServer(router)
  }

  func authRequest(method, url, jwt string, body io.Reader) *http.Request {
      req, _ := http.NewRequest(method, url, body)
      req.Header.Set("Content-Type", "application/json")
      req.AddCookie(&http.Cookie{Name: "access_token", Value: jwt})
      return req
  }
  ```

- [ ] **Step 2: Write `endpoint_auth_test.go`**

  Start with auth since it establishes the authentication flow other tests depend on.
  Port each test case from `test_auth_endpoints.py`.

- [ ] **Step 3: Write `endpoint_archer_test.go`**

  Port tests from `test_archer_endpoints.py`.

- [ ] **Step 4: Write `endpoint_session_test.go`**

  Port tests from `test_session_endpoints.py`. This is the largest test file (25.7KB Python).

- [ ] **Step 5: Write `endpoint_slot_test.go`**

  Port tests from `test_slot_endpoints.py`. This is the largest overall (33.8KB Python).

- [ ] **Step 6: Write `endpoint_shot_test.go`**

  Port tests from `test_shot_endpoints.py`.

- [ ] **Step 7: Write `endpoint_faces_test.go`**

  Port tests from `test_faces_endpoints.py` (smallest file).

- [ ] **Step 8: Write `endpoint_security_test.go`**

  Port tests from `test_security_edge_cases.py`.

- [ ] **Step 9: Run all integration tests**

  ```bash
  cd backend
  go test ./tests/integration/... -v -count=1
  ```

- [ ] **Step 10: Run go vet**

  ```bash
  cd backend
  go vet ./...
  ```

- [ ] **Step 11: Commit**

  ```bash
  git add -A
  git commit -m "test: port Python endpoint tests to Go integration tests"
  ```

## Verification

- `cd backend && go test ./tests/integration/... -v -count=1` — all tests pass.
- `cd backend && go vet ./...` — clean.
- JSON response shapes match the existing API contract (verified by test assertions).
- HTTP status codes match the Python test expectations.
