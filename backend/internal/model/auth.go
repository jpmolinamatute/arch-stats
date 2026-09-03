package model

import (
	"time"

	"github.com/google/uuid"
)

// AuthCreate represents the creation of a server-side authentication session.
type AuthCreate struct {
	ArcherID         uuid.UUID `json:"archer_id"`
	SessionTokenHash []byte    `json:"session_token_hash"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	UA               *string   `json:"ua,omitempty"`
	IPInet           *string   `json:"ip_inet,omitempty"`
}

// AuthFilter represents criteria to query authentication sessions.
type AuthFilter struct {
	SessionTokenHash []byte     `json:"session_token_hash,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

// AuthSet represents updates to an authentication session (e.g. revocation).
type AuthSet struct {
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// AuthUpdate wraps filter and update data for auth sessions.
type AuthUpdate struct {
	Where AuthFilter `json:"where"`
	Data  AuthSet    `json:"data"`
}

// AuthRead represents a persisted authentication session record.
type AuthRead struct {
	AuthID           uuid.UUID  `json:"auth_id"`
	ArcherID         uuid.UUID  `json:"archer_id"`
	SessionTokenHash []byte     `json:"session_token_hash"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	UA               *string    `json:"ua,omitempty"`
	IPInet           *string    `json:"ip_inet,omitempty"`
}

// GoogleOneTapRequest represents the incoming ID token from Google One Tap.
type GoogleOneTapRequest struct {
	Credential string `json:"credential" validate:"required,min=10"`
}

// AuthAuthenticated represents a successful authentication response.
type AuthAuthenticated struct {
	Status      AuthStatus `json:"status"`
	AccessToken string     `json:"access_token"`
	ExpiresAt   time.Time  `json:"expires_at"`
	Archer      ArcherRead `json:"archer"`
}

// AuthNeedsRegistration represents Google identity verified but requiring archer registration.
type AuthNeedsRegistration struct {
	Status             AuthStatus `json:"status"`
	GoogleEmail        string     `json:"google_email"`
	GoogleSubject      string     `json:"google_subject"`
	GivenName          *string    `json:"given_name,omitempty"`
	FamilyName         *string    `json:"family_name,omitempty"`
	GivenNameProvided  bool       `json:"given_name_provided"`
	FamilyNameProvided bool       `json:"family_name_provided"`
	PictureURL         *string    `json:"picture_url,omitempty"`
}

// LogoutResponse indicates logout operation success.
type LogoutResponse struct {
	Success bool `json:"success"`
}

// AuthRegistrationRequest represents profile information submitted on first registration.
type AuthRegistrationRequest struct {
	Credential  string     `json:"credential" validate:"required,min=10"`
	DateOfBirth string     `json:"date_of_birth" validate:"required"`
	Gender      Gender     `json:"gender" validate:"required"`
	Bowstyle    Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight  float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID      *uuid.UUID `json:"club_id,omitempty"`
	FirstName   *string    `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    *string    `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
}
