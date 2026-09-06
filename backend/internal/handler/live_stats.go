package handler

import (
	"context"
	"net/http"

	coderws "github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
	ws "github.com/jpmolinamatute/arch-stats/backend/internal/websocket"
)

// LiveStatsService defines the business operations required by LiveStatsHandler.
type LiveStatsService interface {
	GetStats(ctx context.Context, slotID, archerID uuid.UUID) (*model.LiveStat, error)
}

// WebSocketHub defines the hub operations required by LiveStatsHandler.
type WebSocketHub interface {
	Register(client *ws.Client)
	Unregister(client *ws.Client)
}

// LiveStatsHandler manages HTTP and WebSocket endpoints for real-time and aggregate slot statistics.
type LiveStatsHandler struct {
	svc LiveStatsService
	hub WebSocketHub
}

// NewLiveStatsHandler constructs a LiveStatsHandler with service and hub dependency injection.
func NewLiveStatsHandler(svc LiveStatsService, hub WebSocketHub) *LiveStatsHandler {
	return &LiveStatsHandler{
		svc: svc,
		hub: hub,
	}
}

// Routes registers live stats endpoints on the provided chi Router.
// The literal `/ws/{slot_id}` route is registered before `/{slot_id}` to ensure unambiguous route matching.
func (h *LiveStatsHandler) Routes(r chi.Router) {
	r.Get("/ws/{slot_id}", h.WebSocketStats)
	r.Get("/{slot_id}", h.GetStats)
}

// GetStats godoc
// @Summary     Get Stats
// @Description Returns current live aggregate statistics and recent shot scores for a slot
// @Tags        Stats
// @Produce     json
// @Param       slot_id path string true "Slot UUID"
// @Success     200 {object} model.LiveStat
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @Router      /stats/{slot_id} [get]
func (h *LiveStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		WriteAppError(w, err)
		return
	}

	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid slot_id UUID"))
		return
	}

	stat, err := h.svc.GetStats(r.Context(), slotID, authArcherID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, stat); err != nil {
		WriteAppError(w, err)
		return
	}
}

// WebSocketStats handles GET /api/v0/stats/ws/{slot_id}.
// Upgrades the HTTP request to a WebSocket connection, registers the client with the hub,
// starts WritePump and ReadPump goroutines, and unregisters the client when the connection closes.
func (h *LiveStatsHandler) WebSocketStats(w http.ResponseWriter, r *http.Request) {
	slotIDStr := getURLParam(r, "slot_id")
	if slotIDStr == "" {
		slotIDStr = getURLParam(r, "id")
	}

	if _, err := uuid.Parse(slotIDStr); err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "invalid slot_id UUID"))
		return
	}

	conn, err := coderws.Accept(w, r, &coderws.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		// Accept already writes the HTTP error response if handshake fails
		return
	}

	client := ws.NewClient(conn)
	if h.hub != nil {
		h.hub.Register(client)
		defer h.hub.Unregister(client)
	}

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)
}
