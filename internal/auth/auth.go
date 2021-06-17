package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword takes a password and returns the hash
func HashPassword(pwd []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
}

// PasswordMatches takes a hash and a password and validates if they match
func PasswordMatches(hash []byte, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)

	return err == nil
}
