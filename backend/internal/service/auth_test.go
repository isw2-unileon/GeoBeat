package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type mockUserRepository struct {
	users         map[string]*user.User
	databaseError error
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*user.User), databaseError: errors.New("database error")}
}

func (m *mockUserRepository) FindByEmail(email string) (*user.User, error) {
	u, exists := m.users[email]
	if email == "extrange@mail.com" {
		return nil, m.databaseError
	}
	if !exists {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) Save(u *user.User) error {
	if u.Email == "dbError@error.com" {
		return m.databaseError
	}
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

type mockHasher struct {
	hassingError bool
}

func newMockHasher(hassingError bool) *mockHasher {
	return &mockHasher{hassingError: hassingError}
}

func (m *mockHasher) HashPassword(password string) (string, error) {
	if m.hassingError {
		return "", errors.New("hashing error")
	}
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
		name         string
		setupRepo    func(*mockUserRepository)
		hassingError bool
		email        string
		userName     string
		password     string
		expectedErr  error
		checkUser    func(*testing.T, *mockUserRepository)
	}{
		{
			name: "successfully registers a new user",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "new@mail.com",
			userName:    "Juan",
			password:    "Password_123",
			expectedErr: nil,
			checkUser: func(t *testing.T, r *mockUserRepository) {
				u, exists := r.users["new@mail.com"]
				if !exists {
					t.Errorf("user was not saved in the database")
					return
				}
				if u.PasswordHash == nil || *u.PasswordHash != "hash_Password_123" {
					t.Errorf("password was not hashed correctly")
				}
			},
		},
		{
			name: "fails if missing field",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "",
			userName:    "Juan",
			password:    "Password_123",
			expectedErr: user.ErrEmptyEmailOrUsername,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "fails if email is already in use",
			setupRepo: func(r *mockUserRepository) {
				r.users["used@mail.com"], _ = user.NewUserFromEmail("used@mail.com", "Pedro", "hash")
			},
			email:       "used@mail.com",
			userName:    "Intruder",
			password:    "Password_123",
			expectedErr: service.ErrUserAlreadyExists,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles database error on email lookup",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "extrange@mail.com",
			userName:    "Juan",
			password:    "Password_123",
			expectedErr: service.ErrUserCreationFailed,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles database error on user save",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "dbError@error.com",
			userName:    "Juan",
			password:    "Password_123",
			expectedErr: service.ErrUserCreationFailed,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles hashing error",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			hassingError: true,
			email:        "new@mail.com",
			userName:     "Juan",
			password:     "Password_123",
			expectedErr:  service.ErrUserCreationFailed,
			checkUser:    func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "fails if password is too short",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "new@mail.com",
			userName:    "Juan",
			password:    "short",
			expectedErr: service.ErrPasswordTooWeak,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "fails if password lacks complexity",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "new@mail.com",
			userName:    "Juan",
			password:    "NoComplexity123",
			expectedErr: service.ErrPasswordTooWeak,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			tt.setupRepo(repo)
			svc := service.NewAuthService(repo, &mockTokenizer{}, newMockHasher(tt.hassingError))

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
		name         string
		setupRepo    func(*mockUserRepository)
		hassingError bool
		email        string
		password     string
		expectedErr  error
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
			name: "fails if user not found",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "nonexistent@mail.com",
			password:    "password123",
			expectedErr: service.ErrInvalidCredentials,
		},
		{
			name: "handles database error on user lookup",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			email:       "extrange@mail.com",
			password:    "password123",
			expectedErr: service.ErrUserLoginFailed,
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
				u, _ := user.NewUserExternal("oauth@mail.com", "OAuth User", "g_123", user.ProviderGoogle, true)
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
			svc := service.NewAuthService(repo, &mockTokenizer{}, newMockHasher(tt.hassingError))

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
		hassingError bool
		code         string
		providerMock *mockOAuthProvider
		expectedErr  error
		checkUser    func(*testing.T, *mockUserRepository)
	}{
		{
			name: "successful OAuth registration and login",
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
		{
			name: "successful OAuth login with existing user",
			setupRepo: func(r *mockUserRepository) {
				u, _ := user.NewUserExternal("oauth@mail.com", "OAuth User", "g_123", user.ProviderGoogle, true)
				r.users["oauth@mail.com"] = u
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
					t.Fatalf("user not found")
				}
				if u.Provider != user.ProviderGoogle || *u.ProviderID != "g_123" {
					t.Errorf("incorrect provider data")
				}
			},
		},
		{
			name: "fails if providerID does not match existing user",
			setupRepo: func(r *mockUserRepository) {
				u, _ := user.NewUserExternal("oauth@mail.com", "OAuth User", "g_123", user.ProviderGoogle, true)
				r.users["oauth@mail.com"] = u
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "oauth@mail.com",
					UserName:      "OAuth User",
					ProviderID:    "g_456",
					EmailVerified: true,
				},
			},
			expectedErr: service.ErrInvalidCredentials,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "successfully links OAuth account to existing email-only user",
			setupRepo: func(r *mockUserRepository) {
				u, _ := user.NewUserFromEmail("oauth@mail.com", "Email User", "hash_password123")
				r.users["oauth@mail.com"] = u
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
					t.Fatalf("user not found")
				}
				if u.Provider != user.ProviderGoogle || *u.ProviderID != "g_123" {
					t.Errorf("incorrect provider data")
				}
			},
		},
		{
			name: "fails if email from OAuth provider is not verified",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "oauth@mail.com",
					UserName:      "OAuth User",
					ProviderID:    "g_123",
					EmailVerified: false,
				},
			},
			expectedErr: user.ErrEmailNotVerified,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "fails linking if email from OAuth provider is not verified",
			setupRepo: func(r *mockUserRepository) {
				u, _ := user.NewUserFromEmail("oauth@mail.com", "Email User", "hash_password123")
				r.users["oauth@mail.com"] = u
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "oauth@mail.com",
					UserName:      "OAuth User",
					ProviderID:    "g_123",
					EmailVerified: false,
				},
			},
			expectedErr: user.ErrEmailNotVerified,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "unverified email from OAuth provider, fails linking",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "",
					UserName:      "OAuth User",
					ProviderID:    "g_123",
					EmailVerified: false,
				},
			},
			expectedErr: user.ErrEmailNotVerified,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles error from OAuth provider",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockErr: errors.New("provider error"),
			},
			expectedErr: service.ErrUserLoginFailed,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles database error on user lookup",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "extrange@mail.com",
					UserName:      "Extrange User",
					ProviderID:    "e_123",
					EmailVerified: true,
				},
			},
			expectedErr: service.ErrUserLoginFailed,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
		{
			name: "handles database error on user save",
			setupRepo: func(r *mockUserRepository) {
				// No setup needed
			},
			code: "code_123",
			providerMock: &mockOAuthProvider{
				mockResponse: &service.OAuthUserInfo{
					Email:         "dbError@error.com",
					UserName:      "New User",
					ProviderID:    "n_123",
					EmailVerified: true,
				},
			},
			expectedErr: service.ErrUserCreationFailed,
			checkUser:   func(t *testing.T, r *mockUserRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			tt.setupRepo(repo)
			svc := service.NewAuthService(repo, &mockTokenizer{}, newMockHasher(tt.hassingError))

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
