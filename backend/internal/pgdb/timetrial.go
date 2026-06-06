package pgdb

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTimetrialRepo implements the timetrial repository interface.
type PostgresTimetrialRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresTimetrialRepo creates a new instance of the repository.
func NewPostgresTimetrialRepo(pool *pgxpool.Pool) *PostgresTimetrialRepo {
	return &PostgresTimetrialRepo{pool: pool}
}

// -- Challenge Methods --

// GetChallengeByDate retrieves the time trial challenge for a specific date. If no challenge exists for that date, it returns an error.
func (r *PostgresTimetrialRepo) GetChallengeByDate(ctx context.Context, date time.Time) (*timetrial.Challenge, error) {
	query := `
		SELECT id, target_countries, target_genres, play_date 
		FROM timetrial_challenges 
		WHERE play_date = $1::DATE
	`

	c := &timetrial.Challenge{}
	err := r.pool.QueryRow(ctx, query, date).Scan(
		&c.ID,
		&c.TargetCountries,
		&c.TargetGenres,
		&c.Date,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, timetrial.ErrChallengeNotFound
		}
		return nil, err
	}

	return c, err
}

// SaveChallenge saves a new time trial challenge to the database and updates the Challenge struct with the generated ID.
func (r *PostgresTimetrialRepo) SaveChallenge(ctx context.Context, c *timetrial.Challenge) error {
	query := `
		INSERT INTO timetrial_challenges (target_countries, target_genres, play_date) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query, c.TargetCountries, c.TargetGenres, c.Date).Scan(&c.ID)
	return err
}

// -- Session Methods --

// CreateSession saves a new game session to the database.
func (r *PostgresTimetrialRepo) CreateSession(ctx context.Context, s *timetrial.Session) error {
	query := `
		INSERT INTO timetrial_sessions (user_id, challenge_id, current_index, start_time, status) 
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, s.UserID, s.ChallengeID, s.CurrentIndex, s.StartTime, s.Status)
	return err
}

// GetSession retrieves a user's session for a specific challenge. If no session exists, it returns an error.
func (r *PostgresTimetrialRepo) GetSession(ctx context.Context, userID uuid.UUID, challengeID int) (*timetrial.Session, error) {
	query := `
		SELECT user_id, challenge_id, current_index, start_time, end_time, duration, status 
		FROM timetrial_sessions 
		WHERE user_id = $1 AND challenge_id = $2
	`

	s := &timetrial.Session{}

	var endTime *time.Time
	var duration *int64

	err := r.pool.QueryRow(ctx, query, userID, challengeID).Scan(
		&s.UserID,
		&s.ChallengeID,
		&s.CurrentIndex,
		&s.StartTime,
		&endTime,
		&duration,
		&s.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, timetrial.ErrSessionNotFound
		}
		return nil, err
	}

	if endTime != nil {
		s.EndTime = *endTime
	}
	if duration != nil {
		s.Duration = time.Duration(*duration)
	}

	return s, nil
}

// UpdateSession updates an existing game session in the database. It returns an error if the session does not exist or if there is a database error.
func (r *PostgresTimetrialRepo) UpdateSession(ctx context.Context, s *timetrial.Session) error {
	query := `
		UPDATE timetrial_sessions 
		SET current_index = $1, end_time = $2, duration = $3, status = $4 
		WHERE user_id = $5 AND challenge_id = $6
	`

	var dbEndTime *time.Time
	var dbDuration *int64
	if !s.EndTime.IsZero() {
		dbEndTime = &s.EndTime
	}
	if s.Duration != 0 {
		durationInt := int64(s.Duration)
		dbDuration = &durationInt
	}
	result, err := r.pool.Exec(ctx, query, s.CurrentIndex, dbEndTime, dbDuration, s.Status, s.UserID, s.ChallengeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return timetrial.ErrSessionNotFound
	}
	return nil
}

// -- Leaderboard Methods --

// GetLeaderboard retrieves the leaderboard for a specific challenge, including the user's own entry if it exists.
func (r *PostgresTimetrialRepo) GetLeaderboard(ctx context.Context, challengeID int, userID uuid.UUID) (*timetrial.Leaderboard, error) {
	query := `
	WITH ranked_sessions AS (
		SELECT u.user_name, s.duration, CAST(RANK() OVER (ORDER BY s.duration ASC) AS INT) as rank, s.user_id
		FROM timetrial_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.challenge_id = $1 AND s.status = 'completed'
	)
	SELECT user_name, duration, rank, user_id 
	FROM ranked_sessions
	WHERE rank <= 100 OR user_id = $2
	ORDER BY rank ASC;
	`

	rows, err := r.pool.Query(ctx, query, challengeID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []timetrial.LeaderboardEntry
	var userEntry *timetrial.LeaderboardEntry

	var leaderboard timetrial.Leaderboard
	leaderboard.ChallengeID = challengeID

	for rows.Next() {
		var entry timetrial.LeaderboardEntry
		var entryUserID uuid.UUID
		if err := rows.Scan(&entry.UserName, &entry.Duration, &entry.Rank, &entryUserID); err != nil {
			return nil, err
		}
		if entry.Rank <= 100 {
			entries = append(entries, entry)
		}
		if entryUserID == userID {
			userEntry = &entry
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	leaderboard.Entries = entries
	leaderboard.UserEntry = userEntry

	return &leaderboard, err
}
