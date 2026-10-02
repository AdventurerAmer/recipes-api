package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
)

type HMACTokener struct {
	pepper []byte
}

func NewHMAC(pepper string) ports.Tokener {
	return &HMACTokener{
		pepper: []byte(pepper),
	}
}

func (tm *HMACTokener) Generate() (plain string, hash string, err error) {
	b := make([]byte, 32) // 256 bits of entropy
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(b)
	hash = tm.Hash(plain)
	return plain, hash, nil
}

func (tm *HMACTokener) Hash(plain string) string {
	mac := hmac.New(sha256.New, tm.pepper)
	mac.Write([]byte(plain))
	return hex.EncodeToString(mac.Sum(nil))
}

func (tm *HMACTokener) Verify(plain, hash string) bool {
	expected, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, tm.pepper)
	mac.Write([]byte(plain))
	return hmac.Equal(mac.Sum(nil), expected)
}
