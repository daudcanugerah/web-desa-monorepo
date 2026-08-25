package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := Hash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	password := "mySecurePassword123"

	hash1, err := Hash(password)
	require.NoError(t, err)

	hash2, err := Hash(password)
	require.NoError(t, err)

	// Bcrypt generates different hashes for the same password due to salt
	assert.NotEqual(t, hash1, hash2)
}

func TestVerify_CorrectPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := Hash(password)
	require.NoError(t, err)

	result := Verify(password, hash)
	assert.True(t, result)
}

func TestVerify_IncorrectPassword(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword"

	hash, err := Hash(password)
	require.NoError(t, err)

	result := Verify(wrongPassword, hash)
	assert.False(t, result)
}

func TestVerify_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := Hash(password)
	require.NoError(t, err)

	result := Verify("", hash)
	assert.False(t, result)
}

func TestVerify_InvalidHash(t *testing.T) {
	password := "mySecurePassword123"
	invalidHash := "not-a-valid-hash"

	result := Verify(password, invalidHash)
	assert.False(t, result)
}

func TestHash_EmptyPassword(t *testing.T) {
	hash, err := Hash("")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Should be able to verify empty password
	result := Verify("", hash)
	assert.True(t, result)
}
