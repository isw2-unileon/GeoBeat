package daily

import (
	"errors"
	"strings"
	"time"
)

type GameStatus string

const (
	StatusPlaying GameStatus = "playing"
	StatusWon     GameStatus = "won"
	StatusLost    GameStatus = "lost"
	MaxAttempts   int        = 5
)

var (
	ErrChallengeNotFound = errors.New("no challenge available for today")
	ErrGameOver          = errors.New("game is already over")
	ErrInvalidInput      = errors.New("invalid input, please try again")
)

// Challenge represents the rules for the daily challenge.
type Challenge struct {
	ID            int
	TargetCountry string
	TargetGenre   string
	HintSongs     []string
	Date          time.Time
}

// Session represents the current state of the player.
type Session struct {
	ID           int
	UserID       int
	ChallengeID  int
	AttemptsUsed int
	Status       GameStatus
	UpdatedAt    time.Time
}

// AttemptResult represents the result of a guess attempt.
type AttemptResult struct {
	Correct  bool       `json:"correct"`
	Status   GameStatus `json:"status"`
	Attempts int        `json:"attempts_remaining"`
	Hint     string     `json:"hint,omitempty"`
}

func NewSession(userID, challengeID int) *Session {
	return &Session{
		UserID:       userID,
		ChallengeID:  challengeID,
		AttemptsUsed: 0,
		Status:       StatusPlaying,
	}
}

func (s *Session) MakeAttempt(guess string, challenge *Challenge) (*AttemptResult, error) {
	if s.Status != StatusPlaying {
		return nil, ErrGameOver
	}

	isCorrect := strings.EqualFold(strings.TrimSpace(guess), challenge.TargetGenre)
	s.AttemptsUsed++

	result := &AttemptResult{
		Correct:  isCorrect,
		Attempts: MaxAttempts - s.AttemptsUsed,
	}

	if isCorrect {
		s.Status = StatusWon
	} else if s.AttemptsUsed >= MaxAttempts {
		s.Status = StatusLost
	} else {
		result.Hint = challenge.HintSongs[s.AttemptsUsed-1]
	}

	result.Status = s.Status
	s.UpdatedAt = time.Now()

	return result, nil
}
