// Package hasher
package hasher

import (
	"errors"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 KB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

var ErrInvalidHash = errors.New("invalid argon2id hash format")
