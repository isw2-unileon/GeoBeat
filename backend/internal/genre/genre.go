package genre

import "context"

type Genre struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}

type GenreRepository interface {
	GetAllowedGenres(ctx context.Context) ([]Genre, error)
}
