package backoff

import (
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		name        string
		fn          func() error
		opts        []Option
		expectError bool
	}{
		{
			name: "Success on first try",
			fn: func() error {
				return nil
			},
			opts:        []Option{WithMaxRetries(3), WithRetryInterval(1 * time.Second)},
			expectError: false,
		},
		{
			name: "Fail all retries",
			fn: func() error {
				return errors.New("fail")
			},
			opts:        []Option{WithMaxRetries(3), WithRetryInterval(1 * time.Second)},
			expectError: true,
		},
		{
			name: "Success on second try",
			fn: func() func() error {
				static := 0
				return func() error {
					static++
					if static == 2 {
						return nil
					}
					return errors.New("fail")
				}
			}(),
			opts:        []Option{WithMaxRetries(3), WithRetryInterval(1 * time.Second)},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Retry(tt.fn, tt.opts...)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}
