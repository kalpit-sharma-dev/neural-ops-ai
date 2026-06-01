package security

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	t.Setenv("NEURALOPS_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key))

	encrypted, err := EncryptSecret("oidc-client-secret")
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	plain, err := DecryptSecret(encrypted)
	require.NoError(t, err)
	require.Equal(t, "oidc-client-secret", plain)
}

func TestEncryptMissingKey(t *testing.T) {
	_ = os.Unsetenv("NEURALOPS_ENCRYPTION_KEY")
	_, err := EncryptSecret("secret")
	require.Error(t, err)
}
