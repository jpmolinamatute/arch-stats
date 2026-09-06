// Package websocket provides a real-time WebSocket hub and client connections.
package websocket

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

const (
	// defaultSendBufferSize is the buffer size for a client's outbound message channel.
	defaultSendBufferSize = 256

	// writeTimeout is the maximum duration to wait for a write to complete.
	writeTimeout = 5 * time.Second
)

// Client represents a single connected WebSocket client.
type Client struct {
	// hub is the central hub managing this client.
	hub *Hub

	// Conn is the underlying WebSocket connection.
	Conn *websocket.Conn

	// Send is the buffered channel of outbound messages to be delivered to the client.
	Send chan []byte
}

// NewClient creates a new Client with a buffered send channel.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, defaultSendBufferSize),
	}
}

// WritePump pumps messages from the client's Send channel to the WebSocket connection.
// It terminates when the context is canceled, the Send channel is closed, or a write error occurs.
func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.Send:
			if !ok {
				// Hub closed the channel
				if c.Conn != nil {
					_ = c.Conn.Close(websocket.StatusNormalClosure, "hub closed channel")
				}
				return
			}

			if c.Conn == nil {
				continue
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.Conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				if c.hub != nil {
					c.hub.Unregister(c)
				}
				return
			}
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to detect client disconnect.
// It blocks until the connection is closed or an error occurs, then cleans up the client.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		if c.hub != nil {
			c.hub.Unregister(c)
		}
		if c.Conn != nil {
			_ = c.Conn.Close(websocket.StatusNormalClosure, "read pump closed")
		}
	}()

	if c.Conn == nil {
		<-ctx.Done()
		return
	}

	for {
		_, _, err := c.Conn.Read(ctx)
		if err != nil {
			return
		}
	}
}
