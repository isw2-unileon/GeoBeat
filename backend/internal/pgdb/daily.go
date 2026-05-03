package pgdb

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

// PostgresDailyRepo implements the daily.Repository interface using PostgreSQL as the backend database.
type PostgresDailyRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresDailyRepo creates a new instance of the PostgresDailyRepo with the given database connection pool.
func NewPostgresDailyRepo(pool *pgxpool.Pool) *PostgresDailyRepo {
	return &PostgresDailyRepo{pool: pool}
}

// -- Challenge Methods --

// GetChallengeByDate retrieves the daily challenge for a specific date. If no challenge exists for that date, it returns an error.
func (r *PostgresDailyRepo) GetChallengeByDate(ctx context.Context, date time.Time) (*daily.Challenge, error) {
	query := `
		SELECT id, target_country, target_genre, hint_songs, play_date 
		FROM daily_challenges 
		WHERE play_date = $1::DATE
	`

	var c daily.Challenge
	err := r.pool.QueryRow(ctx, query, date).Scan(
		&c.ID,
		&c.TargetCountry,
		&c.TargetGenre,
		&c.HintSongs,
		&c.Date,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, daily.ErrChallengeNotFound
		}
		return nil, err
	}

	return &c, nil
}

// SaveDailyChallenge saves a new daily challenge to the database and updates the Challenge struct with the generated ID.
func (r *PostgresDailyRepo) SaveDailyChallenge(ctx context.Context, c *daily.Challenge) error {
	query := `
		INSERT INTO daily_challenges (target_country, target_genre, hint_songs, play_date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		c.TargetCountry,
		c.TargetGenre,
		c.HintSongs,
		c.Date,
	).Scan(&c.ID)

	return err
}

// -- Session Methods --

// GetSession retrieves a user's session for a specific challenge. If no session exists, it returns an error.
func (r *PostgresDailyRepo) GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*daily.Session, error) {
	query := `
		SELECT user_id, challenge_id, attempts_used, status, updated_at 
		FROM daily_sessions 
		WHERE user_id = $1 AND challenge_id = $2
	`

	var s daily.Session
	err := r.pool.QueryRow(ctx, query, userID, challengeID).Scan(
		&s.UserID,
		&s.ChallengeID,
		&s.AttemptsUsed,
		&s.Status,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, daily.ErrSessionNotFound
		}
		return nil, err
	}

	return &s, nil
}

// CreateSession creates a new session for a user and challenge. It returns an error if the session already exists or if there is a database error.
func (r *PostgresDailyRepo) CreateSession(ctx context.Context, s *daily.Session) error {
	query := `
		INSERT INTO daily_sessions (user_id, challenge_id, attempts_used, status, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		s.UserID,
		s.ChallengeID,
		s.AttemptsUsed,
		s.Status,
		s.UpdatedAt,
	)

	return err
}

// UpdateSession updates an existing session's attempts used, status, and updated_at timestamp. It returns an error if the session does not exist or if there is a database error.
func (r *PostgresDailyRepo) UpdateSession(ctx context.Context, s *daily.Session) error {
	query := `
		UPDATE daily_sessions 
		SET attempts_used = $1, status = $2, updated_at = $3 
		WHERE user_id = $4 AND challenge_id = $5
	`

	commandTag, err := r.pool.Exec(
		ctx,
		query,
		s.AttemptsUsed,
		s.Status,
		s.UpdatedAt,
		s.UserID,
		s.ChallengeID,
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("concurrent update failure or session deleted")
	}

	return nil
}
