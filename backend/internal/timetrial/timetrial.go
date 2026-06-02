package timetrial

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GameStatus represents the current state of the game session.
type GameStatus string

const (
	// StatusPlaying indicates the game is ongoing, the player can still make attempts.
	StatusPlaying GameStatus = "playing"
	// StatusCompleted indicates the player has guessed correctly and won the game.
	StatusCompleted GameStatus = "completed"
)

var (
	// ErrChallengeNotFound is returned when there is no challenge available for the current day.
	ErrChallengeNotFound = errors.New("no challenge available for today")
	// ErrSessionNotFound is returned when there is no session available for the user and challenge.
	ErrSessionNotFound = errors.New("no session available for this user and challenge")
	// ErrInvalidInput is returned when the player's guess is invalid (e.g., empty or not a valid genre).
	ErrInvalidInput = errors.New("invalid input, please try again")
	// ErrInvalidID is returned when the user ID or challenge ID is invalid (e.g., non-positive integers).
	ErrInvalidID = errors.New("invalid user or challenge ID")
	// ErrAlreadyCompleted is returned when a player tries to make an attempt after the game has already ended.
	ErrAlreadyCompleted = errors.New("time trial challenge already completed")
)

// Challenge represents the rules for the time trial challenge.
type Challenge struct {
	ID              int
	TargetCountries []string
	TargetGenres    []string
	Date            time.Time
}

// Session represents the current state of the player.
type Session struct {
	UserID       uuid.UUID
	ChallengeID  int
	CurrentIndex int
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Status       GameStatus
}

// AttemptResult represents the result of a guess attempt.
type AttemptResult struct {
	Correct     bool          `json:"correct"`
	Status      GameStatus    `json:"status"`
	NextCountry string        `json:"next_country,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// LeaderboardEntry represents a single entry in the leaderboard.
type LeaderboardEntry struct {
	UserName string        `json:"username"`
	Duration time.Duration `json:"duration"`
	Rank     int           `json:"rank"`
}

// Leaderboard groups the rankings strictly tied to a single Challenge.
type Leaderboard struct {
	ChallengeID int                `json:"challenge_id"`
	Entries     []LeaderboardEntry `json:"entries"`
	UserEntry   *LeaderboardEntry  `json:"user_entry,omitempty"`
}

// NewSession creates a new game session for a user and challenge.
func NewSession(userID uuid.UUID, challengeID int) (*Session, error) {
	if challengeID <= 0 || userID == uuid.Nil {
		return nil, ErrInvalidID
	}
	return &Session{
		UserID:       userID,
		ChallengeID:  challengeID,
		CurrentIndex: 0,
		StartTime:    time.Now(),
		Status:       StatusPlaying,
	}, nil
}

// MakeAttempt processes a player's guess and updates the session state accordingly.
func (s *Session) MakeAttempt(guess string, challenge *Challenge) (*AttemptResult, error) {
	if s.Status != StatusPlaying {
		return nil, ErrAlreadyCompleted
	}

	if s.CurrentIndex >= len(challenge.TargetCountries) {
		return nil, errors.New("critical state error: index out of bounds")
	}

	isCorrect := strings.EqualFold(strings.TrimSpace(guess), challenge.TargetGenres[s.CurrentIndex])

	result := &AttemptResult{
		Correct: isCorrect,
	}

	if isCorrect {
		s.CurrentIndex++
		if s.CurrentIndex >= len(challenge.TargetCountries) {
			s.Status = StatusCompleted
			s.EndTime = time.Now()
			s.Duration = s.EndTime.Sub(s.StartTime)
			result.Duration = s.Duration
		}
	}

	result.Status = s.Status

	if s.Status == StatusPlaying {
		result.NextCountry = challenge.TargetCountries[s.CurrentIndex]
	}

	return result, nil
}
