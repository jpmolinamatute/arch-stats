package auth_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"google.golang.org/api/idtoken"
)

type mockArcherRepo struct {
	findByIDFn            func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	findByGoogleSubjectFn func(ctx context.Context, sub string) (*model.ArcherRead, error)
	createFn              func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	updateFn              func(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
}

func (m *mockArcherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockArcherRepo) FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error) {
	if m.findByGoogleSubjectFn != nil {
		return m.findByGoogleSubjectFn(ctx, sub)
	}
	return nil, nil
}

//nolint:gocritic // hugeParam: data matches ArcherRepository interface
func (m *mockArcherRepo) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return uuid.New(), nil
}

//nolint:gocritic // hugeParam: filter matches ArcherRepository interface
func (m *mockArcherRepo) Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, data, filter)
	}
	return nil
}

type mockSessionRepo struct {
	createFn            func(ctx context.Context, data model.AuthSessionCreate) error
	findByTokenHashFn   func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)
	revokeByTokenHashFn func(ctx context.Context, hash []byte, revokedAt time.Time) error
	deleteByArcherIDFn  func(ctx context.Context, archerID uuid.UUID) error
}

//nolint:gocritic // hugeParam: data matches SessionRepository interface
func (m *mockSessionRepo) Create(ctx context.Context, data model.AuthSessionCreate) error {
	if m.createFn != nil {
		return m.createFn(ctx, data)
	}
	return nil
}

func (m *mockSessionRepo) FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
	if m.findByTokenHashFn != nil {
		return m.findByTokenHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockSessionRepo) RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error {
	if m.revokeByTokenHashFn != nil {
		return m.revokeByTokenHashFn(ctx, hash, revokedAt)
	}
	return nil
}

func (m *mockSessionRepo) DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error {
	if m.deleteByArcherIDFn != nil {
		return m.deleteByArcherIDFn(ctx, archerID)
	}
	return nil
}

func defaultTestConfig() auth.Config {
	return auth.Config{
		JWTSecret:           "test-secret-key-that-is-sufficiently-long",
		JWTAlgorithm:        "HS256",
		JWTTTLMinutes:       60,
		SessionTokenBytes:   32,
		GoogleOAuthClientID: "test-client-id",
	}
}

func TestService_LoginExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	archerID := uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	now := time.Now().UTC()

	archer := &model.ArcherRead{
		ArcherID:      archerID,
		FirstName:     "Robin",
		LastName:      "Hood",
		Email:         "robin@sherwood.org",
		DateOfBirth:   "1990-01-01",
		Gender:        model.GenderMale,
		Bowstyle:      model.BowstyleBarebow,
		DrawWeight:    40.0,
		GoogleSubject: "google-sub-1",
	}

	googleData := &auth.GoogleUserData{
		Sub:     "google-sub-1",
		Email:   "robin@sherwood.org",
		Picture: "https://sherwood.org/pic.png",
	}

	t.Run("successfully logs in existing archer, updates last login, creates session and jwt", func(t *testing.T) {
		t.Parallel()
		var (
			updateCalled bool
			sessionSaved bool
		)

		archers := &mockArcherRepo{
			updateFn: func(_ context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
				updateCalled = true
				if filter.ArcherID == nil || *filter.ArcherID != archerID {
					t.Errorf("expected filter by archer id %s", archerID)
				}
				if data.LastLoginAt == nil || *data.LastLoginAt != now {
					t.Errorf("expected LastLoginAt to match now")
				}
				if data.GooglePictureURL == nil || *data.GooglePictureURL != googleData.Picture {
					t.Errorf("expected picture url %s", googleData.Picture)
				}
				return nil
			},
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				updated := *archer
				updated.LastLoginAt = now
				updated.GooglePictureURL = &googleData.Picture
				return &updated, nil
			},
		}

		sessions := &mockSessionRepo{
			createFn: func(_ context.Context, data model.AuthSessionCreate) error {
				sessionSaved = true
				if data.ArcherID != archerID {
					t.Errorf("expected session archerID %s, got %s", archerID, data.ArcherID)
				}
				if len(data.SessionTokenHash) != 32 {
					t.Errorf("expected 32-byte token hash")
				}
				return nil
			},
		}

		svc := auth.NewService(archers, sessions, defaultTestConfig())
		ua := "Mozilla/5.0"
		ip := "127.0.0.1"
		res, err := svc.LoginExisting(ctx, archer, googleData, now, auth.SessionMetadata{
			UserAgent: &ua,
			IPAddress: &ip,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !updateCalled {
			t.Errorf("expected archer update to be called")
		}
		if !sessionSaved {
			t.Errorf("expected session create to be called")
		}
		if res.Status != model.AuthStatusAuthenticated {
			t.Errorf("expected status %s, got %s", model.AuthStatusAuthenticated, res.Status)
		}
		if res.AccessToken == "" {
			t.Errorf("expected access token to be generated")
		}

		// Verify generated JWT
		claims, err := auth.DecodeJWT(res.AccessToken, defaultTestConfig().JWTSecret, defaultTestConfig().JWTAlgorithm)
		if err != nil {
			t.Fatalf("failed to decode generated jwt: %v", err)
		}
		if claims.Sub != archerID.String() {
			t.Errorf("expected jwt sub %s, got %s", archerID.String(), claims.Sub)
		}
	})

	t.Run("nil archer returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.LoginExisting(ctx, nil, googleData, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil archer, got: %v", err)
		}
	})
}

func TestService_Register(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	newArcherID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")

	googleData := &auth.GoogleUserData{
		Sub:        "new-google-sub-99",
		Email:      "will@scarlet.org",
		GivenName:  "Will",
		FamilyName: "Scarlet",
		Picture:    "https://scarlet.org/pic.png",
	}

	payload := model.AuthRegistrationRequest{
		Credential:  "dummy-google-credential",
		DateOfBirth: "1995-05-15",
		Gender:      model.GenderMale,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  34.5,
	}

	t.Run("registers new archer and returns authenticated response", func(t *testing.T) {
		t.Parallel()
		var (
			createdData model.ArcherCreate
			sessionMade bool
		)

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil // Does not exist
			},
			createFn: func(_ context.Context, data model.ArcherCreate) (uuid.UUID, error) {
				createdData = data
				return newArcherID, nil
			},
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return &model.ArcherRead{
					ArcherID:      newArcherID,
					FirstName:     createdData.FirstName,
					LastName:      createdData.LastName,
					Email:         createdData.Email,
					DateOfBirth:   createdData.DateOfBirth,
					Gender:        createdData.Gender,
					Bowstyle:      createdData.Bowstyle,
					DrawWeight:    createdData.DrawWeight,
					GoogleSubject: createdData.GoogleSubject,
				}, nil
			},
		}

		sessions := &mockSessionRepo{
			createFn: func(_ context.Context, _ model.AuthSessionCreate) error {
				sessionMade = true
				return nil
			},
		}

		svc := auth.NewService(archers, sessions, defaultTestConfig())
		res, err := svc.Register(ctx, payload, googleData, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdData.FirstName != "Will" || createdData.LastName != "Scarlet" {
			t.Errorf("expected names from google data, got %s %s", createdData.FirstName, createdData.LastName)
		}
		if createdData.Email != "will@scarlet.org" {
			t.Errorf("expected email will@scarlet.org, got %s", createdData.Email)
		}
		if !sessionMade {
			t.Errorf("expected session to be created")
		}
		if res.Status != model.AuthStatusAuthenticated {
			t.Errorf("expected status %s, got %s", model.AuthStatusAuthenticated, res.Status)
		}
	})

	t.Run("if archer already exists with google subject, performs login instead", func(t *testing.T) {
		t.Parallel()
		existingArcher := &model.ArcherRead{
			ArcherID:      newArcherID,
			FirstName:     "Existing",
			LastName:      "Archer",
			Email:         googleData.Email,
			GoogleSubject: googleData.Sub,
		}

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return existingArcher, nil
			},
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.ArcherRead, error) {
				return existingArcher, nil
			},
		}

		sessions := &mockSessionRepo{}
		svc := auth.NewService(archers, sessions, defaultTestConfig())

		res, err := svc.Register(ctx, payload, googleData, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Archer.ArcherID != newArcherID {
			t.Errorf("expected existing archer ID %s, got %s", newArcherID, res.Archer.ArcherID)
		}
	})

	t.Run("fails when given name and family name are missing", func(t *testing.T) {
		t.Parallel()
		emptyGoogle := &auth.GoogleUserData{
			Sub:   "sub-without-names",
			Email: "noname@example.com",
		}
		emptyPayload := payload
		emptyPayload.FirstName = nil
		emptyPayload.LastName = nil

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(archers, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.Register(ctx, emptyPayload, emptyGoogle, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for missing names, got: %v", err)
		}
	})

	t.Run("fails for invalid date of birth format", func(t *testing.T) {
		t.Parallel()
		invalidPayload := payload
		invalidPayload.DateOfBirth = "15/05/1995" // Not YYYY-MM-DD

		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(_ context.Context, _ string) (*model.ArcherRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(archers, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.Register(ctx, invalidPayload, googleData, now)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid date format, got: %v", err)
		}
	})
}

func TestService_ValidateSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokenHash := []byte("32-byte-hash-of-session-token!")
	now := time.Now().UTC()

	t.Run("valid active session returns session data", func(t *testing.T) {
		t.Parallel()
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			ArcherID:         uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(1 * time.Hour),
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		got, err := svc.ValidateSession(ctx, tokenHash)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AuthID != session.AuthID {
			t.Errorf("expected authID %s, got %s", session.AuthID, got.AuthID)
		}
	})

	t.Run("empty token hash returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, nil)
		if err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil hash, got: %v", err)
		}
	})

	t.Run("non-existent session returns not found", func(t *testing.T) {
		t.Parallel()
		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return nil, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("revoked session returns unauthorized error", func(t *testing.T) {
		t.Parallel()
		revokedTime := now.Add(-10 * time.Minute)
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(1 * time.Hour),
			RevokedAt:        &revokedTime,
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got: %v", err)
		}
	})

	t.Run("expired session returns unauthorized error", func(t *testing.T) {
		t.Parallel()
		session := &model.AuthSessionRead{
			AuthID:           uuid.New(),
			SessionTokenHash: tokenHash,
			ExpiresAt:        now.Add(-10 * time.Minute), // Expired
		}

		sessions := &mockSessionRepo{
			findByTokenHashFn: func(_ context.Context, _ []byte) (*model.AuthSessionRead, error) {
				return session, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		_, err := svc.ValidateSession(ctx, tokenHash)
		if err == nil || !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized for expired session, got: %v", err)
		}
	})
}

func TestService_RevokeSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokenHash := []byte("test-hash")
	now := time.Now().UTC()

	t.Run("calls revoke by token hash", func(t *testing.T) {
		t.Parallel()
		var called bool
		sessions := &mockSessionRepo{
			revokeByTokenHashFn: func(_ context.Context, hash []byte, revokedAt time.Time) error {
				called = true
				if !bytes.Equal(hash, tokenHash) {
					t.Errorf("unexpected hash")
				}
				if revokedAt != now {
					t.Errorf("unexpected revokedAt")
				}
				return nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		if err := svc.RevokeSession(ctx, tokenHash, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected RevokeByTokenHash to be called")
		}
	})

	t.Run("empty token hash returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		if err := svc.RevokeSession(ctx, nil, now); err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty hash, got: %v", err)
		}
	})
}

func TestService_RevokeAllSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	archerID := uuid.New()

	t.Run("calls delete by archer id", func(t *testing.T) {
		t.Parallel()
		var called bool
		sessions := &mockSessionRepo{
			deleteByArcherIDFn: func(_ context.Context, id uuid.UUID) error {
				called = true
				if id != archerID {
					t.Errorf("expected archerID %s, got %s", archerID, id)
				}
				return nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessions, defaultTestConfig())
		if err := svc.RevokeAllSessions(ctx, archerID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected DeleteByArcherID to be called")
		}
	})

	t.Run("nil archer id returns validation error", func(t *testing.T) {
		t.Parallel()
		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, defaultTestConfig())
		if err := svc.RevokeAllSessions(ctx, uuid.Nil); err == nil || !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation for nil archer id, got: %v", err)
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	archerID := uuid.New()
	cfg := defaultTestConfig()

	rawSession := []byte("12345678901234567890123456789012")
	sid := auth.EncodeSessionID(rawSession)
	tokenHash := auth.HashSessionToken(rawSession)
	now := time.Now().UTC()

	t.Run("valid token returns archer id", func(t *testing.T) {
		t.Parallel()
		jwtToken, err := auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), cfg.JWTSecret, cfg.JWTAlgorithm)
		if err != nil {
			t.Fatalf("BuildJWT() error: %v", err)
		}

		sessionRepo := &mockSessionRepo{
			findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
				if !bytes.Equal(hash, tokenHash) {
					return nil, nil
				}
				return &model.AuthSessionRead{
					ArcherID:         archerID,
					SessionTokenHash: tokenHash,
					CreatedAt:        now,
					ExpiresAt:        now.Add(time.Hour),
				}, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessionRepo, cfg)
		gotID, err := svc.Authenticate(ctx, jwtToken)
		if err != nil {
			t.Fatalf("Authenticate() unexpected error: %v", err)
		}
		if gotID != archerID {
			t.Errorf("Authenticate() = %v, want %v", gotID, archerID)
		}
	})

	t.Run("expired jwt returns unauthorized", func(t *testing.T) {
		t.Parallel()
		past := now.Add(-2 * time.Hour)
		jwtToken, err := auth.BuildJWT(archerID, sid, past, past.Add(time.Hour), cfg.JWTSecret, cfg.JWTAlgorithm)
		if err != nil {
			t.Fatalf("BuildJWT() error: %v", err)
		}

		svc := auth.NewService(&mockArcherRepo{}, &mockSessionRepo{}, cfg)
		_, err = svc.Authenticate(ctx, jwtToken)
		if err == nil {
			t.Fatal("Authenticate() expected error for expired JWT, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
		}
	})

	t.Run("revoked session returns unauthorized", func(t *testing.T) {
		t.Parallel()
		jwtToken, err := auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), cfg.JWTSecret, cfg.JWTAlgorithm)
		if err != nil {
			t.Fatalf("BuildJWT() error: %v", err)
		}

		revokedAt := now.Add(-5 * time.Minute)
		sessionRepo := &mockSessionRepo{
			findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
				return &model.AuthSessionRead{
					ArcherID:         archerID,
					SessionTokenHash: tokenHash,
					CreatedAt:        now,
					ExpiresAt:        now.Add(time.Hour),
					RevokedAt:        &revokedAt,
				}, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessionRepo, cfg)
		_, err = svc.Authenticate(ctx, jwtToken)
		if err == nil {
			t.Fatal("Authenticate() expected error for revoked session, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
		}
	})

	t.Run("mismatched session archer id returns unauthorized", func(t *testing.T) {
		t.Parallel()
		jwtToken, err := auth.BuildJWT(archerID, sid, now, now.Add(time.Hour), cfg.JWTSecret, cfg.JWTAlgorithm)
		if err != nil {
			t.Fatalf("BuildJWT() error: %v", err)
		}

		otherArcherID := uuid.New()
		sessionRepo := &mockSessionRepo{
			findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
				return &model.AuthSessionRead{
					ArcherID:         otherArcherID,
					SessionTokenHash: tokenHash,
					CreatedAt:        now,
					ExpiresAt:        now.Add(time.Hour),
				}, nil
			},
		}

		svc := auth.NewService(&mockArcherRepo{}, sessionRepo, cfg)
		_, err = svc.Authenticate(ctx, jwtToken)
		if err == nil {
			t.Fatal("Authenticate() expected error for mismatched archer ID, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
		}
	})
}

func TestService_LoginWithGoogle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	t.Run("returns AuthNeedsRegistration when archer does not exist", func(t *testing.T) {
		t.Parallel()
		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(ctx context.Context, sub string) (*model.ArcherRead, error) {
				return nil, nil
			},
		}
		sessions := &mockSessionRepo{}
		cfg := defaultTestConfig()
		cfg.GoogleVerifier = func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-sub-new",
				Claims: map[string]any{
					"email":       "new@example.com",
					"given_name":  "New",
					"family_name": "Archer",
					"picture":     "https://pic.example.com/new.jpg",
				},
			}, nil
		}

		svc := auth.NewService(archers, sessions, cfg)
		authd, needsReg, err := svc.LoginWithGoogle(ctx, "valid-token", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authd != nil {
			t.Fatalf("expected nil AuthAuthenticated, got %+v", authd)
		}
		if needsReg == nil {
			t.Fatal("expected non-nil AuthNeedsRegistration")
		}
		if needsReg.GoogleEmail != "new@example.com" || needsReg.GoogleSubject != "google-sub-new" {
			t.Fatalf("mismatched Google details: %+v", needsReg)
		}
	})

	t.Run("returns AuthAuthenticated when archer exists", func(t *testing.T) {
		t.Parallel()
		existingID := uuid.New()
		existing := &model.ArcherRead{
			ArcherID:      existingID,
			FirstName:     "Existing",
			LastName:      "User",
			Email:         "existing@example.com",
			GoogleSubject: "google-sub-existing",
		}
		archers := &mockArcherRepo{
			findByGoogleSubjectFn: func(ctx context.Context, sub string) (*model.ArcherRead, error) {
				if sub == "google-sub-existing" {
					return existing, nil
				}
				return nil, nil
			},
			findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
				return existing, nil
			},
			updateFn: func(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
				return nil
			},
		}
		sessions := &mockSessionRepo{
			createFn: func(ctx context.Context, data model.AuthSessionCreate) error {
				return nil
			},
		}
		cfg := defaultTestConfig()
		cfg.GoogleVerifier = func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Subject: "google-sub-existing",
				Claims: map[string]any{
					"email": "existing@example.com",
				},
			}, nil
		}

		svc := auth.NewService(archers, sessions, cfg)
		authd, needsReg, err := svc.LoginWithGoogle(ctx, "valid-token", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if needsReg != nil {
			t.Fatalf("expected nil AuthNeedsRegistration, got %+v", needsReg)
		}
		if authd == nil {
			t.Fatal("expected non-nil AuthAuthenticated")
		}
		if authd.Archer.ArcherID != existingID {
			t.Fatalf("expected archer id %s, got %s", existingID, authd.Archer.ArcherID)
		}
	})

	t.Run("returns unauthorized error when google verification fails", func(t *testing.T) {
		t.Parallel()
		archers := &mockArcherRepo{}
		sessions := &mockSessionRepo{}
		cfg := defaultTestConfig()
		cfg.GoogleVerifier = func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
			return nil, errors.New("bad token")
		}

		svc := auth.NewService(archers, sessions, cfg)
		_, _, err := svc.LoginWithGoogle(ctx, "bad-token", now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestService_RegisterWithGoogle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	createdID := uuid.New()
	archers := &mockArcherRepo{
		findByGoogleSubjectFn: func(ctx context.Context, sub string) (*model.ArcherRead, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
			return createdID, nil
		},
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
			return &model.ArcherRead{
				ArcherID:  createdID,
				Email:     "reg@example.com",
				FirstName: "Reg",
				LastName:  "User",
			}, nil
		},
	}
	sessions := &mockSessionRepo{
		createFn: func(ctx context.Context, data model.AuthSessionCreate) error {
			return nil
		},
	}
	cfg := defaultTestConfig()
	cfg.GoogleVerifier = func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Subject: "google-sub-reg",
			Claims: map[string]any{
				"email":       "reg@example.com",
				"given_name":  "Reg",
				"family_name": "User",
			},
		}, nil
	}

	svc := auth.NewService(archers, sessions, cfg)

	payload := model.AuthRegistrationRequest{
		Credential:  "valid-reg-token",
		DateOfBirth: "1995-05-15",
		Gender:      model.GenderFemale,
		Bowstyle:    model.BowstyleRecurve,
		DrawWeight:  32.5,
	}

	authd, err := svc.RegisterWithGoogle(ctx, payload, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authd == nil || authd.Archer.Email != "reg@example.com" {
		t.Fatalf("unexpected registration result: %+v", authd)
	}
}

func TestService_RevokeTokenAndDecodeToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()

	archerID := uuid.New()
	arch := &model.ArcherRead{
		ArcherID: archerID,
		Email:    "test@example.com",
	}

	archers := &mockArcherRepo{
		updateFn: func(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error {
			return nil
		},
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
			return arch, nil
		},
	}

	var storedHash []byte
	var isRevoked bool
	sessions := &mockSessionRepo{
		createFn: func(ctx context.Context, data model.AuthSessionCreate) error {
			storedHash = data.SessionTokenHash
			return nil
		},
		findByTokenHashFn: func(ctx context.Context, hash []byte) (*model.AuthSessionRead, error) {
			if !bytes.Equal(hash, storedHash) {
				return nil, nil
			}
			var revokedAt *time.Time
			if isRevoked {
				rev := now
				revokedAt = &rev
			}
			return &model.AuthSessionRead{
				ArcherID:         archerID,
				SessionTokenHash: storedHash,
				CreatedAt:        now,
				ExpiresAt:        now.Add(time.Hour),
				RevokedAt:        revokedAt,
			}, nil
		},
		revokeByTokenHashFn: func(ctx context.Context, hash []byte, revokedAt time.Time) error {
			if bytes.Equal(hash, storedHash) {
				isRevoked = true
			}
			return nil
		},
	}

	cfg := defaultTestConfig()
	svc := auth.NewService(archers, sessions, cfg)

	authd, err := svc.LoginExisting(ctx, arch, nil, now)
	if err != nil {
		t.Fatalf("LoginExisting failed: %v", err)
	}

	claims, err := svc.DecodeToken(authd.AccessToken)
	if err != nil {
		t.Fatalf("DecodeToken failed: %v", err)
	}
	if claims.Sub != archerID.String() {
		t.Fatalf("expected subject %s, got %s", archerID, claims.Sub)
	}

	if err := svc.RevokeToken(ctx, authd.AccessToken); err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}

	// Verify token is now invalid due to revoked session
	_, err = svc.Authenticate(ctx, authd.AccessToken)
	if err == nil {
		t.Fatal("expected Authenticate to fail after revocation, but succeeded")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
