package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// ShotRepository defines persistence operations required by ShotService.
type ShotRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error)
	FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error)
	FindAll(ctx context.Context, filter model.ShotFilter) ([]model.ShotRead, error)
	Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.ShotSet, filter model.ShotFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ShotService encapsulates business logic, score constraints, and validation for arrow shots.
type ShotService struct {
	shotRepo ShotRepository
	slotRepo SlotRepository
}

// NewShotService constructs a ShotService with repository dependency injection.
func NewShotService(shotRepo ShotRepository, slotRepo SlotRepository) *ShotService {
	return &ShotService{
		shotRepo: shotRepo,
		slotRepo: slotRepo,
	}
}

// GetByID retrieves a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) GetByID(ctx context.Context, id uuid.UUID) (*model.ShotRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	shot, err := s.shotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching shot: %w", err)
	}
	if shot == nil {
		return nil, apperror.ErrNotFound
	}
	return shot, nil
}

// ListBySlotID retrieves all shots recorded for a given slot.
func (s *ShotService) ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.ShotRead, error) {
	if slotID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	shots, err := s.shotRepo.FindBySlotID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("listing shots by slot: %w", err)
	}
	if shots == nil {
		return []model.ShotRead{}, nil
	}
	return shots, nil
}

// Create validates incoming shot data, ensures score rules and coordinate consistency are met,
// verifies that the target slot exists, and persists the shot record.
// Returns apperror.ErrNotFound if the slot does not exist.
func (s *ShotService) Create(ctx context.Context, data model.ShotCreate) (uuid.UUID, error) {
	if err := validateShotCreate(data); err != nil {
		return uuid.Nil, err
	}

	slot, err := s.slotRepo.FindByID(ctx, data.SlotID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("verifying slot: %w", err)
	}
	if slot == nil {
		return uuid.Nil, apperror.ErrNotFound
	}

	id, err := s.shotRepo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating shot: %w", err)
	}
	return id, nil
}

// Update validates mutation fields, checks score and is_x constraints, verifies shot existence,
// and updates the shot record.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) Update(ctx context.Context, id uuid.UUID, data model.ShotSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	if err := validateShotSet(data); err != nil {
		return err
	}

	existing, err := s.shotRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying shot: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	if data.IsX != nil && *data.IsX && data.Score == nil {
		if existing.Score == nil || *existing.Score != 10 {
			return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
		}
	}

	filter := model.ShotFilter{ShotID: &id}
	if err := s.shotRepo.Update(ctx, data, filter); err != nil {
		return fmt.Errorf("updating shot: %w", err)
	}
	return nil
}

// Delete removes a shot record by its primary key identifier.
// Returns apperror.ErrNotFound if the shot does not exist.
func (s *ShotService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "shot id is required")
	}

	if err := s.shotRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting shot: %w", err)
	}
	return nil
}

func validateShotCreate(data model.ShotCreate) error {
	if data.SlotID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "slot_id is required")
	}

	if data.Score != nil && (*data.Score < 0 || *data.Score > 10) {
		return apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
	}

	if data.IsX && (data.Score == nil || *data.Score != 10) {
		return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
	}

	allNil := data.X == nil && data.Y == nil && data.Score == nil
	allPresent := data.X != nil && data.Y != nil && data.Score != nil
	if !allNil && !allPresent {
		return apperror.Wrap(apperror.ErrValidation, "coordinates x, y, and score must either all be present or all be nil")
	}

	return nil
}

func validateShotSet(data model.ShotSet) error {
	if data.Score != nil && (*data.Score < 0 || *data.Score > 10) {
		return apperror.Wrap(apperror.ErrValidation, "score must be between 0 and 10")
	}

	if data.IsX != nil && *data.IsX && data.Score != nil && *data.Score != 10 {
		return apperror.Wrap(apperror.ErrValidation, "is_x requires score of 10")
	}

	return nil
}
