// Package crypto provides pure Argon2id password hashing and verification.
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 KB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

var ErrInvalidHash = errors.New("invalid argon2id hash format")

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return encodeHash(salt, hash), nil
}

func VerifyPassword(password, storedHash string) (bool, error) {
	salt, expectedHash, err := decodeHash(storedHash)
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

func encodeHash(salt, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash)
}

func decodeHash(encoded string) (salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	// Expected: ["", "argon2id", "v=19", "m=65536,t=1,p=4", "<salt>", "<hash>"]
	if len(parts) != 6 {
		return nil, nil, ErrInvalidHash
	}
	if parts[1] != "argon2id" {
		return nil, nil, ErrInvalidHash
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("decode salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	return salt, hash, nil
}
