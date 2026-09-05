package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// AuthCookieName defines the cookie key for authenticated sessions.
const AuthCookieName = "arch_stats_auth"

// TokenAuthenticator defines the contract for authenticating access tokens.
type TokenAuthenticator interface {
	Authenticate(ctx context.Context, token string) (uuid.UUID, error)
}

// Auth constructs an HTTP middleware that extracts the authentication token
// from either the Authorization header or the arch_stats_auth cookie, validates it
// with the authenticator, and injects the authenticated archer ID into the request context.
func Auth(authenticator TokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				WriteError(w, apperror.Wrap(apperror.ErrUnauthorized, "missing authentication token"))
				return
			}

			archerID, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				WriteError(w, apperror.Wrap(apperror.ErrUnauthorized, err.Error()))
				return
			}

			ctx := WithArcherID(r.Context(), archerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractToken extracts the bearer token from either the Authorization header
// or the arch_stats_auth cookie.
func ExtractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if trimmed := strings.TrimSpace(parts[1]); trimmed != "" {
				return trimmed
			}
		}
	}

	if cookie, err := r.Cookie(AuthCookieName); err == nil {
		if trimmed := strings.TrimSpace(cookie.Value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
