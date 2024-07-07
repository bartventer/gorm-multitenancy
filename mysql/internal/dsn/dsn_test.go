package dsn

import "testing"

func TestStripSchemeFromURL(t *testing.T) {
	tests := []struct {
		name     string
		urlstr   string
		expected string
	}{
		{
			name:     "with scheme",
			urlstr:   "mysql://user:password@tcp(localhost:3306)/dbname",
			expected: "user:password@tcp(localhost:3306)/dbname",
		},
		{
			name:     "without scheme",
			urlstr:   "user:password@tcp(localhost:3306)/dbname",
			expected: "user:password@tcp(localhost:3306)/dbname",
		},
		{
			name:     "empty string",
			urlstr:   "",
			expected: "",
		},
		{
			name:     "only scheme",
			urlstr:   "http://",
			expected: "",
		},
		{
			name:     "scheme with no following slash",
			urlstr:   "http:/malformed/url",
			expected: "http:/malformed/url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripSchemeFromURL(tt.urlstr)
			if got != tt.expected {
				t.Errorf("stripSchemeFromURL(%q) = %q, want %q", tt.urlstr, got, tt.expected)
			}
		})
	}
}
