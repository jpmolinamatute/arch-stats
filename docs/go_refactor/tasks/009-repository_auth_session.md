# Task 009: Build Repository — Auth Session

## Git Branch

`refactor/009-repository-auth-session`

## Objective

Implement the auth session repository for managing login session records (session token hashes,
expiry, user agent, IP). This maps from the Python `auth_model.py` and is used by the
authentication system to create, validate, and expire sessions.

## Dependencies

- Task 007 (auth model structs)
- Task 008 (repository base patterns + DBTX interface)

## Acceptance Criteria

- [x] `backend/internal/repository/auth_session.go` implements `AuthSessionRepo` with methods:
    - `Create(ctx, data model.AuthSessionCreate) error`
    - `FindByTokenHash(ctx, hash []byte) (*model.AuthSessionRead, error)`
    - `DeleteByArcherID(ctx, archerID uuid.UUID) error` (logout all sessions)
    - `DeleteExpired(ctx) (int64, error)` (cleanup expired sessions)
- [x] All queries use squirrel.
- [x] Unit tests verify query building and scan logic.
- [x] `go test ./internal/repository/...` passes.
- [x] `go vet ./...` reports no issues.

## Files to Create

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/repository/auth_session.go` |
| Create | `backend/internal/repository/auth_session_test.go` |

## Reference

- Python auth model: [auth_model.py](file:///home/juanpa/Projects/arch-stats/backend/src/models/auth_model.py)
- Python auth schema: [auth_schema.py](file:///home/juanpa/Projects/arch-stats/backend/src/schema/auth_schema.py)

## Steps

- [x] **Step 1: Write failing tests**

  Create `backend/internal/repository/auth_session_test.go`:
    - Test `Create` builds correct INSERT with session_token_hash, archer_id, expires_at, ua, ip_inet
    - Test `FindByTokenHash` builds correct SELECT with WHERE on session_token_hash
    - Test `DeleteByArcherID` builds correct DELETE with WHERE on archer_id
    - Test `DeleteExpired` builds correct DELETE with WHERE expires_at < NOW()

- [x] **Step 2: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 3: Implement `auth_session.go`**

  Implement `AuthSessionRepo` struct with all methods using squirrel and the `DBTX` interface.

- [x] **Step 4: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/repository/... -v
  ```

- [x] **Step 5: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [x] **Step 6: Commit**

  ```bash
  git add -A
  git commit -m "feat: add auth session repository for login token management"
  ```

## Verification

- `cd backend && go test ./internal/repository/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
