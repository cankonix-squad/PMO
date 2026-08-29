package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain-text password using bcrypt with cost 12.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPassword compares a plain-text password with a bcrypt hash.
// Returns nil if they match.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
