package service

import (
	"context"
	"errors"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

type Repository interface {
	GetChallengeByDate(ctx context.Context, date string) (*daily.Challenge, error)
	GetSession(ctx context.Context, userID, challengeID int) (*daily.Session, error)
	CreateSession(ctx context.Context, session *daily.Session) error
	UpdateSession(ctx context.Context, session *daily.Session) error
}

type Daily struct {
	repo Repository
}

func NewService(r Repository) *Daily {
	return &Daily{repo: r}
}

func (s *Daily) GetCurrentStatus(ctx context.Context, userID int) (*daily.Challenge, *daily.Session, error) {
	challenge, err := s.repo.GetChallengeByDate(ctx, "today")
	if err != nil {
		return nil, nil, daily.ErrChallengeNotFound
	}

	session, err := s.repo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session = daily.NewSession(userID, challenge.ID)
		if err := s.repo.CreateSession(ctx, session); err != nil {
			return nil, nil, errors.New("error al iniciar sesión")
		}
	}

	return challenge, session, nil
}

func (s *Daily) ProcessAttempt(ctx context.Context, userID int, guess string) (*daily.AttemptResult, error) {
	challenge, err := s.repo.GetChallengeByDate(ctx, "today")
	if err != nil {
		return nil, daily.ErrChallengeNotFound
	}

	session, err := s.repo.GetSession(ctx, userID, challenge.ID)
	if err != nil {
		session = daily.NewSession(userID, challenge.ID)
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
