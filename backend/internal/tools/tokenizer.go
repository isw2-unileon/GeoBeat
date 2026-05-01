package tools

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTTokenizer implements the Tokenizer interface using JSON Web Tokens (JWT) for authentication token generation and validation
type JWTTokenizer struct {
	secret []byte
}

// NewJWTTokenizer creates a new instance of JWTTokenizer with the provided secret key used for signing and validating tokens
func NewJWTTokenizer(secret string) *JWTTokenizer {
	return &JWTTokenizer{secret: []byte(secret)}
}

// GenerateToken generates a JWT token containing the user ID as a claim, with an expiration time of 24 hours. It returns the signed token string or an error if the token generation fails.
func (t *JWTTokenizer) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

// ValidateToken parses and validates a JWT token string, returning the user ID if the token is valid. It returns an error if the token is invalid or if the user ID claim is missing or malformed.
func (t *JWTTokenizer) ValidateToken(token string) (int, error) {
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return 0, err
	}
	if !parsed.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDValue, ok := claims["user_id"]
	if !ok {
		return 0, errors.New("user_id claim missing")
	}

	userIDFloat, ok := userIDValue.(float64)
	if !ok {
		return 0, errors.New("invalid user_id claim")
	}

	return int(userIDFloat), nil
}
