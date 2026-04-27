package tools

import "golang.org/x/crypto/bcrypt"

type bCryptHasher struct{}

func NewBCryptHasher() *bCryptHasher {
	return &bCryptHasher{}
}

func (h *bCryptHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (h *bCryptHasher) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
