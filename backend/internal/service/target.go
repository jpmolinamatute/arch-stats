package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// TargetRepository defines persistence operations required by TargetService.
type TargetRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error)
	FindBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error)
	Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error)
	Update(ctx context.Context, data model.TargetSet, filter model.TargetFilter) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TargetService encapsulates business logic, validation, and operations for lane targets.
type TargetService struct {
	repo TargetRepository
}

// NewTargetService constructs a TargetService with repository dependency injection.
func NewTargetService(repo TargetRepository) *TargetService {
	return &TargetService{repo: repo}
}

// GetByID retrieves a lane target configuration by its primary key identifier.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) GetByID(ctx context.Context, id uuid.UUID) (*model.TargetRead, error) {
	if id == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	target, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching target: %w", err)
	}
	if target == nil {
		return nil, apperror.ErrNotFound
	}
	return target, nil
}

// ListBySlotID retrieves target configurations associated with a specific slot.
func (s *TargetService) ListBySlotID(ctx context.Context, slotID uuid.UUID) ([]model.TargetRead, error) {
	if slotID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "slot id is required")
	}

	targets, err := s.repo.FindBySlotID(ctx, slotID)
	if err != nil {
		return nil, fmt.Errorf("listing targets by slot: %w", err)
	}
	if targets == nil {
		return []model.TargetRead{}, nil
	}
	return targets, nil
}

// ListBySessionID retrieves target configurations associated with a specific session.
func (s *TargetService) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.TargetRead, error) {
	if sessionID == uuid.Nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "session id is required")
	}

	targets, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing targets by session: %w", err)
	}
	if targets == nil {
		return []model.TargetRead{}, nil
	}
	return targets, nil
}

// Create validates incoming target fields and persists the record.
func (s *TargetService) Create(ctx context.Context, data model.TargetCreate) (uuid.UUID, error) {
	if err := validateTargetCreate(data); err != nil {
		return uuid.Nil, err
	}

	id, err := s.repo.Create(ctx, data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("creating target: %w", err)
	}
	return id, nil
}

// Update validates mutation fields, verifies target existence, and updates the target record.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) Update(ctx context.Context, id uuid.UUID, data model.TargetSet) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	if err := validateTargetSet(data); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("verifying target existence: %w", err)
	}
	if existing == nil {
		return apperror.ErrNotFound
	}

	if data.Distance == nil && data.Lane == nil {
		return nil
	}

	filter := model.TargetFilter{TargetID: &id}
	if err := s.repo.Update(ctx, data, filter); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("updating target: %w", err)
	}
	return nil
}

// Delete removes a target configuration by its primary key identifier.
// Returns apperror.ErrNotFound if the target does not exist.
func (s *TargetService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "target id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("deleting target: %w", err)
	}
	return nil
}

func validateTargetCreate(data model.TargetCreate) error {
	if data.SessionID == uuid.Nil {
		return apperror.Wrap(apperror.ErrValidation, "session_id is required")
	}
	if data.Distance < 1 || data.Distance > 100 {
		return apperror.Wrap(apperror.ErrValidation, "distance must be between 1 and 100")
	}
	if data.Lane < 1 || data.Lane > 100 {
		return apperror.Wrap(apperror.ErrValidation, "lane must be between 1 and 100")
	}
	return nil
}

func validateTargetSet(data model.TargetSet) error {
	if data.Distance != nil && (*data.Distance < 1 || *data.Distance > 100) {
		return apperror.Wrap(apperror.ErrValidation, "distance must be between 1 and 100")
	}
	if data.Lane != nil && (*data.Lane < 1 || *data.Lane > 100) {
		return apperror.Wrap(apperror.ErrValidation, "lane must be between 1 and 100")
	}
	return nil
}
