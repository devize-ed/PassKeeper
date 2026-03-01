// Package hash provides bcrypt password hashing and verification utilities.
package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against a hash.
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
