package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID    `json:"id"`
	Email        string       `json:"email"`
	UserName     string       `json:"user_name"`
	PasswordHash *string      `json:"password_hash,omitempty"` // Only for email/password users
	Provider     AuthProvider `json:"provider"`
	ProviderID   *string      `json:"provider_id,omitempty"` // Only for external provider users
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type AuthProvider string

const (
	ProviderEmail  AuthProvider = "email"
	ProviderGoogle AuthProvider = "google"
	// Add more providers as needed
)

var (
	ErrEmailNotVerified     = errors.New("email not verified by provider")
	ErrAccountAlreadyLinked = errors.New("account already linked to a different provider ID")
	ErrInvalidProvider      = errors.New("invalid authentication provider")
	ErrEmptyPasswordHash    = errors.New("password hash cannot be empty")
	ErrEmptyEmailOrUsername = errors.New("email and username cannot be empty")
)

func NewUserFromEmail(email, userName, passwordHash string) (*User, error) {
	return nil, nil
}

func NewUserExternal(email, userName, providerID string, provider AuthProvider, emailVerified bool) (*User, error) {
	return nil, nil
}

func (u *User) LinkExternalAccount(providerID string, provider AuthProvider, emailVerified bool) error {
	return nil
}
