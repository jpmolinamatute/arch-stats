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

func TestEndpointSlot_JoinSession_AssignsSlot(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	owner, ownerToken, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher owner failed: %v", err)
	}

	sess, err := createTestSession(ctx, testPool, owner.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	shotPerRound := 3
	joinPayload := model.SlotJoinRequest{
		SessionID:       sess.SessionID,
		ArcherID:        owner.ArcherID,
		Distance:        18,
		FaceType:        model.FaceTypeWA40Full,
		IsShooting:      true,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	body, _ := json.Marshal(joinPayload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session/slot", ownerToken, bytes.NewReader(body))
	var joinResp model.SlotJoinResponse
	resp, _ := doJSONRequest(t, nil, req, &joinResp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("join status = %d, want 200 or 201", resp.StatusCode)
	}
	if joinResp.SlotID == uuid.Nil {
		t.Fatal("expected valid slot_id")
	}
	if joinResp.Slot == "" {
		t.Fatal("expected non-empty slot string (e.g. 1A)")
	}
}

func TestEndpointSlot_GetSlot(t *testing.T) {
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

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/%s", ts.URL, slot.SlotID), token, nil)
	var fullInfo model.FullSlotInfo
	resp, _ := doJSONRequest(t, nil, req, &fullInfo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get slot status = %d, want 200", resp.StatusCode)
	}
	if fullInfo.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", fullInfo.SlotID, slot.SlotID)
	}
	if fullInfo.ArcherID != archer.ArcherID {
		t.Errorf("ArcherID = %v, want %v", fullInfo.ArcherID, archer.ArcherID)
	}
}

func TestEndpointSlot_GetArcherCurrentSlot(t *testing.T) {
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

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/archer/%s", ts.URL, archer.ArcherID), token, nil)
	var fullInfo model.FullSlotInfo
	resp, _ := doJSONRequest(t, nil, req, &fullInfo)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get archer slot status = %d, want 200", resp.StatusCode)
	}
	if fullInfo.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", fullInfo.SlotID, slot.SlotID)
	}
}

func TestEndpointSlot_LeaveSession(t *testing.T) {
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

	req := authRequest(http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/leave/%s", ts.URL, slot.SlotID), token, nil)
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave status = %d, want 200 or 204", resp.StatusCode)
	}

	// Verify slot is_shooting is false
	slotRepo := repository.NewSlotRepo(testPool)
	updatedSlot, err := slotRepo.FindByID(ctx, slot.SlotID)
	if err != nil || updatedSlot == nil {
		t.Fatalf("finding slot after leave: %v", err)
	}
	if updatedSlot.IsShooting {
		t.Fatal("expected is_shooting to be false after leave")
	}
}

func TestEndpointSlot_ReJoinSession(t *testing.T) {
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
	slot, err := createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID, func(s *model.SlotCreate) {
		s.IsShooting = false
	})
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/re-join/%s", ts.URL, slot.SlotID), token, nil)
	var rejoinResp model.SlotJoinResponse
	resp, _ := doJSONRequest(t, nil, req, &rejoinResp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("re-join status = %d, want 200", resp.StatusCode)
	}
	if rejoinResp.SlotID != slot.SlotID {
		t.Errorf("SlotID = %v, want %v", rejoinResp.SlotID, slot.SlotID)
	}
}

func TestEndpointSlot_JoinClosedSession_Unprocessable(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// Create and close session
	sess, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}
	sessionRepo := repository.NewSessionRepo(testPool)
	_ = sessionRepo.Close(ctx, sess.SessionID)

	shotPerRound := 3
	joinPayload := model.SlotJoinRequest{
		SessionID:       sess.SessionID,
		ArcherID:        archer.ArcherID,
		Distance:        18,
		FaceType:        model.FaceTypeWA40Full,
		IsShooting:      true,
		Bowstyle:        model.BowstyleRecurve,
		DrawWeight:      40.0,
		ShotPerRound:    &shotPerRound,
		IntervalSeconds: 20,
	}
	body, _ := json.Marshal(joinPayload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session/slot", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 Unprocessable Entity", resp.StatusCode)
	}
}

func TestEndpointSlot_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPost, ts.URL + "/api/v0/session/slot"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/slot/archer/%s", ts.URL, uuid.New())},
		{http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/leave/%s", ts.URL, uuid.New())},
		{http.MethodPatch, fmt.Sprintf("%s/api/v0/session/slot/re-join/%s", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
