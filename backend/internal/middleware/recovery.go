package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery returns an HTTP middleware that catches unhandled panics,
// logs the stack trace with slog.Error, and writes a standardized 500 JSON error response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				slog.ErrorContext(r.Context(), "panic recovered in HTTP handler",
					slog.Any("panic", rec),
					slog.String("stack", stack),
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
				)

				WriteError(w, fmt.Errorf("panic: %v", rec))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
