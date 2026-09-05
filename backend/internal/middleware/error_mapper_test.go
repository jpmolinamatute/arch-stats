package middleware_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantStatus    int
		wantCode      string
		wantDetailSub string
	}{
		{
			name:          "ErrNotFound",
			err:           apperror.ErrNotFound,
			wantStatus:    http.StatusNotFound,
			wantCode:      "NOT_FOUND",
			wantDetailSub: "not found",
		},
		{
			name:          "Wrapped ErrNotFound",
			err:           apperror.Wrap(apperror.ErrNotFound, "archer 123"),
			wantStatus:    http.StatusNotFound,
			wantCode:      "NOT_FOUND",
			wantDetailSub: "archer 123",
		},
		{
			name:          "ErrUnauthorized",
			err:           apperror.ErrUnauthorized,
			wantStatus:    http.StatusUnauthorized,
			wantCode:      "UNAUTHORIZED",
			wantDetailSub: "unauthorized",
		},
		{
			name:          "Wrapped ErrUnauthorized",
			err:           apperror.Wrap(apperror.ErrUnauthorized, "session expired"),
			wantStatus:    http.StatusUnauthorized,
			wantCode:      "UNAUTHORIZED",
			wantDetailSub: "session expired",
		},
		{
			name:          "ErrForbidden",
			err:           apperror.ErrForbidden,
			wantStatus:    http.StatusForbidden,
			wantCode:      "FORBIDDEN",
			wantDetailSub: "forbidden",
		},
		{
			name:          "Wrapped ErrForbidden",
			err:           apperror.Wrap(apperror.ErrForbidden, "not your session"),
			wantStatus:    http.StatusForbidden,
			wantCode:      "FORBIDDEN",
			wantDetailSub: "not your session",
		},
		{
			name:          "ErrConflict",
			err:           apperror.ErrConflict,
			wantStatus:    http.StatusConflict,
			wantCode:      "CONFLICT",
			wantDetailSub: "conflict",
		},
		{
			name:          "Wrapped ErrConflict",
			err:           apperror.Wrap(apperror.ErrConflict, "slot already taken"),
			wantStatus:    http.StatusConflict,
			wantCode:      "CONFLICT",
			wantDetailSub: "slot already taken",
		},
		{
			name:          "ErrValidation",
			err:           apperror.ErrValidation,
			wantStatus:    http.StatusUnprocessableEntity,
			wantCode:      "VALIDATION",
			wantDetailSub: "validation",
		},
		{
			name:          "Wrapped ErrValidation",
			err:           apperror.Wrap(apperror.ErrValidation, "invalid arrow score"),
			wantStatus:    http.StatusUnprocessableEntity,
			wantCode:      "VALIDATION",
			wantDetailSub: "invalid arrow score",
		},
		{
			name:          "Unknown standard error",
			err:           errors.New("database connection refused"),
			wantStatus:    http.StatusInternalServerError,
			wantCode:      "INTERNAL_ERROR",
			wantDetailSub: "internal server error",
		},
		{
			name:          "Nil error",
			err:           nil,
			wantStatus:    http.StatusOK,
			wantCode:      "",
			wantDetailSub: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, resp := middleware.MapError(tc.err)
			if status != tc.wantStatus {
				t.Errorf("MapError status = %d, want %d", status, tc.wantStatus)
			}
			if resp.Code != tc.wantCode {
				t.Errorf("MapError code = %q, want %q", resp.Code, tc.wantCode)
			}
			if tc.wantDetailSub != "" && !strings.Contains(resp.Detail, tc.wantDetailSub) {
				t.Errorf("MapError detail = %q, want substring %q", resp.Detail, tc.wantDetailSub)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := apperror.Wrap(apperror.ErrNotFound, "archer not found")

	middleware.WriteError(rec, err)

	if rec.Code != http.StatusNotFound {
		t.Errorf("WriteError status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("WriteError Content-Type = %q, want application/json", ct)
	}

	var resp middleware.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("WriteError response is not valid JSON: %v", err)
	}
	if resp.Code != "NOT_FOUND" {
		t.Errorf("WriteError response Code = %q, want NOT_FOUND", resp.Code)
	}
	if resp.Detail != "not found: archer not found" {
		t.Errorf("WriteError response Detail = %q, want %q", resp.Detail, "not found: archer not found")
	}
}

func TestErrorMapper_MiddlewarePassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	wrapped := middleware.ErrorMapper(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ErrorMapper status = %d, want %d", rec.Code, http.StatusOK)
	}
}
