# Task 017: Implement Auth — Google One Tap + JWT + Session Tokens

## Git Branch

`refactor/017-auth-google-onetap-and-jwt`

## Objective

Implement the full authentication system in `internal/auth/`, porting the Python
`authentication.py` logic. This covers Google One Tap ID token verification, JWT minting and
verification, and SHA-256 hashed session tokens.

## Dependencies

- Task 003 (config with JWT and Google OAuth settings)
- Task 009 (auth session repository)
- Task 008 (archer repository)
- Task 007 (auth model structs)

## Acceptance Criteria

- [ ] `backend/internal/auth/google.go` provides:
    - `VerifyGoogleIDToken(ctx, credential, clientID) (*GoogleUserData, error)` — verifies a Google One Tap credential and returns extracted claims
- [ ] `backend/internal/auth/jwt.go` provides:
    - `BuildJWT(archerID, sid, issuedAt, expiresAt, secret, algorithm) (string, error)` — signs a JWT
    - `DecodeJWT(token, secret, algorithm) (*Claims, error)` — verifies and decodes a JWT
    - A `Claims` struct matching the Python JWT payload (sub, sid, exp, iat, iss, typ)
- [ ] `backend/internal/auth/session.go` provides:
    - `GenerateSessionToken(numBytes int) (raw []byte, err error)` — generates crypto/rand bytes
    - `HashSessionToken(raw []byte) []byte` — SHA-256 hash
- [ ] `backend/internal/auth/service.go` provides a `Service` orchestrating the full flow:
    - `LoginExisting(ctx, archer, googleData, now) (*model.AuthAuthenticated, error)`
    - `Register(ctx, payload, googleData, now) (*model.AuthAuthenticated, error)`
    - `ValidateSession(ctx, tokenHash) (*model.AuthSessionRead, error)`
- [ ] Unit tests cover:
    - JWT build + decode round-trip
    - Session token generation entropy (length check)
    - Hash consistency (same input → same hash)
    - Claims extraction from JWT
- [ ] `go test ./internal/auth/...` passes.
- [ ] `go vet ./...` reports no issues.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `backend/internal/auth/google.go` |
| Create | `backend/internal/auth/jwt.go` |
| Create | `backend/internal/auth/jwt_test.go` |
| Create | `backend/internal/auth/session.go` |
| Create | `backend/internal/auth/session_test.go` |
| Create | `backend/internal/auth/service.go` |
| Create | `backend/internal/auth/service_test.go` |
| Delete | `backend/internal/auth/.gitkeep` |
| Modify | `backend/go.mod` (add google idtoken + golang-jwt dependencies) |

## Reference

- Python auth: [authentication.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/core/authentication.py)
- Python auth schema: [auth_schema.py](file:///home/juanpa/Projects/arch-stats/backend-old/src/schema/auth_schema.py)

## Steps

- [ ] **Step 1: Add dependencies**

  ```bash
  cd backend
  go get github.com/golang-jwt/jwt/v5
  go get google.golang.org/api/idtoken
  ```

- [ ] **Step 2: Write failing tests for JWT**

  Create `backend/internal/auth/jwt_test.go`:
    - Test round-trip: `BuildJWT` → `DecodeJWT` → verify claims match
    - Test expired JWT is rejected
    - Test JWT with wrong secret is rejected
    - Test Claims struct fields (sub, sid, exp, iat, iss, typ)

- [ ] **Step 3: Write failing tests for session tokens**

  Create `backend/internal/auth/session_test.go`:
    - Test `GenerateSessionToken(32)` returns 32 bytes
    - Test two generated tokens are different (randomness)
    - Test `HashSessionToken` returns 32 bytes (SHA-256)
    - Test same input produces same hash

- [ ] **Step 4: Run tests to verify they fail**

  ```bash
  cd backend
  go test ./internal/auth/... -v
  ```

- [ ] **Step 5: Implement auth packages**

  Implement `jwt.go`, `session.go`, `google.go`, and `service.go` following the Python
  implementation structure.

- [ ] **Step 6: Run tests to verify they pass**

  ```bash
  cd backend
  go test ./internal/auth/... -v
  ```

- [ ] **Step 7: Run go vet and build**

  ```bash
  cd backend
  go vet ./...
  go build ./...
  ```

- [ ] **Step 8: Commit**

  ```bash
  rm -f backend/internal/auth/.gitkeep
  git add -A
  git commit -m "feat: add auth system with Google One Tap, JWT, and session tokens"
  ```

## Verification

- `cd backend && go test ./internal/auth/... -v` — all tests pass.
- `cd backend && go vet ./...` — clean.
- `cd backend && go build ./...` — compiles.
