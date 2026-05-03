package daily_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

func TestNewSession(t *testing.T) {
	validUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name        string
		userID      uuid.UUID
		challengeID int
		want        *daily.Session
		wantErr     error
	}{
		{
			name:        "Valid session creation",
			userID:      validUserID,
			challengeID: 1,
			want: &daily.Session{
				UserID:       validUserID,
				ChallengeID:  1,
				AttemptsUsed: 0,
				Status:       daily.StatusPlaying,
			},
			wantErr: nil,
		},
		{
			name:        "Invalid user ID (Nil UUID)",
			userID:      uuid.Nil,
			challengeID: 1,
			want:        nil,
			wantErr:     daily.ErrInvalidID,
		},
		{
			name:        "Invalid challenge ID",
			userID:      validUserID,
			challengeID: -1,
			want:        nil,
			wantErr:     daily.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := daily.NewSession(tt.userID, tt.challengeID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				return
			}

			if *got != *tt.want {
				t.Errorf("NewSession() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSession_MakeAttempt(t *testing.T) {
	mockChallenge := &daily.Challenge{
		ID:          1,
		TargetGenre: "Pop",
		HintSongs: []string{
			"Song 1",
			"Song 2",
			"Song 3",
			"Song 4",
			"Song 5",
		},
	}
	tests := []struct {
		name          string
		initialStatus daily.GameStatus
		initialAttmpt int
		guess         string
		want          *daily.AttemptResult
		wantErr       error
	}{
		{
			name:          "1. Correct guess on first try",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 0,
			guess:         "Pop",
			want: &daily.AttemptResult{
				Correct:  true,
				Status:   daily.StatusWon,
				Attempts: 4,
				Hint:     "",
			},
			wantErr: nil,
		},
		{
			name:          "2. Incorrect guess assigns first hint",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 0,
			guess:         "Rock",
			want: &daily.AttemptResult{
				Correct:  false,
				Status:   daily.StatusPlaying,
				Attempts: 4,
				Hint:     "Song 1",
			},
			wantErr: nil,
		},
		{
			name:          "3. Correct guess ignores case and extra spaces",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 2,
			guess:         "   pOp   ",
			want: &daily.AttemptResult{
				Correct:  true,
				Status:   daily.StatusWon,
				Attempts: 2,
				Hint:     "",
			},
			wantErr: nil,
		},
		{
			name:          "4. 5th incorrect guess results in loss and no hint",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 4,
			guess:         "Jazz",
			want: &daily.AttemptResult{
				Correct:  false,
				Status:   daily.StatusLost,
				Attempts: 0,
				Hint:     "",
			},
			wantErr: nil,
		},
		{
			name:          "5. Attempting to play a finished game returns error",
			initialStatus: daily.StatusWon,
			initialAttmpt: 3,
			guess:         "Pop",
			want:          nil,
			wantErr:       daily.ErrGameOver,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &daily.Session{
				UserID:       uuid.New(),
				ChallengeID:  1,
				AttemptsUsed: tt.initialAttmpt,
				Status:       tt.initialStatus,
			}
			got, gotErr := s.MakeAttempt(tt.guess, mockChallenge)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("MakeAttempt() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				return
			}

			if *got != *tt.want {
				t.Errorf("MakeAttempt() = %v, want %v", got, tt.want)
			}
		})
	}
}
