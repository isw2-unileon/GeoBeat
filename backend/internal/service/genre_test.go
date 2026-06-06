package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
)

// mockGetGenreRepository is a mock implementation of the GetGenreRepository interface for testing.
type mockGetGenreRepository struct {
	mockGenres []genre.Genre
	mockError  error
}

func (m *mockGetGenreRepository) GetAllowedGenres(ctx context.Context) ([]genre.Genre, error) {
	return m.mockGenres, m.mockError
}

func TestGenreService_GetAllGenres(t *testing.T) {
	t.Run("returns genres successfully", func(t *testing.T) {
		expectedGenres := []genre.Genre{
			{ID: 1, Name: "Pop", NormalizedName: "pop"},
			{ID: 2, Name: "Rock", NormalizedName: "rock"},
		}
		repo := &mockGetGenreRepository{mockGenres: expectedGenres, mockError: nil}
		svc := NewGenreService(repo)

		got, err := svc.GetAllGenres(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, expectedGenres) {
			t.Errorf("expected %v, got %v", expectedGenres, got)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expectedErr := errors.New("db connection failed")
		repo := &mockGetGenreRepository{mockGenres: nil, mockError: expectedErr}
		svc := NewGenreService(repo)

		got, err := svc.GetAllGenres(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if got != nil {
			t.Errorf("expected nil genres on error, got %v", got)
		}
	})
}
