package middleware

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
)

func TestDefaultTenantFromSubdomain(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid:test-from-subdomain",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.example.com",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:host parts < 2",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "invalid",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid:host empty",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   ".invalid",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid:test-from-subdomain-with-port",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.example.com:8080",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "valid:test-from-subdomain-with-multiple-subdomains",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.sub.example.com",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:no-subdomain",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "example.com",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid:host-is-ip",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "192.168.1.1",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultTenantFromSubdomain(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("TenantFromSubdomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromSubdomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTenantFromHeader(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Host:   "example.com",
					Header: map[string][]string{
						"X-Tenant": {"tenant1"},
					},
					RequestURI: "/",
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:header empty",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Host:   "example.com",
					Header: map[string][]string{
						"X-Tenant": {" "},
					},
					RequestURI: "/",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultTenantFromHeader(tt.args.r)
			if err != nil && !tt.wantErr {
				t.Errorf("TenantFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
