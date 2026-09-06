package handler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpmolinamatute/arch-stats/backend/internal/apperror"
	"github.com/jpmolinamatute/arch-stats/backend/internal/auth"
	"github.com/jpmolinamatute/arch-stats/backend/internal/middleware"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

// AuthService defines the authentication operations required by the HTTP auth handler.
type AuthService interface {
	LoginWithGoogle(ctx context.Context, credential string, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, *model.AuthNeedsRegistration, error)
	RegisterWithGoogle(ctx context.Context, payload model.AuthRegistrationRequest, now time.Time, meta ...auth.SessionMetadata) (*model.AuthAuthenticated, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	DecodeToken(tokenStr string) (*auth.Claims, error)
	Authenticate(ctx context.Context, token string) (uuid.UUID, error)
}

// AuthHandlerConfig specifies runtime configuration for cookie management and token lifetimes.
type AuthHandlerConfig struct {
	JWTTTLMinutes int
	DevMode       bool
}

// AuthHandler manages HTTP endpoints for authentication, registration, session revocation, and current user retrieval.
type AuthHandler struct {
	authSvc   AuthService
	archerSvc ArcherService
	cfg       AuthHandlerConfig
}

// NewAuthHandler constructs an AuthHandler with service dependencies and configuration.
func NewAuthHandler(authSvc AuthService, archerSvc ArcherService, cfg AuthHandlerConfig) *AuthHandler {
	if cfg.JWTTTLMinutes <= 0 {
		cfg.JWTTTLMinutes = 1440
	}
	return &AuthHandler{
		authSvc:   authSvc,
		archerSvc: archerSvc,
		cfg:       cfg,
	}
}

// Login godoc
// @Summary     Google One Tap Login
// @Description Authenticate via Google One Tap credential. For existing archers, sets the auth cookie and returns credentials. For new archers, returns registration requirement.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body model.GoogleOneTapRequest true "Google One Tap credential"
// @Success     200 {object} model.AuthAuthenticated
// @Failure     422 {object} model.ErrorResponse
// @Router      /auth/login [post]
// @Router      /auth/google [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.GoogleOneTapRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteAppError(w, err)
		return
	}

	if strings.TrimSpace(req.Credential) == "" {
		WriteAppError(w, apperror.Wrap(apperror.ErrValidation, "credential is required"))
		return
	}

	now := time.Now().UTC()
	meta := extractSessionMetadata(r)

	authd, needsReg, err := h.authSvc.LoginWithGoogle(r.Context(), req.Credential, now, meta)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	if needsReg != nil {
		_ = WriteJSON(w, http.StatusOK, needsReg)
		return
	}

	h.setAuthCookie(w, r, authd.AccessToken, authd.ExpiresAt)
	_ = WriteJSON(w, http.StatusOK, authd)
}

// Register godoc
// @Summary     Register
// @Description Register a new archer profile with Google One Tap credential and demographic details
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body model.AuthRegistrationRequest true "Registration details"
// @Success     200 {object} model.AuthAuthenticated
// @Success     201 {object} model.AuthAuthenticated
// @Failure     422 {object} model.ErrorResponse
// @Security    BearerAuth
// @ID          register_api_v0_auth_register_post
// @Router      /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRegistrationRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteAppError(w, err)
		return
	}

	if err := validateRegistrationRequest(&req); err != nil {
		WriteAppError(w, err)
		return
	}

	now := time.Now().UTC()
	meta := extractSessionMetadata(r)

	authd, err := h.authSvc.RegisterWithGoogle(r.Context(), req, now, meta)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	h.setAuthCookie(w, r, authd.AccessToken, authd.ExpiresAt)
	_ = WriteJSON(w, http.StatusCreated, authd)
}

// Logout godoc
// @Summary     Logout
// @Description Deletes the auth cookie and revokes the active session token in the database
// @Tags        Auth
// @Produce     json
// @Success     200 {object} model.LogoutResponse
// @Failure     401 {object} model.ErrorResponse
// @Security    BearerAuth
// @ID          logout_api_v0_auth_logout_post
// @Router      /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := middleware.ExtractToken(r)
	if token != "" {
		_ = h.authSvc.RevokeToken(r.Context(), token)
	}

	h.clearAuthCookie(w, r)
	_ = WriteJSON(w, http.StatusOK, model.LogoutResponse{Success: true})
}

// Me godoc
// @Summary     Get Current User
// @Description Returns the currently authenticated archer from the request context or token cookie
// @Tags        Auth
// @Produce     json
// @Success     200 {object} model.AuthAuthenticated
// @Failure     401 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Security    BearerAuth
// @ID          get_current_user_api_v0_auth_me_get
// @Router      /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	archerID, err := middleware.GetArcherID(r.Context())
	token := middleware.ExtractToken(r)

	if err != nil {
		if token == "" {
			WriteAppError(w, apperror.Wrap(apperror.ErrUnauthorized, "no authentication token or context found"))
			return
		}

		authenticatedID, authErr := h.authSvc.Authenticate(r.Context(), token)
		if authErr != nil {
			WriteAppError(w, apperror.Wrap(apperror.ErrUnauthorized, authErr.Error()))
			return
		}
		archerID = authenticatedID
	}

	archer, err := h.archerSvc.GetByID(r.Context(), archerID)
	if err != nil {
		WriteAppError(w, err)
		return
	}

	var expiresAt time.Time
	if token != "" {
		if claims, err := h.authSvc.DecodeToken(token); err == nil && claims != nil {
			expiresAt = time.Unix(claims.Exp, 0).UTC()
		}
	}

	resp := model.AuthAuthenticated{
		Status:      model.AuthStatusAuthenticated,
		AccessToken: token,
		ExpiresAt:   expiresAt,
		Archer:      *archer,
	}

	_ = WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	maxAge := h.cfg.JWTTTLMinutes * 60
	if !expiresAt.IsZero() {
		if remaining := int(time.Until(expiresAt).Seconds()); remaining > 0 {
			maxAge = remaining
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.isCookieSecure(r),
	})
}

func (h *AuthHandler) clearAuthCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.isCookieSecure(r),
	})
}

func (h *AuthHandler) isCookieSecure(r *http.Request) bool {
	if !h.cfg.DevMode {
		return true
	}
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func extractSessionMetadata(r *http.Request) auth.SessionMetadata {
	var ua *string
	if val := strings.TrimSpace(r.Header.Get("User-Agent")); val != "" {
		ua = &val
	}

	var ip *string
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		first := strings.TrimSpace(parts[0])
		if first != "" {
			ip = &first
		}
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		trimmed := strings.TrimSpace(xri)
		if trimmed != "" {
			ip = &trimmed
		}
	} else if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && host != "" {
			ip = &host
		} else {
			trimmed := strings.TrimSpace(r.RemoteAddr)
			if trimmed != "" {
				ip = &trimmed
			}
		}
	}

	return auth.SessionMetadata{
		UserAgent: ua,
		IPAddress: ip,
	}
}

func validateRegistrationRequest(req *model.AuthRegistrationRequest) error {
	if strings.TrimSpace(req.Credential) == "" {
		return apperror.Wrap(apperror.ErrValidation, "credential is required")
	}
	if strings.TrimSpace(req.DateOfBirth) == "" {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth is required")
	}
	if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
		return apperror.Wrap(apperror.ErrValidation, "date_of_birth must be formatted as YYYY-MM-DD")
	}
	if !isValidGender(req.Gender) {
		return apperror.Wrap(apperror.ErrValidation, "invalid gender")
	}
	if !isValidBowstyle(req.Bowstyle) {
		return apperror.Wrap(apperror.ErrValidation, "invalid bowstyle")
	}
	if req.DrawWeight <= 0 || req.DrawWeight > 200 {
		return apperror.Wrap(apperror.ErrValidation, "draw_weight must be between 0 and 200")
	}
	return nil
}

func isValidGender(g model.Gender) bool {
	switch g {
	case model.GenderMale, model.GenderFemale, model.GenderNonBinary, model.GenderOther, model.GenderUnspecified:
		return true
	default:
		return false
	}
}

func isValidBowstyle(b model.Bowstyle) bool {
	switch b {
	case model.BowstyleRecurve, model.BowstyleCompound, model.BowstyleBarebow, model.BowstyleLongbow:
		return true
	default:
		return false
	}
}
