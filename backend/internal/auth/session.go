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
