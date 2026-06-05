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
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
	"github.com/isw2-unileon/GeoBeat/backend/internal/config"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
	"github.com/isw2-unileon/GeoBeat/backend/internal/server"
	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

type mockUserRepository struct {
	users map[string]*geouser.User
}

func (m *mockUserRepository) Save(ctx context.Context, u *geouser.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) Update(ctx context.Context, u *geouser.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*geouser.User, error) {
	if u, exists := m.users[email]; exists {
		return u, nil
	}
	return nil, geouser.ErrNotFound
}

type mockTokenizer struct{}

func (m *mockTokenizer) GenerateToken(userID uuid.UUID) (string, error) {
	return "mock-jwt-token", nil
}

func (m *mockTokenizer) ValidateToken(token string) (uuid.UUID, error) {
	if token == "mock-jwt-token" {
		return uuid.New(), nil
	}
	return uuid.Nil, errors.New("invalid token")
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

func (m *mockHasher) GenerateRawToken() (string, error) {
	return "raw-token", nil
}

func (m *mockHasher) HashToken(rawToken string) (string, error) {
	return "hash-" + rawToken, nil
}

type mockRefreshTokenRepository struct {
	tokens map[string]*authsession.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepository {
	return &mockRefreshTokenRepository{tokens: make(map[string]*authsession.RefreshToken)}
}

func (m *mockRefreshTokenRepository) Save(ctx context.Context, token *authsession.RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*authsession.RefreshToken, error) {
	if token, ok := m.tokens[tokenHash]; ok {
		return token, nil
	}
	return nil, authsession.ErrNotFound
}

func (m *mockRefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*authsession.RefreshToken, error) {
	for _, token := range m.tokens {
		if token.UserID == userID {
			return token, nil
		}
	}
	return nil, authsession.ErrNotFound
}

func (m *mockRefreshTokenRepository) Delete(ctx context.Context, tokenHash string) error {
	if _, ok := m.tokens[tokenHash]; !ok {
		return authsession.ErrNotFound
	}
	delete(m.tokens, tokenHash)
	return nil
}

type mockOAuthProvider struct {
	mockResponse *service.OAuthUserInfo
	authURL      string
	mockErr      error
}

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) GetProviderName() geouser.AuthProvider {
	return geouser.ProviderGoogle
}

func (m *mockOAuthProvider) GetUserInfo(ctx context.Context, code string) (*service.OAuthUserInfo, error) {
	return m.mockResponse, m.mockErr
}

var mockCfg = &config.Config{
	FrontendURL: "https://fake_frontend_url",
	RedirectURL: "fake_url",
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
				repo.users["existing@example.com"] = &geouser.User{
					Email:        "existing@example.com",
					UserName:     "existinguser",
					PasswordHash: &hashed,
					Provider:     geouser.ProviderEmail,
				}
			}

			var body bytes.Buffer
			if tt.requestBody != nil {
				if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
					t.Fatalf("failed to encode request body: %v", err)
				}
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
				userRepo.users["test@example.com"] = &geouser.User{
					ID:           uuid.New(),
					Email:        "test@example.com",
					UserName:     "testuser",
					PasswordHash: &hashed,
					Provider:     geouser.ProviderEmail,
				}
			case "invalid credentials":
				hashed := "hashed-ValidPass123!"
				userRepo.users["test@example.com"] = &geouser.User{
					Email:        "test@example.com",
					UserName:     "testuser",
					PasswordHash: &hashed,
					Provider:     geouser.ProviderEmail,
				}
			case "OAuth-only account":
				userRepo.users["oauth@example.com"] = &geouser.User{
					Email:    "oauth@example.com",
					UserName: "oauthuser",
					Provider: geouser.ProviderGoogle,
				}
			}

			var body bytes.Buffer
			if tt.requestBody != nil {
				if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
					t.Fatalf("failed to encode request body: %v", err)
				}
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
				validateRedirect(t, w)
			}
		})
	}
}

func validateRedirect(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

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
			expectedStatus: http.StatusTemporaryRedirect,
			expectedBody:   "<a href=\"https://fake_frontend_url\">Temporary Redirect</a>",
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
			expectedBody:   "",
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
				userRepo.users["oauth@example.com"] = &geouser.User{
					ID:       uuid.New(),
					Email:    "oauth@example.com",
					UserName: "oauthuser",
					Provider: geouser.ProviderGoogle,
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

func TestHandleRefreshAndLogout(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{name: "refresh", path: "/api/auth/refresh"},
		{name: "logout", path: "/api/auth/logout"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux, _, refreshRepo := newTestAuthServerWithRefresh(t)
			if err := refreshRepo.Save(context.Background(), authsession.NewRefreshToken(uuid.New(), "hash-raw-token")); err != nil {
				t.Fatalf("setup Save failed: %v", err)
			}

			req := httptest.NewRequest("POST", tc.path, nil)
			req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-token", Path: "/"})
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			if tc.name == "refresh" {
				if !strings.Contains(w.Body.String(), `"token":"mock-jwt-token"`) {
					t.Fatalf("expected access token in response, got %s", w.Body.String())
				}
			} else {
				if !strings.Contains(w.Header().Get("Set-Cookie"), "refresh_token=") {
					t.Fatalf("expected refresh_token cookie to be cleared, got %s", w.Header().Get("Set-Cookie"))
				}
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
			userRepo := &mockUserRepository{users: make(map[string]*geouser.User)}
			hasher := &mockHasher{}
			tokenizer := &mockTokenizer{}
			authSvc := service.NewAuthService(userRepo, newMockRefreshTokenRepo(), tokenizer, hasher, nil)
			handler := server.NewAuthHandler(authSvc, nil, mockCfg)

			// Mock next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("handler called")); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
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

	repo := &mockUserRepository{users: make(map[string]*geouser.User)}
	hasher := &mockHasher{}
	tokenizer := &mockTokenizer{}
	provider := &mockOAuthProvider{
		mockResponse: &service.OAuthUserInfo{
			Email:      "oauth@example.com",
			UserName:   "oauthuser",
			ProviderID: "provider-id",
		},
	}
	svcProviders := map[geouser.AuthProvider]service.OAuthProvider{
		geouser.ProviderGoogle: provider,
	}
	hdlProviders := map[geouser.AuthProvider]server.OAuthProvider{
		geouser.ProviderGoogle: provider,
	}

	authSvc := service.NewAuthService(repo, newMockRefreshTokenRepo(), tokenizer, hasher, svcProviders)
	handler := server.NewAuthHandler(authSvc, hdlProviders, mockCfg)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux, repo
}

func newTestAuthServerWithRefresh(t *testing.T) (*http.ServeMux, *mockUserRepository, *mockRefreshTokenRepository) {
	t.Helper()

	repo := &mockUserRepository{users: make(map[string]*geouser.User)}
	refreshRepo := newMockRefreshTokenRepo()
	hasher := &mockHasher{}
	tokenizer := &mockTokenizer{}
	provider := &mockOAuthProvider{
		mockResponse: &service.OAuthUserInfo{
			Email:      "oauth@example.com",
			UserName:   "oauthuser",
			ProviderID: "provider-id",
		},
	}
	svcProviders := map[geouser.AuthProvider]service.OAuthProvider{
		geouser.ProviderGoogle: provider,
	}
	hdlProviders := map[geouser.AuthProvider]server.OAuthProvider{
		geouser.ProviderGoogle: provider,
	}

	authSvc := service.NewAuthService(repo, refreshRepo, tokenizer, hasher, svcProviders)
	handler := server.NewAuthHandler(authSvc, hdlProviders, mockCfg)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux, repo, refreshRepo
}
