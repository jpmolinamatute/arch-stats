package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

func TestVerifyGoogleIDToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clientID := "test-google-client-id.apps.googleusercontent.com"

	t.Run("successfully extracts claims from valid payload", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, idToken, audience string) (*idtoken.Payload, error) {
			if idToken != "valid-credential" || audience != clientID {
				return nil, errors.New("mismatch")
			}
			return &idtoken.Payload{
				Subject: "google-subject-12345",
				Claims: map[string]any{
					"email":          "archer@example.com",
					"email_verified": true,
					"name":           "Robin Hood",
					"given_name":     "Robin",
					"family_name":    "Hood",
					"picture":        "https://example.com/avatar.jpg",
					"locale":         "en",
					"hd":             "example.com",
				},
			}, nil
		}

		data, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "valid-credential", clientID, mockVerifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if data.Sub != "google-subject-12345" {
			t.Errorf("expected sub google-subject-12345, got %s", data.Sub)
		}
		if data.Email != "archer@example.com" {
			t.Errorf("expected email archer@example.com, got %s", data.Email)
		}
		if !data.EmailVerified {
			t.Errorf("expected email_verified true, got false")
		}
		if data.GivenName != "Robin" {
			t.Errorf("expected given_name Robin, got %s", data.GivenName)
		}
		if data.FamilyName != "Hood" {
			t.Errorf("expected family_name Hood, got %s", data.FamilyName)
		}
		if data.Picture != "https://example.com/avatar.jpg" {
			t.Errorf("expected picture url, got %s", data.Picture)
		}
	})

	t.Run("verifier error translates to unauthorized apperror", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return nil, errors.New("signature expired or invalid")
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "bad-credential", clientID, mockVerifier)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("missing subject or email in token returns validation error", func(t *testing.T) {
		t.Parallel()
		mockVerifierMissingEmail := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-subject-12345",
				Claims: map[string]any{
					"name": "No Email Archer",
				},
			}, nil
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", clientID, mockVerifierMissingEmail)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing email, got: %v", err)
		}

		mockVerifierMissingSub := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "",
				Claims: map[string]any{
					"email": "archer@example.com",
				},
			}, nil
		}

		_, err = auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", clientID, mockVerifierMissingSub)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing subject, got: %v", err)
		}
	})

	t.Run("empty credential or clientID returns validation error", func(t *testing.T) {
		t.Parallel()
		mockVerifier := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
			return nil, nil
		}

		_, err := auth.VerifyGoogleIDTokenWithVerifier(ctx, "", clientID, mockVerifier)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty credential, got: %v", err)
		}

		_, err = auth.VerifyGoogleIDTokenWithVerifier(ctx, "credential", "", mockVerifier)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty clientID, got: %v", err)
		}
	})
}

func TestBuildNeedsRegistrationResponse(t *testing.T) {
	t.Parallel()

	t.Run("correctly maps full google claims", func(t *testing.T) {
		t.Parallel()
		googleData := &auth.GoogleUserData{
			Sub:        "google-sub-456",
			Email:      "legolas@woodland.realm",
			GivenName:  "Legolas",
			FamilyName: "Greenleaf",
			Picture:    "https://lotr.realm/legolas.png",
		}

		res := auth.BuildNeedsRegistrationResponse(googleData)
		if res == nil {
			t.Fatal("expected non-nil response")
		}

		if res.Status != model.AuthStatusNeedsRegistration {
			t.Errorf("expected status %s, got %s", model.AuthStatusNeedsRegistration, res.Status)
		}
		if res.GoogleEmail != "legolas@woodland.realm" {
			t.Errorf("expected email legolas@woodland.realm, got %s", res.GoogleEmail)
		}
		if res.GoogleSubject != "google-sub-456" {
			t.Errorf("expected subject google-sub-456, got %s", res.GoogleSubject)
		}
		if res.GivenName == nil || *res.GivenName != "Legolas" {
			t.Errorf("expected given_name Legolas, got %v", res.GivenName)
		}
		if !res.GivenNameProvided {
			t.Errorf("expected GivenNameProvided true")
		}
		if res.FamilyName == nil || *res.FamilyName != "Greenleaf" {
			t.Errorf("expected family_name Greenleaf, got %v", res.FamilyName)
		}
		if !res.FamilyNameProvided {
			t.Errorf("expected FamilyNameProvided true")
		}
		if res.PictureURL == nil || *res.PictureURL != "https://lotr.realm/legolas.png" {
			t.Errorf("expected picture url, got %v", res.PictureURL)
		}
	})

	t.Run("correctly handles empty optional names", func(t *testing.T) {
		t.Parallel()
		googleData := &auth.GoogleUserData{
			Sub:   "google-sub-789",
			Email: "mystery@example.com",
		}

		res := auth.BuildNeedsRegistrationResponse(googleData)
		if res.GivenName != nil {
			t.Errorf("expected nil given_name, got %v", res.GivenName)
		}
		if res.GivenNameProvided {
			t.Errorf("expected GivenNameProvided false")
		}
		if res.FamilyName != nil {
			t.Errorf("expected nil family_name, got %v", res.FamilyName)
		}
		if res.FamilyNameProvided {
			t.Errorf("expected FamilyNameProvided false")
		}
		if res.PictureURL != nil {
			t.Errorf("expected nil picture_url, got %v", res.PictureURL)
		}
	})

	t.Run("nil googleData returns nil", func(t *testing.T) {
		t.Parallel()
		res := auth.BuildNeedsRegistrationResponse(nil)
		if res != nil {
			t.Errorf("expected nil response for nil googleData, got %v", res)
		}
	})
}
