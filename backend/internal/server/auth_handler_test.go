package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type mockUserRepository struct {
	users map[string]*user.User
}

func (m *mockUserRepository) Save(u *user.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) Update(u *user.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*user.User, error) {
	if u, exists := m.users[email]; exists {
		return u, nil
	}
	return nil, user.ErrNotFound
}

type mockTokenizer struct{}

func (m *mockTokenizer) GenerateToken(userID int) (string, error) {
	return "mock-jwt-token", nil
}

func (m *mockTokenizer) ValidateToken(token string) (int, error) {
	if token == "mock-jwt-token" {
		return 1, nil
	}
	return 0, errors.New("invalid token")
}

type mockHasher struct{}

func (m *mockHasher) HashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (m *mockHasher) CompareHashAndPassword(hash, password string) error {
	if hash == "hashed-"+password {
		return nil
	}
	return errors.New("password does not match")
}

type mockOAuthProvider struct {
	mockResponse *service.OAuthUserInfo
	authURL      string
	mockErr      error
}

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) GetProviderName() user.AuthProvider {
	return user.ProviderGoogle
}

func (m *mockOAuthProvider) GetUserInfo(ctx context.Context, code string) (*service.OAuthUserInfo, error) {
	return m.mockResponse, m.mockErr
}

func TestHandleRegister(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"username": "testuser",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "",
		},
		{
			name: "user already exists",
			requestBody: map[string]string{
				"email":    "existing@example.com",
				"username": "existinguser",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"failed to create user"}`,
		},
		{
			name: "invalid email format",
			requestBody: map[string]string{
				"email":    "invalid-email",
				"username": "testuser",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid email or password"}`,
		},
		{
			name: "weak password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"username": "testuser",
				"password": "weak",
			},

			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"password does not meet security requirements"}`,
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, repo := newTestAuthServer(t)
			if tt.name == "user already exists" {
				hashed := "hashed-ValidPass123!"
				repo.users["existing@example.com"] = &user.User{
					Email:        "existing@example.com",
					UserName:     "existinguser",
					PasswordHash: &hashed,
					Provider:     user.ProviderEmail,
				}
			}

			var body bytes.Buffer
			if tt.requestBody != nil {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/auth/register", &body)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"mock-jwt-token"}`,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpass",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid email or password"}`,
		},
		{
			name: "OAuth-only account",
			requestBody: map[string]string{
				"email":    "oauth@example.com",
				"password": "password",
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"failed to login user"}`,
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, userRepo := newTestAuthServer(t)
			switch tt.name {
			case "successful login":
				hashed := "hashed-ValidPass123!"
				userRepo.users["test@example.com"] = &user.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					UserName:     "testuser",
					PasswordHash: &hashed,
					Provider:     user.ProviderEmail,
				}
			case "invalid credentials":
				hashed := "hashed-ValidPass123!"
				userRepo.users["test@example.com"] = &user.User{
					Email:        "test@example.com",
					UserName:     "testuser",
					PasswordHash: &hashed,
					Provider:     user.ProviderEmail,
				}
			case "OAuth-only account":
				userRepo.users["oauth@example.com"] = &user.User{
					Email:    "oauth@example.com",
					UserName: "oauthuser",
					Provider: user.ProviderGoogle,
				}
			}

			var body bytes.Buffer
			if tt.requestBody != nil {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/auth/login", &body)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleOAuthLogin(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		checkRedirect  bool
	}{
		{
			name:           "successful OAuth login redirect",
			endpoint:       "/api/auth/login/google",
			expectedStatus: http.StatusFound,
			checkRedirect:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, _ := newTestAuthServer(t)

			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkRedirect {
				location := w.Header().Get("Location")
				if location == "" {
					t.Errorf("expected redirect location, got empty")
				}
				// Check for state cookie
				cookies := w.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if strings.HasPrefix(cookie.Name, "oauth_state_") {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected oauth state cookie to be set")
				}
			}
		})
	}
}

func TestHandleOAuthRedirect(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedBody   string
		setupCookies   func(*http.Request)
	}{
		{
			name: "successful OAuth callback",
			queryParams: map[string]string{
				"code":  "auth-code",
				"state": "valid-state",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"token"`,
			setupCookies: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "valid-state", Path: "/"})
			},
		},
		{
			name: "invalid state",
			queryParams: map[string]string{
				"code":  "auth-code",
				"state": "invalid-state",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
			setupCookies: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "valid-state", Path: "/"})
			},
		},
		{
			name: "missing code",
			queryParams: map[string]string{
				"state": "valid-state",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
			setupCookies: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "valid-state", Path: "/"})
			},
		},
		{
			name: "OAuth provider error",
			queryParams: map[string]string{
				"error": "access_denied",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
			setupCookies:   func(req *http.Request) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, userRepo := newTestAuthServer(t)

			// Setup test data for successful callback
			if tt.name == "successful OAuth callback" {
				userRepo.users["oauth@example.com"] = &user.User{
					ID:       uuid.New(),
					Email:    "oauth@example.com",
					UserName: "oauthuser",
					Provider: user.ProviderGoogle,
					ProviderID: func() *string {
						id := "provider-id"
						return &id
					}(),
				}
			}

			endpoint := "/api/auth/login/callback/google"
			req := httptest.NewRequest("GET", endpoint, nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			if tt.setupCookies != nil {
				tt.setupCookies(req)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer mock-jwt-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "handler called",
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"missing authorization header"}`,
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header"}`,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepository{users: make(map[string]*user.User)}
			hasher := &mockHasher{}
			tokenizer := &mockTokenizer{}
			authSvc := service.NewAuthService(userRepo, tokenizer, hasher, nil)
			handler := server.NewAuthHandler(authSvc, nil)

			// Mock next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("handler called"))
			})

			middleware := handler.AuthMiddleware(nextHandler)

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func newTestAuthServer(t *testing.T) (*http.ServeMux, *mockUserRepository) {
	t.Helper()

	repo := &mockUserRepository{users: make(map[string]*user.User)}
	hasher := &mockHasher{}
	tokenizer := &mockTokenizer{}
	provider := &mockOAuthProvider{
		mockResponse: &service.OAuthUserInfo{
			Email:      "oauth@example.com",
			UserName:   "oauthuser",
			ProviderID: "provider-id",
		},
	}
	svcProviders := map[user.AuthProvider]service.OAuthProvider{
		user.ProviderGoogle: provider,
	}
	hdlProviders := map[user.AuthProvider]server.OAuthProvider{
		user.ProviderGoogle: provider,
	}

	authSvc := service.NewAuthService(repo, tokenizer, hasher, svcProviders)
	handler := server.NewAuthHandler(authSvc, hdlProviders)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux, repo
}
