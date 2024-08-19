package nethttp

import (
	"errors"
	"testing"
)

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		opts          []SubdomainOption
		want          string
		wantErr       bool
		expectedError error
	}{
		{"valid", "sub.example.com", nil, "sub", false, nil},
		{"valid:with-port", "sub.example.com:8080", nil, "sub", false, nil},
		{"invalid host:no subdomain", "example.com", nil, "", true, ErrInvalidHost},
		{"invalid host:no subdomain (with port)", "example.com:8080", nil, "", true, ErrInvalidHost},
		{"invalid host:IPv4 RFC 3986", "192.168.0.1", nil, "", true, ErrInvalidHost},
		{"invalid host:IPv4 RFC 3986 (with port)", "192.168.0.1:8080", nil, "", true, ErrInvalidHost},
		{"invalid host:IPv6 RFC 3986", "[fe80::1]", nil, "", true, ErrInvalidHost},
		{"invalid host:IPv6 RFC 3986 (with port)", "[fe80::1]:8080", nil, "", true, ErrInvalidHost},
		{"invalid subdomain:disallowed prefix", "pg_sub.example.com", []SubdomainOption{WithDisallowedPrefixes("pg_")}, "", true, ErrInvalidSubdomain},
		{"invalid subdomain:disallowed subdomain", "blacklisted.example.com", []SubdomainOption{WithDisallowedSubdomains("blacklisted")}, "", true, ErrInvalidSubdomain},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSubdomain(tt.host, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSubdomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(err, tt.expectedError) {
				t.Errorf("ExtractSubdomain() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractSubdomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
