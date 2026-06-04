package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

type mockSessionRepo struct {
	mu       sync.RWMutex
	sessions map[sessionKey]*daily.Session

	getSessionErr    error
	createSessionErr error
	updateSessionErr error
}

type mockChallengeRepo struct {
	challenge        *daily.Challenge
	inverseChallenge *daily.InverseChallenge

	getChallengeErr error
}
type sessionKey struct {
	userID      uuid.UUID
	challengeID int
}

func newSessionMockRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[sessionKey]*daily.Session),
	}
}

func newChallengeMockRepo() *mockChallengeRepo {
	return &mockChallengeRepo{
		challenge: &daily.Challenge{
			ID:          1,
			TargetGenre: "Pop",
			HintSongs:   []string{"Song 1", "Song 2", "Song 3", "Song 4", "Song 5"},
		},
		inverseChallenge: &daily.InverseChallenge{
			ID:            2,
			TargetCountry: "spain",
		},
	}
}

func (m *mockChallengeRepo) GetChallengeByDate(ctx context.Context, date time.Time) (*daily.Challenge, error) {
	if m.getChallengeErr != nil {
		return nil, m.getChallengeErr
	}
	return m.challenge, nil
}

func (m *mockChallengeRepo) GetInverseChallengeByDate(ctx context.Context, date time.Time) (*daily.InverseChallenge, error) {
	if m.getChallengeErr != nil {
		return nil, m.getChallengeErr
	}
	return m.inverseChallenge, nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*daily.Session, error) {
	if m.getSessionErr != nil {
		return nil, m.getSessionErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := sessionKey{userID, challengeID}
	session, exists := m.sessions[key]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *daily.Session) error {
	if m.createSessionErr != nil {
		return m.createSessionErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sessionKey{session.UserID, session.ChallengeID}
	m.sessions[key] = session
	return nil
}

func (m *mockSessionRepo) UpdateSession(ctx context.Context, session *daily.Session) error {
	if m.updateSessionErr != nil {
		return m.updateSessionErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sessionKey{session.UserID, session.ChallengeID}
	m.sessions[key] = session
	return nil
}

type currentStatusTestCase struct {
	name               string
	setupSessionRepo   func(*mockSessionRepo)
	setupChallengeRepo func(*mockChallengeRepo)
	wantSession        bool
	wantErr            error
}

func TestDaily_GetCurrentStatus(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	tests := []currentStatusTestCase{
		{
			name: "Fail to get challenge returns error",
			setupSessionRepo: func(m *mockSessionRepo) {
				// No setup needed for session repo
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				m.getChallengeErr = errors.New("DB connection failed")
			},
			wantSession: false,
			wantErr:     daily.ErrChallengeNotFound,
		},
		{
			name: "Existing session is retrieved successfully",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID, 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 2,
					Status:       daily.StatusPlaying,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "If session does not exist, it is created and saved automatically",
			setupSessionRepo: func(m *mockSessionRepo) {
				// No session setup, should trigger creation
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "Error while creating session is handled properly",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.createSessionErr = errors.New("DB insert failed")
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: false,
			wantErr:     errors.New("error while creating session"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeRepo := newChallengeMockRepo()
			tt.setupChallengeRepo(challengeRepo)
			sessionRepo := newSessionMockRepo()
			tt.setupSessionRepo(sessionRepo)
			svc := service.NewService(challengeRepo, sessionRepo)

			_, session, err := svc.GetCurrentStatus(context.Background(), userID)

			assertError(t, err, tt.wantErr)

			if err != nil {
				return
			}

			assertSessionState(t, tt.wantSession, session, sessionRepo, userID)
		})
	}
}

func TestDaily_GetCurrentInverseStatus(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	tests := []currentStatusTestCase{
		{
			name: "Fail to get inverse challenge returns error",
			setupSessionRepo: func(m *mockSessionRepo) {
				// No setup needed for session repo
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				m.getChallengeErr = errors.New("DB connection failed")
			},
			wantSession: false,
			wantErr:     daily.ErrChallengeNotFound,
		},
		{
			name: "Existing session is retrieved successfully",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID, 2}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  2,
					AttemptsUsed: 1,
					Status:       daily.StatusPlaying,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "If session does not exist, it is created and saved automatically",
			setupSessionRepo: func(m *mockSessionRepo) {
				// No session setup, should trigger creation
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "Error while creating session is handled properly",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.createSessionErr = errors.New("DB insert failed")
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantSession: false,
			wantErr:     errors.New("error while creating session"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeRepo := newChallengeMockRepo()
			tt.setupChallengeRepo(challengeRepo)
			sessionRepo := newSessionMockRepo()
			tt.setupSessionRepo(sessionRepo)
			svc := service.NewService(challengeRepo, sessionRepo)

			_, session, err := svc.GetCurrentInverseStatus(context.Background(), userID)

			assertError(t, err, tt.wantErr)

			if err != nil {
				return
			}

			assertSessionState(t, tt.wantSession, session, sessionRepo, userID)
		})
	}
}

// assertError isolates error checking to keep cyclomatic complexity extremely low.
func assertError(t *testing.T, got, want error) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("unexpected error: %v", got)
		}
		return
	}

	if got == nil {
		t.Fatalf("expected error %q, but got nil", want)
	}

	if !errors.Is(got, want) && got.Error() != want.Error() {
		t.Fatalf("got error %q, want %q", got, want)
	}
}

// assertSessionState isolates state checking to keep cyclomatic complexity extremely low.
func assertSessionState(t *testing.T, wantSession bool, session *daily.Session, sessionRepo *mockSessionRepo, userID uuid.UUID) {
	t.Helper()

	if wantSession {
		if session == nil {
			t.Fatal("expected a session but got nil")
		}

		key := sessionKey{userID, session.ChallengeID}
		if _, exists := sessionRepo.sessions[key]; !exists {
			t.Error("did not persist the session in the database")
		}
	} else if session != nil {
		t.Fatal("expected no session, but got one")
	}
}

func TestDaily_ProcessAttempt(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name               string
		guess              string
		setupSessionRepo   func(*mockSessionRepo)
		setupChallengeRepo func(*mockChallengeRepo)
		wantErr            error
	}{
		{
			name:  "1. Successful attempt updates session correctly",
			guess: "Pop",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 0,
					Status:       daily.StatusPlaying,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: nil,
		},
		{
			name:  "2. Domain error if game is already over",
			guess: "Rock",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 5,
					Status:       daily.StatusLost,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: daily.ErrGameOver,
		},
		{
			name:  "3. Error while updating session is handled properly",
			guess: "Pop",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.updateSessionErr = errors.New("db update failed")
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: errors.New("error updating session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeRepo := newChallengeMockRepo()
			tt.setupChallengeRepo(challengeRepo)
			sessionRepo := newSessionMockRepo()
			tt.setupSessionRepo(sessionRepo)
			svc := service.NewService(challengeRepo, sessionRepo)

			_, err := svc.ProcessAttempt(context.Background(), userID, tt.guess)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) {
				t.Errorf("ProcessAttempt() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && tt.wantErr != nil && !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
				t.Errorf("ProcessAttempt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDaily_ProcessInverseAttempt(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name               string
		guess              string
		setupSessionRepo   func(*mockSessionRepo)
		setupChallengeRepo func(*mockChallengeRepo)
		wantErr            error
	}{
		{
			name:  "1. Successful attempt updates session correctly",
			guess: "spain",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 2}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  2,
					AttemptsUsed: 0,
					Status:       daily.StatusPlaying,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: nil,
		},
		{
			name:  "2. Domain error if game is already over",
			guess: "france",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 2}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  2,
					AttemptsUsed: 5,
					Status:       daily.StatusLost,
				}
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: daily.ErrGameOver,
		},
		{
			name:  "3. Error while updating session is handled properly",
			guess: "spain",
			setupSessionRepo: func(m *mockSessionRepo) {
				m.updateSessionErr = errors.New("db update failed")
			},
			setupChallengeRepo: func(m *mockChallengeRepo) {
				// No setup needed for challenge repo
			},
			wantErr: errors.New("error updating session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeRepo := newChallengeMockRepo()
			tt.setupChallengeRepo(challengeRepo)
			sessionRepo := newSessionMockRepo()
			tt.setupSessionRepo(sessionRepo)
			svc := service.NewService(challengeRepo, sessionRepo)

			_, err := svc.ProcessInverseAttempt(context.Background(), userID, tt.guess)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) {
				t.Errorf("ProcessInverseAttempt() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && tt.wantErr != nil && !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
				t.Errorf("ProcessInverseAttempt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
