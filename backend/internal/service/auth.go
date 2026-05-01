package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type Tokenizer interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(token string) (int, error)
}

type Hasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}

type OAuthUserInfo struct {
	Email         string
	UserName      string
	ProviderID    string
	EmailVerified bool
}

type OAuthProvider interface {
	GetProviderName() user.AuthProvider
	GetUserInfo(ctx context.Context, code string) (*OAuthUserInfo, error)
}

type UserRepository interface {
	FindByEmail(email string) (*user.User, error)
	Save(u *user.User) error
	Update(u *user.User) error
}

var (
	ErrUserCreationFailed = errors.New("failed to create user")
	ErrUserLoginFailed    = errors.New("failed to login user")
	ErrPasswordTooWeak    = errors.New("password does not meet security requirements")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrOAuthOnlyAccount   = errors.New("email is associated to an account that only supports OAuth login")
)

type AuthService struct {
	userRepo       UserRepository
	tokenizer      Tokenizer
	hasher         Hasher
	oauthProviders map[user.AuthProvider]OAuthProvider
	logger         *slog.Logger
}

func NewAuthService(userRepo UserRepository, tokenizer Tokenizer, hasher Hasher, oauthProviders map[user.AuthProvider]OAuthProvider) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenizer:      tokenizer,
		hasher:         hasher,
		oauthProviders: oauthProviders,
		logger:         slog.Default(),
	}
}

func (s *AuthService) RegisterWithEmail(ctx context.Context, email, userName, password string) error {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
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

	newUser, err := user.NewUserFromEmail(email, userName, hashedPassword)
	if err != nil {
		return err
	}

	err = s.userRepo.Save(newUser)
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

func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (string, error) {
	storedUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		s.logger.Error("error retrieving user", "email", email, "error", err)
		return "", ErrUserLoginFailed
	}

	if storedUser.Provider != user.ProviderEmail {
		return "", ErrOAuthOnlyAccount
	}

	err = s.hasher.CompareHashAndPassword(*storedUser.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokenizer.GenerateToken(int(storedUser.ID.ID()))
}

func (s *AuthService) ProcessOAuthLogin(ctx context.Context, code string, provider user.AuthProvider) (string, error) {
	oauthProvider := s.oauthProviders[provider]

	userInfo, err := oauthProvider.GetUserInfo(ctx, code)
	if err != nil {
		s.logger.Error("error getting user info from OAuth provider", "error", err)
		return "", ErrUserLoginFailed
	}

	existingUser, err := s.userRepo.FindByEmail(userInfo.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		s.logger.Error("error checking existing user", "email", userInfo.Email, "error", err)
		return "", ErrUserLoginFailed
	}

	if existingUser != nil {
		if existingUser.Provider == oauthProvider.GetProviderName() {
			if *existingUser.ProviderID != userInfo.ProviderID {
				s.logger.Error("provider ID mismatch for existing user", "storedProviderID", *existingUser.ProviderID, "oauthProviderID", userInfo.ProviderID)
				return "", ErrInvalidCredentials
			}
			return s.tokenizer.GenerateToken(int(existingUser.ID.ID()))
		}
		// Currently, we only support google as an external provider
		// Therefore this code will not return ErrAccountAlreadyLinked
		// Just here for extensibility
		err = existingUser.LinkExternalAccount(userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified)
		if err != nil {
			return "", err
		}
		err = s.userRepo.Update(existingUser)
		if err != nil {
			s.logger.Error("error updating existing user", "email", userInfo.Email, "error", err)
			return "", ErrUserLoginFailed
		}
		return s.tokenizer.GenerateToken(int(existingUser.ID.ID()))
	}

	newUser, err := user.NewUserExternal(userInfo.Email, userInfo.UserName, userInfo.ProviderID, oauthProvider.GetProviderName(), userInfo.EmailVerified)
	if err != nil {
		return "", err
	}

	err = s.userRepo.Save(newUser)
	if err != nil {
		s.logger.Error("error saving new user", "email", userInfo.Email, "error", err)
		return "", ErrUserCreationFailed
	}

	return s.tokenizer.GenerateToken(int(newUser.ID.ID()))
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (int, error) {
	return s.tokenizer.ValidateToken(token)
}
