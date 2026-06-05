package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/isw2-unileon/GeoBeat/backend/internal/authsession"
	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
)

// Tokenizer defines the interface for generating and validating authentication tokens
type Tokenizer interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ValidateToken(token string) (uuid.UUID, error)
}

// Hasher defines the interface for password hashing and verification
type Hasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
	GenerateRawToken() (string, error)
	HashToken(rawToken string) (string, error)
}

// OAuthUserInfo represents the user information retrieved from an OAuth provider
type OAuthUserInfo struct {
	Email         string
	UserName      string
	ProviderID    string
	EmailVerified bool
}

// OAuthProvider defines the interface for interacting with an OAuth provider
type OAuthProvider interface {
	GetProviderName() geouser.AuthProvider
	GetUserInfo(ctx context.Context, code string) (*OAuthUserInfo, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*geouser.User, error)
	Save(ctx context.Context, u *geouser.User) error
	Update(ctx context.Context, u *geouser.User) error
}

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	Save(ctx context.Context, token *authsession.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*authsession.RefreshToken, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*authsession.RefreshToken, error)
	Delete(ctx context.Context, tokenHash string) error
}

var (
	// ErrUserCreationFailed indicates a failure during user creation
	ErrUserCreationFailed = errors.New("failed to create user")
	// ErrUserLoginFailed indicates a failure during user login
	ErrUserLoginFailed = errors.New("failed to login user")
	// ErrPasswordTooWeak indicates that the provided password does not meet security requirements
	ErrPasswordTooWeak = errors.New("password does not meet security requirements")
	// ErrUserAlreadyExists indicates that a user with the provided email already exists
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrOAuthOnlyAccount indicates that the email is associated to an account that only supports OAuth login
	ErrOAuthOnlyAccount = errors.New("email is associated to an account that only supports OAuth login")
	// ErrAlreadyLoggedIn indicates that the user is already logged in
	ErrAlreadyLoggedIn = errors.New("user is already logged in")
	// ErrRefreshingToken indicates that there was an error refreshing the token
	ErrRefreshingToken = errors.New("error refreshing token")
	// ErrLoggingOut indicates that there was an error during logout
	ErrLoggingOut = errors.New("error during logout")
)

// AuthService provides methods for user authentication and registration
type AuthService struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	tokenizer        Tokenizer
	hasher           Hasher
	oauthProviders   map[geouser.AuthProvider]OAuthProvider
	logger           *slog.Logger
}

// NewAuthService creates a new instance of AuthService with the provided dependencies
func NewAuthService(userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, tokenizer Tokenizer, hasher Hasher, oauthProviders map[geouser.AuthProvider]OAuthProvider) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenizer:        tokenizer,
		hasher:           hasher,
		oauthProviders:   oauthProviders,
		logger:           slog.Default(),
	}
}

// RegisterWithEmail registers a new user using email and password
func (s *AuthService) RegisterWithEmail(ctx context.Context, email, userName, password string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, geouser.ErrNotFound) {
		s.logger.Error("error checking existing user", "email", email, "error", err)
		return ErrUserCreationFailed
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	if !checkEmailFormat(email) {
		return ErrInvalidCredentials
	}

	if !ensurePasswordSecure(password) {
		return ErrPasswordTooWeak
	}

	hashedPassword, err := s.hasher.HashPassword(password)
	if err != nil {
		s.logger.Error("error hashing password", "password", password, "email", email, "error", err)
		return ErrUserCreationFailed
	}

	newUser, err := geouser.NewUserFromEmail(email, userName, hashedPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.Save(ctx, newUser)
	if err != nil {
		s.logger.Error("error saving new user", "email", email, "error", err)
		return ErrUserCreationFailed
	}

	return nil
}

func checkEmailFormat(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

func ensurePasswordSecure(password string) bool {
	if len(password) < 8 {
		return false
	}
	if !containsUpperCase(password) || !containsNumber(password) || !containsSpecialChar(password) {
		return false
	}
	return true
}

func containsUpperCase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()-_=+[]{}|;:'\",.<>/?`~"
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			return true
		}
	}
	return false
}

func (s *AuthService) checkExistingRefreshToken(ctx context.Context, userID uuid.UUID) (bool, error) {
	oldRefreshToken, err := s.refreshTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, authsession.ErrNotFound) {
			return false, nil
		}
		s.logger.Error("error retrieving refresh token for user", "userID", userID, "error", err)
		return false, ErrUserLoginFailed
	}
	return oldRefreshToken != nil, nil
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rawToken, err := s.hasher.GenerateRawToken()
	if err != nil {
		s.logger.Error("error generating refresh token", "userID", userID, "error", err)
		return "", ErrUserLoginFailed
	}

	tokenHash, err := s.hasher.HashToken(rawToken)
	if err != nil {
		s.logger.Error("error hashing refresh token", "userID", userID, "error", err)
		return "", ErrUserLoginFailed
	}

	refreshToken := authsession.NewRefreshToken(userID, tokenHash)
	err = s.refreshTokenRepo.Save(ctx, refreshToken)
	if err != nil {
		s.logger.Error("error saving refresh token", "userID", userID, "error", err)
		return "", ErrUserLoginFailed
	}

	return rawToken, nil
}

// LoginWithEmail authenticates a user using email and password, returning a token if successful
func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (string, string, error) {
	storedUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, geouser.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		s.logger.Error("error retrieving user", "email", email, "error", err)
		return "", "", ErrUserLoginFailed
	}

	if storedUser.Provider != geouser.ProviderEmail {
		return "", "", ErrOAuthOnlyAccount
	}

	err = s.hasher.CompareHashAndPassword(*storedUser.PasswordHash, password)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	exists, err := s.checkExistingRefreshToken(ctx, storedUser.ID)
	if err != nil {
		return "", "", err
	}
	if exists {
		return "", "", ErrAlreadyLoggedIn
	}

	rawToken, err := s.generateRefreshToken(ctx, storedUser.ID)
	if err != nil {
		return "", "", err
	}

	authToken, err := s.tokenizer.GenerateToken(storedUser.ID)
	if err != nil {
		s.logger.Error("error generating auth token", "userID", storedUser.ID, "error", err)
		return "", "", ErrUserLoginFailed
	}

	return authToken, rawToken, nil
}

// ProcessOAuthLogin processes an OAuth login flow, returning a token if successful
func (s *AuthService) ProcessOAuthLogin(ctx context.Context, code string, provider geouser.AuthProvider) (string, error) {
	oauthProvider := s.oauthProviders[provider]

	userInfo, err := oauthProvider.GetUserInfo(ctx, code)
	if err != nil {
		s.logger.Error("error getting user info from OAuth provider", "error", err)
		return "", ErrUserLoginFailed
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, geouser.ErrNotFound) {
		s.logger.Error("error checking existing user", "email", userInfo.Email, "error", err)
		return "", ErrUserLoginFailed
	}

	if existingUser != nil {
		return s.processOAuthExistingUser(ctx, existingUser, userInfo, oauthProvider)
	}

	return s.createAndSaveUserFromOAuth(ctx, userInfo, oauthProvider)
}

func (s *AuthService) processOAuthExistingUser(ctx context.Context, existingUser *geouser.User, userInfo *OAuthUserInfo, oauthProvider OAuthProvider) (string, error) {
	if existingUser.Provider == oauthProvider.GetProviderName() {
		if existingUser.ProviderID == nil || *existingUser.ProviderID != userInfo.ProviderID {
			s.logger.Error("provider ID mismatch for existing user", "storedProviderID", existingUser.ProviderID, "oauthProviderID", userInfo.ProviderID)
			return "", ErrInvalidCredentials
		}
		// Reissue the refresh token on every provider login to keep flow consistent
		s.logger.Info("Oauth login processed", "Provider", oauthProvider.GetProviderName(), "ID", existingUser.ID)
		return s.generateRefreshToken(ctx, existingUser.ID)
	}

	// Link external account for users that registered via other methods
	if err := existingUser.LinkExternalAccount(userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified); err != nil {
		return "", err
	}
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		s.logger.Error("error updating existing user", "email", userInfo.Email, "error", err)
		return "", ErrUserLoginFailed
	}
	return s.generateRefreshToken(ctx, existingUser.ID)
}

func (s *AuthService) createAndSaveUserFromOAuth(ctx context.Context, userInfo *OAuthUserInfo, oauthProvider OAuthProvider) (string, error) {
	newUser, err := geouser.NewUserExternal(userInfo.Email, userInfo.UserName, userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified)
	if err != nil {
		return "", err
	}
	if err := s.userRepo.Save(ctx, newUser); err != nil {
		s.logger.Error("error saving new user", "email", userInfo.Email, "error", err)
		return "", ErrUserCreationFailed
	}
	return s.generateRefreshToken(ctx, newUser.ID)
}

// ValidateToken validates an authentication token and returns the associated user ID if valid
func (s *AuthService) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	return s.tokenizer.ValidateToken(token)
}

// RefreshToken validates the provided refresh token and issues a new authentication token if valid
func (s *AuthService) RefreshToken(ctx context.Context, rawToken string) (string, error) {
	tokenHash, err := s.hasher.HashToken(rawToken)
	if err != nil {
		s.logger.Error("error hashing refresh token", "rawToken", rawToken, "error", err)
		return "", ErrRefreshingToken
	}

	storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, authsession.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		s.logger.Error("error retrieving refresh token", "tokenHash", tokenHash, "error", err)
		return "", ErrRefreshingToken
	}

	if storedToken.IsExpired() {
		return "", authsession.ErrExpired
	}

	authToken, err := s.tokenizer.GenerateToken(storedToken.UserID)
	if err != nil {
		s.logger.Error("error generating auth token", "userID", storedToken.UserID, "error", err)
		return "", ErrRefreshingToken
	}

	return authToken, nil
}

// Logout invalidates the provided refresh token, effectively logging the user out
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash, err := s.hasher.HashToken(rawToken)
	if err != nil {
		s.logger.Error("error hashing refresh token for logout", "rawToken", rawToken, "error", err)
		return ErrLoggingOut
	}

	err = s.refreshTokenRepo.Delete(ctx, tokenHash)
	if err != nil {
		s.logger.Error("error deleting refresh token during logout", "tokenHash", tokenHash, "error", err)
		return ErrLoggingOut
	}

	return nil
}
