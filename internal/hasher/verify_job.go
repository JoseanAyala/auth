package hasher

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// VerifyJob submits a password + stored hash for verification. Read the result from Result.
type VerifyJob struct {
	Password   string
	StoredHash string
	Result     chan VerifyResult
}

type VerifyResult struct {
	Match bool
	Err   error
}

func (j VerifyJob) execute() {
	match, err := verifyPassword(j.Password, j.StoredHash)
	j.Result <- VerifyResult{Match: match, Err: err}
}

func verifyPassword(password, storedHash string) (bool, error) {
	salt, expectedHash, err := decodeHash(storedHash)
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
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
