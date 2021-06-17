package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
