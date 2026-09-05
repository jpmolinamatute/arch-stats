package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions holds configurable settings for CORS.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// GetCORSOptions returns default CORS options depending on devMode.
func GetCORSOptions(devMode bool) CORSOptions {
	if devMode {
		return CORSOptions{
			AllowedOrigins: []string{
				"http://localhost:5173",
				"http://127.0.0.1:5173",
				"http://localhost:8000",
				"http://localhost:3000",
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
				http.MethodHead,
			},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-CSRF-Token",
				"Origin",
			},
			AllowCredentials: true,
			MaxAge:           300,
		}
	}

	return CORSOptions{
		AllowedOrigins: []string{},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Origin",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// CORS constructs an HTTP middleware that configures Cross-Origin Resource Sharing headers.
func CORS(devMode bool) func(http.Handler) http.Handler {
	opts := GetCORSOptions(devMode)
	allowedMethods := strings.Join(opts.AllowedMethods, ", ")
	allowedHeaders := strings.Join(opts.AllowedHeaders, ", ")
	maxAgeStr := strconv.Itoa(opts.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			isAllowed := false
			if devMode {
				// In development mode, allow localhost/127.0.0.1 origins or any if devMode
				isAllowed = true
			} else {
				for _, allowed := range opts.AllowedOrigins {
					if allowed == origin {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if opts.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == http.MethodOptions {
				if isAllowed {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
					w.Header().Set("Access-Control-Max-Age", maxAgeStr)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
