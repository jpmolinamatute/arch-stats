package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

func newMockGoogleVerifier() func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
	return func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
		switch idToken {
		case "mock-google-existing":
			return &idtoken.Payload{
				Subject: "google-sub-existing",
				Claims: map[string]any{
					"email":       "existing@example.com",
					"given_name":  "Existing",
					"family_name": "User",
				},
			}, nil
		case "mock-google-new":
			return &idtoken.Payload{
				Subject: "google-sub-new",
				Claims: map[string]any{
					"email":       "newarcher@example.com",
					"given_name":  "New",
					"family_name": "Archer",
				},
			}, nil
		default:
			return nil, errors.New("invalid google id token credential")
		}
	}
}

func TestEndpointAuth_Login_ValidGoogleToken_ExistingArcher(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	created, err := createTestArcher(ctx, testPool, func(a *model.ArcherCreate) {
		a.Email = "existing@example.com"
	})
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}
	_, err = testPool.Exec(ctx, "UPDATE auth SET google_subject = $1 WHERE archer_id = $2", "google-sub-existing", created.ArcherID)
	if err != nil {
		t.Fatalf("updating auth google_subject failed: %v", err)
	}

	payload := model.GoogleOneTapRequest{Credential: "mock-google-existing"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var authResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &authResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if authResp.Status != model.AuthStatusAuthenticated {
		t.Fatalf("status = %v, want authenticated", authResp.Status)
	}
	if authResp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}

	// Verify cookie is set
	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == middleware.AuthCookieName {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatalf("expected cookie %q in response", middleware.AuthCookieName)
	}
}

func TestEndpointAuth_Login_ValidGoogleToken_NewArcherNeedsRegistration(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	payload := model.GoogleOneTapRequest{Credential: "mock-google-new"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var needsReg model.AuthNeedsRegistration
	resp, _ := doJSONRequest(t, nil, req, &needsReg)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if needsReg.Status != model.AuthStatusNeedsRegistration {
		t.Fatalf("status = %v, want needs_registration", needsReg.Status)
	}
	if needsReg.GoogleEmail != "newarcher@example.com" {
		t.Fatalf("email = %v, want newarcher@example.com", needsReg.GoogleEmail)
	}
}

func TestEndpointAuth_Login_InvalidCredential(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	payload := model.GoogleOneTapRequest{Credential: "invalid_jwt_token"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEndpointAuth_Register_Success(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	first := "New"
	last := "Archer"
	regPayload := model.AuthRegistrationRequest{
		Credential:  "mock-google-new",
		FirstName:   &first,
		LastName:    &last,
		DateOfBirth: "1995-06-20",
		Gender:      model.GenderUnspecified,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  35.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var authResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &authResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if authResp.Archer.Email != "newarcher@example.com" {
		t.Errorf("archer email = %v, want newarcher@example.com", authResp.Archer.Email)
	}
}

func TestEndpointAuth_Register_MissingFields(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	regPayload := model.AuthRegistrationRequest{
		Credential:  "mock-google-new",
		DateOfBirth: "not-a-valid-date",
		DrawWeight:  -10.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEndpointAuth_Register_InvalidCredential(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t, newMockGoogleVerifier())

	first := "Invalid"
	last := "User"
	regPayload := model.AuthRegistrationRequest{
		Credential:  "invalid_jwt_token",
		FirstName:   &first,
		LastName:    &last,
		DateOfBirth: "1990-01-01",
		Gender:      model.GenderUnspecified,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  30.0,
	}
	body, _ := json.Marshal(regPayload)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v0/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEndpointAuth_Logout_ClearsCookieAndRevokes(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/auth/logout", token, nil)
	var logoutResp model.LogoutResponse
	resp, _ := doJSONRequest(t, nil, req, &logoutResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !logoutResp.Success {
		t.Fatalf("success = false, want true")
	}

	// Verify cookie is cleared (MaxAge < 0 or empty)
	var clearedCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == middleware.AuthCookieName {
			clearedCookie = c
			break
		}
	}
	if clearedCookie == nil || clearedCookie.MaxAge > 0 {
		t.Fatalf("expected cleared cookie with MaxAge <= 0, got: %+v", clearedCookie)
	}

	// Subsequent /me request with same revoked token must fail with 401
	meReq := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	meResp, _ := doJSONRequest(t, nil, meReq, nil)
	if meResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me request after logout status = %d, want 401", meResp.StatusCode)
	}
}

func TestEndpointAuth_Me_Authenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", token, nil)
	var meResp model.AuthAuthenticated
	resp, _ := doJSONRequest(t, nil, req, &meResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if meResp.Archer.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", meResp.Archer.ArcherID, archer.ArcherID)
	}
}

func TestEndpointAuth_Me_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/auth/me", "", nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
