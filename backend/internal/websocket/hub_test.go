package websocket

import (
	"bytes"
	"context"
	"testing"
	"time"
)

// waitForCondition polls until condition is true or timeout expires.
func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	// Initially 0 clients
	if count := hub.ClientCount(); count != 0 {
		t.Fatalf("expected 0 clients, got %d", count)
	}

	// Register client1
	hub.Register(client1)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client after registering client1, got %d", hub.ClientCount())
	}
	if !hub.HasClient(client1) {
		t.Fatal("expected hub to have client1")
	}

	// Register client2
	hub.Register(client2)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 clients after registering client2, got %d", hub.ClientCount())
	}
	if !hub.HasClient(client2) {
		t.Fatal("expected hub to have client2")
	}

	// Unregister client1
	hub.Unregister(client1)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected 1 client after unregistering client1, got %d", hub.ClientCount())
	}
	if hub.HasClient(client1) {
		t.Fatal("expected hub to not have client1")
	}

	// Unregister client2
	hub.Unregister(client2)
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 0 }) {
		t.Fatalf("expected 0 clients after unregistering client2, got %d", hub.ClientCount())
	}
	if hub.HasClient(client2) {
		t.Fatal("expected hub to not have client2")
	}
}

func TestHub_Broadcast_ReachesAllRegisteredClients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)
	client3Unregistered := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)

	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 }) {
		t.Fatalf("expected 2 registered clients, got %d", hub.ClientCount())
	}

	payload := []byte(`{"event":"shot_insert","score":10}`)
	hub.Broadcast(payload)

	// Verify client1 received message
	select {
	case msg := <-client1.Send:
		if !bytes.Equal(msg, payload) {
			t.Fatalf("client1 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client1 timed out waiting for broadcast")
	}

	// Verify client2 received message
	select {
	case msg := <-client2.Send:
		if !bytes.Equal(msg, payload) {
			t.Fatalf("client2 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client2 timed out waiting for broadcast")
	}

	// Verify unregistered client3 did NOT receive message
	select {
	case msg := <-client3Unregistered.Send:
		t.Fatalf("unregistered client received unexpected message: %s", msg)
	default:
		// Succeeded, nothing received
	}
}

func TestHub_Broadcast_SkipsUnregisteredClients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Unregister client1
	hub.Unregister(client1)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 })

	// client1.Send should be closed
	_, ok := <-client1.Send
	if ok {
		t.Fatal("expected client1.Send to be closed upon unregister")
	}

	payload := []byte(`{"event":"update"}`)
	hub.Broadcast(payload)

	// Verify client2 receives message
	select {
	case msg := <-client2.Send:
		if !bytes.Equal(msg, payload) {
			t.Fatalf("client2 expected %s, got %s", payload, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("client2 timed out waiting for broadcast")
	}
}

func TestHub_SlowClientEviction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub("", "", nil)
	go hub.Run(ctx)

	// Create client with tiny buffer of 1 message
	slowClient := &Client{
		Send: make(chan []byte, 1),
	}
	normalClient := NewClient(nil)

	hub.Register(slowClient)
	hub.Register(normalClient)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Fill slowClient buffer to capacity
	slowClient.Send <- []byte("already full")

	// Broadcast should evict slowClient because its send buffer is full
	hub.Broadcast([]byte("broadcast payload"))

	// Normal client receives broadcast
	select {
	case msg := <-normalClient.Send:
		if string(msg) != "broadcast payload" {
			t.Fatalf("normalClient expected broadcast payload, got %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("normalClient timed out")
	}

	// slowClient should be evicted from hub
	if !waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 1 }) {
		t.Fatalf("expected slow client to be evicted, client count: %d", hub.ClientCount())
	}
	if hub.HasClient(slowClient) {
		t.Fatal("expected slowClient to be removed from hub")
	}
}

func TestHub_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	hub := NewHub("", "", nil)
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	client1 := NewClient(nil)
	client2 := NewClient(nil)

	hub.Register(client1)
	hub.Register(client2)
	waitForCondition(t, time.Second, func() bool { return hub.ClientCount() == 2 })

	// Cancel hub context
	cancel()

	select {
	case <-done:
		// Hub exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("hub did not exit cleanly on context cancellation")
	}

	if count := hub.ClientCount(); count != 0 {
		t.Fatalf("expected 0 clients after shutdown, got %d", count)
	}

	// Verify both client channels are closed
	if _, ok := <-client1.Send; ok {
		t.Fatal("expected client1.Send to be closed on shutdown")
	}
	if _, ok := <-client2.Send; ok {
		t.Fatal("expected client2.Send to be closed on shutdown")
	}
}

func TestHub_ListenNotify_ContextCancellation(t *testing.T) {
	// Tests that listenNotify respects context cancellation with an invalid/unreachable DSN
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	hub := NewHub("postgres://invalid_user:invalid_pass@127.0.0.1:54329/invalid_db", "test_channel", nil)
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Successfully exited on context timeout
	case <-time.After(2 * time.Second):
		t.Fatal("hub.Run did not terminate on context cancellation during listenNotify")
	}
}
