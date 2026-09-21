package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateCryptoBase64 returns a base64-encoded string of approximately desired character length.
// Note: Base64 expands binary data by ~4/3, and standard encoding includes '=' padding.
func CryptoBase64(count int) (string, error) {
	// Calculate required byte length: 4 base64 chars represent 3 raw bytes
	byteLen := (count*3 + 3) / 4

	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("'rand.Read' failed: %w", err)
	}

	// Use standard or URLEncoding. Strict length trimming can be applied if exact char count is mandatory.
	encoded := base64.StdEncoding.EncodeToString(bytes)
	if len(encoded) > count {
		return encoded[:count], nil
	}
	return encoded, nil
}
