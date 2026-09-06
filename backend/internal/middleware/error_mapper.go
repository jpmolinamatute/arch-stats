package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// MapError maps domain errors to their corresponding HTTP status codes and response bodies.
func MapError(err error) (int, model.ErrorResponse) {
	if err == nil {
		return http.StatusOK, model.ErrorResponse{}
	}

	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code() {
		case apperror.ErrNotFound.Code():
			return http.StatusNotFound, model.ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrUnauthorized.Code():
			return http.StatusUnauthorized, model.ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrForbidden.Code():
			return http.StatusForbidden, model.ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrConflict.Code():
			return http.StatusConflict, model.ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		case apperror.ErrValidation.Code():
			return http.StatusUnprocessableEntity, model.ErrorResponse{
				Detail: appErr.Error(),
				Code:   appErr.Code(),
			}
		}
	}

	return http.StatusInternalServerError, model.ErrorResponse{
		Detail: "internal server error",
		Code:   "INTERNAL_ERROR",
	}
}

// WriteError serializes and writes an error response formatted as JSON.
func WriteError(w http.ResponseWriter, err error) {
	status, resp := MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ErrorMapper is an HTTP middleware that serves as a boundary filter.
func ErrorMapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
