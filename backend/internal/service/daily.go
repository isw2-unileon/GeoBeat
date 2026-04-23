package service

import (
	"context"
	"errors"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

// Repository defines the interface for data access related to the daily challenge.
type Repository interface {
	GetChallengeByDate(ctx context.Context, date string) (*daily.Challenge, error)
	GetSession(ctx context.Context, userID, challengeID int) (*daily.Session, error)
	CreateSession(ctx context.Context, session *daily.Session) error
	UpdateSession(ctx context.Context, session *daily.Session) error
}

// Daily provides methods to manage the daily challenge game logic.
type Daily struct {
	repo Repository
}

// NewService creates a new Daily service with the given Repository.
func NewService(r Repository) *Daily {
	return &Daily{repo: r}
}

// GetCurrentStatus retrieves the current challenge and session status for a given user.
func (s *Daily) GetCurrentStatus(ctx context.Context, userID int) (*daily.Challenge, *daily.Session, error) {
	challenge, err := s.repo.GetChallengeByDate(ctx, "today")
	if err != nil {
		return nil, nil, daily.ErrChallengeNotFound
	}

	session, err := s.repo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session, err = daily.NewSession(userID, challenge.ID)
		if err != nil {
			return nil, nil, err
		}
		if err := s.repo.CreateSession(ctx, session); err != nil {
			return nil, nil, errors.New("error while creating session")
		}
	}

	return challenge, session, nil
}

// ProcessAttempt processes a user's guess for the daily challenge and updates the session state accordingly.
func (s *Daily) ProcessAttempt(ctx context.Context, userID int, guess string) (*daily.AttemptResult, error) {
	challenge, session, err := s.GetCurrentStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	result, err := session.MakeAttempt(guess, challenge)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, errors.New("error updating session")
	}

	return result, nil
}
