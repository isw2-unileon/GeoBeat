package pgdb

import (
	"context"
	"errors"

	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepo implements the user.Repository interface.
type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepo creates a new instance of the repository.
func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

// FindByEmail retrieves a user by their exact email address.
func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*geouser.User, error) {
	query := `
		SELECT id, email, user_name, password_hash, provider, provider_id, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	var u geouser.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.UserName,
		&u.PasswordHash,
		&u.Provider,
		&u.ProviderID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, geouser.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

// Save inserts a new user into the database and updates the struct with the new ID.
func (r *PostgresUserRepo) Save(ctx context.Context, u *geouser.User) error {
	query := `
		INSERT INTO users (id, email, user_name, password_hash, provider, provider_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		u.ID,
		u.Email,
		u.UserName,
		u.PasswordHash,
		u.Provider,
		u.ProviderID,
		u.CreatedAt,
		u.UpdatedAt,
	)

	return err
}

// Update modifies an existing user's record based on their ID.
func (r *PostgresUserRepo) Update(ctx context.Context, u *geouser.User) error {
	query := `
		UPDATE users 
		SET provider = $1, provider_id = $2, updated_at = $3 
		WHERE id = $4
	`

	commandTag, err := r.pool.Exec(ctx, query, u.Provider, u.ProviderID, u.UpdatedAt, u.ID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("cannot update: user not found")
	}

	return nil
}
