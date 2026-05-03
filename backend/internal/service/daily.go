package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

// ChallengeRepository defines the methods required to manage daily challenges.
type ChallengeRepository interface {
	GetChallengeByDate(ctx context.Context, date time.Time) (*daily.Challenge, error)
}

// SessionRepository defines the methods required to manage user sessions for the daily challenge.
type SessionRepository interface {
	GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*daily.Session, error)
	CreateSession(ctx context.Context, session *daily.Session) error
	UpdateSession(ctx context.Context, session *daily.Session) error
}

// Daily provides methods to manage the daily challenge game logic.
type Daily struct {
	challengeRepo ChallengeRepository
	sessionRepo   SessionRepository
}

// NewService creates a new Daily service with the given Repository.
func NewService(challengeRepo ChallengeRepository, sessionRepo SessionRepository) *Daily {
	return &Daily{
		challengeRepo: challengeRepo,
		sessionRepo:   sessionRepo,
	}
}

// GetCurrentStatus retrieves the current challenge and session status for a given user.
func (s *Daily) GetCurrentStatus(ctx context.Context, userID uuid.UUID) (*daily.Challenge, *daily.Session, error) {
	challenge, err := s.challengeRepo.GetChallengeByDate(ctx, time.Now())
	if err != nil {
		return nil, nil, daily.ErrChallengeNotFound
	}

	session, err := s.sessionRepo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session, err = daily.NewSession(userID, challenge.ID)
		if err != nil {
			return nil, nil, err
		}
		if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
			return nil, nil, errors.New("error while creating session")
		}
	}

	return challenge, session, nil
}

// ProcessAttempt processes a user's guess for the daily challenge and updates the session state accordingly.
func (s *Daily) ProcessAttempt(ctx context.Context, userID uuid.UUID, guess string) (*daily.AttemptResult, error) {
	challenge, session, err := s.GetCurrentStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	result, err := session.MakeAttempt(guess, challenge)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		return nil, errors.New("error updating session")
	}

	return result, nil
}
