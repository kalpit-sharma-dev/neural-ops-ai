package security

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashAPIKey returns the SHA-256 hex digest of an API key for storage/lookup.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
