package desktop

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSecret creates a cryptographically secure random secret
// for per-launch authentication. The secret is 32 bytes encoded as
// URL-safe base64, resulting in a 43-character string.
func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
