package user_test

import (
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type userCreationTest struct {
	name        string
	create      func() (*user.User, error)
	expectedErr error
	verify      func(*testing.T, *user.User)
}

func runUserCreationTests(t *testing.T, tests []userCreationTest) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			u, err := tt.create()
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.verify != nil {
				tt.verify(t, u)
			}
		})
	}
}

func TestNewUserFromEmail(t *testing.T) {
	tests := []userCreationTest{
		{
			name: "creates user correctly by email",
			create: func() (*user.User, error) {
				return user.NewUserFromEmail("internal@test.com", "User Internal", "secure_hashed_password")
			},
			expectedErr: nil,
			verify: func(t *testing.T, u *user.User) {
				if u == nil {
					t.Fatal("expected non-nil user")
				}
				if u.Provider != user.ProviderEmail {
					t.Errorf("expected Provider %v, got %v", user.ProviderEmail, u.Provider)
				}
				if u.ProviderID != nil {
					t.Errorf("expected ProviderID nil, got %v", *u.ProviderID)
				}
			},
		},
		{
			name: "handles empty password hash",
			create: func() (*user.User, error) {
				return user.NewUserFromEmail("internal@test.com", "Empty Password", "")
			},
			expectedErr: user.ErrEmptyPasswordHash,
		},
		{
			name: "handles empty email and username",
			create: func() (*user.User, error) {
				return user.NewUserFromEmail("", "", "secure_hashed_password")
			},
			expectedErr: user.ErrEmptyEmailOrUsername,
		},
	}

	runUserCreationTests(t, tests)
}

func TestNewUserExternal(t *testing.T) {
	tests := []userCreationTest{
		{
			name: "Email not verified",
			create: func() (*user.User, error) {
				return user.NewUserExternal("test@gmail.com", "Test", "g_123", user.ProviderGoogle, false)
			},
			expectedErr: user.ErrEmailNotVerified,
		},
		{
			name: "Correct creation from Google with email verified",
			create: func() (*user.User, error) {
				return user.NewUserExternal("test@gmail.com", "Test", "g_123", user.ProviderGoogle, true)
			},
			expectedErr: nil,
			verify: func(t *testing.T, u *user.User) {
				if u == nil {
					t.Fatal("expected non-nil user")
				}
				if u.Provider != user.ProviderGoogle {
					t.Errorf("expected Provider %v, got %v", user.ProviderGoogle, u.Provider)
				}
				if u.ProviderID == nil || *u.ProviderID != "g_123" {
					t.Errorf("expected ProviderID %v, got %v", "g_123", u.ProviderID)
				}
				if u.PasswordHash != nil {
					t.Errorf("expected PasswordHash nil, got a value")
				}
			},
		},
	}

	runUserCreationTests(t, tests)
}

type userMutationTest struct {
	name        string
	initial     func() *user.User
	action      func(*user.User) error
	expectedErr error
	verify      func(*testing.T, *user.User)
}

func runUserMutationTests(t *testing.T, tests []userMutationTest) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			u := tt.initial()
			err := tt.action(u)
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.verify != nil {
				tt.verify(t, u)
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

	tests := []userMutationTest{
		{
			name:      "Email not verified",
			initial:   newUserEmail,
			action: func(u *user.User) error {
				return u.LinkExternalAccount("g_123", user.ProviderGoogle, false)
			},
			expectedErr: user.ErrEmailNotVerified,
			verify: func(t *testing.T, u *user.User) {
				if u.ProviderID != nil {
					t.Errorf("the user state should not have mutated")
				}
			},
		},
		{
			name:      "Successful link to Google account with verified email",
			initial:   newUserEmail,
			action: func(u *user.User) error {
				return u.LinkExternalAccount("g_123", user.ProviderGoogle, true)
			},
			expectedErr: nil,
			verify: func(t *testing.T, u *user.User) {
				if u.Provider != user.ProviderGoogle {
					t.Errorf("expected provider to mutate to %v, got %v", user.ProviderGoogle, u.Provider)
				}
				if u.ProviderID == nil || *u.ProviderID != "g_123" {
					t.Errorf("expected google provider ID to be stored")
				}
			},
		},
		{
			name:      "Same provider and ID is idempotent",
			initial:   newUserGoogle,
			action: func(u *user.User) error {
				return u.LinkExternalAccount("g_123", user.ProviderGoogle, true)
			},
			expectedErr: nil,
			verify: func(t *testing.T, u *user.User) {
				if u.Provider != user.ProviderGoogle || u.ProviderID == nil || *u.ProviderID != "g_123" {
					t.Errorf("original state was lost after an idempotent link")
				}
			},
		},
		{
			name:    "Overwrite attempt with a different provider ID fails",
			initial: newUserGoogle,
			action: func(u *user.User) error {
				return u.LinkExternalAccount("g_999", user.ProviderGoogle, true)
			},
			expectedErr: user.ErrAccountAlreadyLinked,
			verify: func(t *testing.T, u *user.User) {
				if u.ProviderID == nil || *u.ProviderID != "g_123" {
					t.Errorf("original user ID was overwritten by a failed operation")
				}
			},
		},
	}

	runUserMutationTests(t, tests)
}
