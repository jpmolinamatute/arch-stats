# Auth Flow Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement integration tests for the full authentication flow against a real PostgreSQL 17 database (`backend/tests/integration/auth_flow_test.go`), verifying session lifecycle, JWT round-trip with DB, multiple session isolation/cleanup, expired session pruning, token hash consistency, and auth service orchestration, and mark Task 039 as completed in task tracking documents.

**Architecture:** Integration tests reside in `backend/tests/integration/` under `package integration_test`, using the shared `*pgxpool.Pool` provisioned by `TestMain` via `testcontainers-go`. Tests execute against a real PostgreSQL 17 container with Goose migrations applied, exercising repository layer (`ArcherRepo`, `AuthSessionRepo`) and auth domain services/utilities (`auth.GenerateSessionToken`, `auth.HashSessionToken`, `auth.EncodeSessionID`, `auth.DecodeSessionID`, `auth.BuildJWT`, `auth.DecodeJWT`, and `auth.Service`). Every test method registers table truncation in `t.Cleanup` to ensure hermetic isolation.

**Tech Stack:** Go 1.27+, `pgx/v5` (`pgxpool`), `testcontainers-go` (v0.44+), PostgreSQL 17 container, `golang-jwt/jwt/v5`, `google/uuid`.

**Spec:** [docs/go_refactor/tasks/039-integration_tests_auth_flow.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/039-integration_tests_auth_flow.md)

## Global Constraints

- Target branch: `refactor/039-integration-tests-auth-flow`
- File to create: `backend/tests/integration/auth_flow_test.go`
- Package declaration: `package integration_test`
- Every test function must register table cleanup via `t.Cleanup(func() { _ = truncateAll(ctx, testPool) })`
- Tests must execute against real PostgreSQL database in `testcontainers-go`, validating actual SQL execution and DB constraints
- All tests must pass with `go test ./tests/integration/... -v -count=1 -run Auth` and `go vet ./...` reporting zero issues
- At the end of implementation, mark all checklist items in [docs/go_refactor/tasks/039-integration_tests_auth_flow.md](file:///home/juanpa/Projects/arch-stats/docs/go_refactor/tasks/039-integration_tests_auth_flow.md) as completed (`[x]`) and update [docs/plans/task.md](file:///home/juanpa/Projects/arch-stats/docs/plans/task.md) marking all tasks as `DONE`

---

### Task 1: Git Branch Setup & Plan Tracking

**Files:**
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes: none
- Produces: Live task tracking table in `docs/plans/task.md` and feature branch `refactor/039-integration-tests-auth-flow`

- [ ] **Step 1: Check out feature branch**

```bash
git checkout -b refactor/039-integration-tests-auth-flow
```

- [ ] **Step 2: Initialize live task tracker in `docs/plans/task.md`**

Write table tracker in `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Plan Tracking | IN_PROGRESS | Check out branch `refactor/039-integration-tests-auth-flow` and initialize task tracker |
| Task 2: Session Lifecycle & Token Hash Consistency Tests | TODO | Implement `TestAuthFlow_SessionLifecycle` and `TestAuthFlow_SessionTokenHashConsistency` |
| Task 3: Multiple Sessions & Expired Cleanup Tests | TODO | Implement `TestAuthFlow_MultipleSessions` and `TestAuthFlow_ExpiredSessionCleanup` |
| Task 4: JWT Round-Trip with DB & Service End-to-End Tests | TODO | Implement `TestAuthFlow_JWTRoundTripWithDB` and `TestAuthFlow_ServiceEndToEnd` |
| Task 5: Full Verification, Documentation Updates & Commit | TODO | Run test suite, verify linting/vet, mark Task 039 items done, and commit |
```

- [ ] **Step 3: Verify git branch status**

Run: `git branch --show-current`
Expected: `refactor/039-integration-tests-auth-flow`

---

### Task 2: Session Lifecycle & Token Hash Consistency Tests

**Files:**
- Create: `backend/tests/integration/auth_flow_test.go`

**Interfaces:**
- Consumes:
  - `auth.GenerateSessionToken(numBytes int) ([]byte, error)`
  - `auth.HashSessionToken(raw []byte) []byte`
  - `repository.NewArcherRepo(pool *pgxpool.Pool) *ArcherRepo`
  - `repository.NewAuthSessionRepo(pool *pgxpool.Pool) *AuthSessionRepo`
  - `createTestArcher(ctx context.Context, pool *pgxpool.Pool, overrides ...ArcherOverride) (*model.ArcherRead, error)`
  - `truncateAll(ctx context.Context, pool *pgxpool.Pool) error`
- Produces:
  - `TestAuthFlow_SessionLifecycle(t *testing.T)`
  - `TestAuthFlow_SessionTokenHashConsistency(t *testing.T)`

- [ ] **Step 1: Write initial test file with `TestAuthFlow_SessionLifecycle` and `TestAuthFlow_SessionTokenHashConsistency`**

Create `backend/tests/integration/auth_flow_test.go`:

```go
package integration_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestAuthFlow_SessionLifecycle(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archerRepo := repository.NewArcherRepo(testPool)
	authRepo := repository.NewAuthSessionRepo(testPool)

	// 1. Create an archer profile in PostgreSQL
	archer, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Email = "auth-lifecycle@example.com"
		a.GoogleSubject = "google-sub-lifecycle-test"
	})
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// Verify archer exists
	dbArcher, err := archerRepo.FindByID(ctx, archer.ArcherID)
	if err != nil || dbArcher == nil {
		t.Fatalf("archerRepo.FindByID failed: %v", err)
	}

	// 2. Generate cryptographically secure session token and hash
	rawToken, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}
	if len(rawToken) != 32 {
		t.Fatalf("GenerateSessionToken returned %d bytes, want 32", len(rawToken))
	}
	tokenHash := auth.HashSessionToken(rawToken)

	// 3. Store auth session in PostgreSQL
	ua := "Mozilla/5.0 (IntegrationTest-Lifecycle)"
	ip := "192.168.1.100"
	now := time.Now().UTC().Truncate(time.Microsecond)
	expiresAt := now.Add(24 * time.Hour)

	err = authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UA:               &ua,
		IPInet:           &ip,
	})
	if err != nil {
		t.Fatalf("authRepo.Create failed: %v", err)
	}

	// 4. Find session by token hash
	session, err := authRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("FindByTokenHash failed: %v", err)
	}
	if session == nil {
		t.Fatal("expected auth session, got nil")
	}
	if session.ArcherID != archer.ArcherID {
		t.Errorf("session.ArcherID = %v, want %v", session.ArcherID, archer.ArcherID)
	}
	if !bytes.Equal(session.SessionTokenHash, tokenHash) {
		t.Errorf("session.SessionTokenHash mismatch")
	}
	if session.RevokedAt != nil {
		t.Errorf("expected session not revoked, got %v", session.RevokedAt)
	}

	// 5. Delete session and verify it is removed
	err = authRepo.DeleteByArcherID(ctx, archer.ArcherID)
	if err != nil {
		t.Fatalf("DeleteByArcherID failed: %v", err)
	}

	deleted, err := authRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("FindByTokenHash after delete returned error: %v", err)
	}
	if deleted != nil {
		t.Errorf("expected session to be deleted, but still found: %+v", deleted)
	}
}

func TestAuthFlow_SessionTokenHashConsistency(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	authRepo := repository.NewAuthSessionRepo(testPool)

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// 1. Generate token and hash
	rawToken, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}
	hash1 := auth.HashSessionToken(rawToken)

	// 2. Store session in PostgreSQL using hash1
	expiresAt := time.Now().UTC().Add(time.Hour)
	err = authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: hash1,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		t.Fatalf("authRepo.Create failed: %v", err)
	}

	// 3. Re-hash the exact same raw token bytes
	hash2 := auth.HashSessionToken(rawToken)
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("HashSessionToken produced inconsistent digests for identical raw token bytes")
	}

	// 4. Find session using hash2
	session, err := authRepo.FindByTokenHash(ctx, hash2)
	if err != nil {
		t.Fatalf("FindByTokenHash(hash2) failed: %v", err)
	}
	if session == nil {
		t.Fatal("expected session to be found using regenerated hash")
	}
	if session.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", session.ArcherID, archer.ArcherID)
	}

	// 5. Verify tampered raw token produces a different hash that finds nothing in DB
	tampered := make([]byte, len(rawToken))
	copy(tampered, rawToken)
	tampered[0] ^= 0xFF
	tamperedHash := auth.HashSessionToken(tampered)
	if bytes.Equal(tamperedHash, hash1) {
		t.Fatal("tampered token hash unexpectedly matched original hash")
	}

	notFound, err := authRepo.FindByTokenHash(ctx, tamperedHash)
	if err != nil {
		t.Fatalf("unexpected error searching for tampered hash: %v", err)
	}
	if notFound != nil {
		t.Errorf("expected nil for tampered token hash, but found session: %+v", notFound)
	}
}
```

- [ ] **Step 2: Run test suite to verify tests pass**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run "TestAuthFlow_SessionLifecycle|TestAuthFlow_SessionTokenHashConsistency"`
Expected: Both tests PASS.

- [ ] **Step 3: Update `docs/plans/task.md`**

Update Task 2 status to `DONE` and Task 3 status to `IN_PROGRESS`.

---

### Task 3: Multiple Sessions & Expired Cleanup Tests

**Files:**
- Modify: `backend/tests/integration/auth_flow_test.go`

**Interfaces:**
- Consumes:
  - `authRepo.Create`
  - `authRepo.FindByTokenHash`
  - `authRepo.DeleteByArcherID`
  - `authRepo.DeleteExpired`
  - `createTestArcher`
  - `truncateAll`
- Produces:
  - `TestAuthFlow_MultipleSessions(t *testing.T)`
  - `TestAuthFlow_ExpiredSessionCleanup(t *testing.T)`

- [ ] **Step 1: Append `TestAuthFlow_MultipleSessions` and `TestAuthFlow_ExpiredSessionCleanup` to `backend/tests/integration/auth_flow_test.go`**

Append to `backend/tests/integration/auth_flow_test.go`:

```go
func TestAuthFlow_MultipleSessions(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	authRepo := repository.NewAuthSessionRepo(testPool)

	// Create primary archer and control archer
	archer1, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher 1 failed: %v", err)
	}
	archer2, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher 2 failed: %v", err)
	}

	// 1. Create 3 distinct sessions for archer 1
	const numSessions = 3
	hashes := make([][]byte, numSessions)
	expiresAt := time.Now().UTC().Add(2 * time.Hour)

	for i := 0; i < numSessions; i++ {
		raw, err := auth.GenerateSessionToken(32)
		if err != nil {
			t.Fatalf("GenerateSessionToken(%d) failed: %v", i, err)
		}
		h := auth.HashSessionToken(raw)
		hashes[i] = h

		err = authRepo.Create(ctx, model.AuthSessionCreate{
			ArcherID:         archer1.ArcherID,
			SessionTokenHash: h,
			ExpiresAt:        expiresAt,
		})
		if err != nil {
			t.Fatalf("Create session %d failed: %v", i, err)
		}
	}

	// 2. Create 1 session for control archer 2
	controlRaw, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken(control) failed: %v", err)
	}
	controlHash := auth.HashSessionToken(controlRaw)
	err = authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer2.ArcherID,
		SessionTokenHash: controlHash,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		t.Fatalf("Create control session failed: %v", err)
	}

	// 3. Verify all 3 sessions for archer 1 and 1 session for archer 2 exist
	for i, h := range hashes {
		s, err := authRepo.FindByTokenHash(ctx, h)
		if err != nil || s == nil {
			t.Fatalf("session %d missing before delete: err=%v, session=%+v", i, err, s)
		}
	}
	ctrl, err := authRepo.FindByTokenHash(ctx, controlHash)
	if err != nil || ctrl == nil {
		t.Fatalf("control session missing before delete: err=%v, session=%+v", err, ctrl)
	}

	// 4. Delete all sessions for archer 1
	if err := authRepo.DeleteByArcherID(ctx, archer1.ArcherID); err != nil {
		t.Fatalf("DeleteByArcherID failed: %v", err)
	}

	// 5. Verify all 3 sessions for archer 1 are gone
	for i, h := range hashes {
		s, err := authRepo.FindByTokenHash(ctx, h)
		if err != nil {
			t.Fatalf("FindByTokenHash(%d) after delete returned error: %v", i, err)
		}
		if s != nil {
			t.Errorf("session %d still exists after DeleteByArcherID: %+v", i, s)
		}
	}

	// 6. Verify archer 2's session is untouched
	stillPresent, err := authRepo.FindByTokenHash(ctx, controlHash)
	if err != nil || stillPresent == nil {
		t.Fatalf("control archer session was unexpectedly deleted: %v", err)
	}
	if stillPresent.ArcherID != archer2.ArcherID {
		t.Errorf("control session ArcherID = %v, want %v", stillPresent.ArcherID, archer2.ArcherID)
	}
}

func TestAuthFlow_ExpiredSessionCleanup(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	authRepo := repository.NewAuthSessionRepo(testPool)

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// 1. Create 2 expired sessions (past expiry)
	expired1Raw, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	expired1Hash := auth.HashSessionToken(expired1Raw)
	if err := authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: expired1Hash,
		ExpiresAt:        time.Now().UTC().Add(-2 * time.Hour),
	}); err != nil {
		t.Fatalf("create expired session 1 failed: %v", err)
	}

	expired2Raw, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	expired2Hash := auth.HashSessionToken(expired2Raw)
	if err := authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: expired2Hash,
		ExpiresAt:        time.Now().UTC().Add(-10 * time.Minute),
	}); err != nil {
		t.Fatalf("create expired session 2 failed: %v", err)
	}

	// 2. Create 1 active session (future expiry)
	activeRaw, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	activeHash := auth.HashSessionToken(activeRaw)
	if err := authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: activeHash,
		ExpiresAt:        time.Now().UTC().Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("create active session failed: %v", err)
	}

	// 3. Call DeleteExpired
	deletedCount, err := authRepo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired() failed: %v", err)
	}
	if deletedCount != 2 {
		t.Errorf("deletedCount = %d, want 2", deletedCount)
	}

	// 4. Verify both expired sessions are gone
	for idx, h := range [][]byte{expired1Hash, expired2Hash} {
		exp, err := authRepo.FindByTokenHash(ctx, h)
		if err != nil {
			t.Fatalf("FindByTokenHash(expired %d) failed: %v", idx+1, err)
		}
		if exp != nil {
			t.Errorf("expired session %d should be deleted, but still found: %+v", idx+1, exp)
		}
	}

	// 5. Verify active session remains
	act, err := authRepo.FindByTokenHash(ctx, activeHash)
	if err != nil || act == nil {
		t.Fatalf("active session should still exist: err=%v, session=%+v", err, act)
	}
	if act.ArcherID != archer.ArcherID {
		t.Errorf("active session ArcherID = %v, want %v", act.ArcherID, archer.ArcherID)
	}
}
```

- [ ] **Step 2: Run test suite to verify tests pass**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run "TestAuthFlow_MultipleSessions|TestAuthFlow_ExpiredSessionCleanup"`
Expected: Both tests PASS.

- [ ] **Step 3: Update `docs/plans/task.md`**

Update Task 3 status to `DONE` and Task 4 status to `IN_PROGRESS`.

---

### Task 4: JWT Round-Trip with DB & Service End-to-End Tests

**Files:**
- Modify: `backend/tests/integration/auth_flow_test.go`

**Interfaces:**
- Consumes:
  - `auth.GenerateSessionToken(numBytes int) ([]byte, error)`
  - `auth.HashSessionToken(raw []byte) []byte`
  - `auth.EncodeSessionID(raw []byte) string`
  - `auth.DecodeSessionID(sid string) ([]byte, error)`
  - `auth.BuildJWT(archerID uuid.UUID, sid string, issuedAt, expiresAt time.Time, secret, algorithm string) (string, error)`
  - `auth.DecodeJWT(tokenStr, secret, algorithm string) (*auth.Claims, error)`
  - `auth.NewService(archers auth.ArcherRepository, sessions auth.SessionRepository, cfg auth.Config) *auth.Service`
  - `service.Authenticate(ctx context.Context, tokenStr string) (uuid.UUID, error)`
  - `service.RevokeToken(ctx context.Context, tokenStr string) error`
  - `service.Register(ctx context.Context, payload model.AuthRegistrationRequest, googleData *auth.GoogleUserData, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)`
  - `service.LoginExisting(ctx context.Context, archer *model.ArcherRead, googleData *auth.GoogleUserData, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)`
- Produces:
  - `TestAuthFlow_JWTRoundTripWithDB(t *testing.T)`
  - `TestAuthFlow_ServiceEndToEnd(t *testing.T)`

- [ ] **Step 1: Append `TestAuthFlow_JWTRoundTripWithDB` and `TestAuthFlow_ServiceEndToEnd` to `backend/tests/integration/auth_flow_test.go`**

Ensure imports in `backend/tests/integration/auth_flow_test.go` include `"fmt"` and `"github.com/google/uuid"`.

Append to `backend/tests/integration/auth_flow_test.go`:

```go
func TestAuthFlow_JWTRoundTripWithDB(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archerRepo := repository.NewArcherRepo(testPool)
	authRepo := repository.NewAuthSessionRepo(testPool)

	// 1. Create archer
	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// 2. Generate raw session token bytes & hash
	rawSession, err := auth.GenerateSessionToken(32)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}
	tokenHash := auth.HashSessionToken(rawSession)

	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(2 * time.Hour)

	// 3. Store session in PostgreSQL
	err = authRepo.Create(ctx, model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		t.Fatalf("authRepo.Create failed: %v", err)
	}

	// 4. Build JWT embedding archer ID and base64url encoded session ID
	sid := auth.EncodeSessionID(rawSession)
	secret := "integration-test-secret-key-32bytes!"
	jwtToken, err := auth.BuildJWT(archer.ArcherID, sid, now, expiresAt, secret, "HS256")
	if err != nil {
		t.Fatalf("BuildJWT failed: %v", err)
	}

	// 5. Decode JWT and extract claims
	claims, err := auth.DecodeJWT(jwtToken, secret, "HS256")
	if err != nil {
		t.Fatalf("DecodeJWT failed: %v", err)
	}

	extractedArcherID, err := claims.ArcherID()
	if err != nil {
		t.Fatalf("claims.ArcherID failed: %v", err)
	}
	if extractedArcherID != archer.ArcherID {
		t.Errorf("claims.ArcherID() = %v, want %v", extractedArcherID, archer.ArcherID)
	}
	if claims.SID != sid {
		t.Errorf("claims.SID = %q, want %q", claims.SID, sid)
	}

	// 6. Decode session ID from JWT claims back to raw bytes and compute token hash
	decodedRaw, err := auth.DecodeSessionID(claims.SID)
	if err != nil {
		t.Fatalf("DecodeSessionID failed: %v", err)
	}
	if !bytes.Equal(decodedRaw, rawSession) {
		t.Fatalf("decoded raw bytes do not match original raw session")
	}

	decodedHash := auth.HashSessionToken(decodedRaw)
	if !bytes.Equal(decodedHash, tokenHash) {
		t.Fatalf("decoded token hash does not match original stored token hash")
	}

	// 7. Verify session exists in PostgreSQL and matches the archer
	dbSession, err := authRepo.FindByTokenHash(ctx, decodedHash)
	if err != nil {
		t.Fatalf("FindByTokenHash failed: %v", err)
	}
	if dbSession == nil {
		t.Fatal("expected session in DB for decoded token hash, got nil")
	}
	if dbSession.ArcherID != archer.ArcherID {
		t.Errorf("dbSession.ArcherID = %v, want %v", dbSession.ArcherID, archer.ArcherID)
	}

	// 8. Verify end-to-end authentication via auth.Service against real database
	svc := auth.NewService(archerRepo, authRepo, auth.Config{
		JWTSecret:         secret,
		JWTAlgorithm:      "HS256",
		JWTTTLMinutes:     120,
		SessionTokenBytes: 32,
	})

	authenticatedID, err := svc.Authenticate(ctx, jwtToken)
	if err != nil {
		t.Fatalf("svc.Authenticate failed: %v", err)
	}
	if authenticatedID != archer.ArcherID {
		t.Errorf("svc.Authenticate returned archer ID %v, want %v", authenticatedID, archer.ArcherID)
	}
}

func TestAuthFlow_ServiceEndToEnd(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })

	archerRepo := repository.NewArcherRepo(testPool)
	authRepo := repository.NewAuthSessionRepo(testPool)

	secret := "service-e2e-integration-secret-key-32b"
	svc := auth.NewService(archerRepo, authRepo, auth.Config{
		JWTSecret:         secret,
		JWTAlgorithm:      "HS256",
		JWTTTLMinutes:     60,
		SessionTokenBytes: 32,
	})

	// 1. Register a new archer via auth.Service
	uniqueID := uuid.New().String()
	googleSub := fmt.Sprintf("google-e2e-%s", uniqueID)
	email := fmt.Sprintf("e2e-%s@example.com", uniqueID[:8])
	firstName := "Morgan"
	lastName := "LeFay"
	googlePicture := "https://example.com/avatar.png"

	googleData := &auth.GoogleUserData{
		Sub:        googleSub,
		Email:      email,
		GivenName:  firstName,
		FamilyName: lastName,
		Picture:    googlePicture,
	}

	regPayload := model.AuthRegistrationRequest{
		FirstName:   &firstName,
		LastName:    &lastName,
		DateOfBirth: "1992-07-20",
		Gender:      model.GenderFemale,
		Bowstyle:    model.BowstyleBarebow,
		DrawWeight:  38.5,
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	authResp, err := svc.Register(ctx, regPayload, googleData, now)
	if err != nil {
		t.Fatalf("svc.Register failed: %v", err)
	}

	if authResp == nil {
		t.Fatal("expected authResp, got nil")
	}
	if authResp.Status != model.AuthStatusAuthenticated {
		t.Errorf("Status = %v, want %v", authResp.Status, model.AuthStatusAuthenticated)
	}
	if authResp.Archer.Email != email {
		t.Errorf("Email = %q, want %q", authResp.Archer.Email, email)
	}
	if authResp.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}

	// Verify archer persisted in PostgreSQL
	persistedArcher, err := archerRepo.FindByID(ctx, authResp.Archer.ArcherID)
	if err != nil || persistedArcher == nil {
		t.Fatalf("archer not found in DB: %v", err)
	}
	if persistedArcher.GoogleSubject != googleSub {
		t.Errorf("GoogleSubject = %q, want %q", persistedArcher.GoogleSubject, googleSub)
	}

	// 2. Authenticate using the minted JWT token
	authenticatedID, err := svc.Authenticate(ctx, authResp.AccessToken)
	if err != nil {
		t.Fatalf("svc.Authenticate failed: %v", err)
	}
	if authenticatedID != authResp.Archer.ArcherID {
		t.Errorf("authenticatedID = %v, want %v", authenticatedID, authResp.Archer.ArcherID)
	}

	// 3. Revoke the token and verify subsequent Authenticate fails
	if err := svc.RevokeToken(ctx, authResp.AccessToken); err != nil {
		t.Fatalf("svc.RevokeToken failed: %v", err)
	}

	revokedID, err := svc.Authenticate(ctx, authResp.AccessToken)
	if err == nil {
		t.Errorf("expected Authenticate to fail on revoked token, but got archer ID %v", revokedID)
	}

	// 4. Log in the existing archer again
	loginResp, err := svc.LoginExisting(ctx, persistedArcher, googleData, time.Now().UTC())
	if err != nil {
		t.Fatalf("svc.LoginExisting failed: %v", err)
	}
	if loginResp.AccessToken == "" {
		t.Fatal("expected non-empty access token after LoginExisting")
	}

	// 5. Authenticate with the newly issued access token
	newAuthID, err := svc.Authenticate(ctx, loginResp.AccessToken)
	if err != nil {
		t.Fatalf("svc.Authenticate with fresh token failed: %v", err)
	}
	if newAuthID != authResp.Archer.ArcherID {
		t.Errorf("newAuthID = %v, want %v", newAuthID, authResp.Archer.ArcherID)
	}
}
```

- [ ] **Step 2: Run all tests in `auth_flow_test.go`**

Run: `cd backend && go test ./tests/integration/... -v -count=1 -run TestAuthFlow`
Expected: All 6 `TestAuthFlow_*` tests PASS.

- [ ] **Step 3: Update `docs/plans/task.md`**

Update Task 4 status to `DONE` and Task 5 status to `IN_PROGRESS`.

---

### Task 5: Full Verification, Documentation Updates & Commit

**Files:**
- Modify: `docs/go_refactor/tasks/039-integration_tests_auth_flow.md`
- Modify: `docs/plans/task.md`

**Interfaces:**
- Consumes:
  - All test functions in `backend/tests/integration/auth_flow_test.go`
  - Integration test suite and linter scripts
- Produces:
  - Completed checklist in `docs/go_refactor/tasks/039-integration_tests_auth_flow.md`
  - Fully completed `docs/plans/task.md`
  - Git commit on branch `refactor/039-integration-tests-auth-flow`

- [ ] **Step 1: Run integration tests with `-run Auth` filter**

Run:
```bash
cd backend && go test ./tests/integration/... -v -count=1 -run Auth
```
Expected: All `TestAuthFlow_*` and `TestAuthSessionRepo_*` tests PASS.

- [ ] **Step 2: Run full integration test suite with race detector**

Run:
```bash
cd backend && go test -race ./tests/integration/... -v -count=1
```
Expected: All integration tests pass cleanly with race detection enabled.

- [ ] **Step 3: Run `go vet` and full Go linting suite**

Run:
```bash
cd backend && go vet ./...
./scripts/linting.bash --go
```
Expected: 0 issues reported.

- [ ] **Step 4: Mark Task 039 as completed in `docs/go_refactor/tasks/039-integration_tests_auth_flow.md`**

Update `docs/go_refactor/tasks/039-integration_tests_auth_flow.md` marking all acceptance criteria checkboxes and steps as `[x]`:

```markdown
## Acceptance Criteria

- [x] `backend/tests/integration/auth_flow_test.go` tests the following scenarios:
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
- [x] Each test truncates tables after completion.
- [x] `go test ./tests/integration/... -v -count=1 -run Auth` passes.
- [x] `go vet ./...` reports no issues.

...

## Steps

- [x] **Step 1: Write auth session lifecycle test**
- [x] **Step 2: Write JWT round-trip with DB test**
- [x] **Step 3: Write multiple sessions test**
- [x] **Step 4: Write expired session cleanup test**
- [x] **Step 5: Write token hash consistency test**
- [x] **Step 6: Run integration tests**
- [x] **Step 7: Run go vet**
- [x] **Step 8: Commit**
```

- [ ] **Step 5: Mark all tasks as `DONE` in `docs/plans/task.md`**

Update `docs/plans/task.md`:

```markdown
| Task | Status | Description |
| --- | --- | --- |
| Task 1: Git Branch Setup & Plan Tracking | DONE | Check out branch `refactor/039-integration-tests-auth-flow` and initialize task tracker |
| Task 2: Session Lifecycle & Token Hash Consistency Tests | DONE | Implement `TestAuthFlow_SessionLifecycle` and `TestAuthFlow_SessionTokenHashConsistency` |
| Task 3: Multiple Sessions & Expired Cleanup Tests | DONE | Implement `TestAuthFlow_MultipleSessions` and `TestAuthFlow_ExpiredSessionCleanup` |
| Task 4: JWT Round-Trip with DB & Service End-to-End Tests | DONE | Implement `TestAuthFlow_JWTRoundTripWithDB` and `TestAuthFlow_ServiceEndToEnd` |
| Task 5: Full Verification, Documentation Updates & Commit | DONE | Run test suite, verify linting/vet, mark Task 039 items done, and commit |
```

- [ ] **Step 6: Commit all changes to git**

Run:
```bash
git add backend/tests/integration/auth_flow_test.go docs/go_refactor/tasks/039-integration_tests_auth_flow.md docs/plans/task.md docs/plans/2026-09-06-integration-tests-auth-flow.md
git commit -m "test: add auth flow integration tests with real PostgreSQL"
```
Expected: Clean commit on branch `refactor/039-integration-tests-auth-flow`.
