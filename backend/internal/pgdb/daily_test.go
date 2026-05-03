package pgdb_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupDailyTestDB(t *testing.T) *pgxpool.Pool {
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
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE users, daily_challenges, daily_sessions RESTART IDENTITY CASCADE;")
		pool.Close()
	})

	return pool
}

func TestPostgresDailyRepo_Challenges(t *testing.T) {
	pool := setupDailyTestDB(t)
	repo := pgdb.NewPostgresDailyRepo(pool)
	ctx := context.Background()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	newChallenge := &daily.Challenge{
		TargetCountry: "Spain",
		TargetGenre:   "Flamenco",
		HintSongs:     []string{"Song A", "Song B"},
		Date:          today,
	}

	t.Run("SaveDailyChallenge", func(t *testing.T) {
		err := repo.SaveDailyChallenge(ctx, newChallenge)
		if err != nil {
			t.Fatalf("Failed to save challenge: %v", err)
		}

		if newChallenge.ID == 0 {
			t.Errorf("Expected ID to be populated by RETURNING clause, got 0")
		}
	})

	t.Run("GetChallengeByDate_Success", func(t *testing.T) {
		fetched, err := repo.GetChallengeByDate(ctx, today)
		if err != nil {
			t.Fatalf("Failed to fetch challenge: %v", err)
		}

		if fetched.TargetCountry != newChallenge.TargetCountry {
			t.Errorf("Expected country %s, got %s", newChallenge.TargetCountry, fetched.TargetCountry)
		}

		if len(fetched.HintSongs) != 2 || fetched.HintSongs[0] != "Song A" {
			t.Errorf("Array mapping failed. Got: %v", fetched.HintSongs)
		}
	})

	t.Run("GetChallengeByDate_NotFound", func(t *testing.T) {
		tomorrow := today.AddDate(0, 0, 1)
		_, err := repo.GetChallengeByDate(ctx, tomorrow)

		if err == nil {
			t.Errorf("Expected error for non-existent challenge, got nil")
		}
	})
}

func TestPostgresDailyRepo_Sessions(t *testing.T) {
	pool := setupDailyTestDB(t)
	repo := pgdb.NewPostgresDailyRepo(pool)
	ctx := context.Background()

	userID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, user_name, provider, provider_id, created_at, updated_at) 
		VALUES ($1, 'test@daily.com', 'daily_tester', 'google', 'google-dummy-123', NOW(), NOW())
	`, userID)
	if err != nil {
		t.Fatalf("Failed to seed dummy user for FK constraint: %v", err)
	}

	challenge := &daily.Challenge{
		TargetCountry: "France",
		TargetGenre:   "Pop",
		HintSongs:     []string{"Song C"},
		Date:          time.Now().UTC(),
	}
	err = repo.SaveDailyChallenge(ctx, challenge)
	if err != nil {
		t.Fatalf("Failed to seed dummy challenge for FK constraint: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	session := &daily.Session{
		UserID:       userID,
		ChallengeID:  challenge.ID,
		AttemptsUsed: 1,
		Status:       daily.StatusPlaying,
		UpdatedAt:    now,
	}

	t.Run("CreateSession", func(t *testing.T) {
		err := repo.CreateSession(ctx, session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	})

	t.Run("GetSession_Success", func(t *testing.T) {
		fetched, err := repo.GetSession(ctx, userID, challenge.ID)
		if err != nil {
			t.Fatalf("Failed to fetch session: %v", err)
		}

		if fetched.AttemptsUsed != 1 {
			t.Errorf("Expected 1 attempt, got %d", fetched.AttemptsUsed)
		}
		if fetched.Status != daily.StatusPlaying {
			t.Errorf("Expected status playing, got %v", fetched.Status)
		}
	})

	t.Run("UpdateSession_Success", func(t *testing.T) {
		session.AttemptsUsed = 5
		session.Status = daily.StatusLost
		session.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

		err := repo.UpdateSession(ctx, session)
		if err != nil {
			t.Fatalf("Failed to update session: %v", err)
		}

		fetched, _ := repo.GetSession(ctx, userID, challenge.ID)
		if fetched.AttemptsUsed != 5 || fetched.Status != daily.StatusLost {
			t.Errorf("Update did not persist. Got attempts: %d, status: %v", fetched.AttemptsUsed, fetched.Status)
		}
	})

	t.Run("UpdateSession_NotFound", func(t *testing.T) {
		fakeSession := &daily.Session{
			UserID:       uuid.New(),
			ChallengeID:  challenge.ID,
			AttemptsUsed: 1,
		}

		err := repo.UpdateSession(ctx, fakeSession)
		if err == nil {
			t.Errorf("Expected error updating non-existent session, got nil")
		}
	})
}
