package middleware

import "testing"

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		domainURL string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid domainURL with subdomain",
			domainURL: "https://sub.example.com",
			want:      "sub",
			wantErr:   false,
		},
		{
			name:      "valid domainURL with subdomain and port",
			domainURL: "https://sub.example.com:8080",
			want:      "sub",
			wantErr:   false,
		},
		{
			name:      "valid domainURL with subdomain, no scheme",
			domainURL: "sub.example.com",
			want:      "sub",
			wantErr:   false,
		},
		{
			name:      "valid domainURL with subdomain and port, no scheme",
			domainURL: "sub.example.com:8080",
			want:      "sub",
			wantErr:   false,
		},
		{
			name:      "invalid domainURL without subdomain",
			domainURL: "https://example.com",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid domainURL without subdomain but with port",
			domainURL: "https://example.com:8080",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid domainURL without subdomain, no scheme",
			domainURL: "example.com",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid domainURL without subdomain but with port, no scheme",
			domainURL: "example.com:8080",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid URL",
			domainURL: "this is not a valid URL",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid subdomain starting with pg_",
			domainURL: "https://pg_sub.example.com",
			want:      "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSubdomain(tt.domainURL)
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
