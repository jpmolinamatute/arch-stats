package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

func TestEndpointArcher_CreateSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	payload := model.ArcherCreate{
		FirstName:     "Test",
		LastName:      "Archer",
		Email:         "test.archer@example.com",
		DateOfBirth:   "1990-01-01",
		Gender:        model.GenderUnspecified,
		Bowstyle:      model.BowstyleRecurve,
		DrawWeight:    40.0,
		GoogleSubject: "test_subject_123",
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/archer/", token, bytes.NewReader(body))
	var respData map[string]any
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if _, ok := respData["archer_id"]; !ok {
		t.Fatalf("missing archer_id in response: %+v", respData)
	}
}

func TestEndpointArcher_ListArchers(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/archer/", token, nil)
	var archers []model.ArcherRead
	resp, _ := doJSONRequest(t, nil, req, &archers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(archers) == 0 {
		t.Fatal("expected at least 1 archer")
	}
	found := false
	for _, a := range archers {
		if a.ArcherID == archer.ArcherID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created archer %v not found in list", archer.ArcherID)
	}
}

func TestEndpointArcher_GetArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var res model.ArcherRead
	resp, _ := doJSONRequest(t, nil, req, &res)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if res.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", res.ArcherID, archer.ArcherID)
	}
	if res.Email != archer.Email {
		t.Errorf("Email = %v, want %v", res.Email, archer.Email)
	}
}

func TestEndpointArcher_GetArcherNotFound(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	randomID := uuid.New()
	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, randomID), token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEndpointArcher_UpdateArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	updatedName := "UpdatedRobin"
	updatePayload := model.ArcherUpdate{
		Where: model.ArcherFilter{ArcherID: &archer.ArcherID},
		Data:  model.ArcherSet{FirstName: &updatedName},
	}
	body, _ := json.Marshal(updatePayload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/archer/", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("patch status = %d, want 200", resp.StatusCode)
	}

	// Verify update with GET
	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var res model.ArcherRead
	getResp, _ := doJSONRequest(t, nil, getReq, &res)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want 200", getResp.StatusCode)
	}
	if res.FirstName != updatedName {
		t.Errorf("FirstName = %v, want %v", res.FirstName, updatedName)
	}
}

func TestEndpointArcher_DeleteArcherSuccess(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	// Create another archer to be deleted
	archerToDelete, err := createTestArcher(ctx, testPool)
	if err != nil {
		t.Fatalf("createTestArcher failed: %v", err)
	}

	_, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	delReq := authRequest(http.MethodDelete, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archerToDelete.ArcherID), token, nil)
	delResp, _ := doJSONRequest(t, nil, delReq, nil)
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", delResp.StatusCode)
	}

	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, archerToDelete.ArcherID), token, nil)
	getResp, _ := doJSONRequest(t, nil, getReq, nil)
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404", getResp.StatusCode)
	}
}

func TestEndpointArcher_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodGet, ts.URL + "/api/v0/archer/"},
		{http.MethodPost, ts.URL + "/api/v0/archer/"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, uuid.New())},
		{http.MethodPatch, ts.URL + "/api/v0/archer/"},
		{http.MethodDelete, fmt.Sprintf("%s/api/v0/archer/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
