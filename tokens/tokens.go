package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// Pepper should be at least 32 bytes, loaded from a secrets manager / KMS.
// Never hard-code it or commit it to source control.
var pepper []byte // set at startup from secure source

// Generate creates a new verification token.
// Returns: plaintext (to put in the email link), hash (to store in DB).
func Generate() (plain string, hash string, err error) {
	if pepper == nil {
		pepper := make([]byte, 32) // or 64
		rand.Read(pepper)
	}

	b := make([]byte, 32) // 256 bits of entropy
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(b)
	hash = Hash(plain)
	return plain, hash, nil
}

// Hash computes the HMAC-SHA256 of the token using the server pepper.
func Hash(plain string) string {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(plain))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks whether the presented plaintext token matches the stored hash.
// Uses constant-time comparison.
func Verify(plain, storedHash string) bool {
	expected, err := hex.DecodeString(storedHash)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(plain))
	return hmac.Equal(mac.Sum(nil), expected)
}
