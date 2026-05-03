package server_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

type sessionKey struct {
	userID      uuid.UUID
	challengeID int
}

type mockSessionRepo struct {
	mu       sync.RWMutex
	sessions map[sessionKey]*daily.Session
}

type mockChallengeRepo struct {
	challenge *daily.Challenge
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[sessionKey]*daily.Session),
	}
}

func newMockChallengeRepo() *mockChallengeRepo {
	return &mockChallengeRepo{
		challenge: &daily.Challenge{
			ID:          1,
			TargetGenre: "Pop",
			HintSongs:   []string{"Song 1", "Song 2", "Song 3", "Song 4", "Song 5"},
		},
	}
}

func (m *mockChallengeRepo) GetChallengeByDate(ctx context.Context, date string) (*daily.Challenge, error) {
	return m.challenge, nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*daily.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := sessionKey{userID: userID, challengeID: challengeID}
	session, exists := m.sessions[key]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *daily.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := sessionKey{userID: session.UserID, challengeID: session.ChallengeID}
	m.sessions[key] = session
	return nil
}

func (m *mockSessionRepo) UpdateSession(ctx context.Context, session *daily.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := sessionKey{userID: session.UserID, challengeID: session.ChallengeID}
	m.sessions[key] = session
	return nil
}

// newTestServer wires up the real service with the in-memory fake repository.
func newTestServer(t *testing.T, loggedInUser uuid.UUID) (*http.ServeMux, *mockSessionRepo, *mockChallengeRepo) {
	t.Helper()

	sessionRepo := newMockSessionRepo()
	challengeRepo := newMockChallengeRepo()
	svc := service.NewService(challengeRepo, sessionRepo)
	handler := server.NewHandler(svc)

	mux := http.NewServeMux()
	mockAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// server.UserIDKey must be exported from your server package!
			ctx := context.WithValue(r.Context(), server.UserIDContextKey, loggedInUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Inject the mock middleware instead of the real one
	handler.RegisterRoutes(mux, mockAuthMiddleware)

	return mux, sessionRepo, challengeRepo
}

func TestHandler_GetDailyStatus(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		seedSession    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "1. Creates new session if none exists and returns 200 OK",
			seedSession:    false, // Empty database
			expectedStatus: http.StatusOK,
			expectedBody:   `"attempts_used":0`,
		},
		{
			name:           "2. Retrieves existing session correctly",
			seedSession:    true, // Database already has progress
			expectedStatus: http.StatusOK,
			expectedBody:   `"attempts_used":3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, sessionRepo, _ := newTestServer(t, userID)

			if tt.seedSession {
				sessionRepo.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
					UserID:       userID,
					ChallengeID:  1,
					AttemptsUsed: 3,
					Status:       daily.StatusPlaying,
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/api/game/daily", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			if !strings.Contains(rec.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestHandler_PostAttempt(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		requestBody    string
		seedStatus     daily.GameStatus // Status to inject into the database before the test
		seedAttempts   int              // Attempts already used
		expectedStatus int
	}{
		{
			name:           "Valid correct guess returns 200 OK",
			requestBody:    `{"guess":"Pop"}`,
			seedStatus:     daily.StatusPlaying,
			seedAttempts:   0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON format returns 400 Bad Request",
			requestBody:    `{"guess":"Pop"`, // Broken JSON
			seedStatus:     daily.StatusPlaying,
			seedAttempts:   0,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Guessing on an already won game returns 409 Conflict",
			requestBody:    `{"guess":"Rock"}`,
			seedStatus:     daily.StatusWon, // Game is already over
			seedAttempts:   1,
			expectedStatus: http.StatusConflict, // Matches daily.ErrGameOver
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, sessionRepo, _ := newTestServer(t, userID)

			// Seed the database with the exact state we want to test
			sessionRepo.sessions[sessionKey{userID: userID, challengeID: 1}] = &daily.Session{
				UserID:       userID,
				ChallengeID:  1,
				AttemptsUsed: tt.seedAttempts,
				Status:       tt.seedStatus,
			}

			req := httptest.NewRequest(http.MethodPost, "/api/game/daily/attempt", bytes.NewBufferString(tt.requestBody))
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d. Body: %s", rec.Code, tt.expectedStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_PlayFullGameFlow(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	mux, _, challengeRepo := newTestServer(t, userID)

	challengeRepo.challenge = &daily.Challenge{ID: 1, TargetGenre: "Pop", HintSongs: []string{"H1", "H2", "H3", "H4", "H5"}}

	// 1st Request: Wrong guess
	req1 := httptest.NewRequest(http.MethodPost, "/api/game/daily/attempt", bytes.NewBufferString(`{"guess":"Rock"}`))
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("req1 failed: status = %d", rec1.Code)
	}

	// 2nd Request: Get status to verify it saved the wrong guess
	req2 := httptest.NewRequest(http.MethodGet, "/api/game/daily", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if !strings.Contains(rec2.Body.String(), `"attempts_used":1`) {
		t.Fatalf("expected attempts to be 1, got body: %s", rec2.Body.String())
	}

	// 3rd Request: Correct guess
	req3 := httptest.NewRequest(http.MethodPost, "/api/game/daily/attempt", bytes.NewBufferString(`{"guess":"Pop"}`))
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)

	if !strings.Contains(rec3.Body.String(), `"status":"won"`) {
		t.Errorf("expected game to be won, got body: %s", rec3.Body.String())
	}
}
