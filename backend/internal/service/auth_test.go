package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type mockUserRepository struct {
	users map[string]*user.User
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*user.User)}
}

func (m *mockUserRepository) FindByEmail(email string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) Save(u *user.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) Update(u *user.User) error {
	m.users[u.Email] = u
	return nil
}

type mockTokenizer struct{}

func (m *mockTokenizer) GenerateToken(userID int) (string, error) {
	return "jwt_simulated", nil
}

type mockHasher struct{}

func (m *mockHasher) HashPassword(password string) (string, error) {
	return "hash_" + password, nil
}

func (m *mockHasher) CompareHashAndPassword(hash, password string) error {
	if hash == "hash_"+password {
		return nil
	}
	return errors.New("hash mismatch")
}

type mockOAuthProvider struct {
	mockResponse *service.OAuthUserInfo
	mockErr      error
}

func (m *mockOAuthProvider) GetProviderName() user.AuthProvider {
	return user.ProviderGoogle
}

func (m *mockOAuthProvider) GetUserInfo(ctx context.Context, code string) (*service.OAuthUserInfo, error) {
	return m.mockResponse, m.mockErr
}

func TestRegisterWithEmail(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*mockUserRepository)
		email       string
		userName    string
		password    string
		expectedErr error
		checkUser   func(*testing.T, *mockUserRepository)
	}{
		{
			name: "successfully registers a new user",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "new@mail.com",
			userName:    "Juan",
			password:    "password123",
			expectedErr: nil,
			checkUser: func(t *testing.T, r *mockUserRepository) {
				u, exists := r.users["new@mail.com"]
				if !exists {
					t.Errorf("user was not saved in the database")
				}
				if u.PasswordHash == nil || *u.PasswordHash != "hash_password123" {
					t.Errorf("password was not hashed correctly")
				}
			},
		},
		{
			name: "fails if email is already in use",
			setupRepo: func(r *mockUserRepository) {
				r.users["used@mail.com"], _ = user.NewUserFromEmail("used@mail.com", "Pedro", "hash")
			},
			email:       "used@mail.com",
			userName:    "Intruder",
			password:    "password123",
			expectedErr: service.ErrUserAlreadyExists,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			tt.setupRepo(repo)
			svc := service.NewAuthService(repo, &mockTokenizer{}, &mockHasher{})

			err := svc.RegisterWithEmail(context.Background(), tt.email, tt.userName, tt.password)
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			tt.checkUser(t, repo)
		})
	}
}

func TestLoginWithEmail(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*mockUserRepository)
		email       string
		password    string
		expectedErr error
	}{
		{
			name: "successful login",
			setupRepo: func(r *mockUserRepository) {
				r.users["test@mail.com"], _ = user.NewUserFromEmail("test@mail.com", "Test", "hash_password123")
			},
			email:       "test@mail.com",
			password:    "password123",
			expectedErr: nil,
		},
		{
			name: "fails with incorrect password",
			setupRepo: func(r *mockUserRepository) {
				r.users["test@mail.com"], _ = user.NewUserFromEmail("test@mail.com", "Test", "hash_password123")
			},
			email:       "test@mail.com",
			password:    "wrong_password",
			expectedErr: service.ErrInvalidCredentials,
		},
		{
			name: "fails if OAuth-only account",
			setupRepo: func(r *mockUserRepository) {
				u, _ := user.NewUserExternal("oauth@mail.com", "G", "123", user.ProviderGoogle, true)
				r.users["oauth@mail.com"] = u
			},
			email:       "oauth@mail.com",
			password:    "anything",
			expectedErr: service.ErrOAuthOnlyAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			tt.setupRepo(repo)
			svc := service.NewAuthService(repo, &mockTokenizer{}, &mockHasher{})

			token, err := svc.LoginWithEmail(context.Background(), tt.email, tt.password)

			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if err == nil && token != "jwt_simulated" {
				t.Errorf("unexpected token: %s", token)
			}
		})
	}
}

func TestProcessOAuthLogin(t *testing.T) {
	tests := []struct {
		name         string
		setupRepo    func(*mockUserRepository)
		code         string
		providerMock *mockOAuthProvider
		expectedErr  error
		checkUser    func(*testing.T, *mockUserRepository)
	}{
		{
			name: "successful OAuth login",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "oauth@mail.com",
					UserName:      "OAuth User",
					ProviderID:    "g_123",
					EmailVerified: true,
				},
			},
			expectedErr: nil,
			checkUser: func(t *testing.T, r *mockUserRepository) {
				u, exists := r.users["oauth@mail.com"]
				if !exists {
					t.Fatalf("user not saved")
				}
				if u.Provider != user.ProviderGoogle || *u.ProviderID != "g_123" {
					t.Errorf("incorrect provider data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			tt.setupRepo(repo)
			svc := service.NewAuthService(repo, &mockTokenizer{}, &mockHasher{})

			token, err := svc.ProcessOAuthLogin(context.Background(), tt.code, tt.providerMock)
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
			if err == nil && token != "jwt_simulated" {
				t.Errorf("unexpected token: %s", token)
			}
			tt.checkUser(t, repo)
		})
	}
}
