package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestContext_ArcherIDRoundTrip(t *testing.T) {
	expectedID := uuid.New()
	ctx := context.Background()

	ctxWithArcher := middleware.WithArcherID(ctx, expectedID)

	gotID, err := middleware.GetArcherID(ctxWithArcher)
	if err != nil {
		t.Fatalf("GetArcherID() returned unexpected error: %v", err)
	}
	if gotID != expectedID {
		t.Errorf("GetArcherID() = %v, want %v", gotID, expectedID)
	}
}

func TestContext_GetArcherID_EmptyContextReturnsError(t *testing.T) {
	ctx := context.Background()

	gotID, err := middleware.GetArcherID(ctx)
	if err == nil {
		t.Fatalf("GetArcherID(emptyCtx) expected error, got nil (id=%v)", gotID)
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("GetArcherID(emptyCtx) error = %v, want errors.Is(err, apperror.ErrUnauthorized)", err)
	}
	if gotID != uuid.Nil {
		t.Errorf("GetArcherID(emptyCtx) id = %v, want uuid.Nil", gotID)
	}
}

func TestContext_GetArcherID_NilUUIDReturnsError(t *testing.T) {
	ctx := middleware.WithArcherID(context.Background(), uuid.Nil)

	gotID, err := middleware.GetArcherID(ctx)
	if err == nil {
		t.Fatalf("GetArcherID(nilUUID) expected error, got nil")
	}
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Errorf("GetArcherID(nilUUID) error = %v, want errors.Is(err, apperror.ErrUnauthorized)", err)
	}
	if gotID != uuid.Nil {
		t.Errorf("GetArcherID(nilUUID) id = %v, want uuid.Nil", gotID)
	}
}
