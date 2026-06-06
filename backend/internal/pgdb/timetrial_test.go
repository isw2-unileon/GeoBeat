package pgdb_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/pgdb"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
	"github.com/jackc/pgx/v5/pgxpool"
)

// cleanTimetrialDB clears the timetrial-related tables to ensure a clean state before each test.
func cleanTimetrialDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
        TRUNCATE timetrial_sessions, timetrial_challenges, users CASCADE;
        ALTER SEQUENCE timetrial_challenges_id_seq RESTART WITH 1;
    `)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}
}

// insertTestUser inserts a fake user to satisfy referential integrity (Foreign Keys).
func insertTestUser(t *testing.T, pool *pgxpool.Pool, id uuid.UUID, username string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
        INSERT INTO users (id, email, user_name, provider, provider_id, created_at, updated_at) 
        VALUES ($1, $2, $3, 'google', 'google-dummy-123', NOW(), NOW())
    `, id, username+"@test.com", username)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
}

func TestPostgresTimetrialRepo_Challenge(t *testing.T) {
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
	defer cleanTimetrialDB(t, pool)
	repo := pgdb.NewPostgresTimetrialRepo(pool)

	today := time.Now().Truncate(24 * time.Hour)

	_, err = repo.GetChallengeByDate(ctx, today)
	if !errors.Is(err, timetrial.ErrChallengeNotFound) {
		t.Fatalf("Expected ErrChallengeNotFound, got %v", err)
	}

	c := &timetrial.Challenge{
		TargetCountries: []string{"Spain", "France"},
		TargetGenres:    []string{"Pop", "Rock"},
		Date:            today,
	}

	err = repo.SaveChallenge(ctx, c)
	if err != nil {
		t.Fatalf("Failed to save challenge: %v", err)
	}
	if c.ID == 0 {
		t.Error("Expected challenge ID to be generated and updated in struct")
	}

	retrieved, err := repo.GetChallengeByDate(ctx, today)
	if err != nil {
		t.Fatalf("Failed to retrieve challenge: %v", err)
	}
	if retrieved.ID != c.ID || len(retrieved.TargetCountries) != 2 {
		t.Errorf("Retrieved challenge data mismatch")
	}
}

func TestPostgresTimetrialRepo_Session(t *testing.T) {
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
	defer cleanTimetrialDB(t, pool)
	repo := pgdb.NewPostgresTimetrialRepo(pool)

	userID := uuid.New()
	insertTestUser(t, pool, userID, "player1")

	c := &timetrial.Challenge{
		TargetCountries: []string{"Spain"},
		TargetGenres:    []string{"Pop"},
		Date:            time.Now(),
	}
	if err := repo.SaveChallenge(ctx, c); err != nil {
		t.Fatalf("Failed setup: %v", err)
	}

	s := &timetrial.Session{
		UserID:       userID,
		ChallengeID:  c.ID,
		CurrentIndex: 0,
		StartTime:    time.Now().UTC(),
		Status:       timetrial.StatusPlaying,
	}

	if err := repo.CreateSession(ctx, s); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	retrieved, err := repo.GetSession(ctx, userID, c.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if retrieved.Status != timetrial.StatusPlaying {
		t.Errorf("Expected status playing, got %v", retrieved.Status)
	}
	if !retrieved.EndTime.IsZero() {
		t.Error("Expected EndTime to be zero for a playing session")
	}

	s.CurrentIndex = 1
	s.Status = timetrial.StatusCompleted
	s.EndTime = time.Now().UTC()
	s.Duration = 15000 * time.Millisecond // 15 seconds

	if err := repo.UpdateSession(ctx, s); err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	completed, _ := repo.GetSession(ctx, userID, c.ID)
	if completed.Duration != s.Duration {
		t.Errorf("Expected duration %v, got %v", s.Duration, completed.Duration)
	}
}

func TestPostgresTimetrialRepo_Leaderboard(t *testing.T) {
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
	defer cleanTimetrialDB(t, pool)
	repo := pgdb.NewPostgresTimetrialRepo(pool)

	c := &timetrial.Challenge{
		TargetCountries: []string{"Spain"},
		TargetGenres:    []string{"Pop"},
		Date:            time.Now(),
	}
	_ = repo.SaveChallenge(ctx, c)

	user1 := uuid.New() // Rank 3
	user2 := uuid.New() // Rank 1
	user3 := uuid.New() // Rank 2

	insertTestUser(t, pool, user1, "slow_player")
	insertTestUser(t, pool, user2, "fast_player")
	insertTestUser(t, pool, user3, "mid_player")

	baseTime := time.Now().UTC()

	_ = repo.CreateSession(ctx, &timetrial.Session{UserID: user1, ChallengeID: c.ID, Status: timetrial.StatusPlaying, StartTime: baseTime})
	_ = repo.UpdateSession(ctx, &timetrial.Session{UserID: user1, ChallengeID: c.ID, Status: timetrial.StatusCompleted, Duration: 50 * time.Second})

	_ = repo.CreateSession(ctx, &timetrial.Session{UserID: user2, ChallengeID: c.ID, Status: timetrial.StatusPlaying, StartTime: baseTime})
	_ = repo.UpdateSession(ctx, &timetrial.Session{UserID: user2, ChallengeID: c.ID, Status: timetrial.StatusCompleted, Duration: 10 * time.Second})

	_ = repo.CreateSession(ctx, &timetrial.Session{UserID: user3, ChallengeID: c.ID, Status: timetrial.StatusPlaying, StartTime: baseTime})
	_ = repo.UpdateSession(ctx, &timetrial.Session{UserID: user3, ChallengeID: c.ID, Status: timetrial.StatusCompleted, Duration: 30 * time.Second})

	leaderboard, err := repo.GetLeaderboard(ctx, c.ID, user1)
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	if len(leaderboard.Entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(leaderboard.Entries))
	}

	if leaderboard.Entries[0].UserName != "fast_player" || leaderboard.Entries[0].Rank != 1 {
		t.Errorf("Expected fast_player to be rank 1")
	}
	if leaderboard.Entries[1].UserName != "mid_player" || leaderboard.Entries[1].Rank != 2 {
		t.Errorf("Expected mid_player to be rank 2")
	}
	if leaderboard.Entries[2].UserName != "slow_player" || leaderboard.Entries[2].Rank != 3 {
		t.Errorf("Expected slow_player to be rank 3")
	}

	if leaderboard.UserEntry == nil {
		t.Fatal("Expected UserEntry to be populated")
	}
	if leaderboard.UserEntry.UserName != "slow_player" {
		t.Errorf("Expected UserEntry to match slow_player, got %v", leaderboard.UserEntry.UserName)
	}
}
