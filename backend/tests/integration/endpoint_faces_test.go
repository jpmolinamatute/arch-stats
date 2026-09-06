package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEndpointFaces_ListFaces(t *testing.T) {
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces", http.NoBody)
	var summaries []model.FaceMinimal
	resp, _ := doJSONRequest(t, nil, req, &summaries)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(summaries) == 0 {
		t.Fatal("expected at least 1 face definition")
	}
	for _, f := range summaries {
		if f.FaceType == "" || f.FaceName == "" {
			t.Errorf("incomplete face summary: %+v", f)
		}
	}
}

func TestEndpointFaces_GetFace_Success(t *testing.T) {
	ts, _ := newTestServer(t)

	// Fetch catalog first to get a real face_type
	listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces", http.NoBody)
	var summaries []model.FaceMinimal
	_, _ = doJSONRequest(t, nil, listReq, &summaries)

	if len(summaries) == 0 {
		t.Fatal("no faces in catalog")
	}
	var targetType model.FaceType
	for _, s := range summaries {
		if s.FaceType != "none" {
			targetType = s.FaceType
			break
		}
	}
	if targetType == "" {
		targetType = summaries[0].FaceType
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/faces/%s", ts.URL, targetType), http.NoBody)
	var face model.FaceRead
	resp, _ := doJSONRequest(t, nil, req, &face)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if face.FaceType != targetType {
		t.Errorf("FaceType = %v, want %v", face.FaceType, targetType)
	}
	if targetType != "none" && len(face.Rings) == 0 {
		t.Errorf("expected rings in face %v", targetType)
	}
}

func TestEndpointFaces_GetFace_NotFound(t *testing.T) {
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v0/faces/non_existent_face_type_xyz", http.NoBody)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 404 or 422", resp.StatusCode)
	}
}
