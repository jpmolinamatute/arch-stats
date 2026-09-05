package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherRepository defines persistence operations for archer profiles required by the auth system.
type ArcherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	FindByGoogleSubject(ctx context.Context, sub string) (*model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
}

// SessionRepository defines persistence operations for auth sessions required by the auth system.
type SessionRepository interface {
	Create(ctx context.Context, data model.AuthSessionCreate) error
	FindByTokenHash(ctx context.Context, hash []byte) (*model.AuthSessionRead, error)
	RevokeByTokenHash(ctx context.Context, hash []byte, revokedAt time.Time) error
	DeleteByArcherID(ctx context.Context, archerID uuid.UUID) error
}

// Config provides configuration settings for JWT signing, session lifetime, and Google OAuth.
type Config struct {
	JWTSecret           string
	JWTAlgorithm        string
	JWTTTLMinutes       int
	SessionTokenBytes   int
	GoogleOAuthClientID string
}

// SessionMetadata captures optional client request metadata (e.g. User-Agent and client IP).
type SessionMetadata struct {
	UserAgent *string
	IPAddress *string
}

// Service orchestrates authentication workflows including Google ID token validation,
// archer registration, session token minting, and JWT issuance.
type Service struct {
	archers  ArcherRepository
	sessions SessionRepository
	cfg      Config
}

// NewService constructs an auth Service with repository and config dependencies.
func NewService(archers ArcherRepository, sessions SessionRepository, cfg Config) *Service {
	if cfg.SessionTokenBytes <= 0 {
		cfg.SessionTokenBytes = 32
	}
	if cfg.JWTTTLMinutes <= 0 {
		cfg.JWTTTLMinutes = 1440
	}
	if strings.TrimSpace(cfg.JWTAlgorithm) == "" {
		cfg.JWTAlgorithm = "HS256"
	}
	return &Service{
		archers:  archers,
		sessions: sessions,
		cfg:      cfg,
	}
}

// LoginExisting updates the archer's last login timestamp and Google picture, creates a new session,
// signs an access JWT, and returns an AuthAuthenticated domain payload.
func (s *Service) LoginExisting(
	ctx context.Context,
	archer *model.ArcherRead,
	googleData *GoogleUserData,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	if archer == nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer is required")
	}

	updateData := model.ArcherSet{
		LastLoginAt: &now,
	}
	if googleData != nil && strings.TrimSpace(googleData.Picture) != "" {
		trimmedPic := strings.TrimSpace(googleData.Picture)
		updateData.GooglePictureURL = &trimmedPic
	}

	filter := model.ArcherFilter{ArcherID: &archer.ArcherID}
	if err := s.archers.Update(ctx, updateData, filter); err != nil {
		return nil, fmt.Errorf("updating archer last login: %w", err)
	}

	updatedArcher, err := s.archers.FindByID(ctx, archer.ArcherID)
	if err != nil {
		return nil, fmt.Errorf("fetching updated archer: %w", err)
	}
	if updatedArcher == nil {
		return nil, apperror.ErrNotFound
	}

	return s.createSessionAndToken(ctx, updatedArcher, now, meta...)
}

// Register registers a new archer, validates demographics and names, and authenticates the newly created profile.
// If an archer with the same Google subject already exists, it transparently logs them in.
//
//nolint:gocritic // hugeParam: payload matches API request specification
func (s *Service) Register(
	ctx context.Context,
	payload model.AuthRegistrationRequest,
	googleData *GoogleUserData,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	if googleData == nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "googleData is required")
	}
	if strings.TrimSpace(googleData.Sub) == "" || strings.TrimSpace(googleData.Email) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "googleData must contain email and sub")
	}

	existing, err := s.archers.FindByGoogleSubject(ctx, googleData.Sub)
	if err != nil {
		return nil, fmt.Errorf("verifying existing archer by google subject: %w", err)
	}
	if existing != nil {
		return s.LoginExisting(ctx, existing, googleData, now, meta...)
	}

	given := ""
	if payload.FirstName != nil && strings.TrimSpace(*payload.FirstName) != "" {
		given = strings.TrimSpace(*payload.FirstName)
	} else if strings.TrimSpace(googleData.GivenName) != "" {
		given = strings.TrimSpace(googleData.GivenName)
	}

	family := ""
	if payload.LastName != nil && strings.TrimSpace(*payload.LastName) != "" {
		family = strings.TrimSpace(*payload.LastName)
	} else if strings.TrimSpace(googleData.FamilyName) != "" {
		family = strings.TrimSpace(googleData.FamilyName)
	}

	var missingNames []string
	if given == "" {
		missingNames = append(missingNames, "given_name is missing")
	}
	if family == "" {
		missingNames = append(missingNames, "family_name is missing")
	}
	if len(missingNames) > 0 {
		return nil, apperror.Wrap(apperror.ErrValidation, strings.Join(missingNames, ", "))
	}

	if _, err := time.Parse("2006-01-02", payload.DateOfBirth); err != nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if payload.DrawWeight <= 0 || payload.DrawWeight > 200 {
		return nil, apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}

	var pictureURL *string
	if trimmedPic := strings.TrimSpace(googleData.Picture); trimmedPic != "" {
		pictureURL = &trimmedPic
	}

	createData := model.ArcherCreate{
		FirstName:        given,
		LastName:         family,
		Email:            googleData.Email,
		DateOfBirth:      payload.DateOfBirth,
		Gender:           payload.Gender,
		Bowstyle:         payload.Bowstyle,
		DrawWeight:       payload.DrawWeight,
		ClubID:           payload.ClubID,
		GooglePictureURL: pictureURL,
		GoogleSubject:    googleData.Sub,
	}

	newArcherID, err := s.archers.Create(ctx, createData)
	if err != nil {
		return nil, fmt.Errorf("creating archer: %w", err)
	}

	createdArcher, err := s.archers.FindByID(ctx, newArcherID)
	if err != nil {
		return nil, fmt.Errorf("fetching created archer: %w", err)
	}
	if createdArcher == nil {
		return nil, apperror.ErrNotFound
	}

	return s.createSessionAndToken(ctx, createdArcher, now, meta...)
}

// ValidateSession retrieves an active session by its SHA-256 token hash and validates that it is neither
// revoked nor expired.
func (s *Service) ValidateSession(ctx context.Context, tokenHash []byte) (*model.AuthSessionRead, error) {
	if len(tokenHash) == 0 {
		return nil, apperror.Wrap(apperror.ErrValidation, "token hash is required")
	}

	session, err := s.sessions.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("finding session by token hash: %w", err)
	}
	if session == nil {
		return nil, apperror.ErrNotFound
	}

	if session.RevokedAt != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "session has been revoked")
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "session has expired")
	}

	return session, nil
}

// RevokeSession revokes a single session identified by its token hash.
func (s *Service) RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	if len(tokenHash) == 0 {
		return apperror.Wrap(apperror.ErrValidation, "token hash is required")
	}
	if err := s.sessions.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

// RevokeAllSessions invalidates all sessions belonging to the specified archer.
func (s *Service) RevokeAllSessions(ctx context.Context, archerID uuid.UUID) error {
	if archerID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archerID cannot be nil")
	}
	if err := s.sessions.DeleteByArcherID(ctx, archerID); err != nil {
		return fmt.Errorf("deleting all sessions for archer: %w", err)
	}
	return nil
}

// VerifyGoogleToken verifies a Google One Tap credential using the configured client ID.
func (s *Service) VerifyGoogleToken(ctx context.Context, credential string) (*GoogleUserData, error) {
	return VerifyGoogleIDToken(ctx, credential, s.cfg.GoogleOAuthClientID)
}

// Authenticate verifies and decodes an access JWT, validates the underlying session in the database,
// ensures the session matches the archer, and returns the authenticated archer UUID.
func (s *Service) Authenticate(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	claims, err := DecodeJWT(tokenStr, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
	if err != nil {
		return uuid.Nil, err
	}

	rawSession, err := DecodeSessionID(claims.SID)
	if err != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid session id in token")
	}

	tokenHash := HashSessionToken(rawSession)
	session, err := s.ValidateSession(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, err
	}

	archerID, err := claims.ArcherID()
	if err != nil {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "invalid archer id in token")
	}

	if session.ArcherID != archerID {
		return uuid.Nil, apperror.Wrap(apperror.ErrUnauthorized, "session does not belong to archer")
	}

	return archerID, nil
}

func (s *Service) createSessionAndToken(
	ctx context.Context,
	archer *model.ArcherRead,
	now time.Time,
	meta ...SessionMetadata,
) (*model.AuthAuthenticated, error) {
	rawSession, err := GenerateSessionToken(s.cfg.SessionTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	tokenHash := HashSessionToken(rawSession)
	expiresAt := now.Add(time.Duration(s.cfg.JWTTTLMinutes) * time.Minute)

	var (
		ua *string
		ip *string
	)
	if len(meta) > 0 {
		ua = meta[0].UserAgent
		ip = meta[0].IPAddress
	}

	sessionCreate := model.AuthSessionCreate{
		ArcherID:         archer.ArcherID,
		SessionTokenHash: tokenHash,
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UA:               ua,
		IPInet:           ip,
	}

	if err := s.sessions.Create(ctx, sessionCreate); err != nil {
		return nil, fmt.Errorf("creating auth session: %w", err)
	}

	sid := EncodeSessionID(rawSession)
	jwtToken, err := BuildJWT(archer.ArcherID, sid, now, expiresAt, s.cfg.JWTSecret, s.cfg.JWTAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("building jwt: %w", err)
	}

	return &model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: jwtToken,
		ExpiresAt:   expiresAt,
		Archer:      *archer,
	}, nil
}
