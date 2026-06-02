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

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				// Assert initialized state
				if got.UserID != tt.userID {
					t.Errorf("Expected UserID %v, got %v", tt.userID, got.UserID)
				}
				if got.ChallengeID != tt.challengeID {
					t.Errorf("Expected ChallengeID %v, got %v", tt.challengeID, got.ChallengeID)
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
	baseStartTime := time.Now().UTC().Add(-1 * time.Minute) // Simulate game started 1 min ago

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
				CurrentIndex: 1, // On France
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
				CurrentIndex: 2, // On Japan (Last country)
				Status:       timetrial.StatusPlaying,
				StartTime:    baseStartTime,
			},
			guess:         "J-Pop",
			wantCorrect:   true,
			wantStatus:    timetrial.StatusCompleted,
			wantNext:      "", // No next country
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
				CurrentIndex: 5, // Invalid index manually set
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

			// Assert Error
			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatalf("Expected error %v, got nil", tt.wantErr)
				}
				if gotErr.Error() != tt.wantErr.Error() && !errors.Is(gotErr, tt.wantErr) {
					t.Fatalf("Expected error %v, got %v", tt.wantErr, gotErr)
				}
				return // If we expect an error, we don't evaluate the AttemptResult
			}

			if gotErr != nil {
				t.Fatalf("Unexpected error: %v", gotErr)
			}

			// Assert DTO Result
			if gotResult.Correct != tt.wantCorrect {
				t.Errorf("Expected Correct %v, got %v", tt.wantCorrect, gotResult.Correct)
			}
			if gotResult.Status != tt.wantStatus {
				t.Errorf("Expected Status %v, got %v", tt.wantStatus, gotResult.Status)
			}
			if gotResult.NextCountry != tt.wantNext {
				t.Errorf("Expected NextCountry %q, got %q", tt.wantNext, gotResult.NextCountry)
			}

			// Assert Internal Session State Mutation
			if tt.initialState.Status != tt.wantStatus {
				t.Errorf("Expected Session.Status to mutate to %v, got %v", tt.wantStatus, tt.initialState.Status)
			}

			if tt.checkDuration {
				if tt.initialState.EndTime.IsZero() {
					t.Error("Expected Session.EndTime to be set, got zero time")
				}
				if tt.initialState.Duration <= 0 {
					t.Errorf("Expected Session.Duration to be calculated (>0), got %v", tt.initialState.Duration)
				}
				if gotResult.Duration != tt.initialState.Duration {
					t.Errorf("Expected DTO duration %v to match Session duration %v", gotResult.Duration, tt.initialState.Duration)
				}
			}
		})
	}
}
