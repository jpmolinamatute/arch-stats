# Task 017: Auth System (Google One Tap, JWT, and Session Tokens) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the full authentication system in `internal/auth/`, porting the Python `authentication.py` logic. This covers Google One Tap ID token verification, JWT minting and decoding, SHA-256 hashed session token generation, and the authentication domain service orchestrating login, registration, and session validation.

**Architecture:**
- `session.go`: Cryptographically secure random session token generator and SHA-256 digest hasher, plus Base64URL session identifier encoders.
- `jwt.go`: Access token minting and signature verification using HMAC (`github.com/golang-jwt/jwt/v5`) embedding archer ID, session identifier, expiration, and issuer claims.
- `google.go`: Google One Tap ID token verification (`google.golang.org/api/idtoken`) with claim extraction into `GoogleUserData` and `AuthNeedsRegistration` response builder, supporting pluggable verifiers for offline testing.
- `service.go`: `auth.Service` orchestrator consuming `ArcherRepository` and `SessionRepository` interfaces via dependency injection to handle `LoginExisting`, `Register`, `ValidateSession`, `RevokeSession`, and `RevokeAllSessions` with comprehensive unit test coverage using repository mocks.

**Tech Stack:** Go 1.27+, `github.com/golang-jwt/jwt/v5`, `google.golang.org/api/idtoken`, standard library (`crypto/rand`, `crypto/sha256`, `encoding/base64`, `time`, `errors`, `fmt`, `strings`), internal packages (`model`, `apperror`).

**Spec:** [docs/go_refactor/tasks/017-auth_google_onetap_and_jwt.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/017-auth_google_onetap_and_jwt.md)

## Global Constraints

- Git branch: `refactor/017-auth-google-onetap-and-jwt`
- Package path: `github.com/jpmolinamatute/arch-stats/backend/internal/auth`
- Error handling: Wrap internal errors with `%w` using contextual descriptive messages (`fmt.Errorf("...: %w", err)`). Return sentinel `apperror.ErrNotFound`, `apperror.ErrUnauthorized`, and `apperror.Wrap(apperror.ErrValidation, ...)` as appropriate.
- Dependency injection: `Service` must accept repository interfaces (`ArcherRepository`, `SessionRepository`) and `Config` in its constructor `NewService(archers ArcherRepository, sessions SessionRepository, cfg Config) *Service`.
- Mock testing: Service unit tests must use mock repositories implementing the repository interfaces without database or network dependencies. Google token verification must support mock/offline verifier injection for testing without network calls.
- Formatting must adhere to `gofumpt` and linting must pass `golangci-lint run ./...`.
- `go test -race ./internal/auth/... -v` must pass.
- `go vet ./...` must report no issues.
- `go build ./...` must compile cleanly.

---

## File Structure

```
backend/
├── go.mod                        # [MODIFY] Add github.com/golang-jwt/jwt/v5 and google.golang.org/api/idtoken
├── go.sum                        # [MODIFY] Checksum entries for new dependencies
└── internal/
    └── auth/
        ├── .gitkeep              # [DELETE] Remove placeholder once files are added
        ├── session.go            # [NEW] GenerateSessionToken, HashSessionToken, Base64URL session ID helpers
        ├── session_test.go       # [NEW] Unit tests for session token generation, hashing, and encoding
        ├── jwt.go                # [NEW] Claims struct, BuildJWT, DecodeJWT
        ├── jwt_test.go           # [NEW] Unit tests for JWT minting, decoding, claims assertions, validation
        ├── google.go             # [NEW] GoogleUserData struct, VerifyGoogleIDToken, BuildNeedsRegistrationResponse
        ├── google_test.go        # [NEW] Unit tests for Google claim extraction, verifier mocking, response builders
        ├── service.go            # [NEW] Repository interfaces, Service struct, LoginExisting, Register, ValidateSession
        └── service_test.go       # [NEW] Mock repositories and unit tests for Service orchestration
```

---

### Task 1: Git Branch Setup & Dependencies

**Files:**
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Delete: `backend/internal/auth/.gitkeep`

**Interfaces:**
- Consumes: `main` branch
- Produces: `refactor/017-auth-google-onetap-and-jwt` branch with installed dependencies (`github.com/golang-jwt/jwt/v5`, `google.golang.org/api/idtoken`)

- [x] **Step 1: Check out git branch**

```bash
git switch -c refactor/017-auth-google-onetap-and-jwt
```

- [x] **Step 2: Add Go module dependencies**

```bash
cd backend
go get github.com/golang-jwt/jwt/v5
go get google.golang.org/api/idtoken
go mod tidy
```

- [x] **Step 3: Remove .gitkeep placeholder**

```bash
rm -f internal/auth/.gitkeep
```

- [x] **Step 4: Verify dependencies compile**

```bash
cd backend
go build ./...
```
Expected: Compiles cleanly with new dependencies added to `go.mod`.

- [x] **Step 5: Commit dependency changes**

```bash
git add backend/go.mod backend/go.sum backend/internal/auth/.gitkeep
git commit -m "chore(deps): add golang-jwt and google api idtoken modules"
```

---

### Task 2: Session Token Generation & Hashing (`session_test.go` & `session.go`)

**Files:**
- Create: `backend/internal/auth/session_test.go`
- Create: `backend/internal/auth/session.go`

**Interfaces:**
- Consumes: `crypto/rand`, `crypto/sha256`, `encoding/base64`, `apperror`
- Produces:
  - `GenerateSessionToken(numBytes int) ([]byte, error)`
  - `HashSessionToken(raw []byte) []byte`
  - `EncodeSessionID(raw []byte) string`
  - `DecodeSessionID(sid string) ([]byte, error)`

- [x] **Step 1: Write failing tests in `backend/internal/auth/session_test.go`**

Create `backend/internal/auth/session_test.go`:

```go
package auth_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
)

func TestGenerateSessionToken(t *testing.T) {
	t.Parallel()

	t.Run("generates requested number of bytes", func(t *testing.T) {
		t.Parallel()
		raw, err := auth.GenerateSessionToken(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(raw) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(raw))
		}
	})

	t.Run("two generated tokens have high entropy and differ", func(t *testing.T) {
		t.Parallel()
		raw1, err1 := auth.GenerateSessionToken(32)
		raw2, err2 := auth.GenerateSessionToken(32)
		if err1 != nil || err2 != nil {
			t.Fatalf("unexpected errors: %v, %v", err1, err2)
		}
		if bytes.Equal(raw1, raw2) {
			t.Fatal("expected distinct random tokens, but tokens are identical")
		}
	})

	t.Run("invalid byte count returns validation error", func(t *testing.T) {
		t.Parallel()
		_, err := auth.GenerateSessionToken(0)
		if err == nil {
			t.Fatal("expected error for 0 bytes, got nil")
		}
		appErr, ok := err.(*apperror.AppError)
		if !ok || appErr.Code() != "VALIDATION" {
			t.Fatalf("expected VALIDATION apperror, got: %v", err)
		}

		_, errNeg := auth.GenerateSessionToken(-5)
		if errNeg == nil {
			t.Fatal("expected error for negative bytes, got nil")
		}
	})
}

func TestHashSessionToken(t *testing.T) {
	t.Parallel()

	t.Run("returns 32-byte sha256 digest", func(t *testing.T) {
		t.Parallel()
		raw := []byte("secret-session-token-bytes")
		hash := auth.HashSessionToken(raw)
		if len(hash) != 32 {
			t.Fatalf("expected 32-byte SHA-256 hash, got %d", len(hash))
		}
	})

	t.Run("hash is deterministic for same input", func(t *testing.T) {
		t.Parallel()
		raw := []byte("consistent-session-input")
		hash1 := auth.HashSessionToken(raw)
		hash2 := auth.HashSessionToken(raw)
		if !bytes.Equal(hash1, hash2) {
			t.Fatalf("expected identical hashes for same input, got %x and %x", hash1, hash2)
		}
	})

	t.Run("matches known SHA-256 test vector", func(t *testing.T) {
		t.Parallel()
		// SHA-256("hello world") = b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
		input := []byte("hello world")
		expectedHex := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
		hash := auth.HashSessionToken(input)
		if hex.EncodeToString(hash) != expectedHex {
			t.Fatalf("expected hex %s, got %s", expectedHex, hex.EncodeToString(hash))
		}
	})
}

func TestSessionIDEncoding(t *testing.T) {
	t.Parallel()

	t.Run("encode and decode round-trip", func(t *testing.T) {
		t.Parallel()
		raw := []byte("32-bytes-of-random-session-data!")
		encoded := auth.EncodeSessionID(raw)
		if encoded == "" {
			t.Fatal("expected non-empty base64url encoded string")
		}

		decoded, err := auth.DecodeSessionID(encoded)
		if err != nil {
			t.Fatalf("unexpected error decoding: %v", err)
		}
		if !bytes.Equal(raw, decoded) {
			t.Fatalf("expected %q, got %q", raw, decoded)
		}
	})

	t.Run("decode invalid base64 returns error", func(t *testing.T) {
		t.Parallel()
		_, err := auth.DecodeSessionID("???not-valid-base64url???")
		if err == nil {
			t.Fatal("expected error decoding invalid base64, got nil")
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/auth/... -v
```
Expected: FAIL due to undefined functions in `auth`.

- [x] **Step 3: Implement `backend/internal/auth/session.go`**

Create `backend/internal/auth/session.go`:

```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// GenerateSessionToken creates a cryptographically secure random byte sequence of length numBytes.
func GenerateSessionToken(numBytes int) ([]byte, error) {
	if numBytes <= 0 {
		return nil, apperror.Wrap(apperror.ErrValidation, "session token size must be greater than 0")
	}

	b := make([]byte, numBytes)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("reading cryptographically secure random bytes: %w", err)
	}

	return b, nil
}

// HashSessionToken computes the SHA-256 digest of raw session token bytes.
// Storing this hash prevents token forgery even if session database rows are compromised.
func HashSessionToken(raw []byte) []byte {
	h := sha256.Sum256(raw)
	return h[:]
}

// EncodeSessionID encodes raw session token bytes using unpadded URL-safe Base64.
// This matches the Python session ID string representation embedded in JWT payloads.
func EncodeSessionID(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeSessionID decodes an unpadded URL-safe Base64 session ID string back to raw bytes.
func DecodeSessionID(sid string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(sid)
	if err != nil {
		return nil, fmt.Errorf("decoding base64url session id: %w", err)
	}
	return raw, nil
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test -race ./internal/auth/... -v
```
Expected: PASS with 100% coverage of `session.go`.

- [x] **Step 5: Commit session token implementation**

```bash
git add backend/internal/auth/session.go backend/internal/auth/session_test.go
git commit -m "feat(auth): add session token generation, hashing, and encoding"
```

---

### Task 3: JWT Minting & Verification (`jwt_test.go` & `jwt.go`)

**Files:**
- Create: `backend/internal/auth/jwt_test.go`
- Create: `backend/internal/auth/jwt.go`

**Interfaces:**
- Consumes: `github.com/golang-jwt/jwt/v5`, `github.com/google/uuid`, `apperror`
- Produces:
  - `Claims` struct matching Python JWT payload (`sub`, `sid`, `exp`, `iat`, `iss`, `typ`)
  - `BuildJWT(archerID uuid.UUID, sid string, issuedAt, expiresAt time.Time, secret, algorithm string) (string, error)`
  - `DecodeJWT(tokenStr, secret, algorithm string) (*Claims, error)`

- [x] **Step 1: Write failing tests in `backend/internal/auth/jwt_test.go`**

Create `backend/internal/auth/jwt_test.go`:

```go
package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
)

func TestJWT(t *testing.T) {
	t.Parallel()

	archerID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	sid := "test-session-id-base64"
	secret := "super-secret-jwt-signing-key-minimum-length"
	algorithm := "HS256"

	t.Run("round-trip build and decode succeeds and preserves all claims", func(t *testing.T) {
		t.Parallel()
		issuedAt := time.Now().UTC().Truncate(time.Second)
		expiresAt := issuedAt.Add(24 * time.Hour)

		token, err := auth.BuildJWT(archerID, sid, issuedAt, expiresAt, secret, algorithm)
		if err != nil {
			t.Fatalf("unexpected error building jwt: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token string")
		}

		claims, err := auth.DecodeJWT(token, secret, algorithm)
		if err != nil {
			t.Fatalf("unexpected error decoding jwt: %v", err)
		}

		if claims.Sub != archerID.String() {
			t.Errorf("expected sub %s, got %s", archerID.String(), claims.Sub)
		}
		if claims.SID != sid {
			t.Errorf("expected sid %s, got %s", sid, claims.SID)
		}
		if claims.Exp != expiresAt.Unix() {
			t.Errorf("expected exp %d, got %d", expiresAt.Unix(), claims.Exp)
		}
		if claims.Iat != issuedAt.Unix() {
			t.Errorf("expected iat %d, got %d", issuedAt.Unix(), claims.Iat)
		}
		if claims.Iss != "arch-stats" {
			t.Errorf("expected iss 'arch-stats', got %s", claims.Iss)
		}
		if claims.Typ != "access" {
			t.Errorf("expected typ 'access', got %s", claims.Typ)
		}

		parsedArcherID, err := claims.ArcherID()
		if err != nil {
			t.Fatalf("unexpected error parsing archer ID from claims: %v", err)
		}
		if parsedArcherID != archerID {
			t.Errorf("expected archerID %s, got %s", archerID, parsedArcherID)
		}
	})

	t.Run("expired jwt is rejected with unauthorized error", func(t *testing.T) {
		t.Parallel()
		issuedAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
		expiresAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)

		token, err := auth.BuildJWT(archerID, sid, issuedAt, expiresAt, secret, algorithm)
		if err != nil {
			t.Fatalf("unexpected error building jwt: %v", err)
		}

		_, err = auth.DecodeJWT(token, secret, algorithm)
		if err == nil {
			t.Fatal("expected error decoding expired token, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("jwt with wrong secret is rejected", func(t *testing.T) {
		t.Parallel()
		issuedAt := time.Now().UTC().Truncate(time.Second)
		expiresAt := issuedAt.Add(1 * time.Hour)

		token, err := auth.BuildJWT(archerID, sid, issuedAt, expiresAt, secret, algorithm)
		if err != nil {
			t.Fatalf("unexpected error building jwt: %v", err)
		}

		_, err = auth.DecodeJWT(token, "different-wrong-secret-key", algorithm)
		if err == nil {
			t.Fatal("expected error decoding token with wrong secret, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("jwt with wrong algorithm is rejected", func(t *testing.T) {
		t.Parallel()
		issuedAt := time.Now().UTC().Truncate(time.Second)
		expiresAt := issuedAt.Add(1 * time.Hour)

		token, err := auth.BuildJWT(archerID, sid, issuedAt, expiresAt, secret, "HS256")
		if err != nil {
			t.Fatalf("unexpected error building jwt: %v", err)
		}

		_, err = auth.DecodeJWT(token, secret, "HS512")
		if err == nil {
			t.Fatal("expected error decoding token with algorithm mismatch, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("invalid build arguments return validation error", func(t *testing.T) {
		t.Parallel()
		now := time.Now().UTC()

		_, err := auth.BuildJWT(uuid.Nil, sid, now, now.Add(time.Hour), secret, algorithm)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation for nil archerID, got %v", err)
		}

		_, err = auth.BuildJWT(archerID, "", now, now.Add(time.Hour), secret, algorithm)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation for empty sid, got %v", err)
		}

		_, err = auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), "", algorithm)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation for empty secret, got %v", err)
		}

		_, err = auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), secret, "UNSUPPORTED_ALG")
		if err == nil {
			t.Errorf("expected error for unsupported algorithm, got nil")
		}
	})

	t.Run("invalid decode arguments return validation error", func(t *testing.T) {
		t.Parallel()
		_, err := auth.DecodeJWT("", secret, algorithm)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation for empty token, got %v", err)
		}

		_, err = auth.DecodeJWT("some.jwt.token", "", algorithm)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("expected ErrValidation for empty secret, got %v", err)
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/auth/... -v
```
Expected: FAIL due to undefined functions in `jwt.go`.

- [x] **Step 3: Implement `backend/internal/auth/jwt.go`**

Create `backend/internal/auth/jwt.go`:

```go
package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// Claims represents the JWT payload claims for an authenticated archer session,
// exactly matching the Python authentication payload format (sub, sid, exp, iat, iss, typ).
type Claims struct {
	Sub string `json:"sub"`
	SID string `json:"sid"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
	Typ string `json:"typ"`
}

// ArcherID parses and returns the Subject claim as a uuid.UUID.
func (c *Claims) ArcherID() (uuid.UUID, error) {
	id, err := uuid.Parse(c.Sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing archer id from jwt subject %q: %w", c.Sub, err)
	}
	return id, nil
}

// GetExpirationTime implements jwt.Claims interface.
func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

// GetIssuedAt implements jwt.Claims interface.
func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.Iat == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Iat, 0)), nil
}

// GetNotBefore implements jwt.Claims interface.
func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuer implements jwt.Claims interface.
func (c *Claims) GetIssuer() (string, error) {
	return c.Iss, nil
}

// GetSubject implements jwt.Claims interface.
func (c *Claims) GetSubject() (string, error) {
	return c.Sub, nil
}

// GetAudience implements jwt.Claims interface.
func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// BuildJWT mints and signs a JWT embedding the archer UUID and session identifier.
func BuildJWT(
	archerID uuid.UUID,
	sid string,
	issuedAt, expiresAt time.Time,
	secret, algorithm string,
) (string, error) {
	if archerID == uuid.Nil {
		return "", apperror.Wrap(apperror.ErrValidation, "archerID cannot be nil")
	}
	if strings.TrimSpace(sid) == "" {
		return "", apperror.Wrap(apperror.ErrValidation, "sid cannot be empty")
	}
	if strings.TrimSpace(secret) == "" {
		return "", apperror.Wrap(apperror.ErrValidation, "jwt secret cannot be empty")
	}

	signingMethod := jwt.GetSigningMethod(algorithm)
	if signingMethod == nil || !strings.HasPrefix(algorithm, "HS") {
		return "", fmt.Errorf("unsupported HMAC signing algorithm: %s", algorithm)
	}

	claims := Claims{
		Sub: archerID.String(),
		SID: sid,
		Exp: expiresAt.Unix(),
		Iat: issuedAt.Unix(),
		Iss: "arch-stats",
		Typ: "access",
	}

	token := jwt.NewWithClaims(signingMethod, &claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return signed, nil
}

// DecodeJWT verifies and decodes an access JWT, enforcing signing algorithm, secret, and expiration.
func DecodeJWT(tokenStr, secret, algorithm string) (*Claims, error) {
	if strings.TrimSpace(tokenStr) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "token cannot be empty")
	}
	if strings.TrimSpace(secret) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "jwt secret cannot be empty")
	}

	var claims Claims
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != algorithm {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperror.Wrap(apperror.ErrUnauthorized, "token has expired")
		}
		return nil, apperror.Wrap(apperror.ErrUnauthorized, fmt.Sprintf("invalid token: %v", err))
	}

	if !parsedToken.Valid {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "token is invalid")
	}

	return &claims, nil
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test -race ./internal/auth/... -v
```
Expected: PASS with complete test coverage of `jwt.go`.

- [x] **Step 5: Commit JWT implementation**

```bash
git add backend/internal/auth/jwt.go backend/internal/auth/jwt_test.go
git commit -m "feat(auth): add JWT minting, decoding, and claims verification"
```

---

### Task 4: Google One Tap ID Token Verification (`google_test.go` & `google.go`)

**Files:**
- Create: `backend/internal/auth/google_test.go`
- Create: `backend/internal/auth/google.go`

**Interfaces:**
- Consumes: `google.golang.org/api/idtoken`, `model.AuthNeedsRegistration`, `apperror`
- Produces:
  - `GoogleUserData` struct
  - `GooglePayloadVerifier` type
  - `VerifyGoogleIDToken(ctx context.Context, credential, clientID string) (*GoogleUserData, error)`
  - `VerifyGoogleIDTokenWithVerifier(ctx context.Context, credential, clientID string, verifier GooglePayloadVerifier) (*GoogleUserData, error)`
  - `BuildNeedsRegistrationResponse(googleData *GoogleUserData) *model.AuthNeedsRegistration`

- [x] **Step 1: Write failing tests in `backend/internal/auth/google_test.go`**

Create `backend/internal/auth/google_test.go`:

```go
package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

func TestVerifyGoogleIDToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clientID := "test-google-client-id.apps.googleusercontent.com"

	t.Run("successfully extracts claims from valid payload", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, idToken, audience string) (*idtoken.Payload, error) {
			if idToken != "valid-credential" || audience != clientID {
				return nil, errors.New("mismatch")
			}
			return &idtoken.Payload{
				Subject: "google-subject-12345",
				Claims: map[string]any{
					"email":          "archer@example.com",
					"email_verified": true,
					"name":           "Robin Hood",
					"given_name":     "Robin",
					"family_name":    "Hood",
					"picture":        "https://example.com/avatar.jpg",
					"locale":         "en",
					"hd":             "example.com",
				},
			}, nil
		}

		data, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "valid-credential", clientID, mockVerifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if data.Sub != "google-subject-12345" {
			t.Errorf("expected sub google-subject-12345, got %s", data.Sub)
		}
		if data.Email != "archer@example.com" {
			t.Errorf("expected email archer@example.com, got %s", data.Email)
		}
		if !data.EmailVerified {
			t.Errorf("expected email_verified true, got false")
		}
		if data.GivenName != "Robin" {
			t.Errorf("expected given_name Robin, got %s", data.GivenName)
		}
		if data.FamilyName != "Hood" {
			t.Errorf("expected family_name Hood, got %s", data.FamilyName)
		}
		if data.Picture != "https://example.com/avatar.jpg" {
			t.Errorf("expected picture url, got %s", data.Picture)
		}
	})

	t.Run("verifier error translates to unauthorized apperror", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return nil, errors.New("signature expired or invalid")
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "bad-credential", clientID, mockVerifier)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("missing subject or email in token returns validation error", func(t *testing.T) {
		t.Parallel()
		mockVerifierMissingEmail := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-subject-12345",
				Claims: map[string]any{
					"name": "No Email Archer",
				},
			}, nil
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", clientID, mockVerifierMissingEmail)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing email, got: %v", err)
		}

		mockVerifierMissingSub := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "",
				Claims: map[string]any{
					"email": "archer@example.com",
				},
			}, nil
		}

		_, err = auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", clientID, mockVerifierMissingSub)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing subject, got: %v", err)
		}
	})

	t.Run("empty credential or clientID returns validation error", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return nil, nil
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "", clientID, mockVerifier)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty credential, got: %v", err)
		}

		_, err = auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", "", mockVerifier)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty clientID, got: %v", err)
		}
	})
}

func TestBuildNeedsRegistrationResponse(t *testing.T) {
	t.Parallel()

	t.Run("correctly maps full google claims", func(t *testing.T) {
		t.Parallel()
		googleData := &auth.GoogleUserData{
			Sub:        "google-sub-456",
			Email:      "legolas@woodland.realm",
			GivenName:  "Legolas",
			FamilyName: "Greenleaf",
			Picture:    "https://lotr.realm/legolas.png",
		}

		res := auth.BuildNeedsRegistrationResponse(googleData)
		if res == nil {
			t.Fatal("expected non-nil response")
		}

		if res.Status != model.AuthStatusNeedsRegistration {
			t.Errorf("expected status %s, got %s", model.AuthStatusNeedsRegistration, res.Status)
		}
		if res.GoogleEmail != "legolas@woodland.realm" {
			t.Errorf("expected email legolas@woodland.realm, got %s", res.GoogleEmail)
		}
		if res.GoogleSubject != "google-sub-456" {
			t.Errorf("expected subject google-sub-456, got %s", res.GoogleSubject)
		}
		if res.GivenName == nil || *res.GivenName != "Legolas" {
			t.Errorf("expected given_name Legolas, got %v", res.GivenName)
		}
		if !res.GivenNameProvided {
			t.Errorf("expected GivenNameProvided true")
		}
		if res.FamilyName == nil || *res.FamilyName != "Greenleaf" {
			t.Errorf("expected family_name Greenleaf, got %v", res.FamilyName)
		}
		if !res.FamilyNameProvided {
			t.Errorf("expected FamilyNameProvided true")
		}
		if res.PictureURL == nil || *res.PictureURL != "https://lotr.realm/legolas.png" {
			t.Errorf("expected picture url, got %v", res.PictureURL)
		}
	})

	t.Run("correctly handles empty optional names", func(t *testing.T) {
		t.Parallel()
		googleData := &auth.GoogleUserData{
			Sub:   "google-sub-789",
			Email: "mystery@example.com",
		}

		res := auth.BuildNeedsRegistrationResponse(googleData)
		if res.GivenName != nil {
			t.Errorf("expected nil given_name, got %v", res.GivenName)
		}
		if res.GivenNameProvided {
			t.Errorf("expected GivenNameProvided false")
		}
		if res.FamilyName != nil {
			t.Errorf("expected nil family_name, got %v", res.FamilyName)
		}
		if res.FamilyNameProvided {
			t.Errorf("expected FamilyNameProvided false")
		}
		if res.PictureURL != nil {
			t.Errorf("expected nil picture_url, got %v", res.PictureURL)
		}
	})

	t.Run("nil googleData returns nil", func(t *testing.T) {
		t.Parallel()
		res := auth.BuildNeedsRegistrationResponse(nil)
		if res != nil {
			t.Errorf("expected nil response for nil googleData, got %v", res)
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/auth/... -v
```
Expected: FAIL due to undefined functions in `google.go`.

- [x] **Step 3: Implement `backend/internal/auth/google.go`**

Create `backend/internal/auth/google.go`:

```go
package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

// GoogleUserData represents the verified subset of Google ID Token (OIDC) claims.
type GoogleUserData struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`
	Picture       string `json:"picture,omitempty"`
	Locale        string `json:"locale,omitempty"`
	HostedDomain  string `json:"hd,omitempty"`
}

// GooglePayloadVerifier defines the function signature for verifying Google ID tokens.
type GooglePayloadVerifier func(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)

// defaultGooglePayloadVerifier invokes Google's official idtoken.Validate validator.
var defaultGooglePayloadVerifier GooglePayloadVerifier = idtoken.Validate

// VerifyGoogleIDToken verifies a Google One Tap credential against the expected client ID
// using Google's public key sets and returns extracted user claims.
func VerifyGoogleIDToken(ctx context.Context, credential, clientID string) (*GoogleUserData, error) {
	return VerifyGoogleIDTokenWithVerifier(ctx, credential, clientID, defaultGooglePayloadVerifier)
}

// VerifyGoogleIDTokenWithVerifier verifies credentials using a pluggable verifier, allowing pure unit testing.
func VerifyGoogleIDTokenWithVerifier(
	ctx context.Context,
	credential, clientID string,
	verifier GooglePayloadVerifier,
) (*GoogleUserData, error) {
	if strings.TrimSpace(credential) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google credential cannot be empty")
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google client id cannot be empty")
	}
	if verifier == nil {
		verifier = defaultGooglePayloadVerifier
	}

	payload, err := verifier(ctx, credential, clientID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, fmt.Sprintf("invalid google credential: %v", err))
	}

	sub := strings.TrimSpace(payload.Subject)
	if sub == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google token missing subject claim")
	}

	email, _ := payload.Claims["email"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google token missing email claim")
	}

	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	givenName, _ := payload.Claims["given_name"].(string)
	familyName, _ := payload.Claims["family_name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	locale, _ := payload.Claims["locale"].(string)
	hd, _ := payload.Claims["hd"].(string)

	return &GoogleUserData{
		Sub:           sub,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		GivenName:     givenName,
		FamilyName:    familyName,
		Picture:       picture,
		Locale:        locale,
		HostedDomain:  hd,
	}, nil
}

// BuildNeedsRegistrationResponse constructs an AuthNeedsRegistration payload from Google claims.
func BuildNeedsRegistrationResponse(googleData *GoogleUserData) *model.AuthNeedsRegistration {
	if googleData == nil {
		return nil
	}

	var (
		given   *string
		family  *string
		picture *string
	)

	if trimmed := strings.TrimSpace(googleData.GivenName); trimmed != "" {
		given = &trimmed
	}
	if trimmed := strings.TrimSpace(googleData.FamilyName); trimmed != "" {
		family = &trimmed
	}
	if trimmed := strings.TrimSpace(googleData.Picture); trimmed != "" {
		picture = &trimmed
	}

	return &model.AuthNeedsRegistration{
		Status:             model.AuthStatusNeedsRegistration,
		GoogleEmail:        googleData.Email,
		GoogleSubject:      googleData.Sub,
		GivenName:          given,
		FamilyName:         family,
		GivenNameProvided:  given != nil,
		FamilyNameProvided: family != nil,
		PictureURL:         picture,
	}
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test -race ./internal/auth/... -v
```
Expected: PASS with complete test coverage of `google.go`.

- [x] **Step 5: Commit Google verification implementation**

```bash
git add backend/internal/auth/google.go backend/internal/auth/google_test.go
git commit -m "feat(auth): add Google One Tap token verification and registration response builder"
```

---

### Task 5: Auth Service Layer (`service_test.go` & `service.go`)

**Files:**
- Create: `backend/internal/auth/service_test.go`
- Create: `backend/internal/auth/service.go`

**Interfaces:**
- Consumes:
  - `model.ArcherRead`, `model.ArcherCreate`, `model.ArcherSet`, `model.ArcherFilter`
  - `model.AuthSessionCreate`, `model.AuthSessionRead`, `model.AuthAuthenticated`, `model.AuthRegistrationRequest`
  - `apperror.ErrNotFound`, `apperror.ErrUnauthorized`, `apperror.ErrValidation`
- Produces:
  - `ArcherRepository` interface:
    - `FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)`
    - `FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error)`
    - `Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)`
    - `Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error`
  - `SessionRepository` interface:
    - `Create(ctx context.Context, data model.AuthSessionCreate) error`
    - `FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)`
    - `RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error`
    - `DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error`
  - `Config` struct (`JWTSecret`, `JWTAlgorithm`, `JWTTTLMinutes`, `SessionTokenBytes`, `GoogleOAuthClientID`)
  - `SessionMetadata` struct (`UserAgent *string`, `IPAddress *string`)
  - `Service` struct and constructor `NewService(archers ArcherRepository, sessions SessionRepository, cfg Config) *Service`
  - Methods:
    - `LoginExisting(ctx context.Context, archer *model.ArcherRead, googleData *GoogleUserData, now time.Time, meta ...SessionMetadata) (*model.AuthAuthenticated, error)`
    - `Register(ctx context.Context, payload model.AuthRegistrationRequest, googleData *GoogleUserData, now time.Time, meta ...SessionMetadata) (*model.AuthAuthenticated, error)`
    - `ValidateSession(ctx context.Context, tokenHash []byte) (*model.AuthSessionRead, error)`
    - `RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error`
    - `RevokeAllSessions(ctx context.Context, archerID uuid.UUID) error`
    - `VerifyGoogleToken(ctx context.Context, credential string) (*GoogleUserData, error)`

- [x] **Step 1: Write failing tests in `backend/internal/auth/service_test.go`**

Create `backend/internal/auth/service_test.go`:

```go
package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockArcherRepo struct {
	findByIDFn            func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	findByGoogleSubjectFn func(ctx context.Context, sub string) (*model.ArcherRead, error)
	createFn              func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn              func(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
}

func (m *mockArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockArcherRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error) {
	if m.findByGoogleSubjectFn != nil {
		return m.findByGoogleSubjectFn(ctx, sub)
	}
	return nil, nil
}

func (m *mockArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.New(), nil
}

func (m *mockArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

type mockSessionRepo struct {
	createFn             func(ctx context.Context, data model.AuthSessionCreate) error
	findByTokenHashFn    func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)
	revokeByTokenHashFn  func(ctx context.Context, hash []byte, revokedAt time.Time) error
	deleteByArcherIDFn   func(ctx context.Context, archerID uuid.UUID) error
}

func (m *mockSessionRepo) Create(ctx context.Context, data model.AuthSessionCreate) error {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return nil
}

func (m *mockSessionRepo) FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
	if m.findByTokenHashFn != nil {
		return m.findByTokenHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockSessionRepo) RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error {
	if m.revokeByTokenHashFn != nil {
		return m.revokeByTokenHashFn(ctx, hash, revokedAt)
	}
	return nil
}

func (m *mockSessionRepo) DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error {
	if m.deleteByArcherIDFn != nil {
		return m.deleteByArcherIDFn(ctx, archerID)
	}
	return nil
}

func defaultTestConfig() auth.Config {
	return auth.Config{
		JWTSecret:           "test-secret-key-that-is-sufficiently-long",
		JWTAlgorithm:        "HS256",
		JWTTTLMinutes:       60,
		SessionTokenBytes:   32,
		GoogleOAuthClientID: "test-client-id",
	}
}

func TestService_LoginExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	archerID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	now := time.Now().UTC()

	archer := &model.ArcherRead{
		ArcherID:      archerID,
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         "robin@sherwood.org",
		DateOfBirth:   "1990-01-01",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleBarebow,
		DrawWeight:    40.0,
		GoogleSubject: "google-sub-1",
	}

	googleData := &auth.GoogleUserData{
		Sub:     "google-sub-1",
		Email:   "robin@sherwood.org",
		Picture: "https://sherwood.org/pic.png",
	}

	t.Run("successfully logs in existing archer, updates last login, creates session and jwt", func(t *testing.T) {
		t.Parallel()
		var (
			updateCalled bool
			sessionSaved bool
		)

		archers := &mockArcherRepo{
			updateFn: func(_ context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
				updateCalled = true
				if filter.ArcherID == nil || *filter.ArcherID != archerID {
					t.Errorf("expected filter by archer id %s", archerID)
				}
				if data.LastLoginAt == nil || *data.LastLoginAt != now {
					t.Errorf("expected LastLoginAt to match now")
				}
				if data.GooglePictureURL == nil || *data.GooglePictureURL != googleData.Picture {
					t.Errorf("expected picture url %s", googleData.Picture)
				}
				return nil
			},
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				updated := *archer
				updated.LastLoginAt = now
				updated.GooglePictureURL = &googleData.Picture
				return &updated, nil
			},
		}

		sessions := &mockSessionRepo{
			createFn: func(_ context.Context, data model.AuthSessionCreate) error {
				sessionSaved = true
				if data.ArcherID != archerID {
					t.Errorf("expected session archerID %s, got %s", archerID, data.ArcherID)
				}
				if len(data.SessionTokenHash) != 32 {
					t.Errorf("expected 32-byte token hash")
				}
				return nil
			},
		}

		svc := auth.NewService(archers, sessions, defaultTestConfig())
		ua := "Mozilla/5.0"
		ip := "127.0.0.1"
		res, err := svc.LoginExisting(ctx, archer, googleData, now, auth.SessionMetadata{
			UserAgent: &ua,
			IPAddress: &ip,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !updateCalled {
			t.Errorf("expected archer update to be called")
		}
		if !sessionSaved {
			t.Errorf("expected session create to be called")
		}
		if res.Status != model.AuthStatusAuthenticated {
			t.Errorf("expected status %s, got %s", model.AuthStatusAuthenticated, res.Status)
		}
		if res.AccessToken == "" {
			t.Errorf("expected access token to be generated")
		}

		// Verify generated JWT
		claims, err := auth.DecodeJWT(res.AccessToken, defaultTestConfig().JWTSecret, defaultTestConfig().JWTAlgorithm)
		if err != nil {
			t.Fatalf("failed to decode generated jwt: %v", err)
		}
		if claims.Sub != archerID.String() {
			t.Errorf("expected jwt sub %s, got %s", archerID.String(), claims.Sub)
		}
	})

	t.Run("nil archer returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.LoginExisting(ctx, nil, googleData, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil archer, got: %v", err)
		}
	})
}

func TestService_Register(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	newArcherID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")

	googleData := &auth.GoogleUserData{
		Sub:        "new-google-sub-99",
		Email:      "will@scarlet.org",
		GivenName:  "Will",
		FamilyName: "Scarlet",
		Picture:    "https://scarlet.org/pic.png",
	}

	payload := model.AuthRegistrationRequest{
		Credential:  "dummy-google-credential",
		DateOfBirth: "1995-05-15",
		Gender:      model.GenderMale,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  34.5,
	}

	t.Run("registers new archer and returns authenticated response", func(t *testing.T) {
		t.Parallel()
		var (
			createdData model.ArcherCreate
			sessionMade bool
		)

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil // Does not exist
			},
			createFn: func(_ context.Context, data model.ArcherCreate) (uuid.UUID, error) {
				createdData = data
				return newArcherID, nil
			},
			findByIDFn: func(_ context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				return &model.ArcherRead{
					ArcherID:      newArcherID,
					FirstName:     createdData.FirstName,
					LastName:      createdData.LastName,
					Email:         createdData.Email,
					DateOfBirth:   createdData.DateOfBirth,
					Gender:        createdData.Gender,
					Bowstyle:      createdData.Bowstyle,
					DrawWeight:    createdData.DrawWeight,
					GoogleSubject: createdData.GoogleSubject,
				}, nil
			},
		}

		sessions := &mockSessionRepo{
			createFn: func(_ context.Context, _ model.AuthSessionCreate) error {
				sessionMade = true
				return nil
			},
		}

		svc := auth.NewService(archers, sessions, defaultTestConfig())
		res, err := svc.Register(ctx, payload, googleData, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdData.FirstName != "Will" || createdData.LastName != "Scarlet" {
			t.Errorf("expected names from google data, got %s %s", createdData.FirstName, createdData.LastName)
		}
		if createdData.Email != "will@scarlet.org" {
			t.Errorf("expected email will@scarlet.org, got %s", createdData.Email)
		}
		if !sessionMade {
			t.Errorf("expected session to be created")
		}
		if res.Status != model.AuthStatusAuthenticated {
			t.Errorf("expected status %s, got %s", model.AuthStatusAuthenticated, res.Status)
		}
	})

	t.Run("if archer already exists with google subject, performs login instead", func(t *testing.T) {
		t.Parallel()
		existingArcher := &model.ArcherRead{
			ArcherID:      newArcherID,
			FirstName:     "Existing",
			LastName:      "Archer",
			Email:         googleData.Email,
			GoogleSubject: googleData.Sub,
		}

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return existingArcher, nil
			},
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return existingArcher, nil
			},
		}

		sessions := &mockSessionRepo{}
		svc := auth.NewService(archers, sessions, defaultTestConfig())

		res, err := svc.Register(ctx, payload, googleData, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Archer.ArcherID != newArcherID {
			t.Errorf("expected existing archer ID %s, got %s", newArcherID, res.Archer.ArcherID)
		}
	})

	t.Run("fails when given name and family name are missing", func(t *testing.T) {
		t.Parallel()
		emptyGoogle := &auth.GoogleUserData{
			Sub:   "sub-without-names",
			Email: "noname@example.com",
		}
		emptyPayload := payload
		emptyPayload.FirstName = nil
		emptyPayload.LastName = nil

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(archers, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.Register(ctx, emptyPayload, emptyGoogle, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing names, got: %v", err)
		}
	})

	t.Run("fails for invalid date of birth format", func(t *testing.T) {
		t.Parallel()
		invalidPayload := payload
		invalidPayload.DateOfBirth = "15/05/1995" // Not YYYY-MM-DD

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(archers, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.Register(ctx, invalidPayload, googleData, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid date format, got: %v", err)
		}
	})
}

func TestService_ValidateSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokenHash := []byte("32-byte-hash-of-session-token!")
	now := time.Now().UTC()

	t.Run("valid active session returns session data", func(t *testing.T) {
		t.Parallel()
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			ArcherID:         uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(1 * time.Hour),
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		got, err := svc.ValidateSession(ctx, tokenHash)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AuthID != session.AuthID {
			t.Errorf("expected authID %s, got %s", session.AuthID, got.AuthID)
		}
	})

	t.Run("empty token hash returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, nil)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil hash, got: %v", err)
		}
	})

	t.Run("non-existent session returns not found", func(t *testing.T) {
		t.Parallel()
		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("revoked session returns unauthorized error", func(t *testing.T) {
		t.Parallel()
		revokedTime := now.Add(-10 * time.Minute)
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(1 * time.Hour),
			RevokedAt:        &revokedTime,
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("expired session returns unauthorized error", func(t *testing.T) {
		t.Parallel()
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(-10 * time.Minute), // Expired
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized for expired session, got: %v", err)
		}
	})
}

func TestService_RevokeSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokenHash := []byte("test-hash")
	now := time.Now().UTC()

	t.Run("calls revoke by token hash", func(t *testing.T) {
		t.Parallel()
		var called bool
		sessions := &mockSessionRepo{
			revokeByTokenHashFn: func(_ context.Context, hash []byte, revokedAt time.Time) error {
				called = true
				if string(hash) != string(tokenHash) {
					t.Errorf("unexpected hash")
				}
				if revokedAt != now {
					t.Errorf("unexpected revokedAt")
				}
				return nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		if err := svc.RevokeSession(ctx, tokenHash, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected RevokeByTokenHash to be called")
		}
	})

	t.Run("empty token hash returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		if err := svc.RevokeSession(ctx, nil, now); err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty hash, got: %v", err)
		}
	})
}

func TestService_RevokeAllSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	archerID := uuid.New()

	t.Run("calls delete by archer id", func(t *testing.T) {
		t.Parallel()
		var called bool
		sessions := &mockSessionRepo{
			deleteByArcherIDFn: func(_ context.Context, id uuid.UUID) error {
				called = true
				if id != archerID {
					t.Errorf("expected archerID %s, got %s", archerID, id)
				}
				return nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		if err := svc.RevokeAllSessions(ctx, archerID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected DeleteByArcherID to be called")
		}
	})

	t.Run("nil archer id returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		if err := svc.RevokeAllSessions(ctx, uuid.Nil); err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil archer id, got: %v", err)
		}
	})
}
```

- [x] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/auth/... -v
```
Expected: FAIL due to undefined `auth.NewService` and related types in `service.go`.

- [x] **Step 3: Implement `backend/internal/auth/service.go`**

Create `backend/internal/auth/service.go`:

```go
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherRepository defines persistence operations for archer profiles required by the auth system.
type ArcherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
}

// SessionRepository defines persistence operations for auth sessions required by the auth system.
type SessionRepository interface {
	Create(ctx context.Context, data model.AuthSessionCreate) error
	FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)
	RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error
	DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error
}

// Config provides configuration settings for JWT signing, session lifetime, and Google OAuth.
type Config struct {
	JWTSecret           string
	JWTAlgorithm        string
	JWTTTLMinutes       int
	SessionTokenBytes   int
	GoogleOAuthClientID string
}

// SessionMetadata captures optional client request metadata (e.g. User-Agent and client IP).
type SessionMetadata struct {
	UserAgent *string
	IPAddress *string
}

// Service orchestrates authentication workflows including Google ID token validation,
// archer registration, session token minting, and JWT issuance.
type Service struct {
	archers  ArcherRepository
	sessions SessionRepository
	cfg      Config
}

// NewService constructs an auth Service with repository and config dependencies.
func NewService(archers ArcherRepository, sessions SessionRepository, cfg Config) *Service {
	if cfg.SessionTokenBytes <= 0 {
		cfg.SessionTokenBytes = 32
	}
	if cfg.JWTTTLMinutes <= 0 {
		cfg.JWTTTLMinutes = 1440
	}
	if strings.TrimSpace(cfg.JWTAlgorithm) == "" {
		cfg.JWTAlgorithm = "HS256"
	}
	return &Service{
		archers:  archers,
		sessions: sessions,
		cfg:      cfg,
	}
}

// LoginExisting updates the archer's last login timestamp and Google picture, creates a new session,
// signs an access JWT, and returns an AuthAuthenticated domain payload.
func (s *Service) LoginExisting(
	ctx context.Context,
	archer *model.ArcherRead,
	googleData *GoogleUserData,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	if archer == nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer is required")
	}

	updateData := model.ArcherSet{
		LastLoginAt: &now,
	}
	if googleData != nil && strings.TrimSpace(googleData.Picture) != "" {
		trimmedPic := strings.TrimSpace(googleData.Picture)
		updateData.GooglePictureURL = &trimmedPic
	}

	filter := model.ArcherFilter{ArcherID: &archer.ArcherID}
	if err := s.archers.Update(ctx, updateData, filter); err != nil {
		return nil, fmt.Errorf("updating archer last login: %w", err)
	}

	updatedArcher, err := s.archers.FindByID(ctx, archer.ArcherID)
	if err != nil {
		return nil, fmt.Errorf("fetching updated archer: %w", err)
	}
	if updatedArcher == nil {
		return nil, apperror.ErrNotFound
	}

	return s.createSessionAndToken(ctx, updatedArcher, now, meta...)
}

// Register registers a new archer, validates demographics and names, and authenticates the newly created profile.
// If an archer with the same Google subject already exists, it transparently logs them in.
//
//nolint:gocritic // hugeParam: payload matches API request specification
func (s *Service) Register(
	ctx context.Context,
	payload model.AuthRegistrationRequest,
	googleData *GoogleUserData,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	if googleData == nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "googleData is required")
	}
	if strings.TrimSpace(googleData.Sub) == "" || strings.TrimSpace(googleData.Email) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "googleData must contain email and sub")
	}

	existing, err := s.archers.FindByGoogleSubject(ctx, googleData.Sub)
	if err != nil {
		return nil, fmt.Errorf("verifying existing archer by google subject: %w", err)
	}
	if existing != nil {
		return s.LoginExisting(ctx, existing, googleData, now, meta...)
	}

	given := ""
	if payload.FirstName != nil && strings.TrimSpace(*payload.FirstName) != "" {
		given = strings.TrimSpace(*payload.FirstName)
	} else if strings.TrimSpace(googleData.GivenName) != "" {
		given = strings.TrimSpace(googleData.GivenName)
	}

	family := ""
	if payload.LastName != nil && strings.TrimSpace(*payload.LastName) != "" {
		family = strings.TrimSpace(*payload.LastName)
	} else if strings.TrimSpace(googleData.FamilyName) != "" {
		family = strings.TrimSpace(googleData.FamilyName)
	}

	var missingNames []string
	if given == "" {
		missingNames = append(missingNames, "given_name is missing")
	}
	if family == "" {
		missingNames = append(missingNames, "family_name is missing")
	}
	if len(missingNames) > 0 {
		return nil, apperror.Wrap(apperror.ErrValidation, strings.Join(missingNames, ", "))
	}

	if _, err := time.Parse("2006-01-02", payload.DateOfBirth); err != nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if payload.DrawWeight <= 0 || payload.DrawWeight > 200 {
		return nil, apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}

	var pictureURL *string
	if trimmedPic := strings.TrimSpace(googleData.Picture); trimmedPic != "" {
		pictureURL = &trimmedPic
	}

	createData := model.ArcherCreate{
		FirstName:        given,
		LastName:         family,
		Email:            googleData.Email,
		DateOfBirth:      payload.DateOfBirth,
		Gender:           payload.Gender,
		Bowstyle:         payload.Bowstyle,
		DrawWeight:       payload.DrawWeight,
		ClubID:           payload.ClubID,
		GooglePictureURL: pictureURL,
		GoogleSubject:    googleData.Sub,
	}

	newArcherID, err := s.archers.Create(ctx, createData)
	if err != nil {
		return nil, fmt.Errorf("creating archer: %w", err)
	}

	createdArcher, err := s.archers.FindByID(ctx, newArcherID)
	if err != nil {
		return nil, fmt.Errorf("fetching created archer: %w", err)
	}
	if createdArcher == nil {
		return nil, apperror.ErrNotFound
	}

	return s.createSessionAndToken(ctx, createdArcher, now, meta...)
}

// ValidateSession retrieves an active session by its SHA-256 token hash and validates that it is neither
// revoked nor expired.
func (s *Service) ValidateSession(ctx context.Context, tokenHash []byte) (*model.AuthSessionRead, error) {
	if len(tokenHash) == 0 {
		return nil, apperror.Wrap(apperror.ErrValidation, "token hash is required")
	}

	session, err := s.sessions.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("finding session by token hash: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}

	if session.RevokedAt != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "session has been revoked")
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "session has expired")
	}

	return session, nil
}

// RevokeSession revokes a single session identified by its token hash.
func (s *Service) RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	if len(tokenHash) == 0 {
		return apperror.Wrap(apperror.ErrValidation, "token hash is required")
	}
	if err := s.sessions.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

// RevokeAllSessions invalidates all sessions belonging to the specified archer.
func (s *Service) RevokeAllSessions(ctx context.Context, archerID uuid.UUID) error {
	if archerID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archerID cannot be nil")
	}
	if err := s.sessions.DeleteByArcherID(ctx, archerID); err != nil {
		return fmt.Errorf("deleting all sessions for archer: %w", err)
	}
	return nil
}

// VerifyGoogleToken verifies a Google One Tap credential using the configured client ID.
func (s *Service) VerifyGoogleToken(ctx context.Context, credential string) (*GoogleUserData, error) {
	return VerifyGoogleIDToken(ctx, credential, s.cfg.GoogleOAuthClientID)
}

func (s *Service) createSessionAndToken(
	ctx context.Context,
	archer *model.ArcherRead,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	rawSession, err := GenerateSessionToken(s.cfg.SessionTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	tokenHash := HashSessionToken(rawSession)
	expiresAt := now.Add(time.Duration(s.cfg.JWTTTLMinutes) * time.Minute)

	var (
		ua *string
		ip *string
	)
	if len(meta) > 0 {
		ua = meta[0].UserAgent
		ip = meta[0].IPAddress
	}

	sessionCreate := model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UA:               ua,
		IPInet:           ip,
	}

	if err := s.sessions.Create(ctx, sessionCreate); err != nil {
		return nil, fmt.Errorf("creating auth session: %w", err)
	}

	sid := EncodeSessionID(rawSession)
	jwtToken, err := BuildJWT(archer.ArcherID, sid, now, expiresAt, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("building jwt: %w", err)
	}

	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: jwtToken,
		ExpiresAt:   expiresAt,
		Archer:      *archer,
	}, nil
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test -race ./internal/auth/... -v
```
Expected: PASS with 100% passing tests for all auth packages.

- [x] **Step 5: Commit Service layer implementation**

```bash
git add backend/internal/auth/service.go backend/internal/auth/service_test.go
git commit -m "feat(auth): add auth service orchestrating login, registration, and session validation"
```

---

### Task 6: Full Verification & Code Quality Gates

**Files:**
- Modify: any files needing formatting or linter adjustments

**Interfaces:**
- Consumes: complete `internal/auth` package
- Produces: formatted, fully passing, cleanly building code

- [x] **Step 1: Run unit tests with race detector**

```bash
cd backend
go test -race ./internal/auth/... -v
```
Expected: All tests pass.

- [x] **Step 2: Run all backend tests**

```bash
cd backend
go test -race ./... -v
```
Expected: All packages pass without regressions.

- [x] **Step 3: Run go vet and go build**

```bash
cd backend
go vet ./...
go build ./...
```
Expected: No vet warnings; build succeeds with exit code 0.

- [x] **Step 4: Run linter and code formatter**

```bash
cd backend
golangci-lint run ./...
```
Expected: Clean output with 0 lint errors.

- [x] **Step 5: Verify git status is clean**

```bash
git status
```
Expected: Clean working tree, all changes committed on `refactor/017-auth-google-onetap-and-jwt`.
