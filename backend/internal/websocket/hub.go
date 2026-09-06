package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
)

const (
	// defaultChannelBufferSize is the buffer size for hub management channels.
	defaultChannelBufferSize = 256

	// reconnectBackoff is the wait time between LISTEN connection retry attempts.
	reconnectBackoff = 1 * time.Second
)

// Hub maintains the set of active WebSocket clients and broadcasts PostgreSQL
// NOTIFY events to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	dsn     string
	channel string
	logger  *slog.Logger

	mu sync.RWMutex
}

// NewHub creates a new Hub instance.
// If dsn is non-empty, Run(ctx) will start a dedicated goroutine listening for PostgreSQL NOTIFY events.
// If logger is nil, slog.Default() is used.
func NewHub(dsn, channel string, logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}

	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, defaultChannelBufferSize),
		register:   make(chan *Client, defaultChannelBufferSize),
		unregister: make(chan *Client, defaultChannelBufferSize),
		dsn:        dsn,
		channel:    channel,
		logger:     logger,
	}
}

// Run starts the hub's main event loop and PostgreSQL notification listener.
// It blocks until the provided context is canceled, at which point it gracefully
// shuts down and disconnects all clients.
func (h *Hub) Run(ctx context.Context) {
	if h.dsn != "" && h.channel != "" {
		go h.listenNotify(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					h.logger.Warn("client send channel full, evicting slow client")
					delete(h.clients, client)
					close(client.Send)
					if client.Conn != nil {
						_ = client.Conn.Close(websocket.StatusPolicyViolation, "send buffer full")
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// Register registers a new client for receiving broadcasts.
func (h *Hub) Register(client *Client) {
	client.hub = h
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	default:
		// Fallback for non-blocking unregister if channel is saturated or hub shutting down
		h.mu.Lock()
		if _, ok := h.clients[client]; ok {
			delete(h.clients, client)
			close(client.Send)
		}
		h.mu.Unlock()
	}
}

// Broadcast distributes a payload to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("hub broadcast channel full, dropping message")
	}
}

// ClientCount returns the number of active clients registered with the hub.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HasClient checks whether a client is currently registered with the hub.
func (h *Hub) HasClient(client *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[client]
}

// shutdown disconnects all clients and closes their send channels.
func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		delete(h.clients, client)
		close(client.Send)
		if client.Conn != nil {
			_ = client.Conn.Close(websocket.StatusNormalClosure, "hub shutdown")
		}
	}
	h.logger.Info("websocket hub shut down cleanly")
}

// listenNotify runs on a dedicated goroutine, holding an open dedicated pgx connection
// (outside the connection pool) to execute LISTEN and receive NOTIFY events.
func (h *Hub) listenNotify(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := pgx.Connect(ctx, h.dsn)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			h.logger.Error("failed to connect dedicated pgx connection for LISTEN", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
				continue
			}
		}

		listenErr := h.listenLoop(ctx, conn)
		_ = conn.Close(context.Background())

		if ctx.Err() != nil {
			return
		}

		if listenErr != nil {
			h.logger.Warn("LISTEN connection lost, reconnecting...", "error", listenErr)
			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectBackoff):
			}
		}
	}
}

// listenLoop executes the LISTEN statement and waits for notifications.
func (h *Hub) listenLoop(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", pgx.Identifier{h.channel}.Sanitize()))
	if err != nil {
		return fmt.Errorf("executing LISTEN statement: %w", err)
	}

	h.logger.Info("listening for PostgreSQL notifications", "channel", h.channel)

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("waiting for notification: %w", err)
		}

		h.Broadcast([]byte(notification.Payload))
	}
}
