package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

// GoogleUserData represents the verified subset of Google ID Token (OIDC) claims.
type GoogleUserData struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`
	Picture       string `json:"picture,omitempty"`
	Locale        string `json:"locale,omitempty"`
	HostedDomain  string `json:"hd,omitempty"`
}

// GooglePayloadVerifier defines the function signature for verifying Google ID tokens.
type GooglePayloadVerifier func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error)

// defaultGooglePayloadVerifier invokes Google's official idtoken.Validate validator.
var defaultGooglePayloadVerifier GooglePayloadVerifier = idtoken.Validate

// VerifyGoogleIDToken verifies a Google One Tap credential against the expected client ID
// using Google's public key sets and returns extracted user claims.
func VerifyGoogleIDToken(ctx context.Context, credential, clientID string) (*GoogleUserData, error) {
	return VerifyGoogleIDTokenWithVerifier(ctx, credential, clientID, defaultGooglePayloadVerifier)
}

// VerifyGoogleIDTokenWithVerifier verifies credentials using a pluggable verifier, allowing pure unit testing.
func VerifyGoogleIDTokenWithVerifier(
	ctx context.Context,
	credential, clientID string,
	verifier GooglePayloadVerifier,
) (*GoogleUserData, error) {
	if strings.TrimSpace(credential) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google credential cannot be empty")
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google client id cannot be empty")
	}
	if verifier == nil {
		verifier = defaultGooglePayloadVerifier
	}

	payload, err := verifier(ctx, credential, clientID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, fmt.Sprintf("invalid google credential: %v", err))
	}

	sub := strings.TrimSpace(payload.Subject)
	if sub == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google token missing subject claim")
	}

	email, _ := payload.Claims["email"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "google token missing email claim")
	}

	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	givenName, _ := payload.Claims["given_name"].(string)
	familyName, _ := payload.Claims["family_name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	locale, _ := payload.Claims["locale"].(string)
	hd, _ := payload.Claims["hd"].(string)

	return &GoogleUserData{
		Sub:           sub,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		GivenName:     givenName,
		FamilyName:    familyName,
		Picture:       picture,
		Locale:        locale,
		HostedDomain:  hd,
	}, nil
}

// BuildNeedsRegistrationResponse constructs an AuthNeedsRegistration payload from Google claims.
func BuildNeedsRegistrationResponse(googleData *GoogleUserData) *model.AuthNeedsRegistration {
	if googleData == nil {
		return nil
	}

	var (
		given   *string
		family  *string
		picture *string
	)

	if trimmed := strings.TrimSpace(googleData.GivenName); trimmed != "" {
		given = &trimmed
	}
	if trimmed := strings.TrimSpace(googleData.FamilyName); trimmed != "" {
		family = &trimmed
	}
	if trimmed := strings.TrimSpace(googleData.Picture); trimmed != "" {
		picture = &trimmed
	}

	return &model.AuthNeedsRegistration{
		Status:             model.AuthStatusNeedsRegistration,
		GoogleEmail:        googleData.Email,
		GoogleSubject:      googleData.Sub,
		GivenName:          given,
		FamilyName:         family,
		GivenNameProvided:  given != nil,
		FamilyNameProvided: family != nil,
		PictureURL:         picture,
	}
}
