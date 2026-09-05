package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
	"github.com/jpmolinamatute/arch-stats/backend/internal/service"
)

var (
	_ service.FaceRepository = (*mockFaceRepo)(nil)
	_ service.FaceRepository = (*repository.FaceRepo)(nil)
)

type mockFaceRepo struct {
	findByIDFn   func(ctx context.Context, id string) (*model.FaceRead, error)
	findAllFn    func(ctx context.Context) ([]model.FaceRead, error)
	findByTypeFn func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error)
}

func (m *mockFaceRepo) FindByID(ctx context.Context, id string) (*model.FaceRead, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockFaceRepo) FindAll(ctx context.Context) ([]model.FaceRead, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

func (m *mockFaceRepo) FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	if m.findByTypeFn != nil {
		return m.findByTypeFn(ctx, faceType)
	}
	return nil, nil
}

func sampleFaceRead(faceType model.FaceType, faceName string) model.FaceRead {
	return model.FaceRead{
		FaceType:    faceType,
		FaceName:    faceName,
		Spots:       []model.Spot{{XOffset: 0, YOffset: 0, Diameter: 40}},
		Rings:       []model.Ring{{DataScore: 10, Fill: "#FFD700", R: 2, Stroke: "#000000", StrokeWidth: 0.1}},
		ViewBox:     400,
		RenderCross: true,
	}
}

func TestFaceService_GetByID_Success(t *testing.T) {
	expected := sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full")
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			if id == "wa_40cm_full" {
				return &expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	face, err := svc.GetByID(context.Background(), "wa_40cm_full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face == nil || face.FaceType != model.FaceTypeWA40Full {
		t.Fatalf("unexpected face: %+v", face)
	}
}

func TestFaceService_GetByID_EmptyIDReturnsValidationError(t *testing.T) {
	mock := &mockFaceRepo{}
	svc := service.NewFaceService(mock)

	_, err := svc.GetByID(context.Background(), "  ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestFaceService_GetByID_NotFound(t *testing.T) {
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.GetByID(context.Background(), "unknown_face")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFaceService_GetByID_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.FaceRead, error) {
			return nil, errors.New("database failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.GetByID(context.Background(), "wa_40cm_full")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFaceService_ListAll_Success(t *testing.T) {
	expected := []model.FaceRead{
		sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full"),
		sampleFaceRead(model.FaceTypeWA60Full, "WA 60cm Full"),
	}
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return expected, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 2 {
		t.Fatalf("expected 2 faces, got %d", len(faces))
	}
}

func TestFaceService_ListAll_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if faces == nil || len(faces) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", faces)
	}
}

func TestFaceService_ListAll_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findAllFn: func(ctx context.Context) ([]model.FaceRead, error) {
			return nil, errors.New("query failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.ListAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFaceService_ListByType_Success(t *testing.T) {
	expected := []model.FaceRead{sampleFaceRead(model.FaceTypeWA40Full, "WA 40cm Full")}
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			if faceType == model.FaceTypeWA40Full {
				return expected, nil
			}
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListByType(context.Background(), model.FaceTypeWA40Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("expected 1 face, got %d", len(faces))
	}
}

func TestFaceService_ListByType_InvalidTypeReturnsValidationError(t *testing.T) {
	mock := &mockFaceRepo{}
	svc := service.NewFaceService(mock)

	_, err := svc.ListByType(context.Background(), model.FaceType("invalid_type"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestFaceService_ListByType_NilReturnsEmptySlice(t *testing.T) {
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			return nil, nil
		},
	}

	svc := service.NewFaceService(mock)
	faces, err := svc.ListByType(context.Background(), model.FaceTypeWA80Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if faces == nil || len(faces) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", faces)
	}
}

func TestFaceService_ListByType_RepoError(t *testing.T) {
	mock := &mockFaceRepo{
		findByTypeFn: func(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
			return nil, errors.New("lookup failure")
		},
	}

	svc := service.NewFaceService(mock)
	_, err := svc.ListByType(context.Background(), model.FaceTypeWA40Full)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
