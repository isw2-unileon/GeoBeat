package server_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

type timetrialSessionKey struct {
	userID      uuid.UUID
	challengeID int
}

type mockTimetrialSessionRepo struct {
	mu       sync.RWMutex
	sessions map[timetrialSessionKey]*timetrial.Session
}

func newMockTimetrialSessionRepo() *mockTimetrialSessionRepo {
	return &mockTimetrialSessionRepo{
		sessions: make(map[timetrialSessionKey]*timetrial.Session),
	}
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

func (m *mockTimetrialSessionRepo) CreateSession(ctx context.Context, session *timetrial.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := timetrialSessionKey{userID: session.UserID, challengeID: session.ChallengeID}
	m.sessions[key] = session
	return nil
}

func (m *mockTimetrialSessionRepo) UpdateSession(ctx context.Context, session *timetrial.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := timetrialSessionKey{userID: session.UserID, challengeID: session.ChallengeID}
	m.sessions[key] = session
	return nil
}

type mockTimetrialChallengeRepo struct {
	challenge *timetrial.Challenge
}

func newMockTimetrialChallengeRepo() *mockTimetrialChallengeRepo {
	return &mockTimetrialChallengeRepo{
		challenge: &timetrial.Challenge{
			ID:              1,
			TargetCountries: []string{"Spain", "France"},
			TargetGenres:    []string{"Pop", "Rock"},
			Date:            time.Now(),
		},
	}
}

func (m *mockTimetrialChallengeRepo) GetChallengeByDate(ctx context.Context, date time.Time) (*timetrial.Challenge, error) {
	if m.challenge == nil {
		return nil, timetrial.ErrChallengeNotFound
	}
	return m.challenge, nil
}

type mockTimetrialLeaderboardRepo struct {
	leaderboard *timetrial.Leaderboard
}

func newMockTimetrialLeaderboardRepo() *mockTimetrialLeaderboardRepo {
	return &mockTimetrialLeaderboardRepo{
		leaderboard: &timetrial.Leaderboard{
			ChallengeID: 1,
			Entries: []timetrial.LeaderboardEntry{
				{Rank: 1, UserName: "Player1", Duration: 15 * time.Second},
			},
		},
	}
}

func (m *mockTimetrialLeaderboardRepo) GetLeaderboard(ctx context.Context, challengeID int, userID uuid.UUID) (*timetrial.Leaderboard, error) {
	return m.leaderboard, nil
}

// newTimetrialTestServer wires up the real service with the in-memory fake repositories.
func newTimetrialTestServer(t *testing.T, loggedInUser uuid.UUID) (*http.ServeMux, *mockTimetrialSessionRepo) {
	t.Helper()

	sessionRepo := newMockTimetrialSessionRepo()
	challengeRepo := newMockTimetrialChallengeRepo()
	leaderboardRepo := newMockTimetrialLeaderboardRepo()

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewTimetrialService(challengeRepo, sessionRepo, leaderboardRepo, silentLogger)
	handler := server.NewTimetrialHandler(svc)

	mux := http.NewServeMux()
	mockAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), server.UserIDContextKey, loggedInUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	handler.RegisterRoutes(mux, mockAuthMiddleware)

	return mux, sessionRepo 
}

func TestTimetrialHandler_GetStatus(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		seedSession    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "1. Fails if session doesn't exist (must hit /start first)",
			seedSession:    false, // Empty database
			expectedStatus: http.StatusNotFound,
			expectedBody:   `no session available for this user and challenge`,
		},
		{
			name:           "2. Retrieves existing session correctly",
			seedSession:    true,
			expectedStatus: http.StatusOK,
			expectedBody:   `"target_country":"Spain"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, sessionRepo := newTimetrialTestServer(t, userID)

			if tt.seedSession {
				sessionRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
					UserID:       userID,
					ChallengeID:  1,
					CurrentIndex: 0,
					Status:       timetrial.StatusPlaying,
					StartTime:    time.Now().UTC(),
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/api/game/timetrial/status", nil)
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

func TestTimetrialHandler_StartGame(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		seedSession    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "1. Creates new session if none exists",
			seedSession:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"playing"`,
		},
		{
			name:           "2. Resumes existing session if already started",
			seedSession:    true,
			expectedStatus: http.StatusOK,
			expectedBody:   `"target_country":"France"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, sessionRepo := newTimetrialTestServer(t, userID)

			if tt.seedSession {
				sessionRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
					UserID:       userID,
					ChallengeID:  1,
					CurrentIndex: 1,
					Status:       timetrial.StatusPlaying,
					StartTime:    time.Now().UTC(),
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/start", nil)
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

func TestTimetrialHandler_PostAttempt(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		requestBody    string
		seedStatus     timetrial.GameStatus
		seedIndex      int
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid correct guess returns 200 OK and next country",
			requestBody:    `{"guess":"Pop"}`,
			seedStatus:     timetrial.StatusPlaying,
			seedIndex:      0,
			expectedStatus: http.StatusOK,
			expectedBody:   `"correct":true`,
		},
		{
			name:           "Invalid correct guess returns 200 OK but correct is false",
			requestBody:    `{"guess":"Jazz"}`,
			seedStatus:     timetrial.StatusPlaying,
			seedIndex:      0,
			expectedStatus: http.StatusOK,
			expectedBody:   `"correct":false`,
		},
		{
			name:           "Invalid JSON format returns 400 Bad Request",
			requestBody:    `{"guess":"Pop"`,
			seedStatus:     timetrial.StatusPlaying,
			seedIndex:      0,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `invalid request body`,
		},
		{
			name:           "Guessing on an already completed game returns 409 Conflict",
			requestBody:    `{"guess":"Pop"}`,
			seedStatus:     timetrial.StatusCompleted,
			seedIndex:      2,
			expectedStatus: http.StatusConflict,
			expectedBody:   `time trial challenge already completed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, sessionRepo := newTimetrialTestServer(t, userID)

			sessionRepo.sessions[timetrialSessionKey{userID: userID, challengeID: 1}] = &timetrial.Session{
				UserID:       userID,
				ChallengeID:  1,
				CurrentIndex: tt.seedIndex,
				Status:       tt.seedStatus,
				StartTime:    time.Now().UTC(),
			}

			req := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/attempt", bytes.NewBufferString(tt.requestBody))
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d. Body: %s", rec.Code, tt.expectedStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestTimetrialHandler_PlayFullGameFlow(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	mux, _ := newTimetrialTestServer(t, userID)

	req1 := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/start", nil)
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("req1 (Start) failed: status = %d", rec1.Code)
	}
	if !strings.Contains(rec1.Body.String(), `"target_country":"Spain"`) {
		t.Fatalf("Expected Spain to be first, got: %s", rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/attempt", bytes.NewBufferString(`{"guess":"Jazz"}`))
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if !strings.Contains(rec2.Body.String(), `"correct":false`) {
		t.Fatalf("expected guess to be incorrect, got body: %s", rec2.Body.String())
	}

	req3 := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/attempt", bytes.NewBufferString(`{"guess":"Pop"}`))
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)

	if !strings.Contains(rec3.Body.String(), `"next_country":"France"`) {
		t.Fatalf("expected to advance to France, got body: %s", rec3.Body.String())
	}

	req4 := httptest.NewRequest(http.MethodPost, "/api/game/timetrial/attempt", bytes.NewBufferString(`{"guess":"Rock"}`))
	rec4 := httptest.NewRecorder()
	mux.ServeHTTP(rec4, req4)

	if !strings.Contains(rec4.Body.String(), `"status":"completed"`) {
		t.Errorf("expected game to be completed, got body: %s", rec4.Body.String())
	}

	req5 := httptest.NewRequest(http.MethodGet, "/api/game/timetrial/leaderboard", nil)
	rec5 := httptest.NewRecorder()
	mux.ServeHTTP(rec5, req5)

	if rec5.Code != http.StatusOK {
		t.Fatalf("expected leaderboard status 200, got: %d", rec5.Code)
	}
	if !strings.Contains(rec5.Body.String(), `"Player1"`) {
		t.Errorf("expected leaderboard to contain Player1, got body: %s", rec5.Body.String())
	}
}
