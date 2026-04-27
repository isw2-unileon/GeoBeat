package user_test

import (
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
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
			expectedErr:  user.ErrEmptyPasswordHash,
		},
		{
			name:         "handles empty email and username",
			email:        "",
			userName:     "",
			passwordHash: "secure_hashed_password",
			expectedErr:  user.ErrEmptyEmailOrUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := user.NewUserFromEmail(tt.email, tt.userName, tt.passwordHash)

			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				return
			}

			if tt.expectedErr == nil {
				if u.Provider != user.ProviderEmail {
					t.Errorf("expected Provider %v, got %v", user.ProviderEmail, u.Provider)
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
		provider      user.AuthProvider
		emailVerified bool
		expectedErr   error
	}{
		{
			name:          "Email not verified",
			email:         "test@gmail.com",
			userName:      "Test",
			providerID:    "g_123",
			provider:      user.ProviderGoogle,
			emailVerified: false,
			expectedErr:   user.ErrEmailNotVerified,
		},
		{
			name:          "Correct creation from Google with email verified",
			email:         "test@gmail.com",
			userName:      "Test",
			providerID:    "g_123",
			provider:      user.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := user.NewUserExternal(tt.email, tt.userName, tt.providerID, tt.provider, tt.emailVerified)

			if err != tt.expectedErr {
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
	newUserEmail := func() *user.User {
		u, _ := user.NewUserFromEmail("test@gmail.com", "Test", "hashed_pass")
		return u
	}

	newUserGoogle := func() *user.User {
		u, _ := user.NewUserExternal("test@gmail.com", "Test", "g_123", user.ProviderGoogle, true)
		return u
	}

	tests := []struct {
		name          string
		initialUser   *user.User
		providerID    string
		provider      user.AuthProvider
		emailVerified bool
		expectedErr   error
		checkState    func(*testing.T, *user.User)
	}{
		{
			name:          "Email not verified",
			initialUser:   newUserEmail(),
			providerID:    "g_123",
			provider:      user.ProviderGoogle,
			emailVerified: false,
			expectedErr:   user.ErrEmailNotVerified,
			checkState: func(t *testing.T, u *user.User) {
				if u.ProviderID != nil {
					t.Errorf("the user state should not have mutated")
				}
			},
		},
		{
			name:          "Succesful link to Google account with verified email",
			initialUser:   newUserEmail(),
			providerID:    "g_123",
			provider:      user.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
			checkState: func(t *testing.T, u *user.User) {
				if u.Provider != user.ProviderGoogle {
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
			provider:      user.ProviderGoogle,
			emailVerified: true,
			expectedErr:   nil,
			checkState: func(t *testing.T, u *user.User) {
				if u.Provider != user.ProviderGoogle || *u.ProviderID != "g_123" {
					t.Errorf("original state was lost after an idempotent link")
				}
			},
		},
		{
			name:          "Overwrite attempt with a different provider ID fails",
			initialUser:   newUserGoogle(),
			providerID:    "g_999",
			provider:      user.ProviderGoogle,
			emailVerified: true,
			expectedErr:   user.ErrAccountAlreadyLinked,
			checkState: func(t *testing.T, u *user.User) {
				if *u.ProviderID != "g_123" {
					t.Errorf("original user ID was overwritten by a failed operation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initialUser.LinkExternalAccount(tt.providerID, tt.provider, tt.emailVerified)

			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.checkState != nil {
				tt.checkState(t, tt.initialUser)
			}
		})
	}
}
