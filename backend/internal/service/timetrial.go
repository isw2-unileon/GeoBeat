package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

// TimetrialChallengeRepository defines the methods required to manage daily challenges.
type TimetrialChallengeRepository interface {
	GetChallengeByDate(ctx context.Context, date time.Time) (*timetrial.Challenge, error)
}

// TimetrialSessionRepository defines the methods required to manage user sessions for the time trial challenge.
type TimetrialSessionRepository interface {
	CreateSession(ctx context.Context, s *timetrial.Session) error
	GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*timetrial.Session, error)
	UpdateSession(ctx context.Context, s *timetrial.Session) error
}

// TimetrialLeaderboardRepository defines the methods required to manage the leaderboard for the time trial challenge.
type TimetrialLeaderboardRepository interface {
	GetLeaderboard(ctx context.Context, challengeID int, userID uuid.UUID) (*timetrial.Leaderboard, error)
}

// TimetrialService orchestrates the domain logic and the storage layer.
type TimetrialService struct {
	challengeRepo   TimetrialChallengeRepository
	sessionRepo     TimetrialSessionRepository
	leaderboardRepo TimetrialLeaderboardRepository
	logger          *slog.Logger
}

// NewTimetrialService creates a new instance of the service with injected dependencies.
func NewTimetrialService(challengeRepo TimetrialChallengeRepository, sessionRepo TimetrialSessionRepository, leaderboardRepo TimetrialLeaderboardRepository, logger *slog.Logger) *TimetrialService {
	return &TimetrialService{
		challengeRepo:   challengeRepo,
		sessionRepo:     sessionRepo,
		leaderboardRepo: leaderboardRepo,
		logger:          logger,
	}
}

// GetCurrentStatus retrieves the current game state for a user, including the session and the next target country if the game is still active.
func (s *TimetrialService) GetCurrentStatus(ctx context.Context, userID uuid.UUID) (*timetrial.Session, string, error) {
	challenge, err := s.challengeRepo.GetChallengeByDate(ctx, time.Now())
	if err != nil {
		return nil, "", timetrial.ErrChallengeNotFound
	}

	session, err := s.sessionRepo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		return nil, "", err
	}

	if session.Status == timetrial.StatusPlaying && session.CurrentIndex < len(challenge.TargetCountries) {
		return session, challenge.TargetCountries[session.CurrentIndex], nil
	}

	return session, "", nil
}

// StartGame initializes a new game session for the user or retrieves the existing one if it already exists.
func (s *TimetrialService) StartGame(ctx context.Context, userID uuid.UUID) (*timetrial.Session, string, error) {
	challenge, err := s.challengeRepo.GetChallengeByDate(ctx, time.Now())
	if err != nil {
		return nil, "", timetrial.ErrChallengeNotFound
	}

	_, err = s.sessionRepo.GetSession(ctx, userID, challenge.ID)

	if err == nil {
		return s.GetCurrentStatus(ctx, userID)
	}

	if !errors.Is(err, timetrial.ErrSessionNotFound) {
		s.logger.Error("database error while fetching session", "error", err, "user_id", userID)
		return nil, "", err
	}

	newSession, err := timetrial.NewSession(userID, challenge.ID)
	if err != nil {
		return nil, "", err
	}

	if err := s.sessionRepo.CreateSession(ctx, newSession); err != nil {
		s.logger.Error("failed to persist new session", "error", err, "user_id", userID)
		return nil, "", err
	}

	return newSession, challenge.TargetCountries[newSession.CurrentIndex], nil
}

// ProcessAttempt handles a user's guess, delegates business rules to the domain, and persists state changes.
func (s *TimetrialService) ProcessAttempt(ctx context.Context, userID uuid.UUID, guess string) (*timetrial.AttemptResult, error) {
	challenge, err := s.challengeRepo.GetChallengeByDate(ctx, time.Now())
	if err != nil {
		return nil, timetrial.ErrChallengeNotFound
	}

	session, err := s.sessionRepo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		return nil, err
	}

	result, err := session.MakeAttempt(guess, challenge)
	if err != nil {
		return nil, err
	}

	if result.Correct {
		if updateErr := s.sessionRepo.UpdateSession(ctx, session); updateErr != nil {
			s.logger.Error("failed to update session after attempt", "error", updateErr, "user_id", userID)
			return nil, updateErr
		}
	}

	return result, nil
}

// GetLeaderboard fetches today's ranking.
func (s *TimetrialService) GetLeaderboard(ctx context.Context, userID uuid.UUID) (*timetrial.Leaderboard, error) {
	challenge, err := s.challengeRepo.GetChallengeByDate(ctx, time.Now())
	if err != nil {
		return nil, timetrial.ErrChallengeNotFound
	}

	leaderboard, err := s.leaderboardRepo.GetLeaderboard(ctx, challenge.ID, userID)
	if err != nil {
		s.logger.Error("failed to fetch leaderboard", "error", err, "challenge_id", challenge.ID)
		return nil, err
	}

	return leaderboard, nil
}
