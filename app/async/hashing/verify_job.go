package hasher

import (
	"auth-as-a-service/sdk/crypto"
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
	match, err := crypto.VerifyPassword(j.Password, j.StoredHash)
	j.Result <- VerifyResult{Match: match, Err: err}
}
