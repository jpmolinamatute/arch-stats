package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

// writeJSON marshals data as JSON, sets the Content-Type header to application/json,
// writes the HTTP status code, and writes the response body.
func writeJSON(w http.ResponseWriter, status int, data any) error {
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

// readJSON decodes the JSON request body into the target pointer dst.
// It limits request payload size to 1MB and wraps any decode error in apperror.ErrValidation.
func readJSON(r *http.Request, dst any) error {
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

// writeError writes an error response formatted as JSON with a status code and detail message.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(middleware.ErrorResponse{
		Detail: message,
	})
}

// writeAppError translates a domain error using middleware.WriteError into appropriate status code and JSON.
func writeAppError(w http.ResponseWriter, err error) {
	middleware.WriteError(w, err)
}

// WriteJSON is the exported alias for writeJSON.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, data)
}

// ReadJSON is the exported alias for readJSON.
func ReadJSON(r *http.Request, dst any) error {
	return readJSON(r, dst)
}

// WriteError is the exported alias for writeError.
func WriteError(w http.ResponseWriter, status int, message string) {
	writeError(w, status, message)
}

// WriteAppError is the exported alias for writeAppError.
func WriteAppError(w http.ResponseWriter, err error) {
	writeAppError(w, err)
}
