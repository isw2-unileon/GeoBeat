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
	mux.Handle("GET /api/game/daily", authMiddleware(http.HandlerFunc(h.getDailyStatus)))
	mux.Handle("GET /api/game/inverse", authMiddleware(http.HandlerFunc(h.getInverseDailyStatus)))
	mux.Handle("POST /api/game/daily/attempt", authMiddleware(http.HandlerFunc(h.postAttempt)))
	mux.Handle("POST /api/game/inverse/attempt", authMiddleware(http.HandlerFunc(h.postInverseAttempt)))
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

func handleStatusFetch[C any, R any](
	w http.ResponseWriter,
	r *http.Request,
	fetchFn func(ctx context.Context, uid uuid.UUID) (C, *daily.Session, error),
	buildRespFn func(challenge C, session *daily.Session) R,
) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	challenge, session, err := fetchFn(r.Context(), userID)
	if err != nil {
		dailyError(w, err)
		return
	}

	resp := buildRespFn(challenge, session)
	writeJSON(w, http.StatusOK, resp)
}

// getDailyStatus handles the GET /api/game/daily endpoint to retrieve the current status of the daily challenge for a user.
func (h *Handler) getDailyStatus(w http.ResponseWriter, r *http.Request) {
	handleStatusFetch(w, r, h.svc.GetCurrentStatus, func(challenge *daily.Challenge, session *daily.Session) statusResponse {
		return statusResponse{
			Country:      challenge.TargetCountry,
			AttemptsUsed: session.AttemptsUsed,
			Status:       string(session.Status),
		}
	})
}

func (h *Handler) getInverseDailyStatus(w http.ResponseWriter, r *http.Request) {
	handleStatusFetch(w, r, h.svc.GetCurrentInverseStatus, func(challenge *daily.InverseChallenge, session *daily.Session) statusInverseResponse {
		return statusInverseResponse{
			Song:         challenge.GivenSongName,
			AttemptsUsed: session.AttemptsUsed,
			Status:       string(session.Status),
		}
	})
}

func handleAttemptProcessing[R any](
	w http.ResponseWriter,
	r *http.Request,
	processFn func(ctx context.Context, userID uuid.UUID, guess string) (R, error),
) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	var req attemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := processFn(r.Context(), userID, req.Guess)
	if err != nil {
		dailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// postAttempt handles the POST /api/game/daily/attempt endpoint to process a user's guess for the daily challenge.
func (h *Handler) postAttempt(w http.ResponseWriter, r *http.Request) {
	handleAttemptProcessing(w, r, h.svc.ProcessAttempt)
}

func (h *Handler) postInverseAttempt(w http.ResponseWriter, r *http.Request) {
	handleAttemptProcessing(w, r, h.svc.ProcessInverseAttempt)
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
