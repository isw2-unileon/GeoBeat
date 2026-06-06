package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

// TimetrialService defines the interface for the time trial challenge service.
type TimetrialService interface {
	GetCurrentStatus(ctx context.Context, userID uuid.UUID) (*timetrial.Session, string, error)
	StartGame(ctx context.Context, userID uuid.UUID) (*timetrial.Session, string, error)
	ProcessAttempt(ctx context.Context, userID uuid.UUID, guess string) (*timetrial.AttemptResult, error)
	GetLeaderboard(ctx context.Context, userID uuid.UUID) (*timetrial.Leaderboard, error)
}

// TimetrialHandler handles HTTP requests for the time trial challenge endpoints.
type TimetrialHandler struct {
	svc TimetrialService
}

// NewTimetrialHandler creates a new TimetrialHandler with the given TimetrialService.
func NewTimetrialHandler(svc TimetrialService) *TimetrialHandler {
	return &TimetrialHandler{svc: svc}
}

// RegisterRoutes registers the HTTP routes for the time trial endpoints.
func (h *TimetrialHandler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	getStatusHandler := http.HandlerFunc(h.getStatus)
	startGameHandler := http.HandlerFunc(h.startGame)
	postAttemptHandler := http.HandlerFunc(h.postAttempt)
	getLeaderboardHandler := http.HandlerFunc(h.getLeaderboard)

	mux.Handle("GET /api/game/timetrial/status", authMiddleware(getStatusHandler))
	mux.Handle("POST /api/game/timetrial/start", authMiddleware(startGameHandler))
	mux.Handle("POST /api/game/timetrial/attempt", authMiddleware(postAttemptHandler))
	mux.Handle("GET /api/game/timetrial/leaderboard", authMiddleware(getLeaderboardHandler))
}

// timetrialAttemptRequest represents the expected request body for making a time trial guess.
type timetrialAttemptRequest struct {
	Guess string `json:"guess"`
}

// timetrialStatusResponse represents the response structure mapped for the frontend.
type timetrialStatusResponse struct {
	Status        string    `json:"status"`
	TargetCountry string    `json:"target_country,omitempty"`
	StartTime     time.Time `json:"start_time"`
	DurationMs    int64     `json:"duration_ms,omitempty"`
}

// handleTimetrialFetch is a helper that unifies the logic for fetching and starting a session.
func handleTimetrialFetch(
	w http.ResponseWriter,
	r *http.Request,
	fetchFn func(ctx context.Context, uid uuid.UUID) (*timetrial.Session, string, error),
) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	session, targetCountry, err := fetchFn(r.Context(), userID)
	if err != nil {
		timetrialError(w, err)
		return
	}

	resp := timetrialStatusResponse{
		Status:        string(session.Status),
		TargetCountry: targetCountry,
		StartTime:     session.StartTime,
		DurationMs:    session.Duration.Milliseconds(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// getStatus handles the GET /api/game/timetrial/status endpoint.
func (h *TimetrialHandler) getStatus(w http.ResponseWriter, r *http.Request) {
	handleTimetrialFetch(w, r, h.svc.GetCurrentStatus)
}

// startGame handles the POST /api/game/timetrial/start endpoint.
func (h *TimetrialHandler) startGame(w http.ResponseWriter, r *http.Request) {
	handleTimetrialFetch(w, r, h.svc.StartGame)
}

// postAttempt handles the POST /api/game/timetrial/attempt endpoint.
func (h *TimetrialHandler) postAttempt(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	var req timetrialAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.ProcessAttempt(r.Context(), userID, req.Guess)
	if err != nil {
		timetrialError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// getLeaderboard handles the GET /api/game/timetrial/leaderboard endpoint.
func (h *TimetrialHandler) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDContextKey).(uuid.UUID)
	if userID == uuid.Nil {
		http.Error(w, "server error: missing user context", http.StatusInternalServerError)
		return
	}

	leaderboard, err := h.svc.GetLeaderboard(r.Context(), userID)
	if err != nil {
		timetrialError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leaderboard)
}

// timetrialError maps time trial domain errors to appropriate HTTP responses.
func timetrialError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, timetrial.ErrChallengeNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, timetrial.ErrSessionNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, timetrial.ErrAlreadyCompleted):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
