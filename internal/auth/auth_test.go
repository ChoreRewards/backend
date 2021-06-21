package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testClock struct {
	time time.Time
}

func (t testClock) Now() time.Time {
	return t.time
}

func TestHashPassword(t *testing.T) {
	password := []byte(`testPassword123`)

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	hash2, err := HashPassword(password)
	assert.NoError(t, err)

	t.Run("hash should be different when called twice", func(t *testing.T) {
		assert.NotEqual(t, hash, hash2)
	})
}

func TestPasswordMatches(t *testing.T) {
	password := []byte(`testPassword123`)

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	assert.True(t, PasswordMatches(hash, password))
	assert.False(t, PasswordMatches([]byte(`somethingdifferent`), password))
}

func TestValidateToken(t *testing.T) {
	t.Run("it should not return an error when the token is valid", func(t *testing.T) {
		tm := NewTokenManager("test-key")

		tkn, err := tm.CreateToken("test-user")
		assert.NoError(t, err)

		assert.NoError(t, tm.ValidateToken(tkn))
	})

	t.Run("it should error when the token is malformed", func(t *testing.T) {
		tm := NewTokenManager("test-key")

		assert.EqualError(t, tm.ValidateToken("aaaaa"), "Token is malformed")
	})

	t.Run("it should error when the token has expired", func(t *testing.T) {
		tm := TokenManager{
			key:   "test-key",
			clock: testClock{time: time.Now().Add(-time.Minute * 300)},
		}

		tkn, err := tm.CreateToken("test-user")
		assert.NoError(t, err)

		assert.EqualError(t, tm.ValidateToken(tkn), "Expired token")
	})
}
