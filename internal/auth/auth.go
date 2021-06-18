package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a password and returns the hash
func HashPassword(pwd []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
}

// PasswordMatches takes a hash and a password and validates if they match
func PasswordMatches(hash []byte, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)

	return err == nil
}

type TokenManager struct {
	key string
}

func NewTokenManager(key string) TokenManager {
	return TokenManager{
		key: key,
	}
}

func (t TokenManager) CreateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = username
	claims["expiry"] = time.Now().Add(time.Minute * 30).Unix()

	return token.SignedString([]byte(t.key))
}
