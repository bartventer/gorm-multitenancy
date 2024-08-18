// Package backoff provides exponential backoff retry logic.
package backoff

import (
	"cmp"
	"fmt"
	"log"
	"time"
)

// Default configuration values.
const (
	DefaultMaxRetries    = 3
	DefaultRetryInterval = 2 * time.Second
	DefaultMaxInterval   = 30 * time.Second
)

// Options is the configuration options for retry operations.
type Options struct {
	// MaxRetries is the maximum number of retry attempts.
	// Default is 3.
	MaxRetries int `json:"gmt_max_retries" mapstructure:"gmt_max_retries"`

	// Interval is the initial interval between retry attempts.
	// Default is 2 seconds.
	Interval time.Duration `json:"gmt_retry_interval" mapstructure:"gmt_retry_interval"`

	// MaxInterval is the maximum interval between retry attempts.
	// Default is 30 seconds.
	MaxInterval time.Duration `json:"gmt_retry_max_interval" mapstructure:"gmt_retry_max_interval"`
}

// Option is a function that applies an option to an Options instance.
type Option func(*Options)

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}

	o.MaxRetries = cmp.Or(max(o.MaxRetries, 0), DefaultMaxRetries)
	o.Interval = cmp.Or(max(o.Interval, 0), DefaultRetryInterval)
	o.MaxInterval = cmp.Or(max(o.MaxInterval, 0), DefaultMaxInterval)
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) Option {
	return func(o *Options) {
		o.MaxRetries = n
	}
}

// WithRetryInterval sets the initial interval between retry attempts.
func WithRetryInterval(i time.Duration) Option {
	return func(o *Options) {
		o.Interval = i
	}
}

// WithMaxInterval sets the maximum interval between retry attempts.
func WithMaxInterval(i time.Duration) Option {
	return func(o *Options) {
		o.MaxInterval = i
	}
}

// Retry executes the provided function with retry logic using exponential backoff.
func Retry(fn func() error, opts ...Option) error {
	o := &Options{}
	o.apply(opts...)

	interval := o.Interval
	for i := 0; i < o.MaxRetries; i++ {
		if err := fn(); err != nil {
			if i == o.MaxRetries-1 {
				return fmt.Errorf("backoff: max retries (%d) exceeded: %w", o.MaxRetries, err)
			}
			time.Sleep(interval)
			log.Printf("backoff: retrying after %s (attempt %d of %d) due to error: %v", interval, i+1, o.MaxRetries, err)
			interval *= 2
			if interval > o.MaxInterval {
				interval = o.MaxInterval
			}
		} else {
			return nil
		}
	}
	return nil
}
