package migrator

import (
	"errors"
	"hash/fnv"
	"math/big"
)

var (
	ErrExecSQL     = errors.New("locking: failed to execute SQL")
	ErrAcquireLock = errors.New("locking: failed to acquire advisory lock")
	ErrReleaseLock = errors.New("locking: failed to release advisory lock")
)

// GenerateLockKey generates a lock key from a string.
//
// See the [benchmark results] for performance comparisons between different hashing algorithms.
//
// [benchmark results]: https://github.com/bartventer/gorm-multitenancy/blob/master/docs/LOCKING.md
func GenerateLockKey(s string) int64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	bigInt := new(big.Int).SetUint64(h.Sum64())
	return bigInt.Int64()
}
