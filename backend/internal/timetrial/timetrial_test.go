package timetrial_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

func TestNewSession(t *testing.T) {
	validUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name        string
		userID      uuid.UUID
		challengeID int
		wantErr     error
	}{
		{
			name:        "Valid session creation",
			userID:      validUserID,
			challengeID: 1,
			wantErr:     nil,
		},
		{
			name:        "Invalid user ID (Nil UUID)",
			userID:      uuid.Nil,
			challengeID: 1,
			wantErr:     timetrial.ErrInvalidID,
		},
		{
			name:        "Invalid challenge ID (Zero)",
			userID:      validUserID,
			challengeID: 0,
			wantErr:     timetrial.ErrInvalidID,
		},
		{
			name:        "Invalid challenge ID (Negative)",
			userID:      validUserID,
			challengeID: -5,
			wantErr:     timetrial.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timetrial.NewSession(tt.userID, tt.challengeID)

			assertError(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assertInitialSession(t, got, tt.userID, tt.challengeID)
			}
		})
	}
}

func TestSession_MakeAttempt(t *testing.T) {
	mockChallenge := &timetrial.Challenge{
		ID:              1,
		TargetCountries: []string{"Spain", "France", "Japan"},
		TargetGenres:    []string{"Flamenco", "Pop", "J-Pop"},
		Date:            time.Now(),
	}

	validUserID := uuid.New()
	baseStartTime := time.Now().UTC().Add(-1 * time.Minute)

	tests := []struct {
		name          string
		initialState  *timetrial.Session
		guess         string
		wantCorrect   bool
		wantStatus    timetrial.GameStatus
		wantNext      string
		wantErr       error
		checkDuration bool
	}{
		{
			name: "1. Incorrect guess does not advance index",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 0,
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:       "Rock",
			wantCorrect: false,
			wantStatus:  timetrial.StatusPlaying,
			wantNext:    "Spain",
			wantErr:     nil,
		},
		{
			name: "2. Correct guess advances index and returns next country",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 0,
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:       "Flamenco",
			wantCorrect: true,
			wantStatus:  timetrial.StatusPlaying,
			wantNext:    "France",
			wantErr:     nil,
		},
		{
			name: "3. Correct guess is case-insensitive and trims spaces",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 1,
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:       "   pOp   ",
			wantCorrect: true,
			wantStatus:  timetrial.StatusPlaying,
			wantNext:    "Japan",
			wantErr:     nil,
		},
		{
			name: "4. Correct guess on last country completes game and computes duration",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 2,
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:         "J-Pop",
			wantCorrect:   true,
			wantStatus:    timetrial.StatusCompleted,
			wantNext:      "",
			wantErr:       nil,
			checkDuration: true,
		},
		{
			name: "5. Attempting to play a completed game returns error",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 3,
				Status:       timetrial.StatusCompleted,
				StartTime:    baseStartTime,
				EndTime:      time.Now().UTC(),
			},
			guess:       "Pop",
			wantCorrect: false,
			wantErr:     timetrial.ErrAlreadyCompleted,
		},
		{
			name: "6. Critical state error: index out of bounds",
			initialState: &timetrial.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				CurrentIndex: 5,
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:       "Pop",
			wantCorrect: false,
			wantErr:     errors.New("critical state error: index out of bounds"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := tt.initialState.MakeAttempt(tt.guess, mockChallenge)

			assertError(t, gotErr, tt.wantErr)

			// If we expected an error and got it, there is no result to evaluate.
			if tt.wantErr != nil {
				return
			}

			assertMakeAttemptResult(t, gotResult, tt.wantCorrect, tt.wantStatus, tt.wantNext)
			assertSessionMutation(t, tt.initialState, tt.wantStatus)

			if tt.checkDuration {
				assertDuration(t, tt.initialState, gotResult)
			}
		})
	}
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("Expected no error, got %v", got)
		}
		return
	}
	if got == nil {
		t.Fatalf("Expected error %v, got nil", want)
	}
	if got.Error() != want.Error() && !errors.Is(got, want) {
		t.Fatalf("Expected error %v, got %v", want, got)
	}
}

func assertInitialSession(t *testing.T, got *timetrial.Session, expectedUser uuid.UUID, expectedChallenge int) {
	t.Helper()
	if got.UserID != expectedUser {
		t.Errorf("Expected UserID %v, got %v", expectedUser, got.UserID)
	}
	if got.ChallengeID != expectedChallenge {
		t.Errorf("Expected ChallengeID %v, got %v", expectedChallenge, got.ChallengeID)
	}
	if got.CurrentIndex != 0 {
		t.Errorf("Expected initial CurrentIndex to be 0, got %d", got.CurrentIndex)
	}
	if got.Status != timetrial.StatusPlaying {
		t.Errorf("Expected initial Status to be playing, got %v", got.Status)
	}
	if got.StartTime.IsZero() {
		t.Error("Expected StartTime to be initialized, but got zero time")
	}
}

func assertMakeAttemptResult(t *testing.T, got *timetrial.AttemptResult, wantCorrect bool, wantStatus timetrial.GameStatus, wantNext string) {
	t.Helper()
	if got.Correct != wantCorrect {
		t.Errorf("Expected Correct %v, got %v", wantCorrect, got.Correct)
	}
	if got.Status != wantStatus {
		t.Errorf("Expected Status %v, got %v", wantStatus, got.Status)
	}
	if got.NextCountry != wantNext {
		t.Errorf("Expected NextCountry %q, got %q", wantNext, got.NextCountry)
	}
}

func assertSessionMutation(t *testing.T, session *timetrial.Session, wantStatus timetrial.GameStatus) {
	t.Helper()
	if session.Status != wantStatus {
		t.Errorf("Expected Session.Status to mutate to %v, got %v", wantStatus, session.Status)
	}
}

func assertDuration(t *testing.T, session *timetrial.Session, result *timetrial.AttemptResult) {
	t.Helper()
	if session.EndTime.IsZero() {
		t.Error("Expected Session.EndTime to be set, got zero time")
	}
	if session.Duration <= 0 {
		t.Errorf("Expected Session.Duration to be calculated (>0), got %v", session.Duration)
	} else if result.Duration != session.Duration {
		t.Errorf("Expected DTO duration %v to match Session duration %v", result.Duration, session.Duration)
	}
}
