package nethttp

import (
	"fmt"
	"testing"
)

func ExampleExtractSubdomain() {
	subdomain, err := ExtractSubdomain("test.domain.com")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Subdomain:", subdomain)
	}

	subdomain, err = ExtractSubdomain("test.sub.domain.com")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Subdomain:", subdomain)
	}

	// Output:
	// Subdomain: test
	// Subdomain: test
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		want    string
		wantErr bool
	}{
		{"valid:subdomain", "sub.example.com", "sub", false},
		{"valid:subdomain-with-port", "sub.example.com:8080", "sub", false},
		{"invalid:no-subdomain", "example.com", "", true},
		{"invalid:no-subdomain-with-port", "example.com:8080", "", true},
		{"invalid:pg-subdomain", "pg_sub.example.com", "", true},
		{"invalid:host-is-ipv4-RFC-3986", "192.168.0.1", "", true},
		{"invalid:host-is-ipv4-RFC-3986-with-port", "192.168.0.1:8080", "", true},
		{"invalid:host-is-ipv6-RFC-3986", "[fe80::1]", "", true},
		{"invalid:host-is-ipv6-RFC-3986-with-port", "[fe80::1]:8080", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSubdomain(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSubdomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractSubdomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
