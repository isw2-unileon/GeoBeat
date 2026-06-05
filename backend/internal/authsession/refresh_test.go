package authsession_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
)

func TestNewRefreshToken(t *testing.T) {
	validUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name      string
		userID    uuid.UUID
		tokenHash string
	}{
		{
			name:      "valid input",
			userID:    validUserID,
			tokenHash: "validTokenHash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := authsession.NewRefreshToken(tt.userID, tt.tokenHash)

			if rt.UserID != tt.userID {
				t.Errorf("expected UserID %v, got %v", tt.userID, rt.UserID)
			}
			if rt.TokenHash != tt.tokenHash {
				t.Errorf("expected TokenHash %s, got %s", tt.tokenHash, rt.TokenHash)
			}
			if rt.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be set, got zero value")
			}
			if rt.ExpiresAt.IsZero() {
				t.Error("expected ExpiresAt to be set, got zero value")
			}
		})
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	validUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name     string
		refresh  *authsession.RefreshToken
		expected bool
	}{
		{
			name: "not expired",
			refresh: &authsession.RefreshToken{
				UserID:    validUserID,
				TokenHash: "validTokenHash",
				CreatedAt: time.Now().Add(-1 * time.Hour),
				ExpiresAt: time.Now().Add(23 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			refresh: &authsession.RefreshToken{
				UserID:    validUserID,
				TokenHash: "validTokenHash",
				CreatedAt: time.Now().Add(-25 * time.Hour),
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.refresh.IsExpired() != tt.expected {
				t.Errorf("expected IsExpired to return %v, got %v", tt.expected, tt.refresh.IsExpired())
			}
		})
	}
}
