package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

type DailyService interface {
	GetCurrentStatus(ctx context.Context, userID int) (*daily.Challenge, *daily.Session, error)
	ProcessAttempt(ctx context.Context, userID int, guess string) (*daily.AttemptResult, error)
}

type Handler struct {
	svc DailyService
}

func NewHandler(svc DailyService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/game/daily", h.getDailyStatus)
	mux.HandleFunc("POST /api/game/daily/attempt", h.postAttempt)
}

type attemptRequest struct {
	Guess string `json:"guess"`
}

type statusResponse struct {
	Country      string `json:"country"`
	AttemptsUsed int    `json:"attempts_used"`
	Status       string `json:"status"`
}

func (h *Handler) getDailyStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Obtain userID from middleware context
	// userID := r.Context().Value("userID").(int)
	userID := 1 // hardcoded for testing

	challenge, session, err := h.svc.GetCurrentStatus(r.Context(), userID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	resp := statusResponse{
		Country:      challenge.TargetCountry,
		AttemptsUsed: session.AttemptsUsed,
		Status:       string(session.Status),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) postAttempt(w http.ResponseWriter, r *http.Request) {
	// TODO: Obtain userID from middleware context
	userID := 1

	var req attemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.ProcessAttempt(r.Context(), userID, req.Guess)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func handleDomainError(w http.ResponseWriter, err error) {
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
