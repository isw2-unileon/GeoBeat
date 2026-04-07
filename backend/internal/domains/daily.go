package domains

import (
	"errors"
	"strings"
	"time"
)

// GameStatus represents the current state of the game session.
type GameStatus string

const (
	// StatusPlaying indicates the game is ongoing, the player can still make attempts.
	StatusPlaying GameStatus = "playing"
	// StatusWon indicates the player has guessed correctly and won the game.
	StatusWon GameStatus = "won"
	// StatusLost indicates the player has used all attempts without guessing correctly and lost the game.
	StatusLost GameStatus = "lost"
	// MaxAttempts defines the maximum number of attempts a player has to guess the correct answer.
	MaxAttempts int = 5
)

var (
	// ErrChallengeNotFound is returned when there is no challenge available for the current day.
	ErrChallengeNotFound = errors.New("no challenge available for today")
	// ErrGameOver is returned when a player tries to make an attempt after the game has already ended.
	ErrGameOver = errors.New("game is already over")
	// ErrInvalidInput is returned when the player's guess is invalid (e.g., empty or not a valid genre).
	ErrInvalidInput = errors.New("invalid input, please try again")
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

// NewSession creates a new game session for a user and challenge.
func NewSession(userID, challengeID int) *Session {
	return &Session{
		UserID:       userID,
		ChallengeID:  challengeID,
		AttemptsUsed: 0,
		Status:       StatusPlaying,
	}
}

// MakeAttempt processes a player's guess and updates the session state accordingly.
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

	switch {
	case isCorrect:
		s.Status = StatusWon
	case s.AttemptsUsed >= 5:
		s.Status = StatusLost
	default:
		result.Hint = challenge.HintSongs[s.AttemptsUsed-1]
	}

	result.Status = s.Status
	s.UpdatedAt = time.Now()

	return result, nil
}
