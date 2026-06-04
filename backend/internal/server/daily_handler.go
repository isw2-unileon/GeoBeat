package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

// DailyService defines the interface for the daily challenge service.
type DailyService interface {
	GetCurrentStatus(ctx context.Context, userID uuid.UUID) (*daily.Challenge, *daily.Session, error)
	GetCurrentInverseStatus(ctx context.Context, userID uuid.UUID) (*daily.InverseChallenge, *daily.Session, error)
	ProcessAttempt(ctx context.Context, userID uuid.UUID, guess string) (*daily.AttemptResult, error)
	ProcessInverseAttempt(ctx context.Context, userID uuid.UUID, guess string) (*daily.InverseAttemptResult, error)
}

// Handler handles HTTP requests for the daily challenge endpoints.
type Handler struct {
	svc DailyService
}

// NewHandler creates a new Handler with the given DailyService.
func NewHandler(svc DailyService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers the HTTP routes for the daily challenge endpoints.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	getStatusHandler := http.HandlerFunc(h.getDailyStatus)
	getInverseStatusHandler := http.HandlerFunc(h.getInverseDailyStatus)
	postAttemptHandler := http.HandlerFunc(h.postAttempt)
	postInverseAttemptHandler := http.HandlerFunc(h.postInverseAttempt)

	// Step 2: Wrap them in the middleware and use mux.Handle (NOT HandleFunc)
	mux.Handle("GET /api/game/daily", authMiddleware(getStatusHandler))
	mux.Handle("GET /api/game/inverse", authMiddleware(getInverseStatusHandler))
	mux.Handle("POST /api/game/daily/attempt", authMiddleware(postAttemptHandler))
	mux.Handle("POST /api/game/inverse/attempt", authMiddleware(postInverseAttemptHandler))
}

// attemptRequest represents the expected request body for making an attempt at the daily challenge.
type attemptRequest struct {
	Guess string `json:"guess"`
}

// statusResponse represents the response structure for the daily status endpoint.
type statusResponse struct {
	Country      string `json:"country"`
	AttemptsUsed int    `json:"attempts_used"`
	Status       string `json:"status"`
}

// statusInverseResponse represents the response structure for the daily inverse status endpoint.
type statusInverseResponse struct {
	Song         string `json:"song"`
	AttemptsUsed int    `json:"attempts_used"`
	Status       string `json:"status"`
}

// getDailyStatus handles the GET /api/game/daily endpoint to retrieve the current status of the daily challenge for a user.
func (h *Handler) getDailyStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	challenge, session, err := h.svc.GetCurrentStatus(r.Context(), userID)
	if err != nil {
		dailyError(w, err)
		return
	}

	resp := statusResponse{
		Country:      challenge.TargetCountry,
		AttemptsUsed: session.AttemptsUsed,
		Status:       string(session.Status),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getInverseDailyStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	challenge, session, err := h.svc.GetCurrentInverseStatus(r.Context(), userID)
	if err != nil {
		dailyError(w, err)
		return
	}

	resp := statusInverseResponse{
		Song:         challenge.GivenSongName,
		AttemptsUsed: session.AttemptsUsed,
		Status:       string(session.Status),
	}

	writeJSON(w, http.StatusOK, resp)
}

// postAttempt handles the POST /api/game/daily/attempt endpoint to process a user's guess for the daily challenge.
func (h *Handler) postAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	var req attemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Result contains status
	result, err := h.svc.ProcessAttempt(r.Context(), userID, req.Guess)
	if err != nil {
		dailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) postInverseAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	var req attemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.ProcessInverseAttempt(r.Context(), userID, req.Guess)
	if err != nil {
		dailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// dailyError maps daily errors to appropriate HTTP responses.
func dailyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, daily.ErrChallengeNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, daily.ErrGameOver):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, daily.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// writeJSON writes the given value as a JSON response with the specified HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error coding JSON response: %v", err)
	}
}

// writeError writes a JSON error response with the given HTTP status code and message.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
