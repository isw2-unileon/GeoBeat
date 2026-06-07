package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
	"github.com/isw2-unileon/GeoBeat/backend/internal/config"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

// OAuthProvider defines the interface for OAuth providers to generate authentication URLs.
type OAuthProvider interface {
	GetAuthURL(state string) string
}

// AuthService defines the interface for authentication-related operations used by the AuthHandler.
type AuthService interface {
	RegisterWithEmail(ctx context.Context, email, username, password string) error
	LoginWithEmail(ctx context.Context, email, password string) (string, string, error)
	ProcessOAuthLogin(ctx context.Context, code string, provider geouser.AuthProvider) (string, error)
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
	Logout(ctx context.Context, rawToken string) error
	RefreshToken(ctx context.Context, rawToken string) (string, error)
}

// AuthHandler handles authentication-related HTTP requests, including registration, login, and OAuth flows.
type AuthHandler struct {
	authService AuthService
	providers   map[geouser.AuthProvider]OAuthProvider
	cfg         *config.Config
}

type contextKey string

// UserIDContextKey is the context key used to store the authenticated user's ID in the request context.
const UserIDContextKey = contextKey("userID")

// NewAuthHandler creates a new AuthHandler with the given authentication service and OAuth providers.
func NewAuthHandler(authService AuthService, providers map[geouser.AuthProvider]OAuthProvider, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		providers:   providers,
		cfg:         cfg,
	}
}

// RegisterRoutes registers the authentication-related routes on the provided HTTP mux.
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.handleRegister)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", h.handleLogout)
	mux.HandleFunc("POST /api/auth/refresh", h.handleRefresh)
	for provider := range h.providers {
		p := provider
		mux.HandleFunc("GET /api/auth/login/"+string(p), func(w http.ResponseWriter, r *http.Request) {
			h.handleOAuthLogin(w, r, p)
		})
		mux.HandleFunc("GET /api/auth/login/callback/"+string(p), func(w http.ResponseWriter, r *http.Request) {
			h.handleOAuthRedirect(w, r, p)
		})
	}
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		formatError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.RegisterWithEmail(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		mapErrors(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		formatError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, refreshToken, err := h.authService.LoginWithEmail(r.Context(), req.Email, req.Password)
	if err != nil {
		mapErrors(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(time.Hour) * 24,
	})

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) handleOAuthLogin(w http.ResponseWriter, r *http.Request, provider geouser.AuthProvider) {
	oauthProvider := h.providers[provider]
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		formatError(w, http.StatusInternalServerError, "failed to generate oauth state")
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	setOAuthStateCookie(w, state, provider, r.TLS != nil)
	authURL := oauthProvider.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) handleOAuthRedirect(w http.ResponseWriter, r *http.Request, provider geouser.AuthProvider) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		formatError(w, http.StatusBadRequest, fmt.Sprintf("oauth provider error: %s", errParam))
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		formatError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	cookieState, err := getOAuthStateFromCookie(r, provider)
	if err != nil || cookieState != state {
		formatError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	clearOAuthStateCookie(w, provider, r.TLS != nil)

	code := r.URL.Query().Get("code")
	if code == "" {
		formatError(w, http.StatusBadRequest, "missing oauth code")
		return
	}

	refreshToken, err := h.authService.ProcessOAuthLogin(r.Context(), code, provider)
	if err != nil {
		mapErrors(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(time.Hour) * 24,
	})

	http.Redirect(w, r, h.cfg.FrontendURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		formatError(w, http.StatusBadRequest, "missing refresh token")
		return
	}

	err = h.authService.Logout(r.Context(), cookie.Value)
	if err != nil {
		mapErrors(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		formatError(w, http.StatusBadRequest, "missing refresh token")
		return
	}

	newToken, err := h.authService.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		mapErrors(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": newToken})
}

// AuthMiddleware is an HTTP middleware that validates the presence and validity of a Bearer token in the Authorization header.
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			formatError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			formatError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		token := parts[1]
		userID, err := h.authService.ValidateToken(r.Context(), token)
		if err != nil || userID == uuid.Nil {
			formatError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), UserIDContextKey, userID))
		next.ServeHTTP(w, r)
	})
}

func oauthStateCookieName(provider geouser.AuthProvider) string {
	return "oauth_state_" + string(provider)
}

func setOAuthStateCookie(w http.ResponseWriter, state string, provider geouser.AuthProvider, secure bool) {
	// #nosec G124 - Secure flag is dynamically configured based on deployment environment environment
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName(provider),
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   300,
	})
}

func getOAuthStateFromCookie(r *http.Request, provider geouser.AuthProvider) (string, error) {
	cookie, err := r.Cookie(oauthStateCookieName(provider))
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func clearOAuthStateCookie(w http.ResponseWriter, provider geouser.AuthProvider, secure bool) {
	// #nosec G124 - Secure flag is dynamically configured based on deployment environment environment
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName(provider),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}

func mapErrors(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authsession.ErrExpired), errors.Is(err, geouser.ErrEmailNotVerified), errors.Is(err, geouser.ErrEmptyPassword), errors.Is(err, geouser.ErrEmptyEmailOrUsername), errors.Is(err, geouser.ErrAccountAlreadyLinked):
		formatError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		formatError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrPasswordTooWeak):
		formatError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrUserAlreadyExists):
		formatError(w, http.StatusConflict, service.ErrUserCreationFailed.Error())
	case errors.Is(err, service.ErrAlreadyLoggedIn):
		formatError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrOAuthOnlyAccount):
		formatError(w, http.StatusConflict, service.ErrUserLoginFailed.Error())
	case errors.Is(err, service.ErrUserCreationFailed), errors.Is(err, service.ErrUserLoginFailed):
	default:
		formatError(w, http.StatusInternalServerError, err.Error())
	}
}

func formatError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("error encoding JSON error response: %v", err)
	}
}
