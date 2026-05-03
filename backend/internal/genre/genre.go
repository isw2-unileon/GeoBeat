package genre

import "context"

// Genre represents a music genre with its ID, name, and normalized name.
type Genre struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}

// Repository defines the methods required to manage genres.
type Repository interface {
	GetAllowedGenres(ctx context.Context) ([]Genre, error)
}
