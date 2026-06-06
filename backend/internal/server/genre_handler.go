package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
)

// GenreService defines the interface for fetching genres.
type GenreService interface {
	GetAllGenres(ctx context.Context) ([]genre.Genre, error)
}

// GenreHandler handles HTTP requests related to genres.
type GenreHandler struct {
	svc GenreService
}

// NewGenreHandler creates a new GenreHandler with the given GenreService.
func NewGenreHandler(svc GenreService) *GenreHandler {
	return &GenreHandler{svc: svc}
}

// RegisterRoutes registers the HTTP routes for genre-related endpoints.
func (h *GenreHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/genres", http.HandlerFunc(h.getAllGenres))
}

// getAllGenres handles the GET /api/genres endpoint to retrieve all allowed genres.
func (h *GenreHandler) getAllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.svc.GetAllGenres(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch genres", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(genres); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
