package tools

import "golang.org/x/crypto/bcrypt"

type bCryptHasher struct{}

// NewBCryptHasher creates a new instance of bCryptHasher which implements the Hasher interface using bcrypt for password hashing and verification
func NewBCryptHasher() *bCryptHasher {
	return &bCryptHasher{}
}

// HashPassword takes a plaintext password and returns its bcrypt hash. It returns an error if the hashing process fails.
func (h *bCryptHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CompareHashAndPassword compares a bcrypt hashed password with its possible plaintext equivalent. It returns nil on success, or an error on failure.
func (h *bCryptHasher) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
