package pgdb_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
)

func TestPostgresGenreRepo_GetAllowedGenres(t *testing.T) {
	testDBUrl := os.Getenv("TEST_DATABASE_URL")
	if testDBUrl == "" {
		t.Skip("Skipping integration test: TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDBUrl)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE genres RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE genres RESTART IDENTITY CASCADE;")
	})

	seedQuery := `
		INSERT INTO genres (name, normalized_name) 
		VALUES ('Pop Music', 'pop'), ('Hip Hop', 'hip-hop')
		RETURNING id;
	`
	_, err = pool.Exec(ctx, seedQuery)
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	repo := pgdb.NewPostgresGenreRepo(pool)
	got, err := repo.GetAllowedGenres(ctx)
	if err != nil {
		t.Fatalf("GetAllowedGenres() failed unexpectedly: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("Expected 2 genres, got %d", len(got))
	}

	expected := []genre.Genre{
		{ID: 1, Name: "Pop Music", NormalizedName: "pop"},
		{ID: 2, Name: "Hip Hop", NormalizedName: "hip-hop"},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetAllowedGenres() mismatch.\nGot:  %+v\nWant: %+v", got, expected)
	}
}
