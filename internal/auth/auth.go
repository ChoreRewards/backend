package auth

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

// Clock interface to make testing easier
type clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type TokenManager struct {
	key   string
	clock clock
}

func NewTokenManager(key string) TokenManager {
	return TokenManager{
		key:   key,
		clock: realClock{},
	}
}

func (t TokenManager) CreateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	now := t.clock.Now()

	// See https://tools.ietf.org/html/rfc7519#section-4.1
	claims["authorized"] = true
	claims["username"] = username
	claims["exp"] = now.Add(time.Minute * 30).Unix()
	claims["iat"] = now.Unix()

	return token.SignedString([]byte(t.key))
}

func (t TokenManager) ValidateToken(token string) error {
	jwtToken, err := jwt.Parse(token, func(tkn *jwt.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid Signing Method")
		}

		return []byte(t.key), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch ve.Errors {
			case jwt.ValidationErrorMalformed:
				return errors.New("Token is malformed")
			case jwt.ValidationErrorUnverifiable:
				return errors.New("Token could not be verified because of signing problems")
			case jwt.ValidationErrorSignatureInvalid:
				return errors.New("Signature validation failed")
			case jwt.ValidationErrorExpired:
				return errors.New("Expired token")
			case jwt.ValidationErrorClaimsInvalid:
				return errors.New("Invalid Claims")
			default:
				return errors.Wrap(err, "Validation error")
			}
		}
		return errors.Wrap(err, "Error parsing token")
	}

	if _, ok := jwtToken.Claims.(jwt.MapClaims); !ok {
		return errors.New("Unable to map claims")
	}

	return nil
}

func (t TokenManager) ValidateAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	_, endpoint := path.Split(info.FullMethod)

	// Login route is unprotected
	if endpoint == "Login" {
		return handler(ctx, req)
	}

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("Unable to get metadata from context")
	}

	auth := meta.Get("Authorization")

	if len(auth) != 1 {
		return nil, errors.New("Authorization header is in the wrong format")
	}

	if err := t.ValidateToken(strings.TrimPrefix(auth[0], "Bearer ")); err != nil {
		return nil, errors.Wrap(err, "Invalid token")
	}

	return handler(ctx, req)
}
