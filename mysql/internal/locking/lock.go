// Package locking provides functions for acquiring and releasing MySQL advisory locks.
package locking

import (
	"database/sql"
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

// SQL statements for MySQL advisory locks.
// https://dev.mysql.com/doc/refman/8.4/en/locking-functions.html
const (
	// GET_LOCK(str, timeout) → int (1: lock acquired, 0: lock not acquired, NULL: an error occurred).
	sqlGetLock = "SELECT GET_LOCK('%s', 0)"

	// RELEASE_LOCK(str) → int (1: lock released, 0: lock not released, NULL: lock does not exist).
	sqlReleaseLock = "SELECT RELEASE_LOCK('%s')"
)

// lock represents a MySQL advisory lock.
type lock struct {
	tx      *gorm.DB
	lockKey string
}

func (a *lock) execute(sqlstr string) (bool, error) {
	var result sql.NullInt64
	if err := a.tx.Raw(sqlstr).Scan(&result).Error; err == nil {
		if !result.Valid {
			return false, migrator.ErrAcquireLock
		}
		switch result.Int64 {
		case 1:
			return true, nil
		case 0:
			return false, migrator.ErrAcquireLock
		}
	}
	return false, fmt.Errorf("%w: %s", migrator.ErrExecSQL, sqlstr)
}

func (a *lock) acquire() (func() error, error) {
	sqlstr := fmt.Sprintf(sqlGetLock, a.lockKey)
	ok, err := a.execute(sqlstr)
	if err != nil || !ok {
		return nil, fmt.Errorf("%w for key %s: %v", migrator.ErrAcquireLock, a.lockKey, err)
	}

	return a.release, nil
}

func (a *lock) release() error {
	sqlstr := fmt.Sprintf(sqlReleaseLock, a.lockKey)
	success, err := a.execute(sqlstr)
	if err != nil || !success {
		return fmt.Errorf("%w for key %s: %v", migrator.ErrReleaseLock, a.lockKey, err)
	}
	return nil
}

// Options represents the locking options.
type Options struct {
	Retry *backoff.Options
}

// Option is a function that applies an option to an Options instance.
type Option func(*Options)

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

// acquire acquires a MySQL advisory lock.
func acquire(tx *gorm.DB, lockKey string) (func() error, error) {
	lock := &lock{tx: tx, lockKey: lockKey}
	return lock.acquire()
}

// Acquire acquires a MySQL advisory lock.
// Returns a release function and an error.
//
// It's the responsibility of the caller to release the lock by calling the release function.
func Acquire(tx *gorm.DB, lockKey string, opts ...Option) (release func() error, err error) {
	options := &Options{}
	options.apply(opts...)

	if options.Retry != nil {
		acquireFunc := func() error {
			releaseFn, acquireErr := acquire(tx, lockKey)
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
		release, err = acquire(tx, lockKey)
	}

	return release, err
}
