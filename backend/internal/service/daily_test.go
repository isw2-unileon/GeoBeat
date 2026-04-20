package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

type mockDailyRepo struct {
	mu        sync.RWMutex
	sessions  map[sessionKey]*daily.Session
	challenge *daily.Challenge

	getChallengeErr  error
	getSessionErr    error
	createSessionErr error
	updateSessionErr error
}

type sessionKey struct {
	userID      int
	challengeID int
}

func newMockRepo() *mockDailyRepo {
	return &mockDailyRepo{
		sessions: make(map[sessionKey]*daily.Session),
		challenge: &daily.Challenge{
			ID:          1,
			TargetGenre: "Pop",
			HintSongs:   []string{"Song 1", "Song 2", "Song 3", "Song 4", "Song 5"},
		},
	}
}

func (m *mockDailyRepo) GetChallengeByDate(ctx context.Context, date string) (*daily.Challenge, error) {
	if m.getChallengeErr != nil {
		return nil, m.getChallengeErr
	}
	return m.challenge, nil
}

func (m *mockDailyRepo) GetSession(ctx context.Context, userID, challengeID int) (*daily.Session, error) {
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

func (m *mockDailyRepo) CreateSession(ctx context.Context, session *daily.Session) error {
	if m.createSessionErr != nil {
		return m.createSessionErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sessionKey{session.UserID, session.ChallengeID}
	m.sessions[key] = session
	return nil
}

func (m *mockDailyRepo) UpdateSession(ctx context.Context, session *daily.Session) error {
	if m.updateSessionErr != nil {
		return m.updateSessionErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sessionKey{session.UserID, session.ChallengeID}
	m.sessions[key] = session
	return nil
}

func TestDaily_GetCurrentStatus(t *testing.T) {
	userID := 1
	tests := []struct {
		name        string
		setupRepo   func(*mockDailyRepo)
		wantSession bool
		wantErr     error
	}{
		{
			name: "Fail to get challenge returns error",
			setupRepo: func(m *mockDailyRepo) {
				m.getChallengeErr = errors.New("DB connection failed")
			},
			wantSession: false,
			wantErr:     daily.ErrChallengeNotFound,
		},
		{
			name: "Existing session is retrieved successfully",
			setupRepo: func(m *mockDailyRepo) {
				m.sessions[sessionKey{userID, 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 2,
					Status:       daily.StatusPlaying,
				}
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "If session does not exist, it is created and saved automatically",
			setupRepo: func(m *mockDailyRepo) {
				// No session setup, should trigger creation
			},
			wantSession: true,
			wantErr:     nil,
		},
		{
			name: "Error while creating session is handled properly",
			setupRepo: func(m *mockDailyRepo) {
				m.createSessionErr = errors.New("DB insert failed")
			},
			wantSession: false,
			wantErr:     errors.New("error while creating session"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			tt.setupRepo(repo)
			svc := service.NewService(repo)

			_, session, err := svc.GetCurrentStatus(context.Background(), userID)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error()) {
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("GetCurrentStatus() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if tt.wantSession && session == nil {
				t.Error("GetCurrentStatus() expected a session but got nil")
			}

			if tt.wantSession {
				key := sessionKey{userID, 1}
				if _, exists := repo.sessions[key]; !exists {
					t.Error("GetCurrentStatus() did not persist the session in the database")
				}
			}
		})
	}
}

func TestDaily_ProcessAttempt(t *testing.T) {
	userID := 1

	tests := []struct {
		name      string
		guess     string
		setupRepo func(*mockDailyRepo)
		wantErr   error
	}{
		{
			name:  "1. Successful attempt updates session correctly",
			guess: "Pop",
			setupRepo: func(m *mockDailyRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 0,
					Status:       daily.StatusPlaying,
				}
			},
			wantErr: nil,
		},
		{
			name:  "2. Domain error if game is already over",
			guess: "Rock",
			setupRepo: func(m *mockDailyRepo) {
				m.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 5,
					Status:       daily.StatusLost,
				}
			},
			wantErr: daily.ErrGameOver,
		},
		{
			name:  "3. Error while updating session is handled properly",
			guess: "Pop",
			setupRepo: func(m *mockDailyRepo) {
				m.updateSessionErr = errors.New("db update failed")
			},
			wantErr: errors.New("error updating session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			tt.setupRepo(repo)
			svc := service.NewService(repo)

			_, err := svc.ProcessAttempt(context.Background(), userID, tt.guess)

			if (err != nil && tt.wantErr == nil) || (err == nil && tt.wantErr != nil) {
				t.Errorf("ProcessAttempt() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && tt.wantErr != nil && !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
				t.Errorf("ProcessAttempt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
