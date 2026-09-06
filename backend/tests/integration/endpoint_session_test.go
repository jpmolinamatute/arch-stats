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

func TestEndpointSession_CreateSession(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	payload := model.SessionCreate{
		OwnerArcherID:   archer.ArcherID,
		SessionLocation: "Main Range",
		IsIndoor:        false,
		IsOpened:        true,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session", token, bytes.NewReader(body))
	var respData model.SessionID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if respData.SessionID == nil || *respData.SessionID == uuid.Nil {
		t.Fatalf("missing session_id in response: %+v", respData)
	}
}

func TestEndpointSession_CreateSession_AlreadyOpen_Conflict(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// 1. Create first open session
	_, err = createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession failed: %v", err)
	}

	// 2. Attempt to create second open session
	payload := model.SessionCreate{
		OwnerArcherID:   archer.ArcherID,
		SessionLocation: "Second Range",
		IsIndoor:        false,
		IsOpened:        true,
	}
	body, _ := json.Marshal(payload)

	req := authRequest(http.MethodPost, ts.URL+"/api/v0/session", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409 Conflict", resp.StatusCode)
	}
}

func TestEndpointSession_GetByID(t *testing.T) {
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

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, sess.SessionID), token, nil)
	var readSess model.SessionRead
	resp, _ := doJSONRequest(t, nil, req, &readSess)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if readSess.SessionID != sess.SessionID {
		t.Errorf("SessionID = %v, want %v", readSess.SessionID, sess.SessionID)
	}
}

func TestEndpointSession_GetOpenForArcher(t *testing.T) {
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

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/open-session", ts.URL, archer.ArcherID), token, nil)
	var respData model.SessionID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if respData.SessionID == nil || *respData.SessionID != sess.SessionID {
		t.Errorf("session_id = %v, want %v", respData.SessionID, sess.SessionID)
	}
}

func TestEndpointSession_ListAllOpen(t *testing.T) {
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

	req := authRequest(http.MethodGet, ts.URL+"/api/v0/session/open", token, nil)
	var sessions []model.SessionRead
	resp, _ := doJSONRequest(t, nil, req, &sessions)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	found := false
	for _, s := range sessions {
		if s.SessionID == sess.SessionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("open session %v not found in list", sess.SessionID)
	}
}

func TestEndpointSession_CloseSession(t *testing.T) {
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

	closePayload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(closePayload)

	req := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(body))
	resp, _ := doJSONRequest(t, nil, req, nil)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 200 or 204", resp.StatusCode)
	}

	// Verify session is no longer open
	getReq := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, sess.SessionID), token, nil)
	var readSess model.SessionRead
	_, _ = doJSONRequest(t, nil, getReq, &readSess)
	if readSess.IsOpened {
		t.Fatal("expected session to be closed, but is_opened is true")
	}
}

func TestEndpointSession_ReOpenSession(t *testing.T) {
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

	// Close first
	closePayload := model.SessionID{SessionID: &sess.SessionID}
	body, _ := json.Marshal(closePayload)
	closeReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(body))
	_, _ = doJSONRequest(t, nil, closeReq, nil)

	// Now re-open
	reopenReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/re-open", token, bytes.NewReader(body))
	reopenResp, _ := doJSONRequest(t, nil, reopenReq, nil)
	if reopenResp.StatusCode != http.StatusOK {
		t.Fatalf("re-open status = %d, want 200", reopenResp.StatusCode)
	}
}

func TestEndpointSession_ReOpen_BlockedIfAlreadyOpen(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	archer, token, err := createAuthenticatedArcher(ctx, testPool, testJWTSecret)
	if err != nil {
		t.Fatalf("createAuthenticatedArcher failed: %v", err)
	}

	// 1. Create Session A and close it
	sessA, err := createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession A failed: %v", err)
	}
	bodyA, _ := json.Marshal(model.SessionID{SessionID: &sessA.SessionID})
	closeReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/close", token, bytes.NewReader(bodyA))
	_, _ = doJSONRequest(t, nil, closeReq, nil)

	// 2. Create Session B (now open)
	_, err = createTestSession(ctx, testPool, archer.ArcherID)
	if err != nil {
		t.Fatalf("createTestSession B failed: %v", err)
	}

	// 3. Attempt to re-open Session A -> conflict
	reopenReq := authRequest(http.MethodPatch, ts.URL+"/api/v0/session/re-open", token, bytes.NewReader(bodyA))
	resp, _ := doJSONRequest(t, nil, reopenReq, nil)
	if resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 409 or 422", resp.StatusCode)
	}
}

func TestEndpointSession_GetParticipating(t *testing.T) {
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
	_, err = createTestSlot(ctx, testPool, target.TargetID, archer.ArcherID, sess.SessionID)
	if err != nil {
		t.Fatalf("createTestSlot failed: %v", err)
	}

	req := authRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/participating", ts.URL, archer.ArcherID), token, nil)
	var respData map[string]*uuid.UUID
	resp, _ := doJSONRequest(t, nil, req, &respData)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if respData["session_id"] == nil || *respData["session_id"] != sess.SessionID {
		t.Errorf("session_id = %v, want %v", respData["session_id"], sess.SessionID)
	}
}

func TestEndpointSession_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { _ = truncateAll(ctx, testPool) })
	ts, _ := newTestServer(t)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodGet, ts.URL + "/api/v0/session/open"},
		{http.MethodPost, ts.URL + "/api/v0/session"},
		{http.MethodPatch, ts.URL + "/api/v0/session/close"},
		{http.MethodPatch, ts.URL + "/api/v0/session/re-open"},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/%s", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/open-session", ts.URL, uuid.New())},
		{http.MethodGet, fmt.Sprintf("%s/api/v0/session/archer/%s/participating", ts.URL, uuid.New())},
	}

	for _, ep := range endpoints {
		req := authRequest(ep.method, ep.url, "", nil)
		resp, _ := doJSONRequest(t, nil, req, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", ep.method, ep.url, resp.StatusCode)
		}
	}
}
