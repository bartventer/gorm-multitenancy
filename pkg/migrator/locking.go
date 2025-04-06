package migrator

import (
	"errors"
	"hash/fnv"
)

var (
	// ErrExecSQL is returned when the migrator fails to execute SQL.
	ErrExecSQL = errors.New("locking: failed to execute SQL")

	// ErrAcquireLock is returned when the migrator fails to acquire an advisory lock.
	ErrAcquireLock = errors.New("locking: failed to acquire advisory lock")

	// ErrReleaseLock is returned when the migrator fails to release an advisory lock.
	ErrReleaseLock = errors.New("locking: failed to release advisory lock")
)

// GenerateLockKey generates a lock key from a string.
//
// See the [benchmark results] for performance comparisons between different hashing algorithms.
//
// [benchmark results]: https://github.com/bartventer/gorm-multitenancy/blob/master/docs/LOCKING.md
func GenerateLockKey(s string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	return hasher.Sum64()
}
