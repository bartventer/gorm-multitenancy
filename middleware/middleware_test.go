package middleware

import (
	"net/http"
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
					Method:     http.MethodGet,
					Header:     map[string][]string{},
					Host:       "tenant1.example.com",
					RequestURI: "/",
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:host parts < 2",
			args: args{
				r: &http.Request{
					Method:     http.MethodGet,
					Header:     map[string][]string{},
					Host:       "invalid",
					RequestURI: "/",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid:host empty",
			args: args{
				r: &http.Request{
					Method:     http.MethodGet,
					Header:     map[string][]string{},
					Host:       ".invalid",
					RequestURI: "/",
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
