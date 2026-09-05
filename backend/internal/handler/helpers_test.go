package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestWriteJSON(t *testing.T) {
	t.Run("writes JSON response with headers and status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		err := handler.WriteJSON(rec, http.StatusOK, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body["message"] != "success" {
			t.Fatalf("expected message 'success', got %q", body["message"])
		}
	})

	t.Run("handles 204 No Content without body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := handler.WriteJSON(rec, http.StatusNoContent, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("expected empty body, got %s", rec.Body.String())
		}
	})
}

func TestReadJSON(t *testing.T) {
	type samplePayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("successfully decodes valid JSON", func(t *testing.T) {
		body := `{"name":"Robin","age":30}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Name != "Robin" || dst.Age != 30 {
			t.Fatalf("unexpected decoded struct: %+v", dst)
		}
	})

	t.Run("returns ErrValidation on empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("returns ErrValidation on malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json"))
		var dst samplePayload
		err := handler.ReadJSON(req, &dst)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestWriteErrorAndWriteAppError(t *testing.T) {
	t.Run("writeError writes custom status and detail", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteError(rec, http.StatusBadRequest, "custom error message")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var resp middleware.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if resp.Detail != "custom error message" {
			t.Fatalf("expected detail 'custom error message', got %q", resp.Detail)
		}
	})

	t.Run("writeAppError maps ErrValidation to 422", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteAppError(rec, apperror.Wrap(apperror.ErrValidation, "field required"))

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("writeAppError maps ErrUnauthorized to 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.WriteAppError(rec, apperror.Wrap(apperror.ErrUnauthorized, "invalid credentials"))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
