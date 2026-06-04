package daily_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

func TestSession_MakeInverseAttempt(t *testing.T) {
	mockChallenge := &daily.InverseChallenge{
		ID:            1,
		GivenSongName: "Song 1",
		TargetCountry: "Pop",
		Date:          time.Now(),
	}
	tests := []struct {
		name          string
		initialStatus daily.GameStatus
		initialAttmpt int
		guess         string
		want          *daily.InverseAttemptResult
		wantErr       error
	}{
		{
			name:          "1. Correct guess on first try",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 0,
			guess:         "Pop",
			want: &daily.InverseAttemptResult{
				Correct:  true,
				Status:   daily.StatusWon,
				Attempts: 4,
			},
			wantErr: nil,
		},
		{
			name:          "2. Correct guess ignores case and extra spaces",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 2,
			guess:         "   pOp   ",
			want: &daily.InverseAttemptResult{
				Correct:  true,
				Status:   daily.StatusWon,
				Attempts: 2,
			},
			wantErr: nil,
		},
		{
			name:          "3. 5th incorrect guess results in loss",
			initialStatus: daily.StatusPlaying,
			initialAttmpt: 4,
			guess:         "Jazz",
			want: &daily.InverseAttemptResult{
				Correct:  false,
				Status:   daily.StatusLost,
				Attempts: 0,
			},
			wantErr: nil,
		},
		{
			name:          "4. Attempting to play a finished game returns error",
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
			got, gotErr := s.MakeInverseAttempt(tt.guess, mockChallenge)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("MakeInverseAttempt() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				return
			}

			if *got != *tt.want {
				t.Errorf("MakeInverseAttempt() = %v, want %v", got, tt.want)
			}
		})
	}
}
