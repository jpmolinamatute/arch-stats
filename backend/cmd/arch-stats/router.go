package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/config"
	"github.com/jpmolinamatute/arch-stats/backend/internal/handler"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
)

// RouterDeps encapsulates all router dependencies.
type RouterDeps struct {
	Cfg            *config.Config
	Logger         *slog.Logger
	AuthSvc        middleware.TokenAuthenticator
	AuthHandler    *handler.AuthHandler
	ArcherHandler  *handler.ArcherHandler
	SessionHandler *handler.SessionHandler
	SlotHandler    *handler.SlotHandler
	ShotHandler    *handler.ShotHandler
	FaceHandler    *handler.FaceHandler
	HealthHandler  *handler.HealthHandler
	SPAHandler     http.Handler
}

// buildRouter constructs the chi HTTP router with global middleware and nested route groups.
func buildRouter(deps *RouterDeps) chi.Router {
	r := chi.NewRouter()

	// Global middleware stack: logging -> recovery -> CORS
	r.Use(middleware.RequestLogger(deps.Logger))
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(deps.Cfg.DevMode))

	r.Route("/api/v0", func(r chi.Router) {
		// Public health check endpoint
		r.Get("/health", deps.HealthHandler.Health)

		// Public face catalog endpoints
		r.Route("/faces", deps.FaceHandler.Routes)

		// Auth route group: public login/google/register, protected logout/me
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.ErrorMapper)

			r.Post("/login", deps.AuthHandler.Login)
			r.Post("/google", deps.AuthHandler.Login)
			r.Post("/register", deps.AuthHandler.Register)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(deps.AuthSvc))
				r.Post("/logout", deps.AuthHandler.Logout)
				r.Get("/me", deps.AuthHandler.Me)
			})
		})

		// Protected route groups requiring valid authentication token
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthSvc))
			r.Use(middleware.ErrorMapper)

			r.Route("/archer", deps.ArcherHandler.Routes)
			r.Route("/session", func(r chi.Router) {
				r.Route("/slot", deps.SlotHandler.Routes)
				deps.SessionHandler.Routes(r)
			})
			r.Route("/shot", deps.ShotHandler.Routes)
		})
	})

	if deps.SPAHandler != nil {
		r.Handle("/*", deps.SPAHandler)
	}

	return r
}
