package tools

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenizer struct {
	secret []byte
}

func NewJWTTokenizer(secret string) *JWTTokenizer {
	return &JWTTokenizer{secret: []byte(secret)}
}

func (t *JWTTokenizer) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}
