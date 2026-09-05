package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// FaceRepository defines persistence operations required by FaceService.
type FaceRepository interface {
	FindByID(ctx context.Context, id string) (*model.FaceRead, error)
	FindAll(ctx context.Context) ([]model.FaceRead, error)
	FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)
}

// FaceService encapsulates business logic and catalog queries for target face definitions.
type FaceService struct {
	repo FaceRepository
}

// NewFaceService constructs a FaceService with repository dependency injection.
func NewFaceService(repo FaceRepository) *FaceService {
	return &FaceService{repo: repo}
}

// GetByID retrieves a target face definition by its string identifier (face_type).
// Returns apperror.ErrNotFound if no matching face exists.
func (s *FaceService) GetByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if strings.TrimSpace(id) == "" {
		return nil, apperror.Wrap(apperror.ErrValidation, "face id is required")
	}

	face, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching face: %w", err)
	}
	if face == nil {
		return nil, apperror.ErrNotFound
	}
	return face, nil
}

// ListAll returns all available target face definitions in the catalog.
func (s *FaceService) ListAll(ctx context.Context) ([]model.FaceRead, error) {
	faces, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing faces: %w", err)
	}
	if faces == nil {
		return []model.FaceRead{}, nil
	}
	return faces, nil
}

// ListByType retrieves target face definitions matching the provided face type.
// Returns apperror.ErrValidation if faceType is invalid.
func (s *FaceService) ListByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	if !isValidFaceType(faceType) {
		return nil, apperror.Wrap(apperror.ErrValidation, "invalid face type")
	}

	faces, err := s.repo.FindByType(ctx, faceType)
	if err != nil {
		return nil, fmt.Errorf("listing faces by type: %w", err)
	}
	if faces == nil {
		return []model.FaceRead{}, nil
	}
	return faces, nil
}
