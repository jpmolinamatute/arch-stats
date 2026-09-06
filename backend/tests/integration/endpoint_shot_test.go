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
	"github.com/jpmolinamatute/arch-stats/backend/internal/repository"
)

func TestEndpointShot_CreateSingleShot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	x := 10.5
	y := 20.5
	score := 9
	payload := model.ShotCreate{
		SlotID: slot.SlotID,
		X:      &x,
		Y:      &y,
		Score:  &score,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	var shotResp model.ShotID
	resp, _ := doJSONRequest(t, nil, req, &shotResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if shotResp.ShotID == uuid.Nil {
		t.Fatal("expected valid shot_id")
	}
}

func TestEndpointShot_CreateBatchShots(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	x := 5.0
	y := 5.0
	score := 10
	shots := []model.ShotCreate{
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
		{SlotID: slot.SlotID, X: &x, Y: &y, Score: &score},
	}
	body, _ := json.Marshal(shots)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	var batchResp []model.ShotID
	resp, _ := doJSONRequest(t, nil, req, &batchResp)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if len(batchResp) != 3 {
		t.Fatalf("got %d shot_ids, want 3", len(batchResp))
	}
}

func TestEndpointShot_GetBySlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	shot1, _ := createTestShot(ctx, testPool, slot.SlotID)
	shot2, _ := createTestShot(ctx, testPool, slot.SlotID)

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/shot/by-slot/%s", ts.URL, slot.SlotID), token, nil)
	var shots []model.ShotRead
	resp, _ := doJSONRequest(t, nil, req, &shots)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(shots) != 2 {
		t.Fatalf("got %d shots, want 2", len(shots))
	}
	gotIDs := map[uuid.UUID]bool{shots[0].ShotID: true, shots[1].ShotID: true}
	if !gotIDs[shot1.ShotID] || !gotIDs[shot2.ShotID] {
		t.Errorf("missing created shots in response: %+v", shots)
	}
}

func TestEndpointShot_CountBySlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	_, _ = createTestShot(ctx, testPool, slot.SlotID)
	_, _ = createTestShot(ctx, testPool, slot.SlotID)
	_, _ = createTestShot(ctx, testPool, slot.SlotID)

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/shot/count-by-slot/%s", ts.URL, slot.SlotID), token, nil)
	var count int
	resp, _ := doJSONRequest(t, nil, req, &count)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestEndpointShot_InvalidScore(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	invalidScores := []int{-1, 11, 99}
	for _, sc := range invalidScores {
		score := sc
		payload := model.ShotCreate{
			SlotID: slot.SlotID,
			Score:  &score,
		}
		body, _ := json.Marshal(payload)

		req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
		resp, _ := doJSONRequest(t, nil, req, nil)

		if resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("score %d status = %d, want 422 or 400", sc, resp.StatusCode)
		}
	}
}

func TestEndpointShot_CreateShotInClosedSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	target, err := createTestTarget(ctx, testPool, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestTarget failed: %v", err)
	}
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	// Close the session
	sessionRepo := repository.NewSessionRepo(testPool)
	_ = sessionRepo.Close(ctx, sess.SessionID)

	score := 10
	payload := model.ShotCreate{
		SlotID: slot.SlotID,
		Score:  &score,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/shot", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 Unprocessable Entity", resp.StatusCode)
	}
}

func TestEndpointShot_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPost, ts.URL + "/api/v0/shot"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/shot/by-slot/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/shot/count-by-slot/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
