package services

import (
	"context"
	"errors"

	"github.com/isw2-unileon/GeoBeat/backend/internal/domains"
)

// Repository defines the interface for data access related to the daily challenge.
type Repository interface {
	GetChallengeByDate(ctx context.Context, date string) (*domains.Challenge, error)
	GetSession(ctx context.Context, userID, challengeID int) (*domains.Session, error)
	CreateSession(ctx context.Context, session *domains.Session) error
	UpdateSession(ctx context.Context, session *domains.Session) error
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
func (s *Daily) GetCurrentStatus(ctx context.Context, userID int) (*domains.Challenge, *domains.Session, error) {
	challenge, err := s.repo.GetChallengeByDate(ctx, "today")
	if err != nil {
		return nil, nil, domains.ErrChallengeNotFound
	}

	session, err := s.repo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session = domains.NewSession(userID, challenge.ID)
		if err := s.repo.CreateSession(ctx, session); err != nil {
			return nil, nil, errors.New("error al iniciar sesión")
		}
	}

	return challenge, session, nil
}

// ProcessAttempt processes a user's guess for the daily challenge and updates the session state accordingly.
func (s *Daily) ProcessAttempt(ctx context.Context, userID int, guess string) (*domains.AttemptResult, error) {
	challenge, err := s.repo.GetChallengeByDate(ctx, "today")
	if err != nil {
		return nil, domains.ErrChallengeNotFound
	}

	session, err := s.repo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session = domains.NewSession(userID, challenge.ID)
		if err := s.repo.CreateSession(ctx, session); err != nil {
			return nil, errors.New("error creating session")
		}
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
