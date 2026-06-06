package service

import (
	"context"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
)

// GetGenreRepository defines the interface for fetching genres from the data source.
type GetGenreRepository interface {
	GetAllowedGenres(ctx context.Context) ([]genre.Genre, error)
}

// GenreService provides methods to interact with genres.
type GenreService struct {
	genreRepository GetGenreRepository
}

// NewGenreService creates a new instance of GenreService with the provided repository.
func NewGenreService(gr GetGenreRepository) *GenreService {
	return &GenreService{
		genreRepository: gr,
	}
}

// GetAllGenres retrieves all allowed genres using the repository.
func (s *GenreService) GetAllGenres(ctx context.Context) ([]genre.Genre, error) {
	return s.genreRepository.GetAllowedGenres(ctx)
}
