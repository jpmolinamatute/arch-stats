package model

import (
	"time"

	"github.com/google/uuid"
)

// ArcherCreate represents the payload required to create a new archer profile.
type ArcherCreate struct {
	FirstName        string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName         string     `json:"last_name" validate:"required,min=1,max=100"`
	Email            string     `json:"email" validate:"required,email"`
	DateOfBirth      string     `json:"date_of_birth" validate:"required"`
	Gender           Gender     `json:"gender" validate:"required"`
	Bowstyle         Bowstyle   `json:"bowstyle" validate:"required"`
	DrawWeight       float64    `json:"draw_weight" validate:"required,gt=0,lte=200"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	GoogleSubject    string     `json:"google_subject" validate:"required"`
}

// ArcherSet represents mutable fields when updating an archer profile.
type ArcherSet struct {
	FirstName        *string    `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName         *string    `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Gender           *Gender    `json:"gender,omitempty"`
	Bowstyle         *Bowstyle  `json:"bowstyle,omitempty"`
	DrawWeight       *float64   `json:"draw_weight,omitempty" validate:"omitempty,gt=0,lte=200"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
}

// ArcherFilter represents criteria to query or select archers.
type ArcherFilter struct {
	ArcherID      *uuid.UUID `json:"archer_id,omitempty"`
	FirstName     *string    `json:"first_name,omitempty"`
	LastName      *string    `json:"last_name,omitempty"`
	Gender        *Gender    `json:"gender,omitempty"`
	Bowstyle      *Bowstyle  `json:"bowstyle,omitempty"`
	DrawWeight    *float64   `json:"draw_weight,omitempty"`
	ClubID        *uuid.UUID `json:"club_id,omitempty"`
	GoogleSubject *string    `json:"google_subject,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// ArcherUpdate wraps target filter criteria and field updates for archers.
type ArcherUpdate struct {
	Where ArcherFilter `json:"where"`
	Data  ArcherSet    `json:"data"`
}

// ArcherRead represents the full persisted archer domain model.
type ArcherRead struct {
	ArcherID         uuid.UUID  `json:"archer_id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Email            string     `json:"email"`
	DateOfBirth      string     `json:"date_of_birth"`
	Gender           Gender     `json:"gender"`
	Bowstyle         Bowstyle   `json:"bowstyle"`
	DrawWeight       float64    `json:"draw_weight"`
	ClubID           *uuid.UUID `json:"club_id,omitempty"`
	GooglePictureURL *string    `json:"google_picture_url,omitempty"`
	GoogleSubject    string     `json:"google_subject"`
	LastLoginAt      time.Time  `json:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ArcherID represents a standalone archer identifier wrapper.
type ArcherID struct {
	ArcherID uuid.UUID `json:"archer_id"`
}
