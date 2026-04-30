package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type OAuthProvider interface {
	GetAuthURL(state string) string
}

type AuthService interface {
	RegisterWithEmail(ctx context.Context, email, username, password string) error
	LoginWithEmail(ctx context.Context, email, password string) (string, error)
	ProcessOAuthLogin(ctx context.Context, code, provider user.AuthProvider) (string, error)
}

type AuthHandler struct {
	authService service.AuthService
	providers   map[user.AuthProvider]OAuthProvider
}

func NewAuthHandler(authService service.AuthService, providers map[user.AuthProvider]OAuthProvider) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		providers:   providers,
	}
}

func (h *AuthHandler) RegisterMiddleware(mux *http.ServeMux) {
	// This is where you would add any middleware for authentication, logging, etc.
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.handleRegister)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
	for provider := range h.providers {
		mux.HandleFunc("GET /api/auth/login/"+string(provider), func(w http.ResponseWriter, r *http.Request) {
			h.handleOAuthLogin(w, r, provider)
		})
		mux.HandleFunc("GET /api/auth/login/callback/"+string(provider), func(w http.ResponseWriter, r *http.Request) {
			h.handleOAuthRedirect(w, r, provider)
		})
	}
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {}

func (h *AuthHandler) handleOAuthLogin(w http.ResponseWriter, r *http.Request, provider user.AuthProvider) {
}

func (h *AuthHandler) handleOAuthRedirect(w http.ResponseWriter, r *http.Request, provider user.AuthProvider) {
}

func mapErrors(w http.ResponseWriter, err error) {
	switch err {
	case user.ErrEmailNotVerified, user.ErrEmptyPassword, user.ErrEmptyEmailOrUsername, user.ErrAccountAlreadyLinked:
		formatError(w, http.StatusBadRequest, err.Error())
	case service.ErrInvalidCredentials:
		formatError(w, http.StatusUnauthorized, err.Error())
	case service.ErrPasswordTooWeak:
		formatError(w, http.StatusBadRequest, err.Error())
	case service.ErrUserAlreadyExists:
		formatError(w, http.StatusConflict, service.ErrUserCreationFailed.Error())
	case service.ErrOAuthOnlyAccount:
		formatError(w, http.StatusConflict, service.ErrUserLoginFailed.Error())
	case service.ErrUserCreationFailed, service.ErrUserLoginFailed:
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
