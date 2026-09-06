package integration_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

func waitForCondition(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}

type wsTestEnv struct {
	srv     *httptest.Server
	hub     *ws.Hub
	channel string
	wsURL   string
	slotID  uuid.UUID
	cancel  context.CancelFunc
}

func setupWSTestEnv(t *testing.T, customChannel ...string) *wsTestEnv {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	channel := "archy"
	if len(customChannel) > 0 && customChannel[0] != "" {
		channel = customChannel[0]
	}

	hub := ws.NewHub(testDSN, channel, slog.Default())
	go hub.Run(ctx)

	slotID := uuid.New()
	liveStatsHandler := handler.NewLiveStatsHandler(nil, hub)

	r := chi.NewRouter()
	r.Route("/api/v0", func(r chi.Router) {
		r.Route("/stats", liveStatsHandler.Routes)
		r.Get("/live-stats", func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())
			rctx.URLParams.Add("slot_id", slotID.String())
			liveStatsHandler.WebSocketStats(w, r)
		})
	})

	srv := httptest.NewServer(r)

	t.Cleanup(func() {
		srv.Close()
		cancel()
		_ = truncateAll(context.Background(), testPool)
	})

	// Allow brief window for LISTEN connection to establish on PostgreSQL
	time.Sleep(100 * time.Millisecond)

	wsURL := fmt.Sprintf("ws%s/api/v0/stats/ws/%s", strings.TrimPrefix(srv.URL, "http"), slotID)

	return &wsTestEnv{
		srv:     srv,
		hub:     hub,
		channel: channel,
		wsURL:   wsURL,
		slotID:  slotID,
		cancel:  cancel,
	}
}

func TestWebSocket_SingleClientConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	// Connect WebSocket client
	conn, resp, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	// In coder/websocket, a reader must be active to process control frames (pong)
	clientReadDone := make(chan struct{})
	go func() {
		defer close(clientReadDone)
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				return
			}
		}
	}()

	// Verify connection is alive via ping
	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	defer pingCancel()
	if err := conn.Ping(pingCtx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	// Verify client is registered with the hub
	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client, got %d", env.hub.ClientCount())
	}

	// Close connection cleanly and verify unregistration
	_ = conn.Close(websocket.StatusNormalClosure, "done")
	<-clientReadDone
	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 registered clients after close, got %d", env.hub.ClientCount())
	}
}

func TestWebSocket_NotifyBroadcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client, got %d", env.hub.ClientCount())
	}

	// Fire pg_notify on the configured channel
	payload := fmt.Sprintf(`{"event":"shot_recorded","data":{"slot_id":%q,"score":10}}`, env.slotID)
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Read message from WebSocket client
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	_, msg, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("reading websocket message failed: %v", err)
	}

	if !strings.Contains(string(msg), "shot_recorded") {
		t.Errorf("message %q does not contain 'shot_recorded'", string(msg))
	}
	if !strings.Contains(string(msg), env.slotID.String()) {
		t.Errorf("message %q does not contain slot ID %s", string(msg), env.slotID)
	}
}

func TestWebSocket_MultipleClientsFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	const clientCount = 3
	conns := make([]*websocket.Conn, clientCount)
	t.Cleanup(func() {
		for _, c := range conns {
			if c != nil {
				_ = c.Close(websocket.StatusNormalClosure, "test done")
			}
		}
	})
	for i := 0; i < clientCount; i++ {
		conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
		if err != nil {
			t.Fatalf("client %d dial failed: %v", i, err)
		}
		conns[i] = conn
	}

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == clientCount }) {
		t.Fatalf("expected %d registered clients, got %d", clientCount, env.hub.ClientCount())
	}

	// Fire a single pg_notify event
	payload := `{"event":"fanout_test","data":{"status":"ok"}}`
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Verify all clients receive the exact same message
	for i, conn := range conns {
		readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
		_, msg, err := conn.Read(readCtx)
		readCancel()
		if err != nil {
			t.Fatalf("client %d failed to read broadcast: %v", i, err)
		}
		if string(msg) != payload {
			t.Errorf("client %d expected %q, got %q", i, payload, string(msg))
		}
	}
}

func TestWebSocket_ClientDisconnectCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn1, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("client 1 dial failed: %v", err)
	}
	defer conn1.Close(websocket.StatusNormalClosure, "test done")

	conn2, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("client 2 dial failed: %v", err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 registered clients, got %d", env.hub.ClientCount())
	}

	// Disconnect client 1
	_ = conn1.Close(websocket.StatusNormalClosure, "client 1 disconnecting")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 registered client after disconnect, got %d", env.hub.ClientCount())
	}

	// Fire pg_notify after client 1 disconnected
	payload := `{"event":"after_disconnect","remaining":1}`
	if _, err := testPool.Exec(ctx, "SELECT pg_notify($1, $2)", env.channel, payload); err != nil {
		t.Fatalf("pg_notify failed: %v", err)
	}

	// Verify remaining client (conn2) receives the notification with no panics
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	_, msg, err := conn2.Read(readCtx)
	if err != nil {
		t.Fatalf("client 2 failed to read notification: %v", err)
	}
	if string(msg) != payload {
		t.Errorf("client 2 expected %q, got %q", payload, string(msg))
	}
}

func TestWebSocket_InvalidUpgrade(t *testing.T) {
	env := setupWSTestEnv(t)

	t.Run("regular HTTP GET to valid slot without upgrade headers", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/stats/ws/" + env.slotID.String())
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Errorf("expected error status for non-WebSocket request, got %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUpgradeRequired {
			t.Errorf("expected status 400 or 426, got %d", resp.StatusCode)
		}
	})

	t.Run("regular HTTP GET to /api/v0/live-stats without upgrade headers", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/live-stats")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Errorf("expected error status for non-WebSocket request, got %d", resp.StatusCode)
		}
	})

	t.Run("GET with invalid slot_id UUID returns 422", func(t *testing.T) {
		resp, err := http.Get(env.srv.URL + "/api/v0/stats/ws/not-a-valid-uuid")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422 for invalid UUID, got %d", resp.StatusCode)
		}
	})
}

func TestWebSocket_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := setupWSTestEnv(t)

	conn, _, err := websocket.Dial(ctx, env.wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client, got %d", env.hub.ClientCount())
	}

	// Cancel the hub context to simulate graceful shutdown
	env.cancel()

	// Reading from client should return a close error or EOF
	readCtx, readCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readCancel()
	_, _, err = conn.Read(readCtx)
	if err == nil {
		t.Fatal("expected error reading after hub shutdown, got nil")
	}

	if !waitForCondition(2*time.Second, func() bool { return env.hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 clients after hub shutdown, got %d", env.hub.ClientCount())
	}
}
