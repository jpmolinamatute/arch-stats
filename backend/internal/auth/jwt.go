package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
)

// Claims represents the JWT payload claims for an authenticated archer session,
// exactly matching the Python authentication payload format (sub, sid, exp, iat, iss, typ).
type Claims struct {
	Sub string `json:"sub"`
	SID string `json:"sid"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
	Typ string `json:"typ"`
}

// ArcherID parses and returns the Subject claim as a uuid.UUID.
func (c *Claims) ArcherID() (uuid.UUID, error) {
	id, err := uuid.Parse(c.Sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing archer id from jwt subject %q: %w", c.Sub, err)
	}
	return id, nil
}

// GetExpirationTime implements jwt.Claims interface.
func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

// GetIssuedAt implements jwt.Claims interface.
func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.Iat == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.Iat, 0)), nil
}

// GetNotBefore implements jwt.Claims interface.
func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuer implements jwt.Claims interface.
func (c *Claims) GetIssuer() (string, error) {
	return c.Iss, nil
}

// GetSubject implements jwt.Claims interface.
func (c *Claims) GetSubject() (string, error) {
	return c.Sub, nil
}

// GetAudience implements jwt.Claims interface.
func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// BuildJWT mints and signs a JWT embedding the archer UUID and session identifier.
func BuildJWT(
	archerID uuid.UUID,
	sid string,
	issuedAt, expiresAt time.Time,
	secret, algorithm string,
) (string, error) {
	if archerID == uuid.Nil {
		return "", apperror.Wrap(apperror.ErrValidation, "archerID cannot be nil")
	}
	if strings.TrimSpace(sid) == "" {
		return "", apperror.Wrap(apperror.ErrValidation, "sid cannot be empty")
	}
	if strings.TrimSpace(secret) == "" {
		return "", apperror.Wrap(apperror.ErrValidation, "jwt secret cannot be empty")
	}

	signingMethod := jwt.GetSigningMethod(algorithm)
	if signingMethod == nil || !strings.HasPrefix(algorithm, "HS") {
		return "", fmt.Errorf("unsupported HMAC signing algorithm: %s", algorithm)
	}

	claims := Claims{
		Sub: archerID.String(),
		SID: sid,
		Exp: expiresAt.Unix(),
		Iat: issuedAt.Unix(),
		Iss: "arch-stats",
		Typ: "access",
	}

	token := jwt.NewWithClaims(signingMethod, &claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return signed, nil
}

// DecodeJWT verifies and decodes an access JWT, enforcing signing algorithm, secret, and expiration.
func DecodeJWT(tokenStr, secret, algorithm string) (*Claims, error) {
	if strings.TrimSpace(tokenStr) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "token cannot be empty")
	}
	if strings.TrimSpace(secret) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "jwt secret cannot be empty")
	}

	var claims Claims
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != algorithm {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperror.Wrap(apperror.ErrUnauthorized, "token has expired")
		}
		return nil, apperror.Wrap(apperror.ErrUnauthorized, fmt.Sprintf("invalid token: %v", err))
	}

	if !parsedToken.Valid {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "token is invalid")
	}

	return &claims, nil
}
