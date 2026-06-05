package pgdb_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTokenTestDB(t *testing.T) *pgxpool.Pool {
	testDBUrl := os.Getenv("TEST_DATABASE_URL")
	if testDBUrl == "" {
		t.Skip("Skipping integration test: TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDBUrl)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE refresh_tokens CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate refresh_tokens: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE refresh_tokens CASCADE;")
		pool.Close()
	})

	return pool
}

func TestPostgresTokenRepo_SaveFindDelete(t *testing.T) {
	pool := setupTokenTestDB(t)
	repo := pgdb.NewPostgresTokenRepo(pool)
	ctx := context.Background()

	uid := uuid.New()
	rt := authsession.NewRefreshToken(uid, "hash-test-token")
	// Ensure times are stable
	rt.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	rt.ExpiresAt = rt.CreatedAt.Add(24 * time.Hour)

	// Seed a user because refresh_tokens references users.id
	userRepo := pgdb.NewPostgresUserRepo(pool)
	pwd := "dummy_bcrypt_hash"
	user := &geouser.User{
		ID:           uid,
		Email:        "seed-token-user@example.com",
		UserName:     "seeduser",
		PasswordHash: &pwd,
		Provider:     "email",
		CreatedAt:    rt.CreatedAt,
		UpdatedAt:    rt.CreatedAt,
	}
	if err := userRepo.Save(ctx, user); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	// Save
	if err := repo.Save(ctx, rt); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// FindByTokenHash
	got, err := repo.FindByTokenHash(ctx, "hash-test-token")
	if err != nil {
		t.Fatalf("FindByTokenHash() failed: %v", err)
	}
	if got.UserID != uid {
		t.Errorf("expected user id %v, got %v", uid, got.UserID)
	}

	// FindByUserID
	got2, err := repo.FindByUserID(ctx, uid)
	if err != nil {
		t.Fatalf("FindByUserID() failed: %v", err)
	}
	if got2.TokenHash != "hash-test-token" {
		t.Errorf("expected token hash 'hash-test-token', got %s", got2.TokenHash)
	}

	// Delete
	if err := repo.Delete(ctx, "hash-test-token"); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	if _, err := repo.FindByTokenHash(ctx, "hash-test-token"); err == nil {
		t.Fatalf("expected token to be deleted, but found it")
	}
}
