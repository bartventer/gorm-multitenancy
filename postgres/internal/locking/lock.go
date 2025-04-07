// Package locking provides functions for acquiring and releasing PostgreSQL advisory locks.
package locking

import (
	"errors"
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

// SQL statements for PostgreSQL advisory locks.
// https://www.postgresql.org/docs/16/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
const (
	// pg_try_advisory_xact_lock ( key bigint ) → boolean.
	// This will either obtain the lock immediately and return true, or return false without waiting if the lock cannot be acquired immediately.
	sqlTryAdvisoryXactLock = "SELECT pg_try_advisory_xact_lock(?)"
)

type lock struct {
	tx  *gorm.DB
	key string
}

func (p *lock) acquire() error {
	key := migrator.GenerateLockKey(p.key)
	var ok bool
	if err := p.tx.Raw(sqlTryAdvisoryXactLock, key).Scan(&ok).Error; err != nil || !ok {
		return errors.Join(
			migrator.ErrAcquireLock,
			migrator.ErrExecSQL,
			fmt.Errorf("failed to acquire lock for key %s: %v", p.key, err),
		)
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

	l := &lock{tx: tx, key: lockKey}

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
