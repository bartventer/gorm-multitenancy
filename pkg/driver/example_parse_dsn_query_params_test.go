package driver_test

import (
	"fmt"
	"time"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
)

func ExampleParseDSNQueryParams() {
	type backoffOptions struct {
		MaxRetries  int           `mapstructure:"max_retries"`
		Interval    time.Duration `mapstructure:"retry_interval"`
		MaxInterval time.Duration `mapstructure:"retry_max_interval"`
	}

	type dsnOptions struct {
		DisableRetry bool           `mapstructure:"disable_retry"`
		Retry        backoffOptions `mapstructure:",squash"`
	}

	dsn := "mysql://user:password@tcp(localhost:3306)/dbname?disable_retry=true&max_retries=6&retry_interval=2s&retry_max_interval=30s"
	opts, err := driver.ParseDSNQueryParams[dsnOptions](dsn)
	if err != nil {
		panic(err)
	}

	// Use the parsed options.
	fmt.Printf("DisableRetry: %v\n", opts.DisableRetry)
	fmt.Printf("MaxRetries: %d\n", opts.Retry.MaxRetries)
	fmt.Printf("Interval: %s\n", opts.Retry.Interval)
	fmt.Printf("MaxInterval: %s\n", opts.Retry.MaxInterval)

	// Output:
	// DisableRetry: true
	// MaxRetries: 6
	// Interval: 2s
	// MaxInterval: 30s
}
