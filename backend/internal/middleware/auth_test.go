package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

type mockAuthenticator struct {
	authenticateFn func(ctx context.Context, token string) (uuid.UUID, error)
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, token string) (uuid.UUID, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "unimplemented mock")
}

func TestAuth_MissingTokenReturns401(t *testing.T) {
	authMw := middleware.Auth(&mockAuthenticator{})
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var errResp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if errResp.Code != "UNAUTHORIZED" {
		t.Errorf("error code = %q, want UNAUTHORIZED", errResp.Code)
	}
}

func TestAuth_InvalidTokenReturns401(t *testing.T) {
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid token signature")
		},
	})
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", http.NoBody)
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "bad-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidCookieAttachesArcherID(t *testing.T) {
	expectedID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "valid-cookie-token" {
				return expectedID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := middleware.GetArcherID(r.Context())
		if err != nil {
			t.Errorf("GetArcherID returned error: %v", err)
		}
		receivedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", http.NoBody)
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "valid-cookie-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != expectedID {
		t.Errorf("receivedID = %v, want %v", receivedID, expectedID)
	}
}

func TestAuth_ValidAuthorizationHeaderAttachesArcherID(t *testing.T) {
	expectedID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "valid-header-token" {
				return expectedID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := middleware.GetArcherID(r.Context())
		if err != nil {
			t.Errorf("GetArcherID returned error: %v", err)
		}
		receivedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-header-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != expectedID {
		t.Errorf("receivedID = %v, want %v", receivedID, expectedID)
	}
}

func TestAuth_HeaderTakesPrecedenceOverCookie(t *testing.T) {
	headerID := uuid.New()
	authMw := middleware.Auth(&mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			if token == "header-token" {
				return headerID, nil
			}
			return uuid.Nil, errors.New("unexpected token")
		},
	})

	var receivedID uuid.UUID
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID, _ = middleware.GetArcherID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/archer/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: middleware.AuthCookieName, Value: "cookie-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if receivedID != headerID {
		t.Errorf("receivedID = %v, want %v", receivedID, headerID)
	}
}

func TestExtractToken(t *testing.T) {
	t.Run("extracts from Authorization header with Bearer prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer header-token-xyz")

		got := middleware.ExtractToken(req)
		if got != "header-token-xyz" {
			t.Fatalf("expected 'header-token-xyz', got %q", got)
		}
	})

	t.Run("extracts from cookie when Authorization header is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "cookie-token-abc",
		})

		got := middleware.ExtractToken(req)
		if got != "cookie-token-abc" {
			t.Fatalf("expected 'cookie-token-abc', got %q", got)
		}
	})

	t.Run("prefers Authorization header over cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer header-priority-token")
		req.AddCookie(&http.Cookie{
			Name:  middleware.AuthCookieName,
			Value: "cookie-ignored-token",
		})

		got := middleware.ExtractToken(req)
		if got != "header-priority-token" {
			t.Fatalf("expected 'header-priority-token', got %q", got)
		}
	})

	t.Run("returns empty string when neither header nor cookie present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		got := middleware.ExtractToken(req)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
