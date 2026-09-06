package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestEndpointSecurity_CloseSessionForbiddenForNonOwner(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	owner, _, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("create owner failed: %v", err)
	}

	_, strangerToken, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("create stranger failed: %v", err)
	}

	// Owner creates a session
	sess, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	// Stranger attempts to close owner's session
	payload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", strangerToken, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 Forbidden", resp.StatusCode)
	}
}

func TestEndpointSecurity_ExpiredJWT(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	// Generate expired JWT
	expiredToken, err := auth.BuildJWT(
		archer.ArcherID,
		"expired-sid",
		time.Now().UTC().Add(-2*time.Hour),
		time.Now().UTC().Add(-1*time.Hour),
		testJWTSecret,
		"HS256",
	)
	if err != nil {
		t.Fatalf("BuildJWT failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", expiredToken, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for expired JWT", resp.StatusCode)
	}
}

func TestEndpointSecurity_MalformedJWT(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	malformedTokens := []string{
		"not-a-jwt",
		"header.payload.signature.extra",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.badpayload.badsig",
	}

	for _, tok := range malformedTokens {
		req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", tok, nil)
		resp, _ := doJSONRequest(t, nil, req, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("token %q status = %d, want 401 Unauthorized", tok, resp.StatusCode)
		}
	}
}

func TestEndpointSecurity_MissingAuthCookie(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", http.NoBody)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for missing auth cookie", resp.StatusCode)
	}
}

func TestEndpointSecurity_RevokedSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// Revoke session in database
	authSessionRepo := repository.NewAuthSessionRepo(testPool)
	if err := authSessionRepo.DeleteByArcherID(ctx, archer.ArcherID); err != nil {
		t.Fatalf("DeleteByArcherID failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 Unauthorized for revoked DB session", resp.StatusCode)
	}
}
