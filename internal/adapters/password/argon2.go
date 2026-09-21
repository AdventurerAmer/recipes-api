package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"golang.org/x/crypto/argon2"
)

// Recommended parameters (RFC 9106)
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

type argon2PasswordMgr struct {
}

func NewArgon2() ports.PasswordManager {
	return &argon2PasswordMgr{}
}

func (mgr *argon2PasswordMgr) Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)
	return encoded, nil
}

func (mgr *argon2PasswordMgr) Verify(password, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	var memory, time uint32
	var threads uint8
	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	decodedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
