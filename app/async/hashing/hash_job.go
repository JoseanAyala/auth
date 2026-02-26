package hasher

import (
	"auth-as-a-service/sdk/crypto"
)

// HashJob submits a password for hashing. Read the result from Result.
type HashJob struct {
	Password string
	Result   chan HashResult
}

type HashResult struct {
	Hash string
	Err  error
}

func (j HashJob) execute() {
	hash, err := crypto.HashPassword(j.Password)
	j.Result <- HashResult{Hash: hash, Err: err}
}
