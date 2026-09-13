package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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
