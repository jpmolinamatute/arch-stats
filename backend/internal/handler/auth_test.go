package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

type mockAuthService struct {
	loginWithGoogleFn    func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	registerWithGoogleFn func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	revokeTokenFn        func(ctx context.Context, tokenStr string) error
	decodeTokenFn        func(tokenStr string) (*auth.Claims, error)
	authenticateFn       func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *mockAuthService) LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
	if m.loginWithGoogleFn != nil {
		return m.loginWithGoogleFn(ctx, credential, now, meta...)
	}
	return nil, nil, errors.New("unimplemented")
}

//nolint:gocritic // hugeParam: payload matches AuthService interface
func (m *mockAuthService) RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
	if m.registerWithGoogleFn != nil {
		return m.registerWithGoogleFn(ctx, payload, now, meta...)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockAuthService) RevokeToken(ctx context.Context, tokenStr string) error {
	if m.revokeTokenFn != nil {
		return m.revokeTokenFn(ctx, tokenStr)
	}
	return nil
}

func (m *mockAuthService) DecodeToken(tokenStr string) (*auth.Claims, error) {
	if m.decodeTokenFn != nil {
		return m.decodeTokenFn(tokenStr)
	}
	return nil, errors.New("unimplemented")
}

func (m *mockAuthService) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	return uuid.Nil, errors.New("unimplemented")
}

type mockArcherService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
}

func (m *mockArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("unimplemented")
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("login with valid credential for existing user returns 200 and sets cookie", func(t *testing.T) {
		archerID := uuid.New()
		expectedExpires := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return &model.AuthAuthenticated{
					Status:      model.AuthStatusAuthenticated,
					AccessToken: "valid-jwt-token-123",
					ExpiresAt:   expectedExpires,
					Archer: model.ArcherRead{
						ArcherID:  archerID,
						FirstName: "Katniss",
						LastName:  "Everdeen",
						Email:     "katniss@district12.org",
					},
				}, nil, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "valid-google-credential"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthAuthenticated
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != model.AuthStatusAuthenticated || resp.Archer.ArcherID != archerID {
			t.Fatalf("unexpected response payload: %+v", resp)
		}

		// Check cookie
		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil {
			t.Fatal("expected arch_stats_auth cookie to be set, but found none")
		}
		if authCookie.Value != "valid-jwt-token-123" {
			t.Fatalf("expected cookie value 'valid-jwt-token-123', got %q", authCookie.Value)
		}
		if !authCookie.HttpOnly {
			t.Fatal("expected cookie to be HttpOnly")
		}
		if authCookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("expected SameSite Lax, got %v", authCookie.SameSite)
		}
	})

	t.Run("login with valid credential for new user returns 200 needs registration without cookie", func(t *testing.T) {
		given := "Legolas"
		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return nil, &model.AuthNeedsRegistration{
					Status:             model.AuthStatusNeedsRegistration,
					GoogleEmail:        "legolas@woodland.realm",
					GoogleSubject:      "sub-woodland-elf",
					GivenName:          &given,
					GivenNameProvided:  true,
					FamilyNameProvided: false,
				}, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "new-user-credential"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthNeedsRegistration
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != model.AuthStatusNeedsRegistration || resp.GoogleEmail != "legolas@woodland.realm" {
			t.Fatalf("unexpected response payload: %+v", resp)
		}

		for _, c := range rec.Result().Cookies() {
			if c.Name == middleware.AuthCookieName && c.Value != "" && c.MaxAge > 0 {
				t.Fatalf("expected no auth cookie to be set for needs_registration, but got %+v", c)
			}
		}
	})

	t.Run("login with invalid credential returns 401", func(t *testing.T) {
		authSvc := &mockAuthService{
			loginWithGoogleFn: func(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error) {
				return nil, nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid google credential")
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: "invalid-cred"})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
		}

		var errResp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if !strings.Contains(errResp.Detail, "invalid google credential") {
			t.Fatalf("expected detail containing 'invalid google credential', got %q", errResp.Detail)
		}
	})

	t.Run("login with empty credential returns 422", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		body, _ := json.Marshal(model.GoogleOneTapRequest{Credential: ""})
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("register with missing fields returns 422", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		// Missing date_of_birth, gender, draw_weight <= 0
		invalidPayload := model.AuthRegistrationRequest{
			Credential: "valid-credential",
			DrawWeight: 0,
		}

		body, _ := json.Marshal(invalidPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("register with valid fields returns 201 and sets cookie", func(t *testing.T) {
		archerID := uuid.New()
		authSvc := &mockAuthService{
			registerWithGoogleFn: func(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error) {
				return &model.AuthAuthenticated{
					Status:      model.AuthStatusAuthenticated,
					AccessToken: "new-user-jwt",
					ExpiresAt:   time.Now().Add(24 * time.Hour).UTC(),
					Archer: model.ArcherRead{
						ArcherID:  archerID,
						FirstName: "Hawkeye",
						LastName:  "Pierce",
						Email:     "hawkeye@mash4077.org",
					},
				}, nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{
			JWTTTLMinutes: 1440,
			DevMode:       true,
		})

		firstName := "Hawkeye"
		lastName := "Pierce"
		validPayload := model.AuthRegistrationRequest{
			Credential:  "google-reg-token-ok",
			DateOfBirth: "1980-01-01",
			Gender:      model.GenderMale,
			Bowstyle:    model.BowstyleRecurve,
			DrawWeight:  35.0,
			FirstName:   &firstName,
			LastName:    &lastName,
		}

		body, _ := json.Marshal(validPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
			t.Fatalf("expected status 201 or 200, got %d: %s", rec.Code, rec.Body.String())
		}

		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil || authCookie.Value != "new-user-jwt" {
			t.Fatalf("expected auth cookie with value 'new-user-jwt', got %+v", authCookie)
		}
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("logout clears the session cookie and returns success", func(t *testing.T) {
		revoked := false
		authSvc := &mockAuthService{
			revokeTokenFn: func(ctx context.Context, tokenStr string) error {
				if tokenStr == "token-to-revoke" {
					revoked = true
				}
				return nil
			},
		}

		h := handler.NewAuthHandler(authSvc, &mockArcherService{}, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/logout", http.NoBody)
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "token-to-revoke",
		})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if !revoked {
			t.Fatal("expected RevokeToken to be called")
		}

		var resp model.LogoutResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Fatal("expected success: true")
		}

		// Verify cookie is cleared (MaxAge < 0)
		cookies := rec.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.AuthCookieName {
				authCookie = c
				break
			}
		}
		if authCookie == nil {
			t.Fatal("expected deleted auth cookie in response")
		}
		if authCookie.MaxAge >= 0 {
			t.Fatalf("expected cookie MaxAge < 0, got %d", authCookie.MaxAge)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("me returns the authenticated archer from context", func(t *testing.T) {
		archerID := uuid.New()
		expectedArcher := &model.ArcherRead{
			ArcherID:  archerID,
			FirstName: "Merida",
			LastName:  "DunBroch",
			Email:     "merida@dunbroch.scot",
		}

		archerSvc := &mockArcherService{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				if id == archerID {
					return expectedArcher, nil
				}
				return nil, apperror.ErrNotFound
			},
		}

		h := handler.NewAuthHandler(&mockAuthService{}, archerSvc, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/me", http.NoBody)
		req = req.WithContext(middleware.WithArcherID(req.Context(), archerID))
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp model.AuthAuthenticated
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Archer.ArcherID != archerID || resp.Archer.FirstName != "Merida" {
			t.Fatalf("unexpected archer data: %+v", resp.Archer)
		}
	})

	t.Run("me returns 401 when no auth context or cookie present", func(t *testing.T) {
		h := handler.NewAuthHandler(&mockAuthService{}, &mockArcherService{}, handler.AuthHandlerConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/me", http.NoBody)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
