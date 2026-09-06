package websocket

// Hub maintains the set of active WebSocket clients.
type Hub struct{}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {}
