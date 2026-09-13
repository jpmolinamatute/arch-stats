package model

import (
	"github.com/google/uuid"
)

// ArcherCreate represents the payload required to create a new archer profile.
type ArcherCreate struct {
	ArcherID    *uuid.UUID `json:"archer_id,omitempty"`
	FirstName   string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName    string     `json:"last_name" validate:"required,min=1,max=100"`
	Email       string     `json:"email" validate:"required,email"`
	DateOfBirth string     `json:"date_of_birth" validate:"required"`
	Gender      Gender     `json:"gender" validate:"required"`
}

// ArcherSet represents mutable fields when updating an archer profile.
type ArcherSet struct {
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Email       *string `json:"email,omitempty" validate:"omitempty,email"`
	DateOfBirth *string `json:"date_of_birth,omitempty"`
	Gender      *Gender `json:"gender,omitempty"`
}

// ArcherFilter represents criteria to query or select archers.
type ArcherFilter struct {
	ArcherID  *uuid.UUID `json:"archer_id,omitempty"`
	Email     *string    `json:"email,omitempty"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Gender    *Gender    `json:"gender,omitempty"`
	IsDeleted *bool      `json:"is_deleted,omitempty"`
}

// ArcherUpdate wraps target filter criteria and field updates for archers.
type ArcherUpdate struct {
	Where ArcherFilter `json:"where"`
	Data  ArcherSet    `json:"data"`
}

// ArcherRead represents the full persisted archer personal profile domain model.
type ArcherRead struct {
	ArcherID    uuid.UUID `json:"archer_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      Gender    `json:"gender"`
	IsDeleted   bool      `json:"is_deleted"`
}

// ArcherID represents a standalone archer identifier wrapper.
type ArcherID struct {
	ArcherID uuid.UUID `json:"archer_id"`
}
