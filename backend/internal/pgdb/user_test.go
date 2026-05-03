package pgdb_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Helper function to easily create pointers for tests
func ptr[T any](v T) *T {
	return &v
}

// Helper function to setup the DB and ensure it is cleaned before and after each test
func setupUserTestDB(t *testing.T) *pgxpool.Pool {
	testDBUrl := os.Getenv("TEST_DATABASE_URL")
	if testDBUrl == "" {
		t.Skip("Skipping integration test: TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDBUrl)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE users CASCADE;")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE users CASCADE;")
		pool.Close()
	})

	return pool
}

func TestPostgresUserRepo_SaveAndFindByEmail(t *testing.T) {
	pool := setupUserTestDB(t)
	repo := pgdb.NewPostgresUserRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	expectedUser := &geouser.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		UserName:     "testuser",
		PasswordHash: nil,
		Provider:     "google",
		ProviderID:   ptr("google-12345"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Test Save
	err := repo.Save(ctx, expectedUser)
	if err != nil {
		t.Fatalf("Save() failed unexpectedly: %v", err)
	}

	// Test FindByEmail (Success Case)
	gotUser, err := repo.FindByEmail(ctx, expectedUser.Email)
	if err != nil {
		t.Fatalf("FindByEmail() failed unexpectedly: %v", err)
	}

	// Validate Fields
	if gotUser.ID != expectedUser.ID {
		t.Errorf("Expected ID %v, got %v", expectedUser.ID, gotUser.ID)
	}
	if gotUser.Email != expectedUser.Email {
		t.Errorf("Expected Email %v, got %v", expectedUser.Email, gotUser.Email)
	}
	if gotUser.PasswordHash != nil {
		t.Errorf("Expected PasswordHash to be nil, got %v", *gotUser.PasswordHash)
	}
	if gotUser.ProviderID == nil || *gotUser.ProviderID != "google-12345" {
		t.Errorf("ProviderID mismatch")
	}
	if !gotUser.CreatedAt.Equal(expectedUser.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", expectedUser.CreatedAt, gotUser.CreatedAt)
	}

	// Test FindByEmail (Not Found Case)
	_, err = repo.FindByEmail(ctx, "nonexistent@example.com")
	if err == nil || err.Error() != "user not found" {
		t.Errorf("Expected 'user not found' error, got %v", err)
	}
}

func TestPostgresUserRepo_Update(t *testing.T) {
	pool := setupUserTestDB(t)
	repo := pgdb.NewPostgresUserRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	initialUser := &geouser.User{
		ID:           uuid.New(),
		Email:        "update@example.com",
		UserName:     "updateme",
		PasswordHash: ptr("dummy_bcrypt_hash_123"),
		Provider:     "email",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := repo.Save(ctx, initialUser)
	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	time.Sleep(1 * time.Millisecond)
	newUpdateTime := time.Now().UTC().Truncate(time.Microsecond)

	initialUser.Provider = "google"
	initialUser.ProviderID = ptr("google-999")

	initialUser.PasswordHash = nil
	initialUser.UpdatedAt = newUpdateTime

	err = repo.Update(ctx, initialUser)
	if err != nil {
		t.Fatalf("Update() failed unexpectedly: %v", err)
	}

	updatedUser, _ := repo.FindByEmail(ctx, initialUser.Email)

	if updatedUser.Provider != "google" {
		t.Errorf("Expected Provider 'google', got %v", updatedUser.Provider)
	}
	if updatedUser.ProviderID == nil || *updatedUser.ProviderID != "google-999" {
		t.Errorf("Expected ProviderID 'google-999', got mismatch")
	}

	// Test Update (Not Found Case)
	ghostUser := &geouser.User{
		ID:        uuid.New(),
		Provider:  "email",
		UpdatedAt: time.Now(),
	}

	err = repo.Update(ctx, ghostUser)
	if err == nil || err.Error() != "cannot update: user not found" {
		t.Errorf("Expected 'cannot update: user not found' error, got %v", err)
	}
}
