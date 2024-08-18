// Package locking provides functions for acquiring and releasing PostgreSQL advisory locks.
package locking

import (
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

// SQL statements for PostgreSQL advisory locks.
// https://www.postgresql.org/docs/16/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
const (
	// pg_try_advisory_xact_lock ( key bigint ) → boolean.
	sqlTryAdvisoryXactLock = "SELECT pg_try_advisory_xact_lock(?)"
)

// lock represents a PostgreSQL transaction-level advisory lock.
type lock struct {
	tx      *gorm.DB
	lockKey string
}

func (p *lock) execute(sqlstr string, args ...interface{}) (bool, error) {
	var result bool
	if err := p.tx.Raw(sqlstr, args...).Scan(&result).Error; err == nil && result {
		return true, nil
	}
	return false, fmt.Errorf("%w: %s", migrator.ErrExecSQL, sqlstr)
}

func (p *lock) acquire() error {
	key := migrator.GenerateLockKey(p.lockKey)
	ok, err := p.execute(sqlTryAdvisoryXactLock, key)
	if err != nil || !ok {
		return fmt.Errorf("%w for key %s: %v", migrator.ErrAcquireLock, p.lockKey, err)
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

// AcquireXact acquires a PostgreSQL transaction-level advisory lock.
// The caller is responsible for ensuring that a transaction is active,
// and that the lock is released after use.
func AcquireXact(tx *gorm.DB, lockKey string, opts ...Option) error {
	options := &Options{}
	options.apply(opts...)

	l := &lock{tx: tx, lockKey: lockKey}

	if options.Retry != nil {
		acquireFunc := func() error {
			return l.acquire()
		}
		return backoff.Retry(acquireFunc, func(o *backoff.Options) {
			*o = *options.Retry
		})
	} else {
		return l.acquire()
	}
}
