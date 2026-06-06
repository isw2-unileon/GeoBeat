package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
)

// mockGenreService is a mock implementation of the GenreService interface for testing.
type mockGenreService struct {
	mockGenres []genre.Genre
	mockError  error
}

func (m *mockGenreService) GetAllGenres(ctx context.Context) ([]genre.Genre, error) {
	return m.mockGenres, m.mockError
}

func TestGenreHandler_getAllGenres(t *testing.T) {
	t.Run("returns 200 and valid json on success", func(t *testing.T) {
		expectedGenres := []genre.Genre{
			{ID: 1, Name: "Pop", NormalizedName: "pop"},
			{ID: 2, Name: "Rock", NormalizedName: "rock"},
		}
		svc := &mockGenreService{mockGenres: expectedGenres, mockError: nil}
		handler := NewGenreHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/genres", nil)
		rr := httptest.NewRecorder()

		handler.getAllGenres(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		var gotGenres []genre.Genre
		if err := json.NewDecoder(rr.Body).Decode(&gotGenres); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}

		if !reflect.DeepEqual(gotGenres, expectedGenres) {
			t.Errorf("expected body %v, got %v", expectedGenres, gotGenres)
		}
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockGenreService{mockGenres: nil, mockError: errors.New("service failure")}
		handler := NewGenreHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/genres", nil)
		rr := httptest.NewRecorder()

		handler.getAllGenres(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		expectedBody := "failed to fetch genres\n"
		if rr.Body.String() != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
		}
	})
}
