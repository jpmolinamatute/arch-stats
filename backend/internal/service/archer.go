package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ArcherRepository defines persistence operations required by ArcherService.
type ArcherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error)
	FindAll(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error)
	Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ArcherSet, filter model.ArcherFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ArcherService encapsulates business logic and validation for archer profiles.
type ArcherService struct {
	repo ArcherRepository
}

// NewArcherService constructs an ArcherService with repository dependency injection.
func NewArcherService(repo ArcherRepository) *ArcherService {
	return &ArcherService{repo: repo}
}

// GetByID retrieves an archer profile by primary key identifier.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) GetByID(ctx context.Context, id uuid.UUID) (*model.ArcherRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	archer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching archer: %w", err)
	}
	if archer == nil {
		return nil, apperror.ErrNotFound
	}
	return archer, nil
}

// List queries archer profiles matching the provided filter criteria.
//
//nolint:gocritic // hugeParam: filter value parameter matches service interface specification
func (s *ArcherService) List(ctx context.Context, filter model.ArcherFilter) ([]model.ArcherRead, error) {
	archers, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing archers: %w", err)
	}
	if archers == nil {
		return []model.ArcherRead{}, nil
	}
	return archers, nil
}

// Create validates incoming archer profile fields and persists the record.
//
//nolint:gocritic // hugeParam: data value parameter matches domain model parameter specification
func (s *ArcherService) Create(ctx context.Context, data model.ArcherCreate) (uuid.UUID, error) {
	if err := validateArcherCreate(data); err != nil {
		return uuid.Nil, err
	}

	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating archer: %w", err)
	}
	return id, nil
}

// Update validates mutation fields and updates an archer profile.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) Update(ctx context.Context, id uuid.UUID, data model.ArcherSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	if err := validateArcherSet(data); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying archer existence: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	filter := model.ArcherFilter{ArcherID: &id}
	if err := s.repo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating archer: %w", err)
	}
	return nil
}

// Delete removes an archer profile by primary key identifier.
// Returns apperror.ErrNotFound if the archer does not exist.
func (s *ArcherService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "archer id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting archer: %w", err)
	}
	return nil
}

//nolint:gocritic // hugeParam: data value parameter matches domain model parameter specification
func validateArcherCreate(data model.ArcherCreate) error {
	if strings.TrimSpace(data.FirstName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "first_name is required")
	}
	if strings.TrimSpace(data.LastName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "last_name is required")
	}
	if strings.TrimSpace(data.Email) == "" {
		return apperror.Wrap(apperror.ErrValidation, "email is required")
	}
	if _, err := mail.ParseAddress(data.Email); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "invalid email address")
	}
	if strings.TrimSpace(data.DateOfBirth) == "" {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth is required")
	}
	if _, err := time.Parse("2006-01-02", data.DateOfBirth); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if !isValidGender(data.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if !isValidBowstyle(data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight <= 0 || data.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	if strings.TrimSpace(data.GoogleSubject) == "" {
		return apperror.Wrap(apperror.ErrValidation, "google_subject is required")
	}
	return nil
}

func validateArcherSet(data model.ArcherSet) error {
	if data.FirstName != nil && strings.TrimSpace(*data.FirstName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "first_name cannot be empty")
	}
	if data.LastName != nil && strings.TrimSpace(*data.LastName) == "" {
		return apperror.Wrap(apperror.ErrValidation, "last_name cannot be empty")
	}
	if data.Gender != nil && !isValidGender(*data.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if data.Bowstyle != nil && !isValidBowstyle(*data.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if data.DrawWeight != nil && (*data.DrawWeight <= 0 || *data.DrawWeight > 200) {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	return nil
}

func isValidGender(g model.Gender) bool {
	switch g {
	case model.GenderMale, model.GenderFemale, model.GenderNonBinary, model.GenderOther, model.GenderUnspecified:
		return true
	default:
		return false
	}
}

func isValidBowstyle(b model.Bowstyle) bool {
	switch b {
	case model.BowstyleRecurve, model.BowstyleCompound, model.BowstyleBarebow, model.BowstyleLongbow:
		return true
	default:
		return false
	}
}
