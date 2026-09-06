// Package handler provides HTTP request handling for the arch-stats API.
//
// Handler Response Conventions:
//
//   - 201 Created: return the created entity's ID struct (e.g., model.ArcherID, model.SessionID).
//   - 200 OK: return the resource or a status struct for state mutations.
//   - 204 No Content: return no body for destructive operations (delete, leave).
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// WriteJSON marshals data as JSON, sets the Content-Type header to application/json,
// writes the HTTP status code, and writes the response body.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil || status == http.StatusNoContent {
		return nil
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encoding json response: %w", err)
	}
	return nil
}

// ReadJSON decodes the JSON request body into the target pointer dst.
// It limits request payload size to 1MB and wraps any decode error in apperror.ErrValidation.
func ReadJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperror.Wrap(apperror.ErrValidation, "request body is empty")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, 1048576))
	if err := dec.Decode(dst); err != nil {
		return apperror.Wrap(apperror.ErrValidation, fmt.Sprintf("invalid request body: %v", err))
	}
	return nil
}

// WriteError writes an error response formatted as JSON with a status code and detail message.
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{
		Detail: message,
	})
}

// WriteAppError translates a domain error using middleware.WriteError into appropriate status code and JSON.
func WriteAppError(w http.ResponseWriter, err error) {
	middleware.WriteError(w, err)
}

// getURLParam extracts a URL parameter from the chi route context,
// falling back to the standard library PathValue.
func getURLParam(r *http.Request, key string) string {
	if val := chi.URLParam(r, key); val != "" {
		return val
	}
	return r.PathValue(key)
}

// parseUUIDParam extracts a UUID from the URL path, trying paramNames in order.
// Writes a 422 validation error response and returns uuid.Nil, false on failure.
func parseUUIDParam(w http.ResponseWriter, r *http.Request, paramNames ...string) (uuid.UUID, bool) {
	var raw string
	for _, name := range paramNames {
		if raw = getURLParam(r, name); raw != "" {
			break
		}
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "valid "+paramNames[0]+" is required"))
		return uuid.Nil, false
	}

	return id, true
}

// requireOwnership verifies the authenticated archer matches the archer_id in the URL.
// Writes the error response and returns uuid.Nil, false on failure.
func requireOwnership(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	authArcherID, err := middleware.GetArcherID(r.Context())
	if err != nil {
		WriteAppError(w, err)
		return uuid.Nil, false
	}

	archerID, ok := parseUUIDParam(w, r, "archer_id", "id")
	if !ok {
		return uuid.Nil, false
	}

	if authArcherID != archerID {
		WriteAppError(w, apperror.Wrap(apperror.ErrForbidden, "forbidden"))
		return uuid.Nil, false
	}

	return archerID, true
}
