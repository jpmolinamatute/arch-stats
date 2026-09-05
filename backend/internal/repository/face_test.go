package repository_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestFaceRepo_FindAll(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != len(repository.DefaultFaceCatalog) {
		t.Fatalf("expected %d faces, got %d", len(repository.DefaultFaceCatalog), len(faces))
	}

	foundWA122 := false
	for _, f := range faces {
		if f.FaceType == model.FaceTypeWA122Full {
			foundWA122 = true
			if f.ViewBox != 1342.0 || len(f.Spots) != 1 || len(f.Rings) != 11 {
				t.Fatalf("unexpected WA122 definition: %+v", f)
			}
		}
	}
	if !foundWA122 {
		t.Fatal("WA 122cm full face not found in catalog")
	}
}

func TestFaceRepo_FindByType_Found(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindByType(context.Background(), model.FaceTypeWA40Full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("expected 1 face, got %d", len(faces))
	}
	if faces[0].FaceType != model.FaceTypeWA40Full {
		t.Fatalf("expected WA 40cm full, got %s", faces[0].FaceType)
	}
	if len(faces[0].Rings) != 11 {
		t.Fatalf("expected 11 rings, got %d", len(faces[0].Rings))
	}
}

func TestFaceRepo_FindByType_NotFound(t *testing.T) {
	repo := repository.NewFaceRepo()
	faces, err := repo.FindByType(context.Background(), model.FaceType("unknown_custom_face"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(faces) != 0 {
		t.Fatalf("expected 0 faces, got %d", len(faces))
	}
}

func TestFaceRepo_FindByID_Found(t *testing.T) {
	repo := repository.NewFaceRepo()
	face, err := repo.FindByID(context.Background(), string(model.FaceTypeWA80Full))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face == nil {
		t.Fatal("expected face, got nil")
	}
	if face.FaceType != model.FaceTypeWA80Full {
		t.Fatalf("expected WA 80cm full, got %s", face.FaceType)
	}
}

func TestFaceRepo_FindByID_NotFoundReturnsNil(t *testing.T) {
	repo := repository.NewFaceRepo()
	face, err := repo.FindByID(context.Background(), "non_existent_face")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if face != nil {
		t.Fatalf("expected nil face, got %+v", face)
	}
}

func TestFaceRepo_BuildSelectQuery(t *testing.T) {
	repo := repository.NewFaceRepo()

	// Test FindAll query building
	sqlAll, argsAll, err := repo.BuildSelectQuery(nil)
	if err != nil {
		t.Fatalf("unexpected error building all query: %v", err)
	}
	if !strings.Contains(sqlAll, "SELECT face_type, face_name, viewBox, render_cross FROM face") {
		t.Fatalf("unexpected SQL for FindAll: %s", sqlAll)
	}
	if len(argsAll) != 0 {
		t.Fatalf("expected 0 args for FindAll, got %d", len(argsAll))
	}

	// Test FindByType query building with WHERE clause
	ft := model.FaceTypeWA60Full
	sqlType, argsType, err := repo.BuildSelectQuery(&ft)
	if err != nil {
		t.Fatalf("unexpected error building type query: %v", err)
	}
	if !strings.Contains(sqlType, "WHERE face_type = $1") {
		t.Fatalf("unexpected SQL for FindByType: %s", sqlType)
	}
	if len(argsType) != 1 || argsType[0] != model.FaceTypeWA60Full {
		t.Fatalf("unexpected args for FindByType: %v", argsType)
	}
}

func TestFaceRepo_WithTx(t *testing.T) {
	mock := &mockDBTX{}
	repo := repository.NewFaceRepo(mock)
	txRepo := repo.WithTx(nil)
	if txRepo == nil {
		t.Fatal("expected non-nil repo from WithTx")
	}
}
