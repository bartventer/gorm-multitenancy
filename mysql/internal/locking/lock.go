// Package locking provides functions for acquiring and releasing MySQL advisory locks.
package locking

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

// SQL statements for MySQL advisory locks.
// https://dev.mysql.com/doc/refman/8.4/en/locking-functions.html
const (
	sqlGetLock     = "SELECT GET_LOCK(?, 0)"
	sqlReleaseLock = "SELECT RELEASE_LOCK(?)"
)

// acquireError wraps [migrator.ErrAcquireLock] with additional context.
type acquireError struct {
	result sql.NullInt64
	key    string
	msg    string
}

func (e acquireError) Error() string {
	return fmt.Sprintf("(GET_LOCK) failed to acquire lock for key %s: %v (result; valid: %t, value: %d)", e.key, e.msg, e.result.Valid, e.result.Int64)
}

func (e acquireError) Unwrap() error {
	return migrator.ErrAcquireLock
}

// releaseError wraps [migrator.ErrReleaseLock] with additional context.
type releaseError struct {
	result sql.NullInt64
	key    string
	msg    string
}

func (e releaseError) Error() string {
	return fmt.Sprintf("(RELEASE_LOCK) failed to release lock for key %s: %v (result; valid: %t, value: %d)", e.key, e.msg, e.result.Valid, e.result.Int64)
}

func (e releaseError) Unwrap() error {
	return migrator.ErrReleaseLock
}

func encodeKey(key string) string {
	if len(key) <= 64 {
		return key
	}
	hash := sha1.Sum([]byte(key))
	return hex.EncodeToString(hash[:]) // SHA-1 hash is always 40 characters in hex
}

type LockConfig struct {
	DisableEncode bool
}

// lock represents a MySQL advisory lock.
type lock struct {
	tx  *gorm.DB
	key string
}

// acquire acquires a MySQL advisory lock.
func (a *lock) acquire() (func() error, error) {
	var result sql.NullInt64

	err := a.tx.Raw(sqlGetLock, a.key).Scan(&result).Error
	if err != nil {
		return nil, errors.Join(
			err,
			acquireError{result: result, key: a.key, msg: fmt.Sprintf("unexpected error: %v", err)},
		)
	}

	if !result.Valid {
		return nil, acquireError{result: result, key: a.key, msg: "unexpected error: NULL result"}
	}

	switch result.Int64 {
	case 1:
		// Lock acquired successfully
	case 0:
		return nil, acquireError{result: result, key: a.key, msg: "a timeout occurred"}
	default:
		return nil, acquireError{result: result, key: a.key, msg: "unexpected result, expected a 0 or 1"}
	}

	// Return a release function
	return func() error {
		return a.release()
	}, nil
}

// release releases a MySQL advisory lock.
func (a *lock) release() error {
	encodedKey := encodeKey(a.key)
	var result sql.NullInt64

	err := a.tx.Raw(sqlReleaseLock, encodedKey).Scan(&result).Error
	if err != nil {
		return errors.Join(
			err,
			releaseError{result: result, key: encodedKey, msg: fmt.Sprintf("unexpected error: %v", err)},
		)
	}

	if !result.Valid {
		return releaseError{result: result, key: encodedKey, msg: "lock does not exist (NULL error)"}
	}

	switch result.Int64 {
	case 1:
		return nil
	case 0:
		return releaseError{result: result, key: encodedKey, msg: "lock was not held by this session"}
	default:
		return releaseError{result: result, key: encodedKey, msg: "unexpected result, expected a 0 or 1"}
	}
}

func acquireLock(tx *gorm.DB, key string) (func() error, error) {
	lock := &lock{tx: tx}
	lock.key = encodeKey(key)
	return lock.acquire()
}

// Acquire acquires a MySQL advisory lock with optional retry logic.
func Acquire(tx *gorm.DB, lockKey string, opts ...Option) (release func() error, err error) {
	options := &Options{}
	options.apply(opts...)

	if options.Retry != nil {
		acquireFunc := func() error {
			releaseFn, acquireErr := acquireLock(tx, lockKey)
			if acquireErr != nil {
				return acquireErr
			}
			release = releaseFn
			return nil
		}
		err = backoff.Retry(acquireFunc, func(o *backoff.Options) {
			*o = *options.Retry
		})
	} else {
		release, err = acquireLock(tx, lockKey)
	}

	return release, err
}

// Options represents the locking options.
type Options struct {
	Retry *backoff.Options
}

// Option is a function that applies an option to an Options instance.
type Option func(*Options)

// apply applies the given options to the Options instance.
func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithRetry sets the retry options.
func WithRetry(opts *backoff.Options) Option {
	return func(o *Options) {
		o.Retry = opts
	}
}
