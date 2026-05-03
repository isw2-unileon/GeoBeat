package pgdb

import (
	"context"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresGenreRepo implements the genre.Repository interface.
type PostgresGenreRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresGenreRepo creates a new instance of the repository.
func NewPostgresGenreRepo(pool *pgxpool.Pool) *PostgresGenreRepo {
	return &PostgresGenreRepo{pool: pool}
}

// GetAllowedGenres retrieves the list of allowed genres from the database.
func (r *PostgresGenreRepo) GetAllowedGenres(ctx context.Context) ([]genre.Genre, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name, normalized_name FROM genres")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []genre.Genre
	for rows.Next() {
		var genre genre.Genre
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.NormalizedName); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return genres, nil
}
