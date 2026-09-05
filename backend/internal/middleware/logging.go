package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status       int
	written      bool
	bytesWritten int64
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.written {
		r.status = status
		r.written = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.status = http.StatusOK
		r.written = true
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

// RequestLogger creates an HTTP middleware that records request execution details
// (method, path, status, duration, IP, user-agent) with slog.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			activeLogger := logger
			if activeLogger == nil {
				activeLogger = slog.Default()
			}

			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}

			activeLogger.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Duration("duration", duration),
				slog.Int64("bytes", rec.bytesWritten),
				slog.String("ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// Logging is a convenience middleware using the default slog logger.
func Logging(next http.Handler) http.Handler {
	return RequestLogger(nil)(next)
}
