package authsession

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const refreshTokenValidity = 1 * 24 * time.Hour // 1 day

var (
	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = errors.New("refresh token not found")
	// ErrExpired indicates that the refresh token has expired
	ErrExpired = errors.New("refresh token has expired")
)

// RefreshToken represents a refresh token used for obtaining new authentication tokens.
type RefreshToken struct {
	UserID    uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewRefreshToken creates a new RefreshToken for the given user ID and token hash.
func NewRefreshToken(userID uuid.UUID, tokenHash string) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(refreshTokenValidity),
	}
}

// IsExpired checks if the refresh token has expired based on the current time and the ExpiresAt field.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}
