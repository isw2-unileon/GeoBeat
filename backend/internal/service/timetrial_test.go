package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

type timetrialSessionKey struct {
	userID      uuid.UUID
	challengeID int
}

type mockTimetrialSessionRepo struct {
	mu           sync.RWMutex
	sessions     map[timetrialSessionKey]*timetrial.Session
	createCalled bool
	updateCalled bool
}

func newMockTimetrialSessionRepo() *mockTimetrialSessionRepo {
	return &mockTimetrialSessionRepo{
		sessions: make(map[timetrialSessionKey]*timetrial.Session),
	}
}

func (m *mockTimetrialSessionRepo) CreateSession(ctx context.Context, s *timetrial.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createCalled = true
	key := timetrialSessionKey{userID: s.UserID, challengeID: s.ChallengeID}
	m.sessions[key] = s
	return nil
}

func (m *mockTimetrialSessionRepo) GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*timetrial.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := timetrialSessionKey{userID: userID, challengeID: challengeID}
	session, exists := m.sessions[key]
	if !exists {
		return nil, timetrial.ErrSessionNotFound
	}
	return session, nil
}

func (m *mockTimetrialSessionRepo) UpdateSession(ctx context.Context, s *timetrial.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updateCalled = true
	key := timetrialSessionKey{userID: s.UserID, challengeID: s.ChallengeID}
	m.sessions[key] = s
	return nil
}

type mockTimetrialChallengeRepo struct {
	challenge *timetrial.Challenge
	err       error
}

func newMockTimetrialChallengeRepo(c *timetrial.Challenge) *mockTimetrialChallengeRepo {
	return &mockTimetrialChallengeRepo{challenge: c}
}

func (m *mockTimetrialChallengeRepo) GetChallengeByDate(ctx context.Context, date time.Time) (*timetrial.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.challenge, nil
}

type mockTimetrialLeaderboardRepo struct {
	leaderboard *timetrial.Leaderboard
}

func newMockTimetrialLeaderboardRepo(l *timetrial.Leaderboard) *mockTimetrialLeaderboardRepo {
	return &mockTimetrialLeaderboardRepo{leaderboard: l}
}

func (m *mockTimetrialLeaderboardRepo) GetLeaderboard(ctx context.Context, challengeID int, userID uuid.UUID) (*timetrial.Leaderboard, error) {
	return m.leaderboard, nil
}

func setupTestService(cRepo service.TimetrialChallengeRepository, sRepo service.TimetrialSessionRepository, lRepo service.TimetrialLeaderboardRepository) *service.TimetrialService {
	return service.NewTimetrialService(cRepo, sRepo, lRepo)
}

func TestGetCurrentStatus(t *testing.T) {
	userID := uuid.New()
	baseChallenge := &timetrial.Challenge{ID: 1, TargetCountries: []string{"Spain", "France"}}

	t.Run("Challenge Not Found", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(nil)
		cRepo.err = timetrial.ErrChallengeNotFound
		svc := setupTestService(cRepo, nil, nil)

		_, _, err := svc.GetCurrentStatus(context.Background(), userID)
		if !errors.Is(err, timetrial.ErrChallengeNotFound) {
			t.Errorf("Expected ErrChallengeNotFound, got %v", err)
		}
	})

	t.Run("Active Session Returns Next Country", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo()

		sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
			UserID:       userID,
			ChallengeID:  1,
			Status:       timetrial.StatusPlaying,
			CurrentIndex: 0,
		}
		svc := setupTestService(cRepo, sRepo, nil)

		session, country, err := svc.GetCurrentStatus(context.Background(), userID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if session == nil {
			t.Fatal("Expected session, got nil")
		}
		if country != "Spain" {
			t.Errorf("Expected country 'Spain', got '%s'", country)
		}
	})

	t.Run("Completed Session Returns Empty Country", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo()

		sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
			UserID:       userID,
			ChallengeID:  1,
			Status:       timetrial.StatusCompleted,
			CurrentIndex: 2,
		}
		svc := setupTestService(cRepo, sRepo, nil)

		_, country, err := svc.GetCurrentStatus(context.Background(), userID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if country != "" {
			t.Errorf("Expected empty country for completed session, got '%s'", country)
		}
	})
}

func TestStartGame(t *testing.T) {
	userID := uuid.New()
	baseChallenge := &timetrial.Challenge{ID: 1, TargetCountries: []string{"Spain"}}

	t.Run("Session Already Exists (Resume)", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo()

		sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
			UserID:       userID,
			ChallengeID:  1,
			Status:       timetrial.StatusPlaying,
			CurrentIndex: 0,
		}
		svc := setupTestService(cRepo, sRepo, nil)

		_, country, err := svc.StartGame(context.Background(), userID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if country != "Spain" {
			t.Errorf("Expected to resume and get 'Spain', got '%s'", country)
		}
	})

	t.Run("Create New Session", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo() // Empty database
		svc := setupTestService(cRepo, sRepo, nil)

		session, country, err := svc.StartGame(context.Background(), userID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !sRepo.createCalled {
			t.Error("Expected CreateSession to be called in the repository")
		}
		if session == nil || country != "Spain" {
			t.Errorf("Invalid return state for new session")
		}

		if _, exists := sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}]; !exists {
			t.Error("Expected session to be stored in the fake database")
		}
	})
}

func TestProcessAttempt(t *testing.T) {
	userID := uuid.New()
	baseChallenge := &timetrial.Challenge{
		ID:              1,
		TargetCountries: []string{"Spain", "France"},
		TargetGenres:    []string{"Pop", "Rock"},
	}

	t.Run("Correct Guess Triggers Update", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo()

		sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
			UserID:       userID,
			ChallengeID:  1,
			CurrentIndex: 0,
			Status:       timetrial.StatusPlaying,
			StartTime:    time.Now(),
		}

		svc := setupTestService(cRepo, sRepo, nil)

		result, err := svc.ProcessAttempt(context.Background(), userID, "Pop") // Correct guess
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Correct {
			t.Error("Expected guess to be correct")
		}
		if !sRepo.updateCalled {
			t.Error("Expected UpdateSession to be called on a correct guess")
		}
	})

	t.Run("Incorrect Guess Avoids Update", func(t *testing.T) {
		cRepo := newMockTimetrialChallengeRepo(baseChallenge)
		sRepo := newMockTimetrialSessionRepo()

		sRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
			UserID:       userID,
			ChallengeID:  1,
			CurrentIndex: 0,
			Status:       timetrial.StatusPlaying,
			StartTime:    time.Now(),
		}

		svc := setupTestService(cRepo, sRepo, nil)

		result, err := svc.ProcessAttempt(context.Background(), userID, "Jazz") // Incorrect guess
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Correct {
			t.Error("Expected guess to be incorrect")
		}
		if sRepo.updateCalled {
			t.Error("Expected UpdateSession NOT to be called on an incorrect guess to save I/O")
		}
	})
}

func TestGetLeaderboard(t *testing.T) {
	userID := uuid.New()

	cRepo := newMockTimetrialChallengeRepo(&timetrial.Challenge{ID: 1})
	lRepo := newMockTimetrialLeaderboardRepo(&timetrial.Leaderboard{ChallengeID: 1})
	svc := setupTestService(cRepo, nil, lRepo)

	lb, err := svc.GetLeaderboard(context.Background(), userID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if lb == nil || lb.ChallengeID != 1 {
		t.Error("Failed to retrieve correct leaderboard state")
	}
}
