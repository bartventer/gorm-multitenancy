package driver

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *URL
		wantErr bool
	}{
		{
			name: "URL without special format",
			args: args{rawURL: "postgres://user:password@localhost:5432/dbname?sslmode=disable"},
			want: &URL{
				URL: &url.URL{
					Scheme:   "postgres",
					Path:     "/dbname",
					Host:     "localhost:5432",
					User:     url.UserPassword("user", "password"),
					RawQuery: "sslmode=disable",
				},
				raw:      "postgres://user:password@localhost:5432/dbname?sslmode=disable",
				standard: "postgres://user:password@localhost:5432/dbname?sslmode=disable",
			},
			wantErr: false,
		},
		{
			name: "URL with @tcp format",
			args: args{rawURL: "mysql://user:password@tcp(localhost:3306)/dbname"},
			want: &URL{
				URL: &url.URL{
					Scheme: "mysql",
					Host:   "localhost:3306",
					Path:   "/dbname",
					User:   url.UserPassword("user", "password"),
				},
				raw:      "mysql://user:password@tcp(localhost:3306)/dbname",
				standard: "mysql://user:password@localhost:3306/dbname",
			},
			wantErr: false,
		},
		{
			name:    "invalid URL",
			args:    args{rawURL: ":foo"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.args.rawURL)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.args.rawURL, got.Raw())
			assert.Equal(t, tt.want.standard, got.standard)
		})
	}
}

func Test_normalizeURL(t *testing.T) {
	type args struct {
		urlstr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "standard URL",
			args: args{urlstr: "http://localhost:8080"},
			want: "http://localhost:8080",
		},
		{
			name: "URL with @tcp format",
			args: args{urlstr: "@tcp(localhost:3306)/dbname"},
			want: "@localhost:3306/dbname",
		},
		{
			name: "URL without special format",
			args: args{urlstr: "mysql://user:pass@localhost:3306/dbname"},
			want: "mysql://user:pass@localhost:3306/dbname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeURL(tt.args.urlstr); got != tt.want {
				t.Errorf("normalizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDSNQueryParams(t *testing.T) {
	type backoffOptions struct {
		MaxRetries  int           `mapstructure:"max_retries"`
		Interval    time.Duration `mapstructure:"retry_interval"`
		MaxInterval time.Duration `mapstructure:"retry_max_interval"`
	}

	type dsnOptions struct {
		DisableRetry bool           `mapstructure:"disable_retry"`
		Retry        backoffOptions `mapstructure:",squash"`
	}
	tests := []struct {
		name     string
		dsn      string
		wantOpts dsnOptions
		wantErr  bool
	}{
		{
			name: "empty DSN",
			dsn:  "",
			wantOpts: dsnOptions{
				DisableRetry: false,
				Retry:        backoffOptions{},
			},
			wantErr: false,
		},
		{
			name: "valid DSN",
			dsn:  "mysql://user:password@tcp(localhost:3306)/dbname?disable_retry=true&max_retries=6&retry_interval=2s&retry_max_interval=30s",
			wantOpts: dsnOptions{
				DisableRetry: true,
				Retry: backoffOptions{
					MaxRetries:  6,
					Interval:    2 * time.Second,
					MaxInterval: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid DSN",
			dsn:     "%gh&%ij",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseDSNQueryParams[dsnOptions](tt.dsn)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantOpts, opts)
			}
		})
	}
}
