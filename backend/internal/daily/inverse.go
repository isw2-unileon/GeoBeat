package daily

import (
	"strings"
	"time"
)

// InverseChallenge represents the rules for the daily challenge.
type InverseChallenge struct {
	ID            int
	GivenSongName string
	TargetCountry string
	Date          time.Time
}

// InverseAttemptResult represents the result of a guess attempt.
type InverseAttemptResult struct {
	Correct  bool       `json:"correct"`
	Status   GameStatus `json:"status"`
	Attempts int        `json:"attempts_remaining"`
}

// MakeAttempt processes a player's guess and updates the session state accordingly.
func (s *Session) MakeInverseAttempt(guess string, challenge *InverseChallenge) (*InverseAttemptResult, error) {
	if s.Status != StatusPlaying {
		return nil, ErrGameOver
	}

	isCorrect := strings.EqualFold(strings.TrimSpace(guess), challenge.TargetCountry)
	s.AttemptsUsed++

	result := &InverseAttemptResult{
		Correct:  isCorrect,
		Attempts: MaxAttempts - s.AttemptsUsed,
	}

	switch {
	case isCorrect:
		s.Status = StatusWon
	case s.AttemptsUsed >= 5:
		s.Status = StatusLost
	}

	result.Status = s.Status
	s.UpdatedAt = time.Now()

	return result, nil
}
