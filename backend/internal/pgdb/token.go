package pgdb

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTokenRepo implements the token.Repository interface.
type PostgresTokenRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresTokenRepo creates a new instance of the repository.
func NewPostgresTokenRepo(pool *pgxpool.Pool) *PostgresTokenRepo {
	return &PostgresTokenRepo{pool: pool}
}

// Save inserts a new refresh token into the database.
func (r *PostgresTokenRepo) Save(ctx context.Context, token *authsession.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.CreatedAt,
		token.ExpiresAt,
	)

	return err
}

// FindByTokenHash retrieves a refresh token by its hash value.
func (r *PostgresTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*authsession.RefreshToken, error) {
	query := `
		SELECT user_id, hash, created_at, expires_at
		FROM refresh_tokens
		WHERE hash = $1
	`

	var token authsession.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.UserID,
		&token.TokenHash,
		&token.CreatedAt,
		&token.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authsession.ErrNotFound
		}
		return nil, err
	}

	return &token, nil
}

// FindByUserID retrieves a refresh token by the associated user ID.
func (r *PostgresTokenRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*authsession.RefreshToken, error) {
	query := `
		SELECT user_id, hash, created_at, expires_at
		FROM refresh_tokens
		WHERE user_id = $1
	`

	var token authsession.RefreshToken
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&token.UserID,
		&token.TokenHash,
		&token.CreatedAt,
		&token.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authsession.ErrNotFound
		}
		return nil, err
	}

	return &token, nil
}

// Delete removes a refresh token from the database by its hash value.
func (r *PostgresTokenRepo) Delete(ctx context.Context, tokenHash string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE hash = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return authsession.ErrNotFound
	}

	return nil
}
