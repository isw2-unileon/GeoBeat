package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/isw2-unileon/GeoBeat/backend/internal/geouser"
)

// Tokenizer defines the interface for generating and validating authentication tokens
type Tokenizer interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(token string) (int, error)
}

// Hasher defines the interface for password hashing and verification
type Hasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
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
)

// AuthService provides methods for user authentication and registration
type AuthService struct {
	userRepo       UserRepository
	tokenizer      Tokenizer
	hasher         Hasher
	oauthProviders map[geouser.AuthProvider]OAuthProvider
	logger         *slog.Logger
}

// NewAuthService creates a new instance of AuthService with the provided dependencies
func NewAuthService(userRepo UserRepository, tokenizer Tokenizer, hasher Hasher, oauthProviders map[geouser.AuthProvider]OAuthProvider) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenizer:      tokenizer,
		hasher:         hasher,
		oauthProviders: oauthProviders,
		logger:         slog.Default(),
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

// LoginWithEmail authenticates a user using email and password, returning a token if successful
func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (string, error) {
	storedUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, geouser.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		s.logger.Error("error retrieving user", "email", email, "error", err)
		return "", ErrUserLoginFailed
	}

	if storedUser.Provider != geouser.ProviderEmail {
		return "", ErrOAuthOnlyAccount
	}

	err = s.hasher.CompareHashAndPassword(*storedUser.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokenizer.GenerateToken(int(storedUser.ID.ID()))
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
		if existingUser.Provider == oauthProvider.GetProviderName() {
			if *existingUser.ProviderID != userInfo.ProviderID {
				s.logger.Error("provider ID mismatch for existing user", "storedProviderID", *existingUser.ProviderID, "oauthProviderID", userInfo.ProviderID)
				return "", ErrInvalidCredentials
			}
			s.logger.Info("Oauth login processed", "Provider", oauthProvider.GetProviderName(), "ID", existingUser.ID)
			return s.tokenizer.GenerateToken(int(existingUser.ID.ID()))
		}
		// Currently, we only support google as an external provider
		// Therefore this code will not return ErrAccountAlreadyLinked
		// Just here for extensibility
		err = existingUser.LinkExternalAccount(userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified)
		if err != nil {
			return "", err
		}
		err = s.userRepo.Update(ctx, existingUser)
		if err != nil {
			s.logger.Error("error updating existing user", "email", userInfo.Email, "error", err)
			return "", ErrUserLoginFailed
		}
		return s.tokenizer.GenerateToken(int(existingUser.ID.ID()))
	}

	newUser, err := geouser.NewUserExternal(userInfo.Email, userInfo.UserName, userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified)
	if err != nil {
		return "", err
	}

	err = s.userRepo.Save(ctx, newUser)
	if err != nil {
		s.logger.Error("error saving new user", "email", userInfo.Email, "error", err)
		return "", ErrUserCreationFailed
	}

	return s.tokenizer.GenerateToken(int(newUser.ID.ID()))
}

// ValidateToken validates an authentication token and returns the associated user ID if valid
func (s *AuthService) ValidateToken(ctx context.Context, token string) (int, error) {
	return s.tokenizer.ValidateToken(token)
}
