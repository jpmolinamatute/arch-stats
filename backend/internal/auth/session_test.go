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
		// Standard test vector for the SHA-256 digest of "hello world".
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
