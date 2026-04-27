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
	return nil
}

func (s *AuthService) LoginWithEmail(ctx context.Context, email, password string) (string, error) {
	return "", nil
}

func (s *AuthService) ProcessOAuthLogin(ctx context.Context, code string, provider OAuthProvider) (string, error) {
	return "", nil
}
