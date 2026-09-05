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
