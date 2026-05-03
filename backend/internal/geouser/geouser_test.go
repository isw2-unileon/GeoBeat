package geouser_test

import (
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
)

func TestNewUserFromEmail(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		userName     string
		passwordHash string
		expectedErr  error
	}{
		{
			name:         "creates user correctly by email",
			email:        "internal@test.com",
			userName:     "User Internal",
			passwordHash: "secure_hashed_password",
			expectedErr:  nil,
		},
		{
			name:         "handles empty password hash",
			email:        "internal@test.com",
			userName:     "Empty Password",
			passwordHash: "",
			expectedErr:  geouser.ErrEmptyPassword,
		},
		{
			name:         "handles empty email and username",
			email:        "",
			userName:     "",
			passwordHash: "secure_hashed_password",
			expectedErr:  geouser.ErrEmptyEmailOrUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := geouser.NewUserFromEmail(tt.email, tt.userName, tt.passwordHash)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				return
			}

			if tt.expectedErr == nil {
				if u.Provider != geouser.ProviderEmail {
					t.Errorf("expected Provider %v, got %v", geouser.ProviderEmail, u.Provider)
				}
				if u.ProviderID != nil {
					t.Errorf("expected ProviderID nil, got %v", *u.ProviderID)
				}
			}
		})
	}
}

func TestNewUserExternal(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		userName      string
		providerID    string
		provider      geouser.AuthProvider
		emailVerified bool
		expectedErr   error
	}{
		{
			name:          "Email not verified",
			email:         "test@gmail.com",
			userName:      "Test",
			providerID:    "g_123",
			provider:      geouser.ProviderGoogle,
			emailVerified: false,
			expectedErr:   geouser.ErrEmailNotVerified,
		},
		{
			name:          "Correct creation from Google with email verified",
			email:         "test@gmail.com",
			userName:      "Test",
			providerID:    "g_123",
			provider:      geouser.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := geouser.NewUserExternal(tt.email, tt.userName, tt.providerID, tt.provider, tt.emailVerified)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil {
				if u.Provider != tt.provider {
					t.Errorf("expected Provider %v, got %v", tt.provider, u.Provider)
				}
				if u.ProviderID == nil || *u.ProviderID != tt.providerID {
					t.Errorf("expected ProviderID %v, got a different or nil one", tt.providerID)
				}
				if u.PasswordHash != nil {
					t.Errorf("expected PasswordHash nil, got a value")
				}
			}
		})
	}
}

func TestLinkExternalAccount(t *testing.T) {
	newUserEmail := func() *geouser.User {
		u, _ := geouser.NewUserFromEmail("test@gmail.com", "Test", "hashed_pass")
		return u
	}

	newUserGoogle := func() *geouser.User {
		u, _ := geouser.NewUserExternal("test@gmail.com", "Test", "g_123", geouser.ProviderGoogle, true)
		return u
	}

	tests := []struct {
		name          string
		initialUser   *geouser.User
		providerID    string
		provider      geouser.AuthProvider
		emailVerified bool
		expectedErr   error
		checkState    func(*testing.T, *geouser.User)
	}{
		{
			name:          "Email not verified",
			initialUser:   newUserEmail(),
			providerID:    "g_123",
			provider:      geouser.ProviderGoogle,
			emailVerified: false,
			expectedErr:   geouser.ErrEmailNotVerified,
			checkState: func(t *testing.T, u *geouser.User) {
				if u.ProviderID != nil {
					t.Errorf("the user state should not have mutated")
				}
			},
		},
		{
			name:          "Successful link to Google account with verified email",
			initialUser:   newUserEmail(),
			providerID:    "g_123",
			provider:      geouser.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
			checkState: func(t *testing.T, u *geouser.User) {
				if u.Provider != geouser.ProviderGoogle {
					t.Errorf("expected provider to mutate to 'google', got %v", u.Provider)
				}
				if u.ProviderID == nil || *u.ProviderID != "g_123" {
					t.Errorf("expected google provider ID to be stored")
				}
			},
		},
		{
			name:          "Same provider and ID is idempotent",
			initialUser:   newUserGoogle(),
			providerID:    "g_123",
			provider:      geouser.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
			checkState: func(t *testing.T, u *geouser.User) {
				if u.Provider != geouser.ProviderGoogle || *u.ProviderID != "g_123" {
					t.Errorf("original state was lost after an idempotent link")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initialUser.LinkExternalAccount(tt.providerID, tt.provider, tt.emailVerified)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.checkState != nil {
				tt.checkState(t, tt.initialUser)
			}
		})
	}
}
