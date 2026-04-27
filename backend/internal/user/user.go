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
	if email == "" || userName == "" {
		return nil, ErrEmptyEmailOrUsername
	}
	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}
	return &User{
		ID:           uuid.New(),
		Email:        email,
		UserName:     userName,
		PasswordHash: &passwordHash,
		Provider:     ProviderEmail,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func NewUserExternal(email, userName, providerID string, provider AuthProvider, emailVerified bool) (*User, error) {
	if !emailVerified {
		return nil, ErrEmailNotVerified
	}
	return &User{
		ID:         uuid.New(),
		Email:      email,
		UserName:   userName,
		Provider:   provider,
		ProviderID: &providerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (u *User) LinkExternalAccount(providerID string, provider AuthProvider, emailVerified bool) error {
	if !emailVerified {
		return ErrEmailNotVerified
	}
	if u.ProviderID != nil && *u.ProviderID != providerID {
		return ErrAccountAlreadyLinked
	}
	u.Provider = provider
	u.ProviderID = &providerID
	u.UpdatedAt = time.Now()
	return nil
}
