package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestClient_WritePump_DeliversMessages(t *testing.T) {
	// Create test server accepting websocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept failed: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "done")

		client := NewClient(conn)
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		go client.WritePump(ctx)

		// Send message through client.Send
		client.Send <- []byte("test message payload")

		// Keep connection open briefly for client read
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	// Client dials the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(msg) != "test message payload" {
		t.Fatalf("expected 'test message payload', got %q", string(msg))
	}
}

func TestClient_ReadPump_DetectsDisconnect(t *testing.T) {
	readPumpFinished := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept failed: %v", err)
			return
		}

		client := NewClient(conn)
		go func() {
			client.ReadPump(r.Context())
			close(readPumpFinished)
		}()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Close dial connection to trigger ReadPump exit on server
	_ = conn.Close(websocket.StatusNormalClosure, "client disconnect")

	select {
	case <-readPumpFinished:
		// Succeeded
	case <-time.After(3 * time.Second):
		t.Fatal("ReadPump did not detect disconnect within timeout")
	}
}

func TestClient_NilConn_SafeOperations(t *testing.T) {
	client := NewClient(nil)
	if client.Send == nil {
		t.Fatal("expected non-nil Send channel")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Both pumps should exit gracefully on context cancellation when Conn is nil
	done := make(chan struct{})
	go func() {
		client.WritePump(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("WritePump did not exit on canceled context with nil conn")
	}
}
