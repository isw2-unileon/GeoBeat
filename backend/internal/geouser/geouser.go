package geouser

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system, containing fields for identification, authentication, and metadata. It supports both email/password and external provider authentication methods.
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

// AuthProvider defines the type for authentication providers, allowing for easy extension to support multiple providers in the future
type AuthProvider string

const (
	// ProviderEmail represents users authenticated via email/password
	ProviderEmail AuthProvider = "email"
	// ProviderGoogle represents users authenticated via Google OAuth
	ProviderGoogle AuthProvider = "google"
	// Add more providers as needed
)

var (
	// ErrNotFound is returned when a user is not found in the repository
	ErrNotFound = errors.New("user not found")
	// ErrEmailNotVerified is returned when an external provider indicates that the user's email is not verified
	ErrEmailNotVerified = errors.New("email not verified by provider")
	// ErrAccountAlreadyLinked is returned when trying to link an external account to a user that is already linked to a different provider
	ErrAccountAlreadyLinked = errors.New("account already linked to a different auth provider")
	// ErrEmptyPassword is returned when trying to create a user with an empty password
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrEmptyEmailOrUsername is returned when trying to create a user with an empty email or username
	ErrEmptyEmailOrUsername = errors.New("email and username cannot be empty")
)

// NewUserFromEmail creates a new User instance for email/password registration. It validates that the email, username, and password hash are not empty, returning an error if any of these conditions are not met.
func NewUserFromEmail(email, userName, passwordHash string) (*User, error) {
	if email == "" || userName == "" {
		return nil, ErrEmptyEmailOrUsername
	}
	if passwordHash == "" {
		return nil, ErrEmptyPassword
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

// NewUserExternal creates a new User instance for external provider registration. It validates that the email is verified by the provider, returning an error if it is not. The ProviderID is set to the unique identifier provided by the external authentication provider.
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

// LinkExternalAccount links an external authentication provider to an existing user account. It checks if the email is verified by the provider and if the user is already linked to a different provider, returning appropriate errors in those cases. If the linking is successful, it updates the user's Provider and ProviderID fields.
func (u *User) LinkExternalAccount(providerID string, provider AuthProvider, emailVerified bool) error {
	if !emailVerified {
		return ErrEmailNotVerified
	}
	if u.Provider != ProviderEmail && u.Provider != provider {
		return ErrAccountAlreadyLinked
	}
	u.Provider = provider
	u.ProviderID = &providerID
	u.UpdatedAt = time.Now()
	return nil
}
