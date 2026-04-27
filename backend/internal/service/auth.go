package service

import (
	"context"
	"errors"

	"github.com/isw2-unileon/GeoBeat/backend/internal/user"
)

type Tokenizer interface {
	GenerateToken(userID int) (string, error)
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
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidProvider    = user.ErrInvalidProvider
	ErrOAuthOnlyAccount   = errors.New("email is associated to an account that only supports OAuth login")
)

type AuthService struct {
	userRepo  UserRepository
	tokenizer Tokenizer
	hasher    Hasher
}

func NewAuthService(userRepo UserRepository, tokenizer Tokenizer, hasher Hasher) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenizer: tokenizer,
		hasher:    hasher,
	}
}

func (s *AuthService) RegisterWithEmail(ctx context.Context, email, userName, password string) error {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return err
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := s.hasher.HashPassword(password)
	if err != nil {
		return err
	}

	newUser, err := user.NewUserFromEmail(email, userName, hashedPassword)
	if err != nil {
		return err
	}

	return s.userRepo.Save(newUser)
}

func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (string, error) {
	storedUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if storedUser.Provider != user.ProviderEmail {
		return "", ErrOAuthOnlyAccount
	}

	if storedUser.PasswordHash == nil {
		return "", ErrInvalidCredentials
	}

	err = s.hasher.CompareHashAndPassword(*storedUser.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokenizer.GenerateToken(int(storedUser.ID.ID()))
}

func (s *AuthService) ProcessOAuthLogin(ctx context.Context, code string, provider OAuthProvider) (string, error) {
	userInfo, err := provider.GetUserInfo(ctx, code)
	if err != nil {
		return "", err
	}

	existingUser, err := s.userRepo.FindByEmail(userInfo.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return "", err
	}

	if existingUser != nil {
		if existingUser.Provider != provider.GetProviderName() {
			return "", user.ErrAccountAlreadyLinked
		}
		err = existingUser.LinkExternalAccount(userInfo.ProviderID, provider.GetProviderName(), userInfo.EmailVerified)
		if err != nil {
			return "", err
		}
		err = s.userRepo.Update(existingUser)
		if err != nil {
			return "", err
		}
		return s.tokenizer.GenerateToken(int(existingUser.ID.ID()))
	}

	newUser, err := user.NewUserExternal(userInfo.Email, userInfo.UserName, userInfo.ProviderID, provider.GetProviderName(), userInfo.EmailVerified)
	if err != nil {
		return "", err
	}

	err = s.userRepo.Save(newUser)
	if err != nil {
		return "", err
	}

	return s.tokenizer.GenerateToken(int(newUser.ID.ID()))
}
